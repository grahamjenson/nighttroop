package httputils

import (
	"bytes"
	"encoding/json"
	"fmt"
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
	router.Get(ControlLogsDBPath, echo)

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
			hidden: []string{"querytoken", "bodypassword", "bodysecret", "headertoken", "headersecret", "referertoken", "headerkey"}},
		{name: "a form's password", method: http.MethodPost, path: "/api/v1/treatments", contentType: "application/x-www-form-urlencoded",
			body: "name=admin&password=formpassword", message: "Captured POST /api/v1/treatments", has: []string{"name=admin"}, hidden: []string{"formpassword"}},
		{name: "a token in the path", method: http.MethodGet, path: "/api/v2/authorization/request/pathtoken", message: "Captured GET /api/v2/authorization/request/REDACTED",
			hidden: []string{"pathtoken"}},
		{name: "a connector's login, keys and TOTP code", method: http.MethodPost, path: "/api/v1/treatments", contentType: "application/json",
			body: `{"type":"dexcom","url_or_username":"dexuser","pwd_ns_url":"https://nstoken@ns.example","backup_key":"backupkey1","api_key":"apikey1",` +
				`"totp":"123456","code":"654321","sensor_code":"sensor1","client_email":"me@example.com","enabled":true}`,
			message: "Captured POST /api/v1/treatments", has: []string{`\"type\":\"dexcom\"`, `\"enabled\":true`},
			hidden: []string{"dexuser", "nstoken", "backupkey1", "apikey1", "123456", "654321", "sensor1", "me@example.com"}},
		{name: "a form's TOTP and backup key", method: http.MethodPost, path: "/api/v1/treatments", contentType: "application/x-www-form-urlencoded",
			body: "totp=123456&code=654321&backup_key=backupkey1&language=en", message: "Captured POST /api/v1/treatments", has: []string{"language=en"},
			hidden: []string{"123456", "654321", "backupkey1"}},
		{name: "JSON sent as text is redacted as JSON", method: http.MethodPost, path: "/api/v1/treatments", contentType: "text/plain",
			body: `{"carbs":20,"api_secret":"bodysecret"}`, message: "Captured POST /api/v1/treatments", has: []string{`\"carbs\":20`}, hidden: []string{"bodysecret"}},
		{name: "a body past the most kept keeps only its size", method: http.MethodPost, path: "/api/v1/treatments", contentType: "application/json",
			body: `{"password":"bigpassword","notes":"` + big + `"}`, message: "Captured POST /api/v1/treatments",
			has: []string{`"truncated":true`, `"omitted":true`, fmt.Sprintf(`"bytes":%d`, len(big)+37)}, hidden: []string{"bigpassword", big[:100]}},
		{name: "a body that is not text keeps only its size and type", method: http.MethodPost, path: "/api/v1/treatments", contentType: "application/octet-stream",
			body: "\xff\xfe", message: "Captured POST /api/v1/treatments", has: []string{`"omitted":true`, `"bytes":2`, `"content_type":"application/octet-stream"`},
			hidden: []string{"//4="}},
		{name: "a page showing credentials keeps only its size and type", method: http.MethodPost, path: "/api/v1/treatments", contentType: "text/html",
			body: "<p>Admin password: pagepassword</p>", message: "Captured POST /api/v1/treatments", has: []string{`"omitted":true`, `"content_type":"text/html"`},
			hidden: []string{"pagepassword"}},
		{name: "JSON that does not parse keeps only its size", method: http.MethodPost, path: "/api/v1/treatments", contentType: "application/json",
			body: `{"url_or_username":"dexuser",`, message: "Captured POST /api/v1/treatments", has: []string{`"omitted":true`}, hidden: []string{"dexuser"}},
		{name: "a page's style is not captured", method: http.MethodGet, path: "/style.css", contentType: "text/css", body: "body{}"},
		{name: "a download of the logs DB is captured without its bodies", method: http.MethodGet, path: "http://control.example.com" + ControlLogsDBPath,
			contentType: "application/vnd.sqlite3", body: "\xff\xfe", message: "Captured GET " + ControlLogsDBPath, has: []string{`"status":200`}, hidden: []string{"//4="}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
			r.Header.Set("Content-Type", tt.contentType)
			r.Header.Set("User-Agent", "Loop/3.4")
			r.Header.Set("Authorization", "Bearer headertoken")
			r.Header.Set("Api-Secret", "headersecret")
			r.Header.Set("Referer", "https://site.example/?token=referertoken")
			r.Header.Set("X-Api-Key", "headerkey")
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

// Each key naming a credential or a connector's login is secret, but no other, so what a client sent can still be read
func TestSecret(t *testing.T) {
	tests := []struct {
		key    string
		secret bool
	}{
		{"Authorization", true}, {"Cookie", true}, {"api_secret", true}, {"token", true}, {"read_only_ns_token", true},
		{"password", true}, {"admin_password", true}, {"jwt", true}, {"url_or_username", true}, {"pwd_ns_url", true},
		{"url", true}, {"backup_key", true}, {"api_key", true}, {"X-Api-Key", true}, {"private_key", true},
		{"totp", true}, {"totp_secret", true}, {"code", true}, {"sensor_code", true}, {"client_email", true},
		{"Referer", true}, {"Hx-Current-Url", true}, {"credentials", true}, {"session", true}, {"pwd", true},
		{"User-Agent", false}, {"Content-Type", false}, {"eventType", false}, {"insulin", false}, {"enabled", false},
		{"name", false}, {"region", false}, {"type", false}, {"language", false}, {"count", false}, {"encode", false},
	}
	for _, tt := range tests {
		require.Equal(t, tt.secret, secret(tt.key), tt.key)
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
