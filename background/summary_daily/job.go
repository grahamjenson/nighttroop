package summary_daily

import (
	"context"
	"database/sql"
	"errors"
	"maps"
	"slices"
	"sync"
	"time"

	"nighttroop.t1d.nz/background/cron"
	"nighttroop.t1d.nz/background/onevent"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/events"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/server/siteapi/sitecollections/sitedata"
	"nighttroop.t1d.nz/utils/aslog"
	"nighttroop.t1d.nz/utils/httputils"
)

// Builder builds each live site's yesterday, remembering what it last built for each, so a site's DB is only opened
// when its date moves on rather than on every run
type Builder struct {
	controlDB *db.ControlDB
	// site is the site's context for building its days, which only need the site's own settings and DB
	site func(ctx context.Context, siteID string) (*httputils.SiteContext, error)

	builtLock sync.Mutex
	built     map[string]built

	// stale are, by site, the times of the data written since its summaries were built, waiting to be built again
	staleLock sync.Mutex
	stale     map[string]*staleData
}

// staleData are a site's 15 minute buckets, by when they start, of the data written to it, each with whether it is
// saved in the site's DB, and when the first and last were written; saved is whether any may be in its DB, to be read
// back and deleted once built. Buckets of 15 minutes are kept, as every timezone's midnight starts one.
type staleData struct {
	buckets     map[int64]bool
	first, last time.Time
	saved       bool
}

const (
	staleBucket = 15 * time.Minute
	// staleQuiet is how long a site's data must go unchanged before its days are built again, so the many records a
	// backfill or sync writes are built once, from all of them, and staleWait the longest written data waits, so a
	// long backfill's days are still built as it goes
	staleQuiet = 2 * time.Minute
	staleWait  = 15 * time.Minute
)

// built is a site's yesterday, in the timezone and week start day it was built in, with its last week and month
type built struct {
	date, timezone, weekStart string
}

// NewBuilder is a builder of controlDB's sites that has built none of them yet, and forgets a site whose days are
// changed through its API, such as its yesterday deleted, so the next run checks it again
func NewBuilder(bus *events.SiteBus, controlDB *db.ControlDB) *Builder {
	builder := &Builder{controlDB: controlDB, built: map[string]built{}, stale: map[string]*staleData{}, site: func(ctx context.Context, siteID string) (*httputils.SiteContext, error) {
		siteDB, model, err := controlDB.GetSite(ctx, siteID)
		if err != nil {
			return nil, err
		}
		return &httputils.SiteContext{SiteID: siteID, ControlDB: controlDB, SiteDB: siteDB, SiteModel: model}, nil
	}}
	bus.SubscribeAll(func(event events.Event) {
		if _, ok := event.Model.(*sitedb.SummaryPeriod); ok {
			builder.remember(event.SiteID, nil)
		}
		builder.markStale(event, time.Now())
	})
	return builder
}

// SetupCron builds each site's yesterday soon after midnight in its timezone, and its last week and month once they
// end, checking every 10 minutes for ones not yet built
func SetupCron(c cron.CronScheduler, bus *events.SiteBus, controlDB *db.ControlDB) *Builder {
	builder := NewBuilder(bus, controlDB)
	c.AddFunc("BuildSummaryDaily", "@every 600s", func(ctx context.Context) error { return builder.Build(ctx, time.Now()) })
	c.AddFunc("RebuildStaleSummaries", "@every 60s", func(ctx context.Context) error { return builder.RebuildStale(ctx, time.Now()) })
	return builder
}

// RebuildOnSettingsChange is the job that, when a site's timezone changes, deletes its days, weeks and months, from a
// midnight they no longer follow, and builds the last week of days, and the last week and month, again in the new
// one; and when its week start day changes, does the same for its weeks
func (b *Builder) RebuildOnSettingsChange() onevent.Job {
	return onevent.Job{
		Queue: "RebuildSummaryDaily",
		Filter: func(event events.Event) bool {
			return event.Name == events.SiteTimezoneChanged || event.Name == events.SiteWeekStartChanged
		},
		Run: func(ctx context.Context, fired onevent.Fired) error {
			// Forgotten once done, so a build that ran while its days were deleted is checked again on the next run
			defer b.remember(fired.SiteID, nil)
			siteContext, err := b.site(ctx, fired.SiteID)
			if err != nil {
				return err
			}
			if fired.Name == events.SiteWeekStartChanged {
				return sitedata.RebuildSummaryWeeks(ctx, siteContext.SiteDB.Queries(), siteContext, time.Now())
			}
			return sitedata.RebuildSummaryDailyWeek(ctx, siteContext.SiteDB.Queries(), siteContext, time.Now())
		},
	}
}

