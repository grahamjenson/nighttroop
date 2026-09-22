package server

// User router enables user specific router, and can route to NS router if endpoint hits that layer.

import (
	"nighttroop.t1d.nz/server/assets/static"
	"nighttroop.t1d.nz/server/site"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/consts"
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

		r.Get("/", site.GetDashboardHandler(controlDB, config))

		r.Get("/sync_settings", site.GetSyncSettingsHandler(controlDB, config))
		r.Post("/sync_settings/nightscout", site.POSTSyncSettingsHandler(controlDB, config, consts.SyncSourceTypeNightscout))
		r.Post("/sync_settings/dexcom", site.POSTSyncSettingsHandler(controlDB, config, consts.SyncSourceTypeDexcom))
		r.Delete("/sync_settings/{type}", site.DELETESyncSettingsHandler(controlDB, config))

		// Identities
		r.Get("/identities", site.GETIdentitiesHandler(controlDB, config))
		r.Post("/identities", site.POSTIdentityHandler(controlDB, config))

		// Actions
		r.Post("/logout", site.POSTLogoutHandler(config))

	})

	return r
}
