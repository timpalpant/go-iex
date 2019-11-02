package socketio_test

import (
	"sync"
	"testing"

	"github.com/golang/glog"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/timpalpant/go-iex"
	. "github.com/timpalpant/go-iex/socketio"
)

type fakeTransport struct {
	sync.Mutex

	messages  []string
	callbacks map[string]map[int]func(PacketData)
	nextId    int
	closed    bool
}

func newFakeTransport() *fakeTransport {
	return &fakeTransport{
		messages:  make([]string, 0),
		callbacks: make(map[string]map[int]func(PacketData)),
		nextId:    0,
		closed:    false,
	}
}

func (f *fakeTransport) Write(data []byte) (int, error) {
	glog.Infof("Fake transport writing message: %s", string(data))
	f.messages = append(f.messages, string(data))
	return len(data), nil
}

func (f *fakeTransport) AddPacketCallback(
	namespace string, callback func(PacketData)) (int, error) {
	f.Lock()
	defer f.Unlock()
	f.nextId++
	if _, ok := f.callbacks[namespace]; !ok {
		f.callbacks[namespace] = make(map[int]func(PacketData))
	}
	f.callbacks[namespace][f.nextId] = callback
	return f.nextId, nil
}

func (f *fakeTransport) RemovePacketCallback(namespace string, id int) error {
	f.Lock()
	defer f.Unlock()
	if _, ok := f.callbacks[namespace]; ok {
		delete(f.callbacks[namespace], id)
	}
	return nil
}

func (f *fakeTransport) TriggerCallbacks(pkt PacketData) {
	f.Lock()
	defer f.Unlock()
	if ns, ok := f.callbacks[pkt.Namespace]; ok {
		for _, callback := range ns {
			callback(pkt)
		}
	}
}

func (f *fakeTransport) Close() {
	f.Lock()
	defer f.Unlock()
	f.closed = true
}

func TestClient(t *testing.T) {
	Convey("The Client should", t, func() {
		ft := newFakeTransport()
		Convey("send DEEP connect", func() {
			client := NewClientWithTransport(ft)
			ns := client.GetDEEPNamespace()
			handler := func(msg iex.DEEP) {}
			ns.SubscribeTo(handler, "fb")
			So(ft.messages[0], ShouldEqual, "40/1.0/deep,")
		})
		Convey("send Last connect", func() {
			client := NewClientWithTransport(ft)
			ns := client.GetLastNamespace()
			handler := func(msg iex.Last) {}
			ns.SubscribeTo(handler, "fb")
			So(ft.messages[0], ShouldEqual, "40/1.0/last,")
		})
		Convey("send TOPS connect", func() {
			client := NewClientWithTransport(ft)
			ns := client.GetTOPSNamespace()
			handler := func(msg iex.TOPS) {}
			ns.SubscribeTo(handler, "fb")
			So(ft.messages[0], ShouldEqual, "40/1.0/tops,")
		})
		Convey("close DEEP connect", func() {
			client := NewClientWithTransport(ft)
			ns := client.GetDEEPNamespace()
			handler := func(msg iex.DEEP) {}
			closer1, err := ns.SubscribeTo(handler, "fb")
			So(err, ShouldBeNil)
			closer2, err := ns.SubscribeTo(handler, "goog")
			So(err, ShouldBeNil)
			closer1()
			closer2()
			So("40/1.0/deep,", ShouldBeIn, ft.messages)
			So("41/1.0/deep,", ShouldBeIn, ft.messages)
		})
		Convey("close Last connect", func() {
			client := NewClientWithTransport(ft)
			ns := client.GetLastNamespace()
			handler := func(msg iex.Last) {}
			closer1, err := ns.SubscribeTo(handler, "fb")
			So(err, ShouldBeNil)
			closer2, err := ns.SubscribeTo(handler, "goog")
			So(err, ShouldBeNil)
			closer1()
			closer2()
			So("40/1.0/last,", ShouldBeIn, ft.messages)
			So("41/1.0/last,", ShouldBeIn, ft.messages)
		})
		Convey("close TOPS connect", func() {
			client := NewClientWithTransport(ft)
			ns := client.GetTOPSNamespace()
			handler := func(msg iex.TOPS) {}
			closer1, err := ns.SubscribeTo(handler, "fb")
			So(err, ShouldBeNil)
			closer2, err := ns.SubscribeTo(handler, "goog")
			So(err, ShouldBeNil)
			closer1()
			closer2()
			So("40/1.0/tops,", ShouldBeIn, ft.messages)
			So("41/1.0/tops,", ShouldBeIn, ft.messages)
		})
	})
}
