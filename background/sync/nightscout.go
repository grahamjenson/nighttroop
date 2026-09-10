package sync

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"nighttroop.t1d.nz/clients/nightscout_client"
	"nighttroop.t1d.nz/server/nightscout_api/nsutils"
	"nighttroop.t1d.nz/utils/httputils"
)

func (b *SyncJobManager) syncNightscout(syncInfo *SyncJobInfo) error {
	if len(syncInfo.userSyncSources) == 0 {
		return fmt.Errorf("Nightscout: No user sync sources found for unique hash: %s", syncInfo.UniqueHash)
	}

	nsc := nightscout_client.NewClient(syncInfo.urlOrUsername, syncInfo.password)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	afterTime := time.Now().Add(-time.Duration(20) * time.Minute)
	egvs, err := nsc.GetEntries(ctx, afterTime, 3)
	if err != nil {
		return fmt.Errorf("Nightscout Error: %w", err)
	}

	totalAddedColumns := 0
	for _, userSyncSource := range syncInfo.userSyncSources {
		siteContext := &httputils.SiteContext{
			Subdomain:     userSyncSource.siteModel.Subdomain,
			IdentityModel: userSyncSource.syncIdentity,
			ControlDB:     userSyncSource.controlDB,
			SiteDB:        userSyncSource.siteDB,
			SiteModel:     userSyncSource.siteModel,
			SiteAction:    httputils.SiteActionWrite,
		}

		addedColumns := false

		collectionsStore, ok := b.nightscoutStore.GetStore(nsutils.Entries)
		if !ok {
			return fmt.Errorf("Nightscout: Entries collection not found")
		}

		for _, egv := range egvs {
			_, created, err := collectionsStore.Create(ctx, siteContext, egv, false)
			if err != nil {
				return fmt.Errorf("Nightscout: Failed to create entry: %w", err)
			}
			if created {
				addedColumns = true
			}
		}
		// Update that it has been seen and used recently
		if addedColumns {
			slog.Info("Added new Nightscout entries for user", "subdomain", userSyncSource.siteModel.Subdomain)
			userSyncSource.controlDB.UpdateIdentityDateFlags(userSyncSource.syncIdentity, true, true)
			totalAddedColumns += 1
		}
	}

	return nil
}
