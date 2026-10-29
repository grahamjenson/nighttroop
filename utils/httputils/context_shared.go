package httputils

import (
	"context"
	"net"
	"net/http"
	"net/netip"
	"os"
	"slices"
	"strings"
)

const ipAddressKey = "ip_address"

// SetIPInHTTPContext is the client's IP, from X-Forwarded-For only when the request came through a trusted proxy
func SetIPInHTTPContext(ctx context.Context, r *http.Request, trustedProxies []netip.Prefix) (context.Context, string) {
	ip := getIPAddressFromRequest(r, trustedProxies)
	return context.WithValue(ctx, ipAddressKey, ip), ip
}

func GetIPFromHTTPContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if ip, ok := ctx.Value(ipAddressKey).(string); ok {
		return ip
	}
	return ""
}

func getIPAddressFromRequest(r *http.Request, trustedProxies []netip.Prefix) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}

	// If we are deployed on Fly.io then use the Fly-Client-IP header
	if os.Getenv("FLY_APP_NAME") != "" {
		if ip := strings.TrimSpace(r.Header.Get("Fly-Client-IP")); ip != "" {
			return ip
		}
	}

	// If it is a trusted proxy then we can accept the X-Forwarded For
	if isTrustedProxy(host, trustedProxies) {
		// Check X-Forwarded-For header first (in case of reverse proxy)
		xForwardedFor := r.Header.Get("X-Forwarded-For")
		if xForwardedFor != "" {
			// Use the last IP, the one our proxy appended; earlier ones can be set by the client
			ips := strings.Split(xForwardedFor, ",")
			return strings.TrimSpace(ips[len(ips)-1])
		}
	}

	return host
}

func isTrustedProxy(host string, trustedProxies []netip.Prefix) bool {
	addr, err := netip.ParseAddr(host)
	if err != nil {
		return false
	}
	addr = addr.Unmap()
	return slices.ContainsFunc(trustedProxies, func(p netip.Prefix) bool { return p.Contains(addr) })
}
