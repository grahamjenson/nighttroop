package server

// User router enables user specific router, and can route to NS router if endpoint hits that layer.

import (
	"nighttroop.t1d.nz/server/assets/static"
	"nighttroop.t1d.nz/server/user"
	"nighttroop.t1d.nz/utils/config"

	"github.com/go-chi/chi/v5"
	"nighttroop.t1d.nz/data/db"
)

func NewUserRouter(controlDB *db.ControlDB, config config.ServerConfig) chi.Router {
	r := chi.NewRouter()
	AddStandardMiddleware(r, config)

	r.Use(user.UserRateLimitMiddleware(config))

	static.AddSatticRoutes(r)
	// Rate Limiter is before auth because auth can be expencive

	// Publically availabl (if the person knows the subdomain) endpoints
	r.Get("/login", user.GETLoginPageHandler())
	r.Post("/login", user.POSTLogin(controlDB, config))
	r.Post("/login_totp", user.POSTLoginTOTP(controlDB, config))

	r.Group(func(r chi.Router) {
		// Checks read and write rates and makes sure not over limit
		r.Use(user.JWTAuthMiddleware(controlDB, config)) // Checks for JWT Login and Token to Auth
		r.Use(user.UserDBMiddleware(controlDB))          // Adds User DB and model to the context

		r.Get("/", user.GetDashboardHandler(controlDB, config))

		r.Get("/sync-settings", user.GetSyncSettingsHandler(controlDB, config))
		r.Post("/sync-settings", user.POSTSyncSettingsHandler(controlDB, config))

		// Actions
		r.Post("/logout", user.POSTLogoutHandler(config))

		//// TODO
		// Get Devices probably with pagination returns htmx, include generated tokens and last seen
		// Delete Device
		// Change Device Passowrd, role TOTP
		// Get Device Token

	})

	return r
}
