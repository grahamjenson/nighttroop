package v2

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/api/v2/handlers"
	"nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/di"
)

// NewNightscoutV2Router constructs a router that serves the supported
// Nightscout API v2 endpoints.
func NewNightscoutV2Router(dep di.Dependencies) http.Handler {
	r := chi.NewRouter()

	registerPropertiesRoutes(r, dep)

	return r
}

func registerPropertiesRoutes(r *chi.Mux, dep di.Dependencies) {
	r.Get("/properties", dep.Auth.Require("api:properties:read", true, handlers.PropertiesList(dep)))
	r.Get("/properties.json", dep.Auth.Require("api:properties:read", true, handlers.PropertiesList(dep)))

}
