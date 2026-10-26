package sync

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	gosync "sync"
	"time"

	"nighttroop.t1d.nz/clients/nightscout_client"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/server/nightscoutapi/nsutils"
	"nighttroop.t1d.nz/utils/httputils"
)

// nightscoutClient is a site's Nightscout client, kept between syncs so its JWT is reused
// until it expires rather than requested every sync. Its lock stops two syncs using it at once.
type nightscoutClient struct {
	gosync.Mutex
	url, token string
	*nightscout_client.Client
}

// nightscoutClientFor returns the site's kept client, or a new one when there is none or its URL or token has changed.
// There is at most one per site, so saving new details many times does not grow them.
func (b *SyncJobManager) nightscoutClientFor(syncInfo *SyncJobInfo) *nightscoutClient {
	b.nightscoutClientsLock.Lock()
	defer b.nightscoutClientsLock.Unlock()
	c := b.nightscoutClients[syncInfo.SiteID]
	if c == nil || c.url != syncInfo.urlOrUsername || c.token != syncInfo.password {
		c = &nightscoutClient{url: syncInfo.urlOrUsername, token: syncInfo.password, Client: nightscout_client.NewClient(syncInfo.urlOrUsername, syncInfo.password)}
		b.nightscoutClients[syncInfo.SiteID] = c
	}
	return c
}

func (b *SyncJobManager) syncNightscout(syncInfo *SyncJobInfo) error {
	if len(syncInfo.UserSyncSources) == 0 {
		return fmt.Errorf("Nightscout: No user sync sources found for site_id %s type %s", syncInfo.SiteID, syncInfo.Type)
	}

	nsc := b.nightscoutClientFor(syncInfo)
	nsc.Lock()
	defer nsc.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	// One small call says which collections have changed, so unchanged ones are not read every sync
	lastModified, err := nsc.GetLastModified(ctx)
	if err != nil {
		return fmt.Errorf("NightscoutSync Error: %w", err)
	}

	read := map[nsutils.Collection]*int64{}
	for _, c := range nightscoutCollections {
		modified := lastModified[string(c.collection)]
		if !changedSince(syncInfo.lastModified[c.collection], modified) {
			continue
		}
		if err := b.syncNightscoutCollection(ctx, nsc, syncInfo, c.collection, c.limit); err != nil {
			return err
		}
		read[c.collection] = &modified
	}
	if len(read) == 0 {
		return nil
	}
	return b.controlDB.Queries.UpdateSyncSourceLastModified(ctx, controldb.UpdateSyncSourceLastModifiedParams{
		EntriesLastModified:      read[nsutils.Entries],
		TreatmentsLastModified:   read[nsutils.Treatments],
		DevicestatusLastModified: read[nsutils.Devicestatus],
		Type:                     syncInfo.Type,
		SiteID:                   syncInfo.SiteID,
	})
}

// nightscoutCollections are what a Nightscout sync reads, and at most how many of each one's newest documents
// from the last 20 minutes: a few readings, a meal, its boluses and a loop's SMBs and temp basals, or each uploader's statuses
var nightscoutCollections = []struct {
	collection nsutils.Collection
	limit      int
}{
	{nsutils.Entries, 3},
	{nsutils.Treatments, 20},
	{nsutils.Devicestatus, 10},
}

// syncNightscoutCollection copies a collection's newest documents into each site. A document the site
// refuses, such as a treatment with an eventType it does not know, is logged and skipped, so it does
// not fail every sync until it is 20 minutes old.
func (b *SyncJobManager) syncNightscoutCollection(ctx context.Context, nsc *nightscoutClient, syncInfo *SyncJobInfo, collection nsutils.Collection, limit int) error {
	docs, err := nsc.GetDocuments(ctx, string(collection), time.Now().Add(-20*time.Minute), limit)
	if err != nil {
		return fmt.Errorf("NightscoutSync Error: %w", err)
	}

	// REMOVE WRITE ONLY COLUMNS
	for _, doc := range docs {
		delete(doc, "srvCreated")
		delete(doc, "srvModified")
		delete(doc, "modifiedBy")
	}

	collectionStore, ok := b.nightscoutStore.GetStore(collection)
	if !ok {
		return fmt.Errorf("Nightscout: %s collection not found", collection)
	}
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

		added := 0
		for _, doc := range docs {
			_, duplicateFound, herr := collectionStore.Create(ctx, siteContext, doc, false)
			if herr != nil && herr.Status < http.StatusInternalServerError {
				b.controlDB.Logs.Log(ctx, &syncInfo.SiteID, db.LogLevelWarn, "skipped a Nightscout document the site refused", map[string]any{"collection": collection, "error": herr.Message})
				continue
			}
			if herr != nil {
				return fmt.Errorf("Nightscout: Failed to create %s: %w", collection, herr)
			}
			if !duplicateFound {
				added++
			}
		}
		if added > 0 {
			slog.Info("Added new Nightscout documents for site", "subdomain", userSyncSource.SiteModel.ID, "collection", collection, "added", added)
		}
	}
	return nil
}

// changedSince is whether a collection may have changed since it was last read. A Nightscout
// that does not say when it changed (0) has it read every sync.
func changedSince(lastRead *int64, modified int64) bool {
	return modified == 0 || lastRead == nil || *lastRead != modified
}
