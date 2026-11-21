package flow

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// MIDDAY_TIMEZONE's time zone is always at least 5 hours from its midnight, whatever the hour UTC, in New Zealand's
// summer and winter, so a test of today's data never runs across it
func TestMiddayTimezone(t *testing.T) {
	auckland, err := time.LoadLocation("Pacific/Auckland")
	require.NoError(t, err)
	for _, day := range []time.Time{time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC), time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC)} {
		for minute := 0; minute < 24*60; minute += 15 {
			now := day.Add(time.Duration(minute) * time.Minute)
			location := time.UTC
			if middayTimezone(now) == "Pacific/Auckland" {
				location = auckland
			}
			local := now.In(location)
			sinceMidnight := local.Sub(time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, location))
			require.True(t, sinceMidnight >= 5*time.Hour && sinceMidnight <= 19*time.Hour, "%s is %s in %s", now, local.Format("15:04"), location)
		}
	}
}

// date, daytime and dayms are of the site's days: UTC, or MIDDAY_TIMEZONE's
func TestTemplateDayTimeInTheSitesTimezone(t *testing.T) {
	utc, err := templateDayTime(map[string]any{}, 0, 0)
	require.NoError(t, err)
	require.Equal(t, time.UTC, utc.Location())
	zoned, err := templateDayTime(map[string]any{"site_timezone": "Pacific/Auckland"}, -1, 720)
	require.NoError(t, err)
	require.Equal(t, "Pacific/Auckland", zoned.Location().String())
	require.Equal(t, 12, zoned.Hour())
	require.Equal(t, time.Now().In(zoned.Location()).AddDate(0, 0, -1).Format(time.DateOnly), zoned.Format(time.DateOnly))
}
