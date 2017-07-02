package tops

import (
	"testing"
	"time"
)

func TestSystemEventMessage(t *testing.T) {
	data := []byte{
		0x53,                                           // S = System Event
		0x45,                                           // End of System Hours
		0x00, 0xa0, 0x99, 0x97, 0xe9, 0x3d, 0xb6, 0x14, // 2017-04-17 17:00:00
	}

	msg := SystemEventMessage{}
	if err := msg.Unmarshal(data); err != nil {
		t.Fatal(err)
	}

	expected := SystemEventMessage{
		MessageType: SystemEvent,
		SystemEvent: EndOfSystemHours,
		Timestamp:   time.Date(2017, time.April, 17, 17, 0, 0, 0, time.UTC),
	}

	if msg != expected {
		t.Fatalf("parsed: %v, expected: %v", msg, expected)
	}
}

func TestSecurityDirectoryMessage(t *testing.T) {
	data := []byte{
		0x44,                                           // D = Security Directory
		0x80,                                           // Test security, not an ETP, not a When Issued security
		0x00, 0x20, 0x89, 0x7b, 0x5a, 0x1f, 0xb6, 0x14, // 2017-04-17 07:40:00
		0x5a, 0x49, 0x45, 0x58, 0x54, 0x20, 0x20, 0x20, // ZIEXT
		0x64, 0x00, 0x00, 0x00, // 100 shares
		0x24, 0x1d, 0x0f, 0x00, 0x00, 0x00, 0x00, 0x00, // $99.05
		0x01, // Tier 1 NMS Stock
	}

	msg := SecurityDirectoryMessage{}
	if err := msg.Unmarshal(data); err != nil {
		t.Fatal(err)
	}

	expected := SecurityDirectoryMessage{
		MessageType:      SecurityDirectory,
		Flags:            0x80,
		Timestamp:        time.Date(2017, time.April, 17, 07, 40, 0, 0, time.UTC),
		Symbol:           "ZIEXT",
		RoundLotSize:     100,
		AdjustedPOCPrice: 99.05,
		LULDTier:         LULDTier1,
	}

	if msg != expected {
		t.Fatalf("parsed: %v, expected: %v", msg, expected)
	}

	if !msg.IsTestSecurity() {
		t.Error("message should be a test security")
	}
	if msg.IsETP() {
		t.Error("message should not be ETP")
	}
	if msg.IsWhenIssuedSecurity() {
		t.Error("message should not be a When Issued security")
	}
}

func TestTradingStatusMessage(t *testing.T) {
	data := []byte{
		0x48,                                           // H = Trading Status
		0x48,                                           // H = Trading Halted
		0xac, 0x63, 0xc0, 0x20, 0x96, 0x86, 0x6d, 0x14, // 2016-08-23 15:30:32.572715948
		0x5a, 0x49, 0x45, 0x58, 0x54, 0x20, 0x20, 0x20, // ZIEXT
		0x54, 0x31, 0x20, 0x20, // T1 = Halt News Pending
	}

	msg := TradingStatusMessage{}
	if err := msg.Unmarshal(data); err != nil {
		t.Fatal(err)
	}

	expected := TradingStatusMessage{
		MessageType:   TradingStatus,
		TradingStatus: TradingHalt,
		// NOTE: The TOPS specification says 2016-08-23 15:30:32.572715948,
		// but that is incorrect (probably not UTC).
		Timestamp: time.Date(2016, time.August, 23, 19, 30, 32, 572715948, time.UTC),
		Symbol:    "ZIEXT",
		Reason:    HaltNewsPending,
	}

	if msg != expected {
		t.Fatalf("parsed: %v, expected: %v", msg, expected)
	}
}

func TestOperationalHaltStatusMessage(t *testing.T) {
	data := []byte{
		0x4f,                                           // O = Operational Halt Status
		0x4f,                                           // O = Operationally halted on IEX
		0xac, 0x63, 0xc0, 0x20, 0x96, 0x86, 0x6d, 0x14, // 2016-08-23 15:30:32.572715948
		0x5a, 0x49, 0x45, 0x58, 0x54, 0x20, 0x20, 0x20, // ZIEXT
	}

	msg := OperationalHaltStatusMessage{}
	if err := msg.Unmarshal(data); err != nil {
		t.Fatal(err)
	}

	expected := OperationalHaltStatusMessage{
		MessageType:           OperationalHaltStatus,
		OperationalHaltStatus: IEXSpecificOperationalHalt,
		Timestamp:             time.Date(2016, time.August, 23, 19, 30, 32, 572715948, time.UTC),
		Symbol:                "ZIEXT",
	}

	if msg != expected {
		t.Fatalf("parsed: %v, expected: %v", msg, expected)
	}
}

