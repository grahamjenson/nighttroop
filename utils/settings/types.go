package settings

import (
	"cmp"
	"fmt"
	"slices"
	"strings"
	"time"

	"nighttroop.t1d.nz/data/types"
	"nighttroop.t1d.nz/utils/consts"
)

type SettingAnyType struct {
	Identifier string
	TypeString string
	Value      any
	CreatedAt  int64
	Subject    string
}

type Setting[T any] struct {
	definition *SettingDefinition[T]
	Value      T
	CreatedAt  int64
	Subject    string
}

func (s *Setting[T]) ToSettingAnyType() *SettingAnyType {
	if s == nil {
		return nil
	}
	return &SettingAnyType{
		Identifier: s.definition.Identifier,
		TypeString: s.definition.TypeString,
		Value:      s.Value,
		CreatedAt:  s.CreatedAt,
		Subject:    s.Subject,
	}
}

func NewSetting[T any](definition *SettingDefinition[T], value *T, createdAt *int64, subject *string) *Setting[T] {
	// Setting value to nil is just deleting the setting
	if value == nil {
		return nil
	}

	if createdAt == nil {
		now := time.Now().Unix()
		createdAt = &now
	}

	if subject == nil {
		subject = ptr("default")
	}

	setting := &Setting[T]{
		definition: definition,
		Value:      *value,
		CreatedAt:  *createdAt,
		Subject:    *subject,
	}

	return setting
}

func NewDefaultSetting[T any](definition *SettingDefinition[T]) *Setting[T] {

	setting := &Setting[T]{
		definition: definition,
		Value:      definition.Default,
		CreatedAt:  time.Now().Unix(),
		Subject:    "default",
	}

	return setting
}

func NewSettingAndParseValidate[T any](definition *SettingDefinition[T], value any, createdAt *int64, subject *string) (*Setting[T], error) {
	if value == nil {
		return nil, nil
	}
	parsedValue, err := definition.ParseValidate(value)
	if err != nil {
		return nil, err
	}
	return NewSetting(definition, parsedValue, createdAt, subject), nil
}

// //
// Parsers
// ////
func ParseUnits(value any) (consts.Units, error) {
	switch v := value.(type) {
	case *consts.Units:
		return *v, nil
	case consts.Units:
		return v, nil
	case string:
		u, err := consts.ParseUnits(v)
		if err != nil {
			return consts.UnitsMMOLL, fmt.Errorf("cannot parse units from string: %v", err)
		}
		return u, nil
	default:
		return consts.UnitsMMOLL, fmt.Errorf("unsupported type for ParseUnits: %T", value)
	}
}

func ParseInt(value any) (int64, error) {
	switch v := value.(type) {
	case int:
		return int64(v), nil
	case int64:
		return v, nil
	case int32:
		return int64(v), nil
	case float64:
		return int64(v), nil
	case float32:
		return int64(v), nil
	case string:
		var i int64
		_, err := fmt.Sscanf(v, "%d", &i)
		if err != nil {
			return 0, fmt.Errorf("cannot parse int from string: %v", err)
		}
		return i, nil
	default:
		return 0, fmt.Errorf("unsupported type for ParseInt: %T", value)
	}
}

func ParseFloat(value any) (float64, error) {
	switch v := value.(type) {
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case string:
		var f float64
		_, err := fmt.Sscanf(v, "%f", &f)
		if err != nil {
			return 0, fmt.Errorf("cannot parse float from string: %v", err)
		}
		return f, nil
	default:
		return 0, fmt.Errorf("unsupported type for ParseFloat: %T", value)
	}
}

func ParseBool(value any) (bool, error) {
	switch v := value.(type) {
	case bool:
		return v, nil
	case string:
		v = strings.ToLower(strings.TrimSpace(v))
		if v == "true" {
			return true, nil
		} else if v == "false" {
			return false, nil
		} else {
			return false, fmt.Errorf("cannot parse bool from string: %v", v)
		}
	default:
		return false, fmt.Errorf("unsupported type for ParseBool: %T", value)
	}
}

func ParseString(value any) (string, error) {
	switch v := value.(type) {
	case string:
		return v, nil
	default:
		return "", fmt.Errorf("unsupported type for ParseString: %T", value)
	}
}

func ParseLowerString(value any) (string, error) {
	switch v := value.(type) {
	case string:
		return strings.ToLower(v), nil
	default:
		return "", fmt.Errorf("unsupported type for ParseLowerString: %T", value)
	}
}

// ParseDeviceNames is a list of device names, as a JSON array is decoded
func ParseDeviceNames(value any) (types.DeviceNames, error) {
	list, ok := value.([]any)
	if !ok {
		return nil, fmt.Errorf("unsupported type for ParseDeviceNames: %T", value)
	}
	names := make(types.DeviceNames, len(list))
	for i, item := range list {
		name, ok := item.(string)
		if !ok {
			return nil, fmt.Errorf("device names must be strings, not %T", item)
		}
		names[i] = name
	}
	return names, nil
}

