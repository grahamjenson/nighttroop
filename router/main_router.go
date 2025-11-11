// This is the leading middleware deciding what lower router to use based on subdomain
package router

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"nighttroop.t1d.nz/db"
)

const SubdomainKey ctxSubdomainKey = 1

func GetMainHandler(cdb *db.ControlDB) http.Handler {

	controlRouter := NewControlRouter(cdb)
	userRouter := NewUserRouter(cdb)
	public := NewPublicRouter(cdb)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		host := r.Host
		subdomain := strings.Split(host, ".")[0] // Get the first part of the host
		ctx = context.WithValue(ctx, SubdomainKey, subdomain)

		switch {
		case subdomain == "control" || subdomain == "admin":
			fmt.Println("Routing to control router")
			controlRouter.ServeHTTP(w, r.WithContext(ctx))
		case subdomain == "www" || subdomain == "public" || subdomain == "":
			fmt.Println("Routing to public router")
			public.ServeHTTP(w, r.WithContext(ctx))
		case db.ValidateUserID(subdomain):
			fmt.Println("Routing to user router for user:", subdomain)
			userRouter.ServeHTTP(w, r.WithContext(ctx))
		default:
			fmt.Printf("Main router 404 for subdomain: %s\n", subdomain)
			http.NotFound(w, r)
		}

	})

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
