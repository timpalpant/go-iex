package socketio_test

import (
	"sync"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	. "github.com/timpalpant/go-iex/socketio"
)

type fakeTransport struct {
	sync.Mutex

	closed bool
	Writer channelWriter
}

func (f *fakeTransport) Write(data []byte) (int, error) {
	return f.Writer.Write(data)
}

func (f *fakeTransport) GetReadChannel() (<-chan PacketData, error) {
	return make(chan PacketData, 0), nil
}

func (f *fakeTransport) Close() {
	f.Lock()
	defer f.Unlock()
	f.closed = true
}

func TestClient(t *testing.T) {
	Convey("The Client should", t, func() {
		ft := &fakeTransport{
			closed: false,
			Writer: channelWriter{
				Messages: make([]string, 0),
				internal: make(chan string, 0),
				C:        make(chan interface{}, 0),
			},
		}
		Convey("send DEEP connect", func() {
			ft.Writer.listen(1)
			client := NewClientWithTransport(ft)
			client.GetDEEPNamespace()
			waitOnClose(ft.Writer.C)
			So(ft.Writer.Messages[0], ShouldEqual, "40/1.0/deep,")
		})
		Convey("send Last connect", func() {
			ft.Writer.listen(1)
			client := NewClientWithTransport(ft)
			client.GetLastNamespace()
			waitOnClose(ft.Writer.C)
			So(ft.Writer.Messages[0], ShouldEqual, "40/1.0/last,")
		})
		Convey("send TOPS connect", func() {
			ft.Writer.listen(1)
			client := NewClientWithTransport(ft)
			client.GetTOPSNamespace()
			waitOnClose(ft.Writer.C)
			So(ft.Writer.Messages[0], ShouldEqual, "40/1.0/tops,")
		})
		Convey("close DEEP connect", func() {
			ft.Writer.listen(4)
			client := NewClientWithTransport(ft)
			ns := client.GetDEEPNamespace()
			conn1 := ns.GetConnection("fb")
			conn2 := ns.GetConnection("goog")
			conn1.Close()
			conn2.Close()
			waitOnClose(ft.Writer.C)
			So(ft.Writer.Messages[0], ShouldEqual, "40/1.0/deep,")
			So(ft.Writer.Messages[3], ShouldEqual, "41/1.0/deep,")
		})
		Convey("close Last connect", func() {
			ft.Writer.listen(4)
			client := NewClientWithTransport(ft)
			ns := client.GetLastNamespace()
			conn1 := ns.GetConnection("fb")
			conn2 := ns.GetConnection("goog")
			conn1.Close()
			conn2.Close()
			waitOnClose(ft.Writer.C)
			So(ft.Writer.Messages[0], ShouldEqual, "40/1.0/last,")
			So(ft.Writer.Messages[3], ShouldEqual, "41/1.0/last,")
		})
		Convey("close TOPS connect", func() {
			ft.Writer.listen(4)
			client := NewClientWithTransport(ft)
			ns := client.GetTOPSNamespace()
			conn1 := ns.GetConnection("fb")
			conn2 := ns.GetConnection("goog")
			conn1.Close()
			conn2.Close()
			waitOnClose(ft.Writer.C)
			So(ft.Writer.Messages[0], ShouldEqual, "40/1.0/tops,")
			So(ft.Writer.Messages[3], ShouldEqual, "41/1.0/tops,")
		})
	})
}
