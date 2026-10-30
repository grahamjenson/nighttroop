package summary_daily

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"nighttroop.t1d.nz/background/cron"
	"nighttroop.t1d.nz/background/onevent"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/events"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/server/siteapi/sitecollections/sitedata"
	"nighttroop.t1d.nz/utils/httputils"
)

// SetupCron builds each site's yesterday soon after midnight in its timezone, checking every 10 minutes for one not yet
// built
func SetupCron(c cron.CronScheduler, controlDB *db.ControlDB) {
	c.AddFunc("BuildSummaryDaily", "@every 600s", func() error { return Build(context.Background(), controlDB, time.Now()) })
}

// RebuildOnTimezoneChange is the job that, when a site's timezone changes, deletes its days, from a midnight they no
// longer follow, and builds the last week of them again in the new one
func RebuildOnTimezoneChange(controlDB *db.ControlDB) onevent.Job {
	return onevent.Job{
		Queue:  "RebuildSummaryDaily",
		Filter: func(event events.Event) bool { return event.Name == events.SiteTimezoneChanged },
		Run: func(ctx context.Context, fired onevent.Fired) error {
			siteContext, err := site(controlDB, fired.SiteID)
			if err != nil {
				return err
			}
			return sitedata.RebuildSummaryDailyWeek(ctx, siteContext.SiteDB.Queries, siteContext, time.Now())
		},
	}
}

// Build builds each live site's yesterday, in its timezone, that is not yet built
func Build(ctx context.Context, controlDB *db.ControlDB, now time.Time) error {
	sites, err := controlDB.Queries.GetAllSites(ctx)
	if err != nil {
		return err
	}
	var errs []error
	for _, site := range sites {
		if site.State == "live" {
			errs = append(errs, buildYesterday(ctx, controlDB, site.ID, now))
		}
	}
	return errors.Join(errs...)
}

func buildYesterday(ctx context.Context, controlDB *db.ControlDB, siteID string, now time.Time) error {
	siteContext, err := site(controlDB, siteID)
	if err != nil {
		return err
	}
	q := siteContext.SiteDB.Queries
	yesterday := sitedata.LocalDate(now, siteContext.GetSiteTimezone(), -1)
	if _, err := q.GetSummaryDaily(ctx, sitedb.GetSummaryDailyParams{Date: yesterday}); !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	_, err = sitedata.RebuildSummaryDaily(ctx, q, siteContext, yesterday, now)
	return err
}

// site is the site's context for building its days, which only need the site's own settings and DB
func site(controlDB *db.ControlDB, siteID string) (*httputils.SiteContext, error) {
	siteDB, model, err := controlDB.GetSite(siteID)
	if err != nil {
		return nil, err
	}
	return &httputils.SiteContext{SiteID: siteID, ControlDB: controlDB, SiteDB: siteDB, SiteModel: model}, nil
}
