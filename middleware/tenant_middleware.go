package mtmiddleware

import (
	"context"
	"net/http"
	"strings"

	mtdb "sqlitemt.t1d.nz/db"
)

type ctxTenantDBKey int
type ctxSubdomainKey int

// RequestIDKey is the key that holds the unique request ID in a request context.
const TenantDBKey ctxTenantDBKey = 0
const SubdomainKey ctxSubdomainKey = 1

func GetUserDBMiddbleware(controlDB *mtdb.ControlDB) func(http.Handler) http.Handler {

	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			host := r.Host
			subdomain := strings.Split(host, ".")[0] // Get the first part of the host
			ctx = context.WithValue(ctx, SubdomainKey, subdomain)

			tenant, err := controlDB.GetTenant(subdomain)
			// We ignore errors and this is a best effort middleware
			// Each route needs to check the tenant if needed
			if err != nil && tenant != nil {
				ctx = context.WithValue(ctx, TenantDBKey, tenant)
			}

			// Validate the API ID and KEY in the tenant db

			next.ServeHTTP(w, r.WithContext(ctx))
		}
		return http.HandlerFunc(fn)
	}
}

// GetTenant retrieves the tenant ID from the context.
func GetTenant(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if reqID, ok := ctx.Value(TenantDBKey).(string); ok {
		return reqID
	}
	return ""
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
