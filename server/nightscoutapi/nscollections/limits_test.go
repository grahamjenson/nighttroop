package nscollections

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/models/nightscoutmodels"
	"nighttroop.t1d.nz/server/nightscoutapi/nsutils"
)

var (
	deviceStatuses = NewCollection[nightscoutmodels.DeviceStatus, sitedb.DeviceReport](DeviceStatusCollection{
		named[nightscoutmodels.DeviceStatus, sitedb.DeviceReport]{nsutils.Devicestatus, DBToNightscoutDeviceStatus}})
	entries = NewCollection[nightscoutmodels.Entry, sitedb.GlucoseReading](EntriesCollection{
		named[nightscoutmodels.Entry, sitedb.GlucoseReading]{nsutils.Entries, nightscoutmodels.DBModelToEntry}})
	treatments = NewCollection[nightscoutmodels.Treatment, sitedb.EventGroupsBundle](TreatmentCollection{
		named[nightscoutmodels.Treatment, sitedb.EventGroupsBundle]{nsutils.Treatments, DBModelToTreatment}})
	profiles = NewCollection[nightscoutmodels.ProfileStore, sitedb.ProfilePresetBundle](ProfileCollection{
		named[nightscoutmodels.ProfileStore, sitedb.ProfilePresetBundle]{nsutils.Profile, DBToNightscoutProfileStore}})
)

// checked is a document of a collection checked against its columns' limits, as a create checks it: the refusal, or
// what was left out and the data kept
func checked[NS SharedNSFields, DM SharedDBFields](c Collection[NS, DM], body string) (string, leftOut, map[string]any) {
	var data map[string]any
	if err := json.Unmarshal([]byte(body), &data); err != nil {
		panic(err)
	}
	var left leftOut
	nsModel, err := c.decode(data, &left)
	if err == nil {
		_, err = c.withinLimits(data, nsModel, &left)
	}
	if err != nil {
		return err.Message, left, data
	}
	return "", left, data
}

