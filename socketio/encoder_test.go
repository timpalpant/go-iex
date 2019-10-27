package socketio_test

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	. "github.com/timpalpant/go-iex/socketio"
)

func TestWebsocketEncoding(t *testing.T) {
	Convey("Websocket encoding should", t, func() {
		Convey("correctly encode a nil type", func() {
			encoder := NewWSEncoder("/")
			encoded, err := encoder.EncodePacket(-1, -1)
			So(err, ShouldBeNil)
			val, err := ioutil.ReadAll(encoded)
			So(err, ShouldBeNil)
			So(string(val), ShouldEqual, `/,`)
		})
		Convey("correctly encode an empty type", func() {
			encoder := NewWSEncoder("/")
			encoded, err := encoder.EncodePacket(-1, -1)
			So(err, ShouldBeNil)
			val, err := ioutil.ReadAll(encoded)
			So(err, ShouldBeNil)
			So(string(val), ShouldEqual, `/,`)
		})
		Convey("correctly send an upgrade request", func() {
			encoder := NewWSEncoder("")
			encoded, err := encoder.EncodePacket(5, -1)
			So(err, ShouldBeNil)
			val, err := ioutil.ReadAll(encoded)
			So(err, ShouldBeNil)
			So(string(val), ShouldEqual, `5`)
		})
		Convey("correctly encode a simple type", func() {
			encoder := NewWSEncoder("")
			encoding, err := json.Marshal(struct {
				Name string
				Ints []int
			}{
				Name: "foo",
				Ints: []int{1, 2, 3},
			})
			So(err, ShouldBeNil)
			iexMsg := &IEXMsg{
				EventType: Subscribe,
				Data:      string(encoding),
			}
			encoded, err := encoder.EncodeMsg(-1, -1, iexMsg)
			So(err, ShouldBeNil)
			val, err := ioutil.ReadAll(encoded)
			So(err, ShouldBeNil)
			So(string(val), ShouldEqual,
				`["subscribe","{\"Name\":\"foo\",\"Ints\":[1,2,3]}"]`)
		})
		Convey("correctly encode a namespace", func() {
			encoder := NewWSEncoder("/")
			iexMsg := &IEXMsg{
				EventType: Subscribe,
				Data:      "foo",
			}
			encoded, err := encoder.EncodeMsg(-1, -1, iexMsg)
			So(err, ShouldBeNil)
			val, err := ioutil.ReadAll(encoded)
			So(err, ShouldBeNil)
			So(string(val), ShouldEqual, `/,["subscribe","foo"]`)
		})
		Convey("correctly encode a longer namespace", func() {
			encoder := NewWSEncoder("/1.0/tops")
			iexMsg := &IEXMsg{
				EventType: Subscribe,
				Data:      "foo",
			}
			encoded, err := encoder.EncodeMsg(-1, -1, iexMsg)
			So(err, ShouldBeNil)
			val, err := ioutil.ReadAll(encoded)
			So(err, ShouldBeNil)
			So(string(val), ShouldEqual,
				`/1.0/tops,["subscribe","foo"]`)
		})
		Convey("correctly encode the packet type", func() {
			encoder := NewWSEncoder("/1.0/tops")
			iexMsg := &IEXMsg{
				EventType: Subscribe,
				Data:      "foo",
			}
			encoded, err := encoder.EncodeMsg(4, -1, iexMsg)
			So(err, ShouldBeNil)
			val, err := ioutil.ReadAll(encoded)
			So(err, ShouldBeNil)
			So(string(val), ShouldEqual,
				`4/1.0/tops,["subscribe","foo"]`)
		})
		Convey("correctly encode the packet and message type", func() {
			encoder := NewWSEncoder("/1.0/tops")
			iexMsg := &IEXMsg{
				EventType: Subscribe,
				Data:      "foo",
			}
			encoded, err := encoder.EncodeMsg(4, 2, iexMsg)
			So(err, ShouldBeNil)
			val, err := ioutil.ReadAll(encoded)
			So(err, ShouldBeNil)
			So(string(val), ShouldEqual,
				`42/1.0/tops,["subscribe","foo"]`)
		})
	})
}

