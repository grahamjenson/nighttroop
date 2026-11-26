package server

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
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

		// The control plane's collections, each at its route: sites, jobs, job stats and connectors
		store.Routes(r)
		// Only an admin can change the server, its sites, its push credentials and its admins, send a test push, or
		// download or upload a DB, not a read only identity
		admin := r.With(controlcollections.AdminOnly)
		admin.Post("/sites", (&controlcollections.SitesCollection{}).CreateHandler)
		admin.Post("/settings", (&controlcollections.SitesCollection{}).SettingsHandler)
		admin.Post("/sites/{id}/truncate", (&controlcollections.SitesCollection{}).TruncateHandler)
		admin.Post("/sites/{id}/new-empty-db", (&controlcollections.SitesCollection{}).NewEmptyDBHandler)
		admin.Post("/sites/{id}/restore-version", (&controlcollections.SitesCollection{}).RestoreVersionHandler)
		admin.Get("/sites/{id}/db", (&controlcollections.SitesCollection{}).DownloadDBHandler)
		admin.Post("/sites/{id}/db", (&controlcollections.SitesCollection{}).UploadDBHandler)
		r.Get("/sites/{id}/identities", (&controlcollections.SitesCollection{}).IdentitiesHandler)
		admin.Post("/sites/{id}/identities/{identity}/password", (&controlcollections.SitesCollection{}).ResetPasswordHandler)
		admin.Post("/sites/{id}/identities/{identity}/totp", (&controlcollections.SitesCollection{}).ResetTOTPHandler)
		admin.Post("/sites/{id}/identities/{identity}/assume", (&controlcollections.SitesCollection{}).AssumeHandler(config))
		r.Get("/account", controlapi.GETAccount(controlDB))
		r.Post("/account/passkeys/begin", controlapi.POSTAccountPasskeyBegin(controlDB, config))
		r.Post("/account/passkeys/finish", controlapi.POSTAccountPasskeyFinish(controlDB, config))
		r.Delete("/account/passkeys/{id}", controlapi.DELETEAccountPasskey(controlDB))
		r.Get("/usage", controlapi.UsageHandler)
		r.Get("/connectors/shared-hosts", (&controlcollections.ConnectorsCollection{}).SharedHostsHandler)
		r.Get("/rate-limits", controlapi.RateLimitsHandler)
		admin.Post("/rate-limits", controlapi.SetRateLimitsHandler)
		r.Get("/push-devices", controlapi.PushDevicesHandler)
		admin.Get(httputils.ControlLogsDBPath, controlapi.LogsDBHandler)
		admin.Post("/push-devices/{id}/test", controlapi.PushTestHandler)
		admin.Get("/push-credentials/new", (&controlcollections.PushCredentialsCollection{}).NewHandler)
		admin.Post("/push-credentials", (&controlcollections.PushCredentialsCollection{}).CreateHandler)
		admin.Delete("/push-credentials/{id}", (&controlcollections.PushCredentialsCollection{}).DeleteHandler)
		// Control's admins, which an admin manages, never itself
		// Messages sites show at the top of their pages, which an admin adds and deletes
		r.Get("/banners", controlapi.SiteBannersHandler)
		admin.Post("/banners", controlapi.CreateSiteBannerHandler)
		admin.Delete("/banners/{id}", controlapi.DeleteSiteBannerHandler)
		r.Get("/admins", controlapi.ControlAdminsHandler)
		admin.Post("/admins", controlapi.CreateControlAdminHandler)
		admin.Post("/admins/{id}/role", controlapi.SetControlAdminRoleHandler)
		admin.Post("/admins/{id}/password", controlapi.ResetControlAdminPasswordHandler)
		admin.Post("/admins/{id}/totp", controlapi.ResetControlAdminTOTPHandler)
		admin.Delete("/admins/{id}", controlapi.DeleteControlAdminHandler)
		admin.Post("/admins/test-page", controlapi.TestPageHandler)
		// Go's profiles (/debug/pprof/) and runtime stats (/debug/vars), to see where the server's memory and CPU go. A
		// CPU profile or trace must take under the 30 seconds a request has, e.g. ?seconds=20
		admin.Mount("/debug", middleware.Profiler())

		// Not here yet: a site's API use per device and DB table sizes
	})
	return r
}
