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

	// These are essentially public (if the person knows the subdomain) but are protected by Global and IP rate limiters in the main router
	r.Get("/login", user.GETLoginPageHandler())
	r.Post("/login", user.POSTLogin(controlDB, config))
	r.Post("/login_totp", user.POSTLoginTOTP(controlDB, config))

	r.Group(func(r chi.Router) {
		r.Use(user.JWTPlusTokenAuthMiddleware(controlDB, config))
		r.Use(user.UserRateLimitMiddleware(config))

		r.Get("/", user.GetDashboardHandler(controlDB, config))
		r.Get("/logout", user.GETLogoutHandler(config))
	})

	r.Get("/ping", user.GETPingHandler())
	return r
}
