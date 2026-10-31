package sync

import (
	"context"
	"fmt"
	"log/slog"

	"nighttroop.t1d.nz/clients/dexcom_client"
	"nighttroop.t1d.nz/models/dexcommodels"
	"nighttroop.t1d.nz/utils/consts"
	utilerrors "nighttroop.t1d.nz/utils/errors"
)

// syncDexcom copies the last 15 minutes of Dexcom readings into the site, returning how many were new
func (b *SyncJobManager) syncDexcom(syncInfo *SyncJobInfo) (int, error) {
	slog.InfoContext(syncInfo.ctx, "Syncing Dexcom")

	ctx, cancel := context.WithTimeout(syncInfo.ctx, syncTimeout)
	defer cancel()

	return copyDexcomEGVs(ctx, syncInfo, 15)
}

// copyDexcomEGVs copies the last minutes of Dexcom readings into the site, returning how many were new
func copyDexcomEGVs(ctx context.Context, syncInfo *SyncJobInfo, minutes int) (int, error) {
	client := dexcom_client.NewClient(syncInfo.urlOrUsername, syncInfo.password, consts.NewDexcomRegionFromString(syncInfo.region) == consts.DexcomRegionOther)
	egvs, err := client.FetchEGVs(ctx, minutes)
	if err != nil {
		return 0, fmt.Errorf("Dexcom Error: %w", err)
	}
	added := 0
	for _, e := range egvs {
		createParams := dexcommodels.EGVToCreate(e, "dexcom_sync")
		res, err := syncInfo.Site.SiteDB.Queries.CreateGlucoseReading(ctx, createParams)
		if err != nil {
			if utilerrors.IsSQLDuplicateID(err) {
				continue
			}
			// Probably just non-unique ID
			slog.ErrorContext(ctx, "Error creating entry in database", "error", err, "id", createParams.ID)
			continue
		}
		if res != nil {
			added++
			syncInfo.Site.published(res)
		}
	}
	if added > 0 {
		slog.InfoContext(ctx, "Added new Dexcom entries for site", "added", added)
	}
	return added, nil
}
