// This is the leading middleware deciding what lower router to use based on subdomain
package router

import (
	"context"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"nighttroop.t1d.nz/db"
)

const SubdomainKey ctxSubdomainKey = 1

func GetMainRouter(db *db.ControlDB) chi.Router {
	r := chi.NewRouter()

	controlRouter := NewControlRouter(db)
	userRouter := NewUserRouter(db)
	public := NewPublicRouter(db)

	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			host := r.Host
			subdomain := strings.Split(host, ".")[0] // Get the first part of the host
			ctx = context.WithValue(ctx, SubdomainKey, subdomain)

			// If subdomain is control plane, route to control router
			if subdomain == "control" || subdomain == "admin" {
				controlRouter.ServeHTTP(w, r.WithContext(ctx))
				return
			}
			// If subdomain is public or www, route to public router
			if subdomain == "www" || subdomain == "public" || subdomain == "" {
				public.ServeHTTP(w, r.WithContext(ctx))
				return
			}
			// Otherwise route to user router
			// TODO: route to user, if subdomain matches a users potential id (length, chars etc)
			userRouter.ServeHTTP(w, r.WithContext(ctx))
		})
	})

	return r
}

func GetSubdomain(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if reqID, ok := ctx.Value(SubdomainKey).(string); ok {
		return reqID
	}
	return ""
}
