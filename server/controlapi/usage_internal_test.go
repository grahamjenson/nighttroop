package controlapi

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// The day, week and month start at midnight, the Monday and the 1st in UTC, whatever the day and the time zone
func TestPeriodStarts(t *testing.T) {
	utc := func(day int, hour int) time.Time { return time.Date(2026, 10, day, hour, 30, 0, 0, time.UTC) }
	for _, test := range []struct {
		name   string
		now    time.Time
		monday int
	}{
		{"a Monday", utc(26, 9), 26},
		{"a Wednesday", utc(28, 23), 26},
		{"a Sunday", utc(25, 12), 19},
		{"a week begun last month, on Thursday the 1st", utc(1, 12), -2},
		{"a Saturday", utc(31, 0), 26},
		{"an hour ahead of UTC on Tuesday is still Monday in UTC", time.Date(2026, 10, 27, 0, 30, 0, 0, time.FixedZone("UTC+1", 3600)), 26},
	} {
		t.Run(test.name, func(t *testing.T) {
			midnight, monday, firstOfMonth := periodStarts(test.now)
			day := test.now.UTC()
			require.Equal(t, time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, time.UTC), midnight)
			require.Equal(t, time.Monday, monday.Weekday())
			require.Equal(t, time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, test.monday-1), monday)
			require.False(t, monday.After(midnight))
			require.Equal(t, time.Date(day.Year(), day.Month(), 1, 0, 0, 0, 0, time.UTC), firstOfMonth)
		})
	}
}
