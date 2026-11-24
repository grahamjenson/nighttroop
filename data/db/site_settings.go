package db

import (
	"context"
	"slices"
	"strconv"

	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/utils/settings"
)

// A site's settings and each identity's are rows of the site DB's general_settings, of the type 'settings': one row
// a level, a JSON object of the settings set at it. So a new setting is a new key, not a new column on two tables.
//
// Fast: a request reads both its levels in one query, two lookups by primary key.
// Safe: a site has one row and an identity one, so the rows are bounded by the identities a site may have, and a
// row by settings.MaxJSONBytes. Only this file writes them, from a SettingMap, which holds only known settings with
// validated values; the Nightscout settings API cannot name these keys, and its queries leave the type out.

// siteSettingsKey is the key of the site's own settings, and identitySettingsKey of an identity's
const siteSettingsKey = "nighttroop_site"

func identitySettingsKey(identityID int64) string {
	return "nighttroop_identity_" + strconv.FormatInt(identityID, 10)
}

// LoadSettings is the site's settings and the identity's, each empty when it has no row. An identityID of 0 is no
// identity, so only the site's are read.
func LoadSettings(ctx context.Context, q *sitedb.Queries, identityID int64) (site, identity *settings.SettingMap, err error) {
	site, identity = &settings.SettingMap{}, &settings.SettingMap{}
	identityKey := identitySettingsKey(identityID)
	rows, err := q.GetSettings(ctx, sitedb.GetSettingsParams{SiteKey: siteSettingsKey, IdentityKey: identityKey})
	if err != nil {
		return site, identity, err
	}
	for _, row := range rows {
		switch {
		case row.Key == siteSettingsKey:
			site = settings.FromJSON([]byte(row.ValueJson), &row.UpdatedAt, new("site"))
		case row.Key == identityKey && identityID != 0:
			identity = settings.FromJSON([]byte(row.ValueJson), &row.UpdatedAt, new("identity"))
		}
	}
	return site, identity, nil
}

// SaveSettings replaces the site's settings, or the identity's when identityID is not 0, with those m has set
func SaveSettings(ctx context.Context, q *sitedb.Queries, identityID int64, m *settings.SettingMap, by string) error {
	raw, err := m.JSON()
	if err != nil {
		return err
	}
	key := siteSettingsKey
	if identityID != 0 {
		key = identitySettingsKey(identityID)
	}
	return q.UpsertSettings(ctx, sitedb.UpsertSettingsParams{Key: key, ValueJson: string(raw), By: by})
}

// SaveSiteSettings saves a level as SaveSettings does, then keeps the control DB's alarms_on for the site: whether
// any level of the site now turns an alarm on
func SaveSiteSettings(ctx context.Context, q *sitedb.Queries, control *controldb.Queries, siteID string, identityID int64, m *settings.SettingMap, by string) error {
	if err := SaveSettings(ctx, q, identityID, m, by); err != nil {
		return err
	}
	rows, err := q.AllSettings(ctx)
	if err != nil {
		return err
	}
	isOn := func(enabled *bool) bool { return enabled != nil && *enabled }
	alarmsOn := slices.ContainsFunc(rows, func(row *sitedb.AllSettingsRow) bool {
		level := settings.FromJSON([]byte(row.ValueJson), nil, nil)
		return isOn(level.GetAlarmUrgentLowEnabled()) || isOn(level.GetAlarmLowEnabled()) || isOn(level.GetAlarmUrgentHighEnabled()) ||
			isOn(level.GetAlarmHighEnabled()) || isOn(level.GetAlarmNoDataEnabled()) || isOn(level.GetAlarmDeviceBatteryLowEnabled())
	})
	on := int64(0)
	if alarmsOn {
		on = 1
	}
	return control.SetSiteAlarmsOn(ctx, controldb.SetSiteAlarmsOnParams{ID: siteID, AlarmsOn: on})
}

// DeleteIdentitySettings deletes an identity's settings, when the identity is deleted
func DeleteIdentitySettings(ctx context.Context, q *sitedb.Queries, identityID int64) error {
	return q.DeleteSettings(ctx, sitedb.DeleteSettingsParams{Key: identitySettingsKey(identityID)})
}
