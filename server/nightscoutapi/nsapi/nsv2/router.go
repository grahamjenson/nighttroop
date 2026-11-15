package nsv2

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"nighttroop.t1d.nz/server/nightscoutapi/nsutils"
	"nighttroop.t1d.nz/utils/apiutils"
)

// NewNightscoutV2Router constructs a router that serves the supported
// Nightscout API v2 endpoints.
func NewNightscoutV2Router(store nsutils.Store) http.Handler {
	r := chi.NewRouter()

	registerPropertiesRoutes(r, store)
	r.Get("/authorization/debug/check/{permission}", CheckPermission)
	registerAuthorizationRoutes(r)
	// Nightscout needs notifications:loop:push; here, as a remote command makes a treatment, an identity that can write them
	r.Post("/notifications/loop", nsutils.Require("api:treatments:create", false, LoopNotification))

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

// CheckPermission is Nightscout's check of whether a token or JWT has a permission, such as api:settings:create,
// which AAPS asks when the v3 status does not show every right: 200 when it has, 403 when not
func CheckPermission(w http.ResponseWriter, r *http.Request) {
	permission := chi.URLParam(r, "permission")
	nsutils.Require(permission, false, func(w http.ResponseWriter, _ *http.Request, _ *nsutils.Identity) {
		apiutils.WriteJSON(w, http.StatusOK, map[string]any{"status": http.StatusOK, "permission": permission})
	})(w, r)
}
