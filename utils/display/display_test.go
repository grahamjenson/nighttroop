package display

import (
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
			if got := Ago(tt.ms, now); got != tt.want {
				t.Errorf("Ago() = %q, want %q", got, tt.want)
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
		if got := TimeAgo(now.Add(-tt.before).UnixMilli(), now); got != tt.want {
			t.Errorf("TimeAgo(%s before) = %q, want %q", tt.before, got, tt.want)
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
		12_000_000: "3 h 20 min",
	} {
		if got := Duration(in); got != want {
			t.Errorf("Duration(%d) = %q, want %q", in, got, want)
		}
	}
}

func TestReadingType(t *testing.T) {
	for in, want := range map[string]string{"sgv": "Sensor", "mbg": "Meter", "cal": "cal"} {
		if got := ReadingType(in); got != want {
			t.Errorf("ReadingType(%q) = %q, want %q", in, got, want)
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
}

func TestName(t *testing.T) {
	for in, want := range map[string]string{"site_created": "Site created", "control_backed_up": "Control backed up", "active": "Active", "": ""} {
		if got := Name(in); got != want {
			t.Errorf("Name(%q) = %q, want %q", in, got, want)
		}
	}
}
