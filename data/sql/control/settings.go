package control

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"nighttroop.t1d.nz/utils/consts"
)

var ALL_SETTINGS = []string{
	"units",
	"timezone",
	"language",
	"date_format",
	"time_format",
	"week_start_day",
	"bolus_dia_hours",
	"bolus_onset_minutes",
	"bolus_peak_minutes",
	"basal_dia_hours",
	"basal_onset_minutes",
	"basal_peak_minutes",
	"default_insulin_sensitivity",
	"default_carb_ratio",
	"default_basal_rate",
	"carb_max_absorption_hours",
	"carb_min_5m_impact",
	"alarm_urgent_low_enabled",
	"alarm_urgent_low",
	"alarm_low_enabled",
	"alarm_low",
	"alarm_low_minutes",
	"alarm_low_prediction_enabled",
	"alarm_high_enabled",
	"alarm_high",
	"alarm_high_minutes",
	"alarm_urgent_high_enabled",
	"alarm_urgent_high",
	"alarm_fast_rise_enabled",
	"alarm_fast_rise",
	"alarm_fast_rise_minutes",
	"alarm_fast_drop_enabled",
	"alarm_fast_drop",
	"alarm_fast_drop_minutes",
	"alarm_no_data_enabled",
	"alarm_no_data_minutes",
	"alarm_battery_low_enabled",
	"alarm_battery_low_percent",
	"alarm_battery_low_pwd_enabled",
	"alarm_battery_low_pwd_percent",
	"alarm_low_reservoir_enabled",
	"alarm_low_reservoir_units",
	"notifications_enabled",
}

var ALL_SETTINGS_DEFAULTS = map[string]any{
	"units":                         "mmol/L",
	"timezone":                      "UTC",
	"language":                      "en-NZ",
	"date_format":                   "YYYY-MM-DD",
	"time_format":                   "HH:mm:ss",
	"week_start_day":                "monday",
	"bolus_dia_hours":               6.0,
	"bolus_onset_minutes":           20.0,
	"bolus_peak_minutes":            60.0,
	"basal_dia_hours":               24.0,
	"basal_onset_minutes":           60.0,
	"basal_peak_minutes":            120.0,
	"default_insulin_sensitivity":   10.0,
	"default_carb_ratio":            10.0,
	"default_basal_rate":            1.0,
	"carb_max_absorption_hours":     6.0,
	"carb_min_5m_impact":            10.0,
	"alarm_urgent_low_enabled":      false,
	"alarm_urgent_low":              56,
	"alarm_low_enabled":             false,
	"alarm_low":                     72,
	"alarm_low_minutes":             5,
	"alarm_low_prediction_enabled":  false,
	"alarm_high_enabled":            false,
	"alarm_high":                    180,
	"alarm_high_minutes":            30,
	"alarm_urgent_high_enabled":     false,
	"alarm_urgent_high":             270,
	"alarm_fast_rise_enabled":       false,
	"alarm_fast_rise":               18,
	"alarm_fast_rise_minutes":       15,
	"alarm_fast_drop_enabled":       false,
	"alarm_fast_drop":               18,
	"alarm_fast_drop_minutes":       15,
	"alarm_no_data_enabled":         false,
	"alarm_no_data_minutes":         20,
	"alarm_battery_low_enabled":     false,
	"alarm_battery_low_percent":     20,
	"alarm_battery_low_pwd_enabled": false,
	"alarm_battery_low_pwd_percent": 20,
	"alarm_low_reservoir_enabled":   false,
	"alarm_low_reservoir_units":     25,
	"notifications_enabled":         false,
}

type SettingUnitsType struct{}

func (s SettingUnitsType) FromJSONRaw(raw json.RawMessage) (any, error) {
	var value string
	if err := json.Unmarshal(raw, &value); err != nil {
		return "", err
	}
	unitsVal, err := consts.ParseUnits(value)
	if err != nil {
		return "", err
	}

	return string(unitsVal), nil
}

func (u SettingUnitsType) MarshalJSON() ([]byte, error) {
	return json.Marshal(u.GetTypeString())
}

func (s SettingUnitsType) GetTypeString() string {
	return "string"
}

type SettingShortStringType struct{}

func (s SettingShortStringType) FromJSONRaw(raw json.RawMessage) (any, error) {
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		return nil, fmt.Errorf("expected notes string, got %s", raw)
	}
	var out string
	switch v := v.(type) {
	case string:
		out = v
	case float64:
		out = strconv.FormatFloat(v, 'f', -1, 64)
	case bool:
		out = strconv.FormatBool(v)
	default:
		out = fmt.Sprintf("%v", v)
	}

	if len(out) > 64 {
		out = out[:64]
	}

	return out, nil
}

func (u SettingShortStringType) MarshalJSON() ([]byte, error) {
	return json.Marshal(u.GetTypeString())
}

func (s SettingShortStringType) GetTypeString() string {
	return "string"
}

