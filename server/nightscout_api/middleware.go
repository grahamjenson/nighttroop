package nightscout_api

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"nighttroop.t1d.nz/data/db"
	userSQL "nighttroop.t1d.nz/data/sql/user"
	"nighttroop.t1d.nz/server/user"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/httputils"
	"nighttroop.t1d.nz/utils/security"
)

func NightscoutAuthMiddleware(controlDB *db.ControlDB, cfg config.ServerConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			identityModel, keepGoing, err := user.TryAuthWithJWTToken(cfg, r)
			if err == nil {

				ctx = httputils.SetIdentityModelInUserContext(ctx, identityModel)

				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			// If the JWT token is present but incorrect we need to stop
			if !keepGoing {
				slog.Info("JWT Auth Failed for Nightscout API", "error", err.Error())
				http.Error(w, "Unauthorized: Invalid JWT token", http.StatusUnauthorized)
				return
			}

			// Last Chance
			identityModel, err = tryAuthWithLegacyToken(controlDB, cfg, r)
			if err == nil {
				ctx = httputils.SetIdentityModelInUserContext(ctx, identityModel)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			slog.Info("JWT and Legacy Token Auth Failed for Nightscout API", "error", err.Error())
			http.Error(w, "Unauthorized: No valid authentication method found", http.StatusUnauthorized)
		})
	}
}

func tryAuthWithLegacyToken(controlDB *db.ControlDB, config config.ServerConfig, r *http.Request) (*userSQL.Identity, error) {

	tokenParam := r.URL.Query().Get("token")
	if tokenParam == "" {
		tokenParam = chi.URLParam(r, "token")
	}

	if tokenParam == "" {
		return nil, fmt.Errorf("no token provided")
	}

	userContext := httputils.GetUserContext(r.Context())

	tokenParamSubdomain, tokenParamName, tokenRole, tokenParamPassword, err := config.Security.SplitNightscoutTokenForIdentity(tokenParam)
	if err != nil {
		return nil, fmt.Errorf("invalid token format: %w", err)
	}

	if tokenParamSubdomain != userContext.Subdomain {
		return nil, fmt.Errorf("token subdomain does not match user context subdomain")
	}

	identity, err := user.GetIdentity(r.Context(), tokenParamSubdomain, tokenParamName, consts.IdentityRole(tokenRole))
	if err != nil {
		return nil, fmt.Errorf("identity not found: %w", err)
	}

	if identity.TotpSecret != "" {
		return nil, fmt.Errorf("legacy token cannot be used for identities with TOTP enabled")
	}

	// Constraint: Not allowed to be admin and use legacy token
	if identity.Role == consts.IdentityRoleAdmin {
		return nil, fmt.Errorf("legacy token cannot be used for admin identities")
	}

	if !security.VerifyPassword(tokenParamPassword, identity.Password) {
		return nil, fmt.Errorf("invalid password")
	}

	return identity, nil
}
