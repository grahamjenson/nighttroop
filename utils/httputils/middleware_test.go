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
	"nighttroop.t1d.nz/utils/testutils"
)

func TestRequestLogsRedactCredentials(t *testing.T) {
	var logs bytes.Buffer
	router := chi.NewRouter()
	testutils.UseLogHandler(t, slog.NewTextHandler(&logs, nil))
	AddStandardMiddleware(router, &config.ServerConfig{})
	router.Get("/api/v2/authorization/request/{token}", func(w http.ResponseWriter, r *http.Request) {})

	r := httptest.NewRequest(http.MethodGet, "/api/v2/authorization/request/pathtoken123?token=querytoken456&count=10", nil)
	router.ServeHTTP(httptest.NewRecorder(), r)

	assert.Contains(t, logs.String(), "count=10")
	for _, credential := range []string{"pathtoken123", "querytoken456"} {
		assert.NotContains(t, logs.String(), credential)
	}
}
