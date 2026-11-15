package db_test

import (
	"database/sql"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/data/sql/logsdb"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/utils/testutils"
)

// A site over 1.5 times its DB size limit (20 MB, so over 30720 KB) is truncated harder by the job every 10 minutes:
// first the JSON that matters only while recent is cleared and device reports over 7 days old are deleted; then, only
// if it is still over, readings and treatments over 90 days old are deleted, but the newest of each type, and its
// summaries are kept. A site under 1.5 times its limit gets only the job's usual truncate.
func TestTruncateSitesTruncatesSitesFarOverTheirLimitHarder(t *testing.T) {
	tests := []struct {
		name string
		// sizesKB are the site's sizes, as measured before, after the first step and at the end
		sizesKB                       []int64
		clearsJSON, deletesOldRecords bool
	}{
		{name: "under 1.5 times its limit", sizesKB: []int64{30_000}},
		{name: "over, and under after the first step", sizesKB: []int64{40_000, 30_000, 30_000}, clearsJSON: true},
		{name: "over, and still over after the first step", sizesKB: []int64{40_000, 35_000, 25_000}, clearsJSON: true, deletesOldRecords: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, controlDB, site := testutils.NewControlDBWithSite(t)
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
			oldDate := now.Add(-100 * day).Format(time.DateOnly)
			_, err = q.UpsertSummaryPeriod(ctx, sitedb.UpsertSummaryPeriodParams{Period: "day", Date: oldDate, Timezone: "UTC", StartAt: ago(100 * day), EndAt: ago(99 * day),
				Days: 1, TotalDailyDose: 30, GlucoseMgdlJson: "[]", LoopIobJson: "[]", LoopIobBasalJson: "[]", LoopIobTreatmentJson: "[]", LoopCobJson: "[]"})
			require.NoError(t, err)

			measured := 0
			site.SiteDB.SetSizeKB(func() *int64 {
				measured++
				return new(test.sizesKB[min(measured, len(test.sizesKB))-1])
			})
			require.NoError(t, controlDB.TruncateSites(ctx))

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

			// The first step
			require.Equal(t, test.clearsJSON, getReport("report-2d").LoopStatusJson == nil, "a report over a day old has its loop status cleared")
			require.Equal(t, test.clearsJSON, getReport("report-2d").UploaderJson == nil, "and its uploader")
			require.Equal(t, new(1.2), getReport("report-2d").LoopIob, "its IOB is kept")
			require.NotNil(t, getReport("report-newest").LoopStatusJson, "the device's newest report keeps its JSON")
			_, err = q.GetDeviceReport(ctx, sitedb.GetDeviceReportParams{ID: "report-10d"})
			require.Equal(t, test.clearsJSON, gone(err), "a report over 7 days old is deleted")
			recent, err := getDose("dose-recent")
			require.NoError(t, err)
			require.Equal(t, test.clearsJSON, recent.IcfgJson == nil, "a dose over a day old has its insulin cleared")
			bolus, err := getGroup("bolus-recent")
			require.NoError(t, err)
			require.Equal(t, test.clearsJSON, bolus.EventGroup.BolusCalculatorResultJson == nil, "a group over a day old has its bolus calculation cleared")
			newestSwitch, err := q.GetProfileSwitch(ctx, sitedb.GetProfileSwitchParams{ID: "switch-old"})
			require.NoError(t, err)
			require.NotNil(t, newestSwitch.IcfgJson, "the newest profile switch keeps its insulin")
			require.NotNil(t, newestSwitch.BasalRatesJson, "a profile switch keeps its profile")

			// The second step
			for _, id := range []string{"bolus-old", "bolus-deleted"} {
				_, err := getGroup(id)
				require.Equal(t, test.deletesOldRecords, gone(err), "%s, a group over 90 days old, is deleted", id)
			}
			for _, id := range []string{"dose-old", "dose-ungrouped-old"} {
				_, err := getDose(id)
				require.Equal(t, test.deletesOldRecords, gone(err), "%s, a dose over 90 days old, is deleted", id)
			}
			require.Equal(t, test.deletesOldRecords, readingGone("sgv-old"), "a reading over 90 days old is deleted")
			require.Equal(t, test.deletesOldRecords, readingGone("sgv-deleted"), "a deleted one too")
			require.False(t, readingGone("mbg-old"), "the newest of its reading type is kept")
			for _, id := range []string{"sensor-old", "profile-old", "bolus-recent"} {
				_, err := getGroup(id)
				require.False(t, gone(err), "%s, the newest of its type, is kept", id)
			}
			summary, err := q.GetSummaryPeriod(ctx, sitedb.GetSummaryPeriodParams{Period: "day", Date: oldDate})
			require.NoError(t, err, "the summaries are kept")
			require.Nil(t, summary.StaleAt, "and are not marked stale, so they are not built again from no data")
			require.Equal(t, 30.0, summary.TotalDailyDose)

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
			require.NoError(t, controlDB.TruncateSites(ctx))
			truncated, err = controlDB.Logs.Queries.SearchEvents(ctx, logsdb.SearchEventsParams{SiteID: &siteID, Type: new("site_truncated"), Limit: 10})
			require.NoError(t, err)
			require.Len(t, truncated, 1)
		})
	}
}
