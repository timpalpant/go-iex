package socketio

// Utilties for encoding SocketIO messages as described here:
// https://socket.io/docs/internals/

import (
	"io"
	"strings"
)

// Provides methods for getting the message type and the message data.
type message interface {
	getStringData() io.Reader
	getMessageType() string
}

// Returns the passed in string with quotes around it.
func getQuoted(str string) string {
	return "\"" + str + "\""
}

// Used to subscribe to data for a given list of traded symbols.
type SubscribeMessage struct {
	symbols []string
}

func NewSubscribeMessage(symbols []string) *SubscribeMessage {
	return &SubscribeMessage{symbols}
}

func (s *SubscribeMessage) getMessageType() string {
	return "2"
}

func (s *SubscribeMessage) getStringData() io.Reader {
	return strings.NewReader(
		"[\"subscribe\",\"" + strings.Join(s.symbols, ",") + "\"]")
}

// Encodes messages to be sent to IEX via SocketIO.
type SocketioEncoder struct {
	namespace string
}

func NewSocketioEncoder(namespace string) *SocketioEncoder {
	encoder := &SocketioEncoder{}
	encoder.namespace = strings.TrimRightFunc(
		namespace, func(char rune) bool {
			if string(char) == "/" {
				return true
			}
			return false
		})
	return encoder
}

func (s *SocketioEncoder) Encode(msg message) io.Reader {
	prefix := msg.getMessageType()
	if s.namespace != "/" {
		prefix += s.namespace
	}
	prefix += ","
	return io.MultiReader(
		strings.NewReader(prefix), msg.getStringData())
}
