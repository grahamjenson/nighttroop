package mtmiddleware

import (
	"context"
	"fmt"
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

			apiToken := r.URL.Query().Get("token")

			if subdomain == "control" {
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			// Search for tenants
			tenant, err := controlDB.GetTenant(subdomain, true)
			// We ignore errors and this is a best effort middleware
			// Each route needs to check the tenant if needed
			if err != nil || tenant == nil {
				// If the tenant is not found, we can return a 404
				fmt.Println("Tenant Middleware Error", err)
				http.NotFound(w, r)
				return
			}

			err = tenant.ValidateAPIToken(apiToken)
			if err != nil {
				fmt.Println("Tenant token error", err)
				// write an not auth http error
				http.Error(w, "Not Authorized", http.StatusUnauthorized)
				return
			}

			ctx = context.WithValue(ctx, TenantDBKey, tenant)
			next.ServeHTTP(w, r.WithContext(ctx))

		}

		return http.HandlerFunc(fn)
	}
}

// GetTenant retrieves the tenant ID from the context.
func GetTenant(ctx context.Context) *mtdb.TenantDB {
	if ctx == nil {
		return nil
	}
	if reqID, ok := ctx.Value(TenantDBKey).(*mtdb.TenantDB); ok {
		return reqID
	}
	return nil
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
