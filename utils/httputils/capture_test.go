package httputils

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/logsdb"
	"nighttroop.t1d.nz/utils/aslog"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/testutils"
)

// Each request is captured as a log with its site and identity, headers and bodies, with its credentials redacted
func TestCaptureRequests(t *testing.T) {
	_, controlDB := testutils.NewControlDB(t)
	router := chi.NewRouter()
	AddStandardMiddleware(router, &config.ServerConfig{CaptureRequests: true}, controlDB)
	echo := func(w http.ResponseWriter, r *http.Request) {
		aslog.AddAttrs(r.Context(), slog.String("site_id", "abcdefgh"), slog.String("identity_name", "loop_phone"))
		var body bytes.Buffer
		_, _ = body.ReadFrom(r.Body)
		w.Header().Set("Content-Type", r.Header.Get("Content-Type"))
		_, _ = w.Write(body.Bytes())
	}
	router.Post("/api/v1/treatments", echo)
	router.Get("/api/v2/authorization/request/{token}", echo)
	router.Get("/style.css", echo)

	big := strings.Repeat("a", maxCapturedBody+10)
	tests := []struct {
		name, method, path, contentType, body string
		// message is the log's, or empty when the request is not captured
		message string
		// has are in the captured attributes, and hidden, credentials, are not
		has, hidden []string
	}{
		{name: "a client's JSON, with its headers, query, site and identity", method: http.MethodPost, path: "/api/v1/treatments?token=querytoken&count=10",
			contentType: "application/json", body: `{"eventType":"Bolus","insulin":1.5,"pump":{"password":"bodypassword","api_secret":"bodysecret"}}`,
			message: "Captured POST /api/v1/treatments", has: []string{`\"insulin\":1.5`, `"count":["10"]`, `"Loop/3.4"`, `"identity_name":"loop_phone"`, `"status":200`, "REDACTED"},
			hidden: []string{"querytoken", "bodypassword", "bodysecret", "headertoken", "headersecret"}},
		{name: "a form's password", method: http.MethodPost, path: "/api/v1/treatments", contentType: "application/x-www-form-urlencoded",
			body: "name=admin&password=formpassword", message: "Captured POST /api/v1/treatments", has: []string{"name=admin"}, hidden: []string{"formpassword"}},
		{name: "a token in the path", method: http.MethodGet, path: "/api/v2/authorization/request/pathtoken", message: "Captured GET /api/v2/authorization/request/REDACTED",
			hidden: []string{"pathtoken"}},
		{name: "a body past the most kept is cut", method: http.MethodPost, path: "/api/v1/treatments", contentType: "text/plain", body: big,
			message: "Captured POST /api/v1/treatments", has: []string{`"truncated":true`, `"text":"` + big[:maxCapturedBody] + `"`}},
		{name: "a body that is not text is base64", method: http.MethodPost, path: "/api/v1/treatments", contentType: "application/octet-stream", body: "\xff\xfe",
			message: "Captured POST /api/v1/treatments", has: []string{`"base64":"//4="`}},
		{name: "a page's HTML is kept as it is", method: http.MethodPost, path: "/api/v1/treatments", contentType: "text/html", body: "<p>a & b</p>",
			message: "Captured POST /api/v1/treatments", has: []string{`"text":"<p>a & b</p>"`}},
		{name: "a page's style is not captured", method: http.MethodGet, path: "/style.css", contentType: "text/css", body: "body{}"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
			r.Header.Set("Content-Type", tt.contentType)
			r.Header.Set("User-Agent", "Loop/3.4")
			r.Header.Set("Authorization", "Bearer headertoken")
			r.Header.Set("Api-Secret", "headersecret")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, r)
			require.Equal(t, tt.body, w.Body.String(), "the response is not changed by capturing it")

			captured := capturedLogs(t, controlDB.Logs.Queries)
			if tt.message == "" {
				require.Empty(t, captured)
				return
			}
			require.Len(t, captured, 1)
			require.Equal(t, tt.message, captured[0].Message)
			require.Equal(t, "abcdefgh", *captured[0].SiteID)
			require.NotNil(t, captured[0].RequestID)
			for _, want := range tt.has {
				require.Contains(t, *captured[0].Attributes, want)
			}
			for _, credential := range tt.hidden {
				require.NotContains(t, *captured[0].Attributes, credential)
			}
			require.True(t, json.Valid([]byte(*captured[0].Attributes)))
		})
	}
}

// capturedLogs are the captured requests logged since the last call, which deletes them
func capturedLogs(t *testing.T, q *logsdb.Queries) []*logsdb.Log {
	rows, err := q.SearchLogs(t.Context(), logsdb.SearchLogsParams{Limit: 100})
	require.NoError(t, err)
	captured := []*logsdb.Log{}
	for _, row := range rows {
		if strings.HasPrefix(row.Message, "Captured ") {
			captured = append(captured, row)
		}
	}
	_, err = q.TrimLogs(t.Context(), logsdb.TrimLogsParams{Keep: 0})
	require.NoError(t, err)
	return captured
}
