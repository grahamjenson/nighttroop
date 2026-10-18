package httputils

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"nighttroop.t1d.nz/utils/config"
)

func TestRequestLogsRedactCredentials(t *testing.T) {
	var logs bytes.Buffer
	router := chi.NewRouter()
	AddStandardMiddleware(router, &config.ServerConfig{Logger: slog.New(slog.NewTextHandler(&logs, nil))})
	router.Get("/api/v2/authorization/request/{token}", func(w http.ResponseWriter, r *http.Request) {})

	r := httptest.NewRequest(http.MethodGet, "/api/v2/authorization/request/pathtoken123?token=querytoken456&secret=secret789&count=10", nil)
	router.ServeHTTP(httptest.NewRecorder(), r)

	assert.Contains(t, logs.String(), "count=10")
	for _, credential := range []string{"pathtoken123", "querytoken456", "secret789"} {
		assert.NotContains(t, logs.String(), credential)
	}
}
