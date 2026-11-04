package push

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// Pushes are counted by UTC day, today's and this week's from Monday, and days over a week old are dropped
func TestCounts(t *testing.T) {
	counts.Lock()
	counts.days = map[int64]Count{}
	counts.Unlock()
	wednesday := time.Date(2026, 11, 4, 12, 0, 0, 0, time.UTC)
	countPush(wednesday.AddDate(0, 0, -9), nil) // the Monday a week before last, dropped
	countPush(wednesday.AddDate(0, 0, -3), nil) // last Sunday, before this week
	countPush(wednesday.AddDate(0, 0, -2), nil) // Monday
	countPush(wednesday.AddDate(0, 0, -2), errors.New("refused"))
	countPush(wednesday, nil)
	require.Equal(t, Counts{Today: Count{Sent: 1}, Week: Count{Sent: 3, Failed: 1}}, CountsAt(wednesday))
	// The next Monday starts a new week
	require.Equal(t, Counts{}, CountsAt(wednesday.AddDate(0, 0, 5)))
	counts.Lock()
	require.Len(t, counts.days, 3)
	counts.Unlock()
}
