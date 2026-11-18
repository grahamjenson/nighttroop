package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/data/sql/logsdb"
	"nighttroop.t1d.nz/utils"
	"nighttroop.t1d.nz/utils/testutils"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/page"
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

// A server whose S3 is not working starts without it when its control DB is here, and is told to restart once S3 works
// and it holds the lock. Without its control DB it does not start, as it would start new. A working S3 needs neither.
func TestStartS3(t *testing.T) {
	here := filepath.Join(t.TempDir(), "controldb.sqlite")
	require.NoError(t, os.WriteFile(here, nil, 0o600))
	down := func(context.Context) error { return errors.New("S3 is not working") }
	works := make(chan struct{})
	whenWorking := func(context.Context) bool { <-works; return true }
	pages := []string{}
	admin := func(_ context.Context, level page.Level, message string, _ ...any) {
		pages = append(pages, string(level)+": "+message)
	}

	restart, err := startS3(t.Context(), func(context.Context) error { return nil }, here, whenWorking, admin)
	require.NoError(t, err)
	require.Nil(t, restart, "a working S3 never restarts the server")

	_, err = startS3(t.Context(), down, filepath.Join(t.TempDir(), "missing.sqlite"), whenWorking, admin)
	require.ErrorContains(t, err, "S3 is not working", "without its control DB the server does not start")

	require.Empty(t, pages, "a server that works, or does not start, pages no one")
	restart, err = startS3(t.Context(), down, here, whenWorking, admin)
	require.NoError(t, err, "with its control DB the server starts without S3")
	require.Len(t, pages, 1, "the admin is paged once")
	require.Contains(t, pages[0], string(page.SMS)+": S3 is not working")
	select {
	case <-restart:
		t.Fatal("the server restarts while S3 is not working")
	case <-time.After(50 * time.Millisecond):
	}
	close(works)
	select {
	case <-restart:
	case <-time.After(time.Second):
		t.Fatal("the server does not restart once S3 works")
	}

	// One told to stop while S3 is not working stops, and does not start without S3
	stopped, cancel := context.WithCancel(t.Context())
	cancel()
	_, err = startS3(stopped, down, here, whenWorking, admin)
	require.Error(t, err)
}

// The admin is paged with a call when the server does not start, or stops with an error, each with one message so a
// server that fails each time it is started again is one incident, and not when it is told to stop or restarts with S3
func TestPageFailure(t *testing.T) {
	tests := []struct {
		name    string
		started bool
		err     error
		want    string
	}{
		{"it does not start", false, errors.New("cannot create control db"), "The server did not start, so its sites are down"},
		{"it stops with an error", true, errors.New("cannot serve HTTP"), "The server stopped with an error, so its sites are down"},
		{"it stops as told", true, nil, ""},
		{"it is told to stop as it starts", false, fmt.Errorf("S3: %w", context.Canceled), ""},
		{"it restarts with S3", true, errRestartWithS3, ""},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			paged := ""
			pageFailure(func(_ context.Context, level page.Level, message string, args ...any) {
				require.Equal(t, page.Call, level)
				require.Equal(t, []any{"error", test.err}, args)
				paged = message
			}, test.started, test.err)
			require.Equal(t, test.want, paged)
		})
	}
}

// A start on a nearly full disk makes room from what the server makes again: its local backup copies, then its logs
// DB if still short. It pages the admin, leaves the DBs, and does not start if there is still too little room.
func TestMakeRoom(t *testing.T) {
	mb := float64(1 << 20)
	tests := []struct {
		name            string
		free            []float64 // at each look: before, after the backups, at the end
		backups, logsDB bool      // whether each is still there after
		pages           int
		wantErr         string
	}{
		{"a disk with room is left as it is", []float64{500 * mb}, true, true, 0, ""},
		{"the backup copies are enough", []float64{20 * mb, 300 * mb, 300 * mb}, false, true, 1, ""},
		{"the logs DB goes too", []float64{20 * mb, 60 * mb, 400 * mb}, false, false, 1, ""},
		{"still too little room", []float64{20 * mb, 30 * mb, 40 * mb}, false, false, 1, "the DBs' disk has 40 MB free, and the server needs 100 MB to start"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cfg := *testutils.GetTestServerConfig()
			cfg.DataDir = t.TempDir()
			backup := filepath.Join(cfg.BackupDir(), "site", "a", "abcdefgh.sqlite.lz4")
			control, logs := controldb.GetPath(&cfg), logsdb.GetPath(&cfg)
			for _, file := range []string{backup, control, logs} {
				require.NoError(t, os.MkdirAll(filepath.Dir(file), 0o700))
				require.NoError(t, os.WriteFile(file, []byte("data"), 0o600))
			}
			looks, pages := 0, 0
			err := makeRoom(&cfg, func(_ context.Context, level page.Level, _ string, _ ...any) {
				require.Equal(t, page.SMS, level)
				pages++
			}, func() (float64, bool) {
				looks++
				return test.free[min(looks, len(test.free))-1], true
			})
			if test.wantErr == "" {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, test.wantErr)
			}
			require.Equal(t, test.pages, pages)
			require.Equal(t, test.backups, utils.FileExists(backup))
			require.Equal(t, test.logsDB, utils.FileExists(logs))
			require.FileExists(t, control, "a DB with data is never removed")
		})
	}
}
