# go-iex
A Go library for accessing the IEX Developer API.

[![Build Status](https://travis-ci.org/timpalpant/go-iex.svg?branch=master)](https://travis-ci.org/timpalpant/go-iex)
[![Coverage Status](https://coveralls.io/repos/timpalpant/go-iex/badge.svg?branch=master&service=github)](https://coveralls.io/github/timpalpant/go-iex?branch=master)

go-iex is a library to access the [IEX Developer API](https://www.iextrading.com/developer/docs/) from [Go](http://www.golang.org).
It provides a thin wrapper for working with the JSON REST endpoints and [IEXTP1 pcap data](https://www.iextrading.com/trading/market-data/#specifications).

[IEX](https://www.iextrading.com) is a fair, simple and transparent stock exchange dedicated to investor protection.
IEX provides realtime and historical market data for free through the IEX Developer API.
By using the IEX API, you agree to the [Terms of Use](https://www.iextrading.com/api-terms/). IEX is not affiliated
and does not endorse or recommend this library.

## Usage

```Go
package main

import (
  "fmt"
  "net/http"

  "github.com/timpalpant/go-iex"
)

func main() {
  client := iex.NewClient(&http.Client{})
	
  availableSymbols, err := client.GetSymbols()
  if err != nil {
      panic(err)
  }
  
  quotes, err := client.GetTOPS(availableSymbols[:5])
  if err != nil {
      panic(err)
  }
  
  for _, quote := range quotes {
      fmt.Fprintf("%v: bid $%.02f (%v shares), ask $%.02f (%v shares) [as of %v]",
          quote.Symbol, quote.BidPrice, quote.BidSize,
          quote.AskPrice, quote.AskSize, quote.LastUpdated)
  }
}
```

## Contributing

Pull requests and issues are welcomed!

## License

go-iex is released under the [GNU Lesser General Public License, Version 3.0](https://www.gnu.org/licenses/lgpl-3.0.en.html)
