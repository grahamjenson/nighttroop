package sitedb_test

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/utils/testutils"
)

// Many requests at once share a site DB's queries, which also run in a transaction that is rolled back or committed
func TestSiteDBQueriesShared(t *testing.T) {
	db, err := sitedb.NewDB(context.Background(), "abcdefgh", testutils.GetTestServerConfig())
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	ctx := context.Background()
	create := func(q *sitedb.Queries, id string) error {
		_, err := q.CreateInsulinDose(ctx, sitedb.CreateInsulinDoseParams{ID: id, Dose: 1, OccurredAt: 1, CreatedByIdentityName: "test"})
		return err
	}
	found := func(id string) bool {
		_, err := db.Queries().GetInsulinDose(ctx, sitedb.GetInsulinDoseParams{ID: id})
		return err == nil
	}

	var wg sync.WaitGroup
	queries := make([]*sitedb.Queries, 20)
	for i := range queries {
		wg.Go(func() {
			queries[i] = db.Queries()
			assert.NoError(t, create(queries[i], fmt.Sprintf("at-once-%d", i)))
		})
	}
	wg.Wait()
	for i, q := range queries {
		assert.Same(t, queries[0], q)
		assert.True(t, found(fmt.Sprintf("at-once-%d", i)))
	}

	for _, commit := range []bool{false, true} {
		id := fmt.Sprintf("tx-commit-%v", commit)
		tx, err := db.SQLiteDB.BeginTx(ctx, nil)
		require.NoError(t, err)
		require.NoError(t, create(db.QueriesWithTx(tx), id))
		if commit {
			require.NoError(t, tx.Commit())
		} else {
			require.NoError(t, tx.Rollback())
		}
		require.Equal(t, commit, found(id), id)
	}
}
