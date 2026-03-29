package server

import (
	"fmt"
	"net/http"
	"runtime"

	"github.com/go-chi/chi/v5"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/server/assets/static"
	publicTemplates "nighttroop.t1d.nz/server/assets/templates/public"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/httputils"
)

func NewPublicRouter(controlDB *db.ControlDB, config config.ServerConfig) chi.Router {
	r := chi.NewRouter()

	static.AddSatticRoutes(r)

	// TODO Documentation route, public facing SEO route

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		err := publicTemplates.CreateUser(w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	r.Post("/user", POSTCreateUserHandler(controlDB, config))

	// TODO backupKeyhashFcuntions
	// 1. reset admin passwrod
	// 2. reset admin totp
	// 3. download sqlite backup
	return r
}

func POSTCreateUserHandler(controlDB *db.ControlDB, config config.ServerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		allowed, remainingTime := config.RateLimiters.CreateUserAllowed(httputils.GetIP(r))
		if !allowed {
			w.WriteHeader(http.StatusTooManyRequests)
			publicTemplates.CreateUserRateLimited(w, fmt.Sprintf("%d", remainingTime))
			return
		}

		// Check to see if there is 10k users and return error if there is
		if controlDB.GetNUsersDBs() > config.MaxUsers {
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

		acceptType := httputils.GetAcceptType(r)

		switch acceptType {
		case httputils.AcceptTypeJSON:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(fmt.Appendf(nil, "{\"userID\": \"%s\", \"apiKey\": \"%s\"}", userDB.Subdomain, apiKey))
			return
		case httputils.AcceptTypeHTMX:
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusOK)
			err := publicTemplates.CreatedUserInfo(w, "url", "token", "backup_key")
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		default:
			err = publicTemplates.CreatedUserInfo(w, "url", "token", "backup_key")
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}

	}
}
