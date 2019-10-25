package socketio_test

import (
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/timpalpant/go-iex"
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
	Namespace   string
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
				&fakeDataWithTypes{
					"baz", []int{4, 6}, 4, 4, ""})
		})
		Convey("should populate message and packet type", func() {
			data := strings.NewReader(
				`44{"foo": "baz", "bar": [4, 6]}`)
			parsed := &fakeDataWithTypes{}
			err := HTTPToJSON(data, []interface{}{parsed})
			So(err, ShouldBeNil)
			So(parsed, ShouldResemble,
				&fakeDataWithTypes{
					"baz", []int{4, 6}, 4, 4, ""})
		})
		Convey("should populate only types", func() {
			data := strings.NewReader(`44`)
			parsed := &fakeDataWithTypes{}
			err := HTTPToJSON(data, []interface{}{parsed})
			So(err, ShouldBeNil)
			So(parsed, ShouldResemble,
				&fakeDataWithTypes{"", []int(nil), 4, 4, ""})
		})
		Convey("should handle length encoding", func() {
			data := strings.NewReader(
				`31:44{"foo": "baz", "bar": [4, 6]}`)
			parsed := &fakeDataWithTypes{}
			err := HTTPToJSON(data, []interface{}{parsed})
			So(err, ShouldBeNil)
			So(parsed, ShouldResemble,
				&fakeDataWithTypes{
					"baz", []int{4, 6}, 4, 4, ""})
		})
		Convey("should handle length and namespace encoding", func() {
			data := strings.NewReader(
				`37:42/1.0/tops,{"foo":"baz","bar":[4,6]}`)
			parsed := &fakeDataWithTypes{}
			err := HTTPToJSON(data, []interface{}{parsed})
			So(err, ShouldBeNil)
			So(parsed, ShouldResemble,
				&fakeDataWithTypes{"baz", []int{4, 6},
					2, 4, "/1.0/tops"})
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
				&fakeDataWithTypes{
					"baz", []int{4, 6}, 4, 4, ""})
		})

	})
}
func TestDecodeActualTops(t *testing.T) {
	Convey("For an actual Tops response, HTTPToJSON", t, func() {
		Convey("should populate a Tops message", func() {
			data := strings.NewReader(`348:42/1.0/tops,["message","{\"symbol\":\"SNAP\",\"sector\":\"mediaentertainment\",\"securityType\":\"commonstock\",\"bidPrice\":0.0000,\"bidSize\":0,\"askPrice\":0.0000,\"askSize\":0,\"lastUpdated\":1569873716685,\"lastSalePrice\":15.8000,\"lastSaleSize\":100,\"lastSaleTime\":1569873590063,\"volume\":458065,\"marketPercent\":0.02262,\"seq\":26739}"]344:42/1.0/tops,["message","{\"symbol\":\"FB\",\"sector\":\"mediaentertainment\",\"securityType\":\"commonstock\",\"bidPrice\":0.0000,\"bidSize\":0,\"askPrice\":0.0000,\"askSize\":0,\"lastUpdated\":1569876755318,\"lastSalePrice\":178.0750,\"lastSaleSize\":1,\"lastSaleTime\":1569873595907,\"volume\":411341,\"marketPercent\":0.03700,\"seq\":5904}"]325:42/1.0/tops,["message","{\"symbol\":\"AIG+\",\"sector\":\"n/a\",\"securityType\":\"warrant\",\"bidPrice\":0.0000,\"bidSize\":0,\"askPrice\":0.0000,\"askSize\":0,\"lastUpdated\":1569873600001,\"lastSalePrice\":14.3700,\"lastSaleSize\":200,\"lastSaleTime\":1569859449771,\"volume\":211,\"marketPercent\":0.00632,\"seq\":7281}"]`)
			parsedOne := &iex.TOPS{}
			parsedTwo := &iex.TOPS{}
			parsedThree := &iex.TOPS{}
			err := HTTPToJSON(data,
				[]interface{}{
					parsedOne, parsedTwo, parsedThree})
			So(err, ShouldBeNil)
			So(parsedOne.Symbol, ShouldEqual, "SNAP")
			So(parsedTwo.Symbol, ShouldEqual, "FB")
			So(parsedThree.Symbol, ShouldEqual, "AIG+")
		})

	})
}
