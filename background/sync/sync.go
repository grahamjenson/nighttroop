package sync

import (
	stderrors "errors"
	"log/slog"
	"time"

	"nighttroop.t1d.nz/clients/dexcom_client"
	"nighttroop.t1d.nz/data/events"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/models/sitemodels"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/errors"
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
// 5. Blocked hostnames and private networks are never synced from (httputils.IsNightscoutURLBlocked)
// 6. We only Sync users that have called a read endpoint recently.
// 7. We slow down syncing of users that don't actually get new data (count writes to the DB)

// Dexcom is easier, but if say the dexcom servers are down we might encounter lots of timeouts, which we should detect quickly and slow down all syncing.

// syncTimeout is all the time a Nightscout, Dexcom or Libre sync has, from its first request to its last write. A slow
// or down source fails fast and is tried again later.
const syncTimeout = 10 * time.Second

// Background Jobs

/////////////
// Actual SyncJob Work
//////////////

func (b *SyncJobManager) SyncUserData(syncInfo *SyncJobInfo) error {
	var err error
	added := 0

	startTime := time.Now()

	switch consts.SyncSourceType(syncInfo.Type) {
	case consts.SyncSourceTypeNightscout:
		added, err = b.syncNightscout(syncInfo)
	case consts.SyncSourceTypeDexcom:
		added, err = b.syncDexcom(syncInfo)
	case consts.SyncSourceTypeLibre:
		added, err = b.syncLibre(syncInfo)
	case consts.SyncSourceTypeVirtual:
		added, err = b.syncVirtual(syncInfo)
	}
	elapsedTimeMS := time.Since(startTime).Milliseconds()

	if err != nil {
		// A source refusing its login is its owner's to fix on the sync page, which shows the error, not the server's
		level := slog.LevelError
		if stderrors.Is(err, dexcom_client.ErrLoginRefused) {
			level = slog.LevelWarn
		}
		slog.Log(syncInfo.ctx, level, "error syncing user data", "error", err)
		b.recordSyncFailure(syncInfo, err, elapsedTimeMS)
		return err
	}

	slog.InfoContext(syncInfo.ctx, "finished sync", "elapsed_time_ms", elapsedTimeMS)
	if added > 0 {
		b.controlDB.Logs.SyncAddedEvent(syncInfo.ctx, syncInfo.SiteID, syncInfo.Type, added)
	}
	updateError := b.controlDB.Queries.UpdateSyncSourceReportsSuccess(syncInfo.ctx, controldb.UpdateSyncSourceReportsSuccessParams{
		ElapsedMs: &elapsedTimeMS,
		Type:      syncInfo.Type,
		SiteID:    syncInfo.SiteID,
	})
	if updateError != nil {
		slog.ErrorContext(syncInfo.ctx, "error updating sync source report with success", "error", updateError)
		return updateError
	}
	b.publishSyncSource(syncInfo.SiteID, syncInfo.Type)

	return nil
}

// publishSyncSource says a site's sync source has new stats or errors, so its sync settings page shows them as they
// change. Only its type is sent: the page reads the source again.
func (b *SyncJobManager) publishSyncSource(siteID string, sourceType consts.SyncSourceType) {
	if b.config != nil && b.config.Events != nil {
		b.config.Events.Publish(siteID, events.ModelChange(siteID, events.Updated, &sitemodels.SyncSource{Type: sourceType}))
	}
}

// recordSyncFailure stores the failure so the source backs off, and in the logs DB's errors. The message is
// truncated first: an oversized one would fail the row's CHECK and leave the source retrying every tick.
func (b *SyncJobManager) recordSyncFailure(syncInfo *SyncJobInfo, err error, elapsedTimeMS int64) {
	b.controlDB.Logs.SyncError(syncInfo.ctx, syncInfo.SiteID, syncInfo.Type, err)
	errMsg := errors.Truncate(err.Error())
	if syncInfo.firstSync {
		b.controlDB.Logs.SyncTurnedOffEvent(syncInfo.ctx, syncInfo.SiteID, syncInfo.Type)
		errMsg = "The first sync failed, so syncing has been turned off. Check the details and save to try again. " + errMsg
	}
	updateError := b.controlDB.Queries.UpdateSyncSourceReportsFailure(syncInfo.ctx, controldb.UpdateSyncSourceReportsFailureParams{
		ErrorMessages: errors.ErrorList{errMsg},
		ElapsedMs:     &elapsedTimeMS,
		Type:          syncInfo.Type,
		SiteID:        syncInfo.SiteID,
	})
	if updateError != nil {
		slog.ErrorContext(syncInfo.ctx, "error updating sync source report with failure", "error", updateError)
	}
	b.publishSyncSource(syncInfo.SiteID, syncInfo.Type)
}
