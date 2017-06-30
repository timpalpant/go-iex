package iex

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/google/go-querystring/query"
)

const baseEndpoint = "https://api.iextrading.com/1.0"

// Client provides methods to interact with IEX's HTTP API for developers.
type Client struct {
	client *http.Client
}

func NewClient(client *http.Client) *Client {
	return &Client{client}
}

// TOPS provides IEX’s aggregated best quoted bid and offer
// position in near real time for all securities on IEX’s
// displayed limit order book. TOPS is ideal for developers
// needing both quote and trade data.
//
// Symbols may be any of the available symbols returned by
// GetSymbols(). If symbols is nil, then all symbols will be returned.
func (c *Client) GetTOPS(symbols []string) ([]*TOPS, error) {
	req := &topsRequest{symbols}
	var result []*TOPS
	err := c.getJSON("/tops", req, result)
	return result, err
}

type topsRequest struct {
	Symbols []string `url:"symbols,comma,omitempty"`
}

// Last provides trade data for executions on IEX.
// It is a near real time, intraday API that provides IEX last sale price,
// size and time. Last is ideal for developers that need a lightweight
// stock quote.
//
// Symbols may be any of the available symbols returned by
// GetSymbols(). If symbols is nil, then all symbols will be returned.
func (c *Client) GetLast(symbols []string) ([]*Last, error) {
	req := &lastRequest{symbols}
	var result []*Last
	err := c.getJSON("/tops/last", req, &result)
	return result, err
}

type lastRequest struct {
	Symbols []string `url:"symbols,comma,omitempty"`
}

// HIST will provide the output of IEX data products for download on
// a T+1 basis. Data will remain available for the trailing twelve months.
//
// If date is provided, then only data for that day will be returned.
// If date IsZero(), then the data available for all dates will be returned.
func (c *Client) GetHIST(date time.Time) ([]*HIST, error) {
	req := &histRequest{}
	if !date.IsZero() {
		req.Date = date.Format("20060102")
	}

	var result []*HIST
	err := c.getJSON("/hist", req, &result)
	return result, err
}

type histRequest struct {
	Date string `url:",omitempty"`
}

// DEEP is used to receive real-time depth of book quotations direct from IEX.
// The depth of book quotations received via DEEP provide an aggregated size
// of resting displayed orders at a price and side, and do not indicate the
// size or number of individual orders at any price level. Non-displayed
// orders and non-displayed portions of reserve orders are not represented
// in DEEP.
//
// DEEP also provides last trade price and size information.
// Trades resulting from either displayed or non-displayed orders
// matching on IEX will be reported. Routed executions will not be reported.
func (c *Client) GetDEEP(symbol string) (*DEEP, error) {
	req := &deepRequest{symbol}
	result := &DEEP{}
	err := c.getJSON("/deep", req, &result)
	return result, err
}

type deepRequest struct {
	Symbols string
}

// Book shows IEX’s bids and asks for given symbols.
//
// A maximumum of 10 symbols may be requested.
func (c *Client) GetBook(symbols []string) (map[string]*Book, error) {
	req := &bookRequest{symbols}
	var result map[string]*Book
	err := c.getJSON("/deep/book", req, &result)
	return result, err
}

type bookRequest struct {
	Symbols []string `url:"symbols,comma,omitempty"`
}

// Trade report messages are sent when an order on the IEX Order Book is
// executed in whole or in part. DEEP sends a Trade report message for
// every individual fill.
//
// A maximum of 10 symbols may be requested. Last is the number of trades
// to fetch, and must be <= 500.
func (c *Client) GetTrades(symbols []string, last int) (map[string][]*Trade, error) {
	req := &tradesRequest{symbols, last}
	var result map[string][]*Trade
	err := c.getJSON("/deep/trades", req, &result)
	return result, err
}

type tradesRequest struct {
	Symbols []string `url:"symbols,comma,omitempty"`
	Last    int      `url:",omitempty"`
}

// The System event message is used to indicate events that apply to
// the market or the data feed.
//
// There will be a single message disseminated per channel for each
// System Event type within a given trading session.
//
// A maximumum of 10 symbols may be requested.
func (c *Client) GetSystemEvents(symbols []string) (map[string]*SystemEvent, error) {
	req := &systemEventRequest{symbols}
	var result map[string]*SystemEvent
	err := c.getJSON("/deep/system-event", req, &result)
	return result, err
}

type systemEventRequest struct {
	Symbols []string `url:"symbols,comma,omitempty"`
}

// The Trading status message is used to indicate the current trading status
// of a security. For IEX-listed securities, IEX acts as the primary market
// and has the authority to institute a trading halt or trading pause in a
// security due to news dissemination or regulatory reasons. For
// non-IEX-listed securities, IEX abides by any regulatory trading halts
// and trading pauses instituted by the primary or listing market, as
// applicable.
//
// IEX disseminates a full pre-market spin of Trading status messages
// indicating the trading status of all securities. In the spin, IEX will
// send out a Trading status message with “T” (Trading) for all securities
// that are eligible for trading at the start of the Pre-Market Session.
// If a security is absent from the dissemination, firms should assume
// that the security is being treated as operationally halted in the IEX
// Trading System.
//
// After the pre-market spin, IEX will use the Trading status message to
// relay changes in trading status for an individual security. Messages
// will be sent when a security is:
//
//     Halted
//     Paused*
//     Released into an Order Acceptance Period*
//     Released for trading
//
// *The paused and released into an Order Acceptance Period status will be disseminated for IEX-listed securities only. Trading pauses on non-IEX-listed securities will be treated simply as a halt.
//
// A maximumum of 10 symbols may be requested.
func (c *Client) GetTradingStatus(symbols []string) (map[string]*TradingStatusMessage, error) {
	req := &tradingStatusRequest{symbols}
	var result map[string]*TradingStatusMessage
	err := c.getJSON("/deep/trading-status", req, &result)
	return result, err
}