type SettingFlexibleFloat64Type struct{}

func (f SettingFlexibleFloat64Type) FromJSONRaw(raw json.RawMessage) (any, error) {
	var v any
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	if err := dec.Decode(&v); err != nil {
		return nil, fmt.Errorf("expected float value, got %s", raw)
	}

	var out float64
	switch val := v.(type) {
	case float64:
		out = val
	case json.Number:
		if asFloat, err := val.Float64(); err == nil {
			out = asFloat
			return out, nil
		}
		return nil, fmt.Errorf("invalid float value %q", val)
	case string:
		var err error
		newF, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid float value %q", val)
		}
		out = newF
	default:
		return nil, fmt.Errorf("expected number or string for float, got %v:%T", v, v)
	}
	return out, nil
}

func (u SettingFlexibleFloat64Type) MarshalJSON() ([]byte, error) {
	return json.Marshal(u.GetTypeString())
}

func (s SettingFlexibleFloat64Type) GetTypeString() string {
	return "number"
}

type SettingFlexibleBoolType struct{}

func (b SettingFlexibleBoolType) FromJSONRaw(raw json.RawMessage) (any, error) {
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		return nil, fmt.Errorf("expected bool value, got %s", raw)
	}

	var out bool
	switch val := v.(type) {
	case bool:
		out = val
	case string:
		lower := strings.ToLower(strings.TrimSpace(val))
		if lower == "true" || lower == "1" {
			out = true
		} else if lower == "false" || lower == "0" {
			out = false
		} else {
			return nil, fmt.Errorf("invalid bool value %q", val)
		}
	default:
		return nil, fmt.Errorf("expected bool or string for bool, got %T", v)
	}
	return out, nil
}

func (u SettingFlexibleBoolType) MarshalJSON() ([]byte, error) {
	return json.Marshal(u.GetTypeString())
}

func (s SettingFlexibleBoolType) GetTypeString() string {
	return "boolean"
}

type SettingsType interface {
	FromJSONRaw(raw json.RawMessage) (any, error)
	GetTypeString() string
}

var ALL_SETTINGS_TYPES = map[string]SettingsType{
	"units":                         SettingUnitsType{},
	"timezone":                      SettingShortStringType{},
	"language":                      SettingShortStringType{},
	"date_format":                   SettingShortStringType{},
	"time_format":                   SettingShortStringType{},
	"week_start_day":                SettingShortStringType{},
	"bolus_dia_hours":               SettingFlexibleFloat64Type{},
	"bolus_onset_minutes":           SettingFlexibleFloat64Type{},
	"bolus_peak_minutes":            SettingFlexibleFloat64Type{},
	"basal_dia_hours":               SettingFlexibleFloat64Type{},
	"basal_onset_minutes":           SettingFlexibleFloat64Type{},
	"basal_peak_minutes":            SettingFlexibleFloat64Type{},
	"default_insulin_sensitivity":   SettingFlexibleFloat64Type{},
	"default_carb_ratio":            SettingFlexibleFloat64Type{},
	"default_basal_rate":            SettingFlexibleFloat64Type{},
	"carb_max_absorption_hours":     SettingFlexibleFloat64Type{},
	"carb_min_5m_impact":            SettingFlexibleFloat64Type{},
	"alarm_urgent_low_enabled":      SettingFlexibleBoolType{},
	"alarm_urgent_low":              SettingFlexibleFloat64Type{},
	"alarm_low_enabled":             SettingFlexibleBoolType{},
	"alarm_low":                     SettingFlexibleFloat64Type{},
	"alarm_low_minutes":             SettingFlexibleFloat64Type{},
	"alarm_low_prediction_enabled":  SettingFlexibleBoolType{},
	"alarm_high_enabled":            SettingFlexibleBoolType{},
	"alarm_high":                    SettingFlexibleFloat64Type{},
	"alarm_high_minutes":            SettingFlexibleFloat64Type{},
	"alarm_urgent_high_enabled":     SettingFlexibleBoolType{},
	"alarm_urgent_high":             SettingFlexibleFloat64Type{},
	"alarm_fast_rise_enabled":       SettingFlexibleBoolType{},
	"alarm_fast_rise":               SettingFlexibleFloat64Type{},
	"alarm_fast_rise_minutes":       SettingFlexibleFloat64Type{},
	"alarm_fast_drop_enabled":       SettingFlexibleBoolType{},
	"alarm_fast_drop":               SettingFlexibleFloat64Type{},
	"alarm_fast_drop_minutes":       SettingFlexibleFloat64Type{},
	"alarm_no_data_enabled":         SettingFlexibleBoolType{},
	"alarm_no_data_minutes":         SettingFlexibleFloat64Type{},
	"alarm_battery_low_enabled":     SettingFlexibleBoolType{},
	"alarm_battery_low_percent":     SettingFlexibleFloat64Type{},
	"alarm_battery_low_pwd_enabled": SettingFlexibleBoolType{},
	"alarm_battery_low_pwd_percent": SettingFlexibleFloat64Type{},
	"alarm_low_reservoir_enabled":   SettingFlexibleBoolType{},
	"alarm_low_reservoir_units":     SettingFlexibleFloat64Type{},
	"notifications_enabled":         SettingFlexibleBoolType{},
}

