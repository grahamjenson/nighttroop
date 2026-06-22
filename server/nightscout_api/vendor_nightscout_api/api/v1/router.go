package v1

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/nsutils"
)

// NewNightscoutV1Router constructs a router that serves the supported
// Nightscout API v1 endpoints.
func NewNightscoutV1Router(dep nsutils.Dependencies) *chi.Mux {
	r := chi.NewRouter()

	registerMiscRoutes(r, dep)
	registerCollectionRoutes(r, dep)

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		slog.Info("Nightscout router 404 for path", "path", r.URL.Path)
		http.Error(w, "Not Found", http.StatusNotFound)
	})
	return r
}

func registerMiscRoutes(r *chi.Mux, dep nsutils.Dependencies) {
	r.Get("/status.json", StatusJSON(dep))
	r.Get("/versions", Versions())
	r.Get("/verifyauth", VerifyAuth(dep))
}

func registerCollectionRoutes(r *chi.Mux, dep nsutils.Dependencies) {
	// Get And Search
	r.Get("/{collection}", Search(dep))
	r.Get("/{collection}/", Search(dep))
	r.Get("/{collection}.json", Search(dep))
	r.Get("/{collection}/{identifier_or_type}", Search(dep))
	r.Get("/{collection}/{identifier_or_type}.json", Search(dep))
	r.Get("/{collection}/current", Current(dep))
	r.Get("/{collection}/current.json", Current(dep))

	// Post And Put (Create/Update)
	r.Post("/{collection}", Create(dep))
	r.Post("/{collection}.json", Create(dep))
	r.Post("/{collection}/", Create(dep))

	// DELETE
	r.Delete("/{collection}", Delete(dep))
	r.Delete("/{collection}.json", Delete(dep))
	r.Delete("/{collection}/{identifier_or_type}", Delete(dep))
	r.Delete("/{collection}/{identifier_or_type}.json", Delete(dep))

	// Weird V1 Ones
	r.Get("/echo/{collection}/{spec}.json", Echo(dep))
	r.Get("/slice/{collection}/{spec}.json", Slice(dep))
	r.Get("/times/{collection}/{spec}.json", Times(dep))

}
