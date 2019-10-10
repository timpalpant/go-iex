package socketio_test

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	. "github.com/timpalpant/go-iex/socketio"
)

type response struct {
	resp *http.Response
	err  error
}

type fakeDoClient struct {
	Requests        []*http.Request
	responses       []*response
	currentResponse int
}

func (m *fakeDoClient) Do(req *http.Request) (*http.Response, error) {
	m.Requests = append(m.Requests, req)
	if len(m.responses) > 0 && m.currentResponse < len(m.responses) {
		defer func() {
			m.currentResponse++
		}()
		if m.responses[m.currentResponse].err != nil {
			return nil, m.responses[m.currentResponse].err
		}
		return m.responses[m.currentResponse].resp, nil
	}
	return nil, nil
}

type message struct {
	messageType int
	message     []byte
	err         error
}

type fakeConn struct {
	messagesToReturn []*message
	messagesWritten  [][]byte
	messagesMutex    sync.Mutex
	closed           bool
	closedChan       chan bool
	currentToReturn  int
}

func (f *fakeConn) ReadMessage() (int, []byte, error) {
	numMessages := len(f.messagesToReturn)
	if numMessages > 0 && f.currentToReturn < numMessages {
		defer func() {
			f.currentToReturn++
		}()
		toReturn := f.messagesToReturn[f.currentToReturn]
		return toReturn.messageType, toReturn.message, toReturn.err
	}
	return 0, []byte{}, io.EOF
}

func (f *fakeConn) WriteMessage(messageType int, data []byte) error {
	f.messagesMutex.Lock()
	defer f.messagesMutex.Unlock()
	f.messagesWritten = append(f.messagesWritten, data)
	return nil
}

func (f *fakeConn) Close() error {
	f.closed = true
	f.closedChan <- true
	return nil
}

type fakeWsDialer struct {
	WsUrl string
	resp  *http.Response
	err   error
	conn  WSConn
}

func (w *fakeWsDialer) Dial(urlStr string, reqHeader http.Header) (
	WSConn, *http.Response, error) {
	w.WsUrl = urlStr
	if w.err != nil {
		return nil, w.resp, w.err
	}
	return w.conn, w.resp, nil
}

type fakeError struct {
	message string
}

func (f *fakeError) Error() string {
	return f.message
}

func init() {
	flag.Set("alsologtostderr", fmt.Sprintf("%t", true))
	var logLevel string
	flag.StringVar(&logLevel, "logLevel", "5", "test")
	flag.Lookup("v").Value.Set(logLevel)
}

var hsResponseString = `95:0{"sid":"N1pkgEHs-wEXi4DtAA4m","upgrades":["websocket"],"pingInterval":500,"pingTimeout":60000}`
var hsLongPingResponseString = `98:0{"sid":"N1pkgEHs-wEXi4DtAA4m","upgrades":["websocket"],"pingInterval":100000,"pingTimeout":60000}`
var hsNoUpgradesString = `86:0{"sid":"N1pkgEHs-wEXi4DtAA4m","upgrades":[],"pingInterval":25000,"pingTimeout":60000}`

var goodJoinResponse = `2:40`
var badJoinResponse = `2:22`

