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
				slog.InfoContext(r.Context(), "Control User Logged In", "user_name", controlIdentity.Name)
				// The status banner also warns of resources high at the last check, and of sites near their limit
				status := config.Status
				status.HighResources = controlDB.HighResources(r.Context())
				status.Sites, status.MaxSites = controlDB.GetNSitesDBs(), config.MaxSites
				ctx := httputils.SetAdminControlContext(r.Context(), controlIdentity, status)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			slog.InfoContext(r.Context(), "Control User Not Logged In", "error", err.Error())

			if httputils.GetAcceptType(r).API() {
				http.Error(w, "Unauthorized: Invalid JWT token", http.StatusUnauthorized)
			} else {
				httputils.RedirectToLogin(w, r, "/login")
			}
		})
	}
}

// SignedIn is whether the request has a control admin's valid sign in token, before the control site checks the admin
func SignedIn(config *config.ServerConfig, r *http.Request) bool {
	_, _, _, err := verifyControlJWTToken(nil, config, r)
	return err == nil
}

// tryLoginWithJWTToken is the control admin a JWT is for. A JWT issued before the admin's password or TOTP last
// changed is refused, so a reset signs out whoever had the old ones.
func tryLoginWithJWTToken(controlDB *db.ControlDB, config *config.ServerConfig, r *http.Request) (*controldb.ControlIdentity, error) {
	userName, role, credentials, err := verifyControlJWTToken(controlDB, config, r)

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

	if err := credentials.CheckCurrent(controlIdentity); err != nil {
		return nil, err
	}

	if err != nil {
		return nil, fmt.Errorf("Failed to get control identity: %w", err)
	}

	return controlIdentity, nil
}

// verifyControlJWTToken is the control admin name, role and credentials of the request's JWT
func verifyControlJWTToken(controlDB *db.ControlDB, config *config.ServerConfig, r *http.Request) (string, consts.ControlRole, security.Credentials, error) {
	authHeader := r.Header.Get("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		jwttokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		adminName, role, credentials, err := config.Security.VerifyControlJWTToken(jwttokenStr)
		if err != nil {
			return "", consts.ControlRole(""), security.Credentials{}, fmt.Errorf("bad JWT token: %w", err)
		}

		return adminName, consts.ControlRole(role), credentials, nil
	}

	if authHeader != "" {
		return "", consts.ControlRole(""), security.Credentials{}, fmt.Errorf("Auth Header exists but is not Bearer token")
	}

	// Check for Cookie
	cookieToken, err := security.GetControlTokenFromRequest(r)
	if err == nil {
		jwttokenStr := cookieToken
		adminName, role, credentials, err := config.Security.VerifyControlJWTToken(jwttokenStr)
		if err != nil {
			return "", consts.ControlRole(""), security.Credentials{}, fmt.Errorf("Bad Cookie JWT token: %w", err)
		}
		return adminName, consts.ControlRole(role), credentials, nil
	}

	return "", consts.ControlRole(""), security.Credentials{}, fmt.Errorf("No JWT token found")
}
