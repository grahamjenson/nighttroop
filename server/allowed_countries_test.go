package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/utils/testutils"
)

// With ALLOWED_COUNTRIES set, creating a site, signing in and recovering are refused from other countries, but not
// from an address whose country is not known, and the Nightscout API phones upload with is not limited
func TestAllowedCountries(t *testing.T) {
	testConfig, controlDB, site := testutils.NewControlDBWithSite(t)
	config := *testConfig
	config.AllowedCountries = []string{"AU"}
	handler := GetMainHandler(controlDB, &config)

	const (
		australia  = "139.130.4.5"
		newZealand = "202.36.163.1"
		private    = "10.0.0.1"
	)
	require.Equal(t, "AU", config.GeoIP.Locate(australia).Country)
	require.Equal(t, "NZ", config.GeoIP.Locate(newZealand).Country)

	siteURL := "http://" + site.SiteDB.Subdomain + ".localhost"
	send := func(method, url, from string) int {
		r := httptest.NewRequest(method, url, strings.NewReader(`{"name": "admin", "password": "wrong"}`))
		r.RemoteAddr = "127.0.0.1:1234"
		r.Header.Set("X-Forwarded-For", from)
		r.Header.Set("Accept", "application/json")
		r.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		return w.Code
	}

	limited := []struct{ method, url string }{
		{"POST", "http://localhost/site"},
		{"POST", "http://localhost/recover"},
		{"POST", siteURL + "/login"},
		{"POST", siteURL + "/login/totp"},
		{"POST", siteURL + "/recover"},
		{"POST", "http://control.localhost/login"},
	}
	for _, route := range limited {
		require.Equal(t, http.StatusForbidden, send(route.method, route.url, newZealand), route.url)
		require.NotEqual(t, http.StatusForbidden, send(route.method, route.url, australia), route.url)
		require.NotEqual(t, http.StatusForbidden, send(route.method, route.url, private), route.url)
	}

	// Phones keep uploading from anywhere, e.g. while travelling
	require.NotEqual(t, http.StatusForbidden, send("GET", "http://"+site.SiteDB.Subdomain+".ns.localhost/api/v1/status.json?token="+site.PwdNSToken, newZealand))
}
