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

// The values the date, time and week start settings accept, in the order the settings pages offer them
var (
	DateFormats   = []string{"YYYY/MM/DD", "DD/MM/YYYY", "MM/DD/YYYY", "YYYY-MM-DD", "DD-MM-YYYY", "MM-DD-YYYY"}
	TimeFormats   = []string{"HH:mm:ss", "hh:mm:ss a", "HH:mm", "hh:mm a"}
	WeekStartDays = []string{"monday", "sunday"}
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

var BolusDIAHours = &SettingDefinition[float64]{
	Identifier:   "bolus_dia_hours",
	TypeString:   "number",
	Default:      6.0,
	ValidateFunc: Range(1.0, 10.0),
	ParseFunc:    ParseFloat,
}

var BolusOnsetMinutes = &SettingDefinition[float64]{
	Identifier:   "bolus_onset_minutes",
	TypeString:   "number",
	Default:      20.0,
	ValidateFunc: Range(1.0, 120.0),
	ParseFunc:    ParseFloat,
}

var BolusPeakMinutes = &SettingDefinition[float64]{
	Identifier:   "bolus_peak_minutes",
	TypeString:   "number",
	Default:      60.0,
	ValidateFunc: Range(1.0, 240.0),
	ParseFunc:    ParseFloat,
}

var BasalDIAHours = &SettingDefinition[float64]{
	Identifier:   "basal_dia_hours",
	TypeString:   "number",
	Default:      24.0,
	ValidateFunc: Range(1.0, 170.0),
	ParseFunc:    ParseFloat,
}

var BasalOnsetMinutes = &SettingDefinition[float64]{
	Identifier:   "basal_onset_minutes",
	TypeString:   "number",
	Default:      60.0,
	ValidateFunc: Range(1.0, 360.0),
	ParseFunc:    ParseFloat,
}

var BasalPeakMinutes = &SettingDefinition[float64]{
	Identifier:   "basal_peak_minutes",
	TypeString:   "number",
	Default:      120.0,
	ValidateFunc: Range(1.0, 240.0),
	ParseFunc:    ParseFloat,
}

var DefaultInsulinSensitivity = &SettingDefinition[float64]{
	Identifier:   "default_insulin_sensitivity",
	TypeString:   "number",
	Default:      10.0,
	ValidateFunc: Range(1.0, 100.0),
	ParseFunc:    ParseFloat,
}

var DefaultCarbRatio = &SettingDefinition[float64]{
	Identifier:   "default_carb_ratio",
	TypeString:   "number",
	Default:      10.0,
	ValidateFunc: Range(1.0, 100.0),
	ParseFunc:    ParseFloat,
}

var DefaultBasalRate = &SettingDefinition[float64]{
	Identifier:   "default_basal_rate",
	TypeString:   "number",
	Default:      1.0,
	ValidateFunc: Range(0.1, 10.0),
	ParseFunc:    ParseFloat,
}

var CarbMaxAbsorptionHours = &SettingDefinition[float64]{
	Identifier:   "carb_max_absorption_hours",
	TypeString:   "number",
	Default:      6.0,
	ValidateFunc: Range(1.0, 24.0),
	ParseFunc:    ParseFloat,
}

var CarbMin5mImpact = &SettingDefinition[float64]{
	Identifier:   "carb_min_5m_impact",
	TypeString:   "number",
	Default:      10.0,
	ValidateFunc: Range(1.0, 100.0),
	ParseFunc:    ParseFloat,
}

var AlarmUrgentLowEnabled = &SettingDefinition[bool]{
	Identifier: "alarm_urgent_low_enabled",
	TypeString: "boolean",
	Default:    false,
	ParseFunc:  ParseBool,
}

var AlarmUrgentLowMgdl = &SettingDefinition[int64]{
	Identifier:   "alarm_urgent_low_mgdl",
	TypeString:   "number",
	Default:      56,
	ValidateFunc: Range(int64(20), 100),
	ParseFunc:    ParseInt,
}

var AlarmLowEnabled = &SettingDefinition[bool]{
	Identifier: "alarm_low_enabled",
	TypeString: "boolean",
	Default:    false,
	ParseFunc:  ParseBool,
}

// MaxAlarmMinutes is the most minutes an alarm can wait for, so a check of the alarms reads only the readings it needs
const MaxAlarmMinutes = 120

var AlarmLowMgdl = &SettingDefinition[int64]{
	Identifier:   "alarm_low_mgdl",
	TypeString:   "number",
	Default:      72,
	ValidateFunc: Range(int64(20), 150),
	ParseFunc:    ParseInt,
}

var AlarmLowMinutes = &SettingDefinition[int64]{
	Identifier:   "alarm_low_minutes",
	TypeString:   "number",
	Default:      5,
	ValidateFunc: Range(int64(0), MaxAlarmMinutes),
	ParseFunc:    ParseInt,
}

var AlarmLowPredictionEnabled = &SettingDefinition[bool]{
	Identifier: "alarm_low_prediction_enabled",
	TypeString: "boolean",
	Default:    false,
	ParseFunc:  ParseBool,
}

// AlarmLowPredictionMinutes is how far ahead the predicted low alarm follows the current trend to the low alarm's glucose
var AlarmLowPredictionMinutes = &SettingDefinition[int64]{
	Identifier:   "alarm_low_prediction_minutes",
	TypeString:   "number",
	Default:      20,
	ValidateFunc: Range(int64(5), 60),
	ParseFunc:    ParseInt,
}

var AlarmHighEnabled = &SettingDefinition[bool]{
	Identifier: "alarm_high_enabled",
	TypeString: "boolean",
	Default:    false,
	ParseFunc:  ParseBool,
}

var AlarmHighMgdl = &SettingDefinition[int64]{
	Identifier:   "alarm_high_mgdl",
	TypeString:   "number",
	Default:      180,
	ValidateFunc: Range(int64(70), 600),
	ParseFunc:    ParseInt,
}

var AlarmHighMinutes = &SettingDefinition[int64]{
	Identifier:   "alarm_high_minutes",
	TypeString:   "number",
	Default:      30,
	ValidateFunc: Range(int64(0), MaxAlarmMinutes),
	ParseFunc:    ParseInt,
}

var AlarmUrgentHighEnabled = &SettingDefinition[bool]{
	Identifier: "alarm_urgent_high_enabled",
	TypeString: "boolean",
	Default:    false,
	ParseFunc:  ParseBool,
}

var AlarmUrgentHighMgdl = &SettingDefinition[int64]{
	Identifier:   "alarm_urgent_high_mgdl",
	TypeString:   "number",
	Default:      270,
	ValidateFunc: Range(int64(100), 600),
	ParseFunc:    ParseInt,
}

var AlarmFastRiseEnabled = &SettingDefinition[bool]{
	Identifier: "alarm_fast_rise_enabled",
	TypeString: "boolean",
	Default:    false,
	ParseFunc:  ParseBool,
}

var AlarmFastRiseMgdl = &SettingDefinition[int64]{
	Identifier:   "alarm_fast_rise_mgdl",
	TypeString:   "number",
	Default:      18,
	ValidateFunc: Range(int64(5), 50),
	ParseFunc:    ParseInt,
}

// AlarmFastRiseMinutes is how far back the rise is measured, at least 5 minutes, as no change is seen over none
var AlarmFastRiseMinutes = &SettingDefinition[int64]{
	Identifier:   "alarm_fast_rise_minutes",
	TypeString:   "number",
	Default:      15,
	ValidateFunc: Range(int64(5), MaxAlarmMinutes),
	ParseFunc:    ParseInt,
}

var AlarmFastDropEnabled = &SettingDefinition[bool]{
	Identifier: "alarm_fast_drop_enabled",
	TypeString: "boolean",
	Default:    false,
	ParseFunc:  ParseBool,
}

var AlarmFastDropMgdl = &SettingDefinition[int64]{
	Identifier:   "alarm_fast_drop_mgdl",
	TypeString:   "number",
	Default:      18,
	ValidateFunc: Range(int64(5), 50),
	ParseFunc:    ParseInt,
}

// AlarmFastDropMinutes is how far back the drop is measured, at least 5 minutes, as no change is seen over none
var AlarmFastDropMinutes = &SettingDefinition[int64]{
	Identifier:   "alarm_fast_drop_minutes",
	TypeString:   "number",
	Default:      15,
	ValidateFunc: Range(int64(5), MaxAlarmMinutes),
	ParseFunc:    ParseInt,
}

var AlarmNoDataEnabled = &SettingDefinition[bool]{
	Identifier: "alarm_no_data_enabled",
	TypeString: "boolean",
	Default:    false,
	ParseFunc:  ParseBool,
}

// AlarmNoDataMinutes is at least 10 minutes, as a reading comes every 5 and fewer would alarm between them
var AlarmNoDataMinutes = &SettingDefinition[int64]{
	Identifier:   "alarm_no_data_minutes",
	TypeString:   "number",
	Default:      20,
	ValidateFunc: Range(int64(10), MaxAlarmMinutes),
	ParseFunc:    ParseInt,
}

var AlarmDeviceBatteryLowEnabled = &SettingDefinition[bool]{
	Identifier: "alarm_device_battery_low_enabled",
	TypeString: "boolean",
	Default:    false,
	ParseFunc:  ParseBool,
}

var AlarmDeviceBatteryLowPercent = &SettingDefinition[int64]{
	Identifier:   "alarm_device_battery_low_percent",
	TypeString:   "number",
	Default:      20,
	ValidateFunc: Range(int64(5), 80),
	ParseFunc:    ParseInt,
}

// AlarmDeviceBatteryLowMinutes is how long a listed device can go without a status before it is alarmed for too
var AlarmDeviceBatteryLowMinutes = &SettingDefinition[int64]{
	Identifier:   "alarm_device_battery_low_minutes",
	TypeString:   "number",
	Default:      30,
	ValidateFunc: Range(int64(5), MaxAlarmMinutes),
	ParseFunc:    ParseInt,
}

// AlarmDeviceBatteryLowDevicesJson are the names of the devices whose batteries and statuses the alarm checks, as
// their device statuses name them; none are checked by default
var AlarmDeviceBatteryLowDevicesJson = &SettingDefinition[types.DeviceNames]{
	Identifier:   "alarm_device_battery_low_devices",
	TypeString:   "array",
	Default:      types.DeviceNames{},
	ValidateFunc: ValidDeviceNames,
	ParseFunc:    ParseDeviceNames,
}

var AlarmLowReservoirEnabled = &SettingDefinition[bool]{
	Identifier: "alarm_low_reservoir_enabled",
	TypeString: "boolean",
	Default:    false,
	ParseFunc:  ParseBool,
}

var AlarmLowReservoirUnits = &SettingDefinition[int64]{
	Identifier:   "alarm_low_reservoir_units",
	TypeString:   "number",
	Default:      25,
	ValidateFunc: Range(int64(0), 200),
	ParseFunc:    ParseInt,
}

var NotificationsEnabled = &SettingDefinition[bool]{
	Identifier: "notifications_enabled",
	TypeString: "boolean",
	Default:    false,
	ParseFunc:  ParseBool,
}
