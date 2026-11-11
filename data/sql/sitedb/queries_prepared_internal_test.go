package sitedb

import (
	"database/sql"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/consts"
)

// cached is the statement db keeps for query, or nil
func cached(db *DB, query string) *sql.Stmt {
	s, _ := db.stmts.stmts.Load(query)
	stmt, _ := s.(*sql.Stmt)
	return stmt
}

// A site's queries are each prepared the first time it runs one, then reused, in and out of a transaction, so SQLite
// does not parse them again, while preparing all of them at open kept about 3.5MB of memory for each site
func TestSiteQueriesArePreparedOnFirstUse(t *testing.T) {
	for _, mode := range []consts.ServerMode{consts.Test, consts.Development} { // in memory, then in a file
		t.Run(string(mode), func(t *testing.T) {
			ctx := t.Context()
			db, err := NewDB(ctx, "abcdefgh", &config.ServerConfig{Mode: mode, DataDir: t.TempDir()})
			require.NoError(t, err)
			t.Cleanup(func() { db.Close() })
			require.Nil(t, cached(db, allActivities), "nothing is prepared at open")

			// Prepared once, then reused, while a query never run is never prepared
			_, err = db.Queries().AllActivities(ctx, AllActivitiesParams{Limit: 1})
			require.NoError(t, err)
			stmt := cached(db, allActivities)
			require.NotNil(t, stmt)
			_, err = db.Queries().AllActivities(ctx, AllActivitiesParams{Limit: 1})
			require.NoError(t, err)
			require.Same(t, stmt, cached(db, allActivities))
			require.Nil(t, cached(db, allTempTargets))

			// Many first runs at once each succeed, and keep one statement
			var wg sync.WaitGroup
			for range 10 {
				wg.Go(func() {
					_, err := db.Queries().AllTempTargets(ctx, AllTempTargetsParams{Limit: 1})
					assert.NoError(t, err)
				})
			}
			wg.Wait()
			require.NotNil(t, cached(db, allTempTargets))

			// A transaction runs the cached statement on its own connection. A query first run in one is prepared
			// for later, except in memory, where the transaction holds the only connection, so it runs unprepared.
			tx, err := db.SQLiteDB.BeginTx(ctx, nil)
			require.NoError(t, err)
			_, err = db.QueriesWithTx(tx).AllActivities(ctx, AllActivitiesParams{Limit: 1})
			require.NoError(t, err)
			require.Same(t, stmt, cached(db, allActivities))
			require.NotNil(t, (&txStmts{Tx: tx, cache: db.stmts}).stmt(ctx, allActivities))
			_, err = db.QueriesWithTx(tx).CreateGlucoseReading(ctx, CreateGlucoseReadingParams{ID: "1", Mgdl: 100, ReadingType: "sgv", OccurredAt: 1, CreatedByIdentityName: "test"})
			require.NoError(t, err)
			require.NoError(t, tx.Commit())
			require.Equal(t, mode == consts.Development, cached(db, createGlucoseReading) != nil)

			// A query that fails to prepare is not kept, so it is tried again
			_, err = db.stmts.stmt(ctx, "SELECT nothing FROM nowhere")
			require.Error(t, err)
			require.Nil(t, cached(db, "SELECT nothing FROM nowhere"))

			// Closing the DB closes them all
			require.NoError(t, db.Close())
			count := 0
			db.stmts.stmts.Range(func(_, _ any) bool { count++; return true })
			require.Zero(t, count, "statements left open")
		})
	}
}
