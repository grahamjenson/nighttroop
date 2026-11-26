// Package site_banners deletes the site banners that have expired
package site_banners

import (
	"context"
	"log/slog"
	"time"

	"nighttroop.t1d.nz/background/cron"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/sql/controldb"
)

// SetupCron deletes, every 10 minutes, the site banners past the time they expire at. A site stops showing one at
// that time, so this only keeps the table to the banners in use.
func SetupCron(c cron.CronScheduler, controlDB *db.ControlDB) {
	c.AddFunc("DeleteExpiredSiteBanners", "@every 10m", func(ctx context.Context) error {
		deleted, err := controlDB.Queries.DeleteExpiredSiteBanners(ctx, controldb.DeleteExpiredSiteBannersParams{Now: time.Now().UnixMilli()})
		if deleted > 0 {
			slog.InfoContext(ctx, "Deleted expired site banners", "banners", deleted)
		}
		return err
	})
}
