package user

import (
	"fmt"
	"net/http"
	"strings"

	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/httputils"
	"nighttroop.t1d.nz/utils/security"
)

func UserRateLimitMiddleware(config config.ServerConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			method := strings.ToUpper(r.Method)
			if method == http.MethodGet || method == http.MethodHead {
				// For GET and HEAD requests, we can use a more permissive rate limit
				allowed, remainingTime := httputils.UserReadAllowed(config, r)
				// TODO handle both html and json response types
				if !allowed {
					w.WriteHeader(http.StatusTooManyRequests)
					fmt.Fprintf(w, "Too many requests, read rate limit exceeded. Try again in %d seconds.", remainingTime)
					return
				}
			} else {
				// For POST, PUT, DELETE, etc., we should enforce a stricter rate limit
				allowed, remainingTime := httputils.UserWriteAllowed(config, r)
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

// There are three ways to validate:
// 1. API: JWT token as Authorization header (for API requests)
// 2. API: (not good) Raw Token as http param in URL. This is slow and insecure and not recommended.
// 3. BROWSER: Cookie with JWT token (we are not supporting token in URL that is a bad idea)
func JWTPlusTokenAuthMiddleware(controlDB *db.ControlDB, cfg config.ServerConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			jwttokenStr := ""

			// 1. Check API Header (Bearer Token)
			authHeader := r.Header.Get("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				jwttokenStr = strings.TrimPrefix(authHeader, "Bearer ")
			}

			// 3. Check Browser Cookie (Fallback for non-JSON or missing headers)
			if jwttokenStr == "" {
				cookieToken, err := security.GetUserTokenFromRequest(r)
				if err == nil {
					jwttokenStr = cookieToken
				}
			}

			if jwttokenStr != "" {
				deviceName, role, err := cfg.Security.VerifyUserJWTToken(jwttokenStr)
				if err != nil {
					http.Error(w, "Bad JWT Token", http.StatusInternalServerError)
					return
				}
				httputils.SetDeviceNameInHTTPContext(ctx, deviceName)
				httputils.SetRoleInHTTPContext(ctx, role)
				next.ServeHTTP(w, r.WithContext(ctx))
				return

			}

			/////
			// Legacy slow annoying method of validating the token param
			////
			// The only other means of authing is with token param
			tokenParam := r.URL.Query().Get("token")
			if tokenParam == "" {
				http.Error(w, "Unauthorized: No token provided", http.StatusUnauthorized)
				return
			}

			subdomain := httputils.GetSubdomainFromHTTPContext(ctx)
			userDB, _, err := controlDB.GetUser(subdomain, false)
			if err != nil {
				http.Error(w, "User not found", http.StatusNotFound)
				return
			}

			deviceName, role, err := cfg.Security.VerifyLegacyTokenForDevice(tokenParam, subdomain, userDB.DevicesGet)

			if err != nil {
				http.Error(w, "Bad Token", http.StatusInternalServerError)
				return
			}

			// Optional: Inject claims into context
			ctx = httputils.SetDeviceNameInHTTPContext(ctx, deviceName)
			ctx = httputils.SetRoleInHTTPContext(ctx, role)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
