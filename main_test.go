package main

import (
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

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

// A server whose port is in use gives the error at once, before the slow steps of a start, and one shut down gives
// none
func TestServeGivesTheErrorThatStopsIt(t *testing.T) {
	ok := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	taken, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { taken.Close() })

	_, err = serve(newHTTPServer(taken.Addr().String(), ok))
	require.ErrorContains(t, err, "address already in use")

	httpServer := newHTTPServer("127.0.0.1:0", ok)
	served, err := serve(httpServer)
	require.NoError(t, err)
	require.NoError(t, httpServer.Shutdown(t.Context()))
	select {
	case err := <-served:
		t.Fatalf("a server shut down gave an error: %v", err)
	case <-time.After(100 * time.Millisecond):
	}
}

// Until the server is ready each request, its health check too, is answered 503 "starting", to try again, and then by
// the server's handler
func TestStartingUntilReady(t *testing.T) {
	starting := &startingHandler{}
	get := func(path string) *httptest.ResponseRecorder {
		recorder := httptest.NewRecorder()
		starting.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, path, nil))
		return recorder
	}
	for _, path := range []string{"/health", "/api/v1/entries.json"} {
		response := get(path)
		require.Equal(t, http.StatusServiceUnavailable, response.Code, path)
		require.Equal(t, "5", response.Header().Get("Retry-After"), path)
		require.JSONEq(t, `{"status":"starting"}`, response.Body.String(), path)
	}

	starting.ready(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusTeapot) }))
	require.Equal(t, http.StatusTeapot, get("/health").Code)
}
