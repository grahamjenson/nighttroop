package upkeep_sites_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/background/cron"
	"nighttroop.t1d.nz/background/upkeep_sites"
	"nighttroop.t1d.nz/data/sql/logsdb"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/utils/testutils"
)

// scheduler keeps the jobs added to it, to run them
type scheduler map[string]func(context.Context) error

func (s scheduler) AddFunc(name, _ string, job func(context.Context) error) { s[name] = job }
func (s scheduler) AddFuncSoon(string, time.Duration, time.Duration, func(context.Context) error) {}

var _ cron.CronScheduler = scheduler{}

// Each upkeep truncates the sites, here every site as tests run each job once, as a site's device report from long ago
// is deleted
func TestUpkeepTruncatesTheSites(t *testing.T) {
	cfg, controlDB, site := testutils.NewControlDBWithSite(t)
	ctx, device := t.Context(), "phone"
	old := time.Now().AddDate(0, 0, -40).UnixMilli()
	_, err := site.SiteDB.Queries().CreateDeviceReport(ctx, sitedb.CreateDeviceReportParams{ID: "old", OccurredAt: old, CreatedByDeviceName: &device, CreatedByIdentityName: "phone"})
	require.NoError(t, err)
	jobs := scheduler{}
	upkeep_sites.SetupUpkeepSiteJob(controlDB, cfg).SetupCron(jobs)

	require.NoError(t, jobs["UpkeepSiteJob"](ctx))
	_, err = site.SiteDB.Queries().GetDeviceReport(ctx, sitedb.GetDeviceReportParams{ID: "old"})
	require.ErrorIs(t, err, sql.ErrNoRows)
}

func TestDeleteUnusedSitesRecordsEvent(t *testing.T) {
	cfg, controlDB, site := testutils.NewControlDBWithSite(t)
	ctx, siteID := t.Context(), site.SiteDB.Subdomain
	jobs := upkeep_sites.SetupUpkeepSiteJob(controlDB, cfg)
	// The site has never been used, and counts as unused from the moment it was made
	jobs.DeleteAfterInactivity = -time.Minute

	require.NoError(t, jobs.DeleteUnusedSites(ctx, siteID[:1]))

	events, err := controlDB.Logs.Queries.SearchEvents(ctx, logsdb.SearchEventsParams{SiteID: &siteID, Limit: 10})
	require.NoError(t, err)
	require.Equal(t, "site_deleted", events[0].Name)
	require.Equal(t, `{"reason":"unused"}`, *events[0].Attributes)
}
