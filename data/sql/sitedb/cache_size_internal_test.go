package sitedb

import (
	"database/sql/driver"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"modernc.org/sqlite"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/consts"
)

// A site DB's connection keeps a page cache of the 100KiB it asks for, not SQLite's default of 2MiB, which it kept
// when the size was set with _cache_size, as the driver does not read it. The size is read from SQLite's own count of
// the cache's memory, as a test runs no SQL of its own.
func TestSitePageCacheSize(t *testing.T) {
	ctx := t.Context()
	db, err := NewDB(ctx, "abcdefgh", &config.ServerConfig{Mode: consts.Development, DataDir: t.TempDir()})
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	// About 1MB of readings, read back on one connection
	tx, err := db.SQLiteDB.BeginTx(ctx, nil)
	require.NoError(t, err)
	note := strings.Repeat("n", 300)
	for i := range 3000 {
		_, err = db.QueriesWithTx(tx).CreateGlucoseReading(ctx, CreateGlucoseReadingParams{ID: fmt.Sprint(i), Mgdl: 100,
			ReadingType: "sgv", OccurredAt: int64(i + 1), Notes: &note, CreatedByIdentityName: "test"})
		require.NoError(t, err)
	}
	require.NoError(t, tx.Commit())
	conn, err := db.SQLiteDB.Conn(ctx)
	require.NoError(t, err)
	defer conn.Close()
	for skip := int64(0); skip < 3000; skip += 1000 {
		readings, err := New(conn).SearchGlucoseReadings(ctx, SearchGlucoseReadingsParams{Skip: skip, Limit: 1000})
		require.NoError(t, err)
		require.Len(t, readings, 1000)
	}

	require.NoError(t, conn.Raw(func(dc any) error {
		status := dc.(interface{ Unwrap() driver.Conn }).Unwrap().(sqlite.DBStatus)
		used, _, err := status.Status(sqlite.DBStatusCacheUsed, false)
		require.Greater(t, used, 50*1024, "the cache is used")
		require.Less(t, used, 300*1024, "the cache is about its size, with its pages' headers, while 2MiB kept 1.4MB of this DB")
		return err
	}))
}
