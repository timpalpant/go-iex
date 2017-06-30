package iex

import (
	"encoding/json"
)

type Feed string

func (f *Feed) UnmarshalJSON(b []byte) error {
	var s string
	err := json.Unmarshal(b, &s)
	*f = Feed(s)
	return err
}

func (f *Feed) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(*f))
}

const (
	FeedDEEP Feed = "DEEP"
	FeedTOPS Feed = "TOPS"
)

type Protocol string

func (p *Protocol) UnmarshalJSON(b []byte) error {
	var s string
	err := json.Unmarshal(b, &s)
	*p = Protocol(s)
	return err
}

func (p *Protocol) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(*p))
}

const IEXTP1 Protocol = "IEXTP1"

type SystemEventCode string

func (sec *SystemEventCode) UnmarshalJSON(b []byte) error {
	var s string
	err := json.Unmarshal(b, &s)
	*sec = SystemEventCode(s)
	return err
}

func (sec *SystemEventCode) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(*sec))
}

const (
	StartMessages    SystemEventCode = "O"
	StartSystemHours SystemEventCode = "S"
	StartMarketHours SystemEventCode = "R"
	EndMarketHours   SystemEventCode = "M"
	EndSystemHours   SystemEventCode = "E"
	EndMessages      SystemEventCode = "C"
)

type TradingStatus string

func (ts *TradingStatus) UnmarshalJSON(b []byte) error {
	var s string
	err := json.Unmarshal(b, &s)
	*ts = TradingStatus(s)
	return err
}

func (ts *TradingStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(*ts))
}

const (
	// Trading halted across all US equity markets.
	TradingHalted TradingStatus = "H"
	// Trading halt released into an Order Acceptance Period
	// (IEX-listed securities only)
	TradingOrderAcceptancePeriod = "O"
	// Trading paused and Order Acceptance Period on IEX
	// (IEX-listed securities only)
	TradingPaused = "P"
	// Trading on IEX
	Trading = "T"
)

type TradingReason string

func (tr *TradingReason) UnmarshalJSON(b []byte) error {
	var s string
	err := json.Unmarshal(b, &s)
	*tr = TradingReason(s)
	return err
}

func (tr *TradingReason) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(*tr))
}

const (
	// Trading halt reasons.
	HaltNewsPending            = "T1"
	IPOIssueNotYetTrading      = "IPO1"
	IPOIssueDeferred           = "IPOD"
	MarketCircuitBreakerLevel3 = "MCB3"
	ReasonNotAvailable         = "NA"

	// Order Acceptance Period Reasons
	HaltNewsDisseminations           = "T2"
	IPONewIssueOrderAcceptancePeriod = "IPO2"
	IPOPreLaunchPeriod               = "IPO3"
	MarketCircuitBreakerLevel1       = "MCB1"
	MarketCircuitBreakerLevel2       = "MCB2"
)

type SecurityEvent string

func (se *SecurityEvent) UnmarshalJSON(b []byte) error {
	var s string
	err := json.Unmarshal(b, &s)
	*se = SecurityEvent(s)
	return err
}

func (se *SecurityEvent) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(*se))
}

const (
	MarketOpen  SecurityEvent = "MarketOpen"
	MarketClose SecurityEvent = "MarketClose"
)

type TOPS struct {
	// Refers to the stock ticker.
	Symbol string
	// Refers to IEX’s percentage of the market in the stock.
	MarketPercent float64
	// Refers to amount of shares on the bid on IEX.
	BidSize int
	// Refers to the best bid price on IEX.
	BidPrice float64
	// Refers to amount of shares on the ask on IEX.
	AskSize int
	// Refers to the best ask price on IEX.
	AskPrice float64
	// Refers to shares traded in the stock on IEX.
	Volume int
	// Refers to last sale price of the stock on IEX. (Refer to the attribution section above.)
	LastSalePrice float64
	// Refers to last sale size of the stock on IEX.
	LastSaleSize int
	// Refers to last sale time of the stock on IEX.
	LastSaleTime Time
	// Refers to the last update time of the data.
	// If the value is the zero Time, IEX has not quoted the symbol in
	// the trading day.
	LastUpdated Time
}

type Last struct {
	// Refers to the stock ticker.
	Symbol string
	// Refers to last sale price of the stock on IEX. (Refer to the attribution section above.)
	Price float64
	// Refers to last sale size of the stock on IEX.
	Size int
	// Refers to last sale time in epoch time of the stock on IEX.
	Time Time
}

type HIST struct {
	// URL to the available data file.
	Link string
	// Date of the data contained in this file.
	Date string
	// Which data feed is contained in this file.
	Feed Feed
	// The feed format specification version.
	Version string
	// The protocol version of the data.
	Protocol Protocol
	// The size, in bytes, of the data file.
	Size int64 `json:",string"`
}

type DEEP struct {
	Symbol        string
	MarketPercent float64
	Volume        int
	LastSalePrice float64
	LastSaleSize  int
	LastSaleTime  Time
	LastUpdate    Time
	Bids          []*Quote
	Asks          []*Quote
	SystemEvent   *SystemEvent
	TradingStatus *TradingStatusMessage
	OpHaltStatus  *OpHaltStatus
	SSRStatus     *SSRStatus
	SecurityEvent *SecurityEvent
	Trades        []*Trade
	TradeBreaks   []*TradeBreak
}

type Quote struct {
	Price     float64
	Size      float64
	Timestamp Time
}

type SystemEvent struct {
	SystemEvent SystemEventCode
	Timestamp   Time
}

type TradingStatusMessage struct {
	Status    TradingStatus
	Reason    TradingReason
	Timestamp Time
}

