package settings

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"nighttroop.t1d.nz/data/types"
	"nighttroop.t1d.nz/utils/consts"
)

// Version is the version of Settings that is written. It changes only when what a stored value means changes, with
// code here that reads the older versions: a new field does not change it.
const Version = 1

// MaxJSONBytes is the most a Settings takes as stored, as its column allows. Every setting set, with the longest
// values each allows (10 device names of 64 characters), is under half of it.
const MaxJSONBytes = 4096

// Settings is what the site, or an identity, has set of its own, as its settings_json column holds it. A setting that
// is not set is nil and is left out of the JSON, so the site's value, then the default, shows through.
//
// Rows of every release are read into this struct, so it only grows: a field is added, and never renamed, removed,
// moved or given another type or JSON name. TestSettingsOnlyGrow fails when one is. AGENTS.md has the rule.
// SettingsValues has the same fields with no pointers.
type Settings struct {
	Version int `json:"version"`

	Units        *consts.Units `json:"units,omitempty"`
	Timezone     *string       `json:"timezone,omitempty"`
	Language     *string       `json:"language,omitempty"`
	DateFormat   *string       `json:"date_format,omitempty"`
	TimeFormat   *string       `json:"time_format,omitempty"`
	WeekStartDay *string       `json:"week_start_day,omitempty"`
	NumberFormat *string       `json:"number_format,omitempty"`

	Bolus         Insulin       `json:"bolus,omitzero"`
	Basal         Insulin       `json:"basal,omitzero"`
	Defaults      Defaults      `json:"defaults,omitzero"`
	Carbs         Carbs         `json:"carbs,omitzero"`
	Range         GlucoseRange  `json:"range,omitzero"`
	Alarms        Alarms        `json:"alarms,omitzero"`
	Notifications Notifications `json:"notifications,omitzero"`
}

// Insulin is how an insulin acts: how long, and when it starts and peaks
type Insulin struct {
	DiaHours     *float64 `json:"dia_hours,omitempty"`
	OnsetMinutes *float64 `json:"onset_minutes,omitempty"`
	PeakMinutes  *float64 `json:"peak_minutes,omitempty"`
}

// Defaults are the profile values used when a site has no profile
type Defaults struct {
	InsulinSensitivity *float64 `json:"insulin_sensitivity,omitempty"`
	CarbRatio          *float64 `json:"carb_ratio,omitempty"`
	BasalRate          *float64 `json:"basal_rate,omitempty"`
}

// Carbs is how carbs are absorbed
type Carbs struct {
	MaxAbsorptionHours *float64 `json:"max_absorption_hours,omitempty"`
	Min5mImpact        *float64 `json:"min_5m_impact,omitempty"`
}

// GlucoseRange is the glucose range in mg/dL, which the graphs, the summaries and the glucose alarms share
type GlucoseRange struct {
	UrgentLowMgdl  *int64 `json:"urgent_low_mgdl,omitempty"`
	LowMgdl        *int64 `json:"low_mgdl,omitempty"`
	HighMgdl       *int64 `json:"high_mgdl,omitempty"`
	UrgentHighMgdl *int64 `json:"urgent_high_mgdl,omitempty"`
}

// Alarms is each alarm: whether it is on, and what it waits for
type Alarms struct {
	UrgentLow        Alarm          `json:"urgent_low,omitzero"`
	Low              TimedAlarm     `json:"low,omitzero"`
	LowPrediction    TimedAlarm     `json:"low_prediction,omitzero"`
	High             TimedAlarm     `json:"high,omitzero"`
	UrgentHigh       Alarm          `json:"urgent_high,omitzero"`
	FastRise         ChangeAlarm    `json:"fast_rise,omitzero"`
	FastDrop         ChangeAlarm    `json:"fast_drop,omitzero"`
	NoData           TimedAlarm     `json:"no_data,omitzero"`
	DeviceBatteryLow BatteryAlarm   `json:"device_battery_low,omitzero"`
	LowReservoir     ReservoirAlarm `json:"low_reservoir,omitzero"`
}

// Alarm is an alarm that is on or off
type Alarm struct {
	Enabled *bool `json:"enabled,omitempty"`
}

