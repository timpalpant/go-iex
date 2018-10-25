package iex

import (
	"io"
	"net"
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

func TestUDPScanner(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping UDP test in short mode.")
	}

	nPacketsToSend := 100
	packetConn, err := net.ListenPacket("udp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}
	defer packetConn.Close()
	packetSource := NewPacketConnDataSource(packetConn)
	t.Logf("Listing on udp://%s", packetConn.LocalAddr())

	// Replay the given pcap dump to the given UDP address.
	testFilename := filepath.Join("testdata", "DEEP10.pcap.gz")
	go udpReplay(t, testFilename, packetConn.LocalAddr(), nPacketsToSend)

	time.Sleep(time.Second)
	t.Log("Scanning UDP packets")
	scanner := NewPcapScanner(packetSource)
	for i := 0; i < nPacketsToSend; i++ {
		if _, err := scanner.NextMessage(); err != nil {
			t.Fatal(err)
		}
	}
}

// Replays all packets in the given pcap filename to the given address.
func udpReplay(t *testing.T, pcapFilename string, addr net.Addr, nPacketsToSend int) {
	t.Log("Dialing: ", addr)
	conn, err := net.DialTimeout("udp", addr.String(), time.Second)
	if err != nil {
		t.Fatal("could not connect to server: ", err)
	}
	defer conn.Close()

	f, err := os.Open(pcapFilename)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	packetSource, err := NewPcapDataSource(f)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Replaying first %d packets from %s", nPacketsToSend, pcapFilename)
	for i := 0; i < nPacketsToSend; i++ {
		payload, err := packetSource.NextPayload()
		if err != nil {
			if err == io.EOF {
				return
			}

			t.Fatal("could not write payload to server:", err)
			return
		}

		if _, err := conn.Write(payload); err != nil {
			t.Fatal("could not write payload to server:", err)
			return
		}
	}
}
