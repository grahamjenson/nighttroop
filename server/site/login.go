package site

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/sql/control"

	"nighttroop.t1d.nz/server/components/site_templ"
	"nighttroop.t1d.nz/server/components/site_templ/site_login_templ"
	"nighttroop.t1d.nz/utils/can"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/errors"
	"nighttroop.t1d.nz/utils/httputils"
	"nighttroop.t1d.nz/utils/security"
)

// TODO CSRF Token protection for all
// TODO Form injections, make sure queries are good with that
type LoginRequest struct {
	Name     string `json:"name"`
	Password string `json:"password"`
	TOTP     string `json:"totp"`
}

func POSTLogoutHandler(config *config.ServerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		config.Security.ClearUserLoginCookie(w)

		// 3. Redirect the user to the login page
		w.Header().Set("HX-Redirect", "/")
		w.WriteHeader(http.StatusOK)
	}
}

func GETLoginPageHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//
		w.Header().Set("Content-Type", "text/html")
		err := site_login_templ.Login().Render(r.Context(), w)
		if err != nil {
			http.Error(w, "Template Error", http.StatusInternalServerError)
			return
		}
	}
}

// All users with an admin role must have a TOTP password, this is annoying but I think worth it.
// curl -v -X POST http://control.localhost:8080/token -d '
// {
//   "devicename": "admin",
//   "password": "password",
//   "totp": "123123"
// }'

func POSTLogin(controlDB *db.ControlDB, config *config.ServerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// API login is different enough to just immediatly bail on finding common code
		if httputils.GetAcceptType(r).API() {
			PostLoginFromAPI(controlDB, config, r, w)
			return
		}

		////
		// Browser Login
		////

		siteContext := httputils.GetSiteContext(r.Context())
		subdomain := siteContext.Subdomain

		if err := r.ParseForm(); err != nil {
			identityLoginErrorHTMX(w, r, "Failed to parse", fmt.Errorf("failed to parse form: %w", err))
			return
		}

		formName := r.FormValue("name")
		formPassword := r.FormValue("password")

		// Rate Limit login attempts by user and IP
		allowedToAttempt, _ := httputils.SiteLoginAttemptsAllowed(config, r, subdomain, formName)
		if !allowedToAttempt {
			rateLimitErrorHTMX(w, r)
			return
		}

		// First thing validate the password
		identity, err := loginTheUserPasswordOnly(siteContext.ControlDB, config.Security, siteContext.Subdomain, formName, formPassword)
		if err != nil {
			identityLoginErrorHTMX(w, r, "Invalid username or password.", fmt.Errorf("failed to login with password: %w", err))
			return
		}

		if !can.LoginToWebsite(identity.Role) {
			identityLoginErrorHTMX(w, r, "Identity does not have permission to login to website", fmt.Errorf("identity with role %s cannot login to website", identity.Role))
			return
		}

		if can.MustLoginWithTotp(identity.Role, identity.TotpSecret != "") {
			// If the Identity Requires TOTP validation, replace the original form with a new form that submits to TOTP validation input
			err = config.Security.SetUserPreAuthCookie(w, identity.Name)
			if err != nil {
				identityLoginErrorHTMX(w, r, "", fmt.Errorf("failed to set pre-auth cookie: %w", err))
				return
			}

			w.Header().Set("Content-Type", "text/html")
			err = site_login_templ.LoginTOTP().Render(r.Context(), w)
			if err != nil {
				identityLoginErrorHTMX(w, r, "", fmt.Errorf("failed to render TOTP login page: %w", err))
				return
			}
			return
		} else {
			tokenString, expiresAt, err := config.Security.GenerateUserJWTToken(siteContext.Subdomain, identity.Name, string(identity.Role))
			if err != nil {
				identityLoginErrorHTMX(w, r, "", fmt.Errorf("failed to generate JWT token: %w", err))
				return
			}

			// Set proper cookie and redirect to dashboard
			config.Security.SetUserLoginCookie(w, tokenString, expiresAt)

			slog.Info("User logged in", "device", identity.Name, "role", identity.Role)
			w.Header().Set("HX-Redirect", "/")
			w.WriteHeader(http.StatusOK)
		}

	}
}

