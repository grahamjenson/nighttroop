package sitemodels

import (
	"encoding/json"

	"nighttroop.t1d.nz/data/sql/sitedb"
)

// SiteLoopSettings are a set of a loop app's settings, kept once however often they were sent, with when they were
// first and last seen. ID is the SHA-256 of the app and the settings.
type SiteLoopSettings struct {
	ID                  string          `json:"id"`
	App                 string          `json:"app"`
	Source              string          `json:"source"`
	Settings            json.RawMessage `json:"settings"`
	FirstSeenAt         int64           `json:"first_seen_at"`
	LastSeenAt          int64           `json:"last_seen_at"`
	CreatedByDeviceName *string         `json:"created_by_device_name"`
	CreatedAt           int64           `json:"created_at"`
}

// LoopSettingsFromDB is the site model of a loop_settings row
func LoopSettingsFromDB(row *sitedb.LoopSetting) *SiteLoopSettings {
	if row == nil {
		return nil
	}
	return &SiteLoopSettings{ID: row.ID, App: row.App, Source: row.Source, Settings: json.RawMessage(row.SettingsJson),
		FirstSeenAt: row.FirstSeenAt, LastSeenAt: row.LastSeenAt, CreatedByDeviceName: row.CreatedByDeviceName, CreatedAt: row.CreatedAt}
}
