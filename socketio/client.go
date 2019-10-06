package socketio

// Connects to IEX SocketIO endpoints and routes received messages back to the
// correct handlers.
type SocketIOClient struct {
	// The Transport object used to send and receive SocketIO messages.
	Conn Transport

	// A mapping from namespaces to listening channels.
	outgoing map[string]chan string
}
