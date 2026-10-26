package httputils

import (
	"bytes"
	"compress/gzip"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/testutils"
)

func TestRequestLogsRedactCredentials(t *testing.T) {
	var logs bytes.Buffer
	router := chi.NewRouter()
	testutils.UseLogHandler(t, slog.NewTextHandler(&logs, nil))
	AddStandardMiddleware(router, &config.ServerConfig{}, nil)
	router.Get("/api/v2/authorization/request/{token}", func(w http.ResponseWriter, r *http.Request) {})

	r := httptest.NewRequest(http.MethodGet, "/api/v2/authorization/request/pathtoken123?token=querytoken456&count=10", nil)
	router.ServeHTTP(httptest.NewRecorder(), r)

	assert.Contains(t, logs.String(), "count=10")
	for _, credential := range []string{"pathtoken123", "querytoken456"} {
		assert.NotContains(t, logs.String(), credential)
	}
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
