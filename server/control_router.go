package server

import (
	"github.com/go-chi/chi/v5"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/server/assets/static"

	"nighttroop.t1d.nz/server/controlapi"
	"nighttroop.t1d.nz/server/controlapi/controlcollections"
	"nighttroop.t1d.nz/server/publicapi"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/httputils"
)

func NewControlRouter(controlDB *db.ControlDB, config *config.ServerConfig) chi.Router {
	r := chi.NewRouter()
	store := controlcollections.NewControlStore()
	httputils.AddStandardMiddleware(r, config, controlDB)

	static.AddStaticRoutes(r)

	r.Get("/login", controlapi.GETLoginPageHandler())
	publicapi.LegalRoutes(r)

	r.Post("/login", controlapi.POSTLogin(controlDB, config))

	r.Group(func(r chi.Router) {
		r.Use(controlapi.JWTAuthMiddleware(controlDB, config))
		r.Use(httputils.CountRoutes(config.Usage, "control"))
		r.Get("/", store.Collection("sites").Search) // The dashboard is the list of sites
		r.Post("/logout", controlapi.POSTLogoutHandler(config))

		// The control plane's collections, each at its route: sites, jobs, job stats and sync sources
		store.Routes(r)
		r.Post("/sites", (&controlcollections.SitesCollection{}).CreateHandler)
		r.Post("/sites/{id}/truncate", (&controlcollections.SitesCollection{}).TruncateHandler)
		r.Get("/sites/{id}/identities", (&controlcollections.SitesCollection{}).IdentitiesHandler)
		r.Post("/sites/{id}/identities/{identity}/password", (&controlcollections.SitesCollection{}).ResetPasswordHandler)
		r.Post("/sites/{id}/identities/{identity}/totp", (&controlcollections.SitesCollection{}).ResetTOTPHandler)
		r.Get("/usage", controlapi.UsageHandler)

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
