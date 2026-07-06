// This is the leading middleware deciding what lower router to use based on subdomain
package server

import (
	"context"
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
			slog.Info("Global rate limit exceeded")
			return
		}

		// Global IP Address Rate Limit
		ipAddress := httputils.GetIP(r)
		ctx = httputils.SetIPInHTTPContext(ctx, ipAddress)
		allowed = httputils.GlobalIPRateLimitAllowed(config, ipAddress)
		if !allowed {
			w.WriteHeader(http.StatusTooManyRequests)
			slog.Info("Global IP rate limit exceeded", "ip", ipAddress)
			return
		}

		subdomain, err := GetSubdomainFromHeader(r.Host, config.HostName)
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
		case strings.HasSuffix(subdomain, ".ns") || strings.HasPrefix(subdomain, "ns."):
			userSubdomain := strings.TrimSuffix(subdomain, ".ns")
			userSubdomain = strings.TrimPrefix(userSubdomain, "ns.")

			if !utils.SyntacticallyValidateUserID(userSubdomain) {
				slog.Error("Main router 404 for nightscout subdomain with invalid user id format", "subdomain", subdomain)
				http.NotFound(w, r)
				return
			}
			// so <userid>.ns.domain routes to the nightscout compatible API
			if exists, ctx := UserExistsAndWriteToContext(controlDB, userSubdomain, r); exists {
				nightscoutRouter.ServeHTTP(w, r.WithContext(ctx))
			} else {
				slog.Error("Main router 404 for nightscout subdomain", "subdomain", subdomain)
				http.NotFound(w, r)
			}
		case utils.SyntacticallyValidateUserID(subdomain):
			// Check to see if in memory map
			// SHOULD ATTACH USER DB AND MODEL TO THE CONTEXT AT THIS POINT

			if exists, ctx := UserExistsAndWriteToContext(controlDB, subdomain, r); exists {
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

func UserExistsAndWriteToContext(controlDB *db.ControlDB, subdomain string, r *http.Request) (bool, context.Context) {
	userDB, userGetModel, err := controlDB.GetUser(subdomain)
	if err != nil {
		slog.Error("Error getting user for subdomain", "subdomain", subdomain, "error", err)
		return false, nil
	}

	ctx := httputils.SetSiteDBAndModelAndCreateUserContext(r.Context(), controlDB, userDB, userGetModel, subdomain)

	return true, ctx
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
