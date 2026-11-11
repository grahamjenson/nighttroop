package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/getsentry/sentry-go"
	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/utils/config"
)

// Sentry reports a server's errors under its deployment's name, so each staging server's are apart
func TestSentryEnvironmentIsTheDeployment(t *testing.T) {
	t.Cleanup(func() { sentry.CurrentHub().BindClient(nil) })
	cfg := &config.ServerConfig{SentryDSN: "https://public@sentry.example.invalid/1", Deployment: "nighttroop-staging"}
	setupSentry(cfg)
	require.True(t, cfg.Status.SentryWorking)
	require.Equal(t, "nighttroop-staging", sentry.CurrentHub().Client().Options().Environment)
}

// A request's headers are refused past maxHeaderBytes, so no connection holds Go's default 1 MB of them
func TestRequestHeadersAreLimited(t *testing.T) {
	ok := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	ts := httptest.NewUnstartedServer(nil)
	ts.Config = newHTTPServer("", ok)
	ts.Start()
	t.Cleanup(ts.Close)

	for _, test := range []struct {
		name   string
		bytes  int
		status int
	}{
		{"a large cookie", 16 << 10, http.StatusOK},
		{"more than the limit", maxHeaderBytes + 8<<10, http.StatusRequestHeaderFieldsTooLarge},
	} {
		t.Run(test.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, ts.URL, nil)
			require.NoError(t, err)
			req.Header.Set("Cookie", "c="+strings.Repeat("a", test.bytes))
			res, err := ts.Client().Do(req)
			require.NoError(t, err)
			defer res.Body.Close()
			require.Equal(t, test.status, res.StatusCode)
		})
	}
}
