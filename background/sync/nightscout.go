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
	"nighttroop.t1d.nz/server/nightscoutapi/nsutils"
)

// nightscoutClient is a site's Nightscout client, kept between syncs so its JWT is reused
// until it expires rather than requested every sync. Its lock stops two syncs using it at once.
type nightscoutClient struct {
	gosync.Mutex
	url, token string
	*nightscout_client.Client
	// lastModified is Nightscout's lastModified for each collection when it was last read; a new client reads them all
	lastModified map[nsutils.Collection]int64
}

// nightscoutClientFor returns the site's kept client, or a new one when there is none or its URL or token has changed.
// There is at most one per site, so saving new details many times does not grow them, and at most maxNightscoutClients.
func (b *SyncJobManager) nightscoutClientFor(syncInfo *SyncJobInfo) *nightscoutClient {
	b.nightscoutClientsLock.Lock()
	defer b.nightscoutClientsLock.Unlock()
	c, _ := b.nightscoutClients.Get(syncInfo.SiteID)
	if c == nil || c.url != syncInfo.urlOrUsername || c.token != syncInfo.password {
		c = &nightscoutClient{url: syncInfo.urlOrUsername, token: syncInfo.password, Client: nightscout_client.NewClient(syncInfo.urlOrUsername, syncInfo.password, nightscout_client.ClientOptions{AllowLocalAddresses: b.config.IsDev()}), lastModified: map[nsutils.Collection]int64{}}
		b.nightscoutClients.Add(syncInfo.SiteID, c)
	}
	return c
}

// syncNightscout copies the changed collections' newest documents into the site, returning how many were new
func (b *SyncJobManager) syncNightscout(syncInfo *SyncJobInfo) (int, error) {
	nsc := b.nightscoutClientFor(syncInfo)
	nsc.Lock()
	defer nsc.Unlock()

	ctx, cancel := context.WithTimeout(syncInfo.ctx, syncTimeout)
	defer cancel()

	// One small call says which collections have changed, so unchanged ones are not read every sync
	lastModified, err := nsc.GetLastModified(ctx)
	if err != nil {
		return 0, fmt.Errorf("NightscoutSync Error: %w", err)
	}

	added := 0
	for _, c := range nightscoutCollections {
		modified := lastModified[string(c.collection)]
		if !changedSince(nsc.lastModified[c.collection], modified) {
			continue
		}
		collectionAdded, err := b.syncNightscoutCollection(ctx, nsc, syncInfo, c.collection, c.limit)
		added += collectionAdded
		if err != nil {
			return added, err
		}
		nsc.lastModified[c.collection] = modified
	}
	return added, nil
}

// nightscoutCollections are what a Nightscout sync reads, and at most how many of each one's newest documents
// from the last 20 minutes: a few readings, a meal, its boluses and a loop's SMBs and temp basals, or each uploader's
// statuses. A backfill reads them a page at a time, small enough to stay under the client's response size limit.
var nightscoutCollections = []struct {
	collection   nsutils.Collection
	limit        int
	backfillPage int
}{
	{nsutils.Entries, 3, 500},
	{nsutils.Treatments, 20, 200},
	{nsutils.Devicestatus, 10, 50},
}

// syncNightscoutCollection copies a collection's newest documents into the site, returning how many were new
func (b *SyncJobManager) syncNightscoutCollection(ctx context.Context, nsc *nightscoutClient, syncInfo *SyncJobInfo, collection nsutils.Collection, limit int) (int, error) {
	docs, err := nsc.GetDocuments(ctx, string(collection), time.Now().Add(-20*time.Minute), time.Time{}, limit)
	if err != nil {
		return 0, fmt.Errorf("NightscoutSync Error: %w", err)
	}
	return b.createNightscoutDocuments(ctx, syncInfo, collection, docs)
}

// createNightscoutDocuments copies documents into the site, returning how many were new. Documents a site refuses,
// such as a treatment with an eventType it does not know, are skipped and logged once, so they do not fail every sync.
func (b *SyncJobManager) createNightscoutDocuments(ctx context.Context, syncInfo *SyncJobInfo, collection nsutils.Collection, docs []map[string]any) (int, error) {
	// REMOVE WRITE ONLY COLUMNS
	for _, doc := range docs {
		delete(doc, "srvCreated")
		delete(doc, "srvModified")
		delete(doc, "modifiedBy")
	}

	collectionStore, ok := b.nightscoutStore.GetStore(collection)
	if !ok {
		return 0, fmt.Errorf("Nightscout: %s collection not found", collection)
	}
	siteContext := syncInfo.Site.siteContext("nightscout_sync")
	added, skipped, reason := 0, 0, ""
	for _, doc := range docs {
		_, duplicateFound, herr := collectionStore.Create(ctx, siteContext, doc, false)
		if herr != nil && herr.Status < http.StatusInternalServerError {
			skipped, reason = skipped+1, herr.Message
			continue
		}
		if herr != nil {
			return added, fmt.Errorf("Nightscout: Failed to create %s: %w", collection, herr)
		}
		if !duplicateFound {
			added++
		}
	}
	if skipped > 0 {
		b.controlDB.Logs.Log(ctx, &syncInfo.SiteID, db.LogLevelWarn, "skipped Nightscout documents the site refused", map[string]any{"collection": collection, "skipped": skipped, "error": reason})
	}
	if added > 0 {
		slog.InfoContext(ctx, "Added new Nightscout documents for site", "collection", collection, "added", added)
	}
	return added, nil
}

// changedSince is whether a collection may have changed since it was last read (0 when it never was).
// A Nightscout that does not say when it changed (0) has it read every sync.
func changedSince(lastRead, modified int64) bool {
	return modified == 0 || lastRead != modified
}
