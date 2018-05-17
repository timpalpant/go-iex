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

	"github.com/timpalpant/go-iex"
)

func main() {
	packetSource, err := iex.NewPacketDataSource(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}

	scanner := iex.NewPcapScanner(packetSource)
	output := bufio.NewWriter(os.Stdout)
	defer output.Flush()
	enc := json.NewEncoder(output)

	for {
		msg, err := scanner.NextMessage()
		if err != nil {
			if err == io.EOF {
				break
			}

			log.Fatal(err)
		}

		enc.Encode(msg)
	}
}
