package summary_daily

import (
	"context"
	"database/sql"
	"errors"
	"maps"
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

	// stale are, by site, when the data its summaries are built from was first and last written since they were built
	// again. Which summaries are stale is kept in the site's DB, which a restart keeps, so the first run since the
	// server started builds every site's; checked is whether it has.
	staleLock sync.Mutex
	stale     map[string]*staleData
	checked   bool
}

// staleData is when a site's data was first and last written since its summaries were built again
type staleData struct {
	first, last time.Time
}

const (
	// unbuiltSince is how long before a site's first write since it was built that its glucose is looked through for
	// days never built, as a write is published once saved; and restartSince how long before the server started
	unbuiltSince = time.Minute
	restartSince = time.Hour
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

// markStale keeps when a site's data a day's summary is built from was first and last written, for when to build its
// stale days again; which they are its triggers mark in its DB. It is called as each change is published, so it only
// keeps it.
func (b *Builder) markStale(event events.Event, now time.Time) {
	switch event.Model.(type) {
	case *sitedb.GlucoseReading, *sitedb.InsulinDose, *sitedb.FoodIntake, *sitedb.TempBasal, *sitedb.TempTarget,
		*sitedb.ProfileSwitch, *sitedb.DeviceReport, *sitedb.EventGroupsBundle:
	default:
		return
	}
	if event.SiteID == "" {
		return
	}
	b.staleLock.Lock()
	defer b.staleLock.Unlock()
	data := b.stale[event.SiteID]
	if data == nil {
		data = &staleData{first: now}
		b.stale[event.SiteID] = data
	}
	data.last = now
}

// RebuildStale builds again the days, and their weeks, fortnights and months, marked stale in each site's DB, and those
// of the glucose written to it since in days never built, once its data has gone staleQuiet without more, or waited
// staleWait, and on the first run since the server started every site's, leaving out today, which is built as it is
// read. A site with only today's data written is not opened.
func (b *Builder) RebuildStale(ctx context.Context, now time.Time) error {
	b.staleLock.Lock()
	due := map[string]*staleData{}
	for siteID, data := range b.stale {
		if now.Sub(data.last) >= staleQuiet || now.Sub(data.first) >= staleWait {
			due[siteID] = data
			delete(b.stale, siteID)
		}
	}
	if !b.checked {
		for _, site := range b.controlDB.AllSites() {
			if due[site.Subdomain] == nil {
				due[site.Subdomain] = &staleData{first: time.Now().Add(unbuiltSince - restartSince)}
			}
		}
		b.checked = true
	}
	b.staleLock.Unlock()

	var errs []error
	for siteID, data := range due {
		site, err := b.controlDB.Queries.GetSiteInAnyState(ctx, controldb.GetSiteInAnyStateParams{ID: siteID})
		if err != nil {
			errs = append(errs, err)
			continue
		}
		today := sitedata.PeriodStart("day", now, (&httputils.SiteContext{SiteID: siteID, SiteModel: site}).GetSiteTimezone(), "monday")
		if site.State != "live" {
			continue
		}
		// Taken before its stale days are read, so one marked as they are built is not cleared unbuilt
		began := time.Now()
		siteDB, _, err := b.controlDB.GetSite(ctx, siteID)
		var stale, unbuilt []int64
		if err == nil {
			stale, err = siteDB.Queries().StaleSummaryDays(ctx)
		}
		if err == nil {
			unbuilt, err = siteDB.Queries().UnbuiltGlucoseTimes(ctx, sitedb.UnbuiltGlucoseTimesParams{
				Since: data.first.Add(-unbuiltSince).UnixMilli(), Before: today.UnixMilli()})
		}
		if err != nil {
			b.restale(siteID, data)
			errs = append(errs, err)
			continue
		}
		var times []time.Time
		for _, at := range append(stale, unbuilt...) {
			if at := time.UnixMilli(at); at.Before(today) {
				times = append(times, at)
			}
		}
		if len(times) == 0 {
			continue
		}
		siteCtx := aslog.WithSite(ctx, siteID)
		siteContext, err := b.site(siteCtx, siteID)
		if err == nil {
			err = sitedata.RebuildSummaryDays(siteCtx, siteContext.SiteDB.Queries(), siteContext, times, now)
		}
		if err == nil {
			err = siteContext.SiteDB.Queries().ClearStaleSummaries(siteCtx, sitedb.ClearStaleSummariesParams{Before: new(began.UnixMilli())})
		}
		if err != nil {
			b.restale(siteID, data)
		}
		errs = append(errs, err)
	}
	return errors.Join(errs...)
}

// restale keeps a site whose stale days could not be built to be built on a later run, with any written since
func (b *Builder) restale(siteID string, data *staleData) {
	b.staleLock.Lock()
	defer b.staleLock.Unlock()
	if since := b.stale[siteID]; since != nil {
		data.last = since.last
	}
	b.stale[siteID] = data
}
