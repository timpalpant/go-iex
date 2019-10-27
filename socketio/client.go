package socketio

import (
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"sync"

	"github.com/golang/glog"
	"github.com/gorilla/websocket"
)

const (
	deep string = "/1.0/deep"
	last string = "/1.0/last"
	tops string = "/1.0/tops"
)

// Connects to IEX SocketIO endpoints and routes received messages back to the
// correct handlers.
type Client struct {
	// Allows reference counting of open namespaces.
	CountingSubscriber
	// Protects access to namespaces.
	sync.Mutex

	// The Transport object used to send and receive SocketIO messages.
	transport Transport
	// Points to a DEEP namespace.
	deepNamespace *IexDEEPNamespace
	// Points to a Last namespace.
	lastNamespace *IexLastNamespace
	// Points to a TOPS namespace.
	topsNamespace *IexTOPSNamespace
}

func (c *Client) getCloseNamespace(ns string) func() {
	return func() {
		c.Lock()
		defer c.Unlock()
		c.Unsubscribe(ns)
		if !c.Subscribed(ns) {
			enc := NewWSEncoder(ns)
			r, err := enc.EncodePacket(Message, Disconnect)
			if err != nil {
				glog.Errorf(
					"Error disconnecting from %s: %s",
					ns, err)
			}
			msg, err := ioutil.ReadAll(r)
			if err != nil {
				glog.Errorf(
					"Error disconnecting from %s: %s",
					ns, err)
			}
			if _, err = c.transport.Write(msg); err != nil {
				glog.Errorf(
					"Error disconnecting from %s: %s",
					ns, err)
			}
			switch ns {
			case deep:
				c.deepNamespace = nil
			case last:
				c.lastNamespace = nil
			case tops:
				c.topsNamespace = nil
			}
		}
	}
}

func (c *Client) GetDEEPNamespace() *IexDEEPNamespace {
	if c.deepNamespace != nil {
		return c.deepNamespace
	}
	readChan, err := c.transport.GetReadChannel()
	if err != nil {
		glog.Fatalf("Failed to get DEEP namespace: %s", err)
	}
	c.deepNamespace = NewIexDEEPNamespace(
		readChan, NewWSEncoder(deep),
		c.transport, deepSubUnsubFactory, c.getCloseNamespace(deep))
	return c.deepNamespace
}

func (c *Client) GetLastNamespace() *IexLastNamespace {
	if c.lastNamespace != nil {
		return c.lastNamespace
	}
	readChan, err := c.transport.GetReadChannel()
	if err != nil {
		glog.Fatalf("Failed to get DEEP namespace: %s", err)
	}
	c.lastNamespace = NewIexLastNamespace(
		readChan, NewWSEncoder(last),
		c.transport, simpleSubUnsubFactory, c.getCloseNamespace(last))
	return c.lastNamespace
}

func (c *Client) GetTOPSNamespace() *IexTOPSNamespace {
	if c.topsNamespace != nil {
		return c.topsNamespace
	}
	readChan, err := c.transport.GetReadChannel()
	if err != nil {
		glog.Fatalf("Failed to get DEEP namespace: %s", err)
	}
	c.topsNamespace = NewIexTOPSNamespace(
		readChan, NewWSEncoder(tops),
		c.transport, deepSubUnsubFactory, c.getCloseNamespace(tops))
	return c.topsNamespace
}

type defaultDialerWrapper struct {
	dialer *websocket.Dialer
}

func (d *defaultDialerWrapper) Dial(uri string, hdr http.Header) (
	WSConn, *http.Response, error) {
	return d.dialer.Dial(uri, hdr)
}

// Returns a SocketIO client that will use the passed in transport for
// communication. If it is nil, a default Transport will be created using an
// http.Client and websocket.DefaultDialer. The ability to inject a Tranport
// is mainly meant for testing.
func NewClientWithTransport(conn Transport) *Client {
	toReturn := &Client{
		transport: conn,
	}
	if conn == nil {
		wrapper := &defaultDialerWrapper{websocket.DefaultDialer}
		jar, err := cookiejar.New(nil)
		if err != nil {
			glog.Fatalf("Error creating cookie jar: %s", err)
		}
		transport, err := NewTransport(&http.Client{Jar: jar}, wrapper)
		if err != nil {
			glog.Fatalf(
				"Failed to create default transport: %s",
				err)
		}
		toReturn.transport = transport
	}
	return toReturn

}
func NewClient() *Client {
	return NewClientWithTransport(nil)
}
