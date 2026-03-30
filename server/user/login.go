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
	"nighttroop.t1d.nz/utils"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/httputils"
)

const TOKEN_COOKIE = "user_token"
const PRE_AUTH_TOKEN_COOKIE = "pre_auth_token"

type LoginRequest struct {
	Name     string `json:"name"`
	Password string `json:"password"`
	TOTP     string `json:"totp"`
}

func GETLoginPageHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//
		w.Header().Set("Content-Type", "text/html")
		err := userTemplates.Login(w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
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
		// is this HTML FORM or JSON?
		contentType := httputils.GetContentType(r)

		// API login is different enough to just immediatly bail on finding common code
		if contentType == httputils.ContentTypeJSON {
			postLoginFromAPI(controlDB, config, r, w)
			return
		}

		subdomain := httputils.GetSubdomainFromHTTPContext(r.Context())
		userDB, _, err := controlDB.GetUser(subdomain, false)
		if err != nil {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}

		devicename := ""
		password := ""

		if err := r.ParseForm(); err != nil {
			http.Error(w, "Invalid form data", http.StatusBadRequest)
			return
		}

		devicename = r.FormValue("name")
		password = r.FormValue("password")

		device, err := loginTheUserPasswordOnly(userDB, devicename, password)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid credentials with %s", err.Error()), http.StatusUnauthorized)
			return
		}
		if deviceRequiresTOTPValidation(device) {
			// set a temporary token
			tempToken, expiresAt, err := config.GeneratePreAuthJWTToken(device.Name)
			http.SetCookie(w, &http.Cookie{
				Name:     PRE_AUTH_TOKEN_COOKIE,
				Value:    tempToken,
				Expires:  expiresAt,
				HttpOnly: true,
				Secure:   config.IsHTTPS(), // Set to true if using HTTPS
				Path:     "/",
				SameSite: http.SameSiteStrictMode, // Going Strict here because we shouldnt have external links to the control plane
			})

			w.Header().Set("Content-Type", "text/html")
			err = userTemplates.LoginTotp(w)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			return
		}

		tokenString, expiresAt, err := config.GenerateUserJWTToken(device.Name, device.Role)
		if err != nil {
			http.Error(w, "Internal server error Token", http.StatusInternalServerError)
			return
		}

		// Set proper cookie and redirect to dashboard
		http.SetCookie(w, &http.Cookie{
			Name:     TOKEN_COOKIE,
			Value:    tokenString,
			Expires:  expiresAt,
			HttpOnly: true,
			Secure:   config.IsHTTPS(), // Set to true if using HTTPS
			Path:     "/",
			SameSite: http.SameSiteStrictMode, // Going Strict here because we shouldnt have external links to the control plane
		})

		fmt.Println("Login successful")
		w.Header().Set("HX-Redirect", "/")
		w.WriteHeader(http.StatusOK)

	}
}

