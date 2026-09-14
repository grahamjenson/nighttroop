package httputils

import (
	"context"
	"fmt"

	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/sql/control"
	"nighttroop.t1d.nz/utils/settings"
)

type SiteAction string

const (
	SiteActionRead  SiteAction = "read"
	SiteActionWrite SiteAction = "write"
)

const SiteContextKey = "user_context"

type SiteContext struct {
	ControlDB *db.ControlDB
	SiteDB    *db.SiteDB
	SiteModel *control.Site

	IdentityModel *control.SiteIdentity

	Subdomain string

	SiteAction SiteAction

	settings *settings.CombinedSettingsMap
}

func GetSiteContext(ctx context.Context) *SiteContext {
	if ctx == nil {
		return nil
	}
	if userCtx, ok := ctx.Value(SiteContextKey).(*SiteContext); ok {
		return userCtx
	}
	return nil
}

// //
// Action
// //
func SetSiteActionInSiteContext(ctx context.Context, action SiteAction) context.Context {
	siteContext := GetSiteContext(ctx)
	siteContext.SiteAction = action
	return context.WithValue(ctx, SiteContextKey, siteContext)
}

////
// SiteDB
////

func SetSiteDBAndModelAndCreateSiteContext(ctx context.Context, controlDB *db.ControlDB, siteDB *db.SiteDB, siteModel *control.Site, subdomain string) context.Context {
	siteContext := &SiteContext{}
	siteContext.ControlDB = controlDB
	siteContext.SiteDB = siteDB
	siteContext.SiteModel = siteModel
	siteContext.Subdomain = subdomain

	return context.WithValue(ctx, SiteContextKey, siteContext)
}

// //
// IdentityModel
// //
func SetIdentityModelInSiteContext(ctx context.Context, identityModel *control.SiteIdentity) context.Context {
	siteContext := GetSiteContext(ctx)
	siteContext.IdentityModel = identityModel
	return context.WithValue(ctx, SiteContextKey, siteContext)
}

// /////
// SETTINGS
// /////

func (sc *SiteContext) GetSettings() (*settings.CombinedSettingsMap, error) {
	if sc.SiteModel == nil || sc.IdentityModel == nil {
		return nil, fmt.Errorf("settings unavailable site or identity model is nil")
	}

	if sc.settings == nil {
		sc.settings = settings.NewCombinedSettingsMap(sc.SiteModel, sc.IdentityModel)
	}

	return sc.settings, nil
}
