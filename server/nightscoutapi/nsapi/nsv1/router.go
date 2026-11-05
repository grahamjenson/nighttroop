package nsv1

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"nighttroop.t1d.nz/server/nightscoutapi/nsutils"
)

// NewNightscoutV1Router constructs a router that serves the supported
// Nightscout API v1 endpoints.
func NewNightscoutV1Router(store nsutils.Store) *chi.Mux {
	r := chi.NewRouter()

	registerMiscRoutes(r, store)

	registerCollectionRoutes(r, store)

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		slog.InfoContext(r.Context(), "Nightscout router 404 for path", "path", r.URL.Path)
		http.Error(w, "Not Found", http.StatusNotFound)
	})
	return r
}

func registerMiscRoutes(r *chi.Mux, store nsutils.Store) {
	r.Get("/status", StatusJSON(store))
	r.Get("/status.json", StatusJSON(store))
	r.Get("/versions", Versions())
	r.Get("/versions.json", Versions())
	r.Get("/verifyauth", VerifyAuth(store))
	r.Get("/experiments/test", ExperimentsTest())
	r.Get("/notifications/ack", AckNotifications())
	r.Get("/adminnotifies", AdminNotifies())
}

func registerCollectionRoutes(r *chi.Mux, store nsutils.Store) {
	// Get And Search
	r.Get("/{collection}", Search(store))
	r.Get("/{collection}/", Search(store))
	r.Get("/{collection}.json", Search(store))
	r.Get("/{collection}/{identifier_or_type}", Search(store))
	r.Get("/{collection}/{identifier_or_type}.json", Search(store))
	r.Get("/{collection}/current", Current(store))
	r.Get("/{collection}/current.json", Current(store))

	// Post Create
	r.Post("/{collection}", Create(store))
	r.Post("/{collection}.json", Create(store))
	r.Post("/{collection}/", Create(store))

	// Put/Patch Update
	r.Put("/{collection}", Replace(store))
	r.Put("/{collection}.json", Replace(store))
	r.Put("/{collection}/", Replace(store))
	r.Put("/{collection}/{identifier}", Replace(store))
	r.Put("/{collection}/{identifier}.json", Replace(store))

	// DELETE
	r.Delete("/{collection}", Delete(store))
	r.Delete("/{collection}.json", Delete(store))
	r.Delete("/{collection}/{identifier_or_type}", Delete(store))
	r.Delete("/{collection}/{identifier_or_type}.json", Delete(store))

	// Weird V1 Ones
	// No optional query params in chi
	// We are not doing user submitted Regexs (that is pain)

	// Slice
	r.Get("/slice/{collection}/{field}/{type}/{prefix}", Slice(store))

	// Count
	r.Get("/count/{collection}/where", Count(store))

	//Times - return entries with {prefix} dateString
	r.Get("/times/{prefix}", Times(store))

	// Echo
	r.Get("/echo/{collection}", Echo(store))
	r.Get("/echo/{collection}/{identifier_or_type}", Echo(store))
}