type OpHaltStatus struct {
	IsHalted  bool
	Timestamp Time
}

type SSRStatus struct {
	IsSSR     bool
	Detail    string
	Timestamp Time
}

type SecurityEventMessage struct {
	SecurityEvent SecurityEvent
	Timestamp     Time
}

type Trade struct {
	Price                 float64
	Size                  int
	TradeID               int64
	IsISO                 bool
	IsOddLot              bool
	IsOutsideRegularHours bool
	IsSinglePriceCross    bool
	IsTradeThroughExcempt bool
	Timestamp             Time
}

type TradeBreak struct {
	Price                 float64
	Size                  int
	TradeID               int64
	IsISO                 bool
	IsOddLot              bool
	IsOutsideRegularHours bool
	IsSinglePriceCross    bool
	IsTradeThroughExcempt bool
	Timestamp             Time
}

type Book struct {
	Bids []*Quote
	Asks []*Quote
}

type Market struct {
	// Refers to the Market Identifier Code (MIC).
	MIC string
	// Refers to the tape id of the venue.
	TapeID string
	// Refers to name of the venue defined by IEX.
	VenueName string
	// Refers to the amount of traded shares reported by the venue.
	Volume int
	// Refers to the amount of Tape A traded shares reported by the venue.
	TapeA int
	// Refers to the amount of Tape B traded shares reported by the venue.
	TapeB int
	// Refers to the amount of Tape C traded shares reported by the venue.
	TapeC int
	// Refers to the venue’s percentage of shares traded in the market.
	MarketPercent float64
	// Refers to the last update time of the data.
	LastUpdated Time
}

type Symbol struct {
	// Refers to the symbol represented in Nasdaq Integrated symbology (INET).
	Ticker string
	// Refers to the name of the company or security.
	Name string
	// Refers to the date the symbol reference data was generated.
	Date string
	// Will be true if the symbol is enabled for trading on IEX.
	IsEnabled bool
}

type IntradayStats struct {
	// Refers to single counted shares matched from executions on IEX.
	Volume struct {
		Value       int
		LastUpdated Time
	}
	// Refers to number of symbols traded on IEX.
	SymbolsTraded struct {
		Value       int
		LastUpdated Time
	}
	// Refers to executions received from order routed to away trading centers.
	RoutedVolume struct {
		Value       int
		LastUpdated Time
	}
	// Refers to sum of matched volume times execution price of those trades.
	Notional struct {
		Value       int
		LastUpdated Time
	}
	// Refers to IEX’s percentage of total US Equity market volume.
	MarketShare struct {
		Value       float64
		LastUpdated Time
	}
}

type Stats struct {
	// Refers to the trading day.
	Date string
	// Refers to executions received from order routed to away trading centers.
	Volume int
	// Refers to single counted shares matched from executions on IEX.
	RoutedVolume int
	// Refers to IEX’s percentage of total US Equity market volume.
	MarketShare float64
	// Will be true if the trading day is a half day.
	IsHalfDay bool
	// Refers to the number of lit shares traded on IEX (single-counted).
	LitVolume int
}

type Records struct {
	// Refers to single counted shares matched from executions on IEX.
	Volume *Record
	// Refers to number of symbols traded on IEX.
	SymbolsTraded *Record
	// Refers to executions received from order routed to away trading centers.
	RoutedVolume *Record
	// Refers to sum of matched volume times execution price of those trades.
	Notional *Record
}

type Record struct {
	Value            int    `json:"recordValue"`
	Date             string `json:"recordDate"`
	PreviousDayValue int
	Avg30Value       float64
}

type HistoricalSummary struct {
	AverageDailyVolume          float64
	AverageDailyRoutedVolume    float64
	AverageMarketShare          float64
	AverageOrderSize            float64
	AverageFillSize             float64
	Bin100Percent               float64
	Bin101Percent               float64
	Bin200Percent               float64
	Bin300Percent               float64
	Bin400Percent               float64
	Bin500Percent               float64
	Bin1000Percent              float64
	Bin5000Percent              float64
	Bin10000Percent             float64
	Bin10000Trades              float64
	Bin20000Trades              float64
	Bin50000Trades              float64
	UniqueSymbolsTraded         float64
	BlockPercent                float64
	SelfCrossPercent            float64
	ETFPercent                  float64
	LargeCapPercent             float64
	MidCapPercent               float64
	SmallCapPercent             float64
	VenueARCXFirstWaveWeight    float64
	VenueBATSFirstWaveWeight    float64
	VenueBATYFirstWaveWeight    float64
	VenueEDGAFirstWaveWeight    float64
	VenueEDGXFirstWaveWeight    float64
	VenueOverallFirstWaveWeight float64
	VenueXASEFirstWaveWeight    float64
	VenueXBOSFirstWaveWeight    float64
	VenueXCHIFirstWaveWeight    float64
	VenueXCISFirstWaveWeight    float64
	VenueXNGSFirstWaveWeight    float64
	VenueXNYSFirstWaveWeight    float64
	VenueXPHLFirstWaveWeight    float64
	VenueARCXFirstWaveRate      float64
	VenueBATSFirstWaveRate      float64
	VenueBATYFirstWaveRate      float64
	VenueEDGAFirstWaveRate      float64
	VenueEDGXFirstWaveRate      float64
	VenueOverallFirstWaveRate   float64
	VenueXASEFirstWaveRate      float64
	VenueXBOSFirstWaveRate      float64
	VenueXCHIFirstWaveRate      float64
	VenueXCISFirstWaveRate      float64
	VenueXNGSFirstWaveRate      float64
	VenueXNYSFirstWaveRate      float64
	VenueXPHLFirstWaveRate      float64
}
