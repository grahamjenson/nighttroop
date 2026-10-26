package sync

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"nighttroop.t1d.nz/clients/nightscout_client"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/server/nightscoutapi/nsutils"
	"nighttroop.t1d.nz/utils/httputils"
)

func (b *SyncJobManager) syncNightscout(syncInfo *SyncJobInfo) error {
	if len(syncInfo.UserSyncSources) == 0 {
		return fmt.Errorf("Nightscout: No user sync sources found for site_id %s type %s", syncInfo.SiteID, syncInfo.Type)
	}

	nsc := nightscout_client.NewClient(syncInfo.urlOrUsername, syncInfo.password)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	// One small call says whether entries have changed, so unchanged entries are not read every sync
	lastModified, err := nsc.GetLastModified(ctx)
	if err != nil {
		return fmt.Errorf("NightscoutSync Error: %w", err)
	}
	entriesModified := lastModified[string(nsutils.Entries)]
	if !entriesChanged(syncInfo.entriesLastModified, entriesModified) {
		return nil
	}

	afterTime := time.Now().Add(-time.Duration(20) * time.Minute)
	egvs, err := nsc.GetEntries(ctx, afterTime, 3)
	if err != nil {
		return fmt.Errorf("NightscoutSync Error: %w", err)
	}

	// REMOVE WRITE ONLY COLUMNS
	for _, egv := range egvs {
		delete(egv, "srvCreated")
		delete(egv, "srvModified")
		delete(egv, "modifiedBy")
	}
	////
	//
	////

	totalAddedColumns := 0
	for _, userSyncSource := range syncInfo.UserSyncSources {
		siteContext := &httputils.SiteContext{
			Subdomain: userSyncSource.SiteModel.ID,
			IdentityModel: &controldb.SiteIdentity{
				Name: "nightscout_sync",
			},
			ControlDB:  userSyncSource.ControlDB,
			SiteDB:     userSyncSource.SiteDB,
			SiteModel:  userSyncSource.SiteModel,
			HTTPAction: httputils.HTTPActionWrite,
		}

		addedColumns := false

		collectionsStore, ok := b.nightscoutStore.GetStore(nsutils.Entries)
		if !ok {
			return fmt.Errorf("Nightscout: Entries collection not found")
		}

		for _, egv := range egvs {
			_, duplicateFound, err := collectionsStore.Create(ctx, siteContext, egv, false)
			if err != nil {
				return fmt.Errorf("Nightscout: Failed to create entry: %w", err)
			}
			if !duplicateFound {
				addedColumns = true
				totalAddedColumns += 1
			}
		}
		// Update that it has been seen and used recently
		if addedColumns {
			slog.Info("Added new Nightscout entries for user", "subdomain", userSyncSource.SiteModel.ID)

		}
	}

	return b.controlDB.Queries.UpdateSyncSourceEntriesLastModified(ctx, controldb.UpdateSyncSourceEntriesLastModifiedParams{
		EntriesLastModified: &entriesModified,
		Type:                syncInfo.Type,
		SiteID:              syncInfo.SiteID,
	})
}

// entriesChanged is whether entries may have changed since they were last read. A Nightscout
// that does not say when its entries changed (0) has them read every sync, as before.
func entriesChanged(lastRead *int64, modified int64) bool {
	return modified == 0 || lastRead == nil || *lastRead != modified
}