// Browser only route for validating TOTP tokens
func POSTLoginTOTP(controlDB *db.ControlDB, config *config.ServerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if httputils.GetAcceptType(r).API() {
			http.Error(w, "API login not supported for TOTP endpoint", http.StatusBadRequest)
			return
		}
		siteContext := httputils.GetSiteContext(r.Context())

		// 1. Fetch And Validate Pre-Auth Token
		identityNameFromToken, err := fetchVerifyPreAuthToken(config, r)
		if err != nil {
			identityLoginErrorHTMX(w, r, "", fmt.Errorf("failed to fetch and verify pre-auth token: %w", err))
			return
		}

		allowedToAttempt, _ := httputils.SiteLoginAttemptsAllowed(config, r, siteContext.Subdomain, identityNameFromToken)
		if !allowedToAttempt {
			rateLimitErrorHTMX(w, r)
			return
		}

		// 2. Grab the user db and identity

		identity, err := getTheIdentityeModel(controlDB, r, identityNameFromToken)
		if err != nil {
			identityLoginErrorHTMX(w, r, "", fmt.Errorf("failed to get identity model: %w", err))
			return
		}

		// THIS IS A HTTP ONLY ROUTE
		if err := r.ParseForm(); err != nil {
			identityLoginErrorHTMX(w, r, "", fmt.Errorf("failed to parse form: %w", err))
			return
		}

		totpToken := r.FormValue("totp")

		totpToken, _ = strings.CutSuffix(totpToken, "]")
		totpToken, _ = strings.CutPrefix(totpToken, "[")

		err = validateIdentityTOTP(config, identity, totpToken)
		if err != nil {
			identityLoginErrorHTMX(w, r, "Invalid TOTP token", fmt.Errorf("invalid TOTP token: %w", err))
			return
		}

		// TOTP is valid, generate the real JWT token
		tokenString, expiresAt, err := config.Security.GenerateUserJWTToken(siteContext.Subdomain, identity.Name, string(identity.Role))
		if err != nil {
			identityLoginErrorHTMX(w, r, "", fmt.Errorf("failed to generate JWT token: %w", err))
			return
		}

		// Clear the pre-auth token cookie
		config.Security.ClearUserPreAuthCookie(w)

		// Set proper cookie and redirect to dashboard
		config.Security.SetUserLoginCookie(w, tokenString, expiresAt)

		slog.Info("User logged in", "identity", identity.Name, "role", identity.Role)

		// Just render the dashboard HTMX here

		w.Header().Set("HX-Redirect", "/")
		w.WriteHeader(http.StatusOK)
	}
}

func getTheIdentityeModel(controlDB *db.ControlDB, r *http.Request, identityNameFromToken string) (*control.SiteIdentity, error) {
	// First try the JWT token (either header or cookie)
	siteContext := httputils.GetSiteContext(r.Context())

	identity, err := controlDB.Queries.GetSiteIdentity(context.TODO(), control.GetSiteIdentityParams{
		SiteID: siteContext.Subdomain,
		Name:   identityNameFromToken,
	})
	if err != nil {
		return nil, fmt.Errorf("identity not found: %w", err)
	}
	return identity, nil
}

func loginTheUserPasswordOnly(controlDB *db.ControlDB, securityConfig *security.SecurityConfig, subdomain, formName, formPassword string) (*control.SiteIdentity, error) {
	if formName == "" || formPassword == "" {
		return nil, fmt.Errorf("device name and password are required")
	}

	identity, err := controlDB.Queries.GetSiteIdentity(context.TODO(), control.GetSiteIdentityParams{
		SiteID: subdomain,
		Name:   formName,
	})
	if err != nil {
		return nil, fmt.Errorf("identity not found: %w", err)
	}

	if !securityConfig.VerifyPassword(formPassword, identity.Password) {
		return nil, fmt.Errorf("invalid password")
	}

	return identity, nil
}

func validateIdentityTOTP(config *config.ServerConfig, identity *control.SiteIdentity, totpToken string) error {
	if totpToken == "" {
		return fmt.Errorf("TOTP token is required")
	}

	if identity.TotpSecret == "" {
		return fmt.Errorf("identity does not have a TOTP secret set")
	}

	totpSecret, err := config.Security.Decrypt(identity.TotpSecret)
	if err != nil {
		return err
	}

	if !config.Security.TOTPValidate(totpToken, totpSecret) {
		return fmt.Errorf("invalid TOTP token")
	}

	return nil
}

