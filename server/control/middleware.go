package control

import (
	"fmt"
	"net/http"
	"strings"

	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/httputils"
	"nighttroop.t1d.nz/utils/security"
)

func JWTAuthMiddleware(controlDB *db.ControlDB, config config.ServerConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var err error

			deviceName, deviceRole, err := tryLoginWithJWTToken(controlDB, config, r)
			if err == nil {
				ctx := r.Context()
				httputils.SetDeviceNameInHTTPContext(ctx, deviceName)
				httputils.SetRoleInHTTPContext(ctx, deviceRole)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			if httputils.GetAcceptType(r).API() {
				http.Error(w, "Unauthorized: Invalid JWT token", http.StatusUnauthorized)
			} else {
				w.Header().Set("HX-Redirect", "/login")
				http.Redirect(w, r, "/login", http.StatusSeeOther)
			}
			return
		})
	}
}

func tryLoginWithJWTToken(controlDB *db.ControlDB, config config.ServerConfig, r *http.Request) (string, string, error) {
	authHeader := r.Header.Get("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		jwttokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		deviceName, role, err := config.Security.VerifyUserJWTToken(jwttokenStr)
		if err != nil {
			return "", "", fmt.Errorf("bad JWT token: %w", err)
		}
		return deviceName, role, nil
	}

	if authHeader != "" {
		return "", "", fmt.Errorf("Auth Header exists but is not Bearer token")
	}

	// Check for Cookie
	cookieToken, err := security.GetUserTokenFromRequest(r)
	if err == nil {
		jwttokenStr := cookieToken
		deviceName, role, err := config.Security.VerifyUserJWTToken(jwttokenStr)
		if err != nil {
			return "", "", fmt.Errorf("Bad Cookie JWT token: %w", err)
		}
		return deviceName, role, nil
	}

	return "", "", fmt.Errorf("No JWT token found")
}
