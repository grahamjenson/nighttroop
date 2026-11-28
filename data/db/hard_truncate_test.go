package db_test

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/data/sql/logsdb"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/utils/testutils"
)

// A site over 1.5 times its DB size limit (20 MB, so over 30720 KB) has its device reports truncated harder by each
// backup: the JSON of those over a day old is cleared. No other record is
// touched, however old: a reading or a treatment is deleted only once its day is packaged. A site under 1.5 times its
// limit gets only the job's usual truncate.
func TestTruncateSitesTruncatesSitesFarOverTheirLimitHarder(t *testing.T) {
	tests := []struct {
		name string
		// sizesKB are the site's sizes, as measured before and after
		sizesKB    []int64
		clearsJSON bool
	}{
		{name: "under 1.5 times its limit", sizesKB: []int64{30_000}},
		{name: "over 1.5 times its limit", sizesKB: []int64{40_000, 35_000}, clearsJSON: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cfg, controlDB, site := testutils.NewControlDBWithSite(t)
			// A server that does not delete packaged rows, so only the size limit deletes here
			cfg.RetentionDeleteRows = false
			t.Cleanup(func() { cfg.RetentionDeleteRows = true })
			ctx, q, siteID := t.Context(), site.SiteDB.Queries(), site.SiteDB.Subdomain
			now, day := time.Now(), 24*time.Hour
			ago := func(age time.Duration) int64 { return now.Add(-age).UnixMilli() }
			status := `{"iob":{"iob":1.2}}`
			report := func(id string, occurredAt int64) {
				_, err := q.CreateDeviceReport(ctx, sitedb.CreateDeviceReportParams{ID: id, LoopStatusJson: &status, UploaderJson: &status,
					OverrideJson: &status, LoopIob: new(1.2), OccurredAt: occurredAt, CreatedByDeviceName: new("aaps"), CreatedByIdentityName: "phone"})
				require.NoError(t, err)
			}
			report("report-2d", ago(2*day))
			report("report-10d", ago(10*day))
			report("report-newest", ago(time.Minute))
			dose := func(id string, occurredAt int64) {
				_, err := q.CreateInsulinDose(ctx, sitedb.CreateInsulinDoseParams{ID: id, Dose: 1, IcfgJson: &status, OccurredAt: occurredAt, CreatedByIdentityName: "phone"})
				require.NoError(t, err)
			}
			profileSwitch := func(id string, occurredAt int64) {
				_, err := q.CreateProfileSwitch(ctx, sitedb.CreateProfileSwitchParams{ID: id, Name: "Default", IcfgJson: &status, BasalRatesJson: new(`[{"time":"00:00","value":1}]`),
					OccurredAt: occurredAt, CreatedByIdentityName: "phone"})
				require.NoError(t, err)
			}
			group := func(id, groupType string, occurredAt int64, params sitedb.CreateEventGroupParams) {
				params.ID, params.GroupType, params.OccurredAt, params.CreatedByIdentityName = id, groupType, occurredAt, "phone"
				params.BolusCalculatorResultJson = &status
				_, err := q.CreateEventGroup(ctx, params)
				require.NoError(t, err)
			}
			reading := func(id, readingType string, occurredAt int64) {
				_, err := q.CreateGlucoseReading(ctx, sitedb.CreateGlucoseReadingParams{ID: id, Mgdl: 120, ReadingType: readingType, OccurredAt: occurredAt, CreatedByIdentityName: "phone"})
				require.NoError(t, err)
			}
			dose("dose-old", ago(100*day))
			dose("dose-recent", ago(2*day))
			dose("dose-ungrouped-old", ago(100*day))
			group("bolus-old", "Bolus", ago(100*day), sitedb.CreateEventGroupParams{InsulinDoseID: new("dose-old")})
			group("bolus-deleted", "Bolus", ago(100*day), sitedb.CreateEventGroupParams{})
			group("bolus-recent", "Bolus", ago(2*day), sitedb.CreateEventGroupParams{InsulinDoseID: new("dose-recent")})
			group("sensor-old", "Sensor Change", ago(100*day), sitedb.CreateEventGroupParams{})
			profileSwitch("switch-old", ago(100*day))
			group("profile-old", "Profile Switch", ago(100*day), sitedb.CreateEventGroupParams{ProfileSwitchID: new("switch-old")})
			reading("sgv-old", "sgv", ago(100*day))
			reading("sgv-deleted", "sgv", ago(100*day))
			reading("sgv-recent", "sgv", ago(time.Hour))
			reading("mbg-old", "mbg", ago(100*day))
			_, err := q.SoftDeleteEventGroup(ctx, sitedb.SoftDeleteEventGroupParams{ID: "bolus-deleted", UpdatedByIdentityName: new("phone")})
			require.NoError(t, err)
			_, err = q.SoftDeleteGlucoseReading(ctx, sitedb.SoftDeleteGlucoseReadingParams{ID: "sgv-deleted", UpdatedByIdentityName: new("phone")})
			require.NoError(t, err)

			measured := 0
			site.SiteDB.SetSizeKB(func() *int64 {
				measured++
				return new(test.sizesKB[min(measured, len(test.sizesKB))-1])
			})
			require.NoError(t, controlDB.TruncateSites(ctx, controlDB.AllSites(), nil))

			gone := func(err error) bool {
				if errors.Is(err, sql.ErrNoRows) {
					return true
				}
				require.NoError(t, err)
				return false
			}
			getReport := func(id string) *sitedb.DeviceReport {
				report, err := q.GetDeviceReport(ctx, sitedb.GetDeviceReportParams{ID: id})
				require.NoError(t, err)
				return report
			}
			getDose := func(id string) (*sitedb.InsulinDose, error) {
				return q.GetInsulinDose(ctx, sitedb.GetInsulinDoseParams{ID: id, IncludeDeleted: true})
			}
			getGroup := func(id string) (*sitedb.GetEventGroupRow, error) {
				return q.GetEventGroup(ctx, sitedb.GetEventGroupParams{ID: id, IncludeDeleted: true})
			}
			readingGone := func(id string) bool {
				_, err := q.GetGlucoseReading(ctx, sitedb.GetGlucoseReadingParams{ID: id, IncludeDeleted: true})
				return gone(err)
			}

			// Its device reports
			require.Equal(t, test.clearsJSON, getReport("report-2d").LoopStatusJson == nil, "a report over a day old has its loop status cleared")
			require.Equal(t, test.clearsJSON, getReport("report-2d").UploaderJson == nil, "and its uploader")
			require.Equal(t, new(1.2), getReport("report-2d").LoopIob, "its IOB is kept")
			require.NotNil(t, getReport("report-newest").LoopStatusJson, "the device's newest report keeps its JSON")
			_, err = q.GetDeviceReport(ctx, sitedb.GetDeviceReportParams{ID: "report-10d"})
			require.True(t, gone(err), "a report over 7 days old is deleted, as it is for every site")
			recent, err := getDose("dose-recent")
			require.NoError(t, err)
			require.NotNil(t, recent.IcfgJson, "a dose keeps its insulin")
			bolus, err := getGroup("bolus-recent")
			require.NoError(t, err)
			require.NotNil(t, bolus.EventGroup.BolusCalculatorResultJson, "a group keeps its bolus calculation")
			newestSwitch, err := q.GetProfileSwitch(ctx, sitedb.GetProfileSwitchParams{ID: "switch-old"})
			require.NoError(t, err)
			require.NotNil(t, newestSwitch.IcfgJson, "the newest profile switch keeps its insulin")
			require.NotNil(t, newestSwitch.BasalRatesJson, "a profile switch keeps its profile")

			// No other record is deleted, however old, as this server does not delete packaged rows
			for _, id := range []string{"bolus-old", "bolus-deleted", "sensor-old", "profile-old", "bolus-recent"} {
				_, err := getGroup(id)
				require.False(t, gone(err), "%s is kept", id)
			}
			for _, id := range []string{"dose-old", "dose-ungrouped-old"} {
				_, err := getDose(id)
				require.False(t, gone(err), "%s is kept", id)
			}
			for _, id := range []string{"sgv-old", "sgv-deleted", "mbg-old"} {
				require.False(t, readingGone(id), "%s is kept", id)
			}

			// What changed is recorded, with the size after, in a "Site truncated" event
			truncated, err := controlDB.Logs.Queries.SearchEvents(ctx, logsdb.SearchEventsParams{SiteID: &siteID, Type: new("site_truncated"), Limit: 10})
			require.NoError(t, err)
			model, err := controlDB.Queries.GetSiteInAnyState(ctx, controldb.GetSiteInAnyStateParams{ID: siteID})
			require.NoError(t, err)
			if !test.clearsJSON {
				require.Empty(t, truncated)
				require.Nil(t, model.EstimatedDbSizeKb)
				return
			}
			final := test.sizesKB[len(test.sizesKB)-1]
			require.Len(t, truncated, 1)
			require.JSONEq(t, fmt.Sprintf(`{"by":"size limit","before_size_kb":40000,"after_size_kb":%d}`, final), *truncated[0].Attributes)
			require.Equal(t, &final, model.EstimatedDbSizeKb)

			// A run that changes nothing more logs nothing
			measured = 0
			require.NoError(t, controlDB.TruncateSites(ctx, controlDB.AllSites(), nil))
			truncated, err = controlDB.Logs.Queries.SearchEvents(ctx, logsdb.SearchEventsParams{SiteID: &siteID, Type: new("site_truncated"), Limit: 10})
			require.NoError(t, err)
			require.Len(t, truncated, 1)
		})
	}
}

