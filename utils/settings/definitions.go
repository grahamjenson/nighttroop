package settings

import (
	"fmt"
	"slices"
	"time"

	"golang.org/x/text/language"
	"nighttroop.t1d.nz/data/types"
	"nighttroop.t1d.nz/utils/consts"
)

type SettingDefinition[T any] struct {
	Identifier   string
	Default      T
	TypeString   string
	ValidateFunc func(T) error
	ParseFunc    func(any) (T, error)
}

func (s *SettingDefinition[T]) GetIdentifier() string {
	return s.Identifier
}

func (s *SettingDefinition[T]) Parse(value any) (any, error) {
	return s.ParseFunc(value)
}

func (s *SettingDefinition[T]) ParseValidate(value any) (*T, error) {
	v, err := s.ParseFunc(value)
	if err != nil {
		return nil, err
	}
	if s.ValidateFunc == nil {
		return &v, nil
	}
	err = s.ValidateFunc(v)
	return &v, err
}

// numberSetting is a setting of a number, def by default, that must be from min to max
func numberSetting[T int64 | float64](identifier string, def, min, max T, parse func(any) (T, error)) *SettingDefinition[T] {
	return &SettingDefinition[T]{Identifier: identifier, TypeString: "number", Default: def, ValidateFunc: Range(min, max), ParseFunc: parse}
}

// onOffSetting is a setting that is on or off, and off by default
func onOffSetting(identifier string) *SettingDefinition[bool] {
	return &SettingDefinition[bool]{Identifier: identifier, TypeString: "boolean", Default: false, ParseFunc: ParseBool}
}

// The values the date, time and week start settings accept, in the order the settings pages offer them
var (
	DateFormats   = []string{"YYYY/MM/DD", "DD/MM/YYYY", "MM/DD/YYYY", "YYYY-MM-DD", "DD-MM-YYYY", "MM-DD-YYYY"}
	TimeFormats   = []string{"HH:mm:ss", "hh:mm:ss a", "HH:mm", "hh:mm a"}
	WeekStartDays = []string{"monday", "sunday"}
	// NumberFormats write a thousand and a half with each thousands separator and decimal mark
	NumberFormats = []string{"1,234.5", "1.234,5", "1 234,5"}
)

var Units = &SettingDefinition[consts.Units]{
	Identifier: "units",
	TypeString: "string",
	Default:    consts.UnitsMMOLL,
	ParseFunc:  ParseUnits,
	ValidateFunc: func(value consts.Units) error {
		if value != consts.UnitsMMOLL && value != consts.UnitsMGDL {
			return fmt.Errorf("invalid units: %v - must be either 'mmol/L' or 'mg/dL'", value)
		}
		return nil
	},
}

var Timezone = &SettingDefinition[string]{
	Identifier: "timezone",
	TypeString: "string",
	Default:    "UTC",
	ParseFunc:  ParseString,
	ValidateFunc: func(value string) error {
		if value == "" {
			return fmt.Errorf("timezone cannot be empty")
		}
		_, err := time.LoadLocation(value)
		if err != nil {
			return fmt.Errorf("invalid timezone: %v - must be a valid IANA timezone", value)
		}
		return nil
	},
}

var Language = &SettingDefinition[string]{
	Identifier: "language",
	TypeString: "string",
	Default:    "en-NZ",
	ParseFunc:  ParseString,
	ValidateFunc: func(value string) error {
		if value == "" {
			return fmt.Errorf("language cannot be empty")
		}
		_, err := language.Parse(value)
		if err != nil {
			return fmt.Errorf("invalid language: %v - must be a valid BCP 47 language tag", value)
		}
		return nil
	},
}

// oneOf validates a setting that must be one of options
func oneOf(name string, options []string) func(string) error {
	return func(value string) error {
		if !slices.Contains(options, value) {
			return fmt.Errorf("invalid %s: %v - must be in %v", name, value, quoted(options))
		}
		return nil
	}
}

var DateFormat = &SettingDefinition[string]{
	Identifier: "date_format",
	TypeString: "string",
	Default:    "YYYY-MM-DD",
	ParseFunc:  ParseString,
	ValidateFunc: func(value string) error {
		if value == "" {
			return fmt.Errorf("date format cannot be empty")
		}

		if !slices.Contains(DateFormats, value) {
			return fmt.Errorf("invalid date format: %v -must be in %v", value, quoted(DateFormats))
		}
		return nil
	},
}

var TimeFormat = &SettingDefinition[string]{
	Identifier: "time_format",
	TypeString: "string",
	Default:    "HH:mm:ss",
	ParseFunc:  ParseString,
	ValidateFunc: func(value string) error {
		if value == "" {
			return fmt.Errorf("time format cannot be empty")
		}
		if !slices.Contains(TimeFormats, value) {
			return fmt.Errorf("invalid time format: %v - must be in %v", value, quoted(TimeFormats))
		}
		return nil
	},
}

var WeekStartDay = &SettingDefinition[string]{
	Identifier: "week_start_day",
	TypeString: "string",
	Default:    "monday",
	ParseFunc:  ParseLowerString,
	ValidateFunc: func(value string) error {
		if value == "" {
			return fmt.Errorf("week start day cannot be empty")
		}

		if !slices.Contains(WeekStartDays, value) {
			return fmt.Errorf("invalid week start day: %v - must be one of %v", value, quoted(WeekStartDays))
		}
		return nil
	},
}

var NumberFormat = &SettingDefinition[string]{
	Identifier:   "number_format",
	TypeString:   "string",
	Default:      NumberFormats[0],
	ParseFunc:    ParseString,
	ValidateFunc: oneOf("number format", NumberFormats),
}

