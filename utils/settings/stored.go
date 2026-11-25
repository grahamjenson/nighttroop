package settings

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"nighttroop.t1d.nz/data/types"
	"nighttroop.t1d.nz/utils/consts"
)

// StoredVersion is the version of Stored that is written. It changes only when what a stored value means changes, with
// code here that reads the older versions: a new field does not change it.
const StoredVersion = 1

// Stored is a level's settings, the site's own or an identity's own, as its settings_json column holds them. A
// setting that is not set is nil and left out, so the level below it, the site's or the default, shows through.
//
// Rows of every release are read into this struct, so it only grows: a field is added, and never renamed, removed,
// moved or given another type or JSON name. TestStoredOnlyGrows fails when one is. AGENTS.md has the rule.
type Stored struct {
	Version int `json:"version"`

	// Units is mg/dL or mmol/L, as text, so one that is neither is left out when read and the rest of the row kept
	Units        *string `json:"units,omitempty"`
	Timezone     *string `json:"timezone,omitempty"`
	Language     *string `json:"language,omitempty"`
	DateFormat   *string `json:"date_format,omitempty"`
	TimeFormat   *string `json:"time_format,omitempty"`
	WeekStartDay *string `json:"week_start_day,omitempty"`
	NumberFormat *string `json:"number_format,omitempty"`

	Bolus         StoredInsulin       `json:"bolus,omitzero"`
	Basal         StoredInsulin       `json:"basal,omitzero"`
	Defaults      StoredDefaults      `json:"defaults,omitzero"`
	Carbs         StoredCarbs         `json:"carbs,omitzero"`
	Range         StoredRange         `json:"range,omitzero"`
	Alarms        StoredAlarms        `json:"alarms,omitzero"`
	Notifications StoredNotifications `json:"notifications,omitzero"`
}

// StoredInsulin is how an insulin acts: how long, and when it starts and peaks
type StoredInsulin struct {
	DiaHours     *float64 `json:"dia_hours,omitempty"`
	OnsetMinutes *float64 `json:"onset_minutes,omitempty"`
	PeakMinutes  *float64 `json:"peak_minutes,omitempty"`
}

// StoredDefaults are the profile values used when a site has no profile
type StoredDefaults struct {
	InsulinSensitivity *float64 `json:"insulin_sensitivity,omitempty"`
	CarbRatio          *float64 `json:"carb_ratio,omitempty"`
	BasalRate          *float64 `json:"basal_rate,omitempty"`
}

// StoredCarbs is how carbs are absorbed
type StoredCarbs struct {
	MaxAbsorptionHours *float64 `json:"max_absorption_hours,omitempty"`
	Min5mImpact        *float64 `json:"min_5m_impact,omitempty"`
}

// StoredRange is the glucose range in mg/dL, which the graphs, the summaries and the glucose alarms share
type StoredRange struct {
	UrgentLowMgdl  *int64 `json:"urgent_low_mgdl,omitempty"`
	LowMgdl        *int64 `json:"low_mgdl,omitempty"`
	HighMgdl       *int64 `json:"high_mgdl,omitempty"`
	UrgentHighMgdl *int64 `json:"urgent_high_mgdl,omitempty"`
}

// StoredAlarms is each alarm: whether it is on, and what it waits for
type StoredAlarms struct {
	UrgentLow        StoredAlarm          `json:"urgent_low,omitzero"`
	Low              StoredTimedAlarm     `json:"low,omitzero"`
	LowPrediction    StoredTimedAlarm     `json:"low_prediction,omitzero"`
	High             StoredTimedAlarm     `json:"high,omitzero"`
	UrgentHigh       StoredAlarm          `json:"urgent_high,omitzero"`
	FastRise         StoredChangeAlarm    `json:"fast_rise,omitzero"`
	FastDrop         StoredChangeAlarm    `json:"fast_drop,omitzero"`
	NoData           StoredTimedAlarm     `json:"no_data,omitzero"`
	DeviceBatteryLow StoredBatteryAlarm   `json:"device_battery_low,omitzero"`
	LowReservoir     StoredReservoirAlarm `json:"low_reservoir,omitzero"`
}

// StoredAlarm is an alarm that is on or off
type StoredAlarm struct {
	Enabled *bool `json:"enabled,omitempty"`
}

// StoredTimedAlarm is an alarm that waits for, or looks ahead, some minutes
type StoredTimedAlarm struct {
	Enabled *bool  `json:"enabled,omitempty"`
	Minutes *int64 `json:"minutes,omitempty"`
}

