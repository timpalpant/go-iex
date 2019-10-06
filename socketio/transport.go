package socketio

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
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

type namespaceResponse struct {
	MessageType MessageType
	PacketType  PacketType
	Nsp         string
}

// Fulfilled by http.Client#Do.
type DoClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// An internal interface that is fulfilled by websocket.Conn and allows
// for injecting a test connection.
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

// Negotiates a SocketIO connection with IEX and returns a WSConn for
// communicating with IEX.
type Transport interface {
	ReadMessage() (int, []byte, error)
	WriteMessage(data io.Reader) error
	Close() error
}

type transport struct {
	conn          WSConn
	quitHeartbeat chan int
}

func (t *transport) ReadMessage() (int, []byte, error) {
	return t.conn.ReadMessage()
}

func (t *transport) WriteMessage(data io.Reader) error {
	toWrite, err := ioutil.ReadAll(data)
	if err != nil {
		return err
	}
	return t.conn.WriteMessage(websocket.TextMessage, toWrite)
}

func (t *transport) Close() error {
	t.quitHeartbeat <- stopBeating
	return t.conn.Close()
}

// Indicates an error during initialization of the Transport layer.
type transportConnectError struct {
	message string
}

func (t *transportConnectError) Error() string {
	return t.message
}

// Performs an HTTP request and returns the response body. If there is an error
// the io.ReaderCloser will be nil.
func makeHTTPRequest(client DoClient, method string,
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
		return nil, &transportConnectError{fmt.Sprintf(
			"No response body from %s", to)}
	}
	return resp.Body, nil
}

// Performs the initial GET connection to the SocketIO endpoint. If it it
// successful, it will set the session id (sid) parameter on the endpoint.
func connect(endpoint Endpoint, client DoClient) (*handshakeResponse, error) {
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
		return nil, &transportConnectError{
			"Websocket upgrade not found"}
	}
	endpoint.SetSid(handshake.Sid)
	return &handshake, nil
}

// Joins the default namespace. Returns an error if there was an unexpected
// server response.
func joinDefaultNsp(endpoint Endpoint, client DoClient) error {
	encoder := NewHTTPEncoder("/")
	reader, err := encoder.Encode(4, 0, nil)
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
	var nsp namespaceResponse
	err = HTTPToJSON(resp, []interface{}{&nsp})
	if err != nil {
		glog.Errorf("Error parsing namespace response: %s", err)
		return err
	}
	if nsp.MessageType != 4 || nsp.PacketType != 0 {
		glog.Errorf("Unexpected namespace response: %+v", nsp)
		return &transportConnectError{fmt.Sprintf(
			"Unexpected namespace response: %+v", nsp)}
	}
	return nil
}

func startHeartbeat(transport WSConn, ping int) chan int {
	encoder := NewWSEncoder("")
	send := func(packetType PacketType) {
		reader, err := encoder.Encode(packetType, -1, nil)
		if err != nil {
			glog.Warningf(
				"Could not encode probe message: %s", err)
		}
		data, err := ioutil.ReadAll(reader)
		if err != nil {
			glog.Warningf(
				"Could not read encoded message: %s", err)
		}
		err = transport.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			glog.Warningf(
				"Error writing probe message: %s", err)
		}
	}
	quit := make(chan int)
	duration, err := time.ParseDuration(strconv.Itoa(ping) + "ms")
	if err != nil {
		glog.Fatalf("Could not start heartbeat: %s", err)
	}
	heartbeat := time.NewTicker(duration)
	go func() {
		for {
			select {
			case <-quit:
				send(Close)
				heartbeat.Stop()
			case t := <-heartbeat.C:
				if glog.V(3) {
					glog.Infof("Heartbeating at %v", t)
				}
				send(Ping)
			}
		}
	}()
	return quit
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
	reader, err := encoder.Encode(5, -1, nil)
	if err != nil {
		glog.Errorf("Error upgrading connection: %s", err)
		return nil, err
	}
	quit := startHeartbeat(conn, ping)
	trans := &transport{conn, quit}
	err = trans.WriteMessage(reader)
	if err != nil {
		glog.Errorf("Error upgrading connection: %s", err)
		return nil, err
	}

	return trans, nil
}

// Returns a new Transport object backed by an open Websocket connection
// or an error if one occurs.
func NewTransport(client DoClient, dialer WSDialer) (Transport, error) {
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