// Each value is checked against the limits generated from the site DB's schema: one a document cannot be stored
// without refuses it, naming the field, what its column takes and what was sent, and any other is left out and said
func TestWithinLimits(t *testing.T) {
	long := strings.Repeat("a", 70)
	tests := []struct {
		name    string
		check   func() (string, leftOut, map[string]any)
		refused string
		leftOut []string
		kept    []string
	}{
		// Required: its column is NOT NULL
		{"required in range", func() (string, leftOut, map[string]any) {
			return checked(entries, `{"type": "sgv", "sgv": 1000, "date": 1}`)
		}, "", nil, []string{"sgv"}},
		{"required out of range", func() (string, leftOut, map[string]any) {
			return checked(entries, `{"type": "sgv", "sgv": 1001, "date": 1}`)
		}, "Bad or missing body: sgv must be more than 0 and at most 1000 mg/dL (was 1001)", nil, nil},
		{"required out of range once converted", func() (string, leftOut, map[string]any) {
			return checked(entries, `{"type": "sgv", "sgv": 60, "units": "mmol/L", "date": 1}`)
		}, "Bad or missing body: sgv must be more than 0 and at most 1000 mg/dL (was 60, stored as 1081.08)", nil, nil},
		{"required missing", func() (string, leftOut, map[string]any) {
			return checked(entries, `{"type": "sgv", "date": 1}`)
		}, "Bad or missing body: mgdl must be more than 0 and at most 1000 mg/dL (was 0)", nil, nil},
		{"required wrong type", func() (string, leftOut, map[string]any) {
			return checked(entries, `{"type": "sgv", "sgv": "high", "date": 1}`)
		}, `Bad or missing body: sgv must be a number (was "high")`, nil, nil},
		{"required out of range in a treatment", func() (string, leftOut, map[string]any) {
			return checked(treatments, `{"eventType": "Correction Bolus", "insulin": 501, "date": 1}`)
		}, "Bad or missing body: insulin must be between -500 and 500 (was 501)", nil, nil},
		{"required wrong type in a treatment", func() (string, leftOut, map[string]any) {
			return checked(treatments, `{"eventType": "Correction Bolus", "insulin": "lots", "date": 1}`)
		}, `Bad or missing body: insulin must be a number (was "lots")`, nil, nil},
		{"required too long", func() (string, leftOut, map[string]any) {
			return checked(treatments, `{"eventType": "Announcement", "name": "`+strings.Repeat("a", 130)+`", "date": 1}`)
		}, "Bad or missing body: name must be at most 128 characters (was 130 characters)", nil, nil},
		// Optional: its column takes NULL
		{"optional in range", func() (string, leftOut, map[string]any) {
			return checked(deviceStatuses, `{"uploader": {"battery": 100}, "date": 1}`)
		}, "", nil, []string{"uploader.battery"}},
		{"optional out of range", func() (string, leftOut, map[string]any) {
			return checked(deviceStatuses, `{"uploader": {"battery": -100, "isCharging": true}, "date": 1}`)
		}, "", []string{"uploader.battery must be between 0 and 100 (was -100)"}, []string{"uploader.isCharging"}},
		{"optional too long", func() (string, leftOut, map[string]any) {
			return checked(entries, `{"type": "sgv", "sgv": 100, "device": "`+long+`", "date": 1}`)
		}, "", []string{"device must be at most 64 characters (was 70 characters)"}, []string{"sgv"}},
		{"optional too long in a profile store", func() (string, leftOut, map[string]any) {
			return checked(profiles, `{"enteredBy": "`+long+`", "store": {"Default": {"dia": 5}, "Other": {"dia": 4}}, "date": 1}`)
		}, "", []string{"enteredBy must be at most 64 characters (was 70 characters)"}, []string{"store.Default", "store.Other"}},
		{"optional too long in a profile", func() (string, leftOut, map[string]any) {
			return checked(profiles, `{"store": {"Default": {"dia": 5, "timezone": "`+long+`"}}, "date": 1}`)
		}, "", []string{"store.Default.timezone must be at most 32 characters (was 70 characters)"}, []string{"store.Default.dia"}},
		{"optional wrong type", func() (string, leftOut, map[string]any) {
			return checked(deviceStatuses, `{"uploader": {"battery": "50 days", "isCharging": true}, "date": 1}`)
		}, "", []string{`uploader.battery must be a number (was "50 days")`}, []string{"uploader.isCharging"}},
		{"optional out of range in a treatment", func() (string, leftOut, map[string]any) {
			return checked(treatments, `{"eventType": "Note", "notes": "ok", "duration": 600000, "date": 1}`)
		}, "", []string{"duration must be between 0 and 525600 (was 600000)"}, []string{"notes"}},

		// The exceptions
		{"a treatment's glucose is left out", func() (string, leftOut, map[string]any) {
			return checked(treatments, `{"eventType": "Bolus Wizard", "glucose": 0, "insulin": 1, "date": 1}`)
		}, "", []string{"glucose must be more than 0 and at most 1000 mg/dL (was 0)"}, []string{"insulin"}},
		{"a treatment's carbs refuse it", func() (string, leftOut, map[string]any) {
			return checked(treatments, `{"eventType": "Carb Correction", "carbs": 1001, "date": 1}`)
		}, "Bad or missing body: carbs must be between -1000 and 1000 (was 1001)", nil, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			refused, left, data := tt.check()
			require.Equal(t, tt.refused, refused)
			require.Equal(t, tt.leftOut, []string(left))
			for _, field := range tt.kept {
				require.True(t, has(data, strings.Split(field, ".")), field)
			}
			if refused == "" {
				for _, reason := range left {
					require.False(t, has(data, strings.Split(strings.Fields(reason)[0], ".")), reason)
				}
			}
		})
	}
}

// The limits are the site DB's: a number's range and a text's length, whether it may be NULL, and each column a field
// of its rows, by its name
func TestColumnLimits(t *testing.T) {
	tests := []struct {
		table, field string
		want         sitedb.ColumnLimit
	}{
		{"glucose_readings", "Mgdl", sitedb.ColumnLimit{Column: "mgdl", Required: true, Min: new(0.0), MinExclusive: true, Max: new(1000.0)}},
		{"insulin_doses", "Dose", sitedb.ColumnLimit{Column: "dose", Required: true, Min: new(-500.0), Max: new(500.0)}},
		{"food_intakes", "Carbs", sitedb.ColumnLimit{Column: "carbs", Min: new(-1000.0), Max: new(1000.0)}},
		{"device_reports", "DeviceBatteryPercent", sitedb.ColumnLimit{Column: "device_battery_percent", Min: new(0.0), Max: new(100.0)}},
		{"event_groups", "GroupType", sitedb.ColumnLimit{Column: "group_type", Required: true, MinLength: 1, MaxLength: 32}},
		{"event_groups", "SyncIdentifier", sitedb.ColumnLimit{Column: "sync_identifier", MaxLength: 256}},
		{"device_reports", "ConfigurationID", sitedb.ColumnLimit{Column: "configuration_id", MinLength: 64, MaxLength: 64}},
		{"general_settings", "Key", sitedb.ColumnLimit{Column: "key", Required: true, MaxLength: 128}},
		{"notification_events", "Severity", sitedb.ColumnLimit{Column: "severity", Required: true, Min: new(0.0), Max: new(5.0)}},
	}
	for _, tt := range tests {
		t.Run(tt.table+"."+tt.field, func(t *testing.T) {
			got, ok := sitedb.ColumnLimitOf(tt.table, tt.field)
			require.True(t, ok)
			require.Equal(t, tt.want, got)
		})
	}
}
