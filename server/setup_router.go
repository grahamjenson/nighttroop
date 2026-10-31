package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/server/assets/static"
	"nighttroop.t1d.nz/server/controlapi"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/httputils"
)

// NewSetupRouter is the control site of a server that is not set up yet: its setup page, and everything else sent there
func NewSetupRouter(controlDB *db.ControlDB, config *config.ServerConfig) chi.Router {
	r := chi.NewRouter()
	httputils.AddStandardMiddleware(r, config, controlDB)
	static.AddStaticRoutes(r)
	r.Get("/setup", controlapi.GETSetupPageHandler())
	r.Post("/setup", controlapi.POSTSetupHandler(controlDB, config))
	r.Post("/setup/confirm", controlapi.POSTSetupConfirmHandler(controlDB, config))
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/setup", http.StatusSeeOther)
	})
	return r
}

// serveSetup sends a server that is not set up to its setup page, leaving only the health check, and apps an error
func serveSetup(w http.ResponseWriter, r *http.Request, config *config.ServerConfig, routeTarget RouteTarget, setup, public http.Handler) {
	switch {
	case routeTarget == RouteTargetControl:
		setup.ServeHTTP(w, r)
	case routeTarget == RouteTargetPublic && r.URL.Path == "/health":
		public.ServeHTTP(w, r)
	case httputils.GetAcceptType(r) == httputils.AcceptTypeJSON:
		http.Error(w, "Nighttroop is not set up yet", http.StatusServiceUnavailable)
	default:
		http.Redirect(w, r, config.SetupURL(), http.StatusSeeOther)
	}
}