// StoredChangeAlarm is an alarm for glucose changing by so many mg/dL over some minutes
type StoredChangeAlarm struct {
	Enabled *bool  `json:"enabled,omitempty"`
	Mgdl    *int64 `json:"mgdl,omitempty"`
	Minutes *int64 `json:"minutes,omitempty"`
}

// StoredBatteryAlarm is the alarm for the listed devices' batteries, and for one going quiet
type StoredBatteryAlarm struct {
	Enabled *bool              `json:"enabled,omitempty"`
	Percent *int64             `json:"percent,omitempty"`
	Minutes *int64             `json:"minutes,omitempty"`
	Devices *types.DeviceNames `json:"devices,omitempty"`
}

// StoredReservoirAlarm is the alarm for a pump's reservoir running low
type StoredReservoirAlarm struct {
	Enabled *bool  `json:"enabled,omitempty"`
	Units   *int64 `json:"units,omitempty"`
}

// StoredNotifications is whether the identity is sent notifications
type StoredNotifications struct {
	Enabled *bool `json:"enabled,omitempty"`
}

// storedField is one setting's place in a Stored
type storedField struct {
	identifier string
	// at is the field itself, a pointer to its pointer
	at any
	// get is the stored value, or nil when it is not set, and set stores a value of the setting's own type, or nil
	get func() any
	set func(any)
}

func fieldOf[T any](identifier string, at **T) storedField {
	return storedField{identifier: identifier, at: at,
		get: func() any {
			if *at == nil {
				return nil
			}
			return **at
		},
		set: func(value any) {
			*at = nil
			if typed, ok := value.(T); ok {
				*at = &typed
			}
		}}
}

// unitsField is the units, which are stored as text
func unitsField(at **string) storedField {
	field := fieldOf(Units.Identifier, at)
	text := field.set
	field.set = func(value any) {
		if units, ok := value.(consts.Units); ok {
			value = string(units)
		}
		text(value)
	}
	return field
}

// fields are where s keeps each setting, by the setting's identifier. A setting is in exactly one place
// (TestEverySettingIsStored).
func (s *Stored) fields() []storedField {
	return []storedField{
		unitsField(&s.Units),
		fieldOf(Timezone.Identifier, &s.Timezone),
		fieldOf(Language.Identifier, &s.Language),
		fieldOf(DateFormat.Identifier, &s.DateFormat),
		fieldOf(TimeFormat.Identifier, &s.TimeFormat),
		fieldOf(WeekStartDay.Identifier, &s.WeekStartDay),
		fieldOf(NumberFormat.Identifier, &s.NumberFormat),
		fieldOf(BolusDIAHours.Identifier, &s.Bolus.DiaHours),
		fieldOf(BolusOnsetMinutes.Identifier, &s.Bolus.OnsetMinutes),
		fieldOf(BolusPeakMinutes.Identifier, &s.Bolus.PeakMinutes),
		fieldOf(BasalDIAHours.Identifier, &s.Basal.DiaHours),
		fieldOf(BasalOnsetMinutes.Identifier, &s.Basal.OnsetMinutes),
		fieldOf(BasalPeakMinutes.Identifier, &s.Basal.PeakMinutes),
		fieldOf(DefaultInsulinSensitivity.Identifier, &s.Defaults.InsulinSensitivity),
		fieldOf(DefaultCarbRatio.Identifier, &s.Defaults.CarbRatio),
		fieldOf(DefaultBasalRate.Identifier, &s.Defaults.BasalRate),
		fieldOf(CarbMaxAbsorptionHours.Identifier, &s.Carbs.MaxAbsorptionHours),
		fieldOf(CarbMin5mImpact.Identifier, &s.Carbs.Min5mImpact),
		fieldOf(RangeUrgentLowMgdl.Identifier, &s.Range.UrgentLowMgdl),
		fieldOf(RangeLowMgdl.Identifier, &s.Range.LowMgdl),
		fieldOf(RangeHighMgdl.Identifier, &s.Range.HighMgdl),
		fieldOf(RangeUrgentHighMgdl.Identifier, &s.Range.UrgentHighMgdl),
		fieldOf(AlarmUrgentLowEnabled.Identifier, &s.Alarms.UrgentLow.Enabled),
		fieldOf(AlarmLowEnabled.Identifier, &s.Alarms.Low.Enabled),
		fieldOf(AlarmLowMinutes.Identifier, &s.Alarms.Low.Minutes),
		fieldOf(AlarmLowPredictionEnabled.Identifier, &s.Alarms.LowPrediction.Enabled),
		fieldOf(AlarmLowPredictionMinutes.Identifier, &s.Alarms.LowPrediction.Minutes),
		fieldOf(AlarmHighEnabled.Identifier, &s.Alarms.High.Enabled),
		fieldOf(AlarmHighMinutes.Identifier, &s.Alarms.High.Minutes),
		fieldOf(AlarmUrgentHighEnabled.Identifier, &s.Alarms.UrgentHigh.Enabled),
		fieldOf(AlarmFastRiseEnabled.Identifier, &s.Alarms.FastRise.Enabled),
		fieldOf(AlarmFastRiseMgdl.Identifier, &s.Alarms.FastRise.Mgdl),
		fieldOf(AlarmFastRiseMinutes.Identifier, &s.Alarms.FastRise.Minutes),
		fieldOf(AlarmFastDropEnabled.Identifier, &s.Alarms.FastDrop.Enabled),
		fieldOf(AlarmFastDropMgdl.Identifier, &s.Alarms.FastDrop.Mgdl),
		fieldOf(AlarmFastDropMinutes.Identifier, &s.Alarms.FastDrop.Minutes),
		fieldOf(AlarmNoDataEnabled.Identifier, &s.Alarms.NoData.Enabled),
		fieldOf(AlarmNoDataMinutes.Identifier, &s.Alarms.NoData.Minutes),
		fieldOf(AlarmDeviceBatteryLowEnabled.Identifier, &s.Alarms.DeviceBatteryLow.Enabled),
		fieldOf(AlarmDeviceBatteryLowPercent.Identifier, &s.Alarms.DeviceBatteryLow.Percent),
		fieldOf(AlarmDeviceBatteryLowMinutes.Identifier, &s.Alarms.DeviceBatteryLow.Minutes),
		fieldOf(AlarmDeviceBatteryLowDevicesJson.Identifier, &s.Alarms.DeviceBatteryLow.Devices),
		fieldOf(AlarmLowReservoirEnabled.Identifier, &s.Alarms.LowReservoir.Enabled),
		fieldOf(AlarmLowReservoirUnits.Identifier, &s.Alarms.LowReservoir.Units),
		fieldOf(NotificationsEnabled.Identifier, &s.Notifications.Enabled),
	}
}

