package sitelogin

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/httputils"
)

// testConfig allows one login attempt an IP and name; a config can only be made once per process
var testConfig *config.ServerConfig

func TestMain(m *testing.M) {
	os.Setenv("SITE_LOGIN_PER_MINUTE_LIMIT", "1")
	os.Setenv("SITE_LOGIN_PER_MINUTE_BURST", "1")
	cfg, err := config.NewServerConfig()
	if err != nil {
		panic(err)
	}
	testConfig = cfg
	os.Exit(m.Run())
}

// A name longer than any identity's is refused before it reaches the rate limiter, which keys on it
func TestLoginRejectsLongNameBeforeRateLimit(t *testing.T) {
	longName := strings.Repeat("a", 33)
	tests := []struct {
		name        string
		ip          string
		contentType string
		accept      string
		body        string
	}{
		{"json", "203.0.113.1", "application/json", "application/json", `{"name":"` + longName + `","password":"pw"}`},
		{"form", "203.0.113.2", "application/x-www-form-urlencoded", "text/htmx", "name=" + longName + "&password=pw"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest("POST", "/login", strings.NewReader(tt.body))
			r.RemoteAddr = tt.ip + ":1234"
			r.Header.Set("Content-Type", tt.contentType)
			r.Header.Set("Accept", tt.accept)
			ctx, _ := httputils.SetIPInHTTPContext(context.Background(), r, nil)
			ctx = httputils.CreateSiteContext(ctx, testConfig, nil, nil, nil, "site"+tt.name, tt.ip, httputils.HTTPActionWrite, httputils.AcceptTypeJSON, httputils.ContentTypeJSON)
			r = r.WithContext(ctx)

			w := httptest.NewRecorder()
			POSTLogin(nil, testConfig)(w, r)
			require.Equal(t, http.StatusUnauthorized, w.Code)

			// The IP still has its one attempt, so the refused login never reached the limiter
			allowed, _ := httputils.SiteLoginAttemptsWithinRateLimit(testConfig, r, "site"+tt.name, "phone")
			require.True(t, allowed)
		})
	}
}
