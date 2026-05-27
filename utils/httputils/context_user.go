package httputils

import (
	"context"

	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/sql/control"
	"nighttroop.t1d.nz/utils/consts"
)

const UserContextKey = "user_context"

type UserContext struct {
	UserDB       *db.UserDB
	UserModel    *control.User
	Subdomain    string
	IdentityName string
	IdentityRole consts.IdentityRole
}

func GetUserContext(ctx context.Context) UserContext {
	if ctx == nil {
		return UserContext{}
	}
	if userCtx, ok := ctx.Value(UserContextKey).(*UserContext); ok {
		return *userCtx
	}
	return UserContext{}
}

// //
// Subdomain
// //
func SetSubdomainInUserContext(ctx context.Context, subdomain string) context.Context {
	userContext := GetUserContext(ctx)
	userContext.Subdomain = subdomain
	return context.WithValue(ctx, UserContextKey, &userContext)
}

////
// UserDB
////

func SetUserDBAndModelInUserContext(ctx context.Context, userDB *db.UserDB, userGetModel *control.User) context.Context {
	userContext := GetUserContext(ctx)
	userContext.UserDB = userDB
	userContext.UserModel = userGetModel

	return context.WithValue(ctx, UserContextKey, &userContext)
}

// //
// IdentityName
// //
func SetIdentityNameInUserContext(ctx context.Context, identityName string) context.Context {
	userContext := GetUserContext(ctx)
	userContext.IdentityName = identityName
	return context.WithValue(ctx, UserContextKey, &userContext)
}

func SetIdentityRoleInUserContext(ctx context.Context, role consts.IdentityRole) context.Context {
	userContext := GetUserContext(ctx)
	userContext.IdentityRole = role
	return context.WithValue(ctx, UserContextKey, &userContext)
}
