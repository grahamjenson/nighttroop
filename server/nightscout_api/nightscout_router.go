package nightscout_api

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/server/assets/static"
	v1 "nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/api/v1"
	v3 "nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/api/v3"
	vendorConfig "nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/config"
	"nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/deps"
	"nighttroop.t1d.nz/server/user"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/httputils"
)

func NewNightscoutRouter(controlDB *db.ControlDB, config config.ServerConfig) chi.Router {
	r := chi.NewRouter()
	httputils.AddStandardMiddleware(r, config)
	r.Use(user.UserRateLimitMiddleware(config))
	static.AddSatticRoutes(r)

	// Only Unauthed Endpoints are the ones to generate the JWT token
	// Can pass token asa query param, or JWT in auth header (recommended)
	r.Get("/api/v2/authorization/request", NSGetNightscoutJWTToken)
	r.Post("/api/v2/authorization/request", NSGetNightscoutJWTToken)

	ntDeps := deps.Dependencies{
		Config: vendorConfig.Config{}.WithDefaults(),
		Store: &NightTroopAPIStore{
			controlDB: controlDB,
			config:    config,
		},
		Auth: &NightTroopAPIAuth{
			controlDB: controlDB,
			config:    config,
		},
	}

	v1Router := v1.NewNightscoutV1Router(ntDeps)
	v3Router := v3.NewNightscoutV3Router(ntDeps)

	r.Mount("/api/v1", v1Router)
	r.Mount("/api/v3", v3Router)

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		slog.Info("Nightscout router 404 for path", "path", r.URL.Path)
		http.Error(w, "Not Found", http.StatusNotFound)
	})

	return r
}
