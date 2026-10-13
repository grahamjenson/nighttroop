package controlapi

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/httputils"
	"nighttroop.t1d.nz/utils/security"
)

func JWTAuthMiddleware(controlDB *db.ControlDB, config *config.ServerConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var err error

			controlIdentity, err := tryLoginWithJWTToken(controlDB, config, r)

			if err == nil {
				slog.Info("Control User Logged In", "user_name", controlIdentity.Name)
				ctx := r.Context()
				httputils.SetAdminControlContext(ctx, controlIdentity)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			slog.Info("Control User Not Logged In", "error", err.Error())

			if httputils.GetAcceptType(r).API() {
				http.Error(w, "Unauthorized: Invalid JWT token", http.StatusUnauthorized)
			} else {
				w.Header().Set("HX-Redirect", "/login")
				http.Redirect(w, r, "/login", http.StatusSeeOther)
			}
		})
	}
}

func tryLoginWithJWTToken(controlDB *db.ControlDB, config *config.ServerConfig, r *http.Request) (*controldb.ControlIdentity, error) {
	userName, role, err := verifyControlJWTToken(controlDB, config, r)

	if err != nil {
		return nil, err
	}

	controlIdentity, err := controlDB.Queries.GetControlIdentity(context.Background(), controldb.GetControlIdentityParams{
		Name: userName,
	})

	if err != nil {
		return nil, err
	}

	if role != controlIdentity.Role {
		return nil, fmt.Errorf("Role mismatch: expected %s, got %s", controlIdentity.Role, role)
	}

	if err != nil {
		return nil, fmt.Errorf("Failed to get control identity: %w", err)
	}

	return controlIdentity, nil
}

func verifyControlJWTToken(controlDB *db.ControlDB, config *config.ServerConfig, r *http.Request) (string, consts.ControlRole, error) {
	authHeader := r.Header.Get("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		jwttokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		adminName, role, err := config.Security.VerifyControlJWTToken(jwttokenStr)
		if err != nil {
			return "", consts.ControlRole(""), fmt.Errorf("bad JWT token: %w", err)
		}

		return adminName, consts.ControlRole(role), nil
	}

	if authHeader != "" {
		return "", consts.ControlRole(""), fmt.Errorf("Auth Header exists but is not Bearer token")
	}

	// Check for Cookie
	cookieToken, err := security.GetControlTokenFromRequest(r)
	if err == nil {
		jwttokenStr := cookieToken
		adminName, role, err := config.Security.VerifyControlJWTToken(jwttokenStr)
		if err != nil {
			return "", consts.ControlRole(""), fmt.Errorf("Bad Cookie JWT token: %w", err)
		}
		return adminName, consts.ControlRole(role), nil
	}

	return "", consts.ControlRole(""), fmt.Errorf("No JWT token found")
}
