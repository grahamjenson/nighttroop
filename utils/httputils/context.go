package httputils

import (
	"context"

	"nighttroop.t1d.nz/data/db"
)

type HTTPContextKeys int

const (
	UserDBKey HTTPContextKeys = iota
	SubdomainKey
)

func SetSubdomainInHTTPContext(ctx context.Context, subdomain string) context.Context {
	return context.WithValue(ctx, SubdomainKey, subdomain)
}

func SetUserDBInHTTPContext(ctx context.Context, userDB *db.UserDB) context.Context {
	return context.WithValue(ctx, UserDBKey, userDB)
}

func GetSubdomainFromHTTPContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if reqID, ok := ctx.Value(SubdomainKey).(string); ok {
		return reqID
	}
	return ""
}

// GetUser retrieves the user ID from the context.
func GetUserFromHTTPContext(ctx context.Context) *db.UserDB {
	if ctx == nil {
		return nil
	}
	if reqID, ok := ctx.Value(UserDBKey).(*db.UserDB); ok {
		return reqID
	}
	return nil
}
