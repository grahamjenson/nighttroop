// This is the leading middleware deciding what lower router to use based on subdomain
package server

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/server/nightscout_api"
	"nighttroop.t1d.nz/utils"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/httputils"
)

func GetMainHandler(controlDB *db.ControlDB, config config.ServerConfig) http.Handler {

	controlRouter := NewControlRouter(controlDB, config)
	userRouter := NewUserRouter(controlDB, config)
	public := NewPublicRouter(controlDB, config)
	nightscoutRouter := nightscout_api.NewNightscoutRouter(controlDB, config)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Global Rate Limit
		allowed := config.RateLimiters.GlobalRateLimiter.Allow()
		if !allowed {
			w.WriteHeader(http.StatusTooManyRequests)
			fmt.Fprint(w, "Too many requests, global rate limit exceeded")
			return
		}

		// Global IP Address Rate Limit
		ipAddress := httputils.GetIP(r)
		ctx = httputils.SetIPInHTTPContext(ctx, ipAddress)
		allowed = config.RateLimiters.GetGlobalIPRateLimiter(ipAddress).Allow()
		if !allowed {
			w.WriteHeader(http.StatusTooManyRequests)
			fmt.Fprint(w, "Too many requests, IP rate limit exceeded")
			return
		}

		subdomain, err := GetSubdomainFromHeader(r.Host, config.HostName)
		ctx = httputils.SetSubdomainInHTTPContext(ctx, subdomain)
		if err != nil {
			slog.Error("Main router 404 for host header", "host", r.Host, "error", err)
			http.NotFound(w, r)
			return
		}

		switch {
		case subdomain == "control" || subdomain == "admin":
			controlRouter.ServeHTTP(w, r.WithContext(ctx))
		case subdomain == "www" || subdomain == "public" || subdomain == "":
			public.ServeHTTP(w, r.WithContext(ctx))
		case strings.HasSuffix(subdomain, ".ns"):
			userSubdomain := strings.TrimSuffix(subdomain, ".ns")
			// so <userid>.ns.domain routes to the nightscout compatible API
			if utils.SyntacticallyValidateUserID(userSubdomain) && controlDB.UserExists(userSubdomain) {
				// Reset subdomain without the .ns suffix
				ctx = httputils.SetSubdomainInHTTPContext(ctx, userSubdomain)
				nightscoutRouter.ServeHTTP(w, r.WithContext(ctx))
			} else {
				slog.Error("Main router 404 for nightscout subdomain", "subdomain", subdomain)
				http.NotFound(w, r)
			}
		case utils.SyntacticallyValidateUserID(subdomain):
			// Check to see if in memory map
			if controlDB.UserExists(subdomain) {
				userRouter.ServeHTTP(w, r.WithContext(ctx))
			} else {
				slog.Error("Main router 404 for subdomain", "subdomain", subdomain)
				http.NotFound(w, r)
			}
		default:
			slog.Error("Main router 404 for subdomain", "subdomain", subdomain)
			http.NotFound(w, r)
		}
	})

}

// This is a bit annoying but for subdomain routing we need to get the subdomain from the host header
// For this we need to know the expected host like nt.t1d.nz and then summize the subdomain from there
// What we do is we get the header, remove the port, remove the expected host and leave the subdomain
// We also have to deal with localhost
func GetSubdomainFromHeader(host_header, hostname string) (string, error) {
	// remove port if present
	host := host_header
	// Remove the port if it is there
	if before, _, ok := strings.Cut(host_header, ":"); ok {
		host = before
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
