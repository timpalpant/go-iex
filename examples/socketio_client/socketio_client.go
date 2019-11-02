package socketio

import (
	"flag"
	"fmt"
	"time"

	"github.com/timpalpant/go-iex"
	"github.com/timpalpant/go-iex/socketio"
)

func main() {
	flag.Parse()
	client := socketio.NewClient()
	ns := client.GetTOPSNamespace()
	go ns.SubscribeTo(func(msg iex.TOPS) {
		fmt.Printf("Received message: %+v\n", msg)
	}, "fb", "goog")
	time.Sleep(30 * time.Second)
}
