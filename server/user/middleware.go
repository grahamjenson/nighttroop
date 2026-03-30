package user

import (
	"fmt"
	"net/http"
	"strings"

	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/httputils"
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

			tokenStr := ""

			// 1. Check API Header (Bearer Token)
			authHeader := r.Header.Get("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				tokenStr = strings.TrimPrefix(authHeader, "Bearer ")
			}

			// 2. Check URL Param (Raw Token - "not good/insecure" per docs)
			if tokenStr == "" {
				tokenStr = r.URL.Query().Get("token")
			}

			// 3. Check Browser Cookie (Fallback for non-JSON or missing headers)
			if tokenStr == "" {
				cookie, err := r.Cookie(TOKEN_COOKIE)
				if err == nil {
					tokenStr = cookie.Value
				}
			}

			contentType := httputils.GetAcceptType(r)

			// If still no token, handle based on content type
			if tokenStr == "" {
				if contentType == httputils.AcceptTypeJSON {
					http.Error(w, "Unauthorized: No token provided", http.StatusUnauthorized)
				} else {
					http.Redirect(w, r, "/login", http.StatusSeeOther)
				}
				return
			}

			switch contentType {
			case httputils.AcceptTypeJSON:
				authHeader := r.Header.Get("Authorization")
				if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
					http.Error(w, "Unauthorized Header", http.StatusUnauthorized)
					return
				}

				tokenStr = strings.TrimPrefix(authHeader, "Bearer ")
			default:
				cookie, err := r.Cookie(TOKEN_COOKIE)
				if err != nil {
					fmt.Println("Control Router: No token cookie found, redirecting to login")
					http.Redirect(w, r, "/login", http.StatusSeeOther)
					return
				}
				tokenStr = cookie.Value
			}

			deviceName, role, err := cfg.VerifyUserJWTToken(tokenStr)

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
