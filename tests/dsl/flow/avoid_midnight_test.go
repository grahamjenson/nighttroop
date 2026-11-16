package flow

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// AVOID_MIDNIGHT waits only from 2 minutes before midnight UTC to its minutes after, and then until those minutes past
func TestMidnightWait(t *testing.T) {
	at := func(hour, minute, second int) time.Time { return time.Date(2026, 11, 16, hour, minute, second, 0, time.UTC) }
	tests := []struct {
		name string
		now  time.Time
		wait time.Duration
	}{
		{"midday", at(12, 0, 0), 0},
		{"3 minutes before midnight", at(23, 57, 0), 0},
		{"2 minutes before midnight", at(23, 58, 0), 17 * time.Minute},
		{"30 seconds before midnight", at(23, 59, 30), 15*time.Minute + 30*time.Second},
		{"midnight", at(0, 0, 0), 15 * time.Minute},
		{"2 minutes after", at(0, 2, 0), 13 * time.Minute},
		{"14 minutes 59 after", at(0, 14, 59), time.Second},
		{"15 minutes after", at(0, 15, 0), 0},
		{"an hour after", at(1, 0, 0), 0},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.wait, midnightWait(test.now, 15*time.Minute))
		})
	}
}
