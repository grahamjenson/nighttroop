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

// A report's summaries are out of date when one is marked stale, or data before today in them was written since the
// site was last checked, or since it was made when it has not been; today, built as it is read, never is
func TestSummariesOutOfDate(t *testing.T) {
	_, controlDB, site := testutils.NewControlDBWithSite(t)
	ctx, now, q := t.Context(), time.Now(), site.SiteDB.Queries()
	siteContext := func() *httputils.SiteContext {
		model, err := controlDB.Queries.GetSiteInAnyState(ctx, controldb.GetSiteInAnyStateParams{ID: site.SiteDB.Subdomain})
		require.NoError(t, err)
		return &httputils.SiteContext{SiteID: site.SiteDB.Subdomain, ControlDB: controlDB, SiteDB: site.SiteDB, SiteModel: model}
	}
	today := sitedata.PeriodStart("day", now, siteContext().GetSiteTimezone(), "monday")
	week := &sitedb.SummaryPeriod{StartAt: today.AddDate(0, 0, -6).UnixMilli(), EndAt: today.AddDate(0, 0, 1).UnixMilli()}
	outOfDate := func(rows ...*sitedb.SummaryPeriod) bool {
		stale, err := sitedata.SummariesOutOfDate(ctx, q, siteContext(), rows, now)
		require.NoError(t, err)
		return stale
	}
	reading := func(id string, at time.Time) {
		_, err := q.CreateGlucoseReading(ctx, sitedb.CreateGlucoseReadingParams{ID: id, Mgdl: 120, ReadingType: "sgv", OccurredAt: at.UnixMilli(), CreatedByIdentityName: "phone"})
		require.NoError(t, err)
	}

	reading("today", now)
	require.False(t, outOfDate(week), "today's data is built as it is read")
	reading("late", today.AddDate(0, 0, -2))
	require.True(t, outOfDate(week), "a site not yet checked has had nothing built since it was made")

	require.NoError(t, controlDB.Queries.UpdateSiteSummariesCheckedAt(ctx, controldb.UpdateSiteSummariesCheckedAtParams{SummariesCheckedAt: new(time.Now().UnixMilli() + 1), ID: site.SiteDB.Subdomain}))
	require.False(t, outOfDate(week), "built once checked")
	require.True(t, outOfDate(week, &sitedb.SummaryPeriod{StaleAt: new(now.UnixMilli())}), "one of its days is marked stale")
	month := &sitedb.SummaryPeriod{StartAt: today.AddDate(0, -2, 0).UnixMilli(), EndAt: today.AddDate(0, -1, 0).UnixMilli()}
	reading("other-month", today.AddDate(0, 0, -3))
	require.False(t, outOfDate(month), "data written outside it")
}
