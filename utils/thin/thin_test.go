package thin

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// A report is kept whatever for its first hour, then is in a 10 minute bucket until a week old, a 20 minute one until
// a fortnight old, and an hour's after; at each boundary it is in the younger tier, as the truncation's query has it
func TestBucket(t *testing.T) {
	now := time.Date(2026, 11, 25, 12, 0, 0, 0, time.UTC).UnixMilli()
	ago := func(d time.Duration) int64 { return now - d.Milliseconds() }
	tests := []struct {
		name string
		at   int64
		want time.Duration
	}{
		{"just now", now, 0},
		{"an hour ago", ago(time.Hour), 0},
		{"just over an hour ago", ago(time.Hour) - 1, 10 * time.Minute},
		{"a week ago", ago(7 * 24 * time.Hour), 10 * time.Minute},
		{"just over a week ago", ago(7*24*time.Hour) - 1, 20 * time.Minute},
		{"a fortnight ago", ago(14 * 24 * time.Hour), 20 * time.Minute},
		{"just over a fortnight ago", ago(14*24*time.Hour) - 1, time.Hour},
		{"30 days ago", ago(30 * 24 * time.Hour), time.Hour},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want.Milliseconds(), Bucket(tt.at, now))
		})
	}
}
