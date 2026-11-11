package server

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/utils/testutils"
)

// Every response, a site's page, its Nightscout API, Socket.IO and the public and control hosts alike, says not to
// sniff its type, not to be framed and not to send its URL to other origins, and behind HTTPS to only be reached over
// HTTPS. The API docs page keeps its own policy on images as well.
func TestSecurityHeaders(t *testing.T) {
	testConfig, controlDB, site := testutils.NewControlDBWithSite(t)
	siteHost := site.SiteDB.Subdomain + ".localhost"
	paths := []struct {
		name string
		url  string
		csp  string
	}{
		{name: "site page", url: "http://" + siteHost + "/login", csp: "frame-ancestors 'none'"},
		{name: "site missing page", url: "http://" + siteHost + "/no-such-page", csp: "frame-ancestors 'none'"},
		{name: "nightscout API", url: "http://" + site.SiteDB.Subdomain + ".ns.localhost/api/v1/status.json?token=" + site.PwdNSToken, csp: "frame-ancestors 'none'"},
		{name: "socket.io", url: "http://" + siteHost + "/socket.io/?EIO=4&transport=polling", csp: "frame-ancestors 'none'"},
		{name: "public", url: "http://localhost/", csp: "frame-ancestors 'none'"},
		{name: "control", url: "http://control.localhost/", csp: "frame-ancestors 'none'"},
		{name: "api docs", url: "http://localhost/api-docs", csp: "frame-ancestors 'none'; img-src 'self' data:"},
	}
	for _, protocol := range []string{"http", "https"} {
		config := *testConfig
		config.HostProtocol = protocol
		handler := GetMainHandler(controlDB, &config)
		for _, path := range paths {
			t.Run(protocol+" "+path.name, func(t *testing.T) {
				r := httptest.NewRequest("GET", path.url, nil)
				r.RemoteAddr = "127.0.0.1:1234"
				w := httptest.NewRecorder()
				handler.ServeHTTP(w, r)
				header := w.Result().Header
				require.Equal(t, "nosniff", header.Get("X-Content-Type-Options"))
				require.Equal(t, "DENY", header.Get("X-Frame-Options"))
				require.Equal(t, []string{path.csp}, header.Values("Content-Security-Policy"))
				require.Equal(t, "same-origin", header.Get("Referrer-Policy"))
				hsts := ""
				if protocol == "https" {
					hsts = "max-age=31536000; includeSubDomains"
				}
				require.Equal(t, hsts, header.Get("Strict-Transport-Security"))
			})
		}
	}
}
