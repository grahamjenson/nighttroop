package sitedata

import (
	"bytes"
	"fmt"
	"log/slog"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/data/sql/sqltiming"
	"nighttroop.t1d.nz/models/sitemodels"
	"nighttroop.t1d.nz/server/siteapi/siteapiutils"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/httputils"
	"nighttroop.t1d.nz/utils/testutils"
)

// summarySite is a new site, as its PWD phone's identity reads its summary
func summarySite(t testing.TB) *httputils.SiteContext {
	_, controlDB, site := testutils.NewControlDBWithSite(t)
	id := site.SiteDB.Subdomain
	model, err := controlDB.Queries.GetSiteInAnyState(t.Context(), controldb.GetSiteInAnyStateParams{ID: id})
	require.NoError(t, err)
	identity, err := controlDB.Queries.GetSiteIdentity(t.Context(), controldb.GetSiteIdentityParams{SiteID: id, Name: consts.FIRST_PWD_NAME})
	require.NoError(t, err)
	return &httputils.SiteContext{SiteID: id, ControlDB: controlDB, SiteDB: site.SiteDB, SiteModel: model, IdentityModel: identity}
}

// tempBasal adds a temp basal of minutes at a rate that started at
func tempBasal(t testing.TB, q *sitedb.Queries, id string, at time.Time, minutes int64, rate float64) {
	_, err := q.CreateTempBasal(t.Context(), sitedb.CreateTempBasalParams{ID: id, DurationMinutes: minutes, Rate: rate, OccurredAt: at.UnixMilli(), CreatedByIdentityName: "test"})
	require.NoError(t, err)
}

// The summary's first minutes take the newest temp basal and temp target that started before it and are not deleted,
// which replaced the ones before them, and only that one of each is read, however many a loop set in the week before
func TestSummaryMinutesReadOnlyTheNewestEarlierTempBasalAndTarget(t *testing.T) {
	siteContext := summarySite(t)
	ctx, q, units := t.Context(), siteContext.SiteDB.Queries(), sitemodels.ContextUnits(siteContext)
	until := time.Now().Truncate(time.Minute)
	since := until.Add(-6 * time.Hour)
	// One of 10 hours, 9 U/hr, that would run through the summary had the loop's newer ones not replaced it
	tempBasal(t, q, "long", since.Add(-25*time.Hour), 600, 9)
	// A loop's day of temp basals of 30 minutes, 5 minutes apart: its last, 5 minutes before the summary, runs 25
	// minutes into it
	for i := range 288 {
		tempBasal(t, q, fmt.Sprint("loop", i), since.Add(-time.Duration(i+1)*5*time.Minute), 30, 0.5)
	}
	// A newer one, deleted, does not hide the loop's last
	tempBasal(t, q, "deleted", since.Add(-time.Minute), 600, 7)
	_, err := q.SoftDeleteTempBasal(ctx, sitedb.SoftDeleteTempBasalParams{ID: "deleted"})
	require.NoError(t, err)
	// A temp target of 10 hours, and a newer one of 30 minutes that replaced it 10 minutes before the summary
	for id, target := range map[string]sitedb.CreateTempTargetParams{
		"long":  {DurationMinutes: 600, LowTargetMgdl: 150, HighTargetMgdl: 160, OccurredAt: since.Add(-2 * time.Hour).UnixMilli()},
		"newer": {DurationMinutes: 30, LowTargetMgdl: 80, HighTargetMgdl: 90, OccurredAt: since.Add(-10 * time.Minute).UnixMilli()},
	} {
		target.ID, target.CreatedByIdentityName = id, "test"
		_, err := q.CreateTempTarget(ctx, target)
		require.NoError(t, err)
	}
	// One in the summary, an hour in for 10 minutes, is given, as its temp basals are
	given := []*sitemodels.SiteTempBasal{{OccurredAt: since.Add(time.Hour).UnixMilli(), DurationMinutes: 10, Rate: new(2.0)}}

	var logged bytes.Buffer
	testutils.UseLogHandler(t, slog.NewTextHandler(&logged, nil))
	sqltiming.SetSlowThreshold(time.Nanosecond)
	t.Cleanup(func() { sqltiming.SetSlowThreshold(0) })
	minutes, herr := summaryMinutes(ctx, q, since.UnixMilli(), until.UnixMilli(), units, time.UTC, siteContext, nil, given)
	sqltiming.SetSlowThreshold(0)
	require.Nil(t, herr)
	require.Len(t, minutes, 361)

	require.Equal(t, 0.5, *minutes[0].Basal, "the loop's last temp basal before the summary")
	require.Equal(t, 0.5, *minutes[24].Basal)
	require.Nil(t, minutes[25].Basal, "once it ends, no older temp basal comes back")
	require.Equal(t, 2.0, *minutes[60].Basal, "the temp basal in the summary")
	require.Nil(t, minutes[70].Basal)
	require.Equal(t, units.FromMGDL(80), *minutes[0].LowTarget, "the newer temp target before the summary")
	require.Equal(t, units.FromMGDL(90), *minutes[19].HighTarget)
	require.Nil(t, minutes[20].LowTarget, "once it ends, the temp target it replaced does not come back")

	read := regexp.MustCompile(`query=(SearchTemp(?:Targets|Basals)) ms=\d+ rows=(\d+)`).FindAllStringSubmatch(logged.String(), -1)
	require.Len(t, read, 2, logged.String())
	for _, query := range read {
		require.Equal(t, "1", query[2], "%s read more than the newest before the summary", query[1])
	}
}

