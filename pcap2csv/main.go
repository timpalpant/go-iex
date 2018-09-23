// pcap2csv is a small binary for extracting IEXTP messages
// from a pcap dump and converting them to minute-resolution bars
// in CSV format for research.
//
// The pcap dump is read from stdin, and may be gzipped,
// and the resulting CSV data is written to stdout.
package main

import (
	"encoding/csv"
	"io"
	"log"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/timpalpant/go-iex"
	"github.com/timpalpant/go-iex/consolidator"
	"github.com/timpalpant/go-iex/iextp/tops"
)

var header = []string{
	"symbol",
	"time",
	"open",
	"high",
	"low",
	"close",
	"volume",
}

func makeBars(trades []*tops.TradeReportMessage, openTime, closeTime time.Time) []*consolidator.Bar {
	bars := consolidator.MakeBars(trades)
	for _, bar := range bars {
		bar.OpenTime = openTime
		bar.CloseTime = closeTime
	}

	sort.Slice(bars, func(i, j int) bool {
		return bars[i].Symbol < bars[j].Symbol
	})

	return bars
}

func writeBar(bar *consolidator.Bar, w *csv.Writer) error {
	row := []string{
		bar.Symbol,
		bar.OpenTime.Format(time.RFC3339),
		strconv.FormatFloat(bar.Open, 'f', 4, 64),
		strconv.FormatFloat(bar.High, 'f', 4, 64),
		strconv.FormatFloat(bar.Low, 'f', 4, 64),
		strconv.FormatFloat(bar.Close, 'f', 4, 64),
		strconv.FormatInt(bar.Volume, 10),
	}

	return w.Write(row)
}

func writeBars(bars []*consolidator.Bar, w *csv.Writer) error {
	for _, bar := range bars {
		if err := writeBar(bar, w); err != nil {
			return err
		}
	}

	return nil
}

func main() {
	packetSource, err := iex.NewPacketDataSource(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}

	scanner := iex.NewPcapScanner(packetSource)
	writer := csv.NewWriter(os.Stdout)
	if err := writer.Write(header); err != nil {
		log.Fatal(err)
	}
	defer writer.Flush()

	var trades []*tops.TradeReportMessage
	var openTime, closeTime time.Time
	for {
		msg, err := scanner.NextMessage()
		if err != nil {
			if err == io.EOF {
				break
			}

			log.Fatal(err)
		}

		if msg, ok := msg.(*tops.TradeReportMessage); ok {
			if openTime.IsZero() {
				openTime = msg.Timestamp.Truncate(time.Minute)
				closeTime = openTime.Add(time.Minute)
			}

			if msg.Timestamp.After(closeTime) && len(trades) > 0 {
				bars := makeBars(trades, openTime, closeTime)
				if err := writeBars(bars, writer); err != nil {
					log.Fatal(err)
				}

				trades = trades[:0]
				openTime = msg.Timestamp.Truncate(time.Minute)
				closeTime = openTime.Add(time.Minute)
			}

			trades = append(trades, msg)
		}
	}
}