/////
// Validators
/////

// maxDeviceNames and maxDeviceName keep the list of device names within its column
const maxDeviceNames, maxDeviceName = 10, 64

// ValidDeviceNames is at most 10 different device names, each of 1 to 64 characters
func ValidDeviceNames(names types.DeviceNames) error {
	if len(names) > maxDeviceNames {
		return fmt.Errorf("must be at most %d devices", maxDeviceNames)
	}
	for i, name := range names {
		if name == "" || len(name) > maxDeviceName {
			return fmt.Errorf("each device must be 1 to %d characters", maxDeviceName)
		}
		if slices.Contains(names[:i], name) {
			return fmt.Errorf("%s is listed twice", name)
		}
	}
	return nil
}

func Range[T cmp.Ordered](min, max T) func(T) error {
	return func(value T) error {
		if value < min || value > max {
			return fmt.Errorf("must be between %v and %v", min, max)
		}
		return nil
	}
}

func quoted(values []string) string {
	result := make([]string, len(values))
	for i, value := range values {
		result[i] = fmt.Sprintf("'%v'", value)
	}
	return strings.Join(result, ", ")
}

//////
// INTERFACES
///////

type SettingsGetter interface {
	GetUnits() *consts.Units
	GetUnitsMgdlLabel() *string
	GetUnitsMmolLabel() *string
	GetTimezone() *string
	GetLanguage() *string
	GetDateFormat() *string
	GetTimeFormat() *string
	GetWeekStartDay() *string
	GetBolusDiaHours() *float64
	GetBolusOnsetMinutes() *float64
	GetBolusPeakMinutes() *float64
	GetBasalDiaHours() *float64
	GetBasalOnsetMinutes() *float64
	GetBasalPeakMinutes() *float64
	GetDefaultInsulinSensitivity() *float64
	GetDefaultCarbRatio() *float64
	GetDefaultBasalRate() *float64
	GetCarbMaxAbsorptionHours() *float64
	GetCarbMin5mImpact() *float64
	GetAlarmUrgentLowEnabled() *bool
	GetRangeUrgentLowMgdl() *int64
	GetAlarmLowEnabled() *bool
	GetRangeLowMgdl() *int64
	GetAlarmLowMinutes() *int64
	GetAlarmLowPredictionEnabled() *bool
	GetAlarmLowPredictionMinutes() *int64
	GetAlarmHighEnabled() *bool
	GetRangeHighMgdl() *int64
	GetAlarmHighMinutes() *int64
	GetAlarmUrgentHighEnabled() *bool
	GetRangeUrgentHighMgdl() *int64
	GetAlarmFastRiseEnabled() *bool
	GetAlarmFastRiseMgdl() *int64
	GetAlarmFastRiseMinutes() *int64
	GetAlarmFastDropEnabled() *bool
	GetAlarmFastDropMgdl() *int64
	GetAlarmFastDropMinutes() *int64
	GetAlarmNoDataEnabled() *bool
	GetAlarmNoDataMinutes() *int64
	GetAlarmDeviceBatteryLowEnabled() *bool
	GetAlarmDeviceBatteryLowPercent() *int64
	GetAlarmDeviceBatteryLowMinutes() *int64
	GetAlarmDeviceBatteryLowDevicesJson() *types.DeviceNames
	GetAlarmLowReservoirEnabled() *bool
	GetAlarmLowReservoirUnits() *int64
	GetNotificationsEnabled() *bool
}

