package connectors

import (
	"cmp"
	"net/url"
	"slices"
	"strings"
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
	// timedSyncGap is the shortest time between a site's timed syncs, and between any timed syncs of one Nightscout host,
	// however many sites sync from it, so a host gets at most one every 10 seconds on top of its usual syncs
	timedSyncGap = 10 * time.Second
	// slowSync is how long a sync can take before its Nightscout is left to its usual syncs, as one that slow may be
	// struggling
	slowSync = 2 * time.Second
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

// entriesDue are the sites whose next Nightscout entry is due about now, each at most every timedSyncGap, and at most
// one a host every timedSyncGap, the one that has waited longest. A site whose client is syncing now, or whose last
// sync was slow, is left out.
func (b *ConnectorJobManager) entriesDue(now time.Time) []string {
	b.timedHostsLock.Lock()
	defer b.timedHostsLock.Unlock()
	ms, gap := now.UnixMilli(), timedSyncGap.Milliseconds()
	for host, last := range b.timedHosts {
		if ms >= last+gap {
			delete(b.timedHosts, host)
		}
	}
	type dueSite struct {
		siteID string
		c      *nightscoutClient
	}
	waiting := []dueSite{}
	for _, siteID := range b.nightscoutClients.Keys() {
		c, ok := b.nightscoutClients.Peek(siteID)
		if !ok {
			continue
		}
		next := c.nextEntryAt.Load()
		if next != 0 && ms >= next-entryEarly.Milliseconds() && ms <= next+entryLate.Milliseconds() &&
			ms >= c.lastTimedSync.Load()+gap && !c.syncing.Load() && c.lastSyncTook.Load() <= slowSync.Milliseconds() {
			waiting = append(waiting, dueSite{siteID, c})
		}
	}
	slices.SortStableFunc(waiting, func(a, b dueSite) int { return cmp.Compare(a.c.lastTimedSync.Load(), b.c.lastTimedSync.Load()) })
	due := []string{}
	for _, site := range waiting {
		host := nightscoutHost(site.c.url)
		if _, taken := b.timedHosts[host]; taken {
			continue
		}
		b.timedHosts[host] = ms
		site.c.lastTimedSync.Store(ms)
		due = append(due, site.siteID)
	}
	return due
}

// nightscoutHost is the host a Nightscout URL is on, as one host is written many ways, or the URL when it has none
func nightscoutHost(nightscoutURL string) string {
	parsed, err := url.Parse(nightscoutURL)
	if err != nil || parsed.Hostname() == "" {
		return nightscoutURL
	}
	return strings.TrimSuffix(strings.ToLower(parsed.Hostname()), ".")
}
