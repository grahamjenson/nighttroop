package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"nighttroop.t1d.nz/db"
)

func NewControlRouter(controlDB *db.ControlDB) chi.Router {
	r := chi.NewRouter()
	// TODO: admin auth middleware
	r.Use(BasicAuthMiddleware("admin", "control")) // TODO: change to env vars

	return r
}

func BasicAuthMiddleware(username, password string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, pass, ok := r.BasicAuth()

			// Check if basic auth credentials are provided and valid
			if !ok || user != username || pass != password {
				w.Header().Set("WWW-Authenticate", `Basic realm="control"`)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			// If credentials are valid, proceed to the next handler
			next.ServeHTTP(w, r)
		})
	}
}
