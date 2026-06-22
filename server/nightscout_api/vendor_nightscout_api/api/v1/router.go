package v1

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/api/v1/handlers"
	"nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/nsutils"
)

// NewNightscoutV1Router constructs a router that serves the supported
// Nightscout API v1 endpoints.
func NewNightscoutV1Router(dep nsutils.Dependencies) *chi.Mux {
	r := chi.NewRouter()

	registerMiscRoutes(r, dep)
	registerCollectionRoutes(r, dep)
	registerSearchRoutes(r, dep)
	registerEntryRoutes(r, dep)
	registerTreatmentRoutes(r, dep)
	registerDeviceStatusRoutes(r, dep)
	registerProfileRoutes(r, dep)
	registerSettingRoutes(r, dep)
	registerFoodRoutes(r, dep)

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		slog.Info("Nightscout router 404 for path", "path", r.URL.Path)
		http.Error(w, "Not Found", http.StatusNotFound)
	})
	return r
}

func registerMiscRoutes(r *chi.Mux, dep nsutils.Dependencies) {
	r.Get("/status.json", handlers.StatusJSON(dep))
	r.Get("/versions", handlers.Versions())
	r.Get("/verifyauth", handlers.VerifyAuth(dep))
}

func registerCollectionRoutes(r *chi.Mux, dep nsutils.Dependencies) {
	// Get And Search
	r.Get("/{collection}", handlers.Search(dep))
	r.Get("/{collection}.json", handlers.Search(dep))
	r.Get("/{collection}/{identifier}", handlers.Read(dep))

	// Post And Put (Create/Update)
	r.Post("/{collection}", handlers.Create(dep))
	r.Post("/{collection}.json", handlers.Create(dep))
	r.Post("/{collection}/", handlers.Create(dep))

	// DELETE
	r.Delete("/{collection}", handlers.Delete(dep))
	r.Delete("/{collection}.json", handlers.Delete(dep))
	r.Delete("/{collection}/", handlers.Delete(dep))

	// Weird V1 Ones
	r.Get("/echo/{collection}/{spec}.json", handlers.Echo(dep))
	r.Get("/slice/{collection}/{spec}.json", handlers.Slice(dep))
	r.Get("/times/{collection}/{spec}.json", handlers.Times(dep))

}

func registerEntryRoutes(r *chi.Mux, dep nsutils.Dependencies) {
	r.Get("/entries/current.json", dep.Auth.Require("api:entries:read", true, handlers.EntriesCurrent(dep)))

	r.Get("/entries/{spec}.json", dep.Auth.Require("api:entries:read", true, handlers.EntriesSpec(dep)))
	r.Delete("/entries/{spec}", dep.Auth.Require("api:entries:delete", false, handlers.EntriesDelete(dep)))
}
