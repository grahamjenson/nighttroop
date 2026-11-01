package sitedb_test

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/sitedb"
)

// A DB kept in a file has one connection, so two writes that each read first, like a Nightscout create checking for a
// duplicate, take turns. With two connections the second write's read is stale, and it fails as the database is locked.
func TestFileDBConcurrentReadThenWrite(t *testing.T) {
	ctx := context.Background()
	sqldb, err := sitedb.OpenSqliteDB(filepath.Join(t.TempDir(), "site.sqlite"), 0)
	require.NoError(t, err)
	t.Cleanup(func() { sqldb.Close() })
	require.NoError(t, sitedb.MigrateUp(sqldb))
	q := sitedb.New(sqldb)

	readThenWrite := func(id string, afterRead func()) error {
		tx, err := sqldb.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer tx.Rollback()
		qtx := q.WithTx(tx)
		if _, err := qtx.GetInsulinDose(ctx, sitedb.GetInsulinDoseParams{ID: id}); err != sql.ErrNoRows {
			return err
		}
		afterRead()
		if _, err := qtx.CreateInsulinDose(ctx, sitedb.CreateInsulinDoseParams{ID: id, Dose: 1, OccurredAt: 1, CreatedByIdentityName: "test"}); err != nil {
			return err
		}
		return tx.Commit()
	}

	// a reads, then b reads, or waits for the connection a holds, then a writes, then b writes
	bRead, aDone, bDone := make(chan struct{}), make(chan struct{}), make(chan error, 1)
	require.NoError(t, readThenWrite("a", func() {
		go func() { bDone <- readThenWrite("b", func() { close(bRead); <-aDone }) }()
		for {
			select {
			case <-bRead:
				return
			case <-time.After(time.Millisecond):
				if sqldb.Stats().WaitCount > 0 {
					return
				}
			}
		}
	}))
	close(aDone)
	require.NoError(t, <-bDone)

	for _, id := range []string{"a", "b"} {
		_, err := q.GetInsulinDose(ctx, sitedb.GetInsulinDoseParams{ID: id})
		require.NoError(t, err, id)
	}
}
