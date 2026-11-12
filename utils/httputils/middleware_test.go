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
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/utils/aslog"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/consts"
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

// A Socket.IO session is authorized once, then each long poll and POST to it carries only its sid, so whoever reads
// the logs or Sentry must not get the sid, which would let them read and write the site as the session's identity
func TestRequestLogsAndSentryRedactSocketIOSessionID(t *testing.T) {
	var logs bytes.Buffer
	router := chi.NewRouter()
	testutils.UseLogHandler(t, slog.NewTextHandler(&logs, nil))
	AddStandardMiddleware(router, &config.ServerConfig{}, nil)
	router.Post("/socket.io/", func(w http.ResponseWriter, r *http.Request) {})

	target := "https://abcdefgh.ns.example.com/socket.io/?EIO=4&transport=polling&sid=sessionid789"
	router.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, target, nil))
	assert.Contains(t, logs.String(), "transport=polling")
	assert.NotContains(t, logs.String(), "sessionid789")

	event := RedactSentryEvent(&sentry.Event{Request: sentry.NewRequest(httptest.NewRequest(http.MethodPost, target, nil))}, nil)
	sent, err := json.Marshal(event.Request)
	require.NoError(t, err)
	assert.Contains(t, string(sent), "transport=polling")
	assert.NotContains(t, string(sent), "sessionid789")
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

// Everything logged for a request, and its request log, says which request it was, its site, and once signed in its
// identity, so the logs before an error can be found by any of them
func TestRequestLogsCarryRequestSiteAndIdentity(t *testing.T) {
	var logs bytes.Buffer
	testutils.UseLogHandler(t, aslog.ContextHandler{Handler: slog.NewJSONHandler(&logs, nil)})
	router := chi.NewRouter()
	AddStandardMiddleware(router, &config.ServerConfig{}, nil)
	router.Get("/data", func(w http.ResponseWriter, r *http.Request) {
		slog.InfoContext(r.Context(), "before sign in")
		ctx := SetIdentityModelInSiteContext(r.Context(), &controldb.SiteIdentity{ID: 7, Name: "phone", Role: consts.IdentityRoleUser})
		slog.InfoContext(ctx, "after sign in")
		w.WriteHeader(http.StatusOK)
	})
	// The main router gives a site's requests their site context before its router's middleware
	site := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		router.ServeHTTP(w, r.WithContext(CreateSiteContext(r.Context(), nil, nil, nil, nil, "site-a", "", HTTPActionRead, AcceptTypeJSON, ContentTypeJSON)))
	})

	site.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/data", nil))

	lines := map[string]map[string]any{}
	for line := range bytes.Lines(logs.Bytes()) {
		var record map[string]any
		require.NoError(t, json.Unmarshal(line, &record))
		lines[record["msg"].(string)] = record
	}
	require.Contains(t, lines, "200: OK")
	requestID := lines["before sign in"]["request_id"]
	require.NotEmpty(t, requestID)
	require.Equal(t, "site-a", lines["before sign in"]["site_id"])
	require.NotContains(t, lines["before sign in"], "identity_id")
	for _, msg := range []string{"after sign in", "200: OK"} {
		require.Equal(t, requestID, lines[msg]["request_id"], msg)
		require.Equal(t, "site-a", lines[msg]["site_id"], msg)
		require.EqualValues(t, 7, lines[msg]["identity_id"], msg)
		require.Equal(t, "phone", lines[msg]["identity_name"], msg)
		require.Equal(t, string(consts.IdentityRoleUser), lines[msg]["role"], msg)
	}
}

// A request is cancelled once it has taken too long, but an event stream, held open on purpose, or the control site's
// download of the logs DB, which can take minutes, is not
func TestTimeoutSparesEventStreams(t *testing.T) {
	timeout := requestTimeout
	requestTimeout = 50 * time.Millisecond
	t.Cleanup(func() { requestTimeout = timeout })
	router := chi.NewRouter()
	AddStandardMiddleware(router, &config.ServerConfig{}, nil)
	wait := func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
			_, _ = io.WriteString(w, "cancelled")
		case <-time.After(500 * time.Millisecond):
			_, _ = io.WriteString(w, "finished")
		}
	}
	router.Get("/wait", wait)
	router.Get("/data/{collection}/events", wait)
	router.Get(ControlLogsDBPath, wait)

	for _, test := range []struct{ host, path, accept, want string }{
		{"example.com", "/wait", "text/html", "cancelled"},
		{"example.com", "/wait", "text/event-stream", "cancelled"},
		{"site.example.com", "/data/glucose/events", "text/event-stream", "finished"},
		{"control.example.com", ControlLogsDBPath, "*/*", "finished"},
		{"site.example.com", ControlLogsDBPath, "*/*", "cancelled"},
	} {
		r := httptest.NewRequest(http.MethodGet, "http://"+test.host+test.path, nil)
		r.Header.Set("Accept", test.accept)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, r)
		assert.Equal(t, test.want, w.Body.String(), test.host+test.path+" "+test.accept)
	}
}