func TestHTTPEncoding(t *testing.T) {
	Convey("HTTP encoding should", t, func() {
		Convey("correctly encode a nil type", func() {
			encoder := NewHTTPEncoder("/")
			encoded, err := encoder.EncodePacket(-1, -1)
			So(err, ShouldBeNil)
			val, err := ioutil.ReadAll(encoded)
			So(err, ShouldBeNil)
			So(string(val), ShouldEqual, `2:/,`)
		})
		Convey("correctly encode an empty type", func() {
			encoder := NewHTTPEncoder("/")
			encoded, err := encoder.EncodePacket(4, 0)
			So(err, ShouldBeNil)
			val, err := ioutil.ReadAll(encoded)
			So(err, ShouldBeNil)
			So(string(val), ShouldEqual, `4:40/,`)
		})
		Convey("correctly encode a simple type", func() {
			encoder := NewHTTPEncoder("")
			encoding, err := json.Marshal(struct {
				Name string
				Ints []int
			}{
				Name: "foo",
				Ints: []int{1, 2, 3},
			})
			So(err, ShouldBeNil)
			iexMsg := &IEXMsg{
				EventType: Subscribe,
				Data:      string(encoding),
			}
			encoded, err := encoder.EncodeMsg(-1, -1, iexMsg)
			So(err, ShouldBeNil)
			val, err := ioutil.ReadAll(encoded)
			So(err, ShouldBeNil)
			So(string(val), ShouldEqual,
				`51:["subscribe","{\"Name\":\"foo\",\"Ints\":[1,2,3]}"]`)
		})
		Convey("correctly encode a namespace", func() {
			encoder := NewHTTPEncoder("/")
			iexMsg := &IEXMsg{
				EventType: Subscribe,
				Data:      "foo",
			}
			encoded, err := encoder.EncodeMsg(-1, -1, iexMsg)
			So(err, ShouldBeNil)
			val, err := ioutil.ReadAll(encoded)
			So(err, ShouldBeNil)
			So(string(val), ShouldEqual, `21:/,["subscribe","foo"]`)
		})
		Convey("correctly encode a longer namespace", func() {
			encoder := NewHTTPEncoder("/1.0/tops")
			iexMsg := &IEXMsg{
				EventType: Subscribe,
				Data:      "foo",
			}
			encoded, err := encoder.EncodeMsg(-1, -1, iexMsg)
			So(err, ShouldBeNil)
			val, err := ioutil.ReadAll(encoded)
			So(err, ShouldBeNil)
			So(string(val), ShouldEqual,
				`29:/1.0/tops,["subscribe","foo"]`)
		})
		Convey("correctly encode the packet type", func() {
			encoder := NewHTTPEncoder("/1.0/tops")
			iexMsg := &IEXMsg{
				EventType: Subscribe,
				Data:      "foo",
			}
			encoded, err := encoder.EncodeMsg(4, -1, iexMsg)
			So(err, ShouldBeNil)
			val, err := ioutil.ReadAll(encoded)
			So(err, ShouldBeNil)
			So(string(val), ShouldEqual,
				`30:4/1.0/tops,["subscribe","foo"]`)
		})
		Convey("correctly encode the packet and message type", func() {
			encoder := NewHTTPEncoder("/1.0/tops")
			iexMsg := &IEXMsg{
				EventType: Subscribe,
				Data:      "foo",
			}
			encoded, err := encoder.EncodeMsg(4, 2, iexMsg)
			So(err, ShouldBeNil)
			val, err := ioutil.ReadAll(encoded)
			So(err, ShouldBeNil)
			So(string(val), ShouldEqual,
				`31:42/1.0/tops,["subscribe","foo"]`)
		})
	})
}
