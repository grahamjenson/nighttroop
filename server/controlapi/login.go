package controlapi

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"
	"unicode/utf8"

	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/server/components/control_templ/control_login_templ"
	"nighttroop.t1d.nz/utils/apiutils"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/errors"
	"nighttroop.t1d.nz/utils/httputils"
)

type LoginRequest struct {
	Name     string `json:"name"`
	Password string `json:"password"`
	TOTP     string `json:"totp"`
}

func POSTLogoutHandler(config *config.ServerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		config.Security.ClearControlLoginCookie(w)

		// 3. Redirect the user to the login page
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}
}

func GETLoginPageHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		err := control_login_templ.Login().Render(r.Context(), w)
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

func POSTLogin(controlDB *db.ControlDB, config *config.ServerConfig) http.HandlerFunc {
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
				loginFailed(w, r, http.StatusUnauthorized, wrongLogin, fmt.Errorf("invalid JSON payload: %w", err))
				return
			}

			username = creds.Name
			password = creds.Password
			totpToken = creds.TOTP
		case httputils.ContentTypeForm:

			if err := r.ParseForm(); err != nil {
				loginFailed(w, r, http.StatusUnauthorized, wrongLogin, fmt.Errorf("invalid form data: %w", err))
				return
			}

			username = r.FormValue("name")
			password = r.FormValue("password")
			totpToken = r.FormValue("totp")
		default:
			loginFailed(w, r, http.StatusUnauthorized, wrongLogin, fmt.Errorf("unsupported Content-Type: %s", string(contentType)))
			return
		}

		// Reject names no identity can have before they reach the rate limiter, which keys on them
		if utf8.RuneCountInString(username) > consts.MAX_IDENTITY_NAME_LENGTH {
			loginFailed(w, r, http.StatusUnauthorized, wrongLogin, fmt.Errorf("username longer than %d characters", consts.MAX_IDENTITY_NAME_LENGTH))
			return
		}

		allowedToAttempt, _ := httputils.ControlLoginAttemptsWithinRateLimit(config, r, username)
		if !allowedToAttempt {
			loginFailed(w, r, http.StatusTooManyRequests, tooManyLogins, fmt.Errorf("too many login attempts"))
			return
		}

		// Control sign ins are the server's events, with no site
		from := config.GeoIP.Locate(httputils.GetIPFromHTTPContext(r.Context()))
		admin, err := loginTheControlUser(r.Context(), controlDB, config, username, password, totpToken)
		if err != nil {
			controlDB.Logs.SignInFailedEvent(r.Context(), "", from)
			loginFailed(w, r, http.StatusUnauthorized, wrongLogin, fmt.Errorf("invalid credentials: %w", err))
			return
		}
		controlDB.Logs.SignedInEvent(r.Context(), "", admin.Name, from)

		tokenString, expiresAt, err := config.Security.GenerateControlJWTToken(admin.Name, string(admin.Role))
		if err != nil {
			loginFailed(w, r, http.StatusUnauthorized, wrongLogin, fmt.Errorf("failed to generate JWT token: %w", err))
			return
		}

		if httputils.GetAcceptType(r).API() {
			// Default to JSON because this is API first
			apiutils.WriteJSON(w, http.StatusOK, map[string]string{
				"token":      tokenString,
				"expires_at": expiresAt.Format(time.RFC3339),
			})
			return
		}

		// Default set the cookie
		config.Security.SetControlLoginCookie(w, tokenString, expiresAt)

		slog.InfoContext(r.Context(), "Admin logged in", "admin", admin.Name, "role", admin.Role)
		w.Header().Set("HX-Redirect", "/")

	}
}

// The messages a failed login answers with, which never say how far through it got
const (
	wrongLogin    = "Wrong name, password or code. Codes change every 30 seconds, so use the newest one."
	tooManyLogins = "Too many login attempts. Wait a minute, then try again."
)

// loginFailed logs why a login failed, and answers with the status and message: as a toast for the login page, or as
// text for the API
func loginFailed(w http.ResponseWriter, r *http.Request, status int, message string, err error) {
	slog.InfoContext(r.Context(), "Control User Login Failed", "error", err, "stack", errors.LocalStack())
	if httputils.GetAcceptType(r) != httputils.AcceptTypeHTMX {
		http.Error(w, message, status)
		return
	}
	w.WriteHeader(status)
	if err := control_login_templ.LoginError(message).Render(r.Context(), w); err != nil {
		slog.ErrorContext(r.Context(), "Failed to render the login error", "error", err)
	}
}

func loginTheControlUser(ctx context.Context, controlDB *db.ControlDB, config *config.ServerConfig, username, password, totpToken string) (*controldb.ControlIdentity, error) {
	admin, err := controlDB.Queries.GetControlIdentity(ctx, controldb.GetControlIdentityParams{
		Name: username,
	})

	if err != nil {
		return nil, fmt.Errorf("admin not found: %w", err)
	}
	totpSecret, err := config.Security.Decrypt(admin.TotpSecretEnc)
	if err != nil {
		return nil, err
	}

	if !config.Security.VerifyPassword(password, admin.PasswordHash) {
		return nil, fmt.Errorf("invalid password")
	}
	if !config.Security.TOTPValidate(totpToken, totpSecret) {
		return nil, fmt.Errorf("invalid TOTP token")
	}
	return admin, nil
}
