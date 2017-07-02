// pcap2json is a small binary for extracting IEXTP messages
// from a pcap dump and converting them to JSON.
//
// The pcap dump is read from stdin, and may be gzipped,
// and the resulting JSON messages are written to stdout.
package main

import (
	"bufio"
	"encoding/json"
	"io"
	"log"
	"os"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcapgo"

	"github.com/timpalpant/go-iex/iextp"
	_ "github.com/timpalpant/go-iex/iextp/deep"
	_ "github.com/timpalpant/go-iex/iextp/tops"
)

func main() {
	input := bufio.NewReader(os.Stdin)
	pcapReader, err := pcapgo.NewReader(input)
	if err != nil {
		log.Fatal(err)
	}
	packetSource := gopacket.NewPacketSource(pcapReader, pcapReader.LinkType())

	output := bufio.NewWriter(os.Stdout)
	enc := json.NewEncoder(output)

	for {
		packet, err := packetSource.NextPacket()
		if err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		}

		if app := packet.ApplicationLayer(); app != nil {
			segment := iextp.Segment{}
			if err := segment.Unmarshal(app.Payload()); err != nil {
				log.Fatal(err)
			}

			for _, msg := range segment.Messages {
				enc.Encode(msg)
			}
		}
	}
}
