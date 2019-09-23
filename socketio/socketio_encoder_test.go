package socketio_test

import (
	"io/ioutil"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	. "github.com/timpalpant/go-iex/socketio"
)

func TestEncodeMessages(t *testing.T) {
	encoder := NewSocketioEncoder("/1.0/tops/")
	Convey("The subscribe message", t, func() {
		message := NewSubscribeMessage([]string{"fb", "snap"})
		reader := encoder.Encode(message)
		encoded, err := ioutil.ReadAll(reader)
		if err != nil {
			t.Errorf("Error encoding subscribe: %s", err)
		}
		Convey("should return the correct prefix", func() {
			So(string(encoded), ShouldStartWith, "2/1.0/tops,")
		})
		Convey("should return the correct suffix", func() {
			So(string(encoded), ShouldEndWith, "[\"subscribe\",\"fb,snap\"]")
		})
	})
}
