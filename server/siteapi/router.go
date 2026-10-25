package siteapi

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/server/assets/static"
	"nighttroop.t1d.nz/server/siteapi/siteapiutils"
	"nighttroop.t1d.nz/server/siteapi/sitelogin"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/httputils"
)

func NewSiteRouter(controlDB *db.ControlDB, config *config.ServerConfig) chi.Router {
	//STORES

	r := chi.NewRouter()
	httputils.AddStandardMiddleware(r, config)

	// Rate Limiter is before auth because auth can be expencive
	r.Use(httputils.SiteRateLimitMiddleware(config))

	static.AddStaticRoutes(r)

	store := sitecollections.NewStore()

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

		// Redirect home to Summary page
		r.Get("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/v1/summary", http.StatusSeeOther)
		}))

		// Data Glucose Readings
		r.Get("/data/{collection}", SearchHandler(controlDB, config, store))
		r.Get("/data/{collection}/{id}", GetHandler(controlDB, config, store))
		r.Post("/data/{collection}", CreateHandler(controlDB, config, store))
		r.Post("/data/{collection}/{id}", UpdateHandler(controlDB, config, store))
		r.Delete("/data/{collection}/{id}", DeleteHandler(controlDB, config, store))

		// Actions
		r.Post("/logout", sitelogin.POSTLogoutHandler(config))
	})

	return r
}
