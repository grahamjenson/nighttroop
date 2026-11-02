package nightscoutapi

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/server/nightscoutapi/nsapi/nsv1"
	"nighttroop.t1d.nz/server/nightscoutapi/nsapi/nsv2"
	"nighttroop.t1d.nz/server/nightscoutapi/nsapi/nsv3"
	"nighttroop.t1d.nz/server/nightscoutapi/nsapi/nswebsockets"
	"nighttroop.t1d.nz/server/nightscoutapi/nsutils"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/httputils"
)

func NewNightscoutRouter(controlDB *db.ControlDB, config *config.ServerConfig) chi.Router {
	websocket := nswebsockets.NewWebsocket(controlDB, config, AllCollections(), jwtIdentity, legacyTokenIdentity)
	store := NewNightTroopAPIStore(controlDB, config)

	r := chi.NewRouter()
	httputils.AddStandardMiddleware(r, config, controlDB)
	// Before auth only the IP is known, so an IP spamming bad tokens cannot spend the site's own limit
	r.Use(httputils.SiteUnauthRateLimitMiddleware(config))

	// Socket.IO clients authorize with an event, not on each request
	r.Handle("/socket.io/", websocket)
	// As in Nightscout, v3's version needs no token: Nightguard asks it without one to see if v3 is there
	r.Get("/api/v3/version", nsv3.Version(store))

	// make sure to check auth for JWT/token param/token URL param

	r.Group(func(r chi.Router) {
		r.Use(NightscoutAuthMiddleware(controlDB, config))
		r.Use(httputils.SiteRateLimitMiddleware(config))
		r.Use(httputils.SiteIdentityRateLimitMiddleware(config))
		r.Use(httputils.SiteIdentitySeenFlagsUpdateMiddleware(controlDB)) // Updates user/identity flags
		r.Use(httputils.CountRoutes(config.Usage, "nightscout"))

		r.Get("/api/v2/authorization/request/token={token}", NightscoutAuthRoute(controlDB, config))
		r.Get("/api/v2/authorization/request/{token}", NightscoutAuthRoute(controlDB, config))
		r.Get("/api/v3/authorization/request/token={token}", NightscoutAuthRoute(controlDB, config))
		r.Get("/api/v3/authorization/request/{token}", NightscoutAuthRoute(controlDB, config))

		v1Router := nsv1.NewNightscoutV1Router(store)
		v2Router := nsv2.NewNightscoutV2Router(store)
		v3Router := nsv3.NewNightscoutV3Router(store)

		r.Mount("/api/v1", v1Router)
		r.Mount("/api/v2", v2Router)
		r.Mount("/api/v3", v3Router)

		r.NotFound(func(w http.ResponseWriter, r *http.Request) {
			slog.InfoContext(r.Context(), "Nightscout router 404 for path", "path", r.URL.Path)
			http.Error(w, "Not Found", http.StatusNotFound)
		})
	})
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		slog.InfoContext(r.Context(), "Nightscout router 404 for path", "path", r.URL.Path)
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
		token, expiresAt, err := config.Security.GenerateUserJWTToken(siteContext.SiteID, siteContext.IdentityModel.Name, string(siteContext.IdentityModel.Role))
		if err != nil {
			http.Error(w, "Internal Server Error: Failed to generate JWT token", http.StatusInternalServerError)
			return
		}

		// An identity with no Nightscout role, such as an admin, cannot use the Nightscout API
		nsIdentity, err := nsutils.NighttroopIdentityToNightscoutIdentity(siteContext.IdentityModel.Name, siteContext.IdentityModel.Role)
		if err != nil {
			http.Error(w, "Forbidden: this identity cannot use the Nightscout API", http.StatusForbidden)
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
