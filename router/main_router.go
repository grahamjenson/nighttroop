// This is the leading middleware deciding what lower router to use based on subdomain
package router

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"nighttroop.t1d.nz/config"
	"nighttroop.t1d.nz/db"
	"nighttroop.t1d.nz/utils"
)

const SubdomainKey ctxSubdomainKey = 1

func GetMainHandler(cdb *db.ControlDB, config config.ServerConfig) http.Handler {

	controlRouter := NewControlRouter(cdb)
	userRouter := NewUserRouter(cdb)
	public := NewPublicRouter(cdb)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		subdomain, err := GetSubdomain(r.Host, config.HostName)
		if err != nil {
			fmt.Printf("Main router 404 for host header: %s, error: %v\n", r.Host, err)
			http.NotFound(w, r)
			return
		}

		ctx = context.WithValue(ctx, SubdomainKey, subdomain)

		switch {
		case subdomain == "control" || subdomain == "admin":
			fmt.Println("Routing to control router")
			controlRouter.ServeHTTP(w, r.WithContext(ctx))
		case subdomain == "www" || subdomain == "public" || subdomain == "":
			fmt.Println("Routing to public router")
			public.ServeHTTP(w, r.WithContext(ctx))
		case utils.ValidateUserID(subdomain):
			fmt.Println("Routing to user router for user:", subdomain)
			userRouter.ServeHTTP(w, r.WithContext(ctx))
		default:
			fmt.Printf("Main router 404 for subdomain: %s\n", subdomain)
			http.NotFound(w, r)
		}

	})

}

// This is a bit annoying but for subdomain routing we need to get the subdomain from the host header
// For this we need to know the expected host like nt.t1d.nz and then summize the subdomain from there
// What we do is we get the header, remove the port, split it by '.' and then get the part immediatly
// If the host is wrong, or there are too many subdomains, we return null, to be handled with a 404
func GetSubdomain(host_header, hostname string) (string, error) {
	// remove port if present
	host := host_header
	// Remove the port if it is there
	if i := strings.Index(host_header, ":"); i != -1 {
		host = host_header[:i]
	}

	endsWithHostName := strings.HasSuffix(host, hostname)
	endsWithLocalhost := strings.HasSuffix(host, "localhost")
	localhost_len := 9

	// If the host doesnt end with the host name OR localhost (for dev) then something is wrong, return null to trigger a 404;
	if !endsWithHostName && !endsWithLocalhost {
		return "", fmt.Errorf("invalid host: %s", host)
	}

	if len(host) == len(hostname) || len(host) == len("localhost") {
		return "", nil // This means there is no subdomain, just the main domain, so we return empty string
	}

	if endsWithHostName {
		// If there is no '.' seperating the host from subdomain
		if host[len(host)-len(hostname)-1] != '.' {
			return "", fmt.Errorf("invalid host format: %s", host) // This should always be he case, and incase it isnt means something bad
		}
		return host[0 : len(host)-len(hostname)-1], nil
	}

	if endsWithLocalhost {
		// If there is no '.' seperating the host from subdomain
		if host[len(host)-localhost_len-1] != '.' {
			return "", fmt.Errorf("invalid host format: %s", host) // This should always be he case, and incase it isnt means something bad
		}
		return host[0 : len(host)-localhost_len-1], nil
	}

	return "", fmt.Errorf("unexpected error parsing host: %s", host) // This should never happen but just in case
}

func GetSubdomaiFromHTTPContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if reqID, ok := ctx.Value(SubdomainKey).(string); ok {
		return reqID
	}
	return ""
}
