package publicapi

import (
	"encoding/json"
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

// TestRecoverWelcome checks a recovery from the recover page shows the new credentials on the site's own address, as a
// new site's are, once and only on that site, and never on the recover page
func TestRecoverWelcome(t *testing.T) {
	cfg, controlDB := testutils.NewControlDB(t)
	w, create := httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/site", nil)
	create.Header.Set("Accept", "application/json")
	POSTCreateSiteHandler(controlDB, cfg)(w, create)
	require.Equal(t, http.StatusOK, w.Code)
	var site CreateUserResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &site))

	form := url.Values{"site_id": {site.SiteSubdomain}, "backup_key": {site.BackupKey}}.Encode()
	recover := httptest.NewRequest(http.MethodPost, "/recover", strings.NewReader(form))
	recover.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w = httptest.NewRecorder()
	POSTRecoverSiteHandler(controlDB, cfg)(w, recover)
	require.Equal(t, http.StatusOK, w.Code)
	require.Empty(t, w.Body.String(), "the credentials are not shown on the recover page")
	page, id, found := strings.Cut(w.Header().Get("HX-Redirect"), "#")
	require.True(t, found)
	require.Equal(t, cfg.SiteURL(site.SiteSubdomain)+"/welcome", page)

	show := func(siteID string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(http.MethodPost, "/welcome", strings.NewReader(url.Values{"id": {id}}.Encode()))
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		r = r.WithContext(httputils.CreateSiteContext(r.Context(), cfg, controlDB, nil, nil, siteID, "", 0, 0, ""))
		w := httptest.NewRecorder()
		POSTWelcomeHandler(cfg)(w, r)
		require.Equal(t, http.StatusOK, w.Code)
		return w
	}
	gone := "These credentials have already been shown"

	// Another site does not show them, nor use them up; the site shows its new backup key, not the one used, once,
	// and lets the browser shown them add the admin a passkey in place of the ones the recovery removed
	require.Contains(t, show("othersit").Body.String(), gone)
	shown := show(site.SiteSubdomain)
	require.Equal(t, "no-store", shown.Header().Get("Cache-Control"))
	require.Contains(t, shown.Body.String(), site.SiteSubdomain+":")
	require.NotContains(t, shown.Body.String(), site.BackupKey)
	require.Contains(t, shown.Body.String(), "public--passkey-login-button")
	require.NotEmpty(t, shown.Result().Cookies())
	require.Contains(t, show(site.SiteSubdomain).Body.String(), gone)
}
