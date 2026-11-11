package httputils

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/getsentry/sentry-go"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/data/sql/logsdb"
	"nighttroop.t1d.nz/utils/testutils"
)

// A request refused, or that failed, is recorded with what it sent, its credentials redacted, as JSON or, when it is
// not, as its text, so a client's bad request can be debugged
func TestErrorsKeepTheRequestBody(t *testing.T) {
	tests := []struct {
		name     string
		status   int
		body     string
		contains []string
	}{
		{"a 400 keeps its JSON, its token redacted", http.StatusBadRequest, `{"sgv":"high","token":"bodytoken123"}`,
			[]string{`"kind":"client"`, `\"sgv\":\"high\"`, `\"token\":\"REDACTED\"`}},
		{"a 400 keeps a connector's login redacted", http.StatusBadRequest, `{"url_or_username":"bodytoken123","totp":"123456"}`,
			[]string{`\"url_or_username\":\"REDACTED\"`, `\"totp\":\"REDACTED\"`}},
		{"a 400 keeps only the size and type of JSON that did not parse", http.StatusBadRequest, `{"sgv": 102, "token": "bodytoken123",`,
			[]string{`"omitted":true`, `"bytes":37`, `"content_type":"application/json"`}},
		{"a 500 keeps its JSON", http.StatusInternalServerError, `{"eventType":"Note"}`, []string{`"kind":"server"`, `\"eventType\":\"Note\"`}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cfg, controlDB := testutils.NewControlDB(t)
			router := chi.NewRouter()
			AddStandardMiddleware(router, cfg, controlDB)
			router.Post("/api/v1/{collection}", func(w http.ResponseWriter, r *http.Request) {
				_, _ = io.ReadAll(r.Body)
				w.WriteHeader(test.status)
			})
			r := httptest.NewRequest(http.MethodPost, "/api/v1/entries", strings.NewReader(test.body))
			r.Header.Set("Content-Type", "application/json")
			r = r.WithContext(CreateSiteContext(r.Context(), cfg, controlDB, nil, &controldb.Site{ID: "somesite"}, "somesite", "", HTTPActionWrite, AcceptTypeJSON, ContentTypeJSON))
			router.ServeHTTP(httptest.NewRecorder(), r)

			recorded, err := controlDB.Logs.Queries.SearchErrors(t.Context(), logsdb.SearchErrorsParams{Limit: 10})
			require.NoError(t, err)
			require.Len(t, recorded, 1)
			for _, part := range test.contains {
				require.Contains(t, *recorded[0].Attributes, part)
			}
			require.NotContains(t, *recorded[0].Attributes, "bodytoken123")
		})
	}
}

// What a request sent is kept only up to errorRequestBodyBytes, so a large one cannot fill an error, and as it cannot be
// redacted once cut off, only its size is
func TestErrorRequestBodyIsCapped(t *testing.T) {
	var kept map[string]any
	handler := KeepErrorRequestBodies(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.ReadAll(r.Body)
		kept = ErrorRequestBody(r.Context())
	}))
	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/", strings.NewReader(strings.Repeat("a", 3*errorRequestBodyBytes))))
	require.Equal(t, map[string]any{"truncated": true, "omitted": true, "bytes": 3 * errorRequestBodyBytes, "content_type": ""}, kept)
}

// An error reported to Sentry while handling a request has what the request sent, its credentials redacted
func TestErrorRequestBodyReportsToSentry(t *testing.T) {
	sent := &sentEvents{}
	client, err := sentry.NewClient(sentry.ClientOptions{Dsn: "https://key@sentry.example/1", Transport: sent, BeforeSend: RedactSentryEvent})
	require.NoError(t, err)
	handler := KeepErrorRequestBodies(reportErrorRequestBodies(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.ReadAll(r.Body)
		sentry.GetHubFromContext(r.Context()).CaptureMessage("cannot store the treatment")
	})))
	r := httptest.NewRequest(http.MethodPost, "/api/v1/treatments", strings.NewReader(`{"carbs":20,"api_secret":"bodysecret123"}`))
	r = r.WithContext(sentry.SetHubOnContext(r.Context(), sentry.NewHub(client, sentry.NewScope())))
	handler.ServeHTTP(httptest.NewRecorder(), r)

	require.Len(t, sent.events, 1)
	text := sent.events[0].Contexts["request_body"]["text"]
	require.Contains(t, text, `"carbs":20`)
	require.Contains(t, text, `"api_secret":"REDACTED"`)
}
