package nsv3

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"nighttroop.t1d.nz/server/nightscoutapi/nsutils"
)

// NewNightscoutV3Router constructs a router that serves the supported
// Nightscout API v3 endpoints.
func NewNightscoutV3Router(store nsutils.Store) http.Handler {
	r := chi.NewRouter()

	r.Get("/version", Version(store))
	r.Get("/status", nsutils.Require("api:entries:read", false, Status(store)))
	r.Get("/test", nsutils.Require("api:entries:read", false, Test(store)))
	r.Get("/lastmodified", nsutils.Require("api:entries:read", false, LastModified(store)))
	r.Get("/lastModified", nsutils.Require("api:entries:read", false, LastModified(store)))

	r.Get("/{collection}/history/{since}", History(store))
	r.Get("/{collection}/{identifier}", Read(store))
	r.Put("/{collection}/{identifier}", Replace(store))
	r.Patch("/{collection}/{identifier}", Patch(store))
	r.Delete("/{collection}/{identifier}", Remove(store))
	r.Get("/{collection}", Search(store))
	r.Post("/{collection}", Create(store))

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		slog.InfoContext(r.Context(), "Nightscout router 404 for path", "path", r.URL.Path)
		http.Error(w, "Not Found", http.StatusNotFound)
	})

	return r
}