// TimedAlarm is an alarm that waits for, or looks ahead, some minutes
type TimedAlarm struct {
	Enabled *bool  `json:"enabled,omitempty"`
	Minutes *int64 `json:"minutes,omitempty"`
}

// ChangeAlarm is an alarm for glucose changing by so many mg/dL over some minutes
type ChangeAlarm struct {
	Enabled *bool  `json:"enabled,omitempty"`
	Mgdl    *int64 `json:"mgdl,omitempty"`
	Minutes *int64 `json:"minutes,omitempty"`
}

// BatteryAlarm is the alarm for the listed devices' batteries, and for one going quiet
type BatteryAlarm struct {
	Enabled *bool              `json:"enabled,omitempty"`
	Percent *int64             `json:"percent,omitempty"`
	Minutes *int64             `json:"minutes,omitempty"`
	Devices *types.DeviceNames `json:"devices,omitempty"`
}

// ReservoirAlarm is the alarm for a pump's reservoir running low
type ReservoirAlarm struct {
	Enabled *bool  `json:"enabled,omitempty"`
	Units   *int64 `json:"units,omitempty"`
}

// Notifications is whether the identity is sent notifications
type Notifications struct {
	Enabled *bool `json:"enabled,omitempty"`
}

// SettingsFromJSON is the Settings a settings_json column holds. Each value is validated as one sent to the API is,
// so a value out of its range, not one of its options or of the wrong type is left out and the rest read: a row is
// never trusted because it was stored. A row that is too long, or is not JSON, has no settings. The logs say so.
func SettingsFromJSON(raw string) Settings {
	var s Settings
	// A row made in code and not read from its table has none
	if raw == "" {
		return s
	}
	if len(raw) > MaxJSONBytes {
		slog.Error("stored settings are too long, so none are used", "bytes", len(raw))
		return s
	}
	// A field of the wrong type is skipped and the rest read. Units that are not units stop the read, so the row is
	// read again without them.
	var wrongType *json.UnmarshalTypeError
	err := json.Unmarshal([]byte(raw), &s)
	if err != nil && !errors.As(err, &wrongType) {
		var row map[string]json.RawMessage
		if json.Unmarshal([]byte(raw), &row) != nil {
			slog.Error("stored settings cannot be read, so none are used", "error", err)
			return Settings{}
		}
		delete(row, "units")
		without, _ := json.Marshal(row)
		s = Settings{}
		if err := json.Unmarshal(without, &s); err != nil && !errors.As(err, &wrongType) {
			slog.Error("stored settings cannot be read, so none are used", "error", err)
			return Settings{}
		}
	}
	for _, f := range allSettings {
		if value, ok := f.own(&s); ok {
			if err := f.validate(value); err != nil {
				slog.Warn("a stored setting is not valid, so it is not used", "setting", f.identifier)
				f.clear(&s)
			}
		}
	}
	return s
}

// JSON is the Settings as its column holds it, at Version
func (s Settings) JSON() (string, error) {
	s.Version = Version
	raw, err := json.Marshal(s)
	if err == nil && len(raw) > MaxJSONBytes {
		err = fmt.Errorf("settings are %d bytes, over %d", len(raw), MaxJSONBytes)
	}
	return string(raw), err
}

// StoredOf is the Settings as its column holds it
func StoredOf(s Settings) string {
	stored, _ := s.JSON()
	return stored
}

// Set sets a setting by its identifier to a value sent, parsed and validated, or with nil clears it so the level
// below shows through. A value that is not valid changes nothing.
func (s *Settings) Set(identifier string, value any) error {
	f, ok := byName[identifier]
	if !ok {
		return nil
	}
	if value == nil {
		f.clear(s)
		return nil
	}
	parsed, err := f.parse(value)
	if err != nil {
		return err
	}
	f.set(s, parsed)
	return nil
}

// Get is the value of a setting by its identifier, and whether it is set
func (s *Settings) Get(identifier string) (any, bool) {
	f, ok := byName[identifier]
	if !ok {
		return nil, false
	}
	return f.own(s)
}
