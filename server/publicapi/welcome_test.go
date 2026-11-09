package publicapi

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/utils/httputils"
	"nighttroop.t1d.nz/utils/testutils"
)

// TestWelcome checks a site created from the home page opens its welcome page on its own address, with an ID after the
// # that shows its credentials once, and only on that site
func TestWelcome(t *testing.T) {
	cfg, controlDB := testutils.NewControlDB(t)
	w, create := httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/site", nil)
	create.Header.Set("Accept", "text/html")
	POSTCreateSiteHandler(controlDB, cfg)(w, create)
	require.Equal(t, http.StatusOK, w.Code)
	require.Empty(t, w.Body.String(), "the credentials are not shown on the home page")
	page, id, found := strings.Cut(w.Header().Get("HX-Redirect"), "#")
	require.True(t, found)
	siteURL := strings.TrimSuffix(page, "/welcome")
	siteID := strings.Split(strings.TrimPrefix(siteURL, "http://"), ".")[0]
	require.Equal(t, cfg.SiteURL(siteID)+"/welcome", page)

	show := func(site string) string {
		r := httptest.NewRequest(http.MethodPost, "/welcome", strings.NewReader(url.Values{"id": {id}}.Encode()))
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		r = r.WithContext(httputils.CreateSiteContext(r.Context(), cfg, controlDB, nil, nil, site, "", 0, 0, ""))
		w := httptest.NewRecorder()
		POSTWelcomeHandler(cfg)(w, r)
		require.Equal(t, http.StatusOK, w.Code)
		return w.Body.String()
	}
	gone := "These credentials have already been shown"

	// Another site does not show them, nor use them up
	require.Contains(t, show("othersit"), gone)
	shown := show(siteID)
	require.Contains(t, shown, "Backup key")
	require.Contains(t, shown, siteID+":")
	require.Contains(t, show(siteID), gone)
}
