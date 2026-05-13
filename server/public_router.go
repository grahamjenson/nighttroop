package server

import (
	"github.com/go-chi/chi/v5"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/server/assets/static"
	"nighttroop.t1d.nz/server/public"
	"nighttroop.t1d.nz/utils/config"
)

func NewPublicRouter(controlDB *db.ControlDB, config config.ServerConfig) chi.Router {
	r := chi.NewRouter()
	AddStandardMiddleware(r, config)

	static.AddSatticRoutes(r)

	r.Get("/", public.GETMainPageHandler(controlDB))

	r.Post("/user", public.POSTCreateUserHandler(controlDB, config))

	// TODO public routes
	// 1. Documentation
	// 2. SEO
	// TODO backupKeyhashFcuntions
	// 1. reset admin passwrod
	// 2. reset admin totp
	// 3. download sqlite backup
	return r
}
