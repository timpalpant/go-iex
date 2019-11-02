package main

import (
	"encoding/json"
	"flag"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"

	"github.com/chilts/sid"
	"github.com/golang/glog"
	"github.com/gorilla/websocket"
)

type handshake struct {
	Sid string
}

func makeRequest(client *http.Client, method string,
	uri *url.URL, bodyData *string) []byte {
	glog.Infof("Making %s request:> %v", method, uri)

	var reader io.Reader
	if bodyData != nil {
		data := *bodyData
		glog.Infof("With data:> %s", data)
		reader = strings.NewReader(data)
	}
	req, _ := http.NewRequest(method, uri.String(), reader)
	resp, _ := client.Do(req)

	body, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	glog.Infof("Response:> %v", string(body))
	return body
}

func wsMessage(conn *websocket.Conn, msg []byte) {
	glog.Infof("Writing WS message:> %s", string(msg))
	conn.WriteMessage(websocket.TextMessage, msg)
}

func wsReadMessage(conn *websocket.Conn) {
	_, message, err := conn.ReadMessage()
	if err != nil {
		glog.Fatal(err)
	}
	glog.Infof("WS Response: %s", string(message))
}

func main() {
	flag.Parse()
	glog.Info("Starting handshake sequence")

	jar, _ := cookiejar.New(nil)
	client := &http.Client{Jar: jar}

	uri, _ := url.Parse("https://ws-api.iextrading.com/socket.io/")
	values := uri.Query()
	values.Set("t", sid.IdBase64())
	values.Set("EIO", "3")
	values.Set("transport", "polling")
	uri.RawQuery = values.Encode()

	resp := makeRequest(client, "GET", uri, nil)

	var hs handshake
	json.Unmarshal(resp[4:], &hs)
	values.Set("sid", hs.Sid)
	uri.RawQuery = values.Encode()

	makeRequest(client, "GET", uri, nil)

	uri, _ = url.Parse("wss://ws-api.iextrading.com/socket.io/")
	values.Set("transport", "websocket")
	uri.RawQuery = values.Encode()
	glog.Infof("Websocket connecting to:> %s", uri.String())
	conn, _, err := websocket.DefaultDialer.Dial(uri.String(), nil)
	if err != nil {
		glog.Fatal(err)
	}
	wsMessage(conn, []byte("5"))
	wsMessage(conn, []byte("2"))
	wsReadMessage(conn)
	wsMessage(conn, []byte("40/1.0/last,"))
	wsReadMessage(conn)
	wsMessage(conn, []byte("42/1.0/last,[\"subscribe\",\"fb,goog\"]"))
	wsReadMessage(conn)
	wsReadMessage(conn)
}
