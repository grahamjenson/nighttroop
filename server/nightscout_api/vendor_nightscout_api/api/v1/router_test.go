package v1

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/config"
	"nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/deps"
	"nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/internal/testsupport"
)

func TestNewNightscoutV1Router(t *testing.T) {
	dep := deps.Dependencies{
		Config: config.Config{APISecret: "secret", DefaultRoles: []string{"readable"}}.WithDefaults(),
		Store:  testsupport.NewMemoryStore(),
		Auth:   testsupport.NewAuthManager("secret", []string{"readable"}),
	}
	router := NewNightscoutV1Router(dep)

	for _, path := range []string{"/status.json", "/versions", "/settings", "/food"} {
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		if rec.Code == http.StatusNotFound {
			t.Fatalf("expected route to exist: %s", path)
		}
	}
}
