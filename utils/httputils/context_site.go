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
	"nighttroop.t1d.nz/utils/aslog"
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

	// settings are the site's settings and the identity's, built once from their rows
	settings *settings.Settings
	// mgdl is whether glucose is always in mg/dL, whatever the identity's or site's units, and mgdlSettings the
	// settings in effect with them
	mgdl         bool
	mgdlSettings *settings.Settings
}

// SetMgdlUnits makes glucose read and written in the context always mg/dL, whatever the identity's or site's units,
// as Nightscout's API has it: its clients read sgv, mbg and targets as mg/dL whatever the units
func SetMgdlUnits(ctx context.Context) context.Context {
	if siteContext := GetSiteContext(ctx); siteContext != nil {
		siteContext.mgdl, siteContext.mgdlSettings = true, nil
	}
	return ctx
}

// MgdlUnitsMiddleware gives every request glucose in mg/dL, as SetMgdlUnits does
func MgdlUnitsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r.WithContext(SetMgdlUnits(r.Context())))
	})
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
	siteContext.settings, siteContext.mgdlSettings = nil, nil
	if identityModel != nil {
		aslog.AddAttrs(ctx, slog.Int64("identity_id", identityModel.ID), slog.String("identity_name", identityModel.Name), slog.String("role", string(identityModel.Role)))
	}
	return context.WithValue(ctx, siteContextKey, siteContext)
}

// /////
// SETTINGS
// /////

// GetSettings are the request's settings: Values in effect, the identity's over the site's over the defaults, and
// Site, the site's own whoever is signed in. Where glucose is always mg/dL, the units in effect are mg/dL.
func (sc *SiteContext) GetSettings() *settings.Settings {
	if !sc.mgdl {
		return sc.GetOwnSettings()
	}
	if sc.mgdlSettings == nil {
		mgdl := *sc.GetOwnSettings()
		mgdl.Units = consts.UnitsMGDL
		sc.mgdlSettings = &mgdl
	}
	return sc.mgdlSettings
}

// GetOwnSettings are the settings as the identity and site have them, their own units even where glucose is always
// mg/dL: for showing glucose in those units, and for reading and writing the settings themselves. Without a signed in
// identity, such as for a live update sent to a site's sockets, the site's own settings are in effect.
func (sc *SiteContext) GetOwnSettings() *settings.Settings {
	if sc.settings == nil {
		if sc.SiteModel == nil {
			// Whether each is known, not the models, which hold the backup key's hash and the identity's secrets
			slog.Error("settings unavailable, as the site is not known, so the defaults are used", "site_id", sc.SiteID, "has_identity", sc.IdentityModel != nil)
		}
		sc.settings = settings.New(sc.SiteModel, sc.IdentityModel)
	}
	return sc.settings
}

// GetSiteTimezone is the site's own timezone, or UTC, whatever the identity's
func (sc *SiteContext) GetSiteTimezone() *time.Location {
	return sc.GetOwnSettings().Site.Location()
}

// GetSiteWeekStartDay is the site's own week start day, monday or sunday, whatever the identity's
func (sc *SiteContext) GetSiteWeekStartDay() string {
	return sc.GetOwnSettings().Site.WeekStartDay
}

// GetSiteRangeMgdl is the site's own glucose range, low and high in mg/dL, whatever the identity's
func (sc *SiteContext) GetSiteRangeMgdl() (low, high int64) {
	site := sc.GetOwnSettings().Site.Range
	return site.LowMgdl, site.HighMgdl
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

// GetJWTSiteIdentity is the site identity a user JWT is for. A JWT issued before the identity's password or TOTP last
// changed is refused, so changing them signs out whoever had the old ones.
func GetJWTSiteIdentity(ctx context.Context, cfg *config.ServerConfig, token string) (*controldb.SiteIdentity, error) {
	subdomain, identityName, role, issuedAtMs, err := cfg.Security.VerifyUserJWTToken(token)
	if err != nil {
		return nil, fmt.Errorf("bad JWT token: %w", err)
	}
	identity, err := GetSiteIdentity(ctx, subdomain, identityName, consts.IdentityRole(role))
	if err != nil {
		slog.WarnContext(ctx, "Bad JWT token cant find identity", "error", err)
		return nil, fmt.Errorf("identity not found: %w", err)
	}
	if err := CheckIssuedAfterCredentialsChanged(identity, issuedAtMs); err != nil {
		return nil, err
	}
	return identity, nil
}

// CheckIssuedAfterCredentialsChanged refuses a JWT issued, in milliseconds, before its site or control identity's
// credentials changed, or in the same millisecond, which can be before the change. A sign in after a change checks its
// password first, which takes longer than that.
func CheckIssuedAfterCredentialsChanged(identity interface{ GetCredentialsChangedAt() int64 }, issuedAtMs int64) error {
	if issuedAtMs <= identity.GetCredentialsChangedAt() {
		return fmt.Errorf("JWT was issued before the identity's credentials changed")
	}
	return nil
}

// IdentityRevoked is whether identity, as it was when it authenticated, has since been deleted, disabled or given
// new credentials. A socket or event stream is not authenticated again on each message, so checks it to stop when a
// request would be refused.
func IdentityRevoked(ctx context.Context, controlDB *db.ControlDB, identity *controldb.SiteIdentity) bool {
	current, err := controlDB.Queries.GetSiteIdentityByID(ctx, controldb.GetSiteIdentityByIDParams{SiteID: identity.SiteID, ID: identity.ID})
	return err != nil || !current.Enabled || current.CredentialsChangedAt != identity.CredentialsChangedAt
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