func fetchVerifyPreAuthToken(config *config.ServerConfig, r *http.Request) (string, error) {
	preAuthToken, err := security.GetUserPreAuthTokenFromRequest(r)
	if err != nil {
		return "", fmt.Errorf("Pre-auth token is required: %w", err)
	}

	identityName, err := config.Security.VerifyPreAuthJWTToken(preAuthToken)
	if err != nil {
		return "", fmt.Errorf("Invalid pre-auth token: %w", err)
	}

	return identityName, nil
}

// ///
// API Login functions
// ///
func PostLoginFromAPI(controlDB *db.ControlDB, config *config.ServerConfig, r *http.Request, w http.ResponseWriter) {
	siteContext := httputils.GetSiteContext(r.Context())

	var creds LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		identityLoginErrorAPI(w, r, "Invalid login attempt", fmt.Errorf("invalid login attempt: %w", err))
		return
	}

	// Rate Limit login attempts by user and IP
	allowedToAttempt, _ := httputils.SiteLoginAttemptsAllowed(config, r, siteContext.Subdomain, creds.Name)
	if !allowedToAttempt {
		http.Error(w, "Too many login attempts Wait a bit", http.StatusTooManyRequests)
		return
	}

	identity, err := loginTheUserAPI(siteContext.ControlDB, siteContext.Subdomain, config, creds.Name, creds.Password, creds.TOTP)
	if err != nil {
		identityLoginErrorAPI(w, r, "Invalid login attempt", fmt.Errorf("invalid login attempt: %w", err))
		return
	}

	tokenString, expiresAt, err := config.Security.GenerateUserJWTToken(siteContext.Subdomain, identity.Name, string(identity.Role))
	if err != nil {
		identityLoginErrorAPI(w, r, "Invalid login attempt", fmt.Errorf("invalid login attempt: %w", err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(map[string]string{
		"token":      tokenString,
		"expires_at": expiresAt.Format(time.RFC3339),
	})
	if err != nil {
		identityLoginErrorAPI(w, r, "Invalid login attempt", fmt.Errorf("invalid login attempt: %w", err))
		return
	}

	slog.Info("User logged in", "identity", identity.Name, "role", identity.Role)

}

func loginTheUserAPI(controlDB *db.ControlDB, subdomain string, cfg *config.ServerConfig, devicename, password, totpToken string) (*control.SiteIdentity, error) {
	device, err := loginTheUserPasswordOnly(controlDB, cfg.Security, subdomain, devicename, password)
	if err != nil {
		return nil, err
	}

	// An Identity only requires a TOTP secret if the role is admin
	if can.MustLoginWithTotp(device.Role, device.TotpSecret != "") {
		if totpToken == "" {
			return nil, fmt.Errorf("user requires TOTP secret")
		}
		totpSecret, err := cfg.Security.Decrypt(device.TotpSecret)
		if err != nil {
			return nil, err
		}
		if !cfg.Security.TOTPValidate(totpToken, totpSecret) {
			return nil, fmt.Errorf("invalid TOTP token")
		}
	}

	return device, nil
}

func rateLimitErrorHTMX(w http.ResponseWriter, r *http.Request) {
	slog.Info("Identity Login Failed", "stack", errors.LocalStack())

	w.WriteHeader(http.StatusTooManyRequests)

	err := site_templ.LoginError("Too many attempts").Render(r.Context(), w)
	if err != nil {
		slog.Error("Failed to render login error template", "error", err)
	}

}

// We have a very generic error to not give any hints about how far through it was rejected
func identityLoginErrorHTMX(w http.ResponseWriter, r *http.Request, usrMsg string, err error) {
	allArgs := []any{
		"error", err,
		"stack", errors.LocalStack(),
	}
	slog.Info("Identity Login Failed", allArgs...)

	if usrMsg == "" {
		usrMsg = "Invalid login attempt"
	}
	w.WriteHeader(http.StatusUnauthorized)
	err = site_templ.LoginError(usrMsg).Render(r.Context(), w)
	if err != nil {
		slog.Error("Failed to render login error template", "error", err)
	}
}

// We have a very generic error to not give any hints about how far through it was rejected
func identityLoginErrorAPI(w http.ResponseWriter, r *http.Request, usrMsg string, err error) {
	allArgs := []any{
		"error", err,
		"stack", errors.LocalStack(),
	}
	slog.Info("Identity Login Failed", allArgs...)

	if usrMsg == "" {
		usrMsg = "Invalid login attempt"
	}
	http.Error(w, usrMsg, http.StatusUnauthorized)
}
