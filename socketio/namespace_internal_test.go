package socketio

import (
	"flag"
	"io"
	"strings"
	"sync"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/timpalpant/go-iex"
)

func init() {
	flag.Set("v", "5")
}

type channelWriter struct {
	sync.RWMutex
	io.Writer

	Messages []string
	internal chan string
	C        chan interface{}
}

func (c *channelWriter) Write(data []byte) (int, error) {
	c.internal <- string(data)
	return len(data), nil
}

func (c *channelWriter) listen(closeAfter int) {
	go func(closeAfter int) {
		for msg := range c.internal {
			c.Lock()
			c.Messages = append(c.Messages, msg)
			if len(c.Messages) >= closeAfter {
				c.Unlock()
				close(c.C)
				return
			}
			c.Unlock()
		}
	}(closeAfter)
}

func waitOnClose(c chan interface{}) {
	for {
		_, open := <-c
		if !open {
			return
		}
	}
}

func TestNamespace(t *testing.T) {
	Convey("The IexTOPSNamespace should", t, func() {
		ch := make(chan packetMetadata, 1)
		encoder := NewWSEncoder("/1.0/tops")
		writer := &channelWriter{
			Messages: make([]string, 0),
			internal: make(chan string, 0),
			C:        make(chan interface{}, 0),
		}
		closeFuncCalled := make(chan interface{})
		closeFunc := func() {
			close(closeFuncCalled)
		}
		subFactory := func(
			signal subOrUnsub, symbols []string) *IEXMsg {
			return &IEXMsg{
				EventType: signal,
				Data:      strings.Join(symbols, ","),
			}
		}
		Convey("send a connect message", func() {
			writer.listen(1)
			newIexTOPSNamespace(
				ch, encoder, writer, subFactory, closeFunc)
			waitOnClose(writer.C)
			So(writer.Messages[0], ShouldEqual, "40/1.0/tops,")
		})
		Convey("send no subscription on empty connection", func() {
			writer.listen(1)
			ns := newIexTOPSNamespace(
				ch, encoder, writer, subFactory, closeFunc)
			ns.GetConnection()
			waitOnClose(writer.C)
			So(writer.Messages[0], ShouldEqual, "40/1.0/tops,")
		})
		Convey("send subscription messages", func() {
			writer.listen(2)
			ns := newIexTOPSNamespace(
				ch, encoder, writer, subFactory, closeFunc)
			ns.GetConnection("fb", "snap")
			waitOnClose(writer.C)
			So(writer.Messages[1], ShouldEqual,
				`42/1.0/tops,["subscribe","fb,snap"]`)
		})
		Convey("send multiple subscription messages", func() {
			writer.listen(3)
			ns := newIexTOPSNamespace(
				ch, encoder, writer, subFactory, closeFunc)
			conn := ns.GetConnection("fb", "snap")
			conn.Subscribe("goog")
			waitOnClose(writer.C)
			So(`42/1.0/tops,["subscribe","fb,snap"]`,
				ShouldBeIn, writer.Messages)
			So(`42/1.0/tops,["subscribe","goog"]`,
				ShouldBeIn, writer.Messages)
		})
		Convey("send unsubscribe messages", func() {
			writer.listen(3)
			ns := newIexTOPSNamespace(
				ch, encoder, writer, subFactory, closeFunc)
			conn := ns.GetConnection("fb", "snap")
			conn.Unsubscribe("goog")
			waitOnClose(writer.C)
			So(`42/1.0/tops,["subscribe","fb,snap"]`,
				ShouldBeIn, writer.Messages)
			So(`42/1.0/tops,["unsubscribe","goog"]`,
				ShouldBeIn, writer.Messages)
		})
		Convey("call closeFunc when all connections closed", func() {
			writer.listen(1)
			ns := newIexTOPSNamespace(
				ch, encoder, writer, subFactory, closeFunc)
			waitOnClose(writer.C)
			conn1 := ns.GetConnection()
			conn2 := ns.GetConnection()
			conn1.Close()
			conn2.Close()
			_, ok := <-closeFuncCalled
			So(ok, ShouldBeFalse)
		})
		Convey("fan out messages", func() {
			writer.listen(3)
			ns := newIexTOPSNamespace(
				ch, encoder, writer, subFactory, closeFunc)
			conn1 := ns.GetConnection("fb")
			conn2 := ns.GetConnection("fb")
			waitOnClose(writer.C)
			ch <- packetMetadata{
				Data: "{\"symbol\":\"fb\",\"bidsize\":12}",
			}
			expected := iex.TOPS{
				Symbol:  "fb",
				BidSize: 12,
			}
			So(<-conn1.C, ShouldResemble, expected)
			So(<-conn2.C, ShouldResemble, expected)
		})
		Convey("filter based on subscriptions", func() {
			writer.listen(3)
			ns := newIexTOPSNamespace(
				ch, encoder, writer, subFactory, closeFunc)
			conn1 := ns.GetConnection("fb")
			conn2 := ns.GetConnection("goog")
			waitOnClose(writer.C)
			ch <- packetMetadata{
				Data: "{\"symbol\":\"fb\",\"bidsize\":12}",
			}
			fbExpected := iex.TOPS{
				Symbol:  "fb",
				BidSize: 12,
			}
			So(<-conn1.C, ShouldResemble, fbExpected)
			So(len(conn2.C), ShouldEqual, 0)
			ch <- packetMetadata{
				Data: "{\"symbol\":\"goog\",\"bidsize\":11}",
			}
			googExpected := iex.TOPS{
				Symbol:  "goog",
				BidSize: 11,
			}
			So(len(conn1.C), ShouldEqual, 0)
			So(<-conn2.C, ShouldResemble, googExpected)
			ch <- packetMetadata{
				Data: "{\"symbol\":\"aig+\",\"bidsize\":11}",
			}
			So(len(conn1.C), ShouldEqual, 0)
			So(len(conn2.C), ShouldEqual, 0)
		})
		Convey("close outgoing when incoming closed", func() {
			writer.listen(3)
			ns := newIexTOPSNamespace(
				ch, encoder, writer, subFactory, closeFunc)
			conn1 := ns.GetConnection("fb")
			conn2 := ns.GetConnection("goog")
			waitOnClose(writer.C)
			close(ch)
			_, ok := <-conn1.C
			So(ok, ShouldBeFalse)
			_, ok = <-conn2.C
			So(ok, ShouldBeFalse)
		})
	})
}
