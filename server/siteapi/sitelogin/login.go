package sitelogin

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"nighttroop.t1d.nz/data/db"

	"nighttroop.t1d.nz/data/sql/controldb"

	"nighttroop.t1d.nz/server/components/errortoast"
	"nighttroop.t1d.nz/server/components/site_templ/site_login_templ"
	"nighttroop.t1d.nz/utils/apiutils"
	"nighttroop.t1d.nz/utils/can"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/errors"
	"nighttroop.t1d.nz/utils/httputils"
	"nighttroop.t1d.nz/utils/security"
)

// LoginRequest is an identity's credentials. Cross-site requests cannot log in or act as an identity, as its cookies are
// SameSite=Strict, and its fields only ever reach the DB through prepared statements.
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
		err := site_login_templ.Login(loginName(r)).Render(r.Context(), w)
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
		siteID := siteContext.SiteID

		if err := r.ParseForm(); err != nil {
			identityLoginErrorHTMX(w, r, "Failed to parse", fmt.Errorf("failed to parse form: %w", err))
			return
		}

		formName := r.FormValue("name")
		formPassword := r.FormValue("password")

		// Reject names no identity can have before they reach the rate limiter, which keys on them
		if utf8.RuneCountInString(formName) > consts.MAX_IDENTITY_NAME_LENGTH {
			identityLoginErrorHTMX(w, r, "Invalid username or password.", fmt.Errorf("username longer than %d characters", consts.MAX_IDENTITY_NAME_LENGTH))
			return
		}

		// Rate Limit login attempts by user and IP
		allowedToAttempt, _ := httputils.SiteLoginAttemptsWithinRateLimit(config, r, siteID, formName)
		if !allowedToAttempt {
			rateLimitErrorHTMX(w, r)
			return
		}

		// First thing validate the password
		identity, err := loginTheUserPasswordOnly(r.Context(), siteContext.ControlDB, config.Security, siteContext.SiteID, formName, formPassword)
		if err != nil {
			identityLoginErrorHTMX(w, r, "Invalid username or password.", fmt.Errorf("failed to login with password: %w", err))
			return
		}

		if !can.LoginToWebsite(identity.Role) {
			identityLoginErrorHTMX(w, r, "Identity does not have permission to login to website", fmt.Errorf("identity with role %s cannot login to website", identity.Role))
			return
		}

		if can.MustLoginWithTotp(identity.Role, identity.TotpSecretEnc != "") {
			// If the Identity Requires TOTP validation, replace the original form with a new form that submits to TOTP validation input
			err = config.Security.SetUserPreAuthCookie(w, siteContext.SiteID, identity.Name)
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
			tokenString, expiresAt, err := config.Security.GenerateUserJWTToken(siteContext.SiteID, identity.Name, string(identity.Role))
			if err != nil {
				identityLoginErrorHTMX(w, r, "", fmt.Errorf("failed to generate JWT token: %w", err))
				return
			}

			// Set proper cookie and redirect to dashboard
			config.Security.SetUserLoginCookie(w, tokenString, expiresAt)

			slog.InfoContext(r.Context(), "User logged in", "device", identity.Name, "role", identity.Role)
			httputils.RecordAuthClient(r, identity)
			recordSignIn(r, identity)
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
		identityNameFromToken, issuedAtMs, err := fetchVerifyPreAuthToken(config, r, siteContext.SiteID)
		if err != nil {
			identityLoginErrorHTMX(w, r, "", fmt.Errorf("failed to fetch and verify pre-auth token: %w", err))
			return
		}

		allowedToAttempt, _ := httputils.SiteLoginAttemptsWithinRateLimit(config, r, siteContext.SiteID, identityNameFromToken)
		if !allowedToAttempt {
			rateLimitErrorHTMX(w, r)
			return
		}

		// 2. Grab the user db and identity

		identity, err := getTheIdentityeModel(controlDB, r, identityNameFromToken)
		if err == nil {
			// The password checked for the pre-auth token may have been reset since
			err = httputils.CheckIssuedAfterCredentialsChanged(identity, issuedAtMs)
		}
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
		tokenString, expiresAt, err := config.Security.GenerateUserJWTToken(siteContext.SiteID, identity.Name, string(identity.Role))
		if err != nil {
			identityLoginErrorHTMX(w, r, "", fmt.Errorf("failed to generate JWT token: %w", err))
			return
		}

		// Clear the pre-auth token cookie
		config.Security.ClearUserPreAuthCookie(w)

		// Set proper cookie and redirect to dashboard
		config.Security.SetUserLoginCookie(w, tokenString, expiresAt)

		slog.InfoContext(r.Context(), "User logged in", "identity", identity.Name, "role", identity.Role)
		httputils.RecordAuthClient(r, identity)
		recordSignIn(r, identity)

		// Just render the dashboard HTMX here

		w.Header().Set("HX-Redirect", "/")
		w.WriteHeader(http.StatusOK)
	}
}

