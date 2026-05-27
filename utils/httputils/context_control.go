package httputils

import (
	"context"

	"nighttroop.t1d.nz/utils/consts"
)

const ControlContextKey = "control_context"

type ControlContext struct {
	AdminName string
	AdminRole consts.AdminRole
}

func GetControlContext(ctx context.Context) ControlContext {
	if ctx == nil {
		return ControlContext{}
	}
	if controlCtx, ok := ctx.Value(ControlContextKey).(*ControlContext); ok {
		return *controlCtx
	}
	return ControlContext{}
}

// //
// IdentityName
// //
func SetAdminNameInControlContext(ctx context.Context, adminName string) context.Context {
	controlContext := GetControlContext(ctx)
	controlContext.AdminName = adminName
	return context.WithValue(ctx, ControlContextKey, &controlContext)
}

// //
// Roles
// //
func SetAdminRoleInControlContext(ctx context.Context, role consts.AdminRole) context.Context {
	controlContext := GetControlContext(ctx)
	controlContext.AdminRole = role
	return context.WithValue(ctx, ControlContextKey, &controlContext)
}
