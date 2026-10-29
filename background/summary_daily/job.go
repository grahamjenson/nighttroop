package summary_daily

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"nighttroop.t1d.nz/background/cron"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/server/siteapi/sitecollections/sitedata"
	"nighttroop.t1d.nz/utils/settings"
)

// SetupCron builds each site's yesterday soon after midnight in its timezone, checking every 10 minutes for one not yet
// built
func SetupCron(c cron.CronScheduler, controlDB *db.ControlDB) {
	c.AddFunc("BuildSummaryDaily", "@every 600s", func() error { return Build(context.Background(), controlDB, time.Now()) })
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
	siteDB, site, err := controlDB.GetSite(siteID)
	if err != nil {
		return err
	}
	siteSettings := settings.NewSiteCombinedSettingsMap(site)
	yesterday := sitedata.LocalDate(now, siteSettings, -1)
	if _, err := siteDB.Queries.GetSummaryDaily(ctx, sitedb.GetSummaryDailyParams{Date: yesterday}); !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	_, err = sitedata.RebuildSummaryDaily(ctx, siteDB.Queries, siteSettings, yesterday, now)
	return err
}
