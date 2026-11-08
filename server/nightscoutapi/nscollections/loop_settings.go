package nscollections

import (
	"context"
	"database/sql"
	"errors"

	"nighttroop.t1d.nz/data/sql/sitedb"
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

// recordLoopSettings keeps an app's settings as a new loop_settings row when they differ from those in force when they
// were sent, so there is a row for each change, not each upload, and gives the row of the settings sent: the new one,
// or the one in force that they match, for the record they came in to refer to. Settings too large to keep are left
// out, with no row, as the record they came in is still kept.
func recordLoopSettings(q *sitedb.Queries, ctx context.Context, app, source string, settings any, at int64, device *string) (*int64, error) {
	settingsJSON := TryValueToJSON(settings)
	if settingsJSON == nil || len(*settingsJSON) > loopSettingsLimit {
		return nil, nil
	}
	inForce, err := q.GetLoopSettingsAt(ctx, sitedb.GetLoopSettingsAtParams{App: app, At: at})
	switch {
	case err == nil && inForce.SettingsJson == *settingsJSON:
		return &inForce.ID, nil
	case err != nil && !errors.Is(err, sql.ErrNoRows):
		return nil, err
	}
	row, err := q.CreateLoopSettings(ctx, sitedb.CreateLoopSettingsParams{App: app, Source: source, SettingsJson: *settingsJSON, OccurredAt: at, CreatedByDeviceName: device})
	if err != nil {
		return nil, err
	}
	return &row.ID, nil
}
