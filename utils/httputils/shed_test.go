package httputils_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/utils/httputils"
)

// A busy server works on as many requests as it can, has a few more wait, and refuses the rest at once, though never
// its health check or a live update
func TestShedLoad(t *testing.T) {
	release := make(chan struct{})
	started := make(chan struct{}, 10)
	handler := httputils.ShedLoad(1, 1, 300*time.Millisecond, httputils.LongLived)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/slow" {
			started <- struct{}{}
			<-release
		}
		w.WriteHeader(http.StatusOK)
	}))
	serve := func(path string, accept string) chan *httptest.ResponseRecorder {
		done := make(chan *httptest.ResponseRecorder, 1)
		go func() {
			r := httptest.NewRequest("GET", path, nil)
			r.Header.Set("Accept", accept)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, r)
			done <- w
		}()
		return done
	}

	// One is worked on, and one waits for it
	first := serve("/slow", "")
	<-started
	waiting := serve("/fast", "")
	time.Sleep(50 * time.Millisecond)

	// Past those it is refused at once, saying when to try again
	refused := <-serve("/fast", "")
	require.Equal(t, http.StatusServiceUnavailable, refused.Code)
	require.Equal(t, "5", refused.Header().Get("Retry-After"))

	// The health check and a live update are never held
	require.Equal(t, http.StatusOK, (<-serve("/health", "")).Code)
	require.Equal(t, http.StatusOK, (<-serve("/data/glucose/events", "text/event-stream")).Code)

	// Once the first finishes, the one waiting is worked on
	close(release)
	require.Equal(t, http.StatusOK, (<-first).Code)
	require.Equal(t, http.StatusOK, (<-waiting).Code)

	// One that waits longer than it may is refused
	release = make(chan struct{})
	slow := serve("/slow", "")
	<-started
	start := time.Now()
	require.Equal(t, http.StatusServiceUnavailable, (<-serve("/fast", "")).Code)
	require.GreaterOrEqual(t, time.Since(start), 300*time.Millisecond)
	close(release)
	require.Equal(t, http.StatusOK, (<-slow).Code)
}

// A request is long lived by its route alone: one asking for an event stream from any other route is not, so it cannot
// skip the load shedding and timeout
func TestLongLived(t *testing.T) {
	for _, test := range []struct {
		method, url, accept string
		want                bool
	}{
		{"GET", "http://localhost/health", "", true},
		{"GET", "http://site.ns.localhost/socket.io/?EIO=4&transport=polling", "", true},
		{"POST", "http://site.ns.localhost/socket.io/?EIO=4&transport=polling", "", true},
		{"GET", "http://control.localhost" + httputils.ControlLogsDBPath, "", true},
		{"GET", "http://site.localhost/data/summary/events", "text/event-stream", true},
		{"GET", "http://site.localhost/data/alarms/events", "text/html, text/event-stream", true},
		{"GET", "http://site.localhost/settings/connectors/events", "text/event-stream", true},
		{"GET", "http://site.localhost/presets/profiles/events", "", true},
		{"GET", "http://site.localhost/logs/events/events", "", true},
		{"GET", "http://site.localhost" + httputils.ControlLogsDBPath, "", false},
		{"GET", "http://site.localhost/data/summary", "text/event-stream", false},
		{"POST", "http://site.localhost/login", "text/event-stream", false},
		{"POST", "http://control.localhost/login", "text/event-stream", false},
		{"GET", "http://site.localhost/reports/agp", "text/event-stream", false},
		{"GET", "http://site.ns.localhost/api/v1/entries", "text/event-stream", false},
		{"POST", "http://site.localhost/data/glucose/events", "text/event-stream", false},
		{"GET", "http://site.localhost/data//events", "text/event-stream", false},
		{"GET", "http://site.localhost/data/glucose/1/events", "text/event-stream", false},
		{"GET", "http://site.localhost/reports/glucose/events", "text/event-stream", false},
		{"GET", "http://site.localhost/events", "text/event-stream", false},
	} {
		r := httptest.NewRequest(test.method, test.url, nil)
		r.Header.Set("Accept", test.accept)
		require.Equal(t, test.want, httputils.LongLived(r), test.method+" "+test.url+" "+test.accept)
	}
}

// Uploads keep a share of a busy server to themselves, so they are worked on while page loads fill the rest
func TestShedLoadReserving(t *testing.T) {
	release := make(chan struct{})
	started := make(chan struct{}, 10)
	upload := func(r *http.Request) bool { return r.Method == http.MethodPost }
	// Of 4 in flight and none waiting, 1 is kept for uploads, with 1 to wait for it
	handler := httputils.ShedLoadReserving(4, 0, 300*time.Millisecond, httputils.LongLived, upload)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/slow" {
			started <- struct{}{}
			<-release
		}
		w.WriteHeader(http.StatusOK)
	}))
	serve := func(method string, path string) chan int {
		done := make(chan int, 1)
		go func() {
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, httptest.NewRequest(method, path, nil))
			done <- w.Code
		}()
		return done
	}

	// Page loads fill the 3 that are not kept, and past those are refused
	var pages []chan int
	for range 3 {
		pages = append(pages, serve(http.MethodGet, "/slow"))
		<-started
	}
	require.Equal(t, http.StatusServiceUnavailable, <-serve(http.MethodGet, "/fast"))

	// An upload is still worked on
	require.Equal(t, http.StatusOK, <-serve(http.MethodPost, "/fast"))

	// Uploads are held to their own share: one worked on, one waiting, then refused
	slowUpload := serve(http.MethodPost, "/slow")
	<-started
	waitingUpload := serve(http.MethodPost, "/fast")
	time.Sleep(50 * time.Millisecond)
	require.Equal(t, http.StatusServiceUnavailable, <-serve(http.MethodPost, "/fast"))

	close(release)
	for _, page := range append(pages, slowUpload, waitingUpload) {
		require.Equal(t, http.StatusOK, <-page)
	}
}
