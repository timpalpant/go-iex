package deep

import (
	"reflect"
	"testing"
	"time"

	"github.com/timpalpant/go-iex/iextp"
)

func TestUnmarshal_UnknownMessageType(t *testing.T) {
	data := []byte{0x02} // Not a known message type.
	msg, err := Unmarshal(data)
	if err != nil {
		t.Fatal(err)
	}

	unkMsg, ok := msg.(*iextp.UnsupportedMessage)
	if !ok {
		t.Fatal("expected to decode UnsupportedMessage")
	}

	if !reflect.DeepEqual(unkMsg.Message, data) {
		t.Fatal("message data not equal to input")
	}
}

func TestUnmarshal_Empty(t *testing.T) {
	data := []byte{}
	_, err := Unmarshal(data)
	if err.Error() != "cannot unmarshal 0-length buffer" {
		t.Fatal("expected unmarshal error")
	}
}

func TestSecurityEventMessage(t *testing.T) {
	data := []byte{
		0x45,                                           // E = Security Event
		0x4f,                                           // O = Opening Process Complete
		0x00, 0xf0, 0x30, 0x2a, 0x5b, 0x25, 0xb6, 0x14, // 2017-04-17 09:30:00
		0x5a, 0x49, 0x45, 0x58, 0x54, 0x20, 0x20, 0x20, // ZIEXT
	}

	msg, err := Unmarshal(data)
	if err != nil {
		t.Fatal(err)
	}

	seMsg := *msg.(*SecurityEventMessage)
	expected := SecurityEventMessage{
		MessageType:   SecurityEvent,
		SecurityEvent: OpeningProcessComplete,
		Timestamp:     time.Date(2017, time.April, 17, 9, 30, 0, 0, time.UTC),
		Symbol:        "ZIEXT",
	}

	if seMsg != expected {
		t.Fatalf("parsed: %v, expected: %v", msg, expected)
	}
}

func TestPriceLevelUpdateMessage_BuySide(t *testing.T) {
	data := []byte{
		0x38, // Price level update on the Buy Side
		0x01, // Event processing complete
		// NOTE: The spec document says 15:30:32, but this is actually 19:30:32 UTC.
		0xac, 0x63, 0xc0, 0x20, 0x96, 0x86, 0x6d, 0x14, // 2016-08-23 15:30:32.572715948
		0x5a, 0x49, 0x45, 0x58, 0x54, 0x20, 0x20, 0x20, // ZIEXT
		0xe4, 0x25, 0x00, 0x00, // 9,700 shares
		0x24, 0x1d, 0x0f, 0x00, 0x00, 0x00, 0x00, 0x00, // $99.05
	}

	msg, err := Unmarshal(data)
	if err != nil {
		t.Fatal(err)
	}

	pluMsg := *msg.(*PriceLevelUpdateMessage)
	expected := PriceLevelUpdateMessage{
		MessageType: PriceLevelUpdateBuySide,
		EventFlags:  1,
		Timestamp:   time.Date(2016, time.August, 23, 19, 30, 32, 572715948, time.UTC),
		Symbol:      "ZIEXT",
		Size:        9700,
		Price:       99.05,
	}

	if pluMsg != expected {
		t.Fatalf("parsed: %v, expected: %v", msg, expected)
	}

	if !pluMsg.IsBuySide() {
		t.Fatal("message is buy side")
	}

	if pluMsg.IsSellSide() {
		t.Fatal("message is buy side")
	}
}

func TestPriceLevelUpdateMessage_SellSide(t *testing.T) {
	data := []byte{
		0x35,                                           // Price level update on the Sell Side
		0x01,                                           // Event processing complete
		0xac, 0x63, 0xc0, 0x20, 0x96, 0x86, 0x6d, 0x14, // 2016-08-23 15:30:32.572715948
		0x5a, 0x49, 0x45, 0x58, 0x54, 0x20, 0x20, 0x20, // ZIEXT
		0xe4, 0x25, 0x00, 0x00, // 9,700 shares
		0x24, 0x1d, 0x0f, 0x00, 0x00, 0x00, 0x00, 0x00, // $99.05
	}

	msg, err := Unmarshal(data)
	if err != nil {
		t.Fatal(err)
	}

	pluMsg := *msg.(*PriceLevelUpdateMessage)
	expected := PriceLevelUpdateMessage{
		MessageType: PriceLevelUpdateSellSide,
		EventFlags:  1,
		Timestamp:   time.Date(2016, time.August, 23, 19, 30, 32, 572715948, time.UTC),
		Symbol:      "ZIEXT",
		Size:        9700,
		Price:       99.05,
	}

	if pluMsg != expected {
		t.Fatalf("parsed: %v, expected: %v", msg, expected)
	}

	if pluMsg.IsBuySide() {
		t.Fatal("message is sell side")
	}

	if !pluMsg.IsSellSide() {
		t.Fatal("message is sell side")
	}
}
