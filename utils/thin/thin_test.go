package thin

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// A report is kept whatever for its first hour, then is in a 10 minute bucket; at the boundary it is still kept, as
// the truncation's query has it
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
		{"6 days ago", ago(6 * 24 * time.Hour), 10 * time.Minute},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want.Milliseconds(), Bucket(tt.at, now))
		})
	}
}