// MaxJSONBytes is the most a level's settings take as stored, as their column allows. Every setting set, with the
// longest values each allows (10 device names of 64 characters), is under half of it.
const MaxJSONBytes = 4096

// stored is m as it is stored: each setting it has set, at StoredVersion
func (m *SettingMap) stored() *Stored {
	s := &Stored{Version: StoredVersion}
	for _, field := range s.fields() {
		if setting := m.GetSettingAnyTypeByName(field.identifier); setting != nil {
			field.set(setting.Value)
		}
	}
	return s
}

// JSON is the settings m has set, as their column holds them
func (m *SettingMap) JSON() ([]byte, error) {
	raw, err := json.Marshal(m.stored())
	if err == nil && len(raw) > MaxJSONBytes {
		err = fmt.Errorf("settings are %d bytes, over %d", len(raw), MaxJSONBytes)
	}
	return raw, err
}

// Stored is JSON as a column takes it, for a save's parameters
func (m *SettingMap) Stored() (string, error) {
	raw, err := m.JSON()
	return string(raw), err
}

// FromJSON is the settings a column holds. Each value is parsed and validated as one sent to the API is, so a value
// out of its range, or of the wrong type, is left out and the rest read: a row is never trusted because it was stored.
// A row that is too long, or is not JSON, has no settings. The logs say so.
func FromJSON(raw []byte, createdAt *int64, subject *string) *SettingMap {
	m, s := &SettingMap{}, &Stored{}
	if len(raw) > MaxJSONBytes {
		slog.Error("stored settings are too long, so none are used", "bytes", len(raw))
		return m
	}
	// A field of the wrong type is skipped and the rest read; anything else is not a Stored
	var wrongType *json.UnmarshalTypeError
	if err := json.Unmarshal(raw, s); errors.As(err, &wrongType) {
		slog.Error("a stored setting is not of its type, so it is not used", "field", wrongType.Field)
	} else if err != nil {
		slog.Error("stored settings cannot be read, so none are used", "error", err)
		return m
	}
	for _, field := range s.fields() {
		if value := field.get(); value != nil {
			// A list is parsed from what JSON decodes it to, as one sent to the API is
			if names, ok := value.(types.DeviceNames); ok {
				list := make([]any, len(names))
				for i, name := range names {
					list[i] = name
				}
				value = list
			}
			_ = m.SetSettingByName(field.identifier, value, createdAt, subject)
		}
	}
	return m
}
