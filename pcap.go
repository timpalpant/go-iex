package iex

import (
	"io"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcapgo"

	"github.com/timpalpant/go-iex/iextp"
	_ "github.com/timpalpant/go-iex/iextp/deep"
	_ "github.com/timpalpant/go-iex/iextp/tops"
)

// PcapScanner is a high-level reader for extracting messages from the
// pcap dumps provided by IEX in the HIST endpoint.
type PcapScanner struct {
	packetSource    *gopacket.PacketSource
	currentSegment  []iextp.Message
	currentMsgIndex int
}

func NewPcapScanner(r io.Reader) (*PcapScanner, error) {
	pcapReader, err := pcapgo.NewReader(r)
	if err != nil {
		return nil, err
	}
	packetSource := gopacket.NewPacketSource(pcapReader, pcapReader.LinkType())

	return &PcapScanner{
		packetSource: packetSource,
	}, nil
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
