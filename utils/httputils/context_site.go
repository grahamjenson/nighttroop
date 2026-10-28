package httputils

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/events"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/settings"
)

const siteContextKey = "user_context"

type SiteContext struct {
	Config *config.ServerConfig

	ControlDB *db.ControlDB
	SiteDB    *db.SiteDB
	SiteModel *controldb.Site

	IdentityModel *controldb.SiteIdentity

	SiteID    string
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

func CreateSiteContext(ctx context.Context, config *config.ServerConfig, controlDB *db.ControlDB, siteDB *db.SiteDB, siteModel *controldb.Site, siteID string, ipAddress string, httpAction HTTPAction, acceptType AcceptType, contentType ContentType) context.Context {
	siteContext := &SiteContext{}
	siteContext.Config = config
	siteContext.ControlDB = controlDB
	siteContext.SiteDB = siteDB
	siteContext.SiteModel = siteModel
	siteContext.SiteID = siteID
	siteContext.IPAddress = ipAddress
	siteContext.HTTPAction = httpAction
	siteContext.AcceptType = acceptType
	siteContext.ContentType = contentType

	return context.WithValue(ctx, siteContextKey, siteContext)
}

// PublishModelChange sends the event of a committed change to one of the site's models, the sitedb row, through the
// server's events
func (s *SiteContext) PublishModelChange(action events.Action, model any) {
	if s.Config != nil && s.Config.Events != nil {
		s.Config.Events.Publish(s.SiteID, events.ModelChange(s.SiteID, action, model))
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
		slog.Error("settings unavailable site or identity model is nil", "site_id", sc.SiteID, "siteModel", sc.SiteModel, "identityModel", sc.IdentityModel)
		return settings.DefaultCombinedSettingsMap()
	}

	if sc.settings == nil {
		sc.settings = settings.NewCombinedSettingsMap(sc.SiteModel, sc.IdentityModel)
	}

	return sc.settings
}

// GetSiteIdentity is the enabled identity with the name and role on the site of the request, refusing one for another site
func GetSiteIdentity(ctx context.Context, siteID string, identityName string, role consts.IdentityRole) (*controldb.SiteIdentity, error) {
	siteContext := GetSiteContext(ctx)
	if siteID != siteContext.SiteID {
		return nil, fmt.Errorf("site ID mismatch: expected %s, got %s", siteContext.SiteID, siteID)
	}

	identity, err := siteContext.ControlDB.Queries.GetSiteIdentity(ctx, controldb.GetSiteIdentityParams{
		SiteID: siteID,
		Name:   identityName,
	})
	if err != nil {
		return nil, err
	}

	if !identity.Enabled {
		return nil, fmt.Errorf("identity disabled")
	}

	if consts.IdentityRole(identity.Role) != role {
		return nil, fmt.Errorf("role mismatch")
	}

	return identity, nil
}

// RecordAuthClient records the identity authenticating from the request's app and network, for its site's admin to
// see an identity shared by several
func RecordAuthClient(r *http.Request, identity *controldb.SiteIdentity) {
	siteContext := GetSiteContext(r.Context())
	if identity == nil || siteContext == nil || siteContext.Config == nil || siteContext.Config.RecentClients == nil {
		return
	}
	siteContext.Config.RecentClients.Add(AuthClientKey(identity.SiteID, identity.ID), r.UserAgent(), GetIPFromHTTPContext(r.Context()), time.Now())
}

// AuthClientKey is an identity's key in the server's recent clients
func AuthClientKey(siteID string, identityID int64) string {
	return siteID + "/" + strconv.FormatInt(identityID, 10)
}