// BenchmarkSummary is the summary of a site whose loop set a temp basal every 5 minutes for 8 days, as AAPS, Loop and
// Trio do, with its readings, device statuses and SMBs: the home page's, read again by each open page on each change,
// and its minutes alone.
// Run with go test -run XXX -bench Summary -benchmem ./server/siteapi/sitecollections/sitedata
func BenchmarkSummary(b *testing.B) {
	siteContext := summarySite(b)
	ctx, q, now := b.Context(), siteContext.SiteDB.Queries(), time.Now()
	for i := range 8 * 288 {
		at := now.Add(-time.Duration(i) * 5 * time.Minute)
		tempBasal(b, q, fmt.Sprint("basal", i), at, 30, float64(i%10)/10)
		_, err := q.CreateGlucoseReading(ctx, sitedb.CreateGlucoseReadingParams{ID: fmt.Sprint("reading", i), Mgdl: float64(100 + i%80),
			ReadingType: "sgv", OccurredAt: at.UnixMilli(), CreatedByDeviceName: new("AndroidAPS-DexcomG6"), CreatedByIdentityName: "test"})
		require.NoError(b, err)
		_, err = q.CreateDeviceReport(ctx, sitedb.CreateDeviceReportParams{ID: fmt.Sprint("status", i), LoopName: new("openaps"), LoopIob: new(0.7),
			LoopCob: new(12.0), OccurredAt: at.UnixMilli(), CreatedByDeviceName: new("openaps://Test Phone"), CreatedByIdentityName: "test"})
		require.NoError(b, err)
		if i%3 == 0 {
			_, err = q.CreateInsulinDose(ctx, sitedb.CreateInsulinDoseParams{ID: fmt.Sprint("smb", i), Dose: 0.2, Automatic: new(true), OccurredAt: at.UnixMilli(), CreatedByIdentityName: "test"})
			require.NoError(b, err)
		}
	}
	// As the daily job keeps them, so the page builds only today
	_, err := RebuildSummaryDaily(ctx, q, siteContext, LocalDate(now, siteContext.GetSiteTimezone(), -1), now)
	require.NoError(b, err)
	require.NoError(b, BuildLastSummaryPeriods(ctx, q, siteContext, now))

	b.Run("page", func(b *testing.B) {
		for b.Loop() {
			tx := siteapiutils.NewCollectionTx(ctx, siteContext)
			summary, herr := getSummary(ctx, tx, defaultHours, siteContext)
			tx.Rollback()
			if herr != nil || len(summary.Minutes) != defaultHours*60+1 {
				b.Fatal(herr, len(summary.Minutes))
			}
		}
	})
	b.Run("minutes", func(b *testing.B) {
		since := now.Add(-defaultHours * time.Hour).UnixMilli()
		for b.Loop() {
			if _, herr := summaryMinutes(ctx, q, since, now.UnixMilli(), new(consts.UnitsMGDL), time.UTC, siteContext, nil, nil); herr != nil {
				b.Fatal(herr)
			}
		}
	})
}
