package sync

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"nighttroop.t1d.nz/clients/dexcom_client"
	"nighttroop.t1d.nz/models/dexcommodels"
	"nighttroop.t1d.nz/utils/consts"
	utilerrors "nighttroop.t1d.nz/utils/errors"
)

func (b *SyncJobManager) syncDexcom(syncInfo *SyncJobInfo) error {
	if len(syncInfo.userSyncSources) == 0 {
		return fmt.Errorf("Dexcom, no user sync sources found for site_id %s type %s", syncInfo.SiteID, syncInfo.Type)
	}

	slog.Info("Syncing Dexcom", "site_id", syncInfo.SiteID, "sync_type", syncInfo.Type)

	client := dexcom_client.NewClient(syncInfo.urlOrUsername, syncInfo.password, syncInfo.region == consts.DexcomRegionOther)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	egvs, err := client.FetchEGVs(ctx, 15)
	if err != nil {
		return fmt.Errorf("Dexcom Error: %w", err)
	}
	totalAddedColumns := 0
	for _, userSyncSource := range syncInfo.userSyncSources {
		addedColumns := false
		for _, e := range egvs {
			createParams := dexcommodels.EGVToCreate(e, "dexcom_sync")
			res, err := userSyncSource.siteDB.Queries.CreateGlucoseReading(context.TODO(), createParams)
			if err != nil {
				if utilerrors.IsSQLDuplicateID(err) {
					continue
				}
				// Probably just non-unique ID
				slog.Error("Error creating entry in database", "error", err, "id", createParams.ID, "site_id", syncInfo.SiteID, "sync_type", syncInfo.Type)
				continue
			}

			if res != nil {
				addedColumns = true
			}

		}

		// Update that it has been seen and used recently
		if addedColumns {
			slog.Info("Added new Dexcom entries for user", "subdomain", userSyncSource.siteModel.ID)
			totalAddedColumns += 1
		}
	}

	return nil
}