func POSTLoginTOTP(controlDB *db.ControlDB, config config.ServerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		subdomain := httputils.GetSubdomainFromHTTPContext(r.Context())
		userDB, _, err := controlDB.GetUser(subdomain, false)
		if err != nil {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}

		// THIS IS A HTTP ONLY ROUTE
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Invalid form data", http.StatusBadRequest)
			return
		}

		totpToken := r.FormValue("totp")
		if totpToken == "" {
			http.Error(w, "TOTP token is required", http.StatusBadRequest)
			return
		}

		preAuthCookie, err := r.Cookie(PRE_AUTH_TOKEN_COOKIE)
		if err != nil {
			http.Error(w, "Pre-auth token not found", http.StatusUnauthorized)
			return
		}

		preAuthToken := preAuthCookie.Value

		deviceName, err := config.VerifyPreAuthJWTToken(preAuthToken)
		if err != nil {
			http.Error(w, "Invalid pre-auth token", http.StatusUnauthorized)
			return

		}

		device, err := userDB.DevicesGet.QueryRow(context.TODO(), deviceName)
		if err != nil {
			http.Error(w, "Device not found", http.StatusNotFound)
			return
		}

		totpSecret, err := config.Decrypt(device.TotpSecret)
		if err != nil {
			http.Error(w, "Error decrypting TOTP secret", http.StatusInternalServerError)
			return
		}

		if !config.TOTPValidate(totpToken, totpSecret) {
			http.Error(w, "Invalid TOTP token", http.StatusUnauthorized)
			return
		}

		// TOTP is valid, generate the real JWT token
		tokenString, expiresAt, err := config.GenerateUserJWTToken(device.Name, device.Role)
		if err != nil {
			http.Error(w, "Internal server error Token", http.StatusInternalServerError)
			return
		}

		// Set proper cookie and redirect to dashboard
		http.SetCookie(w, &http.Cookie{
			Name:     TOKEN_COOKIE,
			Value:    tokenString,
			Expires:  expiresAt,
			HttpOnly: true,
			Secure:   config.IsHTTPS(), // Set to true if using HTTPS
			Path:     "/",
			SameSite: http.SameSiteStrictMode, // Going Strict here because we shouldnt have external links to the control plane
		})

		// Clear the pre-auth token cookie
		http.SetCookie(w, &http.Cookie{
			Name:     PRE_AUTH_TOKEN_COOKIE,
			Value:    "",
			MaxAge:   -1,
			Expires:  time.Unix(0, 0),
			HttpOnly: true,
			Secure:   config.IsHTTPS(), // Set to true if using HTTPS
		})

		fmt.Println("TOTP verification successful, user logged in")
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func postLoginFromAPI(controlDB *db.ControlDB, config config.ServerConfig, r *http.Request, w http.ResponseWriter) {
	var creds LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	subdomain := httputils.GetSubdomainFromHTTPContext(r.Context())
	userDB, _, err := controlDB.GetUser(subdomain, false)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	device, err := loginTheUserAPI(userDB, config, creds.Name, creds.Password, creds.TOTP)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid credentials with %s", err.Error()), http.StatusUnauthorized)
		return
	}

	tokenString, expiresAt, err := config.GenerateUserJWTToken(device.Name, device.Role)
	if err != nil {
		http.Error(w, "Internal server error Token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"token":      tokenString,
		"expires_at": expiresAt.Format(time.RFC3339),
	})

}

func loginTheUserPasswordOnly(userDB *db.UserDB, devicename, password string) (*usersql.DevicesGetModel, error) {
	device, err := userDB.DevicesGet.QueryRow(context.TODO(), devicename)
	if err != nil {
		return nil, fmt.Errorf("device not found: %w", err)
	}

	if !utils.VerifyPassword(password, device.Password) {
		return nil, fmt.Errorf("invalid password")
	}

	return &device, nil
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

	totpSecret, err := cfg.Decrypt(device.TotpSecret)
	if err != nil {
		return nil, err
	}
	if !cfg.TOTPValidate(totpToken, totpSecret) {
		return nil, fmt.Errorf("invalid TOTP token")
	}

	return device, nil
}

func deviceRequiresTOTPValidation(device *usersql.DevicesGetModel) bool {
	if device.Role == "admin" {
		return true
	}
	if device.TotpSecret != "" {
		return true
	}
	return false
}
func GETLogoutHandler(config config.ServerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		fmt.Println("Render Logout")
		http.SetCookie(w, &http.Cookie{
			Name:     TOKEN_COOKIE,
			Value:    "",
			MaxAge:   -1,
			Expires:  time.Unix(0, 0),
			HttpOnly: true,
			Secure:   config.IsHTTPS(), // Set to true if using HTTPS
			Path:     "/",
			SameSite: http.SameSiteStrictMode, // Going Strict here because we shouldnt have external links to the control plane
		})

		// 3. Redirect the user to the login page
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}
}
