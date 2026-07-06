package httputils

import (
	"context"

	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/sql/control"
)

type UserAction string

const (
	UserActionRead  UserAction = "read"
	UserActionWrite UserAction = "write"
)

const UserContextKey = "user_context"

type UserContext struct {
	ControlDB *db.ControlDB
	SiteDB    *db.UserDB
	SiteModel *control.Site

	IdentityModel *control.SiteIdentity

	Subdomain string

	UserAction UserAction
}

func GetUserContext(ctx context.Context) *UserContext {
	if ctx == nil {
		return nil
	}
	if userCtx, ok := ctx.Value(UserContextKey).(*UserContext); ok {
		return userCtx
	}
	return nil
}

// //
// Action
// //
func SetUserActionInUserContext(ctx context.Context, action UserAction) context.Context {
	userContext := GetUserContext(ctx)
	userContext.UserAction = action
	return context.WithValue(ctx, UserContextKey, userContext)
}

////
// UserDB
////

func SetUserDBAndModelAndCreateUserContext(ctx context.Context, controlDB *db.ControlDB, siteDB *db.UserDB, siteModel *control.Site, subdomain string) context.Context {
	userContext := &UserContext{}
	userContext.ControlDB = controlDB
	userContext.SiteDB = siteDB
	userContext.SiteModel = siteModel
	userContext.Subdomain = subdomain

	return context.WithValue(ctx, UserContextKey, userContext)
}

// //
// IdentityModel
// //
func SetIdentityModelInUserContext(ctx context.Context, identityModel *control.SiteIdentity) context.Context {
	userContext := GetUserContext(ctx)
	userContext.IdentityModel = identityModel
	return context.WithValue(ctx, UserContextKey, userContext)
}
