package socketio_test

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	. "github.com/timpalpant/go-iex/socketio"
)

func TestPresenceSubscriber(t *testing.T) {
	Convey("The PresenceSubscriber should", t, func() {
		subscriber := NewPresenceSubscriber()
		Convey("return false by default", func() {
			So(subscriber.Subscribed("FB"), ShouldBeFalse)
		})
		Convey("returns true after subscription", func() {
			subscriber.Subscribe("FB")
			So(subscriber.Subscribed("FB"), ShouldBeTrue)
		})
		Convey("returns true after multiple subscriptions", func() {
			subscriber.Subscribe("FB")
			So(subscriber.Subscribed("FB"), ShouldBeTrue)
			subscriber.Subscribe("FB")
			So(subscriber.Subscribed("FB"), ShouldBeTrue)
		})
		Convey("returns false after unsubscription", func() {
			subscriber.Subscribe("FB")
			So(subscriber.Subscribed("FB"), ShouldBeTrue)
			subscriber.Subscribe("FB")
			So(subscriber.Subscribed("FB"), ShouldBeTrue)
			subscriber.Unsubscribe("FB")
			So(subscriber.Subscribed("FB"), ShouldBeFalse)
			subscriber.Unsubscribe("FB")
			So(subscriber.Subscribed("FB"), ShouldBeFalse)
		})
	})
}

func TestCountingSubscriber(t *testing.T) {
	Convey("The CountingSubscriber should", t, func() {
		subscriber := NewCountingSubscriber()
		Convey("return false by default", func() {
			So(subscriber.Subscribed("FB"), ShouldBeFalse)
		})
		Convey("returns true after subscription", func() {
			subscriber.Subscribe("FB")
			So(subscriber.Subscribed("FB"), ShouldBeTrue)
		})
		Convey("returns true after multiple subscriptions", func() {
			subscriber.Subscribe("FB")
			So(subscriber.Subscribed("FB"), ShouldBeTrue)
			subscriber.Subscribe("FB")
			So(subscriber.Subscribed("FB"), ShouldBeTrue)
		})
		Convey("requires corresponding unsubscriptions", func() {
			subscriber.Subscribe("FB")
			So(subscriber.Subscribed("FB"), ShouldBeTrue)
			subscriber.Subscribe("FB")
			So(subscriber.Subscribed("FB"), ShouldBeTrue)

			subscriber.Unsubscribe("FB")
			So(subscriber.Subscribed("FB"), ShouldBeTrue)
			subscriber.Unsubscribe("FB")
			So(subscriber.Subscribed("FB"), ShouldBeFalse)
			subscriber.Unsubscribe("FB")
			So(subscriber.Subscribed("FB"), ShouldBeFalse)
		})
	})
}
