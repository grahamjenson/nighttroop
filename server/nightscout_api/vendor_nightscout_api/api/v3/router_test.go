package v3

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/config"
	"nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/deps"
	"nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/internal/testsupport"
)

func TestNewNightscoutV3Router(t *testing.T) {
	dep := deps.Dependencies{
		Config: config.Config{APISecret: "secret", DefaultRoles: []string{"readable"}}.WithDefaults(),
		Store:  testsupport.NewMemoryStore(),
		Auth:   testsupport.NewAuthManager("secret", []string{"readable"}),
	}
	router := NewNightscoutV3Router(dep)

	for _, path := range []string{"/version", "/status", "/test", "/lastModified"} {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, path, nil)
		if path != "/version" {
			req.Header.Set("api-secret", "secret")
		}
		router.ServeHTTP(rec, req)
		if rec.Code == http.StatusNotFound {
			t.Fatalf("expected route to exist: %s", path)
		}
	}
}
