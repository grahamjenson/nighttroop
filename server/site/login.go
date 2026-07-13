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

func POSTLogoutHandler(config config.ServerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		config.Security.ClearUserLoginCookie(w)

		// 3. Redirect the user to the login page
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}
}

func GETLoginPageHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//
		w.Header().Set("Content-Type", "text/html")
		err := site_templ.Login().Render(r.Context(), w)
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

func POSTLogin(controlDB *db.ControlDB, config config.ServerConfig) http.HandlerFunc {
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
			errors.GenericIdentityLoginError(w, r, err)
			return
		}

		formName := r.FormValue("name")
		formPassword := r.FormValue("password")

		// Rate Limit login attempts by user and IP

		allowedToAttempt, _ := config.RateLimiters.SiteLoginAllow(fmt.Sprintf("%s:%s", subdomain, formName), httputils.GetIP(r))
		if !allowedToAttempt {
			http.Error(w, "Too many login attempts Wait a bit", http.StatusTooManyRequests)
			return
		}

		// First thing validate the password
		identity, err := loginTheUserPasswordOnly(siteContext.ControlDB, siteContext.Subdomain, formName, formPassword)
		if err != nil {
			errors.GenericIdentityLoginError(w, r, err)
			return
		}

		if !can.LoginToWebsite(identity.Role) {
			errors.GenericIdentityLoginError(w, r, fmt.Errorf("identity does not have permission to login to website"))
			return
		}

		if can.MustLoginWithTotp(identity.Role, identity.TotpSecret != "") {
			// If the Identity Requires TOTP validation, replace the original form with a new form that submits to TOTP validation input
			err = config.Security.SetUserPreAuthCookie(w, identity.Name)
			if err != nil {
				errors.GenericIdentityLoginError(w, r, err)
				return
			}

			w.Header().Set("Content-Type", "text/html")
			err = site_templ.LoginTOTP().Render(r.Context(), w)
			if err != nil {
				errors.GenericIdentityLoginError(w, r, err)
				return
			}
			return
		} else {
			tokenString, expiresAt, err := config.Security.GenerateUserJWTToken(siteContext.Subdomain, identity.Name, string(identity.Role))
			if err != nil {
				errors.GenericIdentityLoginError(w, r, err)
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
func POSTLoginTOTP(controlDB *db.ControlDB, config config.ServerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if httputils.GetAcceptType(r).API() {
			errors.GenericIdentityLoginError(w, r, fmt.Errorf("API login not supported for TOTP endpoint"))
			return
		}
		siteContext := httputils.GetSiteContext(r.Context())

		// 1. Fetch And Validate Pre-Auth Token
		identityNameFromToken, err := fetchVerifyPreAuthToken(config, r)
		if err != nil {
			errors.GenericIdentityLoginError(w, r, err)
			return
		}

		// 2. Grab the user db and identity

		identity, err := getTheIdentityeModel(controlDB, r, identityNameFromToken)
		if err != nil {
			errors.GenericIdentityLoginError(w, r, err)
			return
		}

		// THIS IS A HTTP ONLY ROUTE
		if err := r.ParseForm(); err != nil {
			errors.GenericIdentityLoginError(w, r, err)
			return
		}

		totpToken := r.FormValue("totp")

		totpToken, _ = strings.CutSuffix(totpToken, "]")
		totpToken, _ = strings.CutPrefix(totpToken, "[")

		err = validateIdentityTOTP(config, identity, totpToken)
		if err != nil {
			errors.GenericIdentityLoginError(w, r, err)
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// TOTP is valid, generate the real JWT token
		tokenString, expiresAt, err := config.Security.GenerateUserJWTToken(siteContext.Subdomain, identity.Name, string(identity.Role))
		if err != nil {
			errors.GenericIdentityLoginError(w, r, err)
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Clear the pre-auth token cookie
		config.Security.ClearUserPreAuthCookie(w)

		// Set proper cookie and redirect to dashboard
		config.Security.SetUserLoginCookie(w, tokenString, expiresAt)

		slog.Info("User logged in", "identity", identity.Name, "role", identity.Role)

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func getTheIdentityeModel(controlDB *db.ControlDB, r *http.Request, identityNameFromToken string) (*control.Identity, error) {
	// First try the JWT token (either header or cookie)
	siteContext := httputils.GetSiteContext(r.Context())

	identity, err := controlDB.Queries.GetIdentity(context.TODO(), control.GetIdentityParams{
		Subdomain: siteContext.Subdomain,
		Name:      identityNameFromToken,
	})
	if err != nil {
		return nil, fmt.Errorf("identity not found: %w", err)
	}
	return identity, nil
}

func loginTheUserPasswordOnly(controlDB *db.ControlDB, subdomain, formName, formPassword string) (*control.Identity, error) {
	if formName == "" || formPassword == "" {
		return nil, fmt.Errorf("device name and password are required")
	}

	identity, err := controlDB.Queries.GetIdentity(context.TODO(), control.GetIdentityParams{
		Subdomain: subdomain,
		Name:      formName,
	})
	if err != nil {
		return nil, fmt.Errorf("identity not found: %w", err)
	}

	if !security.VerifyPassword(formPassword, identity.Password) {
		return nil, fmt.Errorf("invalid password")
	}

	return identity, nil
}

func validateIdentityTOTP(config config.ServerConfig, identity *control.Identity, totpToken string) error {
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

func fetchVerifyPreAuthToken(config config.ServerConfig, r *http.Request) (string, error) {
	preAuthToken, err := security.GetUserPreAuthTokenFromRequest(r)
	if err != nil {
		return "", fmt.Errorf("Pre-auth token is required: %w", err)
	}

	deviceName, err := config.Security.VerifyPreAuthJWTToken(preAuthToken)
	if err != nil {
		return "", fmt.Errorf("Invalid pre-auth token: %w", err)
	}

	return deviceName, nil
}

// ///
// API Login functions
// ///
func PostLoginFromAPI(controlDB *db.ControlDB, config config.ServerConfig, r *http.Request, w http.ResponseWriter) {
	siteContext := httputils.GetSiteContext(r.Context())

	var creds LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		errors.GenericIdentityLoginError(w, r, err)
		return
	}

	// Rate Limit login attempts by user and IP
	allowedToAttempt, _ := config.RateLimiters.SiteLoginAllow(fmt.Sprintf("%s:%s", siteContext.Subdomain, creds.Name), httputils.GetIP(r))
	if !allowedToAttempt {
		http.Error(w, "Too many login attempts Wait a bit", http.StatusTooManyRequests)
		return
	}

	identity, err := loginTheUserAPI(siteContext.ControlDB, siteContext.Subdomain, config, creds.Name, creds.Password, creds.TOTP)
	if err != nil {
		errors.GenericIdentityLoginError(w, r, err)
		return
	}

	tokenString, expiresAt, err := config.Security.GenerateUserJWTToken(siteContext.Subdomain, identity.Name, string(identity.Role))
	if err != nil {
		errors.GenericIdentityLoginError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(map[string]string{
		"token":      tokenString,
		"expires_at": expiresAt.Format(time.RFC3339),
	})
	if err != nil {
		errors.GenericIdentityLoginError(w, r, err)
		return
	}

	slog.Info("User logged in", "identity", identity.Name, "role", identity.Role)

}

func loginTheUserAPI(controlDB *db.ControlDB, subdomain string, cfg config.ServerConfig, devicename, password, totpToken string) (*control.Identity, error) {
	device, err := loginTheUserPasswordOnly(controlDB, subdomain, devicename, password)
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
