package nightscoutapi

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/httputils"
	"nighttroop.t1d.nz/utils/security"
)

func NightscoutAuthMiddleware(controlDB *db.ControlDB, cfg *config.ServerConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// An IP out of its unauthenticated budget is refused before its token is checked, which can cost a bcrypt
			if httputils.SiteUnauthSpent(cfg, r) {
				w.WriteHeader(http.StatusTooManyRequests)
				slog.InfoContext(ctx, "Site unauthenticated rate limit exceeded")
				return
			}

			identityModel, keepGoing, err := tryAuthWithJWTToken(cfg, r)
			if err == nil {

				ctx = httputils.SetIdentityModelInSiteContext(ctx, identityModel)

				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			// If the JWT token is present but incorrect we need to stop
			if !keepGoing {
				slog.InfoContext(r.Context(), "JWT Auth Failed for Nightscout API", "error", err.Error())
				// A failed auth spends the IP's budget on the site, as an authenticated request does not
				httputils.SiteUnauthWithinRateLimit(cfg, r)
				http.Error(w, "Unauthorized: Invalid JWT token", http.StatusUnauthorized)
				return
			}

			// Last Chance
			identityModel, err = tryAuthWithLegacyToken(cfg, r)
			if err == nil {
				// A JWT is got with a token, so this is where each client authenticates
				httputils.RecordAuthClient(r, identityModel)
				ctx = httputils.SetIdentityModelInSiteContext(ctx, identityModel)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			slog.InfoContext(r.Context(), "JWT and Legacy Token Auth Failed for Nightscout API", "error", err.Error())
			httputils.SiteUnauthWithinRateLimit(cfg, r)
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
		identity, err := jwtIdentity(config, r.Context(), authHeader[len(BearerPrefix):])
		return identity, false, err
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

	return legacyTokenIdentity(config, r.Context(), tokenParam)
}

// jwtIdentity is the site identity a JWT is for
func jwtIdentity(config *config.ServerConfig, ctx context.Context, jwttokenStr string) (*controldb.SiteIdentity, error) {
	return httputils.GetJWTSiteIdentity(ctx, config, jwttokenStr)
}

// legacyTokenIdentity is the site identity a Nightscout token (or API secret) is for
func legacyTokenIdentity(config *config.ServerConfig, ctx context.Context, tokenParam string) (*controldb.SiteIdentity, error) {
	if tokenParam == "" {
		return nil, fmt.Errorf("no token provided")
	}

	siteContext := httputils.GetSiteContext(ctx)

	// xDrip+ sends the SHA1 of the token, 40 hex digits, as Nightscout clients send their API secret
	if len(tokenParam) == 40 {
		identity, err := siteContext.ControlDB.Queries.GetSiteIdentityByTokenDigest(ctx, controldb.GetSiteIdentityByTokenDigestParams{
			SiteID: siteContext.SiteID, Digest: new(security.NightscoutTokenDigest(tokenParam))})
		if err != nil || !identity.Enabled || identity.TotpSecretEnc != "" || identity.Role == consts.IdentityRoleAdmin {
			return nil, fmt.Errorf("no identity that can use a token has one with this SHA1: %v", err)
		}
		return identity, nil
	}

	tokenParamSubdomain, tokenParamName, tokenRole, tokenParamPassword, err := config.Security.SplitNightscoutTokenForIdentity(tokenParam)
	if err != nil {
		return nil, fmt.Errorf("invalid token format: %w", err)
	}

	if tokenParamSubdomain != siteContext.SiteID {
		return nil, fmt.Errorf("token subdomain does not match user context subdomain")
	}

	identity, err := httputils.GetSiteIdentity(ctx, tokenParamSubdomain, tokenParamName, consts.IdentityRole(tokenRole))
	if err != nil {
		return nil, fmt.Errorf("identity not found: %w", err)
	}

	if identity.TotpSecretEnc != "" {
		return nil, fmt.Errorf("legacy token cannot be used for identities with TOTP enabled")
	}

	// Constraint: Not allowed to be admin and use legacy token
	if identity.Role == consts.IdentityRoleAdmin {
		return nil, fmt.Errorf("legacy token cannot be used for admin identities")
	}

	if !config.Security.VerifyPasswordWithCache(tokenParamPassword, identity.PasswordHash) {
		return nil, fmt.Errorf("invalid password")
	}

	return identity, nil
}
