package user

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/sql/user"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/errors"
	"nighttroop.t1d.nz/utils/httputils"
	"nighttroop.t1d.nz/utils/security"
)

func UserRateLimitMiddleware(config config.ServerConfig) func(http.Handler) http.Handler {

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userAction := httputils.UserActionRead
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
				userAction = httputils.UserActionWrite
				if !allowed {
					w.WriteHeader(http.StatusTooManyRequests)
					fmt.Fprintf(w, "Too many requests, write rate limit exceeded. Try again in %d seconds.", remainingTime)
					return
				}
			}
			ctx := httputils.SetUserActionInUserContext(r.Context(), userAction)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// THIS IS SO STUPID THAT I HAVE TO DO THIS TO CHECK THE STATUS OF THE RETURNED RESPONSE.
type statusResponseWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusResponseWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func UserIdentitySeenFlagsUpdateMiddleware(controlDB *db.ControlDB) func(http.Handler) http.Handler {

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			userContext := httputils.GetUserContext(r.Context())

			sw := &statusResponseWriter{
				ResponseWriter: w,
				status:         http.StatusOK, // default if WriteHeader is never called
			}

			next.ServeHTTP(sw, r)

			if sw.status == 200 {
				wasWrite := userContext.UserAction == httputils.UserActionWrite
				controlDB.UpdateUserDateFlags(userContext.UserModel, true, wasWrite)
				userContext.UserDB.UpdateIdentityDateFlags(userContext.IdentityModel, true, wasWrite)
			}
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

			identity, _, err := TryAuthWithJWTToken(cfg, r)
			if err == nil {
				ctx = httputils.SetIdentityModelInUserContext(ctx, identity)
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
func TryAuthWithJWTToken(config config.ServerConfig, r *http.Request) (*user.Identity, bool, error) {
	authHeader := r.Header.Get("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		jwttokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		subdomain, identityName, role, err := config.Security.VerifyUserJWTToken(jwttokenStr)
		if err != nil {
			return nil, false, fmt.Errorf("bad JWT token: %w", err)
		}
		identity, err := GetIdentity(r.Context(), subdomain, identityName, consts.IdentityRole(role))
		if err != nil {
			return nil, false, fmt.Errorf("identity not found: %w", err)
		}
		return identity, false, nil
	}

	if authHeader != "" {
		return nil, false, fmt.Errorf("Auth Header exists but is not Bearer token")
	}

	// Check for Cookie
	cookieToken, err := security.GetUserTokenFromRequest(r)
	if err == nil {
		jwttokenStr := cookieToken
		subdomain, identityName, role, err := config.Security.VerifyUserJWTToken(jwttokenStr)
		if err != nil {
			return nil, false, fmt.Errorf("Bad Cookie JWT token: %w", err)
		}
		identity, err := GetIdentity(r.Context(), subdomain, identityName, consts.IdentityRole(role))
		if err != nil {
			return nil, false, fmt.Errorf("Bad Cookie JWT token: %w", err)
		}
		return identity, false, nil
	}

	return nil, true, fmt.Errorf("No JWT token found")
}

func GetIdentity(ctx context.Context, subdomain string, identityName string, role consts.IdentityRole) (*user.Identity, error) {
	userContext := httputils.GetUserContext(ctx)
	if subdomain != userContext.Subdomain {
		return nil, fmt.Errorf("subdomain mismatch: expected %s, got %s", userContext.Subdomain, subdomain)
	}

	identity, err := userContext.UserDB.Queries.GetIdentity(ctx, identityName)
	if err != nil {
		return nil, err
	}

	if !identity.Enabled {
		return nil, fmt.Errorf("identity disabled")
	}

	if consts.IdentityRole(identity.Role) != role {
		return nil, fmt.Errorf("role mismatch")
	}

	return identity, nil
}
