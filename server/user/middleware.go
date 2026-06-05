package user

import (
	"fmt"
	"net/http"
	"strings"

	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/errors"
	"nighttroop.t1d.nz/utils/httputils"
	"nighttroop.t1d.nz/utils/security"
)

func UserRateLimitMiddleware(config config.ServerConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			method := strings.ToUpper(r.Method)
			if method == http.MethodGet || method == http.MethodHead {
				// For GET and HEAD requests, we can use a more permissive rate limit
				allowed, remainingTime := httputils.UserGETAllowed(config, r)
				// TODO handle both html and json response types
				if !allowed {
					w.WriteHeader(http.StatusTooManyRequests)
					fmt.Fprintf(w, "Too many requests, read rate limit exceeded. Try again in %d seconds.", remainingTime)
					return
				}
			} else {
				// For POST, PUT, DELETE, etc., we should enforce a stricter rate limit
				allowed, remainingTime := httputils.UserPOSTAllowed(config, r)
				if !allowed {
					w.WriteHeader(http.StatusTooManyRequests)
					fmt.Fprintf(w, "Too many requests, write rate limit exceeded. Try again in %d seconds.", remainingTime)
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

// This middleware happens after auth
func UserDBMiddleware(controlDB *db.ControlDB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			subdomain := httputils.GetUserContext(r.Context()).Subdomain
			userDB, userGetModel, err := controlDB.GetUser(subdomain, true)

			if err != nil {
				errors.GenericIdentityLoginError(w, r, fmt.Errorf("User not found: %w", err))
				return
			}
			ctx := httputils.SetUserDBAndModelInUserContext(r.Context(), userDB, userGetModel)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// There are three ways to validate:
// 1. API: JWT token as Authorization header (for API requests)
// 2. API: (not good) Raw Token as http param in URL. This is slow and insecure and not recommended.
// 3. BROWSER: Cookie with JWT token (we are not supporting token in URL that is a bad idea)
func JWTAuthMiddleware(controlDB *db.ControlDB, cfg config.ServerConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			acceptType := httputils.GetAcceptType(r)

			deviceName, deviceRole, _, err := TryAuthWithJWTToken(cfg, r)
			if err == nil {
				ctx = httputils.SetIdentityNameInUserContext(ctx, deviceName)
				ctx = httputils.SetIdentityRoleInUserContext(ctx, deviceRole)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			if acceptType.API() {
				errors.GenericIdentityLoginError(w, r, fmt.Errorf("Unauthorized: Invalid JWT token"))
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
func TryAuthWithJWTToken(config config.ServerConfig, r *http.Request) (string, consts.IdentityRole, bool, error) {
	authHeader := r.Header.Get("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		jwttokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		deviceName, role, err := config.Security.VerifyUserJWTToken(jwttokenStr)
		if err != nil {
			return "", "", false, fmt.Errorf("bad JWT token: %w", err)
		}
		return deviceName, consts.IdentityRole(role), true, nil
	}

	if authHeader != "" {
		return "", "", false, fmt.Errorf("Auth Header exists but is not Bearer token")
	}

	// Check for Cookie
	cookieToken, err := security.GetUserTokenFromRequest(r)
	if err == nil {
		jwttokenStr := cookieToken
		deviceName, role, err := config.Security.VerifyUserJWTToken(jwttokenStr)
		if err != nil {
			return "", "", false, fmt.Errorf("Bad Cookie JWT token: %w", err)
		}
		return deviceName, consts.IdentityRole(role), true, nil
	}

	return "", "", true, fmt.Errorf("No JWT token found")
}
