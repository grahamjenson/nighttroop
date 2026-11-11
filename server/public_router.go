package server

import (
	"github.com/go-chi/chi/v5"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/server/assets/static"
	"nighttroop.t1d.nz/server/publicapi"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/httputils"
)

func NewPublicRouter(controlDB *db.ControlDB, config *config.ServerConfig) chi.Router {
	r := chi.NewRouter()
	httputils.AddStandardMiddleware(r, config, controlDB)

	static.AddStaticRoutes(r)

	r.Get("/", publicapi.GETMainPageHandler(controlDB))
	r.Get("/health", publicapi.GETHealthHandler(controlDB))
	publicapi.FooterRoutes(r, config)

	// Creating a site, signing in and recovering are limited to the server's ALLOWED_COUNTRIES
	r.With(httputils.AllowedCountriesMiddleware(config)).Post("/site", publicapi.POSTCreateSiteHandler(controlDB, config))
	r.Post("/site/totp", publicapi.POSTCheckTOTPHandler(config))
	r.Get("/recover", publicapi.GETRecoverSitePageHandler())
	r.With(httputils.AllowedCountriesMiddleware(config)).Post("/recover", publicapi.POSTRecoverSiteHandler(controlDB, config))

	// Not served here yet: documentation pages, and pages for search engines beyond robots.txt
	return r
}
