package siteapi

import (
	"github.com/go-chi/chi/v5"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/server/assets/static"
	"nighttroop.t1d.nz/server/siteapi/siteapiutils"
	"nighttroop.t1d.nz/server/siteapi/sitecollections"
	"nighttroop.t1d.nz/server/siteapi/sitelogin"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/httputils"
)

func NewSiteRouter(controlDB *db.ControlDB, config *config.ServerConfig) chi.Router {
	r := chi.NewRouter()
	httputils.AddStandardMiddleware(r, config)

	// Rate Limiter is before auth because auth can be expencive
	r.Use(httputils.SiteRateLimitMiddleware(config))

	static.AddStaticRoutes(r)

	store := sitecollections.NewSiteStore()

	// Publically available (if the person knows the subdomain) endpoints
	r.Group(func(r chi.Router) {
		r.Get("/login", sitelogin.GETLoginPageHandler())
		r.Post("/login", sitelogin.POSTLogin(controlDB, config))
		r.Post("/login/totp", sitelogin.POSTLoginTOTP(controlDB, config))
	})

	r.Group(func(r chi.Router) {
		// Checks read and write rates and makes sure not over limit
		r.Use(siteapiutils.JWTAuthMiddleware(controlDB, config)) // Checks for JWT Login and Token to Auth
		r.Use(httputils.SiteIdentityRateLimitMiddleware(config))
		r.Use(httputils.SiteIdentitySeenFlagsUpdateMiddleware(controlDB)) // Updates user/identity flags

		// The summary is the home page
		r.Get("/", store.Collection("data", "summary").Search)
		r.Get("/summary", store.Collection("data", "summary").Search)

		// The settings page shows the identity's own settings, and the site's to an admin
		r.Get("/settings", store.Collection("settings", "identity").Get)
		for _, route := range []string{"site", "identity"} {
			r.Get("/settings/"+route, store.Collection("settings", route).Get)
			r.Post("/settings/"+route, store.Collection("settings", route).Update)
		}

		// The site's collections, each at its route in its section: data, settings or presets
		for _, section := range []string{"data", "settings", "presets"} {
			routes := store.Section(section)
			r.Get("/"+section+"/{collection}", routes.Search)
			r.Get("/"+section+"/{collection}/{id}", routes.Get)
			r.Post("/"+section+"/{collection}", routes.Create)
			r.Post("/"+section+"/{collection}/{id}", routes.Update)
			r.Delete("/"+section+"/{collection}/{id}", routes.Delete)
		}

		// Actions
		r.Post("/logout", sitelogin.POSTLogoutHandler(config))
	})

	return r
}
