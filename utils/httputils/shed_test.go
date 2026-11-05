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
