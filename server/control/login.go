package control

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"nighttroop.t1d.nz/data/db"
	controlsql "nighttroop.t1d.nz/data/sql/control"
	controlTemplates "nighttroop.t1d.nz/server/assets/templates/control"
	"nighttroop.t1d.nz/utils"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/httputils"
)

const TOKEN_COOKIE = "control_token"

type LoginRequest struct {
	Name     string `json:"name"`
	Password string `json:"password"`
	TOTP     string `json:"totp"`
}

func GETLoginPageHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//
		w.Header().Set("Content-Type", "text/html")
		err := controlTemplates.Login(w)
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
				http.Error(w, "Invalid request payload", http.StatusBadRequest)
				return
			}
			username = creds.Name
			password = creds.Password
			totpToken = creds.TOTP
		case httputils.ContentTypeForm:

			if err := r.ParseForm(); err != nil {
				http.Error(w, "Invalid form data", http.StatusBadRequest)
				return
			}

			username = r.FormValue("name")
			password = r.FormValue("password")
			totpToken = r.FormValue("totp")
		default:
			http.Error(w, "Unsupported Content-Type", http.StatusUnsupportedMediaType)
			return
		}

		admin, err := loginTheControlUser(controlDB, config, username, password, totpToken)
		if err != nil {

			http.Error(w, "Invalid credentials "+err.Error(), http.StatusUnauthorized)
			return
		}

		tokenString, expiresAt, err := config.GenerateControlJWTToken(admin.Name, admin.Role)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		acceptType := httputils.GetAcceptType(r)

		switch acceptType {
		case httputils.AcceptTypeHTMX, httputils.AcceptTypeHTML:
			// Default set the cookie
			http.SetCookie(w, &http.Cookie{
				Name:     TOKEN_COOKIE,
				Value:    tokenString,
				Expires:  expiresAt,
				HttpOnly: true,
				Secure:   config.IsHTTPS(), // Set to true if using HTTPS
				Path:     "/",
				SameSite: http.SameSiteStrictMode, // Going Strict here because we shouldnt have external links to the control plane
			})

			fmt.Println("Login successful, redirecting to dashboard...")
			w.Header().Set("HX-Redirect", "/")
			w.WriteHeader(http.StatusOK)

			return
		default:
			// Default to JSON because this is API first
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{
				"token":      tokenString,
				"expires_at": expiresAt.Format(time.RFC3339),
			})
			return
		}
	}
}

func loginTheControlUser(controlDB *db.ControlDB, config config.ServerConfig, username, password, totpToken string) (*controlsql.AdminsGetModel, error) {
	admin, err := controlDB.AdminsGet.QueryRow(context.TODO(), username)
	if err != nil {
		return nil, fmt.Errorf("admin not found: %w", err)
	}
	totpSecret, err := config.Decrypt(admin.TotpSecret)
	if err != nil {
		return nil, err
	}

	if !utils.VerifyPassword(password, admin.Password) {
		return nil, fmt.Errorf("invalid password")
	}
	if !config.TOTPValidate(totpToken, totpSecret) {
		return nil, fmt.Errorf("invalid TOTP token")
	}
	return &admin, nil
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
