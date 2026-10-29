package siteapiutils

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/errors"
	"nighttroop.t1d.nz/utils/httputils"
	"nighttroop.t1d.nz/utils/security"
)

// There are three ways to validate:
// 1. API: JWT token as Authorization header (for API requests)
// 2. API: (not good) Raw Token as http param in URL. This is slow and insecure and not recommended.
// 3. BROWSER: Cookie with JWT token (we are not supporting token in URL that is a bad idea)
func JWTAuthMiddleware(controlDB *db.ControlDB, cfg *config.ServerConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			acceptType := httputils.GetAcceptType(r)

			identity, err := TryAuthWithJWTToken(cfg, r)
			if err == nil {
				ctx = httputils.SetIdentityModelInSiteContext(ctx, identity)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			// A failed sign in spends the IP's budget on the site, not the site's own limit its identities need
			if allowed, _ := httputils.SiteUnauthWithinRateLimit(cfg, r); !allowed {
				w.WriteHeader(http.StatusTooManyRequests)
				slog.Info("Site unauthenticated rate limit exceeded")
				return
			}

			if acceptType.API() {
				slog.Info("Identity Login Failed: invalid JWT token", "error", err, "stack", errors.LocalStack())
				http.Error(w, "Invalid credentials", http.StatusUnauthorized)

				return
			} else {
				w.Header().Set("HX-Redirect", "/login")
				http.Redirect(w, r, "/login", http.StatusSeeOther)
			}

		})
	}
}

// Try get the Auth Header
// Then try get the cookie
const BearerPrefix = "bearer "

func TryAuthWithJWTToken(config *config.ServerConfig, r *http.Request) (*controldb.SiteIdentity, error) {
	authHeader := r.Header.Get("Authorization")
	hasBearer := strings.HasPrefix(strings.ToLower(authHeader), strings.ToLower(BearerPrefix))
	if hasBearer {
		jwttokenStr := authHeader[len(BearerPrefix):]
		subdomain, identityName, role, err := config.Security.VerifyUserJWTToken(jwttokenStr)
		if err != nil {
			return nil, fmt.Errorf("bad JWT token: %w", err)
		}
		identity, err := httputils.GetSiteIdentity(r.Context(), subdomain, identityName, consts.IdentityRole(role))
		if err != nil {
			slog.Warn("Bad JWT token cant find identity", "error", err)
			return nil, fmt.Errorf("identity not found: %w", err)
		}
		return identity, nil
	}

	if authHeader != "" {
		return nil, fmt.Errorf("Auth Header exists but is not Bearer token")
	}

	// Check for Cookie
	cookieToken, err := security.GetUserTokenFromRequest(r)
	if err == nil {
		jwttokenStr := cookieToken
		subdomain, identityName, role, err := config.Security.VerifyUserJWTToken(jwttokenStr)
		if err != nil {
			slog.Info("Bad Cookie JWT token", "error", err)
			return nil, fmt.Errorf("Bad Cookie JWT token: %w", err)
		}
		identity, err := httputils.GetSiteIdentity(r.Context(), subdomain, identityName, consts.IdentityRole(role))
		if err != nil {
			slog.Warn("Bad Cookie JWT token cant find identity", "error", err)
			return nil, fmt.Errorf("Bad Cookie JWT token: %w", err)
		}
		return identity, nil
	}

	return nil, fmt.Errorf("No JWT token found")
}
