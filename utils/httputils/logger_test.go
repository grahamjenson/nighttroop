package httputils

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// logRequest serves one request through a router logging with config, and is the request's log line as JSON
func logRequest(t *testing.T, config Config, status int, r *http.Request) map[string]any {
	t.Helper()
	var logs bytes.Buffer
	router := chi.NewRouter()
	router.Use(NewLoggerWithConfig(slog.New(slog.NewJSONHandler(&logs, &slog.HandlerOptions{Level: slog.LevelDebug})), config))
	router.Post("/things/{token}", func(w http.ResponseWriter, r *http.Request) {
		io.ReadAll(r.Body)
		w.Header().Set("Set-Cookie", "session=cookiesecret")
		w.Header().Set("X-Answer", "42")
		w.WriteHeader(status)
		w.Write([]byte("response body"))
	})
	router.ServeHTTP(httptest.NewRecorder(), r)

	var line map[string]any
	require.NoError(t, json.Unmarshal(logs.Bytes(), &line))
	return line
}

func newLoggedRequest() *http.Request {
	r := httptest.NewRequest(http.MethodPost, "/things/pathtoken123?token=querytoken456&count=10", strings.NewReader("request body"))
	r.Header.Set("Authorization", "Bearer bearertoken111")
	r.Header.Set("X-Question", "life")
	r.Header.Set("User-Agent", "tester/1")
	return r
}

func TestRequestLogHasWhatTheConfigAsksFor(t *testing.T) {
	all := DefaultConfig()
	all.WithUserAgent, all.WithRequestBody, all.WithRequestHeader, all.WithResponseBody, all.WithResponseHeader = true, true, true, true, true

	line := logRequest(t, all, http.StatusOK, newLoggedRequest())
	assert.Equal(t, "200: OK", line["msg"])
	request, response := line["request"].(map[string]any), line["response"].(map[string]any)

	assert.Equal(t, "POST", request["method"])
	assert.Equal(t, "example.com", request["host"])
	assert.Equal(t, "/things/REDACTED", request["path"])
	assert.Equal(t, "count=10&token=REDACTED", request["query"])
	assert.Equal(t, map[string]any{"token": "REDACTED"}, request["params"])
	assert.Equal(t, "/things/{token}", request["route"])
	assert.Equal(t, "192.0.2.1:1234", request["ip"])
	assert.Equal(t, float64(len("request body")), request["length"])
	assert.Equal(t, "request body", request["body"])
	assert.Equal(t, "tester/1", request["user-agent"])
	assert.Equal(t, map[string]any{"X-Question": []any{"life"}, "User-Agent": []any{"tester/1"}}, request["header"])

	assert.Equal(t, float64(http.StatusOK), response["status"])
	assert.Equal(t, float64(len("response body")), response["length"])
	assert.Equal(t, "response body", response["body"])
	assert.Equal(t, map[string]any{"X-Answer": []any{"42"}}, response["header"])
	assert.Contains(t, response, "latency")
	assert.Contains(t, response, "time")

	// The default logs neither body, headers nor user agent
	line = logRequest(t, DefaultConfig(), http.StatusOK, newLoggedRequest())
	request, response = line["request"].(map[string]any), line["response"].(map[string]any)
	for _, key := range []string{"body", "header", "user-agent"} {
		assert.NotContains(t, request, key)
		assert.NotContains(t, response, key)
	}
	assert.Equal(t, "192.0.2.1:1234", request["ip"])
}

func TestRequestLogLevelIsFromItsStatus(t *testing.T) {
	config := Config{DefaultLevel: slog.LevelDebug, ClientErrorLevel: slog.LevelWarn, ServerErrorLevel: slog.LevelError}
	for _, test := range []struct {
		status int
		level  string
	}{
		{http.StatusOK, "DEBUG"},
		{http.StatusFound, "DEBUG"},
		{http.StatusBadRequest, "WARN"},
		{http.StatusTooManyRequests, "WARN"},
		{http.StatusInternalServerError, "ERROR"},
		{http.StatusServiceUnavailable, "ERROR"},
	} {
		line := logRequest(t, config, test.status, newLoggedRequest())
		assert.Equal(t, test.level, line["level"], "status %d", test.status)
	}

	config.WithCustomMessage = func(w http.ResponseWriter, r *http.Request) string { return "custom " + r.Method }
	assert.Equal(t, "custom POST", logRequest(t, config, http.StatusOK, newLoggedRequest())["msg"])
}
