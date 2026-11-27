package siteapiutils_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/server/siteapi/siteapiutils"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/httputils"
	"nighttroop.t1d.nz/utils/testutils"
)

// On DBs on disk, as in production, a page's read (NewReadCollectionTx) takes no write lock, so an upload to the site
// and a write to the control DB beside it do not wait for the page. A handler that writes (NewCollectionTx) holds the
// lock from its first query, so a write beside it waits.
func TestReadCollectionTxTakesNoWriteLock(t *testing.T) {
	cfg := testutils.GetTestServerConfig()
	cfg.Mode, cfg.DataDir = consts.Development, t.TempDir()
	controlDB, err := db.NewControlDB(t.Context(), cfg)
	require.NoError(t, err)
	t.Cleanup(func() { controlDB.Close() })
	site, err := controlDB.CreateSite(t.Context())
	require.NoError(t, err)
	siteID := site.SiteDB.Subdomain
	siteContext := &httputils.SiteContext{SiteID: siteID, ControlDB: controlDB, SiteDB: site.SiteDB, Config: cfg}

	tests := []struct {
		name  string
		newTx func(context.Context, *httputils.SiteContext) *siteapiutils.CollectionTx
		waits bool
	}{
		{"a read", siteapiutils.NewReadCollectionTx, false},
		{"a write", siteapiutils.NewCollectionTx, true},
	}
	for i, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// The handler's transactions on both DBs, each with a query run and not yet finished
			tx := test.newTx(t.Context(), siteContext)
			defer tx.Rollback()
			q, err := tx.Site()
			require.NoError(t, err)
			_, err = q.SearchGlucoseReadings(t.Context(), sitedb.SearchGlucoseReadingsParams{Limit: 1})
			require.NoError(t, err)
			controlQ, err := tx.Control()
			require.NoError(t, err)
			_, err = controlQ.AllSiteIdentities(t.Context(), controldb.AllSiteIdentitiesParams{SiteID: siteID})
			require.NoError(t, err)

			// An upload and a control write beside them, each given half a second
			ctx, cancel := context.WithTimeout(t.Context(), 500*time.Millisecond)
			defer cancel()
			now := time.Now().UnixMilli()
			_, siteErr := site.SiteDB.Queries().CreateGlucoseReading(ctx, sitedb.CreateGlucoseReadingParams{ID: test.name, Mgdl: 120,
				ReadingType: "sgv", OccurredAt: now + int64(i), CreatedByIdentityName: "phone"})
			controlErr := controlDB.Queries.UpdateSiteDateFlags(ctx, controldb.UpdateSiteDateFlagsParams{LastSeenAt: &now, ID: siteID})
			if test.waits {
				require.Error(t, siteErr, "the upload waits for the site's write lock")
				require.Error(t, controlErr, "the control write waits for the control DB's write lock")
				return
			}
			require.NoError(t, siteErr)
			require.NoError(t, controlErr)
		})
	}
}
