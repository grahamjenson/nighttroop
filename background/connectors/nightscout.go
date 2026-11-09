package connectors

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"nighttroop.t1d.nz/clients/nightscout_client"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/server/nightscoutapi/nsutils"
)

// nightscoutClient is a site's Nightscout client, kept between syncs so its JWT is reused
// until it expires rather than requested every sync. Its lock stops two syncs using it at once.
type nightscoutClient struct {
	sync.Mutex
	url, token string
	*nightscout_client.Client
	// lastModified is Nightscout's lastModified for each collection when it was last read; a new client reads them all
	lastModified map[nsutils.Collection]int64
}

// nightscoutClientFor returns the site's kept client, or a new one when there is none or its URL or token has changed.
// There is at most one per site, so saving new details many times does not grow them, and at most maxNightscoutClients.
func (b *ConnectorJobManager) nightscoutClientFor(jobInfo *ConnectorJobInfo) *nightscoutClient {
	b.nightscoutClientsLock.Lock()
	defer b.nightscoutClientsLock.Unlock()
	c, _ := b.nightscoutClients.Get(jobInfo.SiteID)
	if c == nil || c.url != jobInfo.urlOrUsername || c.token != jobInfo.password {
		c = &nightscoutClient{url: jobInfo.urlOrUsername, token: jobInfo.password, Client: nightscout_client.NewClient(jobInfo.urlOrUsername, jobInfo.password, nightscout_client.ClientOptions{AllowLocalAddresses: b.config.IsDev()}), lastModified: map[nsutils.Collection]int64{}}
		b.nightscoutClients.Add(jobInfo.SiteID, c)
	}
	return c
}

// syncNightscout copies the changed collections' newest documents into the site, returning how many were new
func (b *ConnectorJobManager) syncNightscout(jobInfo *ConnectorJobInfo) (int, error) {
	nsc := b.nightscoutClientFor(jobInfo)
	nsc.Lock()
	defer nsc.Unlock()

	ctx, cancel := context.WithTimeout(jobInfo.ctx, syncTimeout)
	defer cancel()

	// One small call says which collections have changed, so unchanged ones are not read every sync
	lastModified, err := nsc.GetLastModified(ctx)
	if err != nil {
		return 0, fmt.Errorf("Nightscout connector error: %w", err)
	}

	added := 0
	for _, c := range nightscoutCollections {
		modified := lastModified[string(c.collection)]
		if !changedSince(nsc.lastModified[c.collection], modified) {
			continue
		}
		collectionAdded, err := b.syncNightscoutCollection(ctx, nsc, jobInfo, c.collection, c.limit)
		added += collectionAdded
		if err != nil {
			return added, err
		}
		nsc.lastModified[c.collection] = modified
	}
	// A profile store is used long after it starts, so its newest is read whenever the profiles change, not by date
	if modified := lastModified[string(nsutils.Profile)]; changedSince(nsc.lastModified[nsutils.Profile], modified) {
		profileAdded, err := b.syncNightscoutProfile(ctx, nsc, jobInfo)
		added += profileAdded
		if err != nil {
			return added, err
		}
		nsc.lastModified[nsutils.Profile] = modified
	}
	return added, nil
}

// syncNightscoutProfile copies the newest profile store into the site, returning 1 when it was new
func (b *ConnectorJobManager) syncNightscoutProfile(ctx context.Context, nsc *nightscoutClient, jobInfo *ConnectorJobInfo) (int, error) {
	profile, err := nsc.GetCurrentProfile(ctx)
	if err != nil {
		return 0, fmt.Errorf("Nightscout connector error: %w", err)
	}
	if profile == nil {
		return 0, nil
	}
	// Nightscout's profile editor saves an edit to the same document, so it replaces the copy of it
	return b.createNightscoutDocuments(ctx, jobInfo, nsutils.Profile, []map[string]any{profile}, replaceDocuments)
}