// RebuildOnRequest is the job that, when a site's admin asks for its summaries to be built again, builds its last
// year of days, weeks, fortnights and months again from its data
func (b *Builder) RebuildOnRequest() onevent.Job {
	return onevent.Job{
		Queue:  "RebuildSummaryYear",
		Filter: func(event events.Event) bool { return event.Name == events.SummariesRegenerationRequested },
		Run: func(ctx context.Context, fired onevent.Fired) error {
			defer b.remember(fired.SiteID, nil)
			siteContext, err := b.site(ctx, fired.SiteID)
			if err != nil {
				return err
			}
			return sitedata.RebuildSummaryYear(ctx, siteContext.SiteDB.Queries(), siteContext, time.Now())
		},
	}
}

// Build builds each live site's yesterday, last week and last month, in its timezone, that are not yet built,
// skipping without opening its DB a site already built for its yesterday
func (b *Builder) Build(ctx context.Context, now time.Time) error {
	sites, err := b.controlDB.Queries.GetAllSites(ctx)
	if err != nil {
		return err
	}
	live := map[string]bool{}
	var errs []error
	for _, site := range sites {
		if site.State != "live" {
			continue
		}
		live[site.ID] = true
		yesterday := yesterdayOf(site, now)
		if b.isBuilt(site.ID, yesterday) {
			continue
		}
		err := b.buildYesterday(aslog.WithSite(ctx, site.ID), site.ID, now)
		if err == nil {
			b.remember(site.ID, &yesterday)
			// Its days saved waiting to be built are read back and built, as the server may have restarted since
			b.restale(site.ID, &staleData{buckets: map[int64]bool{}, saved: true})
		}
		errs = append(errs, err)
	}
	// Sites no longer live are forgotten, so deleted ones are not kept
	b.builtLock.Lock()
	maps.DeleteFunc(b.built, func(siteID string, _ built) bool { return !live[siteID] })
	b.builtLock.Unlock()
	return errors.Join(errs...)
}

// yesterdayOf is the site's yesterday from its own settings, which are in the control DB, so its DB is not opened
func yesterdayOf(site *controldb.Site, now time.Time) built {
	settings := &httputils.SiteContext{SiteID: site.ID, SiteModel: site}
	location := settings.GetSiteTimezone()
	return built{date: sitedata.LocalDate(now, location, -1), timezone: location.String(), weekStart: settings.GetSiteWeekStartDay()}
}

func (b *Builder) isBuilt(siteID string, yesterday built) bool {
	b.builtLock.Lock()
	defer b.builtLock.Unlock()
	last, ok := b.built[siteID]
	return ok && last == yesterday
}

// remember keeps what was built for the site, or forgets it when nil
func (b *Builder) remember(siteID string, yesterday *built) {
	b.builtLock.Lock()
	defer b.builtLock.Unlock()
	if yesterday == nil {
		delete(b.built, siteID)
	} else {
		b.built[siteID] = *yesterday
	}
}

