package server

// User router enables user specific router, and can route to NS router if endpoint hits that layer.

import (
	"nighttroop.t1d.nz/server/assets/static"
	"nighttroop.t1d.nz/server/site"
	"nighttroop.t1d.nz/server/site/settings"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/httputils"

	"github.com/go-chi/chi/v5"
	"nighttroop.t1d.nz/data/db"
)

func NewSiteRouter(controlDB *db.ControlDB, config *config.ServerConfig) chi.Router {
	r := chi.NewRouter()
	httputils.AddStandardMiddleware(r, config)

	// Rate Limiter is before auth because auth can be expencive
	r.Use(site.SiteRateLimitMiddleware(config))

	static.AddStaticRoutes(r)

	// Publically available (if the person knows the subdomain) endpoints
	r.Group(func(r chi.Router) {
		r.Get("/login", site.GETLoginPageHandler())
		r.Post("/login", site.POSTLogin(controlDB, config))
		r.Post("/login/totp", site.POSTLoginTOTP(controlDB, config))
	})

	r.Group(func(r chi.Router) {
		// Checks read and write rates and makes sure not over limit
		r.Use(site.JWTAuthMiddleware(controlDB, config)) // Checks for JWT Login and Token to Auth
		r.Use(site.SiteIdentityRateLimitMiddleware(config))
		r.Use(site.SiteIdentitySeenFlagsUpdateMiddleware(controlDB)) // Updates user/identity flags

		// Summary page
		r.Get("/", site.GetSummaryHandler(controlDB, config))
		r.Get("/summary", site.GetSummaryHandler(controlDB, config))

		// Settings Sync
		r.Get("/settings/sync", settings.GetSyncSettingsHandler(controlDB, config))
		r.Post("/settings/sync/{type}", settings.POSTSyncSettingsHandler(controlDB, config))
		r.Delete("/settings/sync/{type}", settings.DELETESyncSettingsHandler(controlDB, config))
		r.Get("/settings/sync/reports", settings.GetSyncSettingsReportsHandler(controlDB, config))

		// Settings Identities
		r.Get("/settings/identities", settings.GETIdentitiesHandler(controlDB, config))
		r.Get("/settings/identities/{id}", settings.GETIdentitiesHandler(controlDB, config))
		r.Post("/settings/identities", settings.POSTCreateIdentityHandler(controlDB, config))
		r.Post("/settings/identities/{id}", settings.POSTUpdateIdentityHandler(controlDB, config))
		r.Delete("/settings/identities/{id}", settings.DELETEIdentityHandler(controlDB, config))

		// Settings Settings
		r.Get("/settings/settings", settings.GetSyncSettingsHandler(controlDB, config))
		r.Post("/settings/settings", settings.GetSyncSettingsHandler(controlDB, config))

		// Actions
		r.Post("/logout", site.POSTLogoutHandler(config))

	})

	return r
}