var BolusDIAHours = numberSetting("bolus_dia_hours", 6.0, 1.0, 10.0, ParseFloat)

var BolusOnsetMinutes = numberSetting("bolus_onset_minutes", 20.0, 1.0, 120.0, ParseFloat)

var BolusPeakMinutes = numberSetting("bolus_peak_minutes", 60.0, 1.0, 240.0, ParseFloat)

var BasalDIAHours = numberSetting("basal_dia_hours", 24.0, 1.0, 170.0, ParseFloat)

var BasalOnsetMinutes = numberSetting("basal_onset_minutes", 60.0, 1.0, 360.0, ParseFloat)

var BasalPeakMinutes = numberSetting("basal_peak_minutes", 120.0, 1.0, 240.0, ParseFloat)

var DefaultInsulinSensitivity = numberSetting("default_insulin_sensitivity", 10.0, 1.0, 100.0, ParseFloat)

var DefaultCarbRatio = numberSetting("default_carb_ratio", 10.0, 1.0, 100.0, ParseFloat)

var DefaultBasalRate = numberSetting("default_basal_rate", 1.0, 0.1, 10.0, ParseFloat)

var CarbMaxAbsorptionHours = numberSetting("carb_max_absorption_hours", 6.0, 1.0, 24.0, ParseFloat)

var CarbMin5mImpact = numberSetting("carb_min_5m_impact", 10.0, 1.0, 100.0, ParseFloat)

var AlarmUrgentLowEnabled = onOffSetting("alarm_urgent_low_enabled")

var RangeUrgentLowMgdl = numberSetting("range_urgent_low_mgdl", 56, 20, 100, ParseInt)

var AlarmLowEnabled = onOffSetting("alarm_low_enabled")

// MaxAlarmMinutes is the most minutes an alarm can wait for, so a check of the alarms reads only the readings it needs
const MaxAlarmMinutes = 120

var RangeLowMgdl = numberSetting("range_low_mgdl", 72, 20, 150, ParseInt)

var AlarmLowMinutes = numberSetting("alarm_low_minutes", 5, 0, MaxAlarmMinutes, ParseInt)

var AlarmLowPredictionEnabled = onOffSetting("alarm_low_prediction_enabled")

// AlarmLowPredictionMinutes is how far ahead the predicted low alarm follows the current trend to the low alarm's glucose
var AlarmLowPredictionMinutes = numberSetting("alarm_low_prediction_minutes", 20, 5, 60, ParseInt)

var AlarmHighEnabled = onOffSetting("alarm_high_enabled")

var RangeHighMgdl = numberSetting("range_high_mgdl", 180, 70, 600, ParseInt)

var AlarmHighMinutes = numberSetting("alarm_high_minutes", 30, 0, MaxAlarmMinutes, ParseInt)

var AlarmUrgentHighEnabled = onOffSetting("alarm_urgent_high_enabled")

var RangeUrgentHighMgdl = numberSetting("range_urgent_high_mgdl", 270, 100, 600, ParseInt)

var AlarmFastRiseEnabled = onOffSetting("alarm_fast_rise_enabled")

var AlarmFastRiseMgdl = numberSetting("alarm_fast_rise_mgdl", 18, 5, 50, ParseInt)

// AlarmFastRiseMinutes is how far back the rise is measured, at least 5 minutes, as no change is seen over none
var AlarmFastRiseMinutes = numberSetting("alarm_fast_rise_minutes", 15, 5, MaxAlarmMinutes, ParseInt)

var AlarmFastDropEnabled = onOffSetting("alarm_fast_drop_enabled")

var AlarmFastDropMgdl = numberSetting("alarm_fast_drop_mgdl", 18, 5, 50, ParseInt)

// AlarmFastDropMinutes is how far back the drop is measured, at least 5 minutes, as no change is seen over none
var AlarmFastDropMinutes = numberSetting("alarm_fast_drop_minutes", 15, 5, MaxAlarmMinutes, ParseInt)

var AlarmNoDataEnabled = onOffSetting("alarm_no_data_enabled")

// AlarmNoDataMinutes is at least 10 minutes, as a reading comes every 5 and fewer would alarm between them
var AlarmNoDataMinutes = numberSetting("alarm_no_data_minutes", 20, 10, MaxAlarmMinutes, ParseInt)

var AlarmDeviceBatteryLowEnabled = onOffSetting("alarm_device_battery_low_enabled")

var AlarmDeviceBatteryLowPercent = numberSetting("alarm_device_battery_low_percent", 20, 5, 80, ParseInt)

// AlarmDeviceBatteryLowMinutes is how long a listed device can go without a status before it is alarmed for too
var AlarmDeviceBatteryLowMinutes = numberSetting("alarm_device_battery_low_minutes", 30, 5, MaxAlarmMinutes, ParseInt)

// AlarmDeviceBatteryLowDevicesJson are the names of the devices whose batteries and statuses the alarm checks, as
// their device statuses name them; none are checked by default
var AlarmDeviceBatteryLowDevicesJson = &SettingDefinition[types.DeviceNames]{
	Identifier:   "alarm_device_battery_low_devices",
	TypeString:   "array",
	Default:      types.DeviceNames{},
	ValidateFunc: ValidDeviceNames,
	ParseFunc:    ParseDeviceNames,
}

var AlarmLowReservoirEnabled = onOffSetting("alarm_low_reservoir_enabled")

var AlarmLowReservoirUnits = numberSetting("alarm_low_reservoir_units", 25, 0, 200, ParseInt)

var NotificationsEnabled = onOffSetting("notifications_enabled")
