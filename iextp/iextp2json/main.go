package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcapgo"

	"github.com/timpalpant/go-iex/iextp"
	"github.com/timpalpant/go-iex/iextp/tops"
)

func loadProtocol(name string) iextp.Protocol {
	var protocol iextp.Protocol
	switch name {
	case "tops":
		protocol = tops.Protocol{}
	default:
		log.Fatal(
			"Unsupported protocol: %v. Options: tops, deep",
			name)
	}

	return protocol
}

func main() {
	flag.Parse()

	input := bufio.NewReader(os.Stdin)
	pcapReader, err := pcapgo.NewReader(input)
	if err != nil {
		log.Fatal(err)
	}

	packetSource := gopacket.NewPacketSource(pcapReader, pcapReader.LinkType())
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

			fmt.Println(segment)
		}
	}
}