// A shrink writes the whole DB again, so a step of the harder truncate shrinks the DB only if a tenth of it or more is
// free pages, which the next writes fill otherwise. A site that stays over its limit changes a few records at each
// backup, and was shrunk at each one.
func TestTruncateSitesShrinksOnlyADBWithManyFreePages(t *testing.T) {
	tests := []struct {
		name       string
		oldReports int
		shrunk     bool
	}{
		{name: "a few records deleted", oldReports: 5},
		{name: "many records deleted", oldReports: 400, shrunk: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, controlDB, site := testutils.NewControlDBWithSite(t)
			ctx, q := t.Context(), site.SiteDB.Queries()
			status := fmt.Sprintf(`{"note":%q}`, strings.Repeat("x", 4000))
			old := time.Now().Add(-10 * 24 * time.Hour).UnixMilli()
			for i := range test.oldReports {
				_, err := q.CreateDeviceReport(ctx, sitedb.CreateDeviceReportParams{ID: fmt.Sprintf("report-%d", i), LoopStatusJson: &status,
					OccurredAt: old + int64(i), CreatedByDeviceName: new("aaps"), CreatedByIdentityName: "phone"})
				require.NoError(t, err)
			}
			// Far over 1.5 times its limit before the first step, also without the pages the job's usual truncate
			// frees as it thins the reports, and under after it
			measured := 0
			site.SiteDB.SetSizeKB(func() *int64 {
				measured++
				return new([]int64{400_000, 30_000}[min(measured, 2)-1])
			})
			require.NoError(t, controlDB.TruncateSites(ctx, controlDB.AllSites(), nil))

			_, err := q.GetDeviceReport(ctx, sitedb.GetDeviceReportParams{ID: "report-0"})
			require.ErrorIs(t, err, sql.ErrNoRows, "the reports over 7 days old are deleted")
			free, err := site.SiteDB.FreePart(ctx)
			require.NoError(t, err)
			if test.shrunk {
				require.Zero(t, free, "a DB with a tenth or more of it free is shrunk")
				return
			}
			require.Greater(t, free, 0.0, "a DB with less than a tenth of it free keeps its free pages")
			require.Less(t, free, 0.1)
		})
	}
}

