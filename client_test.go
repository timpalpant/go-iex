package iex

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type mockHTTPClient struct {
	body    string
	headers map[string]string
	code    int
	err     error
}

func (c *mockHTTPClient) Get(url string) (*http.Response, error) {
	w := httptest.NewRecorder()
	w.WriteString(c.body)

	for key, value := range c.headers {
		w.Header().Add(key, value)
	}

	resp := w.Result()
	return resp, c.err
}

func setupTestClient() *Client {
	return NewClient(&http.Client{
		Timeout: 5 * time.Second,
	})
}

func TestTOPS_AllSymbols(t *testing.T) {
	// TODO: Add expected field to struct and use it to verify results
	var testCases = []struct {
		symbols []string
		code    int
		body    string
		err     error
		headers map[string]string
	}{
		{symbols: []string{"SNAP", "FB"}, code: 200, body: `[{"symbol":"SNAP","sector":"softwareservices","securityType":"commonstock","bidPrice":0,"bidSize":0,"askPrice":0,"askSize":0,"lastUpdated":1537215438021,"lastSalePrice":9.165,"lastSaleSize":123,"lastSaleTime":1537214395927,"volume":525079,"marketPercent":0.0238},{"symbol":"FB","sector":"softwareservices","securityType":"commonstock","bidPrice":0,"bidSize":0,"askPrice":0,"askSize":0,"lastUpdated":1537216916977,"lastSalePrice":160.6,"lastSaleSize":100,"lastSaleTime":1537214399372,"volume":991898,"marketPercent":0.04741}]`, err: nil, headers: map[string]string{"Content-Type": "application/json"}},
		{symbols: []string{"AIG+"}, code: 200, body: `[{"symbol":"AIG+","sector":"n/a","securityType":"warrant","bidPrice":0,"bidSize":0,"askPrice":0,"askSize":0,"lastUpdated":1537214400001,"lastSalePrice":0,"lastSaleSize":0,"lastSaleTime":0,"volume":0,"marketPercent":0}]`, err: nil, headers: map[string]string{"Content-Type": "application/json"}},
	}

	for _, tt := range testCases {
		httpc := mockHTTPClient{body: tt.body, code: tt.code, err: tt.err, headers: tt.headers}
		c := NewClient(&httpc)

		result, err := c.GetTOPS(tt.symbols)

		if err != nil {
			t.Fatal(err)
		}

		if len(result) != len(tt.symbols) {
			t.Fatalf("Received %v results, expected %v", len(result), len(tt.symbols))
		}
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

func TestHIST_OneDate(t *testing.T) {
	c := setupTestClient()
	testDate := time.Date(2017, time.June, 6, 0, 0, 0, 0, time.UTC)
	result, err := c.GetHIST(testDate)
	if err != nil {
		t.Fatal(err)
	}

	if len(result) == 0 {
		t.Fatalf("Received zero results")
	}
}

func TestHIST_AllDates(t *testing.T) {
	c := setupTestClient()
	result, err := c.GetAllAvailableHIST()
	if err != nil {
		t.Fatal(err)
	}

	if len(result) == 0 {
		t.Fatalf("Received zero results")
	}
}

func TestDEEP(t *testing.T) {
	c := setupTestClient()
	result, err := c.GetDEEP("SPY")
	if err != nil {
		t.Fatal(err)
	}

	if result.Symbol != "SPY" {
		t.Fatalf("Expected symbol = %v, got %v", "SPY", result.Symbol)
	}
}

func TestBook(t *testing.T) {
	c := setupTestClient()
	symbols := []string{"SPY"}
	result, err := c.GetBook(symbols)
	if err != nil {
		t.Fatal(err)
	}

	if len(result) != len(symbols) {
		t.Log(result)
		t.Fatalf("Received %v results, expected %v", len(result), len(symbols))
	}
}

func TestSymbols(t *testing.T) {
	c := setupTestClient()
	symbols, err := c.GetSymbols()
	if err != nil {
		t.Fatal(err)
	}

	if len(symbols) == 0 {
		t.Fatal("Received zero symbols")
	}

	symbol := symbols[0]
	if symbol.Symbol == "" || symbol.Name == "" || symbol.Date == "" {
		t.Fatal("Failed to decode symbol correctly")
	}
}

func TestMarkets(t *testing.T) {
	c := setupTestClient()
	markets, err := c.GetMarkets()
	if err != nil {
		t.Fatal(err)
	}

	if len(markets) == 0 {
		t.Fatal("Received zero markets")
	}
}

func TestGetHistoricalDaily(t *testing.T) {
	c := setupTestClient()
	stats, err := c.GetHistoricalDaily(&HistoricalDailyRequest{Last: 5})
	if err != nil {
		t.Fatal(err)
	}

	if len(stats) != 5 {
		t.Fatalf("Received %d historical daily stats, expected %d", len(stats), 5)
	}
}
