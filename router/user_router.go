package router

// User router enables user specific router, and can route to NS router if endpoint hits that layer.

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	mtdb "nighttroop.t1d.nz/db"

	"github.com/go-chi/chi/v5"
	"nighttroop.t1d.nz/db"
)

type ctxUserDBKey int
type ctxSubdomainKey int

// RequestIDKey is the key that holds the unique request ID in a request context.
const UserDBKey ctxUserDBKey = 0

func NewUserRouter(controlDB *db.ControlDB) chi.Router {
	r := chi.NewRouter()
	r.Use(GetUserMiddleware(controlDB))

	addNSRoutes(r)
	addSocketIORoute(r)

	r.Post("/ping", func(w http.ResponseWriter, r *http.Request) {
		userDB := GetUser(r.Context())

		if userDB == nil {
			fmt.Println("User NOT found")
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}

		accept := r.Header.Get("Accept")

		if accept == "application/json" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(fmt.Sprintf("{\"ping\": \"pong\"}")))
			return
		}

		if accept == "text/htmx" {
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Pong"))
			return
		}

	})
	return r
}

func addNSRoutes(r chi.Router) {
	// Add NS routes here

}

func addSocketIORoute(r chi.Router) {
	// Add SocketIO route here
}

func GetUserMiddleware(controlDB *mtdb.ControlDB) func(http.Handler) http.Handler {

	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			host := r.Host
			subdomain := strings.Split(host, ".")[0] // Get the first part of the host
			ctx = context.WithValue(ctx, SubdomainKey, subdomain)

			apiToken := r.URL.Query().Get("token")

			if subdomain == "control" {
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			// Search for users
			user, err := controlDB.GetUser(subdomain, true)
			// We ignore errors and this is a best effort middleware
			// Each route needs to check the user if needed
			if err != nil || user == nil {
				// If the user is not found, we can return a 404
				fmt.Println("User Middleware Error", err)
				http.NotFound(w, r)
				return
			}

			err = user.ValidateAPIToken(apiToken)
			if err != nil {
				fmt.Println("User token error", err)
				// write an not auth http error
				http.Error(w, "Not Authorized", http.StatusUnauthorized)
				return
			}

			ctx = context.WithValue(ctx, UserDBKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))

		}

		return http.HandlerFunc(fn)
	}
}

// GetUser retrieves the user ID from the context.
func GetUser(ctx context.Context) *mtdb.UserDB {
	if ctx == nil {
		return nil
	}
	if reqID, ok := ctx.Value(UserDBKey).(*mtdb.UserDB); ok {
		return reqID
	}
	return nil
}
