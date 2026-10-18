package api_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	static "nighttroop.t1d.nz/server/assets/static"
)

func TestStaticFilesRevalidate(t *testing.T) {
	r := chi.NewRouter()
	assert.NoError(t, static.AddStaticRoutes(r))

	first := httptest.NewRecorder()
	r.ServeHTTP(first, httptest.NewRequest(http.MethodGet, "/style_gen.css", nil))
	assert.Equal(t, http.StatusOK, first.Code)
	assert.Equal(t, "no-cache", first.Header().Get("Cache-Control"))
	etag := first.Header().Get("ETag")
	assert.NotEmpty(t, etag)

	req := httptest.NewRequest(http.MethodGet, "/style_gen.css", nil)
	req.Header.Set("If-None-Match", etag)
	second := httptest.NewRecorder()
	r.ServeHTTP(second, req)
	assert.Equal(t, http.StatusNotModified, second.Code)
	assert.Empty(t, second.Body.String())
}
