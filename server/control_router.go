package server

import (
	"github.com/go-chi/chi/v5"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/server/assets/static"

	"nighttroop.t1d.nz/server/controlapi"
	"nighttroop.t1d.nz/server/controlapi/controlapiutils"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/httputils"
)

func NewControlRouter(controlDB *db.ControlDB, config *config.ServerConfig) chi.Router {
	r := chi.NewRouter()
	sitesStore := &controlapi.SitesStore{}
	syncSourcesStore := &controlapi.SyncSourcesStore{}
	httputils.AddStandardMiddleware(r, config)

	static.AddStaticRoutes(r)

	r.Get("/login", controlapi.GETLoginPageHandler())

	r.Post("/login", controlapi.POSTLogin(controlDB, config))

	r.Group(func(r chi.Router) {
		r.Use(controlapi.JWTAuthMiddleware(controlDB, config))
		r.Get("/", controlapi.GetDashboardHandler(controlDB, config))
		r.Post("/logout", controlapi.POSTLogoutHandler(config))

		// Sites
		r.Get("/sites", controlapiutils.SearchHandler(controlDB, config, sitesStore))
		r.Get("/sites/{id}", controlapiutils.GetHandler(controlDB, config, sitesStore))
		r.Post("/sites/{id}", controlapiutils.UpdateHandler(controlDB, config, sitesStore))

		// Sync Sources
		r.Get("/sync-sources", controlapiutils.SearchHandler(controlDB, config, syncSourcesStore))
		r.Get("/sync-sources/{id}", controlapiutils.GetHandler(controlDB, config, syncSourcesStore))

		//// TODO
		// Get User details including db size , last seen, API usage per device, list devices and last seen at, list DB table sizes
		// Get Admin Users
		// Reset Admin User Password (cannot do currently logged in user)
		// Reset Admin User TOTP (cannot do currently logged in user)
		// Change role of user (admin, user, read only)
		// Delete User (cannot delete currently logged in user)
	})
	return r
}
