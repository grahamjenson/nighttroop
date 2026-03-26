package server

import (
	"fmt"
	"net/http"
	"runtime"

	"github.com/go-chi/chi/v5"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/server/assets/static"
	"nighttroop.t1d.nz/server/assets/templates"
	"nighttroop.t1d.nz/utils"
	"nighttroop.t1d.nz/utils/config"
)

func NewPublicRouter(controlDB *db.ControlDB, config config.ServerConfig) chi.Router {
	r := chi.NewRouter()

	static.AddSatticRoutes(r)

	// Create New User Route
	// Documentation Route
	// Public Facing SEO route
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		err := templates.PublicCreateUser(w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	r.Post("/user", createUser(controlDB))
	return r
}

func GetCreateUserHandler(controlDB *db.ControlDB, config config.ServerConfig) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		allowed, remainingTime := config.RateLimiters.CreateUserAllowed(utils.GetIP(r))
		if !allowed {
			http.Error(w, fmt.Sprintf("Too many requests wait %d seconds", remainingTime), http.StatusTooManyRequests)
			return
		}

	}
}

func createUser(controlDB *db.ControlDB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// Check to see if there is 10k users and return error if there is
		if controlDB.GetNUsersDBs() > 10000 {
			http.Error(w, "Full: Too many users", http.StatusTooManyRequests)
			return
		}

		userDB, err := controlDB.CreateUser()
		apiKey := "ahahaha"
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Printf("Number of Users %d, Number of Routines %d\n", controlDB.GetNUsersDBs(), runtime.NumGoroutine())

		acceptType := utils.GetAcceptType(r)

		switch acceptType {
		case utils.AcceptTypeJSON:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(fmt.Appendf(nil, "{\"userID\": \"%s\", \"apiKey\": \"%s\"}", userDB.Subdomain, apiKey))
			return
		case utils.AcceptTypeHTMX:
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusOK)
			err := templates.PublicCreateUser(w, "url", "token", "backup_key")
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		default:
			err = templates.PublicCreateUser(w, "url", "token", "backup_key")
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
	}
}
