package nightscout_api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"nighttroop.t1d.nz/data/db"
	v1 "nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/api/v1"
	v2 "nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/api/v2"
	v3 "nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/api/v3"
	"nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/di"
	"nighttroop.t1d.nz/server/user"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/httputils"
)

func NewNightscoutRouter(controlDB *db.ControlDB, config config.ServerConfig) chi.Router {
	r := chi.NewRouter()
	httputils.AddStandardMiddleware(r, config)
	// make sure to continue rate limiting
	r.Use(user.UserRateLimitMiddleware(config))
	// make sure to check auth for JWT/token param/token URL param
	r.Use(NightscoutAuthMiddleware(controlDB, config))
	r.Use(user.UserIdentitySeenFlagsUpdateMiddleware(controlDB)) // Updates user/identity flags

	r.Get("/api/v2/authorization/request/token={token}", NightscoutAuthRoute(controlDB, config))

	ntDeps := di.Dependencies{
		Store: &NightTroopAPIStore{
			controlDB: controlDB,
			config:    config,
		},
		Auth: &NightTroopAPIAuth{
			controlDB: controlDB,
			config:    config,
		},
	}

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

	return r
}

func NightscoutAuthRoute(controlDB *db.ControlDB, config config.ServerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userContext := httputils.GetUserContext(r.Context())
		if userContext.IdentityModel == nil {
			http.Error(w, "Unauthorized: No valid authentication method found", http.StatusUnauthorized)
			return
		}

		// The user should have already been verified
		token, expiresAt, err := config.Security.GenerateUserJWTToken(userContext.Subdomain, userContext.IdentityModel.Name, string(userContext.IdentityModel.Role))
		if err != nil {
			http.Error(w, "Internal Server Error: Failed to generate JWT token", http.StatusInternalServerError)
			return
		}

		nsIdentity := NighttroopIdentityToNightscoutIdentity(userContext.IdentityModel.Name, userContext.IdentityModel.Role)
		response := map[string]any{
			"token":            token,
			"sub":              userContext.IdentityModel.Name,
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
