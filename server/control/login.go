package control

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"nighttroop.t1d.nz/data/db"
	controlsql "nighttroop.t1d.nz/data/sql/control"
	"nighttroop.t1d.nz/server/assets/templates"
	"nighttroop.t1d.nz/utils"
	"nighttroop.t1d.nz/utils/config"
)

const TOKEN_COOKIE = "control_token"

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	TOTP     string `json:"totp"`
}

func GETLoginPageHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//
		w.Header().Set("Content-Type", "text/html")
		err := templates.ControlLogin(w)
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
		contentType := utils.GetContentType(r)

		username := ""
		password := ""
		totpToken := ""

		switch contentType {
		case utils.ContentTypeJSON:
			var creds LoginRequest
			if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
				http.Error(w, "Invalid request payload", http.StatusBadRequest)
				return
			}
			username = creds.Username
			password = creds.Password
			totpToken = creds.TOTP
		case utils.ContentTypeForm:

			if err := r.ParseForm(); err != nil {
				http.Error(w, "Invalid form data", http.StatusBadRequest)
				return
			}

			username = r.FormValue("username")
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

		tokenString, expiresAt, err := config.GenerateControlJWTToken(admin.Name)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		acceptType := utils.GetAcceptType(r)

		switch acceptType {
		case utils.AcceptTypeHTMX, utils.AcceptTypeHTML:
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

			w.Header().Set("Content-Type", "text/html")

			fmt.Println("Login successful, redirecting to dashboard...")
			http.Redirect(w, r, "/", http.StatusSeeOther)

			return
		default:
			// Default to JSON because this is API first
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{
				"token": tokenString,
			})
			return
		}
	}
}

func loginTheControlUser(controlDB *db.ControlDB, config config.ServerConfig, username, password, totpToken string) (*controlsql.AdminsGetModel, error) {
	admin, err := controlDB.AdminsGet.QueryRow(context.TODO(), username)

	totpSecret, err := config.Decrypt(admin.TotpSecret)
	if err != nil {
		return nil, err
	}

	if !utils.VerifyPassword(password, admin.Password) {
		return nil, fmt.Errorf("invalid password")
	}
	if !utils.TOTPValidate(totpToken, totpSecret) {
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

// //
// Middleware
// //

func JWTAuthMiddleware(controlDB *db.ControlDB, config config.ServerConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			contentType := utils.GetAcceptType(r)
			tokenStr := ""
			switch contentType {
			case utils.AcceptTypeJSON:
				authHeader := r.Header.Get("Authorization")
				if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
					http.Error(w, "Unauthorized Header", http.StatusUnauthorized)
					return
				}

				tokenStr = strings.TrimPrefix(authHeader, "Bearer ")
			default:
				cookie, err := r.Cookie(TOKEN_COOKIE)
				if err != nil {
					fmt.Println("Control Router: No token cookie found, redirecting to login")
					http.Redirect(w, r, "/login", http.StatusSeeOther)
					return
				}
				tokenStr = cookie.Value
			}

			adminName, err := config.VerifyControlJWTToken(tokenStr)

			if err != nil {
				http.Error(w, "Bad Token", http.StatusInternalServerError)
				return
			}

			// Optional: Inject claims into context
			ctx := context.WithValue(r.Context(), "adminName", adminName)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
