package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/utils/testutils"
)

// Until a new server's control admin is made, only its setup page and health check are served
func TestSetupGatesEveryHost(t *testing.T) {
	config := testutils.GetTestServerConfig()
	controlDB, err := db.NewControlDB(context.Background(), config)
	require.NoError(t, err)
	handler := GetMainHandler(controlDB, config)

	send := func(method, url, accept, body string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, url, strings.NewReader(body))
		r.Header.Set("Accept", accept)
		if body != "" {
			r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		}
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		return w
	}

	tests := []struct {
		name, method, url, accept, body string
		status                          int
		location                        string
	}{
		{name: "public page", method: "GET", url: "http://localhost/", accept: "text/html", status: http.StatusSeeOther, location: config.SetupURL()},
		{name: "site page", method: "GET", url: "http://abcdefgh.localhost/", accept: "text/html", status: http.StatusSeeOther, location: config.SetupURL()},
		{name: "site API", method: "GET", url: "http://abcdefgh.localhost/data/glucose", accept: "application/json", status: http.StatusServiceUnavailable},
		{name: "Nightscout API", method: "GET", url: "http://abcdefgh.ns.localhost/api/v1/entries", accept: "application/json", status: http.StatusServiceUnavailable},
		{name: "creating a site", method: "POST", url: "http://localhost/site", accept: "application/json", status: http.StatusServiceUnavailable},
		{name: "health check", method: "GET", url: "http://localhost/health", accept: "application/json", status: http.StatusOK},
		{name: "control login", method: "GET", url: "http://control.localhost/login", accept: "text/html", status: http.StatusSeeOther, location: "/setup"},
		{name: "setup page", method: "GET", url: "http://control.localhost/setup", accept: "text/html", status: http.StatusOK},
		{name: "wrong setup code", method: "POST", url: "http://control.localhost/setup", accept: "text/html", body: "code=wrong", status: http.StatusUnauthorized},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := send(tt.method, tt.url, tt.accept, tt.body)
			require.Equal(t, tt.status, w.Code, w.Body.String())
			require.Equal(t, tt.location, w.Header().Get("Location"))
		})
	}

	// The setup code shows the login, a code from its TOTP secret saves it, and the code no longer works
	code := controlDB.SetupCode()
	w := send("POST", "http://control.localhost/setup", "text/html", "code="+code)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	require.Equal(t, "no-store", w.Header().Get("Cache-Control"))
	w = send("POST", "http://control.localhost/setup/confirm", "text/html", "code="+code+"&totp=111111")
	require.Equal(t, http.StatusUnauthorized, w.Code)
	require.False(t, controlDB.SetupDone())
	w = send("POST", "http://control.localhost/setup/confirm", "text/html", "code="+code+"&totp=123123")
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	require.True(t, controlDB.SetupDone())

	// Set up, the server is served as usual and the setup page is gone
	require.Equal(t, http.StatusOK, send("GET", "http://localhost/", "text/html", "").Code)
	require.Equal(t, http.StatusNotFound, send("GET", "http://control.localhost/setup", "text/html", "").Code)
	_, err = controlDB.BeginSetup(code)
	require.ErrorIs(t, err, db.ErrBadSetup)
}
