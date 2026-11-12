package tanswer

import (
	"testing"
	"time"
)

func TestParseTodayRange(t *testing.T) {
	loc := time.FixedZone("CST", 8*60*60)
	now := time.Date(2026, 6, 26, 15, 30, 0, 0, loc)
	rng, err := ParseTimeRange(TimeRangeOptions{Time: "today", Now: now})
	if err != nil {
		t.Fatalf("ParseTimeRange returned error: %v", err)
	}
	if rng.Start != time.Date(2026, 6, 26, 0, 0, 0, 0, loc).UnixMilli() {
		t.Fatalf("start = %d", rng.Start)
	}
	if rng.End != time.Date(2026, 6, 26, 23, 59, 59, 999000000, loc).UnixMilli() {
		t.Fatalf("end = %d", rng.End)
	}
}

func TestParse24hRange(t *testing.T) {
	now := time.Date(2026, 6, 26, 15, 30, 0, 0, time.UTC)
	rng, err := ParseTimeRange(TimeRangeOptions{Time: "24h", Now: now})
	if err != nil {
		t.Fatalf("ParseTimeRange returned error: %v", err)
	}
	if rng.Start != now.Add(-24*time.Hour).UnixMilli() || rng.End != now.UnixMilli() {
		t.Fatalf("range = %#v", rng)
	}
}

func TestParseExplicitRangeRejectsInvertedRange(t *testing.T) {
	_, err := ParseTimeRange(TimeRangeOptions{
		Start: "2026-06-27 00:00:00",
		End:   "2026-06-26 00:00:00",
	})
	if err == nil {
		t.Fatal("expected inverted range error")
	}
}