// nightscoutCollections are what a Nightscout connector reads, and at most how many of each one's newest documents
// from the last 20 minutes: a few readings, a meal, its boluses and a loop's SMBs and temp basals, or each uploader's
// statuses. A backfill reads them a page at a time, small enough to stay under the client's response size limit, and
// saves them as backfillSave says.
var nightscoutCollections = []struct {
	collection   nsutils.Collection
	limit        int
	backfillPage int
	backfillSave saveDocuments
}{
	{nsutils.Entries, 3, 500, createDocuments},
	{nsutils.Treatments, 20, 200, createDocuments},
	// An old device status's loop status is of no use, and most of a looping site's DB; its IOB and COB are kept
	{nsutils.Devicestatus, 10, 50, createWithoutLoopStatus},
}

// syncNightscoutCollection copies a collection's newest documents into the site, returning how many were new
func (b *ConnectorJobManager) syncNightscoutCollection(ctx context.Context, nsc *nightscoutClient, jobInfo *ConnectorJobInfo, collection nsutils.Collection, limit int) (int, error) {
	docs, err := nsc.GetDocuments(ctx, string(collection), time.Now().Add(-20*time.Minute), time.Time{}, limit)
	if err != nil {
		return 0, fmt.Errorf("Nightscout connector error: %w", err)
	}
	return b.createNightscoutDocuments(ctx, jobInfo, collection, docs, createDocuments)
}

// saveDocuments is how createNightscoutDocuments saves the documents it copies
type saveDocuments int

const (
	// createDocuments adds each, unless the site has it already
	createDocuments saveDocuments = iota
	// replaceDocuments replaces the site's copy of a document with an _id, so a change to it is copied too
	replaceDocuments
	// createWithoutLoopStatus adds each, keeping no loop status of a device report it adds, as a backfill's are old
	createWithoutLoopStatus
)

// createNightscoutDocuments copies documents into the site, returning how many were new. Documents a site refuses,
// such as a treatment with an eventType it does not know, are skipped and logged once, so they do not fail every sync.
// With replace, a document with an _id replaces the site's copy of it, so a change to it is copied too.
func (b *ConnectorJobManager) createNightscoutDocuments(ctx context.Context, jobInfo *ConnectorJobInfo, collection nsutils.Collection, docs []map[string]any, save saveDocuments) (int, error) {
	// REMOVE WRITE ONLY COLUMNS, which that Nightscout set. Its subject is who saved the document there, not the sync
	// that copies it here, so a replace would be refused as changing it.
	for _, doc := range docs {
		delete(doc, "srvCreated")
		delete(doc, "srvModified")
		delete(doc, "modifiedBy")
		delete(doc, "subject")
	}

	collectionStore, ok := b.nightscoutStore.GetStore(collection)
	if !ok {
		return 0, fmt.Errorf("Nightscout: %s collection not found", collection)
	}
	siteContext := jobInfo.Site.siteContext("nightscout_connector")
	added, skipped, reason := 0, 0, ""
	for _, doc := range docs {
		var duplicateFound bool
		var herr *nsutils.HTTPError
		if identifier, _ := doc["_id"].(string); save == replaceDocuments && identifier != "" {
			var created bool
			_, created, herr = collectionStore.Replace(ctx, siteContext, identifier, doc, true)
			duplicateFound = !created
		} else {
			var stored nsutils.Document
			stored, duplicateFound, herr = collectionStore.Create(ctx, siteContext, doc)
			// Only a report it made loses its loop status, not one the site already had
			if herr == nil && !duplicateFound && save == createWithoutLoopStatus {
				if err := siteContext.SiteDB.Queries().ClearDeviceReportLoopStatus(ctx, sitedb.ClearDeviceReportLoopStatusParams{ID: stored.Identifier()}); err != nil {
					return added, fmt.Errorf("Nightscout: Failed to clear a backfilled %s's loop status: %w", collection, err)
				}
			}
		}
		if herr != nil && skipped == 0 {
			recordRefused(ctx, jobInfo, fmt.Errorf("Nightscout: the site refused a %s document: %s", collection, herr.Message), doc)
		}
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
		b.controlDB.Logs.Log(ctx, &jobInfo.SiteID, db.LogLevelWarn, "skipped Nightscout documents the site refused", map[string]any{"collection": collection, "skipped": skipped, "error": reason})
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
