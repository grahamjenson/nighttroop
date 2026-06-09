package control

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/httputils"
	"nighttroop.t1d.nz/utils/security"
)

func JWTAuthMiddleware(controlDB *db.ControlDB, config config.ServerConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var err error

			userName, userRole, err := tryLoginWithJWTToken(config, r)

			if err == nil {
				slog.Info("Control User Logged In", "user_name", userName, "role", userRole)
				ctx := r.Context()
				httputils.SetAdminNameInControlContext(ctx, userName)
				httputils.SetAdminRoleInControlContext(ctx, userRole)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			slog.Info("Control User Not Logged In", "user_name", userName, "role", userRole, "error", err.Error())

			if httputils.GetAcceptType(r).API() {
				http.Error(w, "Unauthorized: Invalid JWT token", http.StatusUnauthorized)
			} else {
				w.Header().Set("HX-Redirect", "/login")
				http.Redirect(w, r, "/login", http.StatusSeeOther)
			}
		})
	}
}

func tryLoginWithJWTToken(config config.ServerConfig, r *http.Request) (string, consts.AdminRole, error) {
	authHeader := r.Header.Get("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		jwttokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		adminName, role, err := config.Security.VerifyControlJWTToken(jwttokenStr)
		if err != nil {
			return "", "", fmt.Errorf("bad JWT token: %w", err)
		}
		return adminName, consts.AdminRole(role), nil
	}

	if authHeader != "" {
		return "", "", fmt.Errorf("Auth Header exists but is not Bearer token")
	}

	// Check for Cookie
	cookieToken, err := security.GetControlTokenFromRequest(r)
	if err == nil {
		jwttokenStr := cookieToken
		adminName, role, err := config.Security.VerifyControlJWTToken(jwttokenStr)
		if err != nil {
			return "", "", fmt.Errorf("Bad Cookie JWT token: %w", err)
		}
		return adminName, consts.AdminRole(role), nil
	}

	return "", "", fmt.Errorf("No JWT token found")
}
