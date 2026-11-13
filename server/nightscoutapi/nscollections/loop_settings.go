package nscollections

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log/slog"

	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/models/sitemodels"
	"nighttroop.t1d.nz/utils/httputils"
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
	if settingsJSON == nil || *settingsJSON == "{}" || *settingsJSON == "null" || len(*settingsJSON) > loopSettingsLimit {
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
	return sitemodels.LoopSettingsFor(siteContext.IdentityModel, json.RawMessage(row.SettingsJson))
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
	return sitemodels.LoopSettingsFor(siteContext.IdentityModel, json.RawMessage(rows[0].SettingsJson))
}
