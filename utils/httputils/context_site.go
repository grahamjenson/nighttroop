package httputils

import (
	"context"
	"log/slog"

	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/sql/control"
	"nighttroop.t1d.nz/utils/settings"
)

type SiteAction string

const (
	SiteActionRead  SiteAction = "read"
	SiteActionWrite SiteAction = "write"
)

const siteContextKey = "user_context"

type SiteContext struct {
	ControlDB *db.ControlDB
	SiteDB    *db.SiteDB
	SiteModel *control.Site

	IdentityModel *control.SiteIdentity

	Subdomain string
	IPAddress string

	SiteAction SiteAction

	settings *settings.CombinedSettingsMap
}

func GetSiteContext(ctx context.Context) *SiteContext {
	if ctx == nil {
		return nil
	}
	if userCtx, ok := ctx.Value(siteContextKey).(*SiteContext); ok {
		return userCtx
	}
	return nil
}

////
// SiteDB
////

func CreateSiteContext(ctx context.Context, controlDB *db.ControlDB, siteDB *db.SiteDB, siteModel *control.Site, subdomain string, ipAddress string) context.Context {
	siteContext := &SiteContext{}
	siteContext.ControlDB = controlDB
	siteContext.SiteDB = siteDB
	siteContext.SiteModel = siteModel
	siteContext.Subdomain = subdomain
	siteContext.IPAddress = ipAddress

	return context.WithValue(ctx, siteContextKey, siteContext)
}

// //
// Action
// //
func SetSiteActionInSiteContext(ctx context.Context, action SiteAction) context.Context {
	siteContext := GetSiteContext(ctx)
	siteContext.SiteAction = action
	return context.WithValue(ctx, siteContextKey, siteContext)
}

// //
// IdentityModel
// //
func SetIdentityModelInSiteContext(ctx context.Context, identityModel *control.SiteIdentity) context.Context {
	siteContext := GetSiteContext(ctx)
	siteContext.IdentityModel = identityModel
	return context.WithValue(ctx, siteContextKey, siteContext)
}

// /////
// SETTINGS
// /////

func (sc *SiteContext) GetSettings() *settings.CombinedSettingsMap {
	if sc.SiteModel == nil || sc.IdentityModel == nil {
		slog.Error("settings unavailable site or identity model is nil", "subdomain", sc.Subdomain, "siteModel", sc.SiteModel, "identityModel", sc.IdentityModel)
		return settings.DefaultCombinedSettingsMap()
	}

	if sc.settings == nil {
		sc.settings = settings.NewCombinedSettingsMap(sc.SiteModel, sc.IdentityModel)
	}

	return sc.settings
}
