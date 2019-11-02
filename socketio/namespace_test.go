package socketio_test

import (
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/timpalpant/go-iex"
	. "github.com/timpalpant/go-iex/socketio"
)

func TestNamespace(t *testing.T) {
	Convey("The IexTOPSNamespace should", t, func() {
		ft := newFakeTransport()
		subFactory := func(
			signal SubOrUnsub, symbols []string) *IEXMsg {
			return &IEXMsg{
				EventType: signal,
				Data:      strings.Join(symbols, ","),
			}
		}
		closed := false
		closedNamespace := ""
		closeFunc := func(namespace string) {
			closedNamespace = namespace
			closed = true
		}
		Convey("not send a connect on creation", func() {
			NewIexTOPSNamespace(ft, subFactory, closeFunc)
			So(ft.messages, ShouldHaveLength, 0)
		})
		Convey("error on SubscribeTo with no symbols", func() {
			ns := NewIexTOPSNamespace(ft, subFactory, closeFunc)
			handler := func(msg iex.TOPS) {}
			_, err := ns.SubscribeTo(handler)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "no symbols")
		})
		Convey("send a connect message on first subscription", func() {
			ns := NewIexTOPSNamespace(ft, subFactory, closeFunc)
			handler := func(msg iex.TOPS) {}
			_, err := ns.SubscribeTo(handler, "fb", "snap")
			So(err, ShouldBeNil)
			So(ft.messages[0], ShouldEqual, "40/1.0/tops,")
			So(ft.messages[1], ShouldEqual,
				`42/1.0/tops,["subscribe","fb,snap"]`)
		})
		Convey("send unsubscribe messages", func() {
			ns := NewIexTOPSNamespace(ft, subFactory, closeFunc)
			handler := func(msg iex.TOPS) {}
			closer, err := ns.SubscribeTo(handler, "fb", "snap")
			So(err, ShouldBeNil)
			closer()
			So(`42/1.0/tops,["subscribe","fb,snap"]`,
				ShouldBeIn, ft.messages)
			So(`42/1.0/tops,["unsubscribe","fb,snap"]`,
				ShouldBeIn, ft.messages)
		})
		Convey("unsubscribe when all references removed", func() {
			ns := NewIexTOPSNamespace(ft, subFactory, closeFunc)
			handler := func(msg iex.TOPS) {}
			closer1, err := ns.SubscribeTo(handler, "fb", "snap")
			So(err, ShouldBeNil)
			closer2, err := ns.SubscribeTo(handler, "fb", "goog")
			So(err, ShouldBeNil)
			closer1()
			closer2()
			So(`42/1.0/tops,["subscribe","fb,snap"]`,
				ShouldBeIn, ft.messages)
			So(`42/1.0/tops,["unsubscribe","snap"]`,
				ShouldBeIn, ft.messages)
			So(`42/1.0/tops,["unsubscribe","fb,goog"]`,
				ShouldBeIn, ft.messages)
		})
		Convey("call closeFunc when all connections closed", func() {
			ns := NewIexTOPSNamespace(ft, subFactory, closeFunc)
			handler := func(msg iex.TOPS) {}
			closer1, err := ns.SubscribeTo(handler, "fb")
			So(err, ShouldBeNil)
			closer2, err := ns.SubscribeTo(handler, "fb")
			So(err, ShouldBeNil)
			closer1()
			closer2()
			So(closedNamespace, ShouldEqual, "/1.0/tops")
			So(closed, ShouldBeTrue)
		})
		Convey("fan out messages", func() {
			ns := NewIexTOPSNamespace(ft, subFactory, closeFunc)
			var msg1 iex.TOPS
			handler1 := func(msg iex.TOPS) {
				msg1 = msg
			}
			_, err := ns.SubscribeTo(handler1, "fb")
			So(err, ShouldBeNil)
			var msg2 iex.TOPS
			handler2 := func(msg iex.TOPS) {
				msg2 = msg
			}
			_, err = ns.SubscribeTo(handler2, "fb")
			So(err, ShouldBeNil)
			ft.callbacks["/1.0/tops"][1](PacketData{
				Data: "{\"symbol\":\"fb\",\"bidsize\":12}",
			})
			expected := iex.TOPS{
				Symbol:  "fb",
				BidSize: 12,
			}
			So(msg1, ShouldResemble, expected)
			So(msg2, ShouldResemble, expected)
		})
		Convey("filter based on subscriptions", func() {
			ns := NewIexTOPSNamespace(ft, subFactory, closeFunc)
			var msg1 iex.TOPS
			handler1 := func(msg iex.TOPS) {
				msg1 = msg
			}
			_, err := ns.SubscribeTo(handler1, "fb")
			So(err, ShouldBeNil)
			var msg2 iex.TOPS
			handler2 := func(msg iex.TOPS) {
				msg2 = msg
			}
			_, err = ns.SubscribeTo(handler2, "goog")
			So(err, ShouldBeNil)
			ft.TriggerCallbacks(PacketData{
				Namespace: "/1.0/tops",
				Data:      "{\"symbol\":\"fb\",\"bidsize\":12}",
			})
			fbExpected := iex.TOPS{
				Symbol:  "fb",
				BidSize: 12,
			}
			So(msg1, ShouldResemble, fbExpected)
			So(msg2, ShouldResemble, iex.TOPS{})
			ft.TriggerCallbacks(PacketData{
				Namespace: "/1.0/tops",
				Data:      "{\"symbol\":\"goog\",\"bidsize\":11}",
			})
			googExpected := iex.TOPS{
				Symbol:  "goog",
				BidSize: 11,
			}
			So(msg2, ShouldResemble, googExpected)
			msg1 = iex.TOPS{}
			msg2 = iex.TOPS{}
			ft.TriggerCallbacks(PacketData{
				Namespace: "/1.0/tops",
				Data:      "{\"symbol\":\"aig+\",\"bidsize\":11}",
			})
			So(msg1, ShouldResemble, iex.TOPS{})
			So(msg2, ShouldResemble, iex.TOPS{})
		})
	})
}
