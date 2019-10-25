package socketio

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/chilts/sid"
	"github.com/golang/glog"
	"github.com/gorilla/websocket"
)

type handshakeResponse struct {
	Sid          string
	PingInterval int
	PingTimeout  int
	Upgrades     []string
}

// Fulfilled by http.Client#Do.
type doClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// An interface that is fulfilled by websocket.Conn and allows for injecting a
// test connection.
type WSConn interface {
	ReadMessage() (int, []byte, error)
	WriteMessage(messageType int, data []byte) error
	Close() error
}

// Fulfilled by websocket.DefaultDialer#Dial.
type WSDialer interface {
	Dial(urlStr string, reqHeader http.Header) (
		WSConn, *http.Response, error)
}

// Used to control the heartbeat functionality.
const (
	stopBeating = iota
)

// Indicates an error during initialization of the Transport layer.
type transportError struct {
	message string
}

func (t *transportError) Error() string {
	return t.message
}

// A wrapper that provides thread-safe methods for interacting with the
// underlying Websocket layer.
type Transport interface {
	// Provides a thread-safe io.Writer Write method.
	io.Writer

	// Returns a pointer to a read channel. A new channel is returned with
	// each call, and all channels will receive a copy of all incoming
	// messages.
	GetReadChannel() (<-chan []byte, error)

	// Closes the underlying Websocket connection.
	Close()
}

// A set of channels used to convey incoming messages to listeners.
type outgoing struct {
	sync.RWMutex
	// A collection of channels for transmitting messages to consumers.
	channels []chan []byte
}

type transport struct {
	sync.RWMutex
	sync.Once
	// The wrapped Gorilla websocket.Conn.
	conn WSConn
	// A channel used to kill an ongoing heartbeat.
	quitHeartbeat chan<- int
	// A collection of outgoing channels returned by GetReadChannel.
	outgoing *outgoing
	// Used to buffer incoming message for writing.
	incoming chan []byte
	// True when this transport has been closed.
	closed bool
}

func (t *transport) Write(p []byte) (int, error) {
	t.RLock()
	defer t.RUnlock()
	if t.closed {
		return 0, &transportError{"Cannot write to a closed transport"}
	}
	t.incoming <- p
	return len(p), nil
}

func (t *transport) GetReadChannel() (<-chan []byte, error) {
	t.outgoing.RLock()
	defer t.outgoing.RUnlock()
	if t.closed {
		return nil, &transportError{
			"Cannot read from a closed transport"}
	}
	t.outgoing.channels = append(
		t.outgoing.channels, make(chan []byte, 1))
	return t.outgoing.channels[len(t.outgoing.channels)-1], nil
}

func (t *transport) Close() {
	t.Do(func() {
		// Send the close signal before marking the transport as closed.
		sendPacket(t, Close)

		t.quitHeartbeat <- stopBeating
		for _, ch := range t.outgoing.channels {
			close(ch)
		}
		close(t.incoming)

		t.Lock()
		t.closed = true
		t.Unlock()
	})
}

func (t *transport) startReadAndWriteRoutines() {
	go func(ch <-chan []byte) {
		for message := range ch {
			if glog.V(3) {
				glog.Infof("Writing message: %s", message)
			}
			err := t.conn.WriteMessage(
				websocket.TextMessage, message)
			if err != nil {
				glog.Errorf(
					"Failed to write message %q: %s",
					message, err)
			}
		}
		t.conn.Close()
	}(t.incoming)
	go func() {
		for {
			_, message, err := t.conn.ReadMessage()
			if err != nil {
				glog.Errorf(
					"Error reading from websocket: %s",
					err)
				return
			}
			if len(message) == 0 {
				continue
			}
			if glog.V(3) {
				glog.Infof(
					"Received websocket message: %s",
					message)
			}
			t.RLock()
			if t.closed {
				if glog.V(3) {
					errTxt := "Dropping message %s;" +
						"Transport closed"
					glog.Warningf(errTxt, message)
				}
				t.RUnlock()
				break
			}
			t.RUnlock()
			for _, ch := range t.outgoing.channels {
				ch <- message
			}
		}
	}()
}

// Performs an HTTP request and returns the response body. If there is an error
// the io.ReaderCloser will be nil.
func makeHTTPRequest(client doClient, method string,
	to string, body io.Reader) (io.ReadCloser, error) {
	glog.Infof("Making %s request to: %v", method, to)
	req, err := http.NewRequest(method, to, body)
	if err != nil {
		if glog.V(3) {
			glog.Warningf(
				"Failed to construct request: %s", err)
		}
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		if glog.V(3) {
			glog.Warningf(
				"Failed to make request: %s", err)
		}
		return nil, err
	}
	if resp == nil {
		return nil, &transportError{fmt.Sprintf(
			"No response body from %s", to)}
	}
	return resp.Body, nil
}

