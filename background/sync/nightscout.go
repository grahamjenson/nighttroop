package sync

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"nighttroop.t1d.nz/clients/nightscout_client"
	"nighttroop.t1d.nz/data/sql/user"
	"nighttroop.t1d.nz/utils"
)

func (b *SyncJobManager) syncNightscout(syncInfo *SyncJobInfo) error {
	if len(syncInfo.userSyncSources) == 0 {
		return fmt.Errorf("Nightscout: No user sync sources found for unique hash: %s", syncInfo.UniqueHash)
	}

	nsc := nightscout_client.NewClient(syncInfo.urlOrUsername, syncInfo.password)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	egvs, err := nsc.GetEntries(ctx, 3)
	if err != nil {
		return fmt.Errorf("Nightscout Error: %w", err)
	}

	// Limit the total egvs to the number requested
	if len(egvs) > 3 {
		egvs = egvs[:3]
	}

	totalAddedColumns := 0
	for _, userSyncSource := range syncInfo.userSyncSources {
		addedColumns := false
		for _, e := range egvs {
			eid := utils.CreateEID("nightscout", e.Date.Int64())
			dbModel := e.ToDBModel()
			dbModel.Eid = eid
			dbModel.CreatedByIdentityID = userSyncSource.syncIdentity.ID
			dbModel.CreatedByIdentityName = userSyncSource.syncIdentity.Name

			res, err := userSyncSource.userDB.Queries.CreateGlucoseReading(context.TODO(), dbModel.ToDBCreateParams())
			if err != nil {
				if user.IsDuplicateEID(err) {
					continue
				}
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
			slog.Info("Added new Nightscout entries for user", "subdomain", userSyncSource.userModel.Subdomain)
			userSyncSource.userDB.UpdateIdentityDateFlags(userSyncSource.syncIdentity, true, true)
			totalAddedColumns += 1
		}
	}

	return nil
}
