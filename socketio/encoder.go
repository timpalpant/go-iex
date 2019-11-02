package socketio

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	"github.com/golang/glog"
)

var msgTypeToNamespace = map[string]string{
	"IexDEEP": "/1.0/deep",
	"IexLast": "/1.0/last",
	"IexTOPS": "/1.0/tops",
}

// Signals a subscribe or unsubscribe event.
type SubOrUnsub string

const (
	Subscribe   SubOrUnsub = "subscribe"
	Unsubscribe SubOrUnsub = "unsubscribe"
)

// A subUnsubMsgFactory takes in a set of string symbols to subscribe or
// unsubscribe to and returns an IEXMsg suitable for passing to an Encoder. This
// is used by namespaces to encode subscriptions and unsubscriptions.
type subUnsubMsgFactory func(signal SubOrUnsub, symbols []string) *IEXMsg

// Returns a subscribe/unsubscribe struct for use by all endpoints except DEEP.
var simpleSubUnsubFactory = func(
	signal SubOrUnsub, symbols []string) *IEXMsg {
	return &IEXMsg{
		EventType: signal,
		Data:      strings.Join(symbols, ","),
	}
}

// Returns a subscribe/unsubscribe struct for use with the DEEP endpoint. Only
// a single symbol at a time can be used. If more than one symbol is passed in
// only the first one is used.
var deepSubUnsubFactory = func(
	signal SubOrUnsub, symbols []string) *IEXMsg {
	if len(symbols) > 1 {
		glog.Error("DEEP can only subscribe to one symbol at a time")
	}
	json, err := json.Marshal(struct {
		Symbols  []string `json:"symbols"`
		Channels []string `json:"channels"`
	}{
		Symbols:  symbols,
		Channels: []string{"deep"},
	})
	if err != nil {
		glog.Errorf("Could not encode DEEP %s", signal)
		return nil
	}
	return &IEXMsg{
		EventType: signal,
		Data:      string(json),
	}
}

type IEXMsg struct {
	// Contains a string representing subscribe or unsubscribe events.
	EventType SubOrUnsub
	// A string containing data to send. This is specific to a given
	// endpoint.
	Data string
}

// Encodes messages for use with IEX SocketIO. MessageType and PacketType are
// defined in decoder.go. If the MessageType or PacketType are less than 0,
// they are not set on the output.
type Encoder interface {
	// Encodes only a namespace and packet and message types.
	EncodePacket(p PacketType, m MessageType) (io.Reader, error)
	// Encodes a namespace, packet and message type and data.
	EncodeMessage(p PacketType, m MessageType, msg *IEXMsg) (
		io.Reader, error)
}

// Wraps a strArrayEncoder and returns its contents prepended by <length>:.
type httpEncoder struct {
	content *strArrayEncoder
}

func (enc *httpEncoder) EncodePacket(
	p PacketType, m MessageType) (io.Reader, error) {
	inner, err := enc.content.EncodePacket(p, m)
	if err != nil {
		return nil, err
	}
	val, err := ioutil.ReadAll(inner)
	if err != nil {
		if glog.V(3) {
			glog.Warningf("Failed to read inner encoding: %q", err)
		}
		return nil, err
	}
	if glog.V(3) {
		glog.Infof("Encoded packet: %s", val)
	}
	parts := []string{fmt.Sprintf("%d", len(val)), string(val)}
	return strings.NewReader(strings.Join(parts, ":")), nil
}

func (enc *httpEncoder) EncodeMessage(
	p PacketType, m MessageType, msg *IEXMsg) (io.Reader, error) {
	inner, err := enc.content.EncodeMessage(p, m, msg)
	if err != nil {
		return nil, err
	}
	val, err := ioutil.ReadAll(inner)
	if err != nil {
		if glog.V(3) {
			glog.Warningf("Failed to read inner encoding: %q", err)
		}
		return nil, err
	}
	if glog.V(3) {
		glog.Infof("Inner encoding: %s", val)
	}
	parts := []string{fmt.Sprintf("%d", len(val)), string(val)}
	return strings.NewReader(strings.Join(parts, ":")), nil
}

// The base encoder implementation that performs as described by the interface.
type strArrayEncoder struct {
	namespace string
}

// Used to indicate an encoding error.
type encodeError struct {
	message string
}

func (e *encodeError) Error() string {
	return e.message
}

func (enc *strArrayEncoder) EncodePacket(
	p PacketType, m MessageType) (io.Reader, error) {
	readers := make([]io.Reader, 0)
	if p >= 0 {
		readers = append(readers,
			strings.NewReader(fmt.Sprintf("%d", p)))
	}
	if m >= 0 {
		readers = append(readers,
			strings.NewReader(fmt.Sprintf("%d", m)))
	}
	if len(enc.namespace) > 0 {
		readers = append(readers,
			strings.NewReader(enc.namespace+","))
	}
	return io.MultiReader(readers...), nil
}

// Encodes a message, msg, of the given PacketType and MessageType. The
// resulting format is:
// <PacketType><MessageType><Namespace>,[msg.Event, msg.Data]
func (enc *strArrayEncoder) EncodeMessage(
	p PacketType, m MessageType, msg *IEXMsg) (io.Reader, error) {
	reader, err := enc.EncodePacket(p, m)
	if err != nil {
		return nil, err
	}
	readers := []io.Reader{reader}
	parts := []string{string(msg.EventType), msg.Data}
	if glog.V(3) {
		glog.Infof("Encoding parts: %v", parts)
	}
	encoding, err := json.Marshal(parts)
	if err != nil {
		glog.Errorf("Failed to encode data as JSON: %s", err)
		return nil, err
	}
	if len(parts) > 0 {
		readers = append(readers, bytes.NewBuffer(encoding))
	}
	return io.MultiReader(readers...), nil

}

// Returns an encoder for use with HTTP Post.
func NewHTTPEncoder(namespace string) Encoder {
	return &httpEncoder{&strArrayEncoder{namespace}}
}

// Returns an encoder for use with SocketIO.
func NewWSEncoder(namespace string) Encoder {
	return &strArrayEncoder{namespace}
}
