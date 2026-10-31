package httputils

import (
	"context"
	"log/slog"

	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/utils/aslog"
	"nighttroop.t1d.nz/utils/config"
)

const controlContextKey = "control_context"

type ControlContext struct {
	IdentityModel *controldb.ControlIdentity

	ControlDB *db.ControlDB

	Subdomain string
	IPAddress string

	HTTPAction HTTPAction

	// ServerStatus is set for signed-in identities, so control pages can warn about missing services
	ServerStatus config.ServerStatus
}

func GetControlContext(ctx context.Context) *ControlContext {
	if ctx == nil {
		return nil
	}
	if controlCtx, ok := ctx.Value(controlContextKey).(*ControlContext); ok {
		return controlCtx
	}
	return nil
}

func CreateControlContext(ctx context.Context, controlDB *db.ControlDB, ipAddress string, httpAction HTTPAction) context.Context {
	controlContext := &ControlContext{
		ControlDB:  controlDB,
		IPAddress:  ipAddress,
		HTTPAction: httpAction,
	}
	return context.WithValue(ctx, controlContextKey, controlContext)
}

func SetAdminControlContext(ctx context.Context, identity *controldb.ControlIdentity, status config.ServerStatus) context.Context {
	controlContext := GetControlContext(ctx)
	controlContext.IdentityModel = identity
	controlContext.ServerStatus = status
	if identity != nil {
		aslog.AddAttrs(ctx, slog.Int64("identity_id", identity.ID), slog.String("identity_name", identity.Name), slog.String("role", string(identity.Role)))
	}
	return context.WithValue(ctx, controlContextKey, controlContext)
}
