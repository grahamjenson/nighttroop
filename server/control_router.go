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
	AddStandardMiddleware(r, config)

	static.AddSatticRoutes(r)

	r.Get("/login", control.GETLoginPageHandler())

	r.Post("/login", control.POSTLogin(controlDB, config))

	r.Group(func(r chi.Router) {
		r.Use(control.JWTAuthMiddleware(controlDB, config))
		r.Get("/", control.GetDashboardHandler(controlDB, config))
		r.Post("/logout", control.POSTLogoutHandler(config))

		//// TODO
		// Get Users probably with pagination returns htmx
		// Get User details including db size , last seen, API usage per device, list devices and last seen at, list DB table sizes
		// Get Admin Users
		// Reset Admin User Password (cannot do currently logged in user)
		// Reset Admin User TOTP (cannot do currently logged in user)
		// Change role of user (admin, user, read only)
		// Delete User (cannot delete currently logged in user)
	})
	return r
}
