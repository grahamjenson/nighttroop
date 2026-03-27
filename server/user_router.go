package server

// User router enables user specific router, and can route to NS router if endpoint hits that layer.

import (
	"nighttroop.t1d.nz/server/assets/static"
	"nighttroop.t1d.nz/server/control"
	"nighttroop.t1d.nz/server/user"
	"nighttroop.t1d.nz/utils/config"

	"github.com/go-chi/chi/v5"
	"nighttroop.t1d.nz/data/db"
)

func NewUserRouter(controlDB *db.ControlDB, config config.ServerConfig) chi.Router {
	r := chi.NewRouter()
	r.Use(user.GetUserMiddleware(controlDB))

	static.AddSatticRoutes(r)

	r.Get("/login", user.GETLoginPageHandler())
	r.Post("/login", user.POSTLogin(controlDB, config))

	r.Group(func(r chi.Router) {
		r.Use(control.JWTAuthMiddleware(controlDB, config))
		r.Get("/", control.GetDashboardHandler(controlDB, config))
		r.Get("/logout", control.GETLogoutHandler(config))
	})

	r.Get("/ping", user.GETPingHandler())
	return r
}
