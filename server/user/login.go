package user

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"nighttroop.t1d.nz/data/db"
	usersql "nighttroop.t1d.nz/data/sql/user"
	userTemplates "nighttroop.t1d.nz/server/assets/templates/user"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/consts"
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
		subdomain := httputils.GetSubdomainFromHTTPContext(r.Context())
		err := userTemplates.LoginHtml(w, subdomain)
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

		subdomain := httputils.GetSubdomainFromHTTPContext(r.Context())
		userDB, _, err := controlDB.GetUser(subdomain, false)
		if err != nil {
			genericLoginError(w, r, err)
			return
		}

		if err := r.ParseForm(); err != nil {
			genericLoginError(w, r, err)
			return
		}

		formName := r.FormValue("name")
		formPassword := r.FormValue("password")

		// First thing validate the password
		device, err := loginTheUserPasswordOnly(userDB, formName, formPassword)
		if err != nil {
			genericLoginError(w, r, err)
			return
		}

		if deviceRequiresTOTPValidation(device) {
			// If the Device Requires TOTP validation, replace the original form with a new form that submits to TOTP validation input
			err = config.Security.SetUserPreAuthCookie(w, device.Name)
			if err != nil {
				genericLoginError(w, r, err)
				return
			}

			w.Header().Set("Content-Type", "text/html")
			err = userTemplates.LoginTotpHtmx(w)
			if err != nil {
				genericLoginError(w, r, err)
				return
			}
			return
		} else {
			tokenString, expiresAt, err := config.Security.GenerateUserJWTToken(device.Name, device.Role)
			if err != nil {
				genericLoginError(w, r, err)
				return
			}

			// Set proper cookie and redirect to dashboard
			config.Security.SetUserLoginCookie(w, tokenString, expiresAt)

			w.Header().Set("HX-Redirect", "/")
			w.WriteHeader(http.StatusOK)
		}

	}
}

// Browser only route for validating TOTP tokens
func POSTLoginTOTP(controlDB *db.ControlDB, config config.ServerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if httputils.GetAcceptType(r).API() {
			genericLoginError(w, r, fmt.Errorf("API login not supported for TOTP endpoint"))
			return
		}

		// 1. Fetch And Validate Pre-Auth Token
		deviceNameFromToken, err := fetchVerifyPreAuthToken(config, r)
		if err != nil {
			genericLoginError(w, r, err)
			return
		}

		// 2. Grab the user db and device

		device, err := getTheDeviceModel(controlDB, r, deviceNameFromToken)
		if err != nil {
			genericLoginError(w, r, err)
			return
		}

		// THIS IS A HTTP ONLY ROUTE
		if err := r.ParseForm(); err != nil {
			genericLoginError(w, r, err)
			return
		}

		totpToken := r.FormValue("totp")
		err = validateDeviceTOTP(config, device, totpToken)
		if err != nil {
			genericLoginError(w, r, err)
			return
		}

		// TOTP is valid, generate the real JWT token
		tokenString, expiresAt, err := config.Security.GenerateUserJWTToken(device.Name, device.Role)
		if err != nil {
			genericLoginError(w, r, err)
			return
		}

		// Clear the pre-auth token cookie
		config.Security.ClearUserPreAuthCookie(w)

		// Set proper cookie and redirect to dashboard
		config.Security.SetUserLoginCookie(w, tokenString, expiresAt)

		fmt.Println("TOTP verification successful, user logged in")
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func getTheDeviceModel(controlDB *db.ControlDB, r *http.Request, deviceNameFromToken string) (*usersql.DevicesGetModel, error) {
	// First try the JWT token (either header or cookie)
	subdomain := httputils.GetSubdomainFromHTTPContext(r.Context())
	userDB, _, err := controlDB.GetUser(subdomain, false)
	if err != nil {

		return nil, err
	}

	device, err := userDB.DevicesGet.QueryRow(context.TODO(), deviceNameFromToken)
	if err != nil {
		return nil, fmt.Errorf("device not found: %w", err)
	}
	return &device, nil
}

func loginTheUserPasswordOnly(userDB *db.UserDB, formName, formPassword string) (*usersql.DevicesGetModel, error) {
	if formName == "" || formPassword == "" {
		return nil, fmt.Errorf("device name and password are required")
	}

	device, err := userDB.DevicesGet.QueryRow(context.TODO(), formName)
	if err != nil {
		return nil, fmt.Errorf("device not found: %w", err)
	}

	if !security.VerifyPassword(formPassword, device.Password) {
		return nil, fmt.Errorf("invalid password")
	}

	return &device, nil
}

func deviceRequiresTOTPValidation(device *usersql.DevicesGetModel) bool {
	// Constraint: TOTP validation is required for admin users
	if device.Role == consts.UserRoleAdmin.String() {
		return true
	}
	// Constraint TOTP validation is required for any user with a TOTP secret set (even if not admin)
	if device.TotpSecret != "" {
		return true
	}
	return false
}

func validateDeviceTOTP(config config.ServerConfig, device *usersql.DevicesGetModel, totpToken string) error {
	if totpToken == "" {
		return fmt.Errorf("TOTP token is required")
	}

	totpSecret, err := config.Security.Decrypt(device.TotpSecret)
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
	subdomain := httputils.GetSubdomainFromHTTPContext(r.Context())
	userDB, _, err := controlDB.GetUser(subdomain, false)
	if err != nil {
		genericLoginError(w, r, err)
		return
	}

	var creds LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		genericLoginError(w, r, err)
		return
	}

	device, err := loginTheUserAPI(userDB, config, creds.Name, creds.Password, creds.TOTP)
	if err != nil {
		genericLoginError(w, r, err)
		return
	}

	tokenString, expiresAt, err := config.Security.GenerateUserJWTToken(device.Name, device.Role)
	if err != nil {
		genericLoginError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(map[string]string{
		"token":      tokenString,
		"expires_at": expiresAt.Format(time.RFC3339),
	})
	if err != nil {
		genericLoginError(w, r, err)
		return
	}

}

func loginTheUserAPI(userDB *db.UserDB, cfg config.ServerConfig, devicename, password, totpToken string) (*usersql.DevicesGetModel, error) {
	device, err := loginTheUserPasswordOnly(userDB, devicename, password)
	if err != nil {
		return nil, err
	}

	// A Device only requires a TOTP secret if the role is admin
	if deviceRequiresTOTPValidation(device) && totpToken == "" {
		return nil, fmt.Errorf("User requires TOTP secret")
	}

	totpSecret, err := cfg.Security.Decrypt(device.TotpSecret)
	if err != nil {
		return nil, err
	}
	if !cfg.Security.TOTPValidate(totpToken, totpSecret) {
		return nil, fmt.Errorf("invalid TOTP token")
	}

	return device, nil
}
