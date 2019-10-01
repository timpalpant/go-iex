package socketio

import "net/url"

// HTTP endpoint.
var httpEndpoint, _ = url.Parse("https://ws-api.iextrading.com/socketio/")

// Wbsocket endpoint.
var wsEndpoint, _ = url.Parse("wss://ws-api.iextrading.com/socketio/")

// An interface for hiding iexEndpoint.
type Endpoint interface {
	SetSid(sid string)
	GetHTTPUrl() string
	GetWSUrl() string
}

// Provides methods for manipulating the IEX websocket URL.
type iexEndpoint struct {
	// URL for making HTTP requests.
	httpUrl *url.URL

	// URL for making websocket requests.
	wsUrl *url.URL

	// Method for generating unique timestamps.
	idg func() string
}

func (e *iexEndpoint) SetSid(sid string) {
	httpValues := e.httpUrl.Query()
	httpValues.Set("sid", sid)
	e.httpUrl.RawQuery = httpValues.Encode()

	wsValues := e.wsUrl.Query()
	wsValues.Set("sid", sid)
	e.wsUrl.RawQuery = wsValues.Encode()
}

func (e *iexEndpoint) GetHTTPUrl() string {
	httpValues := e.httpUrl.Query()
	httpValues.Set("t", e.idg())
	e.httpUrl.RawQuery = httpValues.Encode()
	return e.httpUrl.String()
}

func (e *iexEndpoint) GetWSUrl() string {
	wsValues := e.wsUrl.Query()
	wsValues.Set("t", e.idg())
	e.wsUrl.RawQuery = wsValues.Encode()
	return e.wsUrl.String()
}

func (e *iexEndpoint) Initialize() {
	// Initialize the HTTP enpoint query params.
	httpValues := e.httpUrl.Query()
	httpValues.Set("EIO", "3")
	httpValues.Set("transport", "polling")
	e.httpUrl.RawQuery = httpValues.Encode()

	// Initialize the Websocket enpoint query params.
	wsValues := e.wsUrl.Query()
	wsValues.Set("EIO", "3")
	wsValues.Set("transport", "websocket")
	e.wsUrl.RawQuery = wsValues.Encode()
}

func NewIEXEndpoint(idg func() string) Endpoint {
	endpoint := &iexEndpoint{httpEndpoint, wsEndpoint, idg}
	endpoint.Initialize()
	return endpoint
}
