package nscollections

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log/slog"

	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/models/sitemodels"
	"nighttroop.t1d.nz/utils/can"
	"nighttroop.t1d.nz/utils/httputils"
	"nighttroop.t1d.nz/utils/security"
)

// The loop apps whose settings are kept in loop_settings, and the record each sends them in
const (
	loopApp           = "loop"
	trioApp           = "trio"
	aapsApp           = "aaps"
	fromProfile       = "profile"
	fromDeviceStatus  = "devicestatus"
	loopSettingsLimit = 8192
)

// loopSettingsID is the ID of an app's settings: the SHA-256 of the app and the settings, so the same settings are
// always the same row
func loopSettingsID(app, settingsJSON string) string {
	sum := sha256.Sum256([]byte(app + "\n" + settingsJSON))
	return hex.EncodeToString(sum[:])
}

// recordLoopSettings keeps an app's settings in loop_settings once, however often they are sent, and gives their ID
// for the record they came in to refer to. They are written through the app's model, so the same settings are the
// same text. Empty settings, as AAPS sends in most statuses, are none sent, and settings too large to keep are left
// out; neither has an ID, and the record they came in is still kept.
func recordLoopSettings(q *sitedb.Queries, ctx context.Context, app, source string, settings any, at int64, device *string) (*string, error) {
	settingsJSON := TryValueToJSON(settings)
	if settingsJSON == nil || len(*settingsJSON) > loopSettingsLimit {
		return nil, nil
	}
	// The deviceToken is not kept here: ProfileCollection.Stored keeps it encrypted in the app's push registration
	settingsJSON = ptr(string(sitemodels.WithoutDeviceToken(json.RawMessage(*settingsJSON))))
	if *settingsJSON == "" || *settingsJSON == "{}" || *settingsJSON == "null" {
		return nil, nil
	}
	id := loopSettingsID(app, *settingsJSON)
	err := q.UpsertLoopSettings(ctx, sitedb.UpsertLoopSettingsParams{ID: id, App: app, Source: source, SettingsJson: *settingsJSON, SeenAt: at, CreatedByDeviceName: device})
	if err != nil {
		return nil, err
	}
	return &id, nil
}

// loopSettingsJSON are the settings a record refers to, as the request's identity may see them, or none when it refers
// to none or they were deleted, as a record is kept when its settings are deleted
func loopSettingsJSON(ctx context.Context, siteContext *httputils.SiteContext, id *string) json.RawMessage {
	if id == nil || siteContext == nil || siteContext.SiteDB == nil {
		return nil
	}
	row, err := siteContext.SiteDB.Queries().GetLoopSettings(ctx, sitedb.GetLoopSettingsParams{ID: *id})
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			slog.WarnContext(ctx, "Cannot read a record's loop settings", "error", err)
		}
		return nil
	}
	return withDeviceToken(ctx, siteContext, row.App, json.RawMessage(row.SettingsJson))
}

// newestLoopSettingsJSON are Loop's last seen settings, as the request's identity may see them, for a record made at
// date that was sent with none. They are none if Loop sent them after that date, as the record is then not the newest.
// LoopFollow reads Loop's device token and override presets from the newest profile, which is not Loop's when AAPS or
// the careportal also send profiles to the site.
func newestLoopSettingsJSON(ctx context.Context, siteContext *httputils.SiteContext, date int64) json.RawMessage {
	if siteContext == nil || siteContext.SiteDB == nil {
		return nil
	}
	rows, err := siteContext.SiteDB.Queries().SearchLoopSettings(ctx, sitedb.SearchLoopSettingsParams{App: ptr(loopApp), Limit: 1})
	if err != nil {
		slog.WarnContext(ctx, "Cannot read Loop's newest settings", "error", err)
		return nil
	}
	if len(rows) == 0 || rows[0].LastSeenAt > date {
		return nil
	}
	return withDeviceToken(ctx, siteContext, loopApp, json.RawMessage(rows[0].SettingsJson))
}

// withDeviceToken are Loop's or Trio's settings with the app's deviceToken put back, for an identity that is given
// device tokens, as LoopFollow sending remote commands must be. The token is the one of the app's newest push
// registration, the only place it is kept, encrypted, and only when its bundle is the settings'. Any other identity,
// and AAPS's settings, get the settings as kept, with no token.
func withDeviceToken(ctx context.Context, siteContext *httputils.SiteContext, app string, settings json.RawMessage) json.RawMessage {
	settings = sitemodels.WithoutDeviceToken(settings)
	if settings == nil || app == aapsApp || !can.IdentitySeeDeviceTokens(siteContext.IdentityModel) {
		return settings
	}
	registration, err := siteContext.ControlDB.Queries.NewestPushRegistrationForApp(ctx, controldb.NewestPushRegistrationForAppParams{
		SiteID: siteContext.SiteID, App: app, Roles: can.LoopRegistrationRoles})
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			slog.WarnContext(ctx, "Cannot read an app's push registration for its device token", "app", app, "error", err)
		}
		return settings
	}
	var fields map[string]json.RawMessage
	bundle, _ := json.Marshal(registration.BundleID)
	if json.Unmarshal(settings, &fields) != nil || string(fields["bundleIdentifier"]) != string(bundle) {
		return settings
	}
	token, err := siteContext.Config.Security.DecryptFor(security.PurposePushToken, siteContext.SiteID, registration.TokenEnc)
	if err != nil {
		slog.WarnContext(ctx, "Cannot decrypt an app's device token", "app", app, "error", err)
		return settings
	}
	fields["deviceToken"], _ = json.Marshal(token)
	with, _ := json.Marshal(fields)
	return with
}
