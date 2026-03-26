package server

import (
	"github.com/go-chi/chi/v5"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/server/assets/static"
	"nighttroop.t1d.nz/server/control"
	"nighttroop.t1d.nz/utils/config"
)

func NewControlRouter(controlDB *db.ControlDB, config config.ServerConfig) chi.Router {
	r := chi.NewRouter()
	static.AddSatticRoutes(r)

	r.Get("/login", control.GETLoginPageHandler())

	r.Post("/login", control.POSTLogin(controlDB, config))

	r.Group(func(r chi.Router) {
		r.Use(control.JWTAuthMiddleware(controlDB, config))
		r.Get("/", control.GetDashboardHandler(controlDB, config))
		r.Get("/logout", control.GETLogoutHandler(config))
	})
	return r
}
