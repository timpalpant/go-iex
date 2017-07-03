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
	input := bufio.NewReader(os.Stdin)
	scanner, err := iex.NewPcapScanner(input)
	if err != nil {
		log.Fatal(err)
	}

	output := bufio.NewWriter(os.Stdout)
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
