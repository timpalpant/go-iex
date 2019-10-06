package socketio_test

import (
	"flag"
	"fmt"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	. "github.com/timpalpant/go-iex/socketio"
)

type fakeData struct {
	Foo string
	Bar []int
}

type fakeDataWithTypes struct {
	Foo         string
	Bar         []int
	MessageType int
	PacketType  int
}

func init() {
	if flag.Lookup("alsologtostderr").Value == nil {
		flag.Set("alsologtostderr", fmt.Sprintf("%t", true))
		var logLevel string
		flag.StringVar(&logLevel, "logLevel", "5", "test")
		flag.Lookup("v").Value.Set(logLevel)
	}
}

func TestUnsuccessfulDecoding(t *testing.T) {
	Convey("HTTPToJSON", t, func() {
		Convey("should error when the response is not JSON", func() {
			data := strings.NewReader("just some data")
			parsed := &fakeData{}
			err := HTTPToJSON(data, []interface{}{parsed})
			So(err, ShouldNotBeNil)
		})
	})
}

func TestSuccessfulDecoding(t *testing.T) {
	Convey("For a single message, HTTPToJSON", t, func() {
		Convey("should populate a single struct", func() {
			data := strings.NewReader(
				`{"foo": "baz", "bar": [4, 6]}`)
			parsed := &fakeData{}
			err := HTTPToJSON(data, []interface{}{parsed})
			So(err, ShouldBeNil)
			So(parsed, ShouldResemble,
				&fakeData{"baz", []int{4, 6}})
		})
		Convey("should populate a single struct without types", func() {
			data := strings.NewReader(
				`44{"foo": "baz", "bar": [4, 6]}`)
			parsed := &fakeData{}
			err := HTTPToJSON(data, []interface{}{parsed})
			So(err, ShouldBeNil)
			So(parsed, ShouldResemble,
				&fakeData{"baz", []int{4, 6}})
		})
		Convey("should populate message type", func() {
			data := strings.NewReader(
				`44{"foo": "baz", "bar": [4, 6]}`)
			parsed := &fakeDataWithTypes{}
			err := HTTPToJSON(data, []interface{}{parsed})
			So(err, ShouldBeNil)
			So(parsed, ShouldResemble,
				&fakeDataWithTypes{"baz", []int{4, 6}, 4, 4})
		})
		Convey("should populate message and packet type", func() {
			data := strings.NewReader(
				`44{"foo": "baz", "bar": [4, 6]}`)
			parsed := &fakeDataWithTypes{}
			err := HTTPToJSON(data, []interface{}{parsed})
			So(err, ShouldBeNil)
			So(parsed, ShouldResemble,
				&fakeDataWithTypes{"baz", []int{4, 6}, 4, 4})
		})
		Convey("should populate only types", func() {
			data := strings.NewReader(`44`)
			parsed := &fakeDataWithTypes{}
			err := HTTPToJSON(data, []interface{}{parsed})
			So(err, ShouldBeNil)
			So(parsed, ShouldResemble,
				&fakeDataWithTypes{"", []int(nil), 4, 4})
		})
		Convey("should handle length encoding", func() {
			data := strings.NewReader(
				`31:44{"foo": "baz", "bar": [4, 6]}`)
			parsed := &fakeDataWithTypes{}
			err := HTTPToJSON(data, []interface{}{parsed})
			So(err, ShouldBeNil)
			So(parsed, ShouldResemble,
				&fakeDataWithTypes{"baz", []int{4, 6}, 4, 4})
		})

	})
}
func TestSuccessfulDecodingMultipleMessages(t *testing.T) {
	Convey("For a multiple messages, HTTPToJSON", t, func() {
		Convey("should populate many structs", func() {
			data := strings.NewReader(
				`31:44{"foo": "baz", "bar": [4, 6]}31:44{"foo": "baz", "bar": [4, 6]}`)
			parsedOne := &fakeData{}
			parsedTwo := &fakeDataWithTypes{}
			err := HTTPToJSON(data,
				[]interface{}{parsedOne, parsedTwo})
			So(err, ShouldBeNil)
			So(parsedOne, ShouldResemble,
				&fakeData{"baz", []int{4, 6}})
			So(parsedTwo, ShouldResemble,
				&fakeDataWithTypes{"baz", []int{4, 6}, 4, 4})
		})

	})
}
