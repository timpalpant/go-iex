package socketio

// The iexMsgTypeNamespace is a generic class built using Genny.
// https://github.com/cheekybits/genny
// Run "go generate" to re-generate the specific namespace types.

import (
	"bytes"
	"errors"
	"io"
	"sync"

	"github.com/cheekybits/genny/generic"
	"github.com/golang/glog"
)

// The generic type representing the IEX message parsed by the namespace.
type iexMsgType generic.Type

// Contains a channel for receiving namespace specific messages. Only messages
// for the symbols subscribed to will be passed along.
//
// The close method *must* be called before garbage collection.
type iexMsgTypeConnection struct {
	// For closing.
	sync.Once
	// Guards the closed value.
	sync.RWMutex

	// The ID of this endpoint. Used for removing it from the namespace.
	id int
	// A channel for passing along namespace specific messages.
	C chan iexMsgType
	// Used to track which symbols this enpoint is subscribed to.
	subscriptions Subscriber
	// The factory function used to generate subscribe/unsubscribe messages.
	subUnsubMsgFactory subUnsubMsgFactory
	// Sends subscribe/unsubscribe structs to be encoded as JSON. When this
	// channel is closed, the connection is removed from the namespace.
	subUnsubClose chan<- *IEXMsg
	// True when this connection has been closed.
	closed bool
}

// Cleans up references to this connection in the Namespace. Messages will no
// longer be received and the Subscribe/Unsubscribe methods can no longer be
// called.
func (i *iexMsgTypeConnection) Close() {
	i.Do(func() {
		i.Lock()
		defer i.Unlock()
		i.closed = true
		close(i.subUnsubClose)
	})
}

// Subscribes to the given symbols. An error is returned if the connection is
// already closed.
func (i *iexMsgTypeConnection) Subscribe(symbols ...string) error {
	i.RLock()
	defer i.RUnlock()
	if i.closed {
		return errors.New(
			"Cannot call Subscribe on a closed connection")
	}
	go func() {
		for _, symbol := range symbols {
			i.subscriptions.Subscribe(symbol)
		}
		i.subUnsubClose <- i.subUnsubMsgFactory(
			Subscribe, symbols)
	}()
	return nil
}

// Unsubscribes to the given symbols. An error is returned if the connection is
// already closed.
func (i *iexMsgTypeConnection) Unsubscribe(symbols ...string) error {
	i.RLock()
	defer i.RUnlock()
	if i.closed {
		return errors.New(
			"Cannot call Unsubscribe on a closed connection")
	}
	go func() {
		for _, symbol := range symbols {
			i.subscriptions.Unsubscribe(symbol)
		}
		i.subUnsubClose <- i.subUnsubMsgFactory(
			Unsubscribe, symbols)
	}()
	return nil
}

// Returns true if this connection is subscribed to the given symbol.
func (i *iexMsgTypeConnection) Subscribed(symbol string) bool {
	return i.subscriptions.Subscribed(symbol)
}

// Receives messages for a given namespace and forwards them to endpoints.
type iexMsgTypeNamespace struct {
	// Used to guard access to the fanout channels.
	sync.RWMutex

	// The ID to use for the next endpoint created.
	nextId int
	// Active endpoints by ID.
	connections map[int]*iexMsgTypeConnection
	// Receives raw messages from the Transport. Only messages for the
	// current namespace will be received.
	msgChannel <-chan packetMetadata
	// For encoding outgoing messages in this namespace.
	encoder Encoder
	// Used for sending messages to IEX SocketIO.
	writer io.Writer
	// The factory function used to generate subscribe/unsubscribe messages.
	subUnsubMsgFactory subUnsubMsgFactory
	// A function to be called when the namespace has no more endpoints.
	closeFunc func()
}

func (i *iexMsgTypeNamespace) writeToReader(r io.Reader) error {
	var buffer bytes.Buffer
	if _, err := buffer.ReadFrom(r); err != nil {
		return err
	}
	if glog.V(3) {
		glog.Infof("Writing '%s' to reader", buffer.String())
	}
	if _, err := buffer.WriteTo(i.writer); err != nil {
		return err
	}
	return nil
}