func TestShortSalePriceTestStatusMessage(t *testing.T) {
	data := []byte{
		0x50,                                           // P = Short Sale Price Test Status
		0x01,                                           // Short Sale Price Test in effect
		0xac, 0x63, 0xc0, 0x20, 0x96, 0x86, 0x6d, 0x14, // 2016-08-23 15:30:32.572715948
		0x5a, 0x49, 0x45, 0x58, 0x54, 0x20, 0x20, 0x20, // ZIEXT
		0x41, // Activated
	}

	msg := ShortSalePriceTestStatusMessage{}
	if err := msg.Unmarshal(data); err != nil {
		t.Fatal(err)
	}

	expected := ShortSalePriceTestStatusMessage{
		MessageType:              ShortSalePriceTestStatus,
		ShortSalePriceTestStatus: true,
		Timestamp:                time.Date(2016, time.August, 23, 19, 30, 32, 572715948, time.UTC),
		Symbol:                   "ZIEXT",
		Detail:                   ShortSalePriceTestActivated,
	}

	if msg != expected {
		t.Fatalf("parsed: %v, expected: %v", msg, expected)
	}
}

func TestQuoteUpdateMessage(t *testing.T) {
	data := []byte{
		0x51,                                           // Q = Quote Update
		0x00,                                           // Active and regular market session
		0xac, 0x63, 0xc0, 0x20, 0x96, 0x86, 0x6d, 0x14, // 2016-08-23 15:30:32.572715948
		0x5a, 0x49, 0x45, 0x58, 0x54, 0x20, 0x20, 0x20, // ZIEXT
		0xe4, 0x25, 0x00, 0x00, // 9,700 shares
		0x24, 0x1d, 0x0f, 0x00, 0x00, 0x00, 0x00, 0x00, // $99.05
		0xec, 0x1d, 0x0f, 0x00, 0x00, 0x00, 0x00, 0x00, // $99.07
		0xe8, 0x03, 0x00, 0x00, // 1,000 shares
	}

	msg := QuoteUpdateMessage{}
	if err := msg.Unmarshal(data); err != nil {
		t.Fatal(err)
	}

	expected := QuoteUpdateMessage{
		MessageType: QuoteUpdate,
		Flags:       0,
		Timestamp:   time.Date(2016, time.August, 23, 19, 30, 32, 572715948, time.UTC),
		Symbol:      "ZIEXT",
		BidSize:     9700,
		BidPrice:    99.05,
		AskPrice:    99.07,
		AskSize:     1000,
	}

	if msg != expected {
		t.Fatalf("parsed: %v, expected: %v", msg, expected)
	}

	if !msg.IsActive() {
		t.Error("message flags should be active")
	}

	if !msg.IsRegularMarketSession() {
		t.Error("message flags should indicate regular market session")
	}
}

func TestTradeReportMessage(t *testing.T) {
	data := []byte{
		0x54,
		0x00,
		0xac, 0x63, 0xc0, 0x20, 0x96, 0x86, 0x6d, 0x14, // 2016-08-23 15:30:32.572715948
		0x5a, 0x49, 0x45, 0x58, 0x54, 0x20, 0x20, 0x20, // ZIEXT
		0x64, 0x00, 0x00, 0x00, // 100 shares
		0x24, 0x1d, 0x0f, 0x00, 0x00, 0x00, 0x00, 0x00, // $99.05
		0x96, 0x8f, 0x06, 0x00, 0x00, 0x00, 0x00, 0x00, // 429974
	}

	msg := TradeReportMessage{}
	if err := msg.Unmarshal(data); err != nil {
		t.Fatal(err)
	}

	expected := TradeReportMessage{
		MessageType:        TradeReport,
		SaleConditionFlags: 0,
		Timestamp:          time.Date(2016, time.August, 23, 19, 30, 32, 572715948, time.UTC),
		Symbol:             "ZIEXT",
		Size:               100,
		Price:              99.05,
		TradeID:            429974,
	}

	if msg != expected {
		t.Fatalf("parsed: %v, expected: %v", msg, expected)
	}

	if msg.IsISO() {
		t.Error("message should be non-ISO")
	}

	if msg.IsExtendedHoursTrade() {
		t.Error("message is a regular-hours trade")
	}

	if msg.IsOddLot() {
		t.Error("message is a regular or mixed lot")
	}

	if msg.IsTradeThroughExempt() {
		t.Error("message is trade-through exempt")
	}

	if msg.IsSinglePriceCrossTrade() {
		t.Error("message is not single-price cross trade")
	}

	if !msg.IsLastSaleEligible() {
		t.Error("message is last sale eligible")
	}

	if !msg.IsHighLowPriceEligible() {
		t.Error("message is high-low pice eligible")
	}

	if !msg.IsVolumeEligible() {
		t.Error("message is volume eligible")
	}
}

