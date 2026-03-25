// Contains all unquthenticated public routes available to anyone, e.g. create user and onboarding and documnentation
package server

import (
	"fmt"
	"net/http"
	"runtime"
	"sync"

	"github.com/go-chi/chi/v5"
	"golang.org/x/time/rate"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/server/assets/static"
	"nighttroop.t1d.nz/server/assets/templates"
	"nighttroop.t1d.nz/utils/config"
)

var userCreateLimiter = rate.NewLimiter(10, 100)
var userCreateMu sync.Mutex

func NewPublicRouter(controlDB *db.ControlDB, config config.ServerConfig) chi.Router {
	r := chi.NewRouter()

	static.AddSatticRoutes(r)
	// Create New User Route
	// Documentation Route
	// Public Facing SEO route
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		err := templates.PublicIndex(w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	r.Post("/user", createUser(controlDB))
	return r
}

func createUser(controlDB *db.ControlDB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		userCreateMu.Lock()
		allowed := userCreateLimiter.Allow()
		userCreateMu.Unlock()
		if !allowed {
			http.Error(w, "Too many requests", http.StatusTooManyRequests)
			return
		}

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

		acceptType := GetAcceptType(r)

		switch acceptType {
		case AcceptTypeJSON:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(fmt.Appendf(nil, "{\"userID\": \"%s\", \"apiKey\": \"%s\"}", userDB.Subdomain, apiKey))
			return
		case AcceptTypeHTMX:
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
