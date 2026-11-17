package upkeep_sites

import (
	"context"
	"fmt"
	"log/slog"
	"sync/atomic"
	"time"

	"nighttroop.t1d.nz/background/cron"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/utils"
	"nighttroop.t1d.nz/utils/aslog"
	"nighttroop.t1d.nz/utils/config"
)

// UpkeepSiteJobManager deletes each batch's unused sites and truncates its sites, a batch, the sites whose IDs start
// with the same letter, at a time
type UpkeepSiteJobManager struct {
	controlDB    *db.ControlDB
	ServerConfig *config.ServerConfig

	DeleteAfterInactivity time.Duration

	SiteUpkeepIntervalInSeconds int
}

func SetupUpkeepSiteJob(controlDB *db.ControlDB, config *config.ServerConfig) *UpkeepSiteJobManager {
	return &UpkeepSiteJobManager{
		controlDB:                   controlDB,
		DeleteAfterInactivity:       time.Duration(config.DeleteAfterInactivityMins) * time.Minute,
		SiteUpkeepIntervalInSeconds: config.SiteUpkeepIntervalInSeconds,
		ServerConfig:                config,
	}
}

func (b *UpkeepSiteJobManager) SetupCron(c cron.CronScheduler) {
	count := len(utils.SITE_CHAR_SET)

	tick := max(1, b.SiteUpkeepIntervalInSeconds/count)

	var index atomic.Int64

	// The goal here is to smooth out resource usage, if you looped over all baches at once, we would use a ton of CPU and that will slow down responses
	// This nice thing about this method of upkeep is that because the cron task cannot run concurrently for the same batch,
	// we avoid using too many resources, instead it will just delay the upkeep by tick time
	c.AddFunc("UpkeepSiteJob", fmt.Sprintf("@every %ds", tick), func(ctx context.Context) error {
		i := int(index.Add(1)-1) % count
		batch := string(utils.SITE_CHAR_SET[i])
		if err := b.DeleteUnusedSites(ctx, batch); err != nil {
			slog.ErrorContext(ctx, "Error deleting unused sites", "error", err, "batch", batch)
		}
		// Tests run each job once, so they see every site truncated straight away
		sites := b.controlDB.AllSites()
		if !b.ServerConfig.IsTest() {
			sites = b.batchSites(ctx, batch)
		}
		return b.controlDB.TruncateSites(aslog.With(ctx, slog.String("batch", batch)), sites)
	})
}

// batchSites are the open DBs of the live and disabled sites in batch
func (b *UpkeepSiteJobManager) batchSites(ctx context.Context, batch string) []*db.SiteDB {
	res, err := b.controlDB.Queries.GetBatchSitesToUpkeep(ctx, controldb.GetBatchSitesToUpkeepParams{Batch: batch})
	if err != nil {
		slog.ErrorContext(ctx, "Cannot get the batch's sites", "error", err, "batch", batch)
	}
	return b.controlDB.GetSiteDBsFromModels(res)
}

func (b *UpkeepSiteJobManager) DeleteUnusedSites(ctx context.Context, batch string) error {
	log := false
	defer aslog.TrackTime(ctx, time.Now(), "DeleteUnusedSites", &log, nil)

	deleteList, err := b.findSitesToDelete(ctx, batch)
	if err != nil {
		return fmt.Errorf("cant find sites to delete without archiving %w", err)
	}
	if len(deleteList) == 0 {
		return nil
	}
	log = true
	slog.InfoContext(ctx, "Deleting site without archiving", "subdomains", deleteList)

	// Just delete these sites without an archive, as they have not been used in so many minutes
	if err := b.controlDB.DeleteSites(ctx, deleteList); err != nil {
		return err
	}
	for _, siteID := range deleteList {
		b.controlDB.Logs.SiteDeletedEvent(aslog.WithSite(ctx, siteID), siteID, "unused")
	}
	return nil
}

// The sites we delete without archiving are sites that
func (b *UpkeepSiteJobManager) findSitesToDelete(ctx context.Context, batch string) ([]string, error) {
	sites, err := GetNeverUsedSites(ctx, b.controlDB, batch, b.DeleteAfterInactivity)
	if err != nil {
		return nil, fmt.Errorf("cant get never seen sites %w", err)
	}

	if len(sites) == 0 {
		return nil, nil
	}

	deleteSites := []string{}
	for _, site := range sites {
		if site.Subdomain == "" {
			continue
		}
		deleteSites = append(deleteSites, site.Subdomain)
	}

	return deleteSites, nil
}

// GetNeverUsedSites retrieves sites that have never been logged into
func GetNeverUsedSites(ctx context.Context, controlDB *db.ControlDB, char string, lastSeenMinutes time.Duration) ([]*db.SiteDB, error) {
	res, err := controlDB.Queries.GetSitesToDeleteInBatch(ctx, controldb.GetSitesToDeleteInBatchParams{
		Batch:      char,
		LastSeenAt: lastSeenMinutes.Minutes(),
	})
	if err != nil {
		return nil, fmt.Errorf("cannot query never seen sites: %w", err)
	}

	return controlDB.GetSiteDBsFromModels(res), nil
}
