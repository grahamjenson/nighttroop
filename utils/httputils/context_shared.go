package httputils

import (
	"context"
	"net/http"
	"net/netip"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// ClientIP sets each request's client IP, which GetIPFromHTTPContext reads, with chi's middlewares: the TCP peer;
// from a trusted proxy, the X-Forwarded-For hop before the trusted ones; and on Fly.io its Fly-Client-IP header
func ClientIP(trustedProxies []netip.Prefix) func(http.Handler) http.Handler {
	chain := chi.Chain(middleware.ClientIPFromRemoteAddr)
	if len(trustedProxies) > 0 {
		prefixes := make([]string, len(trustedProxies))
		for i, prefix := range trustedProxies {
			prefixes[i] = prefix.String()
		}
		chain = append(chain, middleware.ClientIPFromXFF(prefixes...))
	}
	if os.Getenv("FLY_APP_NAME") != "" {
		chain = append(chain, middleware.ClientIPFromHeader("Fly-Client-IP"))
	}
	return chain.Handler
}

// GetIPFromHTTPContext is the client IP ClientIP set, or empty without one
func GetIPFromHTTPContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	return middleware.GetClientIP(ctx)
}

// WithClientIP is the request's context with its TCP peer as its client IP, as ClientIP sets it with no proxies, for a
// request a handler is given without the main router, as in a test
func WithClientIP(r *http.Request) context.Context {
	ctx := r.Context()
	ClientIP(nil)(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) { ctx = r.Context() })).ServeHTTP(nil, r)
	return ctx
}
