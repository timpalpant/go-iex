package socketio

import (
	"io/ioutil"
	"net/url"

	"github.com/golang/glog"
	"github.com/timpalpant/go-iex"
)

// The base socketio endpoint.
var baseEndpointUrl, _ = url.Parse("https://ws-api.iextrading.com")

// A SocketIO client capable of negotiating the SocketIO handshake and
// connecting with IEX streaming endpoints.
type SocketioClient struct {
	httpClient  iex.HTTPClient
	idGenerator func() string
}

// NewClient creates a new SocketIO client.
func NewSocketioClient(
	client iex.HTTPClient, idg func() string) *SocketioClient {
	return &SocketioClient{client, idg}
}

func (s *SocketioClient) makeRequest(to *url.URL, ch chan<- string) {
	values := to.Query()
	values.Set("t", s.idGenerator())
	to.RawQuery = values.Encode()
	glog.Info("Requesting: %v", to.String())
	resp, _ := s.httpClient.Get(to.String())

	body, _ := ioutil.ReadAll(resp.Body)
	ch <- string(body)
}

// Negotiates the first polling request.
func (s *SocketioClient) negotiate() {
	to, _ := url.Parse(baseEndpointUrl.String())
	to.Path = "/socket.io"
	values := to.Query()
	values.Set("EIO", "3")
	values.Set("transport", "polling")
	to.RawQuery = values.Encode()
	ch := make(chan string)
	go s.makeRequest(to, ch)
	glog.Info(<-ch)
}

func (s *SocketioClient) OpenTops() {
	s.negotiate()
}
