package siteapi

import (
	"github.com/go-chi/chi/v5"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/server/assets/static"
	"nighttroop.t1d.nz/server/publicapi"
	"nighttroop.t1d.nz/server/siteapi/siteapiutils"
	"nighttroop.t1d.nz/server/siteapi/sitecollections"
	"nighttroop.t1d.nz/server/siteapi/sitelogin"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/httputils"
)

func NewSiteRouter(controlDB *db.ControlDB, config *config.ServerConfig) chi.Router {
	r := chi.NewRouter()
	httputils.AddStandardMiddleware(r, config, controlDB)

	// Static files are embedded and cheap, so like the public and control sites only the global limits cover them
	static.AddStaticRoutes(r)

	store := sitecollections.NewSiteStore()

	// Publically available (if the person knows the subdomain) endpoints
	r.Group(func(r chi.Router) {
		// Before sign in only the IP is known, so an IP spamming logins cannot spend the site's own limit
		r.Use(httputils.SiteUnauthRateLimitMiddleware(config))
		r.Get("/login", sitelogin.GETLoginPageHandler())
		publicapi.LegalRoutes(r)
		r.With(httputils.AllowedCountriesMiddleware(config)).Post("/login", sitelogin.POSTLogin(controlDB, config))
		r.With(httputils.AllowedCountriesMiddleware(config)).Post("/login/totp", sitelogin.POSTLoginTOTP(controlDB, config))
		// The backup key proves ownership, giving the site's admin new credentials
		r.With(httputils.AllowedCountriesMiddleware(config)).Post("/recover", sitelogin.POSTRecover(controlDB, config))
	})

	r.Group(func(r chi.Router) {
		r.Use(siteapiutils.JWTAuthMiddleware(controlDB, config)) // Checks for JWT Login and Token to Auth
		// Checks read and write rates and makes sure not over limit, once the site's own identity has signed in
		r.Use(httputils.SiteRateLimitMiddleware(config))
		r.Use(httputils.SiteIdentityRateLimitMiddleware(config))
		r.Use(httputils.SiteIdentitySeenFlagsUpdateMiddleware(controlDB)) // Updates user/identity flags
		r.Use(httputils.CountRoutes(config.Usage, "site"))

		// The summary is the home page
		r.Get("/", store.Collection("data", "summary").Search)
		r.Get("/summary", store.Collection("data", "summary").Search)
		r.Get("/getting-started", sitecollections.GettingStartedHandler)

		// The settings and alarms pages show the identity's own settings, and the site's to an admin. Both save to
		// /settings/site and /settings/identity.
		r.Get("/settings", store.Collection("settings", "identity").Get)
		r.Get("/settings/alarms", store.Collection("settings", "alarms").Get)
		for _, route := range []string{"site", "identity"} {
			r.Get("/settings/"+route, store.Collection("settings", route).Get)
			r.Post("/settings/"+route, store.Collection("settings", route).Update)
		}

		// The site's collections, each at its route in its section: data, settings, presets or logs
		for _, section := range []string{"data", "settings", "presets", "logs"} {
			routes := store.Section(section)
			r.Get("/"+section+"/{collection}", routes.Search)
			r.Get("/"+section+"/{collection}/events", routes.Events)
			r.Get("/"+section+"/{collection}/new", routes.New)
			r.Get("/"+section+"/{collection}/{id}", routes.Get)
			r.Post("/"+section+"/{collection}", routes.Create)
			r.Post("/"+section+"/{collection}/{id}", routes.Update)
			r.Delete("/"+section+"/{collection}/{id}", routes.Delete)
		}

		// Actions
		r.Get("/data/event-groups/choices", sitecollections.EventGroupChoicesHandler)
		r.Get("/data/event-groups/new", sitecollections.EventGroupNewHandler)
		r.Post("/settings/sync/{id}/backfill", sitecollections.BackfillHandler)
		r.Get("/settings/push-devices", sitecollections.PushDevicesHandler)
		r.Post("/settings/push-registrations/{id}/test", sitecollections.PushTestHandler)
		r.Post("/settings/identities/{id}/password", sitecollections.ResetIdentityPasswordHandler)
		r.Post("/settings/identities/{id}/totp", sitecollections.ResetIdentityTOTPHandler)
		r.Get("/settings/site/download", sitecollections.DownloadDataHandler)
		r.Post("/settings/site/delete", sitecollections.DeleteSiteHandler)
		r.Post("/logout", sitelogin.POSTLogoutHandler(config))
	})

	return r
}
