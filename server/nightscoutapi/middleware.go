package nightscoutapi

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/server/siteapi"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/httputils"
)

func NightscoutAuthMiddleware(controlDB *db.ControlDB, cfg *config.ServerConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			identityModel, keepGoing, err := tryAuthWithJWTToken(cfg, r)
			if err == nil {

				ctx = httputils.SetIdentityModelInSiteContext(ctx, identityModel)

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
			identityModel, err = tryAuthWithLegacyToken(cfg, r)
			if err == nil {
				ctx = httputils.SetIdentityModelInSiteContext(ctx, identityModel)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			slog.Info("JWT and Legacy Token Auth Failed for Nightscout API", "error", err.Error())
			http.Error(w, "Unauthorized: No valid authentication method found", http.StatusUnauthorized)
		})
	}
}

// Try get the Auth Header
// Then try get the cookie
const BearerPrefix = "bearer "

func tryAuthWithJWTToken(config *config.ServerConfig, r *http.Request) (*controldb.SiteIdentity, bool, error) {
	authHeader := r.Header.Get("Authorization")
	hasBearer := strings.HasPrefix(strings.ToLower(authHeader), strings.ToLower(BearerPrefix))
	if hasBearer {
		jwttokenStr := authHeader[len(BearerPrefix):]
		subdomain, identityName, role, err := config.Security.VerifyUserJWTToken(jwttokenStr)
		if err != nil {
			return nil, false, fmt.Errorf("bad JWT token: %w", err)
		}
		identity, err := siteapi.GetSiteIdentity(r.Context(), subdomain, identityName, consts.IdentityRole(role))
		if err != nil {
			slog.Error("Bad JWT token cant find identity", "error", err)
			return nil, false, fmt.Errorf("identity not found: %w", err)
		}
		return identity, false, nil
	}

	if authHeader != "" {
		return nil, false, fmt.Errorf("Auth Header exists but is not Bearer token")
	}

	return nil, true, fmt.Errorf("No JWT token found")
}

func tryAuthWithLegacyToken(config *config.ServerConfig, r *http.Request) (*controldb.SiteIdentity, error) {

	// We first try to get the token, then the secret param, both equal in Nighttroop
	tokenParam := r.URL.Query().Get("token")
	if tokenParam == "" {
		tokenParam = chi.URLParam(r, "token")
		if tokenParam == "" {
			tokenParam = r.URL.Query().Get("secret")
			if tokenParam == "" {
				// get api-secret header
				tokenParam = r.Header.Get("api-secret")
			}
		}
	}

	if tokenParam == "" {
		return nil, fmt.Errorf("no token provided")
	}

	siteContext := httputils.GetSiteContext(r.Context())

	tokenParamSubdomain, tokenParamName, tokenRole, tokenParamPassword, err := config.Security.SplitNightscoutTokenForIdentity(tokenParam)
	if err != nil {
		return nil, fmt.Errorf("invalid token format: %w", err)
	}

	if tokenParamSubdomain != siteContext.Subdomain {
		return nil, fmt.Errorf("token subdomain does not match user context subdomain")
	}

	identity, err := siteapi.GetSiteIdentity(r.Context(), tokenParamSubdomain, tokenParamName, consts.IdentityRole(tokenRole))
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

	if !config.Security.VerifyPasswordWithCache(tokenParamPassword, identity.Password) {
		return nil, fmt.Errorf("invalid password")
	}

	return identity, nil
}
