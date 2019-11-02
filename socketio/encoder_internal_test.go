package socketio

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestSubUnsubMsgFactory(t *testing.T) {
	Convey("The simpleSubUnsubFactory should", t, func() {
		Convey("returns IEXMsg", func() {
			msg := simpleSubUnsubFactory(Subscribe, []string{
				"fb", "snap",
			})
			So(msg, ShouldResemble, &IEXMsg{
				EventType: Subscribe,
				Data:      "fb,snap",
			})
			msg = simpleSubUnsubFactory(Unsubscribe, []string{
				"goog", "aig+",
			})
			So(msg, ShouldResemble, &IEXMsg{
				EventType: Unsubscribe,
				Data:      "goog,aig+",
			})
		})
	})
	Convey("The deepSubUnsubFactory should", t, func() {
		Convey("returns IEXMsg", func() {
			msg := deepSubUnsubFactory(Subscribe, []string{
				"fb",
			})
			So(msg, ShouldResemble, &IEXMsg{
				EventType: Subscribe,
				Data:      "{\"symbols\":[\"fb\"],\"channels\":[\"deep\"]}",
			})
			msg = deepSubUnsubFactory(Unsubscribe, []string{
				"goog",
			})
			So(msg, ShouldResemble, &IEXMsg{
				EventType: Unsubscribe,
				Data:      "{\"symbols\":[\"goog\"],\"channels\":[\"deep\"]}",
			})
		})
	})
}
