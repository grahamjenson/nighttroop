package nightscout_api

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/server/assets/static"
	v1 "nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/api/v1"
	v3 "nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/api/v3"
	vendorConfig "nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/config"
	"nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/deps"
	"nighttroop.t1d.nz/server/user"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/httputils"
)

func NewNightscoutRouter(controlDB *db.ControlDB, config config.ServerConfig) chi.Router {
	r := chi.NewRouter()
	httputils.AddStandardMiddleware(r, config)
	r.Use(user.UserRateLimitMiddleware(config))

	static.AddSatticRoutes(r)

	r.Group(func(r chi.Router) {
		r.Use(NightscoutAuthMiddleware(controlDB, config))
		r.Use(user.UserDBMiddleware(controlDB))
		r.Get("/api/v2/authorization/request/token={token}", getAuthRoute(controlDB, config))

		ntDeps := deps.Dependencies{
			Config: vendorConfig.Config{}.WithDefaults(),
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
		v3Router := v3.NewNightscoutV3Router(ntDeps)

		r.Mount("/api/v1", v1Router)
		r.Mount("/api/v3", v3Router)
	})

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		slog.Info("Nightscout router 404 for path", "path", r.URL.Path)
		http.Error(w, "Not Found", http.StatusNotFound)
	})

	return r
}

func getAuthRoute(controlDB *db.ControlDB, config config.ServerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// The user should have already been verified
		token, expiresAt, err := config.Security.GenerateUserJWTToken("asd", "asd")
		if err != nil {
			http.Error(w, "Internal Server Error: Failed to generate JWT token", http.StatusInternalServerError)
			return
		}

		responce := map[string]any{
			"token":            token,
			"sub":              "read",
			"permissionGroups": [][]string{{"*"}, {"*:*:read"}},
			"iat":              time.Now().Unix(),
			"exp":              expiresAt.Unix(),
		}

		jsonResponce, err := json.Marshal(responce)
		if err != nil {
			http.Error(w, "Internal Server Error: Failed to marshal JSON", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonResponce)
	}

}

func NightscoutAuthMiddleware(controlDB *db.ControlDB, cfg config.ServerConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			identityName, identityRole, keepGoing, err := user.TryLoginWithJWTToken(cfg, r)
			if err == nil {
				ctx = httputils.SetIdentityNameInUserContext(ctx, identityName)
				ctx = httputils.SetIdentityRoleInUserContext(ctx, identityRole)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			// If the JWT token is present but incorrect we need to stop
			if !keepGoing {
				http.Error(w, "Unauthorized: Invalid JWT token", http.StatusUnauthorized)
				return
			}

			// Last Chance
			identityName, identityRole, err = tryLoginWithLegacyToken(controlDB, cfg, r)
			if err == nil {
				ctx = httputils.SetIdentityNameInUserContext(ctx, identityName)
				ctx = httputils.SetIdentityRoleInUserContext(ctx, identityRole)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			http.Error(w, "Unauthorized: No valid authentication method found", http.StatusUnauthorized)
		})
	}
}

func tryLoginWithLegacyToken(controlDB *db.ControlDB, config config.ServerConfig, r *http.Request) (string, consts.IdentityRole, error) {

	tokenParam := r.URL.Query().Get("token")
	if tokenParam == "" {
		return "", "", fmt.Errorf("no token provided")
	}

	userContext := httputils.GetUserContext(r.Context())
	userDB, _, err := controlDB.GetUser(userContext.Subdomain, false)
	if err != nil {
		return "", "", fmt.Errorf("user not found: %w", err)
	}

	identityName, role, err := config.Security.VerifyLegacyTokenForIdentity(tokenParam, userContext.Subdomain, userDB.GetIdentitySplat)
	if err != nil {
		return "", "", fmt.Errorf("bad token: %w", err)
	}

	return identityName, consts.IdentityRole(role), nil
}
