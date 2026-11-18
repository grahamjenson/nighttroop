package connectors

import (
	"time"
)

// A Nightscout uploader sends a CGM reading about every 5 minutes, so a working connector also syncs from just before
// the next entry is due until it comes, at most every timedSyncGap, to copy it soon after it is sent. When it is due is
// kept on the site's client in memory only: after a restart, or once its client is dropped, the site syncs as usual.
const (
	// entryEvery is how long after Nightscout's entries last changed the next entry is due
	entryEvery = 5 * time.Minute
	// entryEarly is how long before the next entry is due the timed syncs start
	entryEarly = 15 * time.Second
	// entryLate is how long after the next entry is due the timed syncs stop, if it has not come
	entryLate = 90 * time.Second
	// timedSyncGap is the shortest time between a site's timed syncs
	timedSyncGap = 10 * time.Second
)

// timeNextEntry sets when the client's next entry is due, after a sync that read when entries were last modified. A
// sync on the usual schedule that works sets it. A timed sync only clears it, once it sees a new entry, so the timed
// syncs stop when the usual ones do, as when the site is not seen. A sync that fails clears it.
func (c *nightscoutClient) timeNextEntry(entriesModified int64, timed bool, err error) {
	due := entriesModified + entryEvery.Milliseconds()
	switch {
	case err != nil || entriesModified == 0:
		c.nextEntryAt.Store(0)
	case !timed:
		c.nextEntryAt.Store(due)
	case due > c.nextEntryAt.Load():
		c.nextEntryAt.Store(0)
	}
}

// entriesDue are the sites whose next Nightscout entry is due about now, each at most every timedSyncGap
func (b *ConnectorJobManager) entriesDue(now time.Time) []string {
	ms, due := now.UnixMilli(), []string{}
	for _, siteID := range b.nightscoutClients.Keys() {
		c, ok := b.nightscoutClients.Peek(siteID)
		if !ok {
			continue
		}
		next := c.nextEntryAt.Load()
		if next == 0 || ms < next-entryEarly.Milliseconds() || ms > next+entryLate.Milliseconds() ||
			ms < c.lastTimedSync.Load()+timedSyncGap.Milliseconds() {
			continue
		}
		c.lastTimedSync.Store(ms)
		due = append(due, siteID)
	}
	return due
}
