package httputils

import (
	"context"

	"nighttroop.t1d.nz/data/db"
)

type HTTPContextKeys int

const (
	UserDBKey HTTPContextKeys = iota
	SubdomainKey
	DeviceNameKey
	IPAddressKey
	AdminNameKey
	RoleKey
)

// //
// Subdomain
// //
func SetSubdomainInHTTPContext(ctx context.Context, subdomain string) context.Context {
	return context.WithValue(ctx, SubdomainKey, subdomain)
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

// //
// DeviceName
// //
func SetDeviceNameInHTTPContext(ctx context.Context, deviceID string) context.Context {
	return context.WithValue(ctx, DeviceNameKey, deviceID)
}

func GetDeviceNameFromHTTPContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if reqID, ok := ctx.Value(DeviceNameKey).(string); ok {
		return reqID
	}
	return ""
}

// //
// IPAddress
// //
func SetIPInHTTPContext(ctx context.Context, ip string) context.Context {
	return context.WithValue(ctx, IPAddressKey, ip)
}

func GetIPFromHTTPContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if reqID, ok := ctx.Value(IPAddressKey).(string); ok {
		return reqID
	}
	return ""
}

////
// UserDB
////

func SetUserDBInHTTPContext(ctx context.Context, userDB *db.UserDB) context.Context {
	return context.WithValue(ctx, UserDBKey, userDB)
}

func GetUserFromHTTPContext(ctx context.Context) *db.UserDB {
	if ctx == nil {
		return nil
	}
	if reqID, ok := ctx.Value(UserDBKey).(*db.UserDB); ok {
		return reqID
	}
	return nil
}

// /
// AdminName
func SetAdminNameInHTTPContext(ctx context.Context, adminName string) context.Context {
	return context.WithValue(ctx, AdminNameKey, adminName)
}

func GetAdminNameFromHTTPContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if reqID, ok := ctx.Value(AdminNameKey).(string); ok {
		return reqID
	}
	return ""
}

// //
// Role
// //
func SetRoleInHTTPContext(ctx context.Context, role string) context.Context {
	return context.WithValue(ctx, RoleKey, role)
}

func GetRoleFromHTTPContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if reqID, ok := ctx.Value(RoleKey).(string); ok {
		return reqID
	}
	return ""
}
