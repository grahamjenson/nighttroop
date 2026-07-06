package sync

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"nighttroop.t1d.nz/clients/dexcom_client"
	"nighttroop.t1d.nz/data/sql/user"
	"nighttroop.t1d.nz/models/dexcom"
	"nighttroop.t1d.nz/utils/consts"
)

func (b *SyncJobManager) syncDexcom(syncInfo *SyncJobInfo) error {
	if len(syncInfo.userSyncSources) == 0 {
		return fmt.Errorf("Dexcom, no user sync sources found for unique hash: %s", syncInfo.UniqueHash)
	}

	slog.Info("Syncing Dexcom", "unique_hash", syncInfo.UniqueHash)

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
			createParams := dexcom.EGVToCreate(e, "dexcom", userSyncSource.syncIdentity.ID, userSyncSource.syncIdentity.Name)
			res, err := userSyncSource.siteDB.Queries.CreateGlucoseReading(context.TODO(), createParams)
			if err != nil {
				if user.IsDuplicateID(err) {
					continue
				}
				// Probably just non-unique ID
				slog.Error("Error creating entry in database", "error", err, "id", createParams.ID, "sync_source_id", userSyncSource.syncIdentity.ID)
				continue
			}

			if res != nil {
				addedColumns = true
			}

		}

		// Update that it has been seen and used recently
		if addedColumns {
			slog.Info("Added new Dexcom entries for user", "subdomain", userSyncSource.siteModel.Subdomain)
			userSyncSource.controlDB.UpdateIdentityDateFlags(userSyncSource.syncIdentity, true, true)
			totalAddedColumns += 1
		}
	}

	return nil
}
