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
	publicapi.FooterRoutes(r, config)

	r.With(httputils.AllowedCountriesMiddleware(config)).Post("/login", controlapi.POSTLogin(controlDB, config))
	// A control admin's passkey signs it in on its own; it adds them on its Account page
	r.With(httputils.AllowedCountriesMiddleware(config)).Post("/login/passkey/begin", controlapi.POSTPasskeyLoginBegin(config))
	r.With(httputils.AllowedCountriesMiddleware(config)).Post("/login/passkey/finish", controlapi.POSTPasskeyLoginFinish(controlDB, config))

	r.Group(func(r chi.Router) {
		r.Use(controlapi.JWTAuthMiddleware(controlDB, config))
		r.Use(httputils.CountRoutes(config.Usage, "control"))
		r.Get("/", store.Collection("sites").Search) // The dashboard is the list of sites
		r.Post("/logout", controlapi.POSTLogoutHandler(config))

		// The control plane's collections, each at its route: sites, jobs, job stats and sync sources
		store.Routes(r)
		r.Post("/sites", (&controlcollections.SitesCollection{}).CreateHandler)
		r.Post("/settings", (&controlcollections.SitesCollection{}).SettingsHandler)
		r.Post("/sites/{id}/truncate", (&controlcollections.SitesCollection{}).TruncateHandler)
		r.Get("/sites/{id}/identities", (&controlcollections.SitesCollection{}).IdentitiesHandler)
		r.Post("/sites/{id}/identities/{identity}/password", (&controlcollections.SitesCollection{}).ResetPasswordHandler)
		r.Post("/sites/{id}/identities/{identity}/totp", (&controlcollections.SitesCollection{}).ResetTOTPHandler)
		r.Post("/sites/{id}/identities/{identity}/assume", (&controlcollections.SitesCollection{}).AssumeHandler(config))
		r.Get("/account", controlapi.GETAccount(controlDB))
		r.Post("/account/passkeys/begin", controlapi.POSTAccountPasskeyBegin(controlDB, config))
		r.Post("/account/passkeys/finish", controlapi.POSTAccountPasskeyFinish(controlDB, config))
		r.Delete("/account/passkeys/{id}", controlapi.DELETEAccountPasskey(controlDB))
		r.Get("/usage", controlapi.UsageHandler)
		r.Get("/rate-limits", controlapi.RateLimitsHandler)
		r.Post("/rate-limits", controlapi.SetRateLimitsHandler)
		r.Get("/push-devices", controlapi.PushDevicesHandler)
		r.Get(httputils.ControlLogsDBPath, controlapi.LogsDBHandler)
		r.Post("/push-devices/{id}/test", controlapi.PushTestHandler)

		// Not here yet: a site's API use per device and DB table sizes, and managing the control admins themselves,
		// listing them, resetting another's password or TOTP, changing their role or deleting one (never the one
		// signed in)
	})
	return r
}
