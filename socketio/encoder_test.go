package socketio_test

import (
	"flag"
	"fmt"
	"io/ioutil"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	. "github.com/timpalpant/go-iex/socketio"
)

type testStruct struct {
	Foo string
	Bar []int
}

type biggerTestStruct struct {
	Foo string
	Bar []int
	Baz float64
	Biz []string
}

type invalidComplex struct {
	Invalid map[string]int
}

type invalidArrayComplex struct {
	Invalid []map[int]string
}

func init() {
	flag.Set("alsologtostderr", fmt.Sprintf("%t", true))
	var logLevel string
	flag.StringVar(&logLevel, "logLevel", "5", "test")
	flag.Lookup("v").Value.Set(logLevel)
}

func TestWebsocketEncodingErrors(t *testing.T) {
	Convey("Websocket encoding should fail", t, func() {
		encoder := NewWSEncoder("/")
		Convey("on complex types", func() {
			invalid := &invalidComplex{map[string]int{
				"foo": 3,
				"bar": 5,
			}}
			_, err := encoder.Encode(-1, -1, invalid)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring,
				"Cannot encode type")
		})
		Convey("on Arrays of complex types", func() {
			invalid := &invalidArrayComplex{[]map[int]string{{
				1: "one",
				2: "two",
			}}}
			_, err := encoder.Encode(-1, -1, invalid)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring,
				"Cannot encode Array type")
		})
	})
}

func TestHTTPEncodingErrors(t *testing.T) {
	Convey("HTTP encoding should fail", t, func() {
		encoder := NewHTTPEncoder("/")
		Convey("on complex types", func() {
			invalid := &invalidComplex{map[string]int{
				"foo": 3,
				"bar": 5,
			}}
			_, err := encoder.Encode(-1, -1, invalid)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring,
				"Cannot encode type")
		})
		Convey("on Arrays of complex types", func() {
			invalid := &invalidArrayComplex{[]map[int]string{{
				1: "one",
				2: "two",
			}}}
			_, err := encoder.Encode(-1, -1, invalid)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring,
				"Cannot encode Array type")
		})
	})
}

func TestWebsocketEncoding(t *testing.T) {
	Convey("Websocket encoding should", t, func() {
		Convey("correctly encode a simple type", func() {
			encoder := NewWSEncoder("")
			encoded, err := encoder.Encode(-1, -1, &testStruct{
				"foo", []int{1, 2, 3}})
			So(err, ShouldBeNil)
			val, err := ioutil.ReadAll(encoded)
			So(err, ShouldBeNil)
			So(string(val), ShouldEqual, `["foo","1,2,3"]`)
		})
		Convey("correctly encode a namespace", func() {
			encoder := NewWSEncoder("/")
			encoded, err := encoder.Encode(-1, -1, &testStruct{
				"foo", []int{1, 2, 3}})
			So(err, ShouldBeNil)
			val, err := ioutil.ReadAll(encoded)
			So(err, ShouldBeNil)
			So(string(val), ShouldEqual, `/,["foo","1,2,3"]`)
		})
		Convey("correctly encode a longer namespace", func() {
			encoder := NewWSEncoder("/1.0/tops")
			encoded, err := encoder.Encode(-1, -1, &testStruct{
				"foo", []int{1, 2, 3}})
			So(err, ShouldBeNil)
			val, err := ioutil.ReadAll(encoded)
			So(err, ShouldBeNil)
			So(string(val), ShouldEqual, `/1.0/tops,["foo","1,2,3"]`)
		})
		Convey("correctly encode the packet type", func() {
			encoder := NewWSEncoder("/1.0/tops")
			encoded, err := encoder.Encode(4, -1, &testStruct{
				"foo", []int{1, 2, 3}})
			So(err, ShouldBeNil)
			val, err := ioutil.ReadAll(encoded)
			So(err, ShouldBeNil)
			So(string(val), ShouldEqual,
				`4/1.0/tops,["foo","1,2,3"]`)
		})
		Convey("correctly encode the packet and message type", func() {
			encoder := NewWSEncoder("/1.0/tops")
			encoded, err := encoder.Encode(4, 2, &testStruct{
				"foo", []int{1, 2, 3}})
			So(err, ShouldBeNil)
			val, err := ioutil.ReadAll(encoded)
			So(err, ShouldBeNil)
			So(string(val), ShouldEqual,
				`42/1.0/tops,["foo","1,2,3"]`)
		})
		Convey("correctly encode a more complex type", func() {
			encoder := NewWSEncoder("/1.0/tops")
			encoded, err := encoder.Encode(4, 2, &biggerTestStruct{
				"foo", []int{1, 2, 3}, 32, []string{"a", "b"}})
			So(err, ShouldBeNil)
			val, err := ioutil.ReadAll(encoded)
			So(err, ShouldBeNil)
			So(string(val), ShouldEqual,
				`42/1.0/tops,["foo","1,2,3","32","a,b"]`)
		})
	})
}