func (b *Builder) buildYesterday(ctx context.Context, siteID string, now time.Time) error {
	siteContext, err := b.site(ctx, siteID)
	if err != nil {
		return err
	}
	q := siteContext.SiteDB.Queries()
	yesterday := sitedata.LocalDate(now, siteContext.GetSiteTimezone(), -1)
	if _, err := q.GetSummaryPeriod(ctx, sitedb.GetSummaryPeriodParams{Period: "day", Date: yesterday}); errors.Is(err, sql.ErrNoRows) {
		if _, err := sitedata.RebuildSummaryDaily(ctx, q, siteContext, yesterday, now); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	// Yesterday is built first, as it may be the last day of the week or month
	return sitedata.BuildLastSummaryPeriods(ctx, q, siteContext, now)
}

// markStale keeps when the record a model change is of happened, and an update's when it happened before, as it may
// have moved from another day, for their days' summaries to be built again, when it is one a day's summary is built
// from. It is called as each change is published, so it only keeps it.
func (b *Builder) markStale(event events.Event, now time.Time) {
	switch event.Model.(type) {
	case *sitedb.GlucoseReading, *sitedb.InsulinDose, *sitedb.FoodIntake, *sitedb.TempBasal, *sitedb.TempTarget,
		*sitedb.ProfileSwitch, *sitedb.DeviceReport, *sitedb.EventGroupsBundle:
	default:
		return
	}
	model, ok := event.Model.(interface{ GetOccurredAt() int64 })
	if !ok || event.SiteID == "" {
		return
	}
	times := []int64{model.GetOccurredAt()}
	if previous, ok := event.Previous.(interface{ GetOccurredAt() int64 }); ok {
		times = append(times, previous.GetOccurredAt())
	}
	b.staleLock.Lock()
	defer b.staleLock.Unlock()
	data := b.stale[event.SiteID]
	if data == nil {
		data = &staleData{buckets: map[int64]bool{}, first: now}
		b.stale[event.SiteID] = data
	}
	for _, at := range times {
		if bucket := at / staleBucket.Milliseconds(); !data.buckets[bucket] {
			data.buckets[bucket] = false
		}
	}
	data.last = now
}

// RebuildStale builds again the days, and their weeks, fortnights and months, of the data written to each site that
// has gone staleQuiet without more, or waited staleWait, leaving out today, which is built as it is read, and saves the
// days of the rest in their DBs, so a restart does not lose them. A site with only today's data written is not opened.
func (b *Builder) RebuildStale(ctx context.Context, now time.Time) error {
	b.staleLock.Lock()
	due, waiting := map[string]*staleData{}, map[string][]int64{}
	for siteID, data := range b.stale {
		if now.Sub(data.last) >= staleQuiet || now.Sub(data.first) >= staleWait {
			due[siteID] = data
			delete(b.stale, siteID)
			continue
		}
		for bucket, saved := range data.buckets {
			if !saved {
				waiting[siteID] = append(waiting[siteID], bucket)
				data.buckets[bucket] = true
			}
		}
	}
	b.staleLock.Unlock()

	var errs []error
	for siteID, buckets := range waiting {
		errs = append(errs, b.saveStale(aslog.WithSite(ctx, siteID), siteID, buckets, now))
	}
	for siteID, data := range due {
		errs = append(errs, b.rebuildStale(aslog.WithSite(ctx, siteID), siteID, data, now))
	}
	return errors.Join(errs...)
}

// siteToday is when the site's today starts in its timezone, zero once it is no longer live
func (b *Builder) siteToday(ctx context.Context, siteID string, now time.Time) (time.Time, error) {
	site, err := b.controlDB.Queries.GetSiteInAnyState(ctx, controldb.GetSiteInAnyStateParams{ID: siteID})
	if err != nil || site.State != "live" {
		return time.Time{}, err
	}
	return sitedata.PeriodStart("day", now, (&httputils.SiteContext{SiteID: siteID, SiteModel: site}).GetSiteTimezone(), "monday"), nil
}

// beforeToday are when each of the buckets before today starts
func beforeToday(buckets []int64, today time.Time) []time.Time {
	var times []time.Time
	for _, bucket := range buckets {
		if at := time.UnixMilli(bucket * staleBucket.Milliseconds()); at.Before(today) {
			times = append(times, at)
		}
	}
	return times
}

// saveStale saves the site's buckets before its today in its DB, in one transaction, so they are built after a restart
func (b *Builder) saveStale(ctx context.Context, siteID string, buckets []int64, now time.Time) error {
	today, err := b.siteToday(ctx, siteID, now)
	times := beforeToday(buckets, today)
	if err != nil || len(times) == 0 {
		return err
	}
	siteContext, err := b.site(ctx, siteID)
	var tx *sql.Tx
	if err == nil {
		tx, err = siteContext.SiteDB.SQLiteDB.BeginTx(ctx, nil)
	}
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	for _, at := range times {
		if err := siteContext.SiteDB.QueriesWithTx(tx).SaveSummaryStale(ctx, sitedb.SaveSummaryStaleParams{At: at.UnixMilli()}); err != nil {
			return err
		}
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	b.staleLock.Lock()
	defer b.staleLock.Unlock()
	if data := b.stale[siteID]; data != nil {
		data.saved = true
	}
	return nil
}

// rebuildStale builds again the site's days before its today of its data, with any saved in its DB, which are then
// deleted, keeping them to be built on a later run when it cannot
func (b *Builder) rebuildStale(ctx context.Context, siteID string, data *staleData, now time.Time) error {
	today, err := b.siteToday(ctx, siteID, now)
	if err != nil || today.IsZero() || (!data.saved && len(beforeToday(slices.Collect(maps.Keys(data.buckets)), today)) == 0) {
		return err
	}
	siteContext, err := b.site(ctx, siteID)
	var saved []int64
	if err == nil && data.saved {
		saved, err = siteContext.SiteDB.Queries().GetSummaryStale(ctx)
	}
	for _, at := range saved {
		data.buckets[at/staleBucket.Milliseconds()] = true
	}
	if times := beforeToday(slices.Collect(maps.Keys(data.buckets)), today); err == nil && len(times) > 0 {
		err = sitedata.RebuildSummaryDays(ctx, siteContext.SiteDB.Queries(), siteContext, times, now)
	}
	if err == nil && len(saved) > 0 {
		err = siteContext.SiteDB.Queries().DeleteSummaryStale(ctx)
	}
	if err != nil {
		b.restale(siteID, data)
	}
	return err
}

// restale keeps a site's data that could not be built to be built on a later run, with any written since
func (b *Builder) restale(siteID string, data *staleData) {
	b.staleLock.Lock()
	defer b.staleLock.Unlock()
	if since := b.stale[siteID]; since != nil {
		maps.Copy(data.buckets, since.buckets)
		data.last, data.saved = since.last, data.saved || since.saved
	}
	b.stale[siteID] = data
}
