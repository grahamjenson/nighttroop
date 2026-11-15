package publicapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/time/rate"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/utils/testutils"
)

// TestPOSTCreateSite checks the public can create a site only while a control admin has public site creation on, and
// while the server has fewer sites than MAX_SITES
func TestPOSTCreateSite(t *testing.T) {
	cfg, controlDB := testutils.NewControlDB(t)
	maxSites, limiter := cfg.MaxSites, cfg.RateLimiters.SiteCreateLimiter
	t.Cleanup(func() { cfg.MaxSites, cfg.RateLimiters.SiteCreateLimiter = maxSites, limiter })
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

	// The JSON reply gives the site's URL and its Nightscout URL, so a script does not build them
	created := create()
	require.Equal(t, http.StatusOK, created.Code)
	var site map[string]any
	require.NoError(t, json.Unmarshal(created.Body.Bytes(), &site))
	subdomain := site["site_subdomain"].(string)
	require.Equal(t, cfg.SiteURL(subdomain), site["url"])
	require.Equal(t, cfg.LinkURL(subdomain+".ns"), site["pwd_ns_url"])
	sites := controlDB.GetNSitesDBs()

	setOn(false)
	off := create()
	require.Equal(t, http.StatusForbidden, off.Code)
	require.JSONEq(t, `{"error": "Creating sites is turned off"}`, off.Body.String())
	require.Equal(t, sites, controlDB.GetNSitesDBs())

	// MAX_SITES is a hard limit: a server with that many sites creates no more
	setOn(true)
	cfg.MaxSites = sites
	full := create()
	require.Equal(t, http.StatusTooManyRequests, full.Code)
	require.JSONEq(t, `{"error": "Full: Too many sites"}`, full.Body.String())
	require.Equal(t, sites, controlDB.GetNSitesDBs())
	cfg.MaxSites = sites + 1
	require.Equal(t, http.StatusOK, create().Code)
	require.Equal(t, sites+1, controlDB.GetNSitesDBs())

	// Over the rate limit, Retry-After says the seconds to wait, as the page does
	cfg.MaxSites = sites + 3
	cfg.RateLimiters.SiteCreateLimiter = rate.NewLimiter(rate.Every(time.Minute), 1)
	require.Equal(t, http.StatusOK, create().Code)
	limited := create()
	require.Equal(t, http.StatusTooManyRequests, limited.Code)
	require.JSONEq(t, `{"error": "Too many requests"}`, limited.Body.String())
	require.Contains(t, []string{"60", "61"}, limited.Header().Get("Retry-After"))
}
