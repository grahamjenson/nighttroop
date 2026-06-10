package sync

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"nighttroop.t1d.nz/clients/dexcom_client"
	"nighttroop.t1d.nz/data/sql/user"
	"nighttroop.t1d.nz/utils"
	"nighttroop.t1d.nz/utils/consts"
)

func (b *SyncJobManager) syncDexcom(syncInfo *SyncJobInfo) error {
	if len(syncInfo.userSyncSources) == 0 {
		return fmt.Errorf("No user sync sources found for unique hash: %s", syncInfo.UniqueHash)
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
			eid := utils.CreateEID("dexcom", e.SystemTime.UnixMilli())
			res, err := userSyncSource.userDB.Queries.CreateGlucoseReadingIgnoreEIDConstraint(context.TODO(), user.CreateGlucoseReadingIgnoreEIDConstraintParams{
				Eid:                   eid,
				OccurredAt:            e.SystemTime.UnixMilli(),
				Value:                 float64(e.Value),
				CreatedByIdentityID:   userSyncSource.syncIdentity.ID,
				RawJsonData:           &e.RawJson,
				CreatedByIdentityName: userSyncSource.syncIdentity.Name,
			})
			if err != nil {
				// Probably just non-unique eid
				slog.Error("Error creating entry in database", "error", err, "eid", eid, "sync_source_id", userSyncSource.syncIdentity.ID)
				continue
			}
			count, _ := res.RowsAffected()
			if count > 0 {
				addedColumns = true
			}

		}

		// Update that it has been seen and used recently
		if addedColumns {
			slog.Info("Added new Dexcom entries for user", "subdomain", userSyncSource.userModel.Subdomain)
			userSyncSource.userDB.UpdateIdentityDateFlags(userSyncSource.syncIdentity, true, true)
			totalAddedColumns += 1
		}
	}

	return nil
}
