package socketio

import (
	"bytes"
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

	// Adds a callback to be triggered when packets from the given
	// namespace are received. Returns a unique ID that can be used
	// to remove the callback later. If an error occurs, returns -1
	// and the error.
	AddPacketCallback(
		namespace string, callback func(PacketData)) (int, error)

	// Removes a callback using the ID that was returned when the
	// callback was added. If either the namespace or the ID is does
	// not exist, this method is a no-op.
	RemovePacketCallback(namespace string, id int) error

	// Closes the underlying websocket connection.
	Close()
}

// A set of callbacks used to convey incoming messages to listeners.
type outgoing struct {
	sync.RWMutex

	// The next ID to use when adding a callback.
	nextId int
	// A collection of channels for transmitting messages to consumers.
	callbacks map[int]func(PacketData)
}

func newOutgoing() *outgoing {
	return &outgoing{
		nextId:    0,
		callbacks: make(map[int]func(PacketData)),
	}
}

// Adds a PacketData callback and returns an identifier to be used when later
// removing the callback.
func (o *outgoing) AddCallback(callback func(PacketData)) int {
	o.Lock()
	defer o.Unlock()
	o.nextId++
	o.callbacks[o.nextId] = callback
	return o.nextId
}

// Deletes the callback associated with the given ID. This function is a no-op
// if the ID is non-existent.
func (o *outgoing) RemoveCallback(id int) {
	o.Lock()
	defer o.Unlock()
	delete(o.callbacks, id)
}

func (o *outgoing) Callbacks() map[int]func(PacketData) {
	o.Lock()
	defer o.Unlock()
	// Make a copy for thread safety.
	copy := make(map[int]func(PacketData))
	for key, val := range o.callbacks {
		copy[key] = val
	}
	return copy
}

type transport struct {
	sync.RWMutex
	sync.Once

	// The wrapped Gorilla websocket.Conn.
	conn WSConn
	// A collection of callbacks keyed by namespace names.
	outgoing map[string]*outgoing
	// True when this transport has been closed.
	closed bool
}

func (t *transport) Write(message []byte) (int, error) {
	t.RLock()
	closed := t.closed
	t.RUnlock()
	if closed {
		return 0, &transportError{"Cannot write to a closed transport"}
	}
	if glog.V(3) {
		glog.Infof("Writing message: %s", string(message))
	}
	err := t.conn.WriteMessage(
		websocket.TextMessage, message)
	if err != nil {
		glog.Errorf(
			"Failed to write message %q: %s",
			string(message), err)
	}
	return len(message), nil
}

func (t *transport) AddPacketCallback(
	namespace string, callback func(PacketData)) (int, error) {
	t.Lock()
	closed := t.closed
	t.Unlock()
	if closed {
		return -1, &transportError{
			"Cannot add a callback to a closed transport"}
	}
	t.Lock()
	defer t.Unlock()
	if _, ok := t.outgoing[namespace]; !ok {
		t.outgoing[namespace] = newOutgoing()
	}
	return t.outgoing[namespace].AddCallback(callback), nil
}

func (t *transport) RemovePacketCallback(namespace string, id int) error {
	t.Lock()
	closed := t.closed
	t.Unlock()
	if closed {
		return &transportError{
			"Cannot remove a callback from a closed transport"}
	}
	t.Lock()
	defer t.Unlock()
	if val, ok := t.outgoing[namespace]; ok {
		val.RemoveCallback(id)
		if len(val.Callbacks()) == 0 {
			delete(t.outgoing, namespace)
		}
	}
	return nil
}

func (t *transport) Close() {
	t.Do(func() {
		// Send the close signal before marking the transport as closed.
		sendPacket(t, Close)
		t.conn.Close()
		t.Lock()
		t.closed = true
		t.Unlock()
	})
}

func (t *transport) startReadLoop() {
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
		closed := t.closed
		t.RUnlock()
		if closed {
			if glog.V(3) {
				errTxt := "Dropping message %s;" +
					"Transport closed"
				glog.Warningf(errTxt, message)
			}
			break
		}
		var metadata PacketData
		remaining := ParseMetadata(string(message), &metadata)
		metadata.Data = remaining
		if val, ok := t.outgoing[metadata.Namespace]; ok {
			callbacks := val.Callbacks()
			for _, callback := range callbacks {
				go callback(metadata)
			}
		}
	}

}

// Starts a go routine that sends a ping message on the given Transport every
// "ping" milliseconds.
func (t *transport) startHeartbeat(pingMillis int) {
	duration, err := time.ParseDuration(strconv.Itoa(pingMillis) + "ms")
	if err != nil {
		glog.Fatalf("Could not start heartbeat: %s", err)
	}
	heartbeat := time.NewTicker(duration)
	go func() {
		for {
			select {
			case time := <-heartbeat.C:
				t.RLock()
				closed := t.closed
				t.RUnlock()
				if closed {
					if glog.V(5) {
						glog.Info("Stop heart beat")
					}
					return
				}
				if glog.V(3) {
					glog.Infof("Heartbeating at %v", time)
				}
				sendPacket(t, Ping)
			}
		}
	}()
}

// Performs an HTTP request and returns the body. If there is an error the
// io.Reader will be nil.
func makeHTTPRequest(client doClient, to string) (io.Reader, error) {
	if glog.V(3) {
		glog.Infof("Making GET request to: %v", to)
	}
	req, err := http.NewRequest("GET", to, nil)

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
	if glog.V(5) {
		glog.Infof("Response: %v", resp)
		glog.Infof("Status: %v", resp.Status)
		glog.Infof("Headers: %v", resp.Header)
	}
	defer resp.Body.Close()
	respBytes, _ := ioutil.ReadAll(resp.Body)
	respBuffer := bytes.NewBuffer(respBytes)
	return respBuffer, nil
}

// Performs the initial GET connection to the SocketIO endpoint. If it it
// successful, it will set the session id (sid) parameter on the endpoint.
func connect(endpoint Endpoint, client doClient) (*handshakeResponse, error) {
	handshakeUrl := endpoint.GetHTTPUrl()
	resp, err := makeHTTPRequest(client, handshakeUrl)
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
	// Making a get request with the SID automatically joins the default
	// namespace.
	resp, err = makeHTTPRequest(client, endpoint.GetHTTPUrl())
	if err != nil {
		glog.Errorf("Error making status GET: %s", err)
		return nil, err
	}
	var packetData PacketData
	err = HTTPToJSON(resp, []interface{}{&packetData})
	if err != nil {
		glog.Errorf("Error parsing handshake response: %s", err)
		return nil, err
	}
	if packetData.PacketType != Message &&
		packetData.MessageType != Connect {
		return nil, fmt.Errorf("Unexpected namespace response: %v",
			packetData)
	}
	return &handshake, nil
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
	trans := &transport{
		conn:     conn,
		outgoing: make(map[string]*outgoing),
	}
	go trans.startReadLoop()
	trans.startHeartbeat(ping)

	// Upgrade the websocket connection.
	sendPacket(trans, Upgrade)

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
	return upgrade(endpoint, dialer, handshake.PingInterval)
}
