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
