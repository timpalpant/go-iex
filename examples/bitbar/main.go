package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/johnmccabe/go-bitbar"
	"github.com/timpalpant/go-iex"
)

func main() {
	client := iex.NewClient(&http.Client{
		Timeout: 5 * time.Second,
	})

	symbols := []string{"AAPL", "FB"}
	quotes, err := client.GetLast(symbols)
	if err != nil {
		panic(err)
	}

	app := bitbar.New()
	for i := range quotes {
		app.StatusLine(fmt.Sprintf("%s: $%f", symbols[i], quotes[i].Price))
	}
	app.Render()
}
