package server

// User router enables user specific router, and can route to NS router if endpoint hits that layer.

import (
	"nighttroop.t1d.nz/models/sitemodels"
	"nighttroop.t1d.nz/server/assets/static"
	"nighttroop.t1d.nz/server/siteapi"
	"nighttroop.t1d.nz/server/siteapi/settings"
	"nighttroop.t1d.nz/server/siteapi/siteutils"

	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/httputils"

	"github.com/go-chi/chi/v5"
	"nighttroop.t1d.nz/data/db"
)

func NewSiteRouter(controlDB *db.ControlDB, config *config.ServerConfig) chi.Router {
	//STORES
	syncSettingsStore := &settings.SyncSettingsStore{}

	r := chi.NewRouter()
	httputils.AddStandardMiddleware(r, config)

	// Rate Limiter is before auth because auth can be expencive
	r.Use(siteapi.SiteRateLimitMiddleware(config))

	static.AddStaticRoutes(r)

	// Publically available (if the person knows the subdomain) endpoints
	r.Group(func(r chi.Router) {
		r.Get("/login", siteapi.GETLoginPageHandler())
		r.Post("/login", siteapi.POSTLogin(controlDB, config))
		r.Post("/login/totp", siteapi.POSTLoginTOTP(controlDB, config))
	})

	r.Group(func(r chi.Router) {
		// Checks read and write rates and makes sure not over limit
		r.Use(siteapi.JWTAuthMiddleware(controlDB, config)) // Checks for JWT Login and Token to Auth
		r.Use(siteapi.SiteIdentityRateLimitMiddleware(config))
		r.Use(siteapi.SiteIdentitySeenFlagsUpdateMiddleware(controlDB)) // Updates user/identity flags

		// Summary page
		r.Get("/", siteapi.GetSummaryHandler(controlDB, config))
		r.Get("/summary", siteapi.GetSummaryHandler(controlDB, config))

		// Settings Sync
		r.Get("/settings/sync", settings.GetSyncSettingsHandler(controlDB, config))
		r.Post("/settings/sync/{id}", siteutils.CreateUpdateHandler[sitemodels.SyncSource](controlDB, config, syncSettingsStore))
		r.Delete("/settings/sync/{id}", settings.DELETESyncSettingsHandler(controlDB, config))
		r.Get("/settings/sync/reports", settings.GetSyncSettingsReportsHandler(controlDB, config))

		// Settings Identities
		r.Get("/settings/identities", settings.GETIdentitiesHandler(controlDB, config))
		r.Get("/settings/identities/{id}", settings.GETIdentitiesHandler(controlDB, config))
		r.Post("/settings/identities", settings.POSTCreateIdentityHandler(controlDB, config))
		r.Post("/settings/identities/{id}", settings.POSTUpdateIdentityHandler(controlDB, config))
		r.Delete("/settings/identities/{id}", settings.DELETEIdentityHandler(controlDB, config))

		// Settings Settings
		r.Get("/settings/config", settings.GETConfigSettingsHandler(controlDB, config))
		r.Post("/settings/config", settings.POSTConfigSettingsHandler(controlDB, config))

		// Actions
		r.Post("/logout", siteapi.POSTLogoutHandler(config))

	})

	return r
}
