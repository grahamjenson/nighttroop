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
	r.Use(httputils.CountRoutes(config.Usage, "public"))

	static.AddStaticRoutes(r)

	r.Get("/", publicapi.GETMainPageHandler(controlDB))
	r.Get("/health", publicapi.GETHealthHandler(controlDB))
	publicapi.LegalRoutes(r)

	r.Post("/site", publicapi.POSTCreateSiteHandler(controlDB, config))

	// TODO public routes
	// 1. Documentation
	// 2. SEO
	// TODO backupKeyhashFcuntions
	// 3. download sqlite backup
	return r
}
