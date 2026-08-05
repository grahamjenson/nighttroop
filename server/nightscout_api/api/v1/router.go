package v1

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"nighttroop.t1d.nz/server/nightscout_api/nsutils"
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

	// Post Create
	r.Post("/{collection}", Create(dep))
	r.Post("/{collection}.json", Create(dep))
	r.Post("/{collection}/", Create(dep))

	// Put/Patch Update
	r.Put("/{collection}", Update(dep))
	r.Put("/{collection}.json", Update(dep))
	r.Put("/{collection}/", Update(dep))
	r.Put("/{collection}/{identifier}", Update(dep))
	r.Put("/{collection}/{identifier}.json", Update(dep))

	r.Patch("/{collection}", Update(dep))
	r.Patch("/{collection}.json", Update(dep))
	r.Patch("/{collection}/", Update(dep))
	r.Patch("/{collection}/{identifier}", Update(dep))
	r.Patch("/{collection}/{identifier}.json", Update(dep))

	// DELETE
	r.Delete("/{collection}/{identifier_or_type}", Delete(dep))
	r.Delete("/{collection}/{identifier_or_type}.json", Delete(dep))

	// Weird V1 Ones
	// No optional query params in chi
	// We are not doing user submitted Regexs (that is pain)

	// Slice
	r.Get("/slice/{collection}/{field}", Slice(dep))
	r.Get("/slice/{collection}/{field}/{type}", Slice(dep))
	r.Get("/slice/{collection}/{field}/{type}/{prefix}", Slice(dep))
	r.Get("/slice/{collection}/{field}/{type}/{prefix}/{regex}", Slice(dep))

	// Count
	r.Get("/count/{collection}/where", Count(dep))

	//Times
	r.Get("/times", Times(dep))
	r.Get("/times/{prefix}", Times(dep))
	r.Get("/times/{prefix}/{regex}", Times(dep))

	// Echo & Times Echo
	r.Get("/echo", Echo(dep))
	r.Get("/echo/{collection}", Echo(dep))
	r.Get("/echo/{collection}/{identifier_or_type}", Echo(dep))

	r.Get("/times/echo", Echo(dep))
	r.Get("/times/echo/{prefix}", Echo(dep))
	r.Get("/times/echo/{prefix}/{regex}", Echo(dep))
}
