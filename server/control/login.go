package control

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/sql/control"
	"nighttroop.t1d.nz/server/templates"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/errors"
	"nighttroop.t1d.nz/utils/httputils"
	"nighttroop.t1d.nz/utils/security"
)

type LoginRequest struct {
	Name     string `json:"name"`
	Password string `json:"password"`
	TOTP     string `json:"totp"`
}

func POSTLogoutHandler(config config.ServerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		config.Security.ClearControlLoginCookie(w)

		// 3. Redirect the user to the login page
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}
}

func GETLoginPageHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		controlContext := httputils.GetControlContext(r.Context())
		w.Header().Set("Content-Type", "text/html")
		err := templates.ControlLoginHTML(w, controlContext)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// curl -v -X POST http://control.localhost:8080/token -d '
// {
//   "username": "admin",
//   "password": "password",
//   "totp": "123123"
// }'

func POSTLogin(controlDB *db.ControlDB, config config.ServerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// is this HTML FORM or JSON?
		contentType := httputils.GetContentType(r)

		username := ""
		password := ""
		totpToken := ""

		switch contentType {
		case httputils.ContentTypeJSON:
			var creds LoginRequest
			if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
				errors.GenericAdminLoginError(w, r, fmt.Errorf("Invalid JSON payload: %w", err))
				return
			}

			username = creds.Name
			password = creds.Password
			totpToken = creds.TOTP
		case httputils.ContentTypeForm:

			if err := r.ParseForm(); err != nil {
				errors.GenericAdminLoginError(w, r, fmt.Errorf("Invalid form data: %w", err))
				return
			}

			username = r.FormValue("name")
			password = r.FormValue("password")
			totpToken = r.FormValue("totp")
		default:
			errors.GenericAdminLoginError(w, r, fmt.Errorf("Unsupported Content-Type: %s", contentType))
			return
		}

		allowedToAttempt, _ := config.RateLimiters.UserLoginAllow(fmt.Sprintf("%s:%s", "control", username), r.RemoteAddr)
		if !allowedToAttempt {
			http.Error(w, "Too many login attempts Wait a bit", http.StatusTooManyRequests)
			return
		}

		admin, err := loginTheControlUser(controlDB, config, username, password, totpToken)
		if err != nil {
			errors.GenericAdminLoginError(w, r, fmt.Errorf("Invalid credentials: %w", err))
			return
		}

		tokenString, expiresAt, err := config.Security.GenerateControlJWTToken(admin.Name, string(admin.Role))
		if err != nil {
			errors.GenericAdminLoginError(w, r, fmt.Errorf("Failed to generate JWT token: %w", err))
			return
		}

		if httputils.GetAcceptType(r).API() {
			// Default to JSON because this is API first
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{
				"token":      tokenString,
				"expires_at": expiresAt.Format(time.RFC3339),
			})
			return
		}

		// Default set the cookie
		config.Security.SetControlLoginCookie(w, tokenString, expiresAt)

		slog.Info("Admin logged in", "admin", admin.Name, "role", admin.Role)
		w.Header().Set("HX-Redirect", "/")

	}
}

func loginTheControlUser(controlDB *db.ControlDB, config config.ServerConfig, username, password, totpToken string) (*control.Admin, error) {
	admin, err := controlDB.Queries.GetAdmin(context.TODO(), username)
	if err != nil {
		return nil, fmt.Errorf("admin not found: %w", err)
	}
	totpSecret, err := config.Security.Decrypt(admin.TotpSecret)
	if err != nil {
		return nil, err
	}

	if !security.VerifyPassword(password, admin.Password) {
		return nil, fmt.Errorf("invalid password")
	}
	if !config.Security.TOTPValidate(totpToken, totpSecret) {
		return nil, fmt.Errorf("invalid TOTP token")
	}
	return admin, nil
}
