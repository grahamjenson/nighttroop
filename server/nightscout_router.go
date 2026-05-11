package server

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/server/assets/static"
	"nighttroop.t1d.nz/server/nightscout"
	"nighttroop.t1d.nz/server/user"
	"nighttroop.t1d.nz/utils/config"
)

func NewNightscoutRouter(controlDB *db.ControlDB, config config.ServerConfig) chi.Router {
	r := chi.NewRouter()
	AddStandardMiddleware(r, config)
	r.Use(user.UserRateLimitMiddleware(config))

	static.AddSatticRoutes(r)

	// Only Unauthed Endpoints are the ones to generate the JWT token
	// Can pass token asa query param, or JWT in auth header (recommended)
	r.Get("/api/v2/authorization/request", nightscout.NSGetNightscoutJWTToken)
	r.Post("/api/v2/authorization/request", nightscout.NSGetNightscoutJWTToken)

	// Versioned APIs
	r.Mount("/api", apiRoutes(controlDB, config))

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		slog.Info("Nightscout router 404 for path", "path", r.URL.Path)
		http.Error(w, "Not Found", http.StatusNotFound)
	})

	return r
}

func apiRoutes(controlDB *db.ControlDB, config config.ServerConfig) chi.Router {
	// I am going to be SUPER VERBOSE on these routes to make sure we know 100% implemented scope
	r := chi.NewRouter()
	r.Use(nightscout.NightscoutAuthMiddleware(controlDB, config))

	// ---- AUTH ----
	r.Post("/v2/authorization/refresh", nightscout.NSRefreshNightscoutJWTToken)
	r.Post("/v2/authorization/revoke", nightscout.NSRevokeNightscoutJWTToken)

	// ---- ENTRIES (sgv) ----
	r.Get("/v1/entries", nightscout.NSGetEntries)
	r.Get("/v1/entries.json", nightscout.NSGetEntries)
	r.Post("/v1/entries", nightscout.NSCreateEntry)
	r.Post("/v1/entries.json", nightscout.NSCreateEntry)
	//
	r.Get("/v3/entries", nightscout.NSGetEntries)
	r.Get("/v3/entries.json", nightscout.NSGetEntries)
	r.Post("/v3/entries", nightscout.NSCreateEntry)
	r.Post("/v3/entries.json", nightscout.NSCreateEntry)

	// Treatments

	// ---- TREATMENTS ----
	r.Get("/v1/treatments", nightscout.NSGetTreatments)
	r.Get("/v1/treatments.json", nightscout.NSGetTreatments)
	r.Post("/v1/treatments", nightscout.NSCreateTreatment)
	r.Post("/v1/treatments.json", nightscout.NSCreateTreatment)
	// v3
	r.Get("/v3/treatments", nightscout.NSGetTreatments)
	r.Get("/v3/treatments.json", nightscout.NSGetTreatments)
	r.Post("/v3/treatments", nightscout.NSCreateTreatment)
	r.Post("/v3/treatments.json", nightscout.NSCreateTreatment)

	// ---- DEVICE STATUS ----
	r.Get("/v1/devicestatus", nightscout.NSGetDeviceStatus)
	r.Get("/v1/devicestatus.json", nightscout.NSGetDeviceStatus)
	r.Post("/v1/devicestatus", nightscout.NSCreateDeviceStatus)
	r.Post("/v1/devicestatus.json", nightscout.NSCreateDeviceStatus)

	// ---- PROFILE ----
	r.Get("/v1/profile", nightscout.NSGetProfiles)
	r.Get("/v1/profile.json", nightscout.NSGetProfiles)
	r.Post("/v1/profile", nightscout.NSCreateProfile)
	r.Post("/v1/profile.json", nightscout.NSCreateProfile)

	r.Get("/v1/profile/current.json", nightscout.NSGetCurrentProfile)
	r.Get("/v1/profile/current", nightscout.NSGetCurrentProfile)

	// ---- STATUS ----
	r.Get("/v1/status", nightscout.NSGetStatus)
	r.Get("/v1/status.json", nightscout.NSGetStatus)

	// ---- PROPERTIES ----
	r.Get("/v2/properties", nightscout.NSGetProperties)
	r.Get("/v2/properties.json", nightscout.NSGetProperties)
	return r
}
