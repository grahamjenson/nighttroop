package usage_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/utils/usage"
)

// The memory shown is all the OS counts for the process, with what SQLite allocates outside Go's heap: preparing every
// query of a few site DBs takes tens of MB of it, which Go's heap barely sees
func TestServerMemoryCountsSQLite(t *testing.T) {
	tracker := usage.New()
	before := tracker.Server(t.TempDir(), time.Now())
	require.Positive(t, before.ResidentBytes)

	for range 20 {
		db, err := sitedb.OpenInMemorySqliteDB()
		require.NoError(t, err)
		t.Cleanup(func() { db.Close() })
		require.NoError(t, sitedb.MigrateUp(db))
		queries, err := sitedb.Prepare(context.Background(), db)
		require.NoError(t, err)
		t.Cleanup(func() { queries.Close() })
	}

	after := tracker.Server(t.TempDir(), time.Now())
	resident, heap := int64(after.ResidentBytes)-int64(before.ResidentBytes), int64(after.HeapBytes)-int64(before.HeapBytes)
	t.Logf("resident +%d MB, Go heap +%d MB", resident>>20, heap>>20)
	require.Greater(t, resident, int64(30<<20))
	require.Greater(t, resident, 3*heap)
}