// The job's usual truncate shrinks a site within its size limit too, when a tenth or more of its DB is free pages, as
// after its old rows are deleted: nothing else gives its disk space back
func TestTruncateSitesShrinksASiteWithinItsLimit(t *testing.T) {
	_, controlDB, site := testutils.NewControlDBWithSite(t)
	ctx, q := t.Context(), site.SiteDB.Queries()
	status := fmt.Sprintf(`{"note":%q}`, strings.Repeat("x", 4000))
	// Device reports over 30 days old, which the usual truncate deletes
	old := time.Now().Add(-40 * 24 * time.Hour).UnixMilli()
	for i := range 400 {
		_, err := q.CreateDeviceReport(ctx, sitedb.CreateDeviceReportParams{ID: fmt.Sprintf("report-%d", i), LoopStatusJson: &status,
			OccurredAt: old + int64(i), CreatedByDeviceName: new("aaps"), CreatedByIdentityName: "phone"})
		require.NoError(t, err)
	}
	require.NoError(t, controlDB.TruncateSites(ctx, controlDB.AllSites(), nil))
	_, err := q.GetDeviceReport(ctx, sitedb.GetDeviceReportParams{ID: "report-0"})
	require.ErrorIs(t, err, sql.ErrNoRows)
	free, err := site.SiteDB.FreePart(ctx)
	require.NoError(t, err)
	require.Zero(t, free, "the pages of the deleted rows are given back")
}
