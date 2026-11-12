package tanswer

import (
	"fmt"
	"strings"
	"time"
)

type TimeRangeOptions struct {
	Time  string
	Start string
	End   string
	Now   time.Time
}

type TimeRange struct {
	Type  string `json:"type"`
	Start int64  `json:"start"`
	End   int64  `json:"end"`
}

func ParseTimeRange(opts TimeRangeOptions) (TimeRange, error) {
	now := opts.Now
	if now.IsZero() {
		now = time.Now()
	}
	if opts.Start != "" || opts.End != "" {
		start, err := parseLocalTime(opts.Start)
		if err != nil {
			return TimeRange{}, err
		}
		end, err := parseLocalTime(opts.End)
		if err != nil {
			return TimeRange{}, err
		}
		if start.After(end) {
			return TimeRange{}, fmt.Errorf("start time cannot be later than end time")
		}
		return TimeRange{Type: "custom", Start: start.UnixMilli(), End: end.UnixMilli()}, nil
	}
	switch strings.ToLower(firstNonEmpty(opts.Time, "today")) {
	case "today":
		start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		end := start.Add(24*time.Hour - time.Millisecond)
		return TimeRange{Type: "today", Start: start.UnixMilli(), End: end.UnixMilli()}, nil
	case "24h":
		return TimeRange{Type: "24h", Start: now.Add(-24 * time.Hour).UnixMilli(), End: now.UnixMilli()}, nil
	case "7d":
		return TimeRange{Type: "7d", Start: now.Add(-7 * 24 * time.Hour).UnixMilli(), End: now.UnixMilli()}, nil
	default:
		return TimeRange{}, fmt.Errorf("unsupported time range %q", opts.Time)
	}
}

func parseLocalTime(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, fmt.Errorf("start and end must be both set for custom time range")
	}
	return time.ParseInLocation("2006-01-02 15:04:05", value, time.Local)
}
