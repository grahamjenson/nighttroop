package publicapi

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/utils/testutils"
)

// TestPOSTCreateSite checks the public can create a site only while a control admin has public site creation on, and
// while the server has fewer sites than MAX_SITES
func TestPOSTCreateSite(t *testing.T) {
	cfg, controlDB := testutils.NewControlDB(t)
	maxSites := cfg.MaxSites
	t.Cleanup(func() { cfg.MaxSites = maxSites })
	create := func() *httptest.ResponseRecorder {
		r := httptest.NewRequest(http.MethodPost, "/site", nil)
		r.Header.Set("Accept", "application/json")
		w := httptest.NewRecorder()
		POSTCreateSiteHandler(controlDB, cfg)(w, r)
		return w
	}
	setOn := func(on bool) {
		require.NoError(t, controlDB.Queries.UpdatePublicSiteCreation(t.Context(), controldb.UpdatePublicSiteCreationParams{PublicSiteCreation: on}))
	}

	require.Equal(t, http.StatusOK, create().Code)
	sites := controlDB.GetNSitesDBs()

	setOn(false)
	off := create()
	require.Equal(t, http.StatusForbidden, off.Code)
	require.JSONEq(t, `{"error": "Creating sites is turned off"}`, off.Body.String())
	require.Equal(t, sites, controlDB.GetNSitesDBs())

	// MAX_SITES is a hard limit: a server with that many sites creates no more
	setOn(true)
	cfg.MaxSites = sites
	require.Equal(t, http.StatusTooManyRequests, create().Code)
	require.Equal(t, sites, controlDB.GetNSitesDBs())
	cfg.MaxSites = sites + 1
	require.Equal(t, http.StatusOK, create().Code)
	require.Equal(t, sites+1, controlDB.GetNSitesDBs())
}