func GetSettingByName(name string, in SettingsGetter) any {
	switch name {
	case "units":
		return in.GetUnits()
	case "units_mgdl_label":
		return in.GetUnitsMgdlLabel()
	case "units_mmol_label":
		return in.GetUnitsMmolLabel()
	case "timezone":
		return in.GetTimezone()
	case "language":
		return in.GetLanguage()
	case "date_format":
		return in.GetDateFormat()
	case "time_format":
		return in.GetTimeFormat()
	case "week_start_day":
		return in.GetWeekStartDay()
	case "bolus_dia_hours":
		return in.GetBolusDiaHours()
	case "bolus_onset_minutes":
		return in.GetBolusOnsetMinutes()
	case "bolus_peak_minutes":
		return in.GetBolusPeakMinutes()
	case "basal_dia_hours":
		return in.GetBasalDiaHours()
	case "basal_onset_minutes":
		return in.GetBasalOnsetMinutes()
	case "basal_peak_minutes":
		return in.GetBasalPeakMinutes()
	case "default_insulin_sensitivity":
		return in.GetDefaultInsulinSensitivity()
	case "default_carb_ratio":
		return in.GetDefaultCarbRatio()
	case "default_basal_rate":
		return in.GetDefaultBasalRate()
	case "carb_max_absorption_hours":
		return in.GetCarbMaxAbsorptionHours()
	case "carb_min_5m_impact":
		return in.GetCarbMin5mImpact()
	case "alarm_urgent_low_enabled":
		return in.GetAlarmUrgentLowEnabled()
	case "range_urgent_low_mgdl":
		return in.GetRangeUrgentLowMgdl()
	case "alarm_low_enabled":
		return in.GetAlarmLowEnabled()
	case "range_low_mgdl":
		return in.GetRangeLowMgdl()
	case "alarm_low_minutes":
		return in.GetAlarmLowMinutes()
	case "alarm_low_prediction_enabled":
		return in.GetAlarmLowPredictionEnabled()
	case "alarm_low_prediction_minutes":
		return in.GetAlarmLowPredictionMinutes()
	case "alarm_high_enabled":
		return in.GetAlarmHighEnabled()
	case "range_high_mgdl":
		return in.GetRangeHighMgdl()
	case "alarm_high_minutes":
		return in.GetAlarmHighMinutes()
	case "alarm_urgent_high_enabled":
		return in.GetAlarmUrgentHighEnabled()
	case "range_urgent_high_mgdl":
		return in.GetRangeUrgentHighMgdl()
	case "alarm_fast_rise_enabled":
		return in.GetAlarmFastRiseEnabled()
	case "alarm_fast_rise_mgdl":
		return in.GetAlarmFastRiseMgdl()
	case "alarm_fast_rise_minutes":
		return in.GetAlarmFastRiseMinutes()
	case "alarm_fast_drop_enabled":
		return in.GetAlarmFastDropEnabled()
	case "alarm_fast_drop_mgdl":
		return in.GetAlarmFastDropMgdl()
	case "alarm_fast_drop_minutes":
		return in.GetAlarmFastDropMinutes()
	case "alarm_no_data_enabled":
		return in.GetAlarmNoDataEnabled()
	case "alarm_no_data_minutes":
		return in.GetAlarmNoDataMinutes()
	case "alarm_device_battery_low_enabled":
		return in.GetAlarmDeviceBatteryLowEnabled()
	case "alarm_device_battery_low_percent":
		return in.GetAlarmDeviceBatteryLowPercent()
	case "alarm_device_battery_low_minutes":
		return in.GetAlarmDeviceBatteryLowMinutes()
	case "alarm_device_battery_low_devices":
		return in.GetAlarmDeviceBatteryLowDevicesJson()
	case "alarm_low_reservoir_enabled":
		return in.GetAlarmLowReservoirEnabled()
	case "alarm_low_reservoir_units":
		return in.GetAlarmLowReservoirUnits()
	case "notifications_enabled":
		return in.GetNotificationsEnabled()
	default:
		return nil
	}
}

type SettingsSetter interface {
	SetUnits(v *consts.Units)
	SetUnitsMgdlLabel(v *string)
	SetUnitsMmolLabel(v *string)
	SetTimezone(v *string)
	SetLanguage(v *string)
	SetDateFormat(v *string)
	SetTimeFormat(v *string)
	SetWeekStartDay(v *string)
	SetBolusDiaHours(v *float64)
	SetBolusOnsetMinutes(v *float64)
	SetBolusPeakMinutes(v *float64)
	SetBasalDiaHours(v *float64)
	SetBasalOnsetMinutes(v *float64)
	SetBasalPeakMinutes(v *float64)
	SetDefaultInsulinSensitivity(v *float64)
	SetDefaultCarbRatio(v *float64)
	SetDefaultBasalRate(v *float64)
	SetCarbMaxAbsorptionHours(v *float64)
	SetCarbMin5mImpact(v *float64)
	SetAlarmUrgentLowEnabled(v *bool)
	SetRangeUrgentLowMgdl(v *int64)
	SetAlarmLowEnabled(v *bool)
	SetRangeLowMgdl(v *int64)
	SetAlarmLowMinutes(v *int64)
	SetAlarmLowPredictionEnabled(v *bool)
	SetAlarmLowPredictionMinutes(v *int64)
	SetAlarmHighEnabled(v *bool)
	SetRangeHighMgdl(v *int64)
	SetAlarmHighMinutes(v *int64)
	SetAlarmUrgentHighEnabled(v *bool)
	SetRangeUrgentHighMgdl(v *int64)
	SetAlarmFastRiseEnabled(v *bool)
	SetAlarmFastRiseMgdl(v *int64)
	SetAlarmFastRiseMinutes(v *int64)
	SetAlarmFastDropEnabled(v *bool)
	SetAlarmFastDropMgdl(v *int64)
	SetAlarmFastDropMinutes(v *int64)
	SetAlarmNoDataEnabled(v *bool)
	SetAlarmNoDataMinutes(v *int64)
	SetAlarmDeviceBatteryLowEnabled(v *bool)
	SetAlarmDeviceBatteryLowPercent(v *int64)
	SetAlarmDeviceBatteryLowMinutes(v *int64)
	SetAlarmDeviceBatteryLowDevicesJson(v *types.DeviceNames)
	SetAlarmLowReservoirEnabled(v *bool)
	SetAlarmLowReservoirUnits(v *int64)
	SetNotificationsEnabled(v *bool)
}

