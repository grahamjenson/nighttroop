package backup_sites

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// The job looks for sites due a backup each minute, whatever the interval, so a due site waits at most a minute; an
// interval under a minute, as in tests, is checked each interval, and never more than each second
func TestTick(t *testing.T) {
	for interval, want := range map[time.Duration]time.Duration{
		time.Hour:        time.Minute,
		6 * time.Hour:    time.Minute,
		time.Minute:      time.Minute,
		30 * time.Second: 30 * time.Second,
		0:                time.Second,
	} {
		require.Equal(t, want, tick(interval), interval.String())
	}
}
