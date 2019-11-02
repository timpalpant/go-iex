package socketio

// The iexMsgTypeNamespace is a generic class built using Genny.
// https://github.com/cheekybits/genny
// Run "go generate" to re-generate the specific namespace types.

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/cheekybits/genny/generic"
	"github.com/golang/glog"
)

// The generic type representing the IEX message parsed by the namespace.
type IEXMsgType generic.Type

// Contains callbacks and the symbols they correspond to.
type subIEXMsgType struct {
	Callback func(IEXMsgType)
	Symbols  map[string]struct{}
}

// Receives messages for a given namespace and forwards them to endpoints.
type IEXMsgTypeNamespace struct {
	// Used to guard access to the fanout channels.
	sync.RWMutex

	// A set of symbols that this namespace is currently subscribed to.
	// This spans across subcriptions so that unsubscribing from a symbol
	// only occurs if there are no subscriptions listening for that symbol.
	symbols Subscriber
	// The ID to use for the next connection created.
	nextId int
	// Active subscriptions by ID.
	subscriptions map[int]*subIEXMsgType
	// For encoding outgoing messages in this namespace.
	encoder Encoder
	// Used for sending messages to the Transport.
	writer io.Writer
	// The factory function used to generate subscribe/unsubscribe messages.
	// Subscribe and unsubscribe messages can differe by IEX namespace.
	subUnsubMsgFactory subUnsubMsgFactory
	// A function to be called when the namespace has no more endpoints.
	closeFunc func(string)
}

// Sends a subscribe message. This is performed when the number of subscriptions
// goes from 0 to 1.
func (i *IEXMsgTypeNamespace) sendPacket(msgType MessageType) error {
	r, err := i.encoder.EncodePacket(Message, msgType)
	if err != nil {
		return err
	}
	buffer := &bytes.Buffer{}
	_, err = buffer.ReadFrom(r)
	_, err = buffer.WriteTo(i.writer)
	return err
}

// Encodes and sends a subscribe or unsubscribe message on the transport layer.
func (i *IEXMsgTypeNamespace) sendSubUnsub(subUnsubMsg *IEXMsg) error {
	r, err := i.encoder.EncodeMessage(Message, Event, subUnsubMsg)
	if err != nil {
		return fmt.Errorf("Error encoding %+v: %s", subUnsubMsg, err)
	}
	buffer := &bytes.Buffer{}
	_, err = buffer.ReadFrom(r)
	_, err = buffer.WriteTo(i.writer)
	return err
}

// Given a string representing a JSON IEX message type, parse out the symbol and
// message and pass the message to each connection subscribed to the symbol.
func (i *IEXMsgTypeNamespace) fanout(pkt PacketData) {
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
	var decoded IEXMsgType
	if err := ParseToJSON(pkt.Data, &decoded); err != nil {
		glog.Errorf("Could not decode iexMsgType: %s - %v",
			err, pkt)
	}
	if glog.V(5) {
		glog.Infof("Extracted symbol: %v", symbol)
		glog.Infof("Extracted message: %v", decoded)
	}
	i.RLock()
	defer i.RUnlock()
	for _, sub := range i.subscriptions {
		if glog.V(5) {
			glog.Infof("Checking for subscription to %s",
				symbol.Symbol)
		}
		if _, ok := sub.Symbols[symbol.Symbol]; ok {
			if glog.V(5) {
				glog.Infof("Calling subscription to %s",
					symbol.Symbol)
			}
			sub.Callback(decoded)
		}
	}
}

// Returns a method that is passed to new Connections, to be called when the
// connection is being closed.
func (i *IEXMsgTypeNamespace) getCloseSubscriptionFunc(id int) func() {
	return func() {
		i.Lock()
		unsub := make([]string, 0)
		sub := i.subscriptions[id]
		// Unsubscribe from the subscription symbols. For any that are
		// no longer being listened to by any subscription, send an
		// unsubscribe event to IEX. If there are no more subscriptions
		// in the namespace, disconnect from the namespace.
		for key, _ := range sub.Symbols {
			i.symbols.Unsubscribe(key)
			if !i.symbols.Subscribed(key) {
				unsub = append(unsub, key)
			}
		}
		delete(i.subscriptions, id)
		i.Unlock()
		if len(unsub) > 0 {
			err := i.sendSubUnsub(i.subUnsubMsgFactory(
				Unsubscribe, unsub))
			if err != nil {
				glog.Errorf("Error unsubscrubing from %v: %s",
					unsub, err)
			}
		}
		if len(i.subscriptions) == 0 {
			i.closeFunc(msgTypeToNamespace["IEXMsgType"])
		}
	}
}

// Receive messages for the passed in symbols using the passed in callback.
// Returns a close function that should be called when the client does not wish
// to receive any further messages. If symbols is empty, an error is returned.
func (i *IEXMsgTypeNamespace) SubscribeTo(
	msgReceived func(msg IEXMsgType), symbols ...string) (func(), error) {
	if len(symbols) == 0 {
		return nil, errors.New(
			"Cannot call SubscribeTo with no symbols")
	}
	i.Lock()
	defer i.Unlock()
	// Connect to the namespace when adding the first subscription.
	if len(i.subscriptions) == 0 {
		i.sendPacket(Connect)
	}
	i.nextId++
	newSub := &subIEXMsgType{
		Callback: msgReceived,
		Symbols:  make(map[string]struct{}),
	}
	if len(symbols) > 0 {
		for _, symbol := range symbols {
			symbol = strings.ToUpper(symbol)
			newSub.Symbols[symbol] = struct{}{}
			i.symbols.Subscribe(symbol)
		}
		err := i.sendSubUnsub(i.subUnsubMsgFactory(Subscribe, symbols))
		if err != nil {
			return nil, err
		}
	}
	i.subscriptions[i.nextId] = newSub
	return i.getCloseSubscriptionFunc(i.nextId), nil
}

// Create a new namespace for a specific IEX endpoint. Because the IEX
// namespaces use different message types for representing the received data,
// these classes are represented as generics using Genny.
func NewIEXMsgTypeNamespace(
	transport Transport, subUnsubMsgFactory subUnsubMsgFactory,
	closeFunc func(string)) *IEXMsgTypeNamespace {
	namespace := msgTypeToNamespace["IEXMsgType"]
	encoder := NewWSEncoder(namespace)
	newNs := &IEXMsgTypeNamespace{
		symbols:            NewCountingSubscriber(),
		nextId:             0,
		subscriptions:      make(map[int]*subIEXMsgType),
		encoder:            encoder,
		writer:             transport,
		subUnsubMsgFactory: subUnsubMsgFactory,
		closeFunc:          closeFunc,
	}
	transport.AddPacketCallback(namespace, newNs.fanout)
	return newNs
}

//go:generate genny -in=$GOFILE -out=gen-$GOFILE gen "IEXMsgType=iex.TOPS,iex.Last,iex.DEEP"
