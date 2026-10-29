package control_sites_templ

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/models/controlmodels"
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
