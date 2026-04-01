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
	static.AddSatticRoutes(r)
	// Rate Limiter is before auth because auth can be expencive
	r.Use(user.UserRateLimitMiddleware(config))

	// Publically availabl (if the person knows the subdomain) endpoints
	r.Get("/login", user.GETLoginPageHandler())
	r.Post("/login", user.POSTLogin(controlDB, config))
	r.Post("/login_totp", user.POSTLoginTOTP(controlDB, config))

	r.Group(func(r chi.Router) {
		// Checks read and write rates and makes sure not over limit
		r.Use(user.JWTPlusTokenAuthMiddleware(controlDB, config)) // Checks for JWT Login and Token to Auth
		r.Use(user.UserDBMiddleware(controlDB))                   // Adds User DB and model to the context

		r.Get("/", user.GetDashboardHandler(controlDB, config))
		r.Get("/logout", user.GETLogoutHandler(config))
	})

	r.Get("/ping", user.GETPingHandler())
	return r
}
