package nsv2

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"nighttroop.t1d.nz/server/nightscoutapi/nsutils"
)

// NewNightscoutV2Router constructs a router that serves the supported
// Nightscout API v2 endpoints.
func NewNightscoutV2Router(store nsutils.Store) http.Handler {
	r := chi.NewRouter()

	registerPropertiesRoutes(r, store)

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		slog.InfoContext(r.Context(), "Nightscout router 404 for path", "path", r.URL.Path)
		http.Error(w, "Not Found", http.StatusNotFound)
	})

	return r
}

func registerPropertiesRoutes(r *chi.Mux, store nsutils.Store) {
	r.Get("/properties", nsutils.Require("api:properties:read", true, PropertiesList(store)))
	r.Get("/properties.json", nsutils.Require("api:properties:read", true, PropertiesList(store)))
	r.Get("/properties/{names}", nsutils.Require("api:properties:read", true, PropertiesList(store)))

	r.Get("/summary", nsutils.Require("api:summary:read", true, Summary(store)))
	r.Get("/summary.json", nsutils.Require("api:summary:read", true, Summary(store)))
}
