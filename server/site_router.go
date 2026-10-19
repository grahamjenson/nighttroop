package server

// User router enables user specific router, and can route to NS router if endpoint hits that layer.

import (
	"nighttroop.t1d.nz/server/assets/static"
	"nighttroop.t1d.nz/server/siteapi/siteapiutils"
	"nighttroop.t1d.nz/server/siteapi/sitelogin"
	"nighttroop.t1d.nz/server/siteapi/sitesettings"
	"nighttroop.t1d.nz/server/siteapi/sitesummary"

	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/httputils"

	"github.com/go-chi/chi/v5"
	"nighttroop.t1d.nz/data/db"
)

func NewSiteRouter(controlDB *db.ControlDB, config *config.ServerConfig) chi.Router {
	//STORES
	syncSettingsStore := &sitesettings.SyncSettingsStore{}

	r := chi.NewRouter()
	httputils.AddStandardMiddleware(r, config)

	// Rate Limiter is before auth because auth can be expencive
	r.Use(httputils.SiteRateLimitMiddleware(config))

	static.AddStaticRoutes(r)

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

		// Summary page
		r.Get("/", sitesummary.GetSummaryHandler(controlDB, config))
		r.Get("/summary", sitesummary.GetSummaryHandler(controlDB, config))

		// Settings Sync
		r.Get("/settings/sync", siteapiutils.SearchHandler(controlDB, config, syncSettingsStore))
		r.Post("/settings/sync/{id}", siteapiutils.UpsertHandler(controlDB, config, syncSettingsStore))
		r.Delete("/settings/sync/{id}", siteapiutils.DeleteHandler(controlDB, config, syncSettingsStore))

		// Settings Identities
		r.Get("/settings/identities", sitesettings.GETIdentitiesHandler(controlDB, config))
		r.Get("/settings/identities/{id}", sitesettings.GETIdentitiesHandler(controlDB, config))
		r.Post("/settings/identities", sitesettings.POSTCreateIdentityHandler(controlDB, config))
		r.Post("/settings/identities/{id}", sitesettings.POSTUpdateIdentityHandler(controlDB, config))
		r.Delete("/settings/identities/{id}", sitesettings.DELETEIdentityHandler(controlDB, config))

		// Settings Settings
		r.Get("/settings/config", sitesettings.GETConfigSettingsHandler(controlDB, config))
		r.Post("/settings/config", sitesettings.POSTConfigSettingsHandler(controlDB, config))

		// Actions
		r.Post("/logout", sitelogin.POSTLogoutHandler(config))

	})

	return r
}
