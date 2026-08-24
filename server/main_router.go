// This is the leading middleware deciding what lower router to use based on subdomain
package server

import (
	"context"
	"fmt"
	"log/slog"
	"net"
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
	siteRouter := NewSiteRouter(controlDB, config)
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
		host, _, err := net.SplitHostPort(r.Host)
		if err != nil {
			host = r.Host
		}
		routeTarget, siteSubdomain, err := GetRouteTargetFromSubdomain(host, []string{config.HostName, "localhost", "t1d.nz"})
		if err != nil {
			slog.Error("Main router 404 for host header", "host", host, "error", err)
			http.NotFound(w, r)
			return
		}

		switch routeTarget {
		case RouteTargetControl:
			controlRouter.ServeHTTP(w, r.WithContext(ctx))
		case RouteTargetPublic:
			public.ServeHTTP(w, r.WithContext(ctx))
		case RouteTargetNightscout:
			// so <siteid>.ns.domain routes to the nightscout compatible API
			if exists, ctx := SiteExistsAndWriteToContext(controlDB, siteSubdomain, r); exists {
				nightscoutRouter.ServeHTTP(w, r.WithContext(ctx))
			} else {
				slog.Error("Main router 404 for nightscout subdomain", "subdomain", siteSubdomain)
				http.NotFound(w, r)
			}
		case RouteTargetSite:
			// Attach userdb and model to the context
			if exists, ctx := SiteExistsAndWriteToContext(controlDB, siteSubdomain, r); exists {
				siteRouter.ServeHTTP(w, r.WithContext(ctx))
			} else {
				slog.Error("Main router 404 for subdomain", "subdomain", siteSubdomain)
				http.NotFound(w, r)
			}
		default:
			slog.Error("Main router 404 for subdomain", "host", r.Host, "siteSubdomain", siteSubdomain, "routeTarget", routeTarget)
			http.NotFound(w, r)
		}
	})

}

func SiteExistsAndWriteToContext(controlDB *db.ControlDB, subdomain string, r *http.Request) (bool, context.Context) {
	siteDB, siteModel, err := controlDB.GetSite(subdomain)
	if err != nil {
		slog.Error("Error getting site for subdomain", "subdomain", subdomain, "error", err)
		return false, nil
	}

	ctx := httputils.SetSiteDBAndModelAndCreateSiteContext(r.Context(), controlDB, siteDB, siteModel, subdomain)

	return true, ctx
}

type RouteTarget int

const (
	RouteTargetPublic RouteTarget = iota
	RouteTargetControl
	RouteTargetSite
	RouteTargetNightscout
)

func GetRouteTargetFromSubdomain(host_header string, validHostnames []string) (RouteTarget, string, error) {

	allowedHost := false
	for _, hostname := range validHostnames {
		if strings.HasSuffix(host_header, hostname) {
			allowedHost = true
			break
		}
	}
	if !allowedHost {
		return RouteTargetPublic, "", fmt.Errorf("host header %s does not match any valid hostnames", host_header)
	}

	if strings.HasPrefix(host_header, "control.") || strings.HasPrefix(host_header, "admin.") {
		return RouteTargetControl, "", nil
	}

	if strings.HasPrefix(host_header, "public.") || strings.HasPrefix(host_header, "www.") {
		return RouteTargetPublic, "", nil
	}

	subdomains := strings.Split(host_header, ".")
	// 0 is undefined
	// e.g. localhost = 1
	// e.g alicezzz.localhost = 2
	// e.g alicezzz.ns.localhost = 3
	subdomainLength := len(subdomains)

	if subdomainLength <= 1 {
		return RouteTargetPublic, "", nil
	}

	// First subdomain should always be site
	siteSubdomain := subdomains[0]
	if utils.SyntacticallyValidateSiteID(siteSubdomain) {
		if subdomains[1] == "ns" {
			return RouteTargetNightscout, siteSubdomain, nil
		}
		return RouteTargetSite, siteSubdomain, nil
	}

	return RouteTargetPublic, siteSubdomain, nil
}