func TestHTTPEncoding(t *testing.T) {
	Convey("HTTP encoding should", t, func() {
		Convey("correctly encode a simple type", func() {
			encoder := NewHTTPEncoder("")
			encoded, err := encoder.Encode(-1, -1, &testStruct{
				"foo", []int{1, 2, 3}})
			So(err, ShouldBeNil)
			val, err := ioutil.ReadAll(encoded)
			So(err, ShouldBeNil)
			So(string(val), ShouldEqual, `15:["foo","1,2,3"]`)
		})
		Convey("correctly encode a namespace", func() {
			encoder := NewHTTPEncoder("/")
			encoded, err := encoder.Encode(-1, -1, &testStruct{
				"foo", []int{1, 2, 3}})
			So(err, ShouldBeNil)
			val, err := ioutil.ReadAll(encoded)
			So(err, ShouldBeNil)
			So(string(val), ShouldEqual, `17:/,["foo","1,2,3"]`)
		})
		Convey("correctly encode a longer namespace", func() {
			encoder := NewHTTPEncoder("/1.0/tops")
			encoded, err := encoder.Encode(-1, -1, &testStruct{
				"foo", []int{1, 2, 3}})
			So(err, ShouldBeNil)
			val, err := ioutil.ReadAll(encoded)
			So(err, ShouldBeNil)
			So(string(val), ShouldEqual,
				`25:/1.0/tops,["foo","1,2,3"]`)
		})
		Convey("correctly encode the packet type", func() {
			encoder := NewHTTPEncoder("/1.0/tops")
			encoded, err := encoder.Encode(4, -1, &testStruct{
				"foo", []int{1, 2, 3}})
			So(err, ShouldBeNil)
			val, err := ioutil.ReadAll(encoded)
			So(err, ShouldBeNil)
			So(string(val), ShouldEqual,
				`26:4/1.0/tops,["foo","1,2,3"]`)
		})
		Convey("correctly encode the packet and message type", func() {
			encoder := NewHTTPEncoder("/1.0/tops")
			encoded, err := encoder.Encode(4, 2, &testStruct{
				"foo", []int{1, 2, 3}})
			So(err, ShouldBeNil)
			val, err := ioutil.ReadAll(encoded)
			So(err, ShouldBeNil)
			So(string(val), ShouldEqual,
				`27:42/1.0/tops,["foo","1,2,3"]`)
		})
		Convey("correctly encode a more complex type", func() {
			encoder := NewHTTPEncoder("/1.0/tops")
			encoded, err := encoder.Encode(4, 2, &biggerTestStruct{
				"foo", []int{1, 2, 3}, 32, []string{"a", "b"}})
			So(err, ShouldBeNil)
			val, err := ioutil.ReadAll(encoded)
			So(err, ShouldBeNil)
			So(string(val), ShouldEqual,
				`38:42/1.0/tops,["foo","1,2,3","32","a,b"]`)
		})
	})
}
