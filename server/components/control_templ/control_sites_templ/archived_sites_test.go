package control_sites_templ

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/models/controlmodels"
	"nighttroop.t1d.nz/utils/httputils"
	"nighttroop.t1d.nz/utils/testutils"
)

// TestArchivedSitesTableHydrate checks each archived site has a Hydrate button changing it to live, which the browser
// tests cannot reach as they have no S3 to archive a site to
func TestArchivedSitesTableHydrate(t *testing.T) {
	var buf bytes.Buffer
	archivedAt := int64(1)
	sites := []*controlmodels.Site{{ID: "site", State: "archived", ArchivedAt: &archivedAt}}
	require.NoError(t, ArchivedSitesTable(sites, 0, 50).Render(context.Background(), &buf))
	for _, want := range []string{`id="control--archived-site--site--hydrate"`, `hx-post="/archived-sites/site"`, `name="state" value="live"`} {
		require.Contains(t, buf.String(), want)
	}
}

// TestSiteHydratedUnavailable checks a site restored with a DB that does not open is said to be unavailable, with its
// fixes on its page, which the browser tests cannot reach as their site DBs are in memory and always open
func TestSiteHydratedUnavailable(t *testing.T) {
	_, controlDB, site := testutils.NewControlDBWithSite(t)
	ctx := httputils.CreateControlContext(t.Context(), controlDB, "", httputils.HTTPActionWrite)
	siteModel := &controlmodels.Site{ID: site.SiteDB.Subdomain}
	var buf bytes.Buffer
	require.NoError(t, SiteHydrated(siteModel).Render(ctx, &buf))
	require.Contains(t, buf.String(), "is live again")

	controlDB.MakeSiteUnavailable(siteModel.ID, db.UnavailableCannotOpen)
	buf.Reset()
	require.NoError(t, SiteHydrated(siteModel).Render(ctx, &buf))
	require.Contains(t, buf.String(), "Site Restored, but Unavailable")
	require.Contains(t, buf.String(), "is live, but its DB does not open or migrate")
}
