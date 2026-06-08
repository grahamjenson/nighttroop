package httputils

import (
	"context"

	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/sql/control"
	"nighttroop.t1d.nz/data/sql/user"
)

type UserAction string

const (
	UserActionRead  UserAction = "read"
	UserActionWrite UserAction = "write"
)

const UserContextKey = "user_context"

type UserContext struct {
	UserDB    *db.UserDB
	UserModel *control.User

	IdentityModel *user.Identity

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

// //
// Subdomain
// //
func SetSubdomainInUserContext(ctx context.Context, subdomain string) context.Context {
	userContext := GetUserContext(ctx)
	userContext.Subdomain = subdomain
	return context.WithValue(ctx, UserContextKey, userContext)
}

////
// UserDB
////

func SetUserDBAndModelAndCreateUserContext(ctx context.Context, userDB *db.UserDB, userGetModel *control.User) context.Context {
	userContext := &UserContext{}
	userContext.UserDB = userDB
	userContext.UserModel = userGetModel

	return context.WithValue(ctx, UserContextKey, userContext)
}

// //
// IdentityModel
// //
func SetIdentityModelInUserContext(ctx context.Context, identityModel *user.Identity) context.Context {
	userContext := GetUserContext(ctx)
	userContext.IdentityModel = identityModel
	return context.WithValue(ctx, UserContextKey, userContext)
}
