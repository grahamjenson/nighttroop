package sync

import (
	"context"
	"log/slog"
	"time"

	"nighttroop.t1d.nz/data/sql/control"
	"nighttroop.t1d.nz/utils/consts"
)

// This is a background job runner that loops over users and syncs some data (initially just entries) from Nighscout and/or Dexcom.
//
// The big problem is the user defines their own nightscout URL.
// If the user is malicious, they could intend to slow down Nighttroop or DDOS another site.
//
// They could slow down Nighttroop by:
// 1. having it point at Nighttroop (syncing to itself)
// 1. Pointing at a server that takes a long time to respond or never responds
// 1. Pointing to a server that sometimes responds and sometimes doesn't, causing timeouts and retries
// 1. Returning more legit data than requested and causing massive times spent processing data
//
// They could DDOS another site by:
// 1. pointing many users at the same site.
// 1. pointing at an expencive call on another site causing issues there
// 1. Pointing to different sumdomains of a DNS they control that points to the same server (e.g. a.example.com, b.example.com, c.example.com all point to the same server)
//
// To deal with these issues:
// 1. Cant point URL at Nighttroop (or any subdomain of Nighttroop)
// 2. If a users sync is taking too long or failing decrease the rate at which they sync (until they are disabled)
// 3. Only sync once from a source:password combination (prevents multiple users pointing at the same URL and causing a DDOS)
// 4. Agressive timeouts for jobs as well as logging
// 5. List of bad and blocked hostnames, that are never allowed to be synced from (TODO)
// 6. We only Sync users that have called a read endpoint recently.
// 7. We slow down syncing of users that don't actually get new data (count writes to the DB)

// Dexcom is easier, but if say the dexcom servers are down we might encounter lots of timeouts, which we should detect quickly and slow down all syncing.

// Background Jobs

/////////////
// Actual SyncJob Work
//////////////

func (b *SyncJobManager) SyncUserData(syncInfo *SyncJobInfo) error {
	var err error

	// TODO CHECK REPORT AND MAKE SURE NOT ABUSING
	// Check that enought time has passed since the last sync to rerun sync

	startTime := time.Now()

	switch consts.SyncSourceType(syncInfo.Type) {
	case consts.SyncSourceTypeNightscout:
		err = b.syncNightscout(syncInfo)
	case consts.SyncSourceTypeDexcom:
		err = b.syncDexcom(syncInfo)
	}
	elapsedTimeMS := time.Since(startTime).Milliseconds()

	if err != nil {
		errMsg := err.Error()
		slog.Error("error syncing user data", "unique_hash", syncInfo.UniqueHash, "sync_type", syncInfo.Type, "error", err)
		updateError := b.controlDB.Queries.UpdateSyncSourceReportsFailure(context.TODO(), control.UpdateSyncSourceReportsFailureParams{
			ErrorMessages: []string{errMsg},
			ElapsedMs:     &elapsedTimeMS,
			UniqueHash:    syncInfo.UniqueHash,
		})
		if updateError != nil {
			slog.Error("error updating sync source report with failure", "unique_hash", syncInfo.UniqueHash, "sync_type", syncInfo.Type, "error", updateError)
		}
		return err
	}

	slog.Info("finished sync", "unique_hash", syncInfo.UniqueHash, "sync_type", syncInfo.Type, "elapsed_time_ms", elapsedTimeMS)
	updateError := b.controlDB.Queries.UpdateSyncSourceReportsSuccess(context.TODO(), control.UpdateSyncSourceReportsSuccessParams{
		ElapsedMs:  &elapsedTimeMS,
		UniqueHash: syncInfo.UniqueHash,
	})
	if updateError != nil {
		slog.Error("error updating sync source report with success", "unique_hash", syncInfo.UniqueHash, "sync_type", syncInfo.Type, "error", updateError)
		return updateError
	}

	return nil
	// TODO this is where we will sync data from the user DB to the control DB, and also do any processing needed
	// This will be called on a schedule, and also on demand when a user logs in or something like that
}
