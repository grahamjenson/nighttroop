package backup_sites

import (
	"context"
	"fmt"
	"time"

	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/sql/control"
)

func HasSitesInBatch(controlDB *db.ControlDB, char string) (bool, error) {
	res, err := controlDB.Queries.GetBatchSitesToBackUp(context.TODO(), control.GetBatchSitesToBackUpParams{
		Batch: char,
	})
	if err != nil {
		return false, fmt.Errorf("cannot query sites: %w", err)
	}
	return len(res) > 0, nil
}

func GetSiteBatch(controlDB *db.ControlDB, char string) ([]*db.SiteDB, error) {

	sites := []*db.SiteDB{}
	res, err := controlDB.Queries.GetBatchSitesToBackUp(context.TODO(), control.GetBatchSitesToBackUpParams{
		Batch: char,
	})
	if err != nil {
		return sites, fmt.Errorf("cannot query sites: %w", err)
	}

	return controlDB.GetSiteDBsFromModels(res)
}

// GetNeverUsedSites retrieves sites that have never been logged into
func GetNeverUsedSites(controlDB *db.ControlDB, char string, lastSeenMinutes time.Duration) ([]*db.SiteDB, error) {
	res, err := controlDB.Queries.GetSitesToDeleteInBatch(context.TODO(), control.GetSitesToDeleteInBatchParams{
		Batch:      char,
		LastSeenAt: lastSeenMinutes.Minutes(),
	})
	if err != nil {
		return nil, fmt.Errorf("cannot query never seen sites: %w", err)
	}

	return controlDB.GetSiteDBsFromModels(res)
}
