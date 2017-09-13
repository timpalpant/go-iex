package iex

import (
	"bufio"
	"compress/gzip"
	"encoding/binary"
	"io"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcapgo"
	"github.com/kor44/pcapng"

	"github.com/timpalpant/go-iex/iextp"
	_ "github.com/timpalpant/go-iex/iextp/deep"
	_ "github.com/timpalpant/go-iex/iextp/tops"
)

const (
	magicGzip1         = 0x1f
	magicGzip2         = 0x8b
	pcapNGMagic uint32 = 0x0A0D0D0A
)

type PacketDataSource interface {
	gopacket.PacketDataSource
	LinkType() layers.LinkType
}

// Initializes a new packet source for pcap or pcap-ng files.
// Looks at the first 4 bytes to determine if the given Reader
// is a pcap or pcap-ng format.
func NewPacketDataSource(r io.Reader) (PacketDataSource, error) {
	input := bufio.NewReader(r)
	gzipMagic, err := input.Peek(2)
	if err != nil {
		return nil, err
	}

	if gzipMagic[0] == magicGzip1 && gzipMagic[1] == magicGzip2 {
		if gzf, err := gzip.NewReader(input); err != nil {
			return nil, err
		} else {
			input = bufio.NewReader(gzf)
		}
	}

	magicBuf, err := input.Peek(4)
	if err != nil {
		return nil, err
	}
	magic := binary.LittleEndian.Uint32(magicBuf)

	var packetSource PacketDataSource
	if magic == pcapNGMagic {
		packetSource, err = pcapng.NewReader(input)
	} else {
		packetSource, err = pcapgo.NewReader(input)
	}

	return packetSource, err
}

// PcapScanner is a high-level reader for extracting messages from the
// pcap dumps provided by IEX in the HIST endpoint.
type PcapScanner struct {
	packetSource    *gopacket.PacketSource
	currentSegment  []iextp.Message
	currentMsgIndex int
}

func NewPcapScanner(packetDataSource PacketDataSource) *PcapScanner {
	packetSource := gopacket.NewPacketSource(packetDataSource, packetDataSource.LinkType())
	return &PcapScanner{
		packetSource: packetSource,
	}
}

// Get the next Message in the pcap dump.
func (p *PcapScanner) NextMessage() (iextp.Message, error) {
	for p.currentMsgIndex >= len(p.currentSegment) {
		if err := p.nextSegment(); err != nil {
			return nil, err
		}
	}

	msg := p.currentSegment[p.currentMsgIndex]
	p.currentMsgIndex++
	return msg, nil
}

// Read packets until we find the next one with > 0 messages.
func (p *PcapScanner) nextSegment() error {
	for {
		packet, err := p.packetSource.NextPacket()
		if err != nil {
			return err
		}

		if app := packet.ApplicationLayer(); app != nil {
			segment := iextp.Segment{}
			if err := segment.Unmarshal(app.Payload()); err != nil {
				return err
			}

			if len(segment.Messages) != 0 {
				p.currentSegment = segment.Messages
				p.currentMsgIndex = 0
				return nil
			}
		}
	}
}
