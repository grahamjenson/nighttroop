package display

import (
	"context"
	"testing"
	"time"
)

func TestAgo(t *testing.T) {
	now := time.Date(2026, 10, 26, 12, 0, 0, 0, time.UTC)
	at := func(before time.Duration) *int64 {
		ms := now.Add(-before).UnixMilli()
		return &ms
	}
	zero := int64(0)
	tests := []struct {
		name string
		ms   *int64
		want string
	}{
		{"no time", nil, "Never"},
		{"zero", &zero, "Never"},
		{"seconds", at(30 * time.Second), "just now"},
		{"minutes", at(5*time.Minute + 20*time.Second), "5 min ago"},
		{"hours", at(3*time.Hour + 10*time.Minute), "3 h ago"},
		{"a day is still hours", at(47 * time.Hour), "47 h ago"},
		{"days", at(5 * 24 * time.Hour), "5 days ago"},
		{"in the future", at(-time.Hour), "just now"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Ago(context.Background(), tt.ms, now); got != tt.want {
				t.Errorf("Ago(context.Background(), ) = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDateTime(t *testing.T) {
	at := func(when time.Time) *int64 {
		ms := when.UnixMilli()
		return &ms
	}
	zero := int64(0)
	tests := []struct {
		name string
		ms   *int64
		want string
	}{
		{"no time", nil, "-"},
		{"zero", &zero, "-"},
		{"utc", at(time.Date(2026, 10, 26, 12, 0, 5, 0, time.UTC)), "2026-10-26 12:00:05"},
		{"another zone is shown in utc", at(time.Date(2026, 10, 27, 1, 30, 0, 0, time.FixedZone("NZDT", 13*60*60))), "2026-10-26 12:30:00"},
		{"ms are dropped", at(time.Date(2026, 1, 2, 3, 4, 5, 999e6, time.UTC)), "2026-01-02 03:04:05"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DateTime(tt.ms); got != tt.want {
				t.Errorf("DateTime() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTimeAgo(t *testing.T) {
	now := time.Date(2026, 10, 26, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		before time.Duration
		want   string
	}{
		{30 * time.Second, "0m ago"},
		{5 * time.Minute, "5m ago"},
		{59*time.Minute + 59*time.Second, "59m ago"},
		{time.Hour, "1h0m ago"},
		{time.Hour + 5*time.Minute, "1h5m ago"},
		{23*time.Hour + 59*time.Minute, "23h59m ago"},
		{24 * time.Hour, "1d0h0m ago"},
		{49*time.Hour + 5*time.Minute, "2d1h5m ago"},
		{-10 * time.Minute, "0m ago"},
	}
	for _, tt := range tests {
		if got := TimeAgo(context.Background(), now.Add(-tt.before).UnixMilli(), now); got != tt.want {
			t.Errorf("TimeAgo(context.Background(), %s before) = %q, want %q", tt.before, got, tt.want)
		}
	}
}

func TestRoundedAgo(t *testing.T) {
	now := time.Date(2026, 10, 26, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		before time.Duration
		want   string
	}{
		{30 * time.Second, "just now"},
		{-10 * time.Minute, "just now"},
		{time.Minute, "1 minute ago"},
		{59*time.Minute + 59*time.Second, "59 minutes ago"},
		{time.Hour + 29*time.Minute, "1 hour ago"},
		{time.Hour + 30*time.Minute, "2 hours ago"},
		{23*time.Hour + 29*time.Minute, "23 hours ago"},
		{23*time.Hour + 30*time.Minute, "1 day ago"},
		{24*time.Hour + 30*time.Minute, "1 day 1 hour ago"},
		{48*time.Hour + 49*time.Minute, "2 days 1 hour ago"},
		{48*time.Hour + 45*time.Minute, "2 days 1 hour ago"},
		{9*24*time.Hour + 23*time.Hour + 50*time.Minute, "10 days ago"},
	}
	for _, tt := range tests {
		if got := RoundedAgo(context.Background(), now.Add(-tt.before).UnixMilli(), now); got != tt.want {
			t.Errorf("RoundedAgo(context.Background(), %s before) = %q, want %q", tt.before, got, tt.want)
		}
	}
}

func TestDuration(t *testing.T) {
	for in, want := range map[int64]string{
		0:          "0 ms",
		250:        "250 ms",
		1500:       "1.5 s",
		59_940:     "59.9 s",
		125_000:    "2 min 5 s",
		120_000:    "2 min",
		12_000_000: "3 h 20 min",
		10_800_000: "3 h",
	} {
		if got := Duration(context.Background(), in); got != want {
			t.Errorf("Duration(context.Background(), %d) = %q, want %q", in, got, want)
		}
	}
}

func TestReadingType(t *testing.T) {
	for in, want := range map[string]string{"sgv": "Sensor", "mbg": "Meter", "cal": "cal"} {
		if got := ReadingType(context.Background(), in); got != want {
			t.Errorf("ReadingType(context.Background(), %q) = %q, want %q", in, got, want)
		}
	}
}

func TestDirection(t *testing.T) {
	str := func(s string) *string { return &s }
	tests := []struct {
		in   *string
		want string
	}{
		{nil, "-"},
		{str("Flat"), "→"},
		{str("FortyFiveDown"), "↘"},
		{str("DoubleUp"), "⇈"},
		{str("NOT COMPUTABLE"), "NOT COMPUTABLE"},
	}
	for _, tt := range tests {
		if got := Direction(tt.in); got != tt.want {
			t.Errorf("Direction(%v) = %q, want %q", tt.in, got, tt.want)
		}
	}
	// Beside a reading, a direction with no arrow, or none, shows nothing
	for in, want := range map[*string]string{nil: "", str("Flat"): "→", str("NOT COMPUTABLE"): ""} {
		if got := Arrow(in); got != want {
			t.Errorf("Arrow(%v) = %q, want %q", in, got, want)
		}
	}
}

func TestValueAndPerDay(t *testing.T) {
	number := func(v float64) *float64 { return &v }
	tests := []struct {
		name string
		got  string
		want string
	}{
		{"no value", Value(nil, " U"), "–"},
		{"a value", Value(number(12.5), " U"), "12.5 U"},
		{"rounded to 2 places", Value(number(1.23456), "%"), "1.23%"},
		{"a real 0", Value(number(0), " U"), "0 U"},
		{"no days with data", PerDay(0, 0, " U"), "–"},
		{"days with data and no insulin", PerDay(3, 0, " U"), "0 U"},
		{"days with data", PerDay(7, 31.256, " g"), "31.26 g"},
	}
	for _, test := range tests {
		if test.got != test.want {
			t.Errorf("%s: got %q, want %q", test.name, test.got, test.want)
		}
	}
}
