// This is the leading middleware deciding what lower router to use based on subdomain
package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"runtime"
	"strings"
	"time"

	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/server/controlapi"
	"nighttroop.t1d.nz/server/nightscoutapi"
	"nighttroop.t1d.nz/server/siteapi"
	"nighttroop.t1d.nz/utils"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/httputils"
)

func GetMainHandler(controlDB *db.ControlDB, config *config.ServerConfig) http.Handler {

	controlRouter := NewControlRouter(controlDB, config)
	siteRouter := siteapi.NewSiteRouter(controlDB, config)
	public := NewPublicRouter(controlDB, config)
	nightscoutRouter := nightscoutapi.NewNightscoutRouter(controlDB, config)
	setupRouter := NewSetupRouter(controlDB, config)

	// Only the configured host is served; loopback names are for local runs only
	validHostnames := []string{config.HostName}
	if config.IsDevOrTest() {
		validHostnames = append(validHostnames, "localhost", "127.0.0.1")
	}

	// Past what the server can work on at once, requests are refused rather than all queueing until they time out. Uploads
	// to the Nightscout API keep a share to themselves, so glucose keeps arriving while page loads fill the rest
	cpus := runtime.GOMAXPROCS(0)
	upload := func(r *http.Request) bool { return nightscoutWrite(r, validHostnames) }
	shed := httputils.ShedLoadReserving(config.RequestsInFlightPerCPU*cpus, config.RequestsBacklogPerCPU*cpus, 5*time.Second, httputils.LongLived, upload)

	return shed(httputils.ClientIP(config.TrustedProxies)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Global IP Address Rate Limit, which every request has, the control site's too
		ipAddress := httputils.GetIPFromHTTPContext(ctx)
		allowed := httputils.GlobalIPWithinRateLimit(config, ipAddress)
		if !allowed {
			w.WriteHeader(http.StatusTooManyRequests)
			slog.InfoContext(r.Context(), "Global IP rate limit exceeded", "ip", ipAddress)
			return
		}
		if !httputils.LimitRequestBody(w, r, config.MaxRequestBodyBytes) {
			slog.InfoContext(r.Context(), "Request body too large", "ip", ipAddress, "bytes", r.ContentLength)
			return
		}
		host, _, err := net.SplitHostPort(r.Host)
		if err != nil {
			host = r.Host
		}
		routeTarget, siteSubdomain, err := GetRouteTargetFromSubdomain(host, validHostnames)
		if err != nil {
			slog.InfoContext(r.Context(), "Main router 404 for host header", "host", host, "error", err)
			http.NotFound(w, r)
			return
		}

		// Global Rate Limit
		if !globalWithinRateLimit(config, routeTarget, r) {
			w.WriteHeader(http.StatusTooManyRequests)
			slog.InfoContext(r.Context(), "Global rate limit exceeded", "route_target", routeTarget)
			return
		}

		// Until its control admin is made, a new server is only its setup page
		if !controlDB.SetupDone() {
			serveSetup(w, r.WithContext(ctx), config, routeTarget, setupRouter, public)
			return
		}

		switch routeTarget {
		case RouteTargetControl:
			controlRouter.ServeHTTP(w, r.WithContext(httputils.CreateControlContext(ctx, controlDB, ipAddress, httputils.GetHTTPAction(r))))
		case RouteTargetPublic:
			public.ServeHTTP(w, r.WithContext(ctx))
		case RouteTargetNightscout:
			// so <siteid>.ns.domain routes to the nightscout compatible API
			serveSite(ctx, w, r, nightscoutRouter, config, controlDB, siteSubdomain, ipAddress)
		case RouteTargetSite:
			serveSite(ctx, w, r, siteRouter, config, controlDB, siteSubdomain, ipAddress)
		default:
			slog.InfoContext(r.Context(), "Main router 404 for subdomain", "host", r.Host, "siteSubdomain", siteSubdomain, "routeTarget", routeTarget)
			http.NotFound(w, r)
		}
	})))

}

// nightscoutWrite is whether a request writes through a site's Nightscout API, as a device uploading its readings does
func nightscoutWrite(r *http.Request, validHostnames []string) bool {
	if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
		return false
	}
	host, _, err := net.SplitHostPort(r.Host)
	if err != nil {
		host = r.Host
	}
	target, _, err := GetRouteTargetFromSubdomain(host, validHostnames)
	return err == nil && target == RouteTargetNightscout
}

// globalWithinRateLimit limits every request to the server together, but the control site's: a control admin signed in
// is never held up by it, so they can fix whatever is using it up, and the rest of the control site, like its login, has
// a limit of its own
func globalWithinRateLimit(config *config.ServerConfig, routeTarget RouteTarget, r *http.Request) bool {
	if routeTarget != RouteTargetControl {
		return config.RateLimiters.GlobalRateLimiter.Allow()
	}
	return controlapi.SignedIn(config, r) || config.RateLimiters.ControlRateLimiter.Allow()
}

// serveSite serves a request for a site with its context: 404 for a site there is not, but 503 when it cannot be
// looked up, as when the server is too busy to answer in time, so an uploader tries again rather than take its URL as
// wrong
func serveSite(ctx context.Context, w http.ResponseWriter, r *http.Request, router http.Handler, config *config.ServerConfig, controlDB *db.ControlDB, subdomain string, ipAddress string) {
	siteDB, siteModel, err := controlDB.GetSite(ctx, subdomain)
	if errors.Is(err, db.ErrSiteNotFound) {
		slog.InfoContext(ctx, "Main router 404 for subdomain", "subdomain", subdomain, "error", err)
		http.NotFound(w, r)
		return
	}
	if err != nil {
		httputils.RefuseBusy(w, r.WithContext(ctx), "cannot look up the site: "+err.Error())
		return
	}
	ctx = httputils.CreateSiteContext(ctx, config, controlDB, siteDB, siteModel, subdomain, ipAddress, httputils.GetHTTPAction(r), httputils.GetAcceptType(r), httputils.GetContentType(r))
	router.ServeHTTP(w, r.WithContext(ctx))
}

type RouteTarget int

const (
	RouteTargetPublic RouteTarget = iota
	RouteTargetControl
	RouteTargetSite
	RouteTargetNightscout
)

func GetRouteTargetFromSubdomain(host_header string, validHostnames []string) (RouteTarget, string, error) {

	// A fully qualified host ends with a dot; match only on whole labels
	host_header = strings.TrimSuffix(host_header, ".")
	allowedHost := false
	for _, hostname := range validHostnames {
		if host_header == hostname || strings.HasSuffix(host_header, "."+hostname) {
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
