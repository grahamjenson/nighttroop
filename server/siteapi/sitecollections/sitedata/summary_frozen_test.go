package sitedata_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/server/siteapi/sitecollections/sitedata"
	"nighttroop.t1d.nz/utils/httputils"
	"nighttroop.t1d.nz/utils/testutils"
)

// A summary that ended over a month ago, and was built while it had all its data, is frozen: no rebuild, timezone or
// week start change, record changed in it, or delete, builds it again or deletes it, so it is kept when its data is gone
func TestSummariesBuiltWithAllTheirDataAreFrozenAfterAMonth(t *testing.T) {
	_, controlDB, site := testutils.NewControlDBWithSite(t)
	ctx, now, q := t.Context(), time.Now(), site.SiteDB.Queries()
	model, err := controlDB.Queries.GetSiteInAnyState(ctx, controldb.GetSiteInAnyStateParams{ID: site.SiteDB.Subdomain})
	require.NoError(t, err)
	siteContext := &httputils.SiteContext{SiteID: site.SiteDB.Subdomain, ControlDB: controlDB, SiteDB: site.SiteDB, SiteModel: model}
	location := siteContext.GetSiteTimezone()
	day := func(daysAgo int) time.Time { return sitedata.PeriodStart("day", now.AddDate(0, 0, -daysAgo), location, "monday") }
	// build builds the day with a reading in it, as if builtAfter its end
	build := func(daysAgo int, builtAfter time.Duration) {
		at := day(daysAgo).Add(12 * time.Hour)
		_, err := q.CreateGlucoseReading(ctx, sitedb.CreateGlucoseReadingParams{ID: at.Format(time.DateOnly), Mgdl: 120, ReadingType: "sgv", OccurredAt: at.UnixMilli(), CreatedByIdentityName: "phone"})
		require.NoError(t, err)
		row, err := sitedata.RebuildSummaryDaily(ctx, q, siteContext, day(daysAgo).Format(time.DateOnly), now)
		require.NoError(t, err)
		_, err = site.SiteDB.SQLiteDB.ExecContext(ctx, "UPDATE summary_period SET updated_at = ? WHERE period = 'day' AND date = ?", row.EndAt+builtAfter.Milliseconds(), row.Date)
		require.NoError(t, err)
	}
	get := func(daysAgo int) *sitedb.SummaryPeriod {
		row, err := q.GetSummaryPeriod(ctx, sitedb.GetSummaryPeriodParams{Period: "day", Date: day(daysAgo).Format(time.DateOnly)})
		if err != nil {
			return nil
		}
		return row
	}

	build(40, time.Hour)        // Built the night after: it had all its data
	build(41, 35*24*time.Hour) // Built over a month after it, as from a backfill: its device reports were already gone
	build(10, time.Hour)        // Not a month old
	require.NoError(t, sitedata.FreezeSummaries(ctx, q, now))
	frozen := get(40)
	require.NotNil(t, frozen.FrozenAt, "a month old and built with all its data")
	require.Nil(t, get(41).FrozenAt, "built after its data was gone")
	require.Nil(t, get(10).FrozenAt, "not a month old")

	// Its raw data is deleted, as truncation will. Nothing builds it again or deletes it.
	_, err = site.SiteDB.SQLiteDB.ExecContext(ctx, "DELETE FROM glucose_readings")
	require.NoError(t, err)
	require.NoError(t, sitedata.RebuildSummaryYear(ctx, q, siteContext, now))
	require.Equal(t, frozen, get(40), "Regenerate summaries keeps it as it was")
	require.Nil(t, get(41), "one not frozen is built again from its data, here none")
	require.Nil(t, get(10))
	require.NoError(t, sitedata.RebuildSummaryDailyWeek(ctx, q, siteContext, now))
	require.NoError(t, sitedata.RebuildSummaryWeeks(ctx, q, siteContext, now))
	require.NoError(t, sitedata.RebuildSummaryDays(ctx, q, siteContext, []time.Time{day(40).Add(time.Hour)}, now))
	require.NoError(t, q.MarkSummariesStale(ctx, sitedb.MarkSummariesStaleParams{At: day(40).Add(time.Hour).UnixMilli()}))
	_, err = q.HardDeleteSummaryPeriod(ctx, sitedb.HardDeleteSummaryPeriodParams{Period: "day", Date: frozen.Date})
	require.NoError(t, err)
	require.Equal(t, frozen, get(40), "a timezone or week start change, a record changed in it, or a delete, keeps it as it was")
}
