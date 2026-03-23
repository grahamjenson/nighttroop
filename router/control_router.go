package router

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/pquerna/otp/totp"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/utils"
	"nighttroop.t1d.nz/utils/config"
)

func NewControlRouter(controlDB *db.ControlDB, config config.ServerConfig) chi.Router {
	r := chi.NewRouter()

	totp.Validate("", "") // TODO
	r.Post("/token", generateJWTToken(controlDB, config))

	r.Group(func(r chi.Router) {
		r.Use(JWTAuthMiddleware(controlDB, config))
		r.Get("/users", getUsersHandler(controlDB, config))
	})
	return r
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	TOTP     string `json:"totp"`
}

// curl -v -X POST http://control.localhost:8080/token -d '
// {
//   "username": "admin",
//   "password": "QHIVFBRPXAXBAJZ3VD2BDMCGLAYT3S2D",
//   "totp": "123123"
// }'

func generateJWTToken(controlDB *db.ControlDB, config config.ServerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// This handler will read the passed username and password and totp token an generate a JWT token if the credentials are correct
		var creds LoginRequest

		// 1. Decode the JSON request body
		if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		// 2. Authenticate Username/Password/TOTP
		// Assuming controlDB has a method like GetUserByName
		admin, err := controlDB.AdminsGet.QueryRow(context.TODO(), creds.Username)

		totpSecret, err := config.Decrypt(admin.TotpSecret)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		err = loginTheControlUser(creds.Username, creds.Password, creds.TOTP, admin.Password, totpSecret)

		if err != nil {
			http.Error(w, "Invalid credentials "+err.Error(), http.StatusUnauthorized)
			return
		}

		tokenString, err := config.GenerateControlJWTToken(admin.Name)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"token": tokenString,
		})

	}
}

func loginTheControlUser(username, password, totpToken, passwordHash, totpSecret string) error {
	if !utils.VerifyPassword(password, passwordHash) {
		return fmt.Errorf("invalid password")
	}
	if !totp.Validate(totpToken, totpSecret) {
		return fmt.Errorf("invalid TOTP token")
	}
	return nil
}

func getUsersHandler(controlDB *db.ControlDB, config config.ServerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		users, err := controlDB.UsersAll.Query(context.TODO())
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(users)
	}
}

// //
// Middleware
// //

func JWTAuthMiddleware(controlDB *db.ControlDB, config config.ServerConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

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
