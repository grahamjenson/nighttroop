package summary_daily

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/background/onevent"
	"nighttroop.t1d.nz/data/events"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/utils/httputils"
	"nighttroop.t1d.nz/utils/testutils"
)

// rig is a builder of one site that counts each time it opens the site, and can fail to
type rig struct {
	t       *testing.T
	builder *Builder
	siteID  string
	site    *sitedb.Queries
	control *controldb.Queries
	bus     *events.SiteBus
	opened  int
	fail    error
}

func newRig(t *testing.T) *rig {
	cfg, controlDB, site := testutils.NewControlDBWithSite(t)
	r := &rig{t: t, builder: NewBuilder(cfg.Events, controlDB), siteID: site.SiteDB.Subdomain, site: site.SiteDB.Queries(), control: controlDB.Queries, bus: cfg.Events}
	open := r.builder.site
	r.builder.site = func(ctx context.Context, siteID string) (*httputils.SiteContext, error) {
		r.opened++
		if r.fail != nil {
			return nil, r.fail
		}
		return open(ctx, siteID)
	}
	return r
}

// built is whether the site's day named date is built
func (r *rig) built(date string) bool {
	_, err := r.site.GetSummaryPeriod(r.t.Context(), sitedb.GetSummaryPeriodParams{Period: "day", Date: date})
	return err == nil
}

func TestBuildOpensASiteOnlyWhenItsYesterdayMovesOn(t *testing.T) {
	r := newRig(t)
	now := time.Date(2026, 10, 28, 12, 0, 0, 0, time.UTC)
	require.NoError(t, r.builder.Build(t.Context(), now))
	require.Equal(t, 1, r.opened)
	require.True(t, r.built("2026-10-27"))

	// The same yesterday, later in the day, does not open the site again
	require.NoError(t, r.builder.Build(t.Context(), now))
	require.NoError(t, r.builder.Build(t.Context(), now.Add(11*time.Hour)))
	require.Equal(t, 1, r.opened)

	// The next day builds the new yesterday
	require.NoError(t, r.builder.Build(t.Context(), now.Add(24*time.Hour)))
	require.Equal(t, 2, r.opened)
	require.True(t, r.built("2026-10-28"))
}

func TestBuildTriesASiteAgainAfterAnError(t *testing.T) {
	r := newRig(t)
	now := time.Date(2026, 10, 28, 12, 0, 0, 0, time.UTC)
	r.fail = errors.New("site cannot open")
	require.Error(t, r.builder.Build(t.Context(), now))
	require.False(t, r.built("2026-10-27"))

	r.fail = nil
	require.NoError(t, r.builder.Build(t.Context(), now))
	require.Equal(t, 2, r.opened)
	require.True(t, r.built("2026-10-27"))
}

func TestBuildOpensASiteAgainAfterItsSettingsChange(t *testing.T) {
	r := newRig(t)
	ctx := t.Context()
	now := time.Date(2026, 10, 28, 12, 0, 0, 0, time.UTC)
	require.NoError(t, r.builder.Build(ctx, now))
	require.Equal(t, 1, r.opened)

	// Rebuilding on a settings change forgets the site, so the next run checks it again
	require.NoError(t, r.builder.RebuildOnSettingsChange().Run(ctx, onevent.Fired{Name: events.SiteWeekStartChanged, SiteID: r.siteID}))
	require.Equal(t, 2, r.opened)
	require.NoError(t, r.builder.Build(ctx, now))
	require.Equal(t, 3, r.opened)

	// A new timezone is a new yesterday, even before its rebuild runs
	require.NoError(t, r.control.UpdateSitesSettings(ctx, controldb.UpdateSitesSettingsParams{ID: r.siteID, Timezone: new("Pacific/Auckland")}))
	require.NoError(t, r.builder.Build(ctx, now))
	require.Equal(t, 4, r.opened)
	require.True(t, r.built("2026-10-28"))
}

func TestBuildBuildsYesterdayAgainOnceItIsDeletedThroughTheAPI(t *testing.T) {
	r := newRig(t)
	ctx := t.Context()
	now := time.Date(2026, 10, 28, 12, 0, 0, 0, time.UTC)
	require.NoError(t, r.builder.Build(ctx, now))
	// The API deletes the day and publishes the row as it was
	row, err := r.site.GetSummaryPeriod(ctx, sitedb.GetSummaryPeriodParams{Period: "day", Date: "2026-10-27"})
	require.NoError(t, err)
	_, err = r.site.HardDeleteSummaryPeriod(ctx, sitedb.HardDeleteSummaryPeriodParams{Period: "day", Date: "2026-10-27"})
	require.NoError(t, err)
	r.bus.Publish(r.siteID, events.ModelChange(r.siteID, events.Deleted, row))

	require.NoError(t, r.builder.Build(ctx, now))
	require.Equal(t, 2, r.opened)
	require.True(t, r.built("2026-10-27"))
}

func TestBuildForgetsSitesNoLongerLive(t *testing.T) {
	r := newRig(t)
	ctx := t.Context()
	now := time.Date(2026, 10, 28, 12, 0, 0, 0, time.UTC)
	require.NoError(t, r.builder.Build(ctx, now))
	require.Len(t, r.builder.built, 1)

	_, err := r.control.DisableSite(ctx, controldb.DisableSiteParams{ID: r.siteID})
	require.NoError(t, err)
	require.NoError(t, r.builder.Build(ctx, now))
	require.Empty(t, r.builder.built)
	require.Equal(t, 1, r.opened)
}

// Asking for the summaries to be regenerated builds again each day of the last year with data, and yesterday, and the
// weeks, fortnights and months of them, leaving out the days with none
func TestRebuildOnRequestRegeneratesTheLastYear(t *testing.T) {
	r := newRig(t)
	ctx, now := t.Context(), time.Now().UTC()
	read := now.AddDate(0, 0, -40)
	_, err := r.site.CreateGlucoseReading(ctx, sitedb.CreateGlucoseReadingParams{ID: "old", Mgdl: 120, ReadingType: "sgv", OccurredAt: read.UnixMilli(), CreatedByIdentityName: "phone"})
	require.NoError(t, err)

	require.NoError(t, r.builder.RebuildOnRequest().Run(ctx, onevent.Fired{Name: events.SummariesRegenerationRequested, SiteID: r.siteID}))
	require.True(t, r.built(read.Format(time.DateOnly)))
	require.True(t, r.built(now.AddDate(0, 0, -1).Format(time.DateOnly)))
	require.False(t, r.built(now.AddDate(0, 0, -50).Format(time.DateOnly)))
	month, err := r.site.GetSummaryPeriod(ctx, sitedb.GetSummaryPeriodParams{Period: "month", Date: time.Date(read.Year(), read.Month(), 1, 0, 0, 0, 0, time.UTC).Format(time.DateOnly)})
	require.NoError(t, err)
	require.Equal(t, int64(1), month.Days)
	weeks, err := r.site.SearchSummaryPeriods(ctx, sitedb.SearchSummaryPeriodsParams{Period: "week", Limit: 60})
	require.NoError(t, err)
	require.Len(t, weeks, 1)
}