func SetSettings(s SettingsSetter, g SettingsGetter) {
	s.SetUnits(g.GetUnits())
	s.SetUnitsMgdlLabel(g.GetUnitsMgdlLabel())
	s.SetUnitsMmolLabel(g.GetUnitsMmolLabel())
	s.SetTimezone(g.GetTimezone())
	s.SetLanguage(g.GetLanguage())
	s.SetDateFormat(g.GetDateFormat())
	s.SetTimeFormat(g.GetTimeFormat())
	s.SetWeekStartDay(g.GetWeekStartDay())
	s.SetBolusDiaHours(g.GetBolusDiaHours())
	s.SetBolusOnsetMinutes(g.GetBolusOnsetMinutes())
	s.SetBolusPeakMinutes(g.GetBolusPeakMinutes())
	s.SetBasalDiaHours(g.GetBasalDiaHours())
	s.SetBasalOnsetMinutes(g.GetBasalOnsetMinutes())
	s.SetBasalPeakMinutes(g.GetBasalPeakMinutes())
	s.SetDefaultInsulinSensitivity(g.GetDefaultInsulinSensitivity())
	s.SetDefaultCarbRatio(g.GetDefaultCarbRatio())
	s.SetDefaultBasalRate(g.GetDefaultBasalRate())
	s.SetCarbMaxAbsorptionHours(g.GetCarbMaxAbsorptionHours())
	s.SetCarbMin5mImpact(g.GetCarbMin5mImpact())
	s.SetAlarmUrgentLowEnabled(g.GetAlarmUrgentLowEnabled())
	s.SetRangeUrgentLowMgdl(g.GetRangeUrgentLowMgdl())
	s.SetAlarmLowEnabled(g.GetAlarmLowEnabled())
	s.SetRangeLowMgdl(g.GetRangeLowMgdl())
	s.SetAlarmLowMinutes(g.GetAlarmLowMinutes())
	s.SetAlarmLowPredictionEnabled(g.GetAlarmLowPredictionEnabled())
	s.SetAlarmLowPredictionMinutes(g.GetAlarmLowPredictionMinutes())
	s.SetAlarmHighEnabled(g.GetAlarmHighEnabled())
	s.SetRangeHighMgdl(g.GetRangeHighMgdl())
	s.SetAlarmHighMinutes(g.GetAlarmHighMinutes())
	s.SetAlarmUrgentHighEnabled(g.GetAlarmUrgentHighEnabled())
	s.SetRangeUrgentHighMgdl(g.GetRangeUrgentHighMgdl())
	s.SetAlarmFastRiseEnabled(g.GetAlarmFastRiseEnabled())
	s.SetAlarmFastRiseMgdl(g.GetAlarmFastRiseMgdl())
	s.SetAlarmFastRiseMinutes(g.GetAlarmFastRiseMinutes())
	s.SetAlarmFastDropEnabled(g.GetAlarmFastDropEnabled())
	s.SetAlarmFastDropMgdl(g.GetAlarmFastDropMgdl())
	s.SetAlarmFastDropMinutes(g.GetAlarmFastDropMinutes())
	s.SetAlarmNoDataEnabled(g.GetAlarmNoDataEnabled())
	s.SetAlarmNoDataMinutes(g.GetAlarmNoDataMinutes())
	s.SetAlarmDeviceBatteryLowEnabled(g.GetAlarmDeviceBatteryLowEnabled())
	s.SetAlarmDeviceBatteryLowPercent(g.GetAlarmDeviceBatteryLowPercent())
	s.SetAlarmDeviceBatteryLowMinutes(g.GetAlarmDeviceBatteryLowMinutes())
	s.SetAlarmDeviceBatteryLowDevicesJson(g.GetAlarmDeviceBatteryLowDevicesJson())
	s.SetAlarmLowReservoirEnabled(g.GetAlarmLowReservoirEnabled())
	s.SetAlarmLowReservoirUnits(g.GetAlarmLowReservoirUnits())
	s.SetNotificationsEnabled(g.GetNotificationsEnabled())
}