func TestTransport(t *testing.T) {
	Convey("The Transport layer should", t, func() {
		Convey("return an error on no response body", func() {
			requests := make([]*http.Request, 0)
			responses := make([]*response, 0)
			fdc := &fakeDoClient{requests, responses, 0}
			fw := &fakeWsDialer{}
			_, err := NewTransport(fdc, fw)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldStartWith, "No response body")
			So(len(fdc.Requests), ShouldEqual, 1)
			to := fdc.Requests[0].URL
			So(fdc.Requests[0].Method, ShouldEqual, "GET")
			So(to.Scheme, ShouldEqual, "https")
			So(to.Host, ShouldStartWith, "ws-api.iextrading.com")
			So(to.Path, ShouldEqual, "/socket.io/")
		})
		Convey("return an error on no handshake response", func() {
			requests := make([]*http.Request, 0)
			hsResponse := &response{nil, &fakeError{"No connection"}}
			responses := []*response{hsResponse}
			fdc := &fakeDoClient{requests, responses, 0}
			fw := &fakeWsDialer{}
			_, err := NewTransport(fdc, fw)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "No connection")
		})
		Convey("return an error no websocket upgrade", func() {
			requests := make([]*http.Request, 0)
			hsResponse := &http.Response{
				Body: ioutil.NopCloser(
					strings.NewReader(hsNoUpgradesString)),
			}
			responses := []*response{&response{
				resp: hsResponse,
			}}
			fdc := &fakeDoClient{requests, responses, 0}
			fw := &fakeWsDialer{}
			_, err := NewTransport(fdc, fw)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "Websocket upgrade not found")
		})
		Convey("return an error on wrong message type", func() {
			requests := make([]*http.Request, 0)
			hsResponse := &http.Response{
				Body: ioutil.NopCloser(
					strings.NewReader(hsResponseString)),
			}
			nspResponse := &http.Response{
				Body: ioutil.NopCloser(
					strings.NewReader(badJoinResponse)),
			}
			responses := []*response{&response{
				resp: hsResponse,
			}, &response{
				resp: nspResponse,
			}}
			fdc := &fakeDoClient{requests, responses, 0}
			fw := &fakeWsDialer{}
			_, err := NewTransport(fdc, fw)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldStartWith,
				"Unexpected namespace response")
		})
		Convey("return an error on failure to open websocket", func() {
			requests := make([]*http.Request, 0)
			hsResponse := &http.Response{
				Body: ioutil.NopCloser(
					strings.NewReader(hsResponseString)),
			}
			nspResponse := &http.Response{
				Body: ioutil.NopCloser(
					strings.NewReader(goodJoinResponse)),
			}
			responses := []*response{&response{
				resp: hsResponse,
			}, &response{
				resp: nspResponse,
			}}
			fdc := &fakeDoClient{requests, responses, 0}
			fw := &fakeWsDialer{
				err: &fakeError{"could not open"},
			}
			_, err := NewTransport(fdc, fw)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring,
				"could not open")
		})
		Convey("successfully handshake and upgrade", func() {
			requests := make([]*http.Request, 0)
			hsResponse := &http.Response{
				Body: ioutil.NopCloser(
					strings.NewReader(hsResponseString)),
			}
			nspResponse := &http.Response{
				Body: ioutil.NopCloser(
					strings.NewReader(goodJoinResponse)),
			}
			responses := []*response{&response{
				resp: hsResponse,
			}, &response{
				resp: nspResponse,
			}}
			fdc := &fakeDoClient{requests, responses, 0}
			fc := &fakeConn{
				closedChan: make(chan bool),
			}
			fw := &fakeWsDialer{
				conn: fc,
			}
			trans, err := NewTransport(fdc, fw)
			So(err, ShouldBeNil)
			So(len(fdc.Requests), ShouldEqual, 2)
			to := fdc.Requests[1].URL
			So(fdc.Requests[1].Method, ShouldEqual, "POST")
			So(to.Scheme, ShouldEqual, "https")
			So(to.Host, ShouldStartWith, "ws-api.iextrading.com")
			So(to.Path, ShouldEqual, "/socket.io/")
			So(to.Query().Get("sid"), ShouldEqual,
				"N1pkgEHs-wEXi4DtAA4m")
			// This should allow at least 2 heartbeats at 500ms.
			dur, _ := time.ParseDuration("1.2s")
			time.Sleep(dur)
			fc.messagesMutex.Lock()
			So(len(fc.messagesWritten), ShouldEqual, 3)
			msgs := fc.messagesWritten
			So(string(msgs[0]), ShouldEqual, "5")
			So(string(msgs[1]), ShouldEqual, "2")
			So(string(msgs[2]), ShouldEqual, "2")
			fc.messagesMutex.Unlock()

			trans.Close()

			<-fc.closedChan

			fc.messagesMutex.Lock()
			msgs = fc.messagesWritten
			So(len(fc.messagesWritten), ShouldEqual, 4)
			So(string(msgs[3]), ShouldEqual, "1")
			So(fc.closed, ShouldEqual, true)
			fc.messagesMutex.Unlock()
		})
		Convey("prevent writing to a closed transport", func() {
			requests := make([]*http.Request, 0)
			hsResponse := &http.Response{
				Body: ioutil.NopCloser(
					strings.NewReader(hsResponseString)),
			}
			nspResponse := &http.Response{
				Body: ioutil.NopCloser(
					strings.NewReader(goodJoinResponse)),
			}
			responses := []*response{&response{
				resp: hsResponse,
			}, &response{
				resp: nspResponse,
			}}
			fdc := &fakeDoClient{requests, responses, 0}
			fc := &fakeConn{}
			fw := &fakeWsDialer{
				conn: fc,
			}
			trans, err := NewTransport(fdc, fw)
			trans.Close()
			_, err = trans.Write([]byte("String"))
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring,
				"Cannot write to a closed transport")
		})
		Convey("prevent reading from a closed transport", func() {
			requests := make([]*http.Request, 0)
			hsResponse := &http.Response{
				Body: ioutil.NopCloser(
					strings.NewReader(hsResponseString)),
			}
			nspResponse := &http.Response{
				Body: ioutil.NopCloser(
					strings.NewReader(goodJoinResponse)),
			}
			responses := []*response{&response{
				resp: hsResponse,
			}, &response{
				resp: nspResponse,
			}}
			fdc := &fakeDoClient{requests, responses, 0}
			fc := &fakeConn{}
			fw := &fakeWsDialer{
				conn: fc,
			}
			trans, err := NewTransport(fdc, fw)
			trans.Close()
			_, err = trans.GetReadChannel()
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring,
				"Cannot read from a closed transport")
		})
		Convey("successfully write from multiple threads", func() {
			requests := make([]*http.Request, 0)
			// For the sake of this test, make the heartbeat long to
			// prevent from interferring.
			hsResponse := &http.Response{
				Body: ioutil.NopCloser(
					strings.NewReader(
						hsLongPingResponseString)),
			}
			nspResponse := &http.Response{
				Body: ioutil.NopCloser(
					strings.NewReader(goodJoinResponse)),
			}
			responses := []*response{&response{
				resp: hsResponse,
			}, &response{
				resp: nspResponse,
			}}
			fdc := &fakeDoClient{requests, responses, 0}
			fc := &fakeConn{
				closedChan: make(chan bool),
			}
			fw := &fakeWsDialer{
				conn: fc,
			}
			trans, err := NewTransport(fdc, fw)
			So(err, ShouldBeNil)
			var wg sync.WaitGroup
			for i := 10; i < 20; i++ {
				wg.Add(1)
				go func(i int) {
					defer wg.Done()
					trans.Write([]byte(strconv.Itoa(i)))
				}(i)
			}
			for i := 20; i < 30; i++ {
				wg.Add(1)
				go func(i int) {
					defer wg.Done()
					trans.Write([]byte(strconv.Itoa(i)))

				}(i)
			}
			wg.Wait()
			trans.Close()
			<-fc.closedChan

			fc.messagesMutex.Lock()
			So(fc.messagesWritten, ShouldHaveLength, 22)
			for i := 10; i < 30; i++ {
				So(fc.messagesWritten, ShouldContain,
					[]byte(strconv.Itoa(i)))
			}
			fc.messagesMutex.Unlock()
		})
	})
}
