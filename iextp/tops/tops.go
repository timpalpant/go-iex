package tops

import (
	"encoding/binary"
	"fmt"
	"strings"
	"time"

	"github.com/timpalpant/go-iex/iextp"
)

const (
	ChannelID         uint32 = 1
	MessageProtocolID uint16 = 0x8003
)

const (
	MessageTypeSystemEvent       uint8 = 0x53
	MessageTypeSecurityDirectory uint8 = 0x44
	MessageTypeTradingStatus     uint8 = 0x48
)

// Protocol implements the TOPS protocol, v1.6.
type Protocol struct{}

func (p Protocol) ID() uint16 {
	return MessageProtocolID
}

func (p Protocol) Unmarshal(buf []byte) (iextp.Message, error) {
	if len(buf) == 0 {
		return nil, fmt.Errorf("cannot unmarshal %v-length buffer", len(buf))
	}

	var msg iextp.Message

	messageType := uint8(buf[0])
	switch messageType {
	case MessageTypeSystemEvent:
		msg = &SystemEventMessage{}
	case MessageTypeSecurityDirectory:
		msg = &SecurityDirectoryMessage{}
	case MessageTypeTradingStatus:
		msg = &TradingStatusMessage{}
	default:
		msg = &iextp.UnsupportedMessage{}
	}

	err := msg.Unmarshal(buf)
	return msg, err
}

// Parse the TOPS timestamp type: 8 bytes, signed integer containing
// a counter of nanoseconds since POSIX (Epoch) time UTC,
// into a native time.Time.
func parseTimestamp(buf []byte) time.Time {
	timestampNs := int64(binary.LittleEndian.Uint64(buf))
	return time.Unix(0, timestampNs).In(time.UTC)
}

// Parse the TOPS price type: 8 bytes, signed integer containing
// a fixed-point number with 4 digits to the right of an implied
// decimal point, into a float64.
func parseFloat(buf []byte) float64 {
	n := int64(binary.LittleEndian.Uint64(buf))
	return float64(n) / 10000
}

// Parse the TOPS string type: fixed-length ASCII byte sequence,
// left justified and space filled on the right.
func parseString(buf []byte) string {
	return strings.TrimRight(string(buf), " ")
}

// SystemEventMessage is used to indicate events that apply
// to the market or the data feed.
//
// There will be a single message disseminated per channel for each
// System Event type within a given trading session.
type SystemEventMessage struct {
	// System event identifier.
	SystemEvent uint8
	// Time stamp of the system event.
	Timestamp time.Time
}

func (m *SystemEventMessage) Unmarshal(buf []byte) error {
	if len(buf) < 10 {
		return fmt.Errorf(
			"cannot unmarshal SystemEventMessage from %v-length buffer",
			len(buf))
	}

	m.SystemEvent = uint8(buf[1])
	m.Timestamp = parseTimestamp(buf[2:10])

	return nil
}

const (
	// Outside of heartbeat messages on the lower level protocol,
	// the start of day message is the first message in any trading session.
	StartOfMessages uint8 = 0x4f
	// This message indicates that IEX is open and ready to start accepting
	// orders.
	StartOfSystemHours uint8 = 0x53
	// This message indicates that DAY and GTX orders, as well as
	// market orders and pegged orders, are available for execution on IEX.
	StartOfRegularMarketHours uint8 = 0x52
	// This message indicates that DAY orders, market orders, and pegged
	// orders are no longer accepted by IEX.
	EndOfRegularMarketHours uint8 = 0x4d
	// This message indicates that IEX is now closed and will not accept
	// any new orders during this trading session. It is still possible to
	// receive messages after the end of day.
	EndOfSystemHours uint8 = 0x45
	// This is always the last message sent in any trading session.
	EndOfMessages uint8 = 0x43
)

// IEX disseminates a full pre-market spin of SecurityDirectoryMessages for
// all IEX-listed securities. After the pre-market spin, IEX will use the
// SecurityDirectoryMessage to relay changes for an individual security.
type SecurityDirectoryMessage struct {
	// See Appendix A for flag values.
	Flags uint8
	// The time of the update event as set by the IEX Trading System logic.
	Timestamp time.Time
	// IEX-listed security represented in Nasdaq Integrated symbology.
	Symbol string
	// The number of shares that represent a round lot for the security.
	RoundLotSize uint32
	// The corporate action adjusted previous official closing price for
	// the security (e.g. stock split, dividend, rights offering).
	// When no corporate action has occurred, the Adjusted POC Price
	// will be populated with the previous official close price. For
	// new issues (e.g., an IPO), this field will be the issue price.
	AdjustedPOCPrice float64
	// Indicates which Limit Up-Limit Down price band calculation
	// parameter is to be used.
	LULDTier uint8
}

func (m *SecurityDirectoryMessage) Unmarshal(buf []byte) error {
	if len(buf) < 31 {
		return fmt.Errorf(
			"cannot unmarshal SecurityDirectoryMessage from %v-length buffer",
			len(buf))
	}

	m.Flags = uint8(buf[1])
	m.Timestamp = parseTimestamp(buf[2:10])
	m.Symbol = strings.TrimRight(string(buf[10:18]), " ")
	m.RoundLotSize = binary.LittleEndian.Uint32(buf[18:22])
	m.AdjustedPOCPrice = parseFloat(buf[22:30])
	m.LULDTier = uint8(buf[30])

	return nil
}

func (m *SecurityDirectoryMessage) IsTestSecurity() bool {
	return m.Flags&0x80 != 0
}

func (m *SecurityDirectoryMessage) IsWhenIssuedSecurity() bool {
	return m.Flags&0x40 != 0
}

func (m *SecurityDirectoryMessage) IsETP() bool {
	return m.Flags&0x20 != 0
}

const (
	// Not applicable.
	LULDTier0 uint8 = 0x0
	// Tier 1 NMS Stock.
	LULDTier1 uint8 = 0x1
	// Tier 2 NMS Stock.
	LULDTier2 uint8 = 0x2
)

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
// *The paused and released into an Order Acceptance Period status will be
// disseminated for IEX-listed securities only. Trading pauses on
// non-IEX-listed securities will be treated simply as a halt.
type TradingStatusMessage struct {
	// Trading status.
	TradingStatus uint8
	// The time of the update event as set by the IEX Trading System logic.
	Timestamp time.Time
	// Security represented in Nasdaq integrated symbology.
	Symbol string
	// IEX populates the Reason field for IEX-listed securities when the
	// TradingStatus is TradingHalted or OrderAcceptancePeriod.
	// For non-IEX listed securities, the Reason field will be set to
	// ReasonNotAvailable when the trading status is TradingHalt.
	// The Reason will be blank when the trading status is TradingPause
	// or Trading.
	Reason string
}

func (m *TradingStatusMessage) Unmarshal(buf []byte) error {
	if len(buf) < 22 {
		return fmt.Errorf(
			"cannot unmarshal SecurityDirectoryMessage from %v-length buffer",
			len(buf))
	}

	m.TradingStatus = uint8(buf[1])
	m.Timestamp = parseTimestamp(buf[2:10])
	m.Symbol = parseString(buf[10:18])
	m.Reason = parseString(buf[18:22])
	return nil
}

const (
	// Trading halted across all US equity markets.
	TradingHalt uint8 = 0x48
	// Trading halt released into an Order Acceptance Period
	// (IEX-listed securities only)
	TradingOrderAcceptancePeriod uint8 = 0x4f
	// Trading paused and Order Acceptance Period on IEX
	// (IEX-listed securities only)
	TradingPaused uint8 = 0x50
	// Trading on IEX
	Trading uint8 = 0x54
)

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