func TestTradeBreakMessage(t *testing.T) {
	data := []byte{
		0x42,                                           // B = Trade Break
		0x00,                                           // Non-ISO, Regular Market Session, Round or mixed lot, subject to Rule 611.
		0xb2, 0x8f, 0xa5, 0xa0, 0xab, 0x86, 0x6d, 0x14, // 2016-08-23 15:32:04.912754610
		0x5a, 0x49, 0x45, 0x58, 0x54, 0x20, 0x20, 0x20, // ZIEXT
		0x64, 0x00, 0x00, 0x00, // 100 shares
		0x24, 0x1d, 0x0f, 0x00, 0x00, 0x00, 0x00, 0x00, // $99.05
		0x96, 0x8f, 0x06, 0x00, 0x00, 0x00, 0x00, 0x00, // 429974
	}

	msg := TradeBreakMessage{}
	if err := msg.Unmarshal(data); err != nil {
		t.Fatal(err)
	}

	expected := TradeBreakMessage{
		MessageType:        TradeBreak,
		SaleConditionFlags: 0,
		Timestamp:          time.Date(2016, time.August, 23, 19, 32, 04, 912754610, time.UTC),
		Symbol:             "ZIEXT",
		Size:               100,
		Price:              99.05,
		TradeID:            429974,
	}

	if msg != expected {
		t.Fatalf("parsed: %v, expected: %v", msg, expected)
	}
}

func TestAuctionInformationMessage(t *testing.T) {
	data := []byte{
		0x41,                                           // A = Auction Information
		0x43,                                           // C = Closing Auction
		0xdd, 0xc7, 0xf0, 0x9a, 0x1a, 0x3a, 0xb6, 0x14, // 2017-04-17 15:50:12.462929885
		0x5a, 0x49, 0x45, 0x58, 0x54, 0x20, 0x20, 0x20, // ZIEXT
		// NOTE: The spec example says 100,000 shares, but this is not correct.
		// It's actually a 27,160 shares as a little endian 4-byte integer.
		0x18, 0x6a, 0x00, 0x00, // 100,000 shares
		0x24, 0x1d, 0x0f, 0x00, 0x00, 0x00, 0x00, 0x00, // $99.05
		0x18, 0x1f, 0x0f, 0x00, 0x00, 0x00, 0x00, 0x00, // $99.10
		// NOTE: The spec example says 10,000 shares, but this is not correct.
		// It's actually a 4,135 shares as a little endian 4-byte integer.
		0x27, 0x10, 0x00, 0x00, // 10,000 shares
		0x42,                   // B = buy-side imbalance
		0x00,                   // 0 extensions
		0x80, 0xe6, 0xf4, 0x58, // 2017-04-17 16:00:00
		0x0c, 0x21, 0x0f, 0x00, 0x00, 0x00, 0x00, 0x00, // $99.15
		0xc0, 0x1c, 0x0f, 0x00, 0x00, 0x00, 0x00, 0x00, // $99.04
		0xa4, 0x99, 0x0d, 0x00, 0x00, 0x00, 0x00, 0x00, // $89.13
		0xdc, 0x9f, 0x10, 0x00, 0x00, 0x00, 0x00, 0x00, // $108.95
	}

	msg := AuctionInformationMessage{}
	if err := msg.Unmarshal(data); err != nil {
		t.Fatal(err)
	}

	expected := AuctionInformationMessage{
		MessageType:              AuctionInformation,
		AuctionType:              ClosingAuction,
		Timestamp:                time.Date(2017, time.April, 17, 15, 50, 12, 462929885, time.UTC),
		Symbol:                   "ZIEXT",
		PairedShares:             27160,
		ReferencePrice:           99.05,
		IndicativeClearingPrice:  99.10,
		ImbalanceShares:          4135,
		ImbalanceSide:            BuySideImbalance,
		ExtensionNumber:          0,
		ScheduledAuctionTime:     time.Date(2017, time.April, 17, 16, 0, 0, 0, time.UTC),
		AuctionBookClearingPrice: 99.15,
		CollarReferencePrice:     99.04,
		LowerAuctionCollar:       89.13,
		UpperAuctionCollar:       108.95,
	}

	if msg != expected {
		t.Fatalf("parsed: %v, expected: %v", msg, expected)
	}
}
