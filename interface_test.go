package iex

import (
	"encoding/json"
	"testing"
)

func TestStatsUnmarshal_IntFalse(t *testing.T) {
	intStats := []byte(`{
		"date": "2017-05-09",
		"volume": 152907569,
		"routedVolume": 46943802,
		"marketShare": 0.02246,
		"isHalfday": 0,
		"litVolume": 35426666
	}`)

	var stats *Stats
	if err := json.Unmarshal(intStats, &stats); err != nil {
		t.Fatal(err)
	}

	if stats.IsHalfDay {
		t.Fatalf("did not unmarshal halfday correctly: got %v, expected %v",
			stats.IsHalfDay, false)
	}
}

func TestStatsUnmarshal_IntTrue(t *testing.T) {
	intStats := []byte(`{
		"date": "2017-05-09",
		"volume": 152907569,
		"routedVolume": 46943802,
		"marketShare": 0.02246,
		"isHalfday": 1,
		"litVolume": 35426666
	}`)

	var stats *Stats
	if err := json.Unmarshal(intStats, &stats); err != nil {
		t.Fatal(err)
	}

	if !stats.IsHalfDay {
		t.Fatalf("did not unmarshal halfday correctly: got %v, expected %v",
			stats.IsHalfDay, true)
	}
}

func TestStatsUnmarshal_BoolFalse(t *testing.T) {
	boolStats := []byte(`{
		"date": "2017-01-11",
		"volume": 128048723,
		"routedVolume": 38314207,
		"marketShare": 0.01769,
		"isHalfday": false,
		"litVolume": 30520534
	}`)

	var stats *Stats
	if err := json.Unmarshal(boolStats, &stats); err != nil {
		t.Fatal(err)
	}

	if stats.IsHalfDay {
		t.Fatalf("did not unmarshal halfday correctly: got %v, expected %v",
			stats.IsHalfDay, false)
	}
}

func TestStatsUnmarshal_BoolTrue(t *testing.T) {
	boolStats := []byte(`{
		"date": "2017-01-11",
		"volume": 128048723,
		"routedVolume": 38314207,
		"marketShare": 0.01769,
		"isHalfday": true,
		"litVolume": 30520534
	}`)

	var stats *Stats
	if err := json.Unmarshal(boolStats, &stats); err != nil {
		t.Fatal(err)
	}

	if !stats.IsHalfDay {
		t.Fatalf("did not unmarshal halfday correctly: got %v, expected %v",
			stats.IsHalfDay, true)
	}
}
