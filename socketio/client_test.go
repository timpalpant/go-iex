package socketio_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	. "github.com/timpalpant/go-iex/socketio"
)

type mockHTTPClient struct {
	body     string
	headers  map[string]string
	code     int
	err      error
	requests []*url.URL
}

func (c *mockHTTPClient) Get(to string) (*http.Response, error) {
	parsed, _ := url.Parse(to)
	c.requests = append(c.requests, parsed)
	w := httptest.NewRecorder()
	w.WriteString(c.body)

	for key, value := range c.headers {
		w.Header().Add(key, value)
	}

	w.WriteHeader(c.code)

	resp := w.Result()
	return resp, c.err
}

func TestClient(t *testing.T) {
	Convey("The SocketIO client", t, func() {
		Convey("should make the initial polling request", func() {
			mock := &mockHTTPClient{}
			c := NewSocketioClient(mock, func() string {
				return "1234"
			})
			c.OpenTops()
			expected, _ := url.Parse("https://ws-api.iextrading.com")
			expected.Path = "socket.io"
			values := expected.Query()
			values.Set("EIO", "3")
			values.Set("transport", "polling")
			values.Set("t", "1234")
			expected.RawQuery = values.Encode()
			So(mock.requests[0].String(), ShouldEqual,
				expected.String())
		})
	})
}
