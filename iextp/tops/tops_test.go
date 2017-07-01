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
