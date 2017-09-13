package iex

import (
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestPcapScanner(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping pcap test in short mode.")
	}

	testFilename := filepath.Join("testdata", "DEEP10.pcap.gz")
	count := testPcapScanner(t, testFilename)

	if count != 392000 {
		t.Fatalf("expected to process 392000 messages, got: %v", count)
	}
}

func TestPcapNgScanner(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping pcap-ng test in short mode.")
	}

	testFilename := filepath.Join("testdata", "TOPS16.pcapng.gz")
	count := testPcapScanner(t, testFilename)
	if count != 57675 {
		t.Fatalf("expected to process 57675 messages, got: %v", count)
	}
}

func testPcapScanner(t *testing.T, filename string) int {
	f, err := os.Open(filename)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		t.Fatal(err)
	}

	packetDataSource, err := NewPacketDataSource(f)
	if err != nil {
		t.Fatal(err)
	}

	scanner := NewPcapScanner(packetDataSource)

	start := time.Now()
	count := 0
	for err = nil; err == nil; count++ {
		_, err = scanner.NextMessage()
	}
	elapsed := time.Since(start)

	msgsPerSec := float64(count) / elapsed.Seconds()
	mbPerSec := float64(stat.Size()) / 1000 / 1000 / elapsed.Seconds()
	t.Logf("Processed %d messages (%.0f msgs/sec, %.1f MB/s)",
		count, msgsPerSec, mbPerSec)

	// The sample pcap file ends with an unexpected EOF.
	// TODO(palpant): Fix it so that we can assert a clean ending here.
	if err != io.EOF && err != io.ErrUnexpectedEOF {
		t.Fatal(err)
	}

	return count
}
