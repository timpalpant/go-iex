package socketio_test

import (
	"net/url"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	. "github.com/timpalpant/go-iex/socketio"
)

func TestIEXEndpoint(t *testing.T) {
	Convey("The Endpoint", t, func() {
		Convey("should have the correct base http URL", func() {
			endpoint := NewIEXEndpoint(func() string {
				return "123"
			})
			to, err := url.Parse(endpoint.GetHTTPUrl())
			So(err, ShouldBeNil)
			So(to.Scheme, ShouldEqual, "https")
			So(to.Host, ShouldEqual, "ws-api.iextrading.com")
			So(to.Path, ShouldEqual, "/socketio/")
			values := to.Query()
			So(values.Get("EIO"), ShouldEqual, "3")
			So(values.Get("transport"), ShouldEqual, "polling")
			So(values.Get("b64"), ShouldEqual, "1")
		})
		Convey("should have the correct base Websocket URL", func() {
			endpoint := NewIEXEndpoint(func() string {
				return "123"
			})
			to, err := url.Parse(endpoint.GetWSUrl())
			So(err, ShouldBeNil)
			So(to.Scheme, ShouldEqual, "wss")
			So(to.Host, ShouldEqual, "ws-api.iextrading.com")
			So(to.Path, ShouldEqual, "/socketio/")
			values := to.Query()
			So(values.Get("EIO"), ShouldEqual, "3")
			So(values.Get("transport"), ShouldEqual, "websocket")
		})
		Convey("should set the SID on the HTTP URL", func() {
			sid := "4567"
			endpoint := NewIEXEndpoint(func() string {
				return "123"
			})
			endpoint.SetSid(sid)
			to, err := url.Parse(endpoint.GetHTTPUrl())
			So(err, ShouldBeNil)
			values := to.Query()
			So(values.Get("sid"), ShouldEqual, sid)
		})
		Convey("should set the SID on the Websocket URL", func() {
			sid := "4567"
			endpoint := NewIEXEndpoint(func() string {
				return "123"
			})
			endpoint.SetSid(sid)
			to, err := url.Parse(endpoint.GetWSUrl())
			So(err, ShouldBeNil)
			values := to.Query()
			So(values.Get("sid"), ShouldEqual, sid)
		})
		Convey("should change 't' for HTTP URLs", func() {
			timestamps := []string{"123", "456"}
			index := 0
			endpoint := NewIEXEndpoint(func() string {
				timestamp := timestamps[index]
				index++
				return timestamp
			})
			to, err := url.Parse(endpoint.GetHTTPUrl())
			So(err, ShouldBeNil)
			values := to.Query()
			first := values.Get("t")

			to, err = url.Parse(endpoint.GetHTTPUrl())
			So(err, ShouldBeNil)
			values = to.Query()
			second := values.Get("t")

			So(first, ShouldNotEqual, second)
		})
		Convey("should change 't' for WS URLs", func() {
			timestamps := []string{"123", "456"}
			index := 0
			endpoint := NewIEXEndpoint(func() string {
				timestamp := timestamps[index]
				index++
				return timestamp
			})
			to, err := url.Parse(endpoint.GetWSUrl())
			So(err, ShouldBeNil)
			values := to.Query()
			first := values.Get("t")

			to, err = url.Parse(endpoint.GetWSUrl())
			So(err, ShouldBeNil)
			values = to.Query()
			second := values.Get("t")

			So(first, ShouldNotEqual, second)
		})
	})
}