// Performs the initial GET connection to the SocketIO endpoint. If it it
// successful, it will set the session id (sid) parameter on the endpoint.
func connect(endpoint Endpoint, client doClient) (*handshakeResponse, error) {
	handshakeUrl := endpoint.GetHTTPUrl()
	resp, err := makeHTTPRequest(client, "GET", handshakeUrl, nil)
	if err != nil {
		glog.Errorf("Error connecting to IEX: %s", err)
		return nil, err
	}
	var handshake handshakeResponse
	err = HTTPToJSON(resp, []interface{}{&handshake})
	if err != nil {
		glog.Errorf("Error parsing handshake response: %s", err)
		return nil, err
	}
	canUpgradeToWs := false
	for _, val := range handshake.Upgrades {
		if val == "websocket" {
			canUpgradeToWs = true
		}

	}
	if !canUpgradeToWs {
		return nil, &transportError{
			"Websocket upgrade not found"}
	}
	endpoint.SetSid(handshake.Sid)
	return &handshake, nil
}

// Joins the default namespace. Returns an error if there was an unexpected
// server response.
func joinDefaultNsp(endpoint Endpoint, client doClient) error {
	encoder := NewHTTPEncoder("/")
	reader, err := encoder.EncodePacket(4, 0)
	if err != nil {
		glog.Errorf("Error encoding namespace connection: %s", err)
		return err
	}
	resp, err := makeHTTPRequest(
		client, "POST", endpoint.GetHTTPUrl(), reader)
	if err != nil {
		glog.Errorf("Error connecting to the empty room: %s", err)
		return err
	}
	var packetMetadata packetMetadata
	err = HTTPToJSON(resp, []interface{}{&packetMetadata})
	if err != nil {
		glog.Errorf("Error parsing namespace response: %s", err)
		return err
	}
	if packetMetadata.PacketType != 4 || packetMetadata.MessageType != 0 {
		glog.Errorf("Unexpected namespace response: %+v",
			packetMetadata)
		return &transportError{fmt.Sprintf(
			"Unexpected namespace response: %+v",
			packetMetadata)}
	}
	return nil
}

func sendPacket(transport Transport, packetType PacketType) {
	encoder := NewWSEncoder("")
	reader, err := encoder.EncodePacket(packetType, -1)
	if err != nil {
		glog.Warningf(
			"Could not encode probe message: %s", err)
	}
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		glog.Warningf(
			"Could not read encoded message: %s", err)
	}
	_, err = transport.Write(data)
	if err != nil {
		glog.Warningf(
			"Error writing probe message: %s", err)
		return
	}
	if glog.V(3) {
		glog.Infof("Sent packet %q", data)
	}
}

// Starts a go routine that sends a ping message on the given Transport every
// "ping" milliseconds.
func startHeartbeat(
	transport Transport, quitChan <-chan int, pingMillis int) {
	duration, err := time.ParseDuration(strconv.Itoa(pingMillis) + "ms")
	if err != nil {
		glog.Fatalf("Could not start heartbeat: %s", err)
	}
	heartbeat := time.NewTicker(duration)
	go func() {
		for {
			select {
			case <-quitChan:
				heartbeat.Stop()
			case t := <-heartbeat.C:
				if glog.V(3) {
					glog.Infof("Heartbeating at %v", t)
				}
				sendPacket(transport, Ping)
			}
		}
	}()
}

// Upgrades from an HTTPS to a Websocket connection. This method starts
// regular probe polling and sends an upgrade message before returning
// the websocket.Conn object. If an error occurs, the returned Transport
// is nil. The ping interval is used to start a hearbeat polling mechanism.
func upgrade(endpoint Endpoint, dialer WSDialer, ping int) (Transport, error) {
	to := endpoint.GetWSUrl()
	if glog.V(3) {
		glog.Infof("Opening websocket connection to: %s", to)
	}
	conn, _, err := dialer.Dial(to, nil)
	if err != nil {
		glog.Errorf("Error opening websocket connection: %s", err)
		return nil, err
	}
	if glog.V(3) {
		glog.Info("Websocket connection established; sending upgrade")
	}
	encoder := NewWSEncoder("")
	reader, err := encoder.EncodePacket(5, -1)
	if err != nil {
		glog.Errorf("Error upgrading connection: %s", err)
		return nil, err
	}
	quitChannel := make(chan int)
	trans := &transport{
		conn:          conn,
		quitHeartbeat: quitChannel,
		outgoing:      &outgoing{channels: make([]chan []byte, 0)},
		incoming:      make(chan []byte, 1),
	}
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		glog.Errorf("Error reading upgrade: %s", err)
	}
	_, err = trans.Write(data)
	if err != nil {
		glog.Errorf("Error upgrading connection: %s", err)
		return nil, err
	}
	startHeartbeat(trans, quitChannel, ping)
	trans.startReadAndWriteRoutines()

	return trans, nil
}

// Returns a new Transport object backed by an open Websocket connection
// or an error if one occurs.
func NewTransport(client doClient, dialer WSDialer) (Transport, error) {
	endpoint := NewIEXEndpoint(sid.IdBase64)
	handshake, err := connect(endpoint, client)
	if err != nil {
		return nil, err
	}
	err = joinDefaultNsp(endpoint, client)
	if err != nil {
		return nil, err
	}
	return upgrade(endpoint, dialer, handshake.PingInterval)
}
