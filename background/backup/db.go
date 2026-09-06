package backup

import (
	"context"
	"fmt"
	"time"

	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/sql/control"
)

func HasSitesInBatch(controlDB *db.ControlDB, char string) (bool, error) {
	res, err := controlDB.Queries.GetBatchSites(context.TODO(), control.GetBatchSitesParams{
		Batch: char,
	})
	if err != nil {
		return false, fmt.Errorf("cannot query sites: %w", err)
	}
	return len(res) > 0, nil
}

func GetSiteBatch(controlDB *db.ControlDB, char string) ([]*db.SiteDB, error) {

	sites := []*db.SiteDB{}
	res, err := controlDB.Queries.GetBatchSites(context.TODO(), control.GetBatchSitesParams{
		Batch: char,
	})
	if err != nil {
		return sites, fmt.Errorf("cannot query sites: %w", err)
	}

	return controlDB.GetSiteDBsFromModels(res)
}

// GetNeverUsedSites retrieves sites that have never been logged into
func GetNeverUsedSites(controlDB *db.ControlDB, lastSeenMinutes time.Duration, lastWriteMinutes time.Duration) ([]*db.SiteDB, error) {
	res, err := controlDB.Queries.GetNeverUsedSites(context.TODO(), control.GetNeverUsedSitesParams{
		LastSeenAt:  lastSeenMinutes.Minutes(),
		LastWriteAt: lastWriteMinutes.Minutes(),
	})
	if err != nil {
		return nil, fmt.Errorf("cannot query never seen sites: %w", err)
	}

	return controlDB.GetSiteDBsFromModels(res)
}

// GetInactiveSites retrieves sites that have been inactive for a specified duration
func GetInactiveSites(controlDB *db.ControlDB, lastSeenMinutes time.Duration, lastWriteMinutes time.Duration) ([]*db.SiteDB, error) {
	res, err := controlDB.Queries.GetInactiveSites(context.TODO(), control.GetInactiveSitesParams{
		LastSeenAt:  lastSeenMinutes.Minutes(),
		LastWriteAt: lastWriteMinutes.Minutes(),
	})
	if err != nil {
		return nil, fmt.Errorf("cannot query inactive sites: %w", err)
	}

	return controlDB.GetSiteDBsFromModels(res)

}
