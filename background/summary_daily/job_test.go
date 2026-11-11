package summary_daily

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/background/onevent"
	"nighttroop.t1d.nz/data/events"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/server/siteapi/sitecollections/sitedata"
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

// reading writes a glucose reading at at and marks it as written at marked, as its model change does
func (r *rig) reading(id string, at, marked time.Time) {
	reading, err := r.site.CreateGlucoseReading(r.t.Context(), sitedb.CreateGlucoseReadingParams{ID: id, Mgdl: 120, ReadingType: "sgv", OccurredAt: at.UnixMilli(), CreatedByIdentityName: "phone"})
	require.NoError(r.t, err)
	r.builder.markStale(events.ModelChange(r.siteID, events.Created, reading), marked)
}

// Data written to old days builds only those days again, and their weeks, fortnights and months, once the site's data
// has been quiet for a while, so a backfill's many records are built once, from all of them
func TestRebuildStaleBuildsTheWrittenDaysOnceQuiet(t *testing.T) {
	r := newRig(t)
	ctx, now := t.Context(), time.Now().UTC()
	week := sitedata.PeriodStart("week", now.AddDate(0, 0, -14), time.UTC, "monday")
	monday, tuesday := week.Add(12*time.Hour), week.AddDate(0, 0, 1).Add(12*time.Hour)

	// The week was built before the backfill wrote its Tuesday
	r.reading("monday", monday, now)
	require.NoError(t, r.builder.RebuildStale(ctx, now.Add(staleQuiet)))
	require.True(t, r.built(monday.Format(time.DateOnly)))
	r.opened = 0

	r.reading("tuesday", tuesday, now)
	r.reading("tuesday-later", tuesday.Add(time.Hour), now.Add(30*time.Second))
	require.NoError(t, r.builder.RebuildStale(ctx, now.Add(staleQuiet)))
	require.False(t, r.built(tuesday.Format(time.DateOnly)), "not until its data has gone quiet")
	require.Zero(t, r.opened)

	require.NoError(t, r.builder.RebuildStale(ctx, now.Add(30*time.Second+staleQuiet)))
	require.True(t, r.built(tuesday.Format(time.DateOnly)))
	require.False(t, r.built(week.AddDate(0, 0, 2).Format(time.DateOnly)), "a day with nothing written is not built")
	require.Equal(t, 1, r.opened)
	built, err := r.site.GetSummaryPeriod(ctx, sitedb.GetSummaryPeriodParams{Period: "week", Date: week.Format(time.DateOnly)})
	require.NoError(t, err)
	require.Equal(t, int64(2), built.Days, "the kept week is built again with its new day")

	// Once built, nothing is left to build
	require.NoError(t, r.builder.RebuildStale(ctx, now.Add(time.Hour)))
	require.Equal(t, 1, r.opened)
}

// A site written to without a pause, as by a long backfill, has its days built once its data has waited staleWait
func TestRebuildStaleBuildsAfterTheLongestWait(t *testing.T) {
	r := newRig(t)
	now := time.Now().UTC()
	day := now.AddDate(0, 0, -3)
	for minute := time.Duration(0); minute < staleWait; minute += time.Minute {
		r.reading(fmt.Sprint("reading", minute), day.Add(minute), now.Add(minute))
	}
	require.NoError(t, r.builder.RebuildStale(t.Context(), now.Add(staleWait)))
	require.True(t, r.built(day.Format(time.DateOnly)))
}

// Only today's data, which is built as it is read, and data not of a day's summary, open no site
func TestRebuildStaleLeavesToday(t *testing.T) {
	r := newRig(t)
	now := time.Now().UTC()
	r.reading("now", now, now)
	activity, err := r.site.CreateActivity(t.Context(), sitedb.CreateActivityParams{ID: "walk", ActivityType: "walk", OccurredAt: now.AddDate(0, 0, -3).UnixMilli(), CreatedByIdentityName: "phone"})
	require.NoError(t, err)
	r.builder.markStale(events.ModelChange(r.siteID, events.Created, activity), now)
	require.NoError(t, r.builder.RebuildStale(t.Context(), now.Add(staleQuiet)))
	require.Zero(t, r.opened)
}

// A record changed to another day builds the day it moved from again as well as the one it moved to
func TestRebuildStaleBuildsTheDayARecordMovedFrom(t *testing.T) {
	r := newRig(t)
	ctx, now := t.Context(), time.Now().UTC()
	from, to := now.AddDate(0, 0, -3), now.AddDate(0, 0, -2)
	r.reading("before", from, now)
	require.NoError(t, r.builder.RebuildStale(ctx, now.Add(staleQuiet)))

	_, err := r.site.HardDeleteGlucoseReading(ctx, sitedb.HardDeleteGlucoseReadingParams{ID: "before"})
	require.NoError(t, err)
	moved, err := r.site.CreateGlucoseReading(ctx, sitedb.CreateGlucoseReadingParams{ID: "after", Mgdl: 120, ReadingType: "sgv", OccurredAt: to.UnixMilli(), CreatedByIdentityName: "phone"})
	require.NoError(t, err)
	event := events.ModelChange(r.siteID, events.Updated, moved)
	event.Previous = &sitedb.GlucoseReading{ID: "before", OccurredAt: from.UnixMilli()}
	r.builder.markStale(event, now)
	require.NoError(t, r.builder.RebuildStale(ctx, now.Add(staleQuiet)))

	require.True(t, r.built(to.Format(time.DateOnly)))
	day, err := r.site.GetSummaryPeriod(ctx, sitedb.GetSummaryPeriodParams{Period: "day", Date: from.Format(time.DateOnly)})
	if err == nil {
		require.Zero(t, day.GlucoseReadings, "the day it moved from no longer has it")
	}
}
