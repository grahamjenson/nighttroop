package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/testutils"
)

// A read only control identity sees the server, its sites and its admins, but cannot change any of them, download a
// DB, profile the server or send a push: each route only an admin uses refuses it before doing anything
func TestControlReadOnlyIdentityCannotChangeTheServer(t *testing.T) {
	config, controlDB, site := testutils.NewControlDBWithSite(t)
	handler := GetMainHandler(controlDB, config)
	_, err := controlDB.CreateControlIdentity(t.Context(), "reader", consts.ControlRoleRead, nil)
	require.NoError(t, err)
	// A token issued in the millisecond its credentials were made is refused, as one issued before they changed is
	time.Sleep(2 * time.Millisecond)
	token, _, err := config.Security.GenerateControlJWTToken("reader", string(consts.ControlRoleRead))
	require.NoError(t, err)
	siteID := site.SiteDB.Subdomain

	tests := []struct {
		method, path string
		status       int
	}{
		// Only an admin changes the server and its sites
		{http.MethodPost, "/sites", http.StatusForbidden},
		{http.MethodPost, "/settings", http.StatusForbidden},
		{http.MethodPost, "/sites/" + siteID, http.StatusForbidden},
		{http.MethodPost, "/archived-sites/" + siteID, http.StatusForbidden},
		{http.MethodPost, "/sites/" + siteID + "/truncate", http.StatusForbidden},
		{http.MethodPost, "/sites/" + siteID + "/new-empty-db", http.StatusForbidden},
		{http.MethodPost, "/sites/" + siteID + "/restore-version", http.StatusForbidden},
		{http.MethodGet, "/sites/" + siteID + "/db", http.StatusForbidden},
		{http.MethodPost, "/sites/" + siteID + "/db", http.StatusForbidden},
		{http.MethodPost, "/sites/" + siteID + "/identities/1/password", http.StatusForbidden},
		{http.MethodPost, "/sites/" + siteID + "/identities/1/totp", http.StatusForbidden},
		{http.MethodPost, "/sites/" + siteID + "/identities/1/assume", http.StatusForbidden},
		{http.MethodPost, "/rate-limits", http.StatusForbidden},
		{http.MethodGet, "/logs-db", http.StatusForbidden},
		// and its push credentials and devices
		{http.MethodGet, "/push-credentials/new", http.StatusForbidden},
		{http.MethodPost, "/push-credentials", http.StatusForbidden},
		{http.MethodDelete, "/push-credentials/1", http.StatusForbidden},
		{http.MethodPost, "/push-devices/1/test", http.StatusForbidden},
		// and its admins
		{http.MethodPost, "/admins", http.StatusForbidden},
		{http.MethodPost, "/admins/1/role", http.StatusForbidden},
		{http.MethodPost, "/admins/1/password", http.StatusForbidden},
		{http.MethodPost, "/admins/1/totp", http.StatusForbidden},
		{http.MethodDelete, "/admins/1", http.StatusForbidden},
		{http.MethodPost, "/admins/test-page", http.StatusForbidden},
		// The profiles show the server's code and its memory
		{http.MethodGet, "/debug/pprof/heap", http.StatusForbidden},
		{http.MethodGet, "/debug/vars", http.StatusForbidden},
		// What it can see
		{http.MethodGet, "/push-credentials", http.StatusOK},
		{http.MethodGet, "/push-devices", http.StatusOK},
		{http.MethodGet, "/rate-limits", http.StatusOK},
		{http.MethodGet, "/admins", http.StatusOK},
	}
	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			r := httptest.NewRequest(tt.method, "http://control.localhost"+tt.path, nil)
			r.Header.Set("Authorization", "Bearer "+token)
			r.Header.Set("Accept", "application/json")
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, r)
			require.Equal(t, tt.status, w.Code, w.Body.String())
		})
	}
}
