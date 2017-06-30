package iex

import (
	"net/http"
	"testing"
	"time"
)

func setupTestClient() *Client {
	return NewClient(&http.Client{
		Timeout: 5 * time.Second,
	})
}

func testTOPS(t *testing.T, symbols []string) {
	c := setupTestClient()
	result, err := c.GetTOPS(symbols)
	if err != nil {
		t.Fatal(err)
	}

	if len(result) != len(symbols) {
		t.Fatalf("Received %v results, expected %v", len(result), len(symbols))
	}

	// TODO(palpant): Test parsing with a mock http client.
}

func TestTOPS_OneSymbol(t *testing.T) {
	testTOPS(t, []string{"SPY"})
}

func TestTOPS_TwoSymbols(t *testing.T) {
	testTOPS(t, []string{"SPY", "AAPL"})
}

func TestTOPS_AllSymbols(t *testing.T) {
	c := setupTestClient()
	result, err := c.GetTOPS(nil)
	if err != nil {
		t.Fatal(err)
	}

	if len(result) == 0 {
		t.Fatalf("Received %v results", len(result))
	}
}

func TestLast(t *testing.T) {
	c := setupTestClient()
	symbols := []string{"SPY", "AAPL"}
	result, err := c.GetLast(symbols)
	if err != nil {
		t.Fatal(err)
	}

	if len(result) != len(symbols) {
		t.Fatalf("Received %v results, expected %v", len(result), len(symbols))
	}
}