type tradingStatusRequest struct {
	Symbols []string `url:"symbols,comma,omitempty"`
}

// The Exchange may suspend trading of one or more securities on IEX
// for operational reasons and indicates such operational halt using
// the Operational halt status message.
//
// IEX disseminates a full pre-market spin of Operational halt status
// messages indicating the operational halt status of all securities.
// In the spin, IEX will send out an Operational Halt Message with “N”
// (Not operationally halted on IEX) for all securities that are
// eligible for trading at the start of the Pre-Market Session. If a
// security is absent from the dissemination, firms should assume that
// the security is being treated as operationally halted in the IEX
// Trading System at the start of the Pre-Market Session.
//
// After the pre-market spin, IEX will use the Operational halt status
// message to relay changes in operational halt status for an
// individual security.
//
// A maximumum of 10 symbols may be requested.
func (c *Client) GetOperationalHaltStatus(symbols []string) (map[string]*OpHaltStatus, error) {
	req := &opHaltStatusRequest{symbols}
	var result map[string]*OpHaltStatus
	err := c.getJSON("/deep/op-halt-status", req, &result)
	return result, err
}

type opHaltStatusRequest struct {
	Symbols []string `url:"symbols,comma,omitempty"`
}

// In association with Rule 201 of Regulation SHO, the Short Sale
// Price Test Message is used to indicate when a short sale price
// test restriction is in effect for a security.
//
// IEX disseminates a full pre-market spin of Short sale price test
// status messages indicating the Rule 201 status of all securities.
// After the pre-market spin, IEX will use the Short sale price test
// status message in the event of an intraday status change.
//
// The IEX Trading System will process orders based on the latest
// short sale price test restriction status.
//
// A maximumum of 10 symbols may be requested.
func (c *Client) GetShortSaleRestriction(symbols []string) (map[string]*SSRStatus, error) {
	req := &ssrStatusRequest{symbols}
	var result map[string]*SSRStatus
	err := c.getJSON("/deep/ssr-status", req, &result)
	return result, err
}

type ssrStatusRequest struct {
	Symbols []string `url:"symbols,comma,omitempty"`
}

// The Security event message is used to indicate events that
// apply to a security. A Security event message will be sent
// whenever such event occurs.
//
// A maximumum of 10 symbols may be requested.
func (c *Client) GetSecurityEvents(symbols []string) (map[string]*SecurityEventMessage, error) {
	req := &securityEventRequest{symbols}
	var result map[string]*SecurityEventMessage
	err := c.getJSON("/deep/security-event", req, &result)
	return result, err
}

type securityEventRequest struct {
	Symbols []string `url:"symbols,comma,omitempty"`
}

// Trade break messages are sent when an execution on IEX is broken
// on that same trading day. Trade breaks are rare and only affect
// applications that rely upon IEX execution based data.
//
// A maximum of 10 symbols may be requested. Last is the number of trades
// to fetch, and must be <= 500.
func (c *Client) GetTradeBreaks(symbols []string, last int) (map[string][]*TradeBreak, error) {
	req := &tradeBreaksRequest{symbols, last}
	var result map[string][]*TradeBreak
	err := c.getJSON("/deep/trade-breaks", req, &result)
	return result, err
}

type tradeBreaksRequest struct {
	Symbols []string `url:"symbols,comma,omitempty"`
	Last    int      `url:",omitempty"`
}

// This endpoint returns near real time traded volume on the markets.
// Market data is captured by the IEX system from approximately
// 7:45 a.m. to 5:15 p.m. ET.
func (c *Client) GetMarkets() ([]*Market, error) {
	var result []*Market
	err := c.getJSON("/market", nil, &result)
	return result, err
}

// GetSymbols returns an array of symbols IEX supports for trading.
// This list is updated daily as of 7:45 a.m. ET. Symbols may be added
// or removed by IEX after the list was produced.
func (c *Client) GetSymbols() ([]*Symbol, error) {
	var result []*Symbol
	err := c.getJSON("/ref-data/symbols", nil, &result)
	return result, err
}

func (c *Client) getJSON(route string, request interface{}, response interface{}) error {
	url := c.endpoint(route)

	values, err := query.Values(request)
	if err != nil {
		return err
	}
	queryString := values.Encode()
	if queryString != "" {
		url = url + "?" + queryString
	}

	resp, err := c.client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("%v: %v", resp.Status, string(body))
	}

	dec := json.NewDecoder(resp.Body)
	return dec.Decode(response)
}

func (c *Client) endpoint(route string) string {
	return baseEndpoint + route
}
