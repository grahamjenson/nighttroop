package v2

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"nighttroop.t1d.nz/server/nightscoutapi/nsutils"
)

// NewNightscoutV2Router constructs a router that serves the supported
// Nightscout API v2 endpoints.
func NewNightscoutV2Router(dep nsutils.Dependencies) http.Handler {
	r := chi.NewRouter()

	registerPropertiesRoutes(r, dep)

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		slog.Info("Nightscout router 404 for path", "path", r.URL.Path)
		http.Error(w, "Not Found", http.StatusNotFound)
	})

	return r
}

func registerPropertiesRoutes(r *chi.Mux, dep nsutils.Dependencies) {
	r.Get("/properties", dep.Auth.Require("api:properties:read", true, PropertiesList(dep)))
	r.Get("/properties.json", dep.Auth.Require("api:properties:read", true, PropertiesList(dep)))

	r.Get("/summary", dep.Auth.Require("api:summary:read", true, Summary(dep)))
	r.Get("/summary.json", dep.Auth.Require("api:summary:read", true, Summary(dep)))
}
