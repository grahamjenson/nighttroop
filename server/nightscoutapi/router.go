package nightscoutapi

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"nighttroop.t1d.nz/data/db"
	v1 "nighttroop.t1d.nz/server/nightscoutapi/api/v1"
	v2 "nighttroop.t1d.nz/server/nightscoutapi/api/v2"
	v3 "nighttroop.t1d.nz/server/nightscoutapi/api/v3"
	"nighttroop.t1d.nz/server/nightscoutapi/nsutils"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/httputils"
)

func NewNightscoutRouter(controlDB *db.ControlDB, config *config.ServerConfig) chi.Router {
	websocket := NewWebsocket(controlDB, config)
	ntDeps := nsutils.Dependencies{
		Store: NewNightTroopAPIStore(controlDB, config),

		Auth: &NightTroopAPIAuth{
			controlDB: controlDB,
			config:    config,
		},
	}

	r := chi.NewRouter()
	httputils.AddStandardMiddleware(r, config)
	// Before auth only the IP is known, so an IP spamming bad tokens cannot spend the site's own limit
	r.Use(httputils.SiteUnauthRateLimitMiddleware(config))

	// Socket.IO clients authorize with an event, not on each request
	r.Handle("/socket.io/", websocket)

	// make sure to check auth for JWT/token param/token URL param

	r.Group(func(r chi.Router) {
		r.Use(NightscoutAuthMiddleware(controlDB, config))
		r.Use(httputils.SiteRateLimitMiddleware(config))
		r.Use(httputils.SiteIdentityRateLimitMiddleware(config))
		r.Use(httputils.SiteIdentitySeenFlagsUpdateMiddleware(controlDB)) // Updates user/identity flags

		r.Get("/api/v2/authorization/request/token={token}", NightscoutAuthRoute(controlDB, config))
		r.Get("/api/v2/authorization/request/{token}", NightscoutAuthRoute(controlDB, config))
		r.Get("/api/v3/authorization/request/token={token}", NightscoutAuthRoute(controlDB, config))
		r.Get("/api/v3/authorization/request/{token}", NightscoutAuthRoute(controlDB, config))

		v1Router := v1.NewNightscoutV1Router(ntDeps)
		v2Router := v2.NewNightscoutV2Router(ntDeps)
		v3Router := v3.NewNightscoutV3Router(ntDeps)

		r.Mount("/api/v1", v1Router)
		r.Mount("/api/v2", v2Router)
		r.Mount("/api/v3", v3Router)

		r.NotFound(func(w http.ResponseWriter, r *http.Request) {
			slog.Info("Nightscout router 404 for path", "path", r.URL.Path)
			http.Error(w, "Not Found", http.StatusNotFound)
		})
	})
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		slog.Info("Nightscout router 404 for path", "path", r.URL.Path)
		http.Error(w, "Not Found", http.StatusNotFound)
	})
	return r
}

func NightscoutAuthRoute(controlDB *db.ControlDB, config *config.ServerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		siteContext := httputils.GetSiteContext(r.Context())
		if siteContext.IdentityModel == nil {
			http.Error(w, "Unauthorized: No valid authentication method found", http.StatusUnauthorized)
			return
		}

		// The user should have already been verified
		token, expiresAt, err := config.Security.GenerateUserJWTToken(siteContext.Subdomain, siteContext.IdentityModel.Name, string(siteContext.IdentityModel.Role))
		if err != nil {
			http.Error(w, "Internal Server Error: Failed to generate JWT token", http.StatusInternalServerError)
			return
		}

		nsIdentity, err := NighttroopIdentityToNightscoutIdentity(siteContext.IdentityModel.Name, siteContext.IdentityModel.Role)
		if err != nil {
			http.Error(w, "Internal Server Error: Failed to convert identity", http.StatusInternalServerError)
			return
		}
		response := map[string]any{
			"token":            token,
			"sub":              siteContext.IdentityModel.Name,
			"permissionGroups": [][]string{nsIdentity.Permissions},
			"iat":              time.Now().Unix(),
			"exp":              expiresAt.Unix(),
		}

		jsonResponse, err := json.Marshal(response)
		if err != nil {
			http.Error(w, "Internal Server Error: Failed to marshal JSON", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonResponse)
	}

}
