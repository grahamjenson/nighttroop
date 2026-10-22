package controlapi

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
func TestControlLoginRejectsLongNameBeforeRateLimit(t *testing.T) {
	longName := strings.Repeat("a", 33)
	tests := []struct {
		name        string
		ip          string
		contentType string
		body        string
	}{
		{"json", "203.0.113.1", "application/json", `{"name":"` + longName + `","password":"pw"}`},
		{"form", "203.0.113.2", "application/x-www-form-urlencoded", "name=" + longName + "&password=pw"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest("POST", "/login", strings.NewReader(tt.body))
			r.RemoteAddr = tt.ip + ":1234"
			r.Header.Set("Content-Type", tt.contentType)
			ctx, _ := httputils.SetIPInHTTPContext(context.Background(), r)
			r = r.WithContext(ctx)

			w := httptest.NewRecorder()
			POSTLogin(nil, testConfig)(w, r)
			require.Equal(t, http.StatusUnauthorized, w.Code)

			// The IP still has its one attempt, so the refused login never reached the limiter
			allowed, _ := httputils.ControlLoginAttemptsWithinRateLimit(testConfig, r, "admin"+tt.name)
			require.True(t, allowed)
		})
	}
}
