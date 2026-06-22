package v3

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/nsutils"
)

// NewNightscoutV3Router constructs a router that serves the supported
// Nightscout API v3 endpoints.
func NewNightscoutV3Router(dep nsutils.Dependencies) http.Handler {
	r := chi.NewRouter()

	r.Get("/version", Version(dep))
	r.Get("/status", dep.Auth.Require("api:entries:read", false, Status(dep)))
	r.Get("/test", dep.Auth.Require("api:entries:read", false, Test(dep)))
	r.Get("/lastModified", dep.Auth.Require("api:entries:read", false, LastModified(dep)))

	r.Get("/{collection}/history/{since}", History(dep))
	r.Get("/{collection}/{identifier}", Read(dep))
	r.Put("/{collection}/{identifier}", Update(dep))
	r.Patch("/{collection}/{identifier}", Patch(dep))
	r.Delete("/{collection}/{identifier}", Remove(dep))
	r.Get("/{collection}", Search(dep))
	r.Post("/{collection}", Create(dep))

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		slog.Info("Nightscout router 404 for path", "path", r.URL.Path)
		http.Error(w, "Not Found", http.StatusNotFound)
	})

	return r
}
