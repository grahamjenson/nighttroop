package httputils

import (
	"context"
	"log/slog"

	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/events/sitemodels"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/settings"
)

const siteContextKey = "user_context"

type SiteContext struct {
	Config *config.ServerConfig

	ControlDB *db.ControlDB
	SiteDB    *db.SiteDB
	SiteModel *controldb.Site

	IdentityModel *controldb.SiteIdentity

	Subdomain string
	IPAddress string

	HTTPAction  HTTPAction
	AcceptType  AcceptType
	ContentType ContentType

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

func CreateSiteContext(ctx context.Context, config *config.ServerConfig, controlDB *db.ControlDB, siteDB *db.SiteDB, siteModel *controldb.Site, subdomain string, ipAddress string, httpAction HTTPAction, acceptType AcceptType, contentType ContentType) context.Context {
	siteContext := &SiteContext{}
	siteContext.Config = config
	siteContext.ControlDB = controlDB
	siteContext.SiteDB = siteDB
	siteContext.SiteModel = siteModel
	siteContext.Subdomain = subdomain
	siteContext.IPAddress = ipAddress
	siteContext.HTTPAction = httpAction
	siteContext.AcceptType = acceptType
	siteContext.ContentType = contentType

	return context.WithValue(ctx, siteContextKey, siteContext)
}

// PublishModelChange tells the site's model events of a committed change to one of its models, the sitedb row
func (s *SiteContext) PublishModelChange(action sitemodels.Action, model any) {
	if s.Config != nil && s.Config.SiteModelEvents != nil {
		s.Config.SiteModelEvents.Publish(s.Subdomain, sitemodels.Event{SiteID: s.Subdomain, Action: action, Model: model})
	}
}

// //
// Action
// //

// //
// IdentityModel
// //
func SetIdentityModelInSiteContext(ctx context.Context, identityModel *controldb.SiteIdentity) context.Context {
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
