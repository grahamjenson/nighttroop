package user

import (
	"fmt"
	"net/http"

	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/utils"
	"nighttroop.t1d.nz/utils/httputils"
)

func GetUserMiddleware(controlDB *db.ControlDB) func(http.Handler) http.Handler {

	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			subdomain := httputils.GetSubdomainFromHTTPContext(ctx)

			apiToken := r.URL.Query().Get("token")

			if subdomain == "control" {
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			// Search for users
			userDB, userModel, err := controlDB.GetUser(subdomain, true)
			// We ignore errors and this is a best effort middleware
			// Each route needs to check the user if needed
			if err != nil || userDB == nil || userModel == nil {
				// If the user is not found, we can return a 404
				fmt.Println("User Middleware Error", err)
				http.NotFound(w, r)
				return
			}

			verified := utils.VerifyPassword("", apiToken)
			if !verified {
				fmt.Println("User token error", err)
				// write an not auth http error
				http.Error(w, "Not Authorized", http.StatusUnauthorized)
				return
			}

			ctx = httputils.SetUserDBInHTTPContext(ctx, userDB)
			next.ServeHTTP(w, r.WithContext(ctx))

		}

		return http.HandlerFunc(fn)
	}
}

func GETPingHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		userDB := httputils.GetUserFromHTTPContext(r.Context())

		if userDB == nil {
			fmt.Println("User NOT found")
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}

		accept := r.Header.Get("Accept")

		if accept == "application/json" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(fmt.Appendf(nil, "{\"ping\": \"pong\"}"))
			return
		}

		if accept == "text/htmx" {
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Pong"))
			return
		}

	}
}
