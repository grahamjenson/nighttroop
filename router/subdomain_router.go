package router

import (
	"context"
	"net/http"
	"strings"
)

const SubdomainKey ctxSubdomainKey = 1

// SubdomainRouter routes based on subdomain and path
type SubdomainRouter struct {
	controlRouter http.Handler
	tenantRouter  http.Handler
}

func (s *SubdomainRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	host := r.Host
	subdomain := strings.Split(host, ".")[0] // Get the first part of the host
	ctx = context.WithValue(ctx, SubdomainKey, subdomain)

	if subdomain == "control" {
		s.controlRouter.ServeHTTP(w, r.WithContext(ctx))
		return
	}
	s.tenantRouter.ServeHTTP(w, r.WithContext(ctx))
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