func (i *SiteIdentity) GetSetting(name string) any {
	switch name {
	case "units":
		return i.Units
	case "timezone":
		return i.Timezone
	case "language":
		return i.Language
	case "date_format":
		return i.DateFormat
	case "time_format":
		return i.TimeFormat
	case "week_start_day":
		return i.WeekStartDay
	case "bolus_dia_hours":
		return i.BolusDiaHours
	case "bolus_onset_minutes":
		return i.BolusOnsetMinutes
	case "bolus_peak_minutes":
		return i.BolusPeakMinutes
	case "basal_dia_hours":
		return i.BasalDiaHours
	case "basal_onset_minutes":
		return i.BasalOnsetMinutes
	case "basal_peak_minutes":
		return i.BasalPeakMinutes
	case "default_insulin_sensitivity":
		return i.DefaultInsulinSensitivity
	case "default_carb_ratio":
		return i.DefaultCarbRatio
	case "default_basal_rate":
		return i.DefaultBasalRate
	case "carb_max_absorption_hours":
		return i.CarbMaxAbsorptionHours
	case "carb_min_5m_impact":
		return i.CarbMin5mImpact
	case "alarm_urgent_low_enabled":
		return i.AlarmUrgentLowEnabled
	case "alarm_urgent_low":
		return i.AlarmUrgentLow
	case "alarm_low_enabled":
		return i.AlarmLowEnabled
	case "alarm_low":
		return i.AlarmLow
	case "alarm_low_minutes":
		return i.AlarmLowMinutes
	case "alarm_low_prediction_enabled":
		return i.AlarmLowPredictionEnabled
	case "alarm_high_enabled":
		return i.AlarmHighEnabled
	case "alarm_high":
		return i.AlarmHigh
	case "alarm_high_minutes":
		return i.AlarmHighMinutes
	case "alarm_urgent_high_enabled":
		return i.AlarmUrgentHighEnabled
	case "alarm_urgent_high":
		return i.AlarmUrgentHigh
	case "alarm_fast_rise_enabled":
		return i.AlarmFastRiseEnabled
	case "alarm_fast_rise":
		return i.AlarmFastRise
	case "alarm_fast_rise_minutes":
		return i.AlarmFastRiseMinutes
	case "alarm_fast_drop_enabled":
		return i.AlarmFastDropEnabled
	case "alarm_fast_drop":
		return i.AlarmFastDrop
	case "alarm_fast_drop_minutes":
		return i.AlarmFastDropMinutes
	case "alarm_no_data_enabled":
		return i.AlarmNoDataEnabled
	case "alarm_no_data_minutes":
		return i.AlarmNoDataMinutes
	case "alarm_battery_low_enabled":
		return i.AlarmBatteryLowEnabled
	case "alarm_battery_low_percent":
		return i.AlarmBatteryLowPercent
	case "alarm_battery_low_pwd_enabled":
		return i.AlarmBatteryLowPwdEnabled
	case "alarm_battery_low_pwd_percent":
		return i.AlarmBatteryLowPwdPercent
	case "alarm_low_reservoir_enabled":
		return i.AlarmLowReservoirEnabled
	case "alarm_low_reservoir_units":
		return i.AlarmLowReservoirUnits
	case "notifications_enabled":
		return i.NotificationsEnabled
	default:
		return nil
	}
}

func (i *Site) GetSetting(name string) any {
	switch name {
	case "units":
		return i.Units
	case "timezone":
		return i.Timezone
	case "language":
		return i.Language
	case "date_format":
		return i.DateFormat
	case "time_format":
		return i.TimeFormat
	case "week_start_day":
		return i.WeekStartDay
	case "bolus_dia_hours":
		return i.BolusDiaHours
	case "bolus_onset_minutes":
		return i.BolusOnsetMinutes
	case "bolus_peak_minutes":
		return i.BolusPeakMinutes
	case "basal_dia_hours":
		return i.BasalDiaHours
	case "basal_onset_minutes":
		return i.BasalOnsetMinutes
	case "basal_peak_minutes":
		return i.BasalPeakMinutes
	case "default_insulin_sensitivity":
		return i.DefaultInsulinSensitivity
	case "default_carb_ratio":
		return i.DefaultCarbRatio
	case "default_basal_rate":
		return i.DefaultBasalRate
	case "carb_max_absorption_hours":
		return i.CarbMaxAbsorptionHours
	case "carb_min_5m_impact":
		return i.CarbMin5mImpact
	default:
		return nil
	}
}
