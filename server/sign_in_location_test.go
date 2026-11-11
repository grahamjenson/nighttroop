package server

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/utils/testutils"
)

// A sign in, to a site or the control site, is recorded with the country its IP address is from, read from
// X-Forwarded-For as a reverse proxy on the same machine sends it, and a refused one too
func TestSignInsRecordWhereFrom(t *testing.T) {
	config, controlDB, site := testutils.NewControlDBWithSite(t)
	handler := GetMainHandler(controlDB, config)

	send := func(method, url, token, body string) map[string]any {
		r := httptest.NewRequest(method, url, strings.NewReader(body))
		r.RemoteAddr = "127.0.0.1:1234"
		r.Header.Set("X-Forwarded-For", "202.36.163.1") // in New Zealand
		r.Header.Set("Accept", "application/json")
		r.Header.Set("Content-Type", "application/json")
		if token != "" {
			r.Header.Set("Authorization", "Bearer "+token)
		}
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		var out map[string]any
		_ = json.Unmarshal(w.Body.Bytes(), &out)
		return out
	}
	newest := func(out map[string]any) map[string]any {
		// A site's search is a page of them in items; control's is a list
		result := out["result"]
		if page, ok := result.(map[string]any); ok {
			result = page["items"]
		}
		events := result.([]any)
		require.NotEmpty(t, events)
		return events[0].(map[string]any)["attributes"].(map[string]any)
	}

	siteURL := "http://" + site.SiteDB.Subdomain + ".localhost"
	send("POST", siteURL+"/login", "", `{"name": "admin", "password": "wrong", "totp": "123123"}`)
	login := send("POST", siteURL+"/login", "", `{"name": "admin", "password": "password", "totp": "123123"}`)
	token, _ := login["token"].(string)
	require.NotEmpty(t, token, login)

	signedIn := newest(send("GET", siteURL+"/logs/events?type=signed_in", token, ""))
	require.Equal(t, "admin", signedIn["identity_name"])
	require.Equal(t, map[string]any{"country": "NZ"}, signedIn["from"])
	failed := newest(send("GET", siteURL+"/logs/events?type=sign_in_failed", token, ""))
	require.Equal(t, map[string]any{"country": "NZ"}, failed["from"])

	control := send("POST", "http://control.localhost/login", "", `{"name": "admin", "password": "password", "totp": "123123"}`)
	controlToken, _ := control["token"].(string)
	require.NotEmpty(t, controlToken, control)
	controlSignedIn := newest(send("GET", "http://control.localhost/events?type=signed_in", controlToken, ""))
	require.Equal(t, "admin", controlSignedIn["identity_name"])
	require.Equal(t, map[string]any{"country": "NZ"}, controlSignedIn["from"])
}

// A page showing DB-IP's locations credits DB-IP in its footer, and a part of one loaded into a page sends the credit
// to that page's footer; pages without locations do not
func TestEventsPagesCreditDBIP(t *testing.T) {
	config, controlDB, site := testutils.NewControlDBWithSite(t)
	handler := GetMainHandler(controlDB, config)
	siteURL := "http://" + site.SiteDB.Subdomain + ".localhost"

	r := httptest.NewRequest("POST", siteURL+"/login", strings.NewReader(`{"name": "admin", "password": "password", "totp": "123123"}`))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	var login struct{ Token string }
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &login))

	page := func(path string, htmx bool) string {
		r := httptest.NewRequest("GET", siteURL+path, nil)
		r.Header.Set("Accept", "text/html")
		r.Header.Set("Authorization", "Bearer "+login.Token)
		if htmx {
			r.Header.Set("HX-Request", "true")
		}
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		return w.Body.String()
	}
	require.Contains(t, page("/logs/events", false), `id="footer--dbip-link"`)
	require.NotContains(t, page("/logs/events", false), `hx-swap-oob="innerHTML:#footer--dbip"`)
	require.Contains(t, page("/logs/events", true), `hx-swap-oob="innerHTML:#footer--dbip"`)
	require.NotContains(t, page("/logs/errors", false), `id="footer--dbip-link"`)
	require.NotContains(t, page("/logs/errors", true), `footer--dbip`)
}
