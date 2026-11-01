package sitedb_test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/sitedb"
)

func reading(id string) sitedb.CreateGlucoseReadingParams {
	return sitedb.CreateGlucoseReadingParams{ID: id, Mgdl: 120, ReadingType: "sgv", OccurredAt: time.Now().UnixMilli(), CreatedByIdentityName: "phone"}
}

// A site's DB on disk has many connections, so a write can come between a transaction's read and its own write, as a
// background job's can between a Nightscout upload's duplicate check and its insert. The transaction holds the write
// lock from its start, so the other write waits for it rather than the transaction failing with SQLITE_BUSY.
func TestWriteTransactionsWaitForEachOther(t *testing.T) {
	ctx := context.Background()
	db, err := sitedb.OpenSqliteDB(filepath.Join(t.TempDir(), "site.sqlite"), 0)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	require.NoError(t, sitedb.MigrateUp(db))
	q, err := sitedb.Prepare(ctx, db)
	require.NoError(t, err)

	tx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err)
	_, err = q.WithTx(tx).SearchGlucoseReadings(ctx, sitedb.SearchGlucoseReadingsParams{Limit: 1})
	require.NoError(t, err)

	// Another connection writes after the transaction's read, before its write
	other := make(chan error, 1)
	go func() { _, err := q.CreateGlucoseReading(ctx, reading("other")); other <- err }()
	time.Sleep(200 * time.Millisecond)

	_, err = q.WithTx(tx).CreateGlucoseReading(ctx, reading("in transaction"))
	require.NoError(t, err)
	require.NoError(t, tx.Commit())
	require.NoError(t, <-other)

	readings, err := q.SearchGlucoseReadings(ctx, sitedb.SearchGlucoseReadingsParams{Limit: 10})
	require.NoError(t, err)
	require.Len(t, readings, 2)
}
