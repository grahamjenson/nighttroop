package httputils

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/testutils"
	"nighttroop.t1d.nz/utils/usage"
)

func TestRequestLogsRedactCredentials(t *testing.T) {
	var logs bytes.Buffer
	router := chi.NewRouter()
	testutils.UseLogHandler(t, slog.NewTextHandler(&logs, nil))
	AddStandardMiddleware(router, &config.ServerConfig{}, nil)
	router.Get("/api/v2/authorization/request/{token}", func(w http.ResponseWriter, r *http.Request) {})

	r := httptest.NewRequest(http.MethodGet, "/api/v2/authorization/request/pathtoken123?token=querytoken456&secret=querysecret789&count=10", nil)
	router.ServeHTTP(httptest.NewRecorder(), r)

	assert.Contains(t, logs.String(), "count=10")
	for _, credential := range []string{"pathtoken123", "querytoken456", "querysecret789"} {
		assert.NotContains(t, logs.String(), credential)
	}
}

// The Nightscout API takes a token in the query, the api-secret header and the path, none of which may reach Sentry
func TestSentryEventsRedactCredentials(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "https://abcdefgh.ns.example.com/api/v2/authorization/request/pathtoken123?token=querytoken456&secret=querysecret789&count=10", nil)
	r.Header.Set("Api-Secret", "headersecret000")
	r.Header.Set("Authorization", "Bearer bearertoken111")
	r.Header.Set("Accept", "application/json")

	event := RedactSentryEvent(&sentry.Event{Request: sentry.NewRequest(r)}, nil)
	sent, err := json.Marshal(event.Request)
	require.NoError(t, err)

	assert.Contains(t, string(sent), "count=10")
	assert.Contains(t, string(sent), "application/json")
	for _, credential := range []string{"pathtoken123", "querytoken456", "querysecret789", "headersecret000", "bearertoken111"} {
		assert.NotContains(t, string(sent), credential)
	}

	// An event without a request is sent as it is
	assert.Nil(t, RedactSentryEvent(&sentry.Event{}, nil).Request)
}

func TestResponsesAreGzippedWhenAccepted(t *testing.T) {
	router := chi.NewRouter()
	AddStandardMiddleware(router, &config.ServerConfig{}, nil)
	css := ".a { color: red; }"
	router.Get("/style_gen.css", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
		_, _ = w.Write([]byte(css))
	})

	r := httptest.NewRequest(http.MethodGet, "/style_gen.css", nil)
	r.Header.Set("Accept-Encoding", "gzip")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)
	require.Equal(t, "gzip", w.Header().Get("Content-Encoding"))
	reader, err := gzip.NewReader(w.Body)
	require.NoError(t, err)
	body, err := io.ReadAll(reader)
	require.NoError(t, err)
	assert.Equal(t, css, string(body))

	// A client that does not accept gzip gets it as it is
	r = httptest.NewRequest(http.MethodGet, "/style_gen.css", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, r)
	assert.Empty(t, w.Header().Get("Content-Encoding"))
	assert.Equal(t, css, w.Body.String())
}

// Only signed in requests are counted, by their route's pattern, not their path
func TestCountRoutes(t *testing.T) {
	tracker := usage.New()
	router := chi.NewRouter()
	router.Get("/htmx.js", func(w http.ResponseWriter, r *http.Request) {})
	router.Group(func(r chi.Router) {
		// A stand in for a router's sign in
		r.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("Authorization") == "" {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
				next.ServeHTTP(w, r)
			})
		})
		r.Use(CountRoutes(tracker, "site"))
		r.Get("/data/{collection}", func(w http.ResponseWriter, r *http.Request) {})
	})
	for _, request := range []struct {
		path     string
		signedIn bool
	}{
		{"/data/glucose", true}, {"/data/insulin", true},
		{"/data/glucose", false}, {"/secret-token", true}, {"/secret-token", false}, {"/htmx.js", true}, {"/htmx.js", false},
	} {
		r := httptest.NewRequest(http.MethodGet, request.path, nil)
		if request.signedIn {
			r.Header.Set("Authorization", "Bearer token")
		}
		router.ServeHTTP(httptest.NewRecorder(), r)
	}
	routes := tracker.Snapshot(time.Now(), 10).Routes
	require.Len(t, routes, 1)
	require.Equal(t, "site GET /data/{collection}", routes[0].Route)
	require.Equal(t, int64(2), routes[0].Count)
}
