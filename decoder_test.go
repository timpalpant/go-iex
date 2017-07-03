package iex

import (
	"bufio"
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
	f, err := os.Open(testFilename)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		t.Fatal(err)
	}

	r := bufio.NewReader(f)
	scanner, err := NewPcapScanner(r)
	if err != nil {
		t.Fatal(err)
	}

	start := time.Now()
	count := 0
	for err = nil; err == nil; count++ {
		_, err = scanner.NextMessage()
	}
	elapsed := time.Since(start)

	if count != 392000 {
		t.Fatalf("expected to process 392000 messages, got: %v", count)
	}

	msgsPerSec := float64(count) / elapsed.Seconds()
	mbPerSec := float64(stat.Size()) / 1000 / 1000 / elapsed.Seconds()
	t.Logf("Processed %d messages (%.0f msgs/sec, %.1f MB/s)",
		count, msgsPerSec, mbPerSec)

	// The sample pcap file ends with an unexpected EOF.
	// TODO(palpant): Fix it so that we can assert a clean ending here.
	if err != io.EOF && err != io.ErrUnexpectedEOF {
		t.Fatal(err)
	}
}
