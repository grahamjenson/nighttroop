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
}

// built is a site's yesterday, in the timezone and week start day it was built in, with its last week and month
type built struct {
	date, timezone, weekStart string
}

// NewBuilder is a builder of controlDB's sites that has built none of them yet, and forgets a site whose days are
// changed through its API, such as its yesterday deleted, so the next run checks it again
func NewBuilder(bus *events.SiteBus, controlDB *db.ControlDB) *Builder {
	builder := &Builder{controlDB: controlDB, built: map[string]built{}, site: func(ctx context.Context, siteID string) (*httputils.SiteContext, error) {
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
	})
	return builder
}

// SetupCron builds each site's yesterday soon after midnight in its timezone, and its last week and month once they
// end, checking every 10 minutes for ones not yet built
func SetupCron(c cron.CronScheduler, bus *events.SiteBus, controlDB *db.ControlDB) *Builder {
	builder := NewBuilder(bus, controlDB)
	c.AddFunc("BuildSummaryDaily", "@every 600s", func(ctx context.Context) error { return builder.Build(ctx, time.Now()) })
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
