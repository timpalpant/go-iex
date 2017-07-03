package iextp

import (
	"os"
	"testing"
	"time"
)

var header = []byte{
	0x01,       // Version: 1
	0x00,       // (Reserved)
	0x04, 0x80, // DEEP v1.0
	0x01, 0x00, 0x00, 0x00, // Channel: 1
	0x00, 0x00, 0x87, 0x42, // Today's current Session ID
	0x48, 0x00, // 72 bytes
	0x02, 0x00, // 2 messages
	0x8c, 0xa6, 0x21, 0x00, 0x00, 0x00, 0x00, 0x00, // 2,205,324 bytes
	0xca, 0x3, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // Sequence: 50122
	0xec, 0x45, 0xc2, 0x20, 0x96, 0x86, 0x6d, 0x14, // 2016-08-23 15:30:32.572839404
}

var payload = []byte{
	// Message Block 1
	0x26, 0x00, // 38 bytes
	0x54, // T = Trade Report
	0x00,
	0xac, 0x63, 0xc0, 0x20, 0x96, 0x86, 0x6d, 0x14, // 2016-08-23 15:30:32.572715948
	0x5a, 0x49, 0x45, 0x58, 0x54, 0x20, 0x20, 0x20, // ZIEXT
	0x64, 0x00, 0x00, 0x00, // 100 shares
	0x24, 0x1d, 0x0f, 0x00, 0x00, 0x00, 0x00, 0x00, // $99.05
	0x96, 0x8f, 0x06, 0x00, 0x00, 0x00, 0x00, 0x00, // 429974

	// Message Block 2
	0x1e, 0x00, // 30 bytes
	0x38, // Price level update on the Buy Side
	0x01, // Event processing complete
	// NOTE: The spec document says 15:30:32, but this is actually 19:30:32 UTC.
	0xac, 0x63, 0xc0, 0x20, 0x96, 0x86, 0x6d, 0x14, // 2016-08-23 15:30:32.572715948
	0x5a, 0x49, 0x45, 0x58, 0x54, 0x20, 0x20, 0x20, // ZIEXT
	0xe4, 0x25, 0x00, 0x00, // 9,700 shares
	0x24, 0x1d, 0x0f, 0x00, 0x00, 0x00, 0x00, 0x00, // $99.05
}

// testUnmarshal is unmarshals all messages into UnsupportedMessage,
// a simulated higher-level protocol for testing IEX-TP.
func testUnmarshal(buf []byte) (Message, error) {
	msg := &UnsupportedMessage{}
	err := msg.Unmarshal(buf)
	return msg, err
}

func TestMain(m *testing.M) {
	RegisterProtocol(0x8004, testUnmarshal)
	os.Exit(m.Run())
}

func TestUnmarshalSegmentHeader(t *testing.T) {
	h := SegmentHeader{}
	if err := h.Unmarshal(header); err != nil {
		t.Fatal(err)
	}

	expected := SegmentHeader{
		Version:                    1,
		MessageProtocolID:          0x8004,
		ChannelID:                  1,
		SessionID:                  1116143616,
		PayloadLength:              72,
		MessageCount:               2,
		StreamOffset:               2205324,
		FirstMessageSequenceNumber: 970,
		SendTime:                   time.Date(2016, time.August, 23, 19, 30, 32, 572839404, time.UTC),
	}

	if h != expected {
		t.Fatalf("parsed: %v, expected: %v", h, expected)
	}
}

func TestUnmarshalSegment(t *testing.T) {
	var data []byte
	data = append(data, header...)
	data = append(data, payload...)

	var segment Segment
	if err := segment.Unmarshal(data); err != nil {
		t.Fatal(err)
	}

	if len(segment.Messages) != 2 {
		t.Fatalf("should have unmarshaled 2 messages, got %v", len(segment.Messages))
	}
}

func TestUnmarshalSegment_UnknownProtocol(t *testing.T) {
	data := []byte{
		0x01,       // Version: 1
		0x00,       // (Reserved)
		0x10, 0x10, // Unknown protocol
		0x01, 0x00, 0x00, 0x00, // Channel: 1
		0x00, 0x00, 0x87, 0x42, // Today's current Session ID
		0x00, 0x00, // 0 bytes
		0x00, 0x00, // 0 messages
		0x8c, 0xa6, 0x21, 0x00, 0x00, 0x00, 0x00, 0x00, // 2,205,324 bytes
		0xca, 0x3, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // Sequence: 50122
		0xec, 0x45, 0xc2, 0x20, 0x96, 0x86, 0x6d, 0x14, // 2016-08-23 15:30:32.572839404
	}

	var segment Segment
	if err := segment.Unmarshal(data); err == nil {
		t.Fatal("expected unknown protocol")
	} else if err.Error() != "unknown message protocol: 4112" {
		t.Fatal(err)
	}
}

func TestUnmarshalSegment_Empty(t *testing.T) {
	data := []byte{}

	var segment Segment
	if err := segment.Unmarshal(data); err == nil {
		t.Fatal("expected error")
	} else if err.Error() != "cannot unmarshal SegmentHeader from 0-length buffer" {
		t.Fatal(err)
	}
}

func TestUnmarshalSegment_TooShort(t *testing.T) {
	data := []byte{
		0x01,       // Version: 1
		0x00,       // (Reserved)
		0x04, 0x80, // DEEP v1.0
	}

	var segment Segment
	if err := segment.Unmarshal(data); err == nil {
		t.Fatal("expected error")
	} else if err.Error() != "cannot unmarshal SegmentHeader from 4-length buffer" {
		t.Fatal(err)
	}
}

func TestUnmarshalSegment_NoMessages(t *testing.T) {
	data := []byte{
		0x01,       // Version: 1
		0x00,       // (Reserved)
		0x04, 0x80, // DEEP v1.0
		0x01, 0x00, 0x00, 0x00, // Channel: 1
		0x00, 0x00, 0x87, 0x42, // Today's current Session ID
		0x00, 0x00, // 0 bytes
		0x00, 0x00, // 0 messages
		0x8c, 0xa6, 0x21, 0x00, 0x00, 0x00, 0x00, 0x00, // 2,205,324 bytes
		0xca, 0x3, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // Sequence: 50122
		0xec, 0x45, 0xc2, 0x20, 0x96, 0x86, 0x6d, 0x14, // 2016-08-23 15:30:32.572839404
	}

	var segment Segment
	if err := segment.Unmarshal(data); err != nil {
		t.Fatal(err)
	}

	if len(segment.Messages) != 0 {
		t.Fatal("should have unmarshaled 0 messages")
	}
}
