package sitedata

import (
	"bytes"
	"fmt"
	"log/slog"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/data/sql/sqltiming"
	"nighttroop.t1d.nz/utils/testutils"
)

// A day's summary has the share of an extended dose that started before the day and runs into it, and of the week
// before the day it reads only such a dose: not a loop's SMBs, a dose whose extended part ended before the day, nor a
// deleted one
func TestSummaryDailyReadsOnlyTheExtendedDosesFromBeforeTheDay(t *testing.T) {
	siteContext := summarySite(t)
	ctx, q, now := t.Context(), siteContext.SiteDB.Queries(), time.Now()
	date := LocalDate(now, siteContext.GetSiteTimezone(), -1)
	midnight, err := time.ParseInLocation(time.DateOnly, date, siteContext.GetSiteTimezone())
	require.NoError(t, err)
	dose := func(id string, at time.Time, units float64, extended *float64, extendedFor time.Duration) {
		params := sitedb.CreateInsulinDoseParams{ID: id, Dose: units, ExtendedDose: extended, OccurredAt: at.UnixMilli(), CreatedByIdentityName: "test"}
		if extended != nil {
			params.ExtendedDurationMs = new(extendedFor.Milliseconds())
		}
		_, err := q.CreateInsulinDose(ctx, params)
		require.NoError(t, err)
	}
	// A loop's day of SMBs before the day, 5 minutes apart
	for i := range 288 {
		dose(fmt.Sprint("smb", i), midnight.Add(-time.Duration(i+1)*5*time.Minute), 0.1, nil, 0)
	}
	// 4 U over 2 hours from an hour before the day gives the day 2 U, and its 1 U at once is the day before's
	dose("over midnight", midnight.Add(-time.Hour), 1, new(4.0), 2*time.Hour)
	// The longest extended part, a week, from a week less a day before the day, gives the day 1 U of its 7 U
	dose("a week", midnight.AddDate(0, 0, -6), 0, new(7.0), 7*24*time.Hour)
	dose("ended before the day", midnight.Add(-3*time.Hour), 0, new(2.0), 2*time.Hour)
	dose("deleted", midnight.Add(-time.Hour), 0, new(4.0), 2*time.Hour)
	_, err = q.SoftDeleteInsulinDose(ctx, sitedb.SoftDeleteInsulinDoseParams{ID: "deleted"})
	require.NoError(t, err)
	// In the day: a bolus of 3 U, and a combo bolus of 1 U at once and 1 U over an hour
	dose("bolus", midnight.Add(12*time.Hour), 3, nil, 0)
	dose("combo", midnight.Add(13*time.Hour), 1, new(1.0), time.Hour)

	var logged bytes.Buffer
	testutils.UseLogHandler(t, slog.NewTextHandler(&logged, nil))
	sqltiming.SetSlowThreshold(time.Nanosecond)
	t.Cleanup(func() { sqltiming.SetSlowThreshold(0) })
	day, err := buildSummaryDaily(ctx, q, siteContext, date, now.UnixMilli())
	sqltiming.SetSlowThreshold(0)
	require.NoError(t, err)
	require.Equal(t, 8.0, day.TotalDailyDose, "2 U and 1 U from before the day, and 3 U and 2 U in it")

	read := regexp.MustCompile(`query=(SearchInsulinDoses|InsulinDosesExtendedInto) ms=\d+ rows=(\d+)`).FindAllStringSubmatch(logged.String(), -1)
	require.Len(t, read, 2, logged.String())
	require.Equal(t, []string{"SearchInsulinDoses", "2"}, read[0][1:], "the day's doses")
	require.Equal(t, []string{"InsulinDosesExtendedInto", "2"}, read[1][1:], "the doses still running into the day")
}