func getTheIdentityeModel(controlDB *db.ControlDB, r *http.Request, identityNameFromToken string) (*controldb.SiteIdentity, error) {
	// First try the JWT token (either header or cookie)
	siteContext := httputils.GetSiteContext(r.Context())

	identity, err := controlDB.Queries.GetSiteIdentity(r.Context(), controldb.GetSiteIdentityParams{
		SiteID: siteContext.SiteID,
		Name:   identityNameFromToken,
	})
	if err != nil {
		return nil, fmt.Errorf("identity not found: %w", err)
	}
	// It may have been disabled since the password step
	if !identity.Enabled {
		return nil, fmt.Errorf("identity disabled")
	}
	return identity, nil
}

func loginTheUserPasswordOnly(ctx context.Context, controlDB *db.ControlDB, securityConfig *security.SecurityConfig, subdomain, formName, formPassword string) (*controldb.SiteIdentity, error) {
	if formName == "" || formPassword == "" {
		return nil, fmt.Errorf("device name and password are required")
	}

	identity, err := controlDB.Queries.GetSiteIdentity(ctx, controldb.GetSiteIdentityParams{
		SiteID: subdomain,
		Name:   formName,
	})
	if err != nil {
		// Checked anyway, so a name with no identity is refused as slowly as a wrong password
		securityConfig.VerifyNoOnesPassword(formPassword)
		return nil, fmt.Errorf("identity not found: %w", err)
	}

	if !securityConfig.VerifyPassword(formPassword, identity.PasswordHash) {
		return nil, fmt.Errorf("invalid password")
	}

	// Checked after the password, so a failed login does not reveal which identities are disabled
	if !identity.Enabled {
		return nil, fmt.Errorf("identity disabled")
	}

	return identity, nil
}

func validateIdentityTOTP(config *config.ServerConfig, identity *controldb.SiteIdentity, totpToken string) error {
	if totpToken == "" {
		return fmt.Errorf("TOTP token is required")
	}

	if identity.TotpSecretEnc == "" {
		return fmt.Errorf("identity does not have a TOTP secret set")
	}

	totpSecret, err := config.Security.Decrypt(identity.TotpSecretEnc)
	if err != nil {
		return err
	}

	if !config.Security.TOTPSignIn(fmt.Sprintf("site %s %d", identity.SiteID, identity.ID), totpToken, totpSecret) {
		return fmt.Errorf("invalid TOTP token")
	}

	return nil
}

func fetchVerifyPreAuthToken(config *config.ServerConfig, r *http.Request, subdomain string) (string, int64, error) {
	preAuthToken, err := security.GetUserPreAuthTokenFromRequest(r)
	if err != nil {
		return "", 0, fmt.Errorf("Pre-auth token is required: %w", err)
	}

	identityName, issuedAtMs, err := config.Security.VerifyPreAuthJWTToken(preAuthToken, subdomain)
	if err != nil {
		return "", 0, fmt.Errorf("Invalid pre-auth token: %w", err)
	}

	return identityName, issuedAtMs, nil
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

	// Reject names no identity can have before they reach the rate limiter, which keys on them
	if utf8.RuneCountInString(creds.Name) > consts.MAX_IDENTITY_NAME_LENGTH {
		identityLoginErrorAPI(w, r, "Invalid login attempt", fmt.Errorf("username longer than %d characters", consts.MAX_IDENTITY_NAME_LENGTH))
		return
	}

	// Rate Limit login attempts by user and IP
	allowedToAttempt, _ := httputils.SiteLoginAttemptsWithinRateLimit(config, r, siteContext.SiteID, creds.Name)
	if !allowedToAttempt {
		http.Error(w, tooManyAttempts, http.StatusTooManyRequests)
		return
	}

	identity, err := loginTheUserAPI(r.Context(), siteContext.ControlDB, siteContext.SiteID, config, creds.Name, creds.Password, creds.TOTP)
	if err != nil {
		identityLoginErrorAPI(w, r, "Invalid login attempt", fmt.Errorf("invalid login attempt: %w", err))
		return
	}

	tokenString, expiresAt, err := config.Security.GenerateUserJWTToken(siteContext.SiteID, identity.Name, string(identity.Role))
	if err != nil {
		identityLoginErrorAPI(w, r, "Invalid login attempt", fmt.Errorf("invalid login attempt: %w", err))
		return
	}

	apiutils.WriteJSON(w, http.StatusOK, map[string]string{
		"token":      tokenString,
		"expires_at": expiresAt.Format(time.RFC3339),
	})

	slog.InfoContext(r.Context(), "User logged in", "identity", identity.Name, "role", identity.Role)
	recordSignIn(r, identity)
}

