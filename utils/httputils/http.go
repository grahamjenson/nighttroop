package httputils

import (
	"net"
	"net/http"
	"os"
	"strings"
)

type AcceptType int

const (
	AcceptTypeHTML AcceptType = iota
	AcceptTypeJSON
	AcceptTypeHTMX
)

func GetAcceptType(r *http.Request) AcceptType {
	accept := r.Header.Get("Accept")
	hxRequest := r.Header.Get("HX-Request")

	switch {
	case strings.Contains(accept, "application/json"):
		return AcceptTypeJSON
	case strings.Contains(accept, "text/htmx") || hxRequest == "true":
		return AcceptTypeHTMX
	case strings.Contains(accept, "text/html"):
		return AcceptTypeHTML
	default:
		// If no Accept header use ContentType
		contentType := GetContentType(r)
		switch contentType {
		case ContentTypeForm:
			return AcceptTypeHTML
		case ContentTypeJSON:
			return AcceptTypeJSON

		default:
			return AcceptTypeJSON
		}
	}
}

func (at AcceptType) API() bool {
	return at == AcceptTypeJSON
}

type ContentType string

const (
	ContentTypeForm ContentType = "application/x-www-form-urlencoded"
	ContentTypeJSON ContentType = "application/json"
)

func GetContentType(r *http.Request) ContentType {
	contentType := r.Header.Get("Content-Type")
	// Defaults to JSON
	switch {
	case strings.Contains(contentType, "application/x-www-form-urlencoded"):
		return ContentTypeForm
	case strings.Contains(contentType, "application/json"):
		return ContentTypeJSON
	default:
		return ContentTypeJSON
	}
}

func GetIP(r *http.Request) string {
	// Check if it has been cached in the context to not repeat this work
	ipAddress := GetIPFromHTTPContext(r.Context())
	if ipAddress != "" {
		return ipAddress
	}

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
