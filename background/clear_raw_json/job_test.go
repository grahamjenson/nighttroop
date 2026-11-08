package clear_raw_json_test

import (
	"context"
	"testing"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/background/clear_raw_json"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/utils/testutils"
)

// scheduler keeps the jobs added to it, with their specs, to run them straight away
type scheduler map[string]func(context.Context) error

func (s scheduler) AddFunc(name string, spec string, cmd func(context.Context) error) { s[name+" "+spec] = cmd }

// Each hour, at a fixed minute, every site's old raw JSON is cleared
func TestClearRawJSONCron(t *testing.T) {
	_, controlDB, site := testutils.NewControlDBWithSite(t)
	ctx, q := t.Context(), site.SiteDB.Queries()
	raw := `{"device":"loop"}`
	_, err := q.CreateDeviceReport(ctx, sitedb.CreateDeviceReportParams{ID: "old", OccurredAt: time.Now().AddDate(0, 0, -3).UnixMilli(), RawJsonData: &raw, CreatedByIdentityName: "phone"})
	require.NoError(t, err)

	jobs := scheduler{}
	clear_raw_json.SetupCron(jobs, controlDB)
	require.NoError(t, jobs["ClearRawJSON "+clear_raw_json.Spec](ctx))
	report, err := q.GetDeviceReport(ctx, sitedb.GetDeviceReportParams{ID: "old"})
	require.NoError(t, err)
	require.Nil(t, report.RawJsonData)

	schedule, err := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow).Parse(clear_raw_json.Spec)
	require.NoError(t, err)
	at := time.Date(2026, 11, 8, 10, 20, 0, 0, time.UTC)
	require.Equal(t, time.Date(2026, 11, 8, 11, 17, 0, 0, time.UTC), schedule.Next(at))
}