// recordSignIn records a sign in to the site, or a refused one when there is no identity, with where it is from
func recordSignIn(r *http.Request, identity *controldb.SiteIdentity) {
	siteContext := httputils.GetSiteContext(r.Context())
	if siteContext == nil || siteContext.ControlDB == nil {
		return
	}
	from := siteContext.Config.GeoIP.Locate(siteContext.IPAddress)
	if identity == nil {
		siteContext.ControlDB.Logs.SignInFailedEvent(r.Context(), siteContext.SiteID, from)
		return
	}
	siteContext.ControlDB.Logs.SignedInEvent(r.Context(), siteContext.SiteID, identity.Name, from)
}

func loginTheUserAPI(ctx context.Context, controlDB *db.ControlDB, subdomain string, cfg *config.ServerConfig, devicename, password, totpToken string) (*controldb.SiteIdentity, error) {
	device, err := loginTheUserPasswordOnly(ctx, controlDB, cfg.Security, subdomain, devicename, password)
	if err != nil {
		return nil, err
	}

	// The API gives the same site token as the website, so only the roles that may sign in to the website get one
	if !can.LoginToWebsite(device.Role) {
		return nil, fmt.Errorf("identity with role %s cannot login to website", device.Role)
	}

	// An Identity only requires a TOTP secret if the role is admin
	if can.MustLoginWithTotp(device.Role, device.TotpSecretEnc != "") {
		if err := validateIdentityTOTP(cfg, device, totpToken); err != nil {
			return nil, err
		}
	}

	return device, nil
}

// tooManyAttempts is what a login refused by the rate limiter says, on the page and to the API
const tooManyAttempts = "Too many login attempts. Wait a minute, then try again."

func rateLimitErrorHTMX(w http.ResponseWriter, r *http.Request) {
	slog.InfoContext(r.Context(), "Identity Login Failed", "stack", errors.LocalStack())

	w.WriteHeader(http.StatusTooManyRequests)

	err := errortoast.Toast("", tooManyAttempts).Render(r.Context(), w)
	if err != nil {
		slog.ErrorContext(r.Context(), "Failed to render login error template", "error", err)
	}

}

// We have a very generic error to not give any hints about how far through it was rejected
func identityLoginErrorHTMX(w http.ResponseWriter, r *http.Request, usrMsg string, err error) {
	recordSignIn(r, nil)
	allArgs := []any{
		"error", err,
		"stack", errors.LocalStack(),
	}
	slog.InfoContext(r.Context(), "Identity Login Failed", allArgs...)

	if usrMsg == "" {
		usrMsg = "Invalid login attempt"
	}
	w.WriteHeader(http.StatusUnauthorized)
	err = errortoast.Toast("", usrMsg).Render(r.Context(), w)
	if err != nil {
		slog.ErrorContext(r.Context(), "Failed to render login error template", "error", err)
	}
}

// We have a very generic error to not give any hints about how far through it was rejected
func identityLoginErrorAPI(w http.ResponseWriter, r *http.Request, usrMsg string, err error) {
	recordSignIn(r, nil)
	allArgs := []any{
		"error", err,
		"stack", errors.LocalStack(),
	}
	slog.InfoContext(r.Context(), "Identity Login Failed", allArgs...)

	if usrMsg == "" {
		usrMsg = "Invalid login attempt"
	}
	http.Error(w, usrMsg, http.StatusUnauthorized)
}

// maxLoginName is the longest an identity's name is, as its table checks
const maxLoginName = 32

// loginName is the name a login link fills in, such as the new site's admin from its welcome page, at most an
// identity's name long
func loginName(r *http.Request) string {
	if name := r.URL.Query().Get("name"); len(name) <= maxLoginName {
		return name
	}
	return ""
}
