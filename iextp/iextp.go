package iextp

import (
	"encoding/binary"
	"fmt"
	"io"
	"time"
)

// Segment represents an IEXTP Segment.
type Segment struct {
	Header   SegmentHeader
	Messages []Message
}

// Message represents an IEXTP message.
type Message interface {
	// Unmarshal unmarshals the given byte content into the Message.
	// Note that buf includes the entire message content, including the
	// leading message type byte.
	//
	// IEX reserves the right to grow the message length without notice,
	// but only by adding additional data to the end of the message, so
	// decoders should handle messages that grow beyond the expected
	// length.
	Unmarshal(buf []byte) error
}

// UnsupportedMessage may be returned by a protocol for any
// message types it does not know how to decode.
type UnsupportedMessage []byte

func (m *UnsupportedMessage) Unmarshal(buf []byte) error {
	*m = buf
	return nil
}

type SegmentHeader struct {
	// Version of the IEX-TP protocol.
	Version uint8
	// A unique identifier for the higher-layer specification that describes
	// the messages contaiend within a segment. See the higher-layer protocol
	// specification for the protocol's message identification in IEX-TP.
	MessageProtocolID uint16
	// An identifier for a given stream of bytes/sequenced messages. Messages
	// received from multiple sources which use the same Channel ID are
	// guaranteed to be duplicates by sequence number and/or offset. See the
	// higher-layer protocol specification for the protocol's channel
	// identification on IEX-TP.
	ChannelID uint32
	// SessionID uniquely identifies a stream of messages produced by the
	// system. A given message is uniquely identified within a message
	// protocol by its Session ID and Sequence Number.
	SessionID uint32
	// StreamOffset is a counter representing the byte offset of the payload
	// in the data stream.
	StreamOffset int64
	// PayloadLength is an unsigned binary count representing the number
	// of bytes contained in the segment's payload. Note that the Payload
	// Length field value does not include the length of the IEX-TP
	// header.
	PayloadLength uint16
	// FirstMessageSequenceNumber is a counter representing the sequence
	// number of the first message in the segment. If there is more than one
	// message in a segment, all subsequent messages are implicitly
	// numbered sequentially.
	FirstMessageSequenceNumber int64
	// MessageCount is a count representing the number of Message Blocks
	// in the segment.
	MessageCount uint16
	// The time the outbound segment was sent as set by the sender.
	SendTime time.Time
}

func (sh *SegmentHeader) Unmarshal(r io.Reader) error {
	if err := binary.Read(r, binary.LittleEndian, &sh.Version); err != nil {
		return err
	}

	if err := binary.Read(r, binary.LittleEndian, &sh.MessageProtocolID); err != nil {
		return err
	}

	if err := binary.Read(r, binary.LittleEndian, &sh.ChannelID); err != nil {
		return err
	}

	if err := binary.Read(r, binary.LittleEndian, &sh.SessionID); err != nil {
		return err
	}

	if err := binary.Read(r, binary.LittleEndian, &sh.PayloadLength); err != nil {
		return err
	}

	if err := binary.Read(r, binary.LittleEndian, &sh.MessageCount); err != nil {
		return err
	}

	if err := binary.Read(r, binary.LittleEndian, &sh.StreamOffset); err != nil {
		return err
	}

	if err := binary.Read(r, binary.LittleEndian, &sh.FirstMessageSequenceNumber); err != nil {
		return err
	}

	var timestampNs int64
	if err := binary.Read(r, binary.LittleEndian, &timestampNs); err != nil {
		return err
	}
	sh.SendTime = time.Unix(0, timestampNs)

	return nil
}

// Protocol represents a higher-level IEXTP protocol, such as TOPS or DEEP.
type Protocol interface {
	// ID returns the MessageProtocolID for this protocol.
	// During scanning, it is an error if a Segment's MessageProtocolID
	// does not match the one expected by the Protocol.
	ID() uint16
	// Unmarshal a Message received in an IEXTP segment.
	// Note that buf contains only the message content.
	Unmarshal(buf []byte) (Message, error)
}

// Scanner provides a convenient interface for reading IEXTP Segments
// from a data stream. It behaves like a bufio.Scanner, advancing
// to the next Segment with each call to Scan until EOF or an I/O error.
type Scanner struct {
	reader   io.Reader
	protocol Protocol

	current *Segment
	err     error
}

func NewScanner(r io.Reader, p Protocol) *Scanner {
	return &Scanner{reader: r, protocol: p}
}

// Scan advances the Scanner to the next segment, which will
// then be available through the Segment method. It returns
// false when the scan stops, either by reaching the end of
// the input or an error. After Scan returns false, the Err
// method will return any error that occurred during scanning,
// except that if it was io.EOF, Err will return nil.
func (s *Scanner) Scan() bool {
	// Unmarshal segment header.
	header := SegmentHeader{}
	if err := header.Unmarshal(s.reader); err != nil {
		s.err = err
		return false
	}

	if header.MessageProtocolID != s.protocol.ID() {
		s.err = fmt.Errorf(
			"Incorrect segment protocol id: segment %v != protocol %v",
			header.MessageProtocolID, s.protocol.ID())
		return false
	}

	// Unmarshal segment messages.
	segment := &Segment{
		Header:   header,
		Messages: make([]Message, 0, header.MessageCount),
	}
	more := true
	for i := uint16(0); i < segment.Header.MessageCount; i++ {
		var messageLength uint16
		if err := binary.Read(s.reader, binary.LittleEndian, &messageLength); err != nil {
			s.err = err
			return false
		}

		buf := make([]byte, messageLength)
		if _, err := io.ReadFull(s.reader, buf); err != nil {
			s.err = err
			return false
		}

		msg, err := s.protocol.Unmarshal(buf)
		if err != nil {
			if err != io.EOF {
				s.err = err
				return false
			} else if i != segment.Header.MessageCount-1 {
				s.err = io.ErrUnexpectedEOF
				return false
			} else {
				// (Expected) EOF and end of messages.
				more = false
			}
		}

		segment.Messages = append(segment.Messages, msg)
	}

	s.current = segment
	return more
}

// Segment returns the current Segment parsed from a recent call to Scan.
func (s *Scanner) Segment() *Segment {
	return s.current
}

// Err returns the first non-EOF error that was encountered by the Scanner.
func (s *Scanner) Err() error {
	return s.err
}
