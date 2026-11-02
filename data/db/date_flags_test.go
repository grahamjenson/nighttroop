package db_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/utils/testutils"
)

// countingDBTX counts the writes the control DB's generated queries make through it
type countingDBTX struct {
	controldb.DBTX
	writes int
}

func (c *countingDBTX) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	c.writes++
	return c.DBTX.ExecContext(ctx, query, args...)
}

// TestDateFlagsWriteOnlyWhenDue checks a request writes the site's and identity's date flags only when each was last
// recorded over 10 minutes ago, as each write waits on the control DB's one write lock
func TestDateFlagsWriteOnlyWhenDue(t *testing.T) {
	_, controlDB, site := testutils.NewControlDBWithSite(t)
	ctx, siteID := t.Context(), site.SiteDB.Subdomain
	counter := &countingDBTX{DBTX: controlDB.SQLiteDB}
	prepared := controlDB.Queries
	controlDB.Queries = controldb.New(counter)
	t.Cleanup(func() { controlDB.Queries = prepared })

	getSite := func() *controldb.Site {
		model, err := controlDB.Queries.GetSiteInAnyState(ctx, controldb.GetSiteInAnyStateParams{ID: siteID})
		require.NoError(t, err)
		return model
	}
	identities, err := controlDB.Queries.AllSiteIdentities(ctx, controldb.AllSiteIdentitiesParams{SiteID: siteID})
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(identities), 2)
	getIdentity := func(identity *controldb.SiteIdentity) *controldb.SiteIdentity {
		model, err := controlDB.Queries.GetSiteIdentityByID(ctx, controldb.GetSiteIdentityByIDParams{SiteID: siteID, ID: identity.ID})
		require.NoError(t, err)
		return model
	}
	request := func(identity *controldb.SiteIdentity, write bool) int {
		counter.writes = 0
		require.NoError(t, controlDB.UpdateSiteAndIdentityDateFlags(ctx, getSite(), site.SiteDB, getIdentity(identity), true, write))
		return counter.writes
	}
	first, second := identities[0], identities[1]

	// Never seen, so the site and identity are both written
	require.Equal(t, 2, request(first, true))
	require.NotNil(t, getSite().LastWriteAt)
	require.NotNil(t, getIdentity(first).LastSeenAt)

	// Seen and written again within 10 minutes writes nothing
	require.Equal(t, 0, request(first, true))
	require.Equal(t, 0, request(first, false))

	// Another identity is recorded on its own, though the site was just seen
	require.Equal(t, 1, request(second, false))
	require.NotNil(t, getIdentity(second).LastSeenAt)
	require.Nil(t, getIdentity(second).LastWriteAt)

	// Over 10 minutes later both are due again
	before := time.Now().Add(-11 * time.Minute).UnixMilli()
	require.NoError(t, controlDB.Queries.UpdateSiteDateFlags(ctx, controldb.UpdateSiteDateFlagsParams{LastSeenAt: &before, LastWriteAt: &before, ID: siteID}))
	require.NoError(t, controlDB.Queries.UpdateSiteIdentityDateFlags(ctx, controldb.UpdateSiteIdentityDateFlagsParams{LastSeenAt: &before, LastWriteAt: &before, ID: first.ID}))
	require.Equal(t, 2, request(first, true))
	require.Greater(t, *getSite().LastWriteAt, before)
	require.Greater(t, *getIdentity(first).LastSeenAt, before)
}