// Sends a subscribe message and starts listening for incoming data. This is
// called when the namespace is created.
func (i *iexMsgTypeNamespace) connect() error {
	r, err := i.encoder.EncodePacket(Message, Connect)
	if err != nil {
		return err
	}
	if err := i.writeToReader(r); err != nil {
		return err
	}
	// Start listening for messages from the Transport layer.
	go func() {
		for msg := range i.msgChannel {
			i.fanout(msg)
		}
		// Close all outgoing connections.
		i.RLock()
		defer i.RUnlock()
		for _, connection := range i.connections {
			close(connection.C)
		}
	}()
	return nil
}

// Given a string representing a JSON IEX message type, parse out the symbol and
// the message and pass the message to each connection subscribed to the symbol.
// Use a go routine to prevent from blocking.
func (i *iexMsgTypeNamespace) fanout(pkt packetMetadata) {
	go func() {
		// This "symbol only" struct is necessary because this class
		// is a genny generic. Therefore, even though all IEX messages
		// have a "symbol" field, iexMsgType.symbol is not type safe.
		var symbol struct {
			Symbol string
		}
		if err := ParseToJSON(pkt.Data, &symbol); err != nil {
			glog.Errorf("No symbol found for iexMsgType: %s - %v",
				err, pkt)
		}
		// Now that the symbol has been extraced, the specific message
		// can be extracted from the data.
		var decoded iexMsgType
		if err := ParseToJSON(pkt.Data, &decoded); err != nil {
			glog.Errorf("Could not decode iexMsgType: %s - %v",
				err, pkt)
		}
		i.RLock()
		defer i.RUnlock()
		for _, connection := range i.connections {
			if connection.Subscribed(symbol.Symbol) {
				connection.C <- decoded
			}
		}
	}()
}

// Returns a connection that will receive messages for the passed in symbols.
// If no symbols are passed in, they can be added/removed later.
func (i *iexMsgTypeNamespace) GetConnection(
	symbols ...string) *iexMsgTypeConnection {
	i.Lock()
	defer i.Unlock()
	i.nextId++
	subUnsubClose := make(chan *IEXMsg, 0)
	connection := &iexMsgTypeConnection{
		id:                 i.nextId,
		C:                  make(chan iexMsgType, 1),
		subscriptions:      NewPresenceSubscriber(),
		subUnsubMsgFactory: i.subUnsubMsgFactory,
		subUnsubClose:      subUnsubClose,
		closed:             false,
	}
	// Start listening for close, subscribe and unsubscribe messages on the
	// new connection.
	go func(id int) {
		for subUnsubMsg := range subUnsubClose {
			r, err := i.encoder.EncodeMsg(
				Message, Event, subUnsubMsg)
			if err != nil {
				glog.Errorf("Error encoding %+v: %s",
					subUnsubMsg, err)
				continue
			}
			if err := i.writeToReader(r); err != nil {
				glog.Errorf("Error encoding %+v: %s",
					subUnsubMsg, err)
				continue
			}

		}
		i.Lock()
		defer i.Unlock()
		delete(i.connections, id)
		if len(i.connections) == 0 {
			i.closeFunc()
		}

	}(i.nextId)
	i.connections[i.nextId] = connection
	if len(symbols) > 0 {
		connection.Subscribe(symbols...)
	}
	return connection
}

func newiexMsgTypeNamespace(
	ch <-chan packetMetadata, encoder Encoder,
	writer io.Writer, subUnsubMsgFactory subUnsubMsgFactory,
	closeFunc func()) *iexMsgTypeNamespace {
	newNs := &iexMsgTypeNamespace{
		nextId:             0,
		connections:        make(map[int]*iexMsgTypeConnection),
		msgChannel:         ch,
		encoder:            encoder,
		writer:             writer,
		subUnsubMsgFactory: subUnsubMsgFactory,
		closeFunc:          closeFunc,
	}
	newNs.connect()
	return newNs
}

//go:generate genny -in=$GOFILE -out=gen-$GOFILE gen "iexMsgType=iex.TOPS,iex.Last,iex.DEEP"
