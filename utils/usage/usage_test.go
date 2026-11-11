package usage

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestSnapshotCountsActivityByDay(t *testing.T) {
	now := time.Date(2026, 10, 30, 12, 0, 0, 0, time.UTC)
	day := func(daysAgo int) time.Time { return now.AddDate(0, 0, -daysAgo) }
	tracker := New()
	// Today: site a twice by identity 1 sending 100 and 300 bytes, site b once by identity 2
	tracker.Seen("a", 1, 100, day(0))
	tracker.Seen("a", 1, 300, day(0))
	tracker.Seen("b", 2, -1, day(0))
	// Earlier this week: site a by identity 3; and before the week, site c, which is left out of it
	tracker.Seen("a", 3, 200, day(3))
	tracker.Seen("c", 4, 50, day(7))

	snapshot := tracker.Snapshot(now, 10)
	require.Equal(t, Activity{Sites: 2, Identities: 2, Requests: 3, Ingested: 400}, snapshot.Today)
	require.Equal(t, Activity{Sites: 2, Identities: 3, Requests: 4, Ingested: 600}, snapshot.Week)
	require.Equal(t, int64(650), snapshot.Ingested, "every byte since start, an unknown length counted as none")
	// The week's site days are a today 2 requests 400 bytes, b today 1 request 0 bytes, a 3 days ago 1 request 200
	require.Equal(t, 3, snapshot.RequestsPerSiteDay.N)
	require.InDelta(t, 4.0/3, snapshot.RequestsPerSiteDay.Mean, 0.001)
	require.InDelta(t, 0.4714, snapshot.RequestsPerSiteDay.Std, 0.001)
	require.InDelta(t, 200, snapshot.IngestedPerSiteDay.Mean, 0.001)

	// Only the last 8 days are kept: a week on, today is the oldest left
	tracker.Seen("d", 5, 0, day(-7))
	require.Len(t, tracker.days, 2)
}

func TestSnapshotRoutesMostAskedForFirst(t *testing.T) {
	tracker := New()
	for range 3 {
		tracker.Route("site GET /", 10*time.Millisecond)
	}
	tracker.Route("site POST /data/{collection}", 30*time.Millisecond)
	tracker.Route("control GET /usage", 20*time.Millisecond)

	routes := tracker.Snapshot(time.Now(), 2).Routes
	require.Equal(t, []RouteCount{{"site GET /", 3, 10}, {"control GET /usage", 1, 20}}, routes)
}

func TestNilTrackerCountsNothing(t *testing.T) {
	var tracker *Tracker
	tracker.Seen("a", 1, 10, time.Now())
	tracker.Route("site GET /", time.Second)
	require.Equal(t, Snapshot{}, tracker.Snapshot(time.Now(), 10))
	require.Positive(t, tracker.Server(".", time.Now()).CPUs)
}

func TestServerReadsTheProcessAndDisk(t *testing.T) {
	// The DB folder's files, a site's in a batch folder, are all that is counted as used
	dbDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dbDir, "site", "a"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dbDir, "control.sqlite"), make([]byte, 3000), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dbDir, "site", "a", "abc.sqlite"), make([]byte, 1234), 0o600))

	tracker := New()
	server := tracker.Server(dbDir, time.Now().Add(time.Second))
	require.Positive(t, server.CPUs)
	require.Positive(t, server.Goroutines)
	require.Positive(t, server.HeapBytes)
	require.GreaterOrEqual(t, server.CPUPercent, 0.0)
	require.Equal(t, uint64(4234), server.DiskUsed)
	require.Positive(t, server.DiskFree)
	require.Positive(t, server.UptimeMs)

	// Before the DB folder is made, nothing is used and the working folder's disk is free
	server = tracker.Server(filepath.Join(dbDir, "missing"), time.Now())
	require.Zero(t, server.DiskUsed)
	require.Positive(t, server.DiskFree)
}
