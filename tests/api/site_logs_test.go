package api_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/sql/logsdb"
	"nighttroop.t1d.nz/utils/consts"
)

// A site's admin reads a log of the site by its ID, and another site's admin does not. The server writes its logs to
// the logs DB once a second, so the log is written here: logs.http, which checks an event and an error the same way,
// would have to wait for one.
func TestAnotherSitesLogIsNotRead(t *testing.T) {
	own := newEventsSite(t)
	site, err := own.controlDB.CreateSite(t.Context())
	require.NoError(t, err)
	other := *own
	other.site = site

	own.controlDB.Logs.Log(t.Context(), &site.SiteDB.Subdomain, db.LogLevelInfo, "other synced", nil)
	logs, err := own.controlDB.Logs.Queries.SearchSiteLogs(t.Context(), logsdb.SearchSiteLogsParams{SiteID: site.SiteDB.Subdomain, Limit: 1})
	require.NoError(t, err)
	require.Len(t, logs, 1)

	path := fmt.Sprintf("/logs/logs/%d", logs[0].ID)
	require.Equal(t, http.StatusOK, other.request(consts.FIRST_ADMIN_NAME, consts.IdentityRoleAdmin, http.MethodGet, path, "").StatusCode)
	require.Equal(t, http.StatusNotFound, own.request(consts.FIRST_ADMIN_NAME, consts.IdentityRoleAdmin, http.MethodGet, path, "").StatusCode)
}
