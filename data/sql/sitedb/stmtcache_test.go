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

// The site DB prepares its queries when they first run, so the same query runs on its own, in a transaction that is
// rolled back or committed, and from many requests at once
func TestSiteDBQueriesPreparedOnFirstUse(t *testing.T) {
	_, _, site := testutils.NewControlDBWithSite(t)
	ctx := context.Background()
	create := func(q *sitedb.Queries, id string) error {
		_, err := q.CreateInsulinDose(ctx, sitedb.CreateInsulinDoseParams{ID: id, Dose: 1, OccurredAt: 1, CreatedByIdentityName: "test"})
		return err
	}
	found := func(id string) bool {
		_, err := site.SiteDB.Queries.GetInsulinDose(ctx, sitedb.GetInsulinDoseParams{ID: id})
		return err == nil
	}

	require.NoError(t, create(site.SiteDB.Queries, "alone"))
	require.True(t, found("alone"))

	for _, commit := range []bool{false, true} {
		id := fmt.Sprintf("tx-commit-%v", commit)
		tx, err := site.SiteDB.SQLiteDB.BeginTx(ctx, nil)
		require.NoError(t, err)
		require.NoError(t, create(site.SiteDB.WithTx(tx), id))
		if commit {
			require.NoError(t, tx.Commit())
		} else {
			require.NoError(t, tx.Rollback())
		}
		require.Equal(t, commit, found(id), id)
	}

	var wg sync.WaitGroup
	for i := range 20 {
		wg.Go(func() {
			assert.NoError(t, create(site.SiteDB.Queries, fmt.Sprintf("at-once-%d", i)))
			assert.True(t, found(fmt.Sprintf("at-once-%d", i)))
		})
	}
	wg.Wait()
}
