package httputils

import (
	"context"
	"net"
	"net/http"
	"os"
	"strings"
)

const ipAddressKey = "ip_address"

func SetIPInHTTPContext(ctx context.Context, r *http.Request) (context.Context, string) {
	ip := getIPAddressFromRequest(r)
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

func getIPAddressFromRequest(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}

	// If we are deployed on Fly.io then use the Fly-Client-IP header
	if os.Getenv("FLY_APP_NAME") != "" {
		if ip := strings.TrimSpace(r.Header.Get("Fly-Client-IP")); ip != "" {
			return ip
		}
	}

	// If it is a trusted proxy then we can accept the X-Forwarded For
	if isTrustedProxy(host) {
		// Check X-Forwarded-For header first (in case of reverse proxy)
		xForwardedFor := r.Header.Get("X-Forwarded-For")
		if xForwardedFor != "" {
			// The X-Forwarded-For header can contain multiple IPs, the first one is the client's IP
			ips := strings.Split(xForwardedFor, ",")
			return strings.TrimSpace(ips[0])
		}
	}

	return host
}

func isTrustedProxy(ip string) bool {
	// Define a list of trusted proxy IPs or CIDR ranges
	trustedProxies := []string{
		"127.0.0.1",
		"::1",
	}

	for _, trustedProxy := range trustedProxies {
		if ip == trustedProxy {
			return true
		}
	}

	return false
}
