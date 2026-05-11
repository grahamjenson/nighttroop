package httputils

import (
	"context"

	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/sql/control/user"
)

type HTTPContextKeys int

const (
	UserDBKey HTTPContextKeys = iota
	UserGetModelKey
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

func SetUserDBAndModelInHTTPContext(ctx context.Context, userDB *db.UserDB, userGetModel user.IUser) context.Context {
	ctx = context.WithValue(ctx, UserDBKey, userDB)
	ctx = context.WithValue(ctx, UserGetModelKey, userGetModel)
	return ctx
}

func GetUserDBAndModelFromHTTPContext(ctx context.Context) (*db.UserDB, user.IUser) {
	if ctx == nil {
		return nil, nil
	}
	var userDB *db.UserDB
	var userGetModel user.IUser
	if reqID, ok := ctx.Value(UserDBKey).(*db.UserDB); ok {
		userDB = reqID
	}
	if reqID, ok := ctx.Value(UserGetModelKey).(user.IUser); ok {
		userGetModel = reqID
	}
	return userDB, userGetModel
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
