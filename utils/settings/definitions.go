package settings

import "nighttroop.t1d.nz/utils/consts"

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

var Units = &SettingDefinition[consts.Units]{
	Identifier: "units",
	TypeString: "string",
	Default:    consts.UnitsMMOLL,
	ParseFunc:  ParseUnits,
}

var Timezone = &SettingDefinition[string]{
	Identifier: "timezone",
	TypeString: "string",
	Default:    "UTC",
	ParseFunc:  ParseString,
}

var Language = &SettingDefinition[string]{
	Identifier: "language",
	TypeString: "string",
	Default:    "en-NZ",
	ParseFunc:  ParseString,
}

var DateFormat = &SettingDefinition[string]{
	Identifier: "date_format",
	TypeString: "string",
	Default:    "YYYY-MM-DD",
	ParseFunc:  ParseString,
}

var TimeFormat = &SettingDefinition[string]{
	Identifier: "time_format",
	TypeString: "string",
	Default:    "HH:mm:ss",
	ParseFunc:  ParseString,
}

var WeekStartDay = &SettingDefinition[string]{
	Identifier: "week_start_day",
	TypeString: "string",
	Default:    "monday",
	ParseFunc:  ParseString,
}

var BolusDIAHours = &SettingDefinition[float64]{
	Identifier:   "bolus_dia_hours",
	TypeString:   "float",
	Default:      6.0,
	ValidateFunc: Range(1.0, 10.0),
	ParseFunc:    ParseFloat,
}

var BolusOnsetMinutes = &SettingDefinition[float64]{
	Identifier:   "bolus_onset_minutes",
	TypeString:   "float",
	Default:      20.0,
	ValidateFunc: Range(1.0, 60.0),
	ParseFunc:    ParseFloat,
}

var BolusPeakMinutes = &SettingDefinition[float64]{
	Identifier:   "bolus_peak_minutes",
	TypeString:   "float",
	Default:      60.0,
	ValidateFunc: Range(1.0, 120.0),
	ParseFunc:    ParseFloat,
}

var BasalDIAHours = &SettingDefinition[float64]{
	Identifier:   "basal_dia_hours",
	TypeString:   "float",
	Default:      24.0,
	ValidateFunc: Range(1.0, 48.0),
	ParseFunc:    ParseFloat,
}

var BasalOnsetMinutes = &SettingDefinition[float64]{
	Identifier:   "basal_onset_minutes",
	TypeString:   "float",
	Default:      60.0,
	ValidateFunc: Range(1.0, 120.0),
	ParseFunc:    ParseFloat,
}

var BasalPeakMinutes = &SettingDefinition[float64]{
	Identifier:   "basal_peak_minutes",
	TypeString:   "float",
	Default:      120.0,
	ValidateFunc: Range(1.0, 240.0),
	ParseFunc:    ParseFloat,
}

var DefaultInsulinSensitivity = &SettingDefinition[float64]{
	Identifier:   "default_insulin_sensitivity",
	TypeString:   "float",
	Default:      10.0,
	ValidateFunc: Range(1.0, 100.0),
	ParseFunc:    ParseFloat,
}

var DefaultCarbRatio = &SettingDefinition[float64]{
	Identifier:   "default_carb_ratio",
	TypeString:   "float",
	Default:      10.0,
	ValidateFunc: Range(1.0, 100.0),
	ParseFunc:    ParseFloat,
}

var DefaultBasalRate = &SettingDefinition[float64]{
	Identifier:   "default_basal_rate",
	TypeString:   "float",
	Default:      1.0,
	ValidateFunc: Range(0.1, 10.0),
	ParseFunc:    ParseFloat,
}

var CarbMaxAbsorptionHours = &SettingDefinition[float64]{
	Identifier:   "carb_max_absorption_hours",
	TypeString:   "float",
	Default:      6.0,
	ValidateFunc: Range(1.0, 24.0),
	ParseFunc:    ParseFloat,
}

var CarbMin5mImpact = &SettingDefinition[float64]{
	Identifier:   "carb_min_5m_impact",
	TypeString:   "float",
	Default:      10.0,
	ValidateFunc: Range(1.0, 100.0),
	ParseFunc:    ParseFloat,
}

var AlarmUrgentLowEnabled = &SettingDefinition[bool]{
	Identifier:   "alarm_urgent_low_enabled",
	TypeString:   "bool",
	Default:      false,
	ValidateFunc: nil,
	ParseFunc:    ParseBool,
}

var AlarmUrgentLow = &SettingDefinition[int64]{
	Identifier:   "alarm_urgent_low",
	TypeString:   "int64",
	Default:      56,
	ValidateFunc: nil,
	ParseFunc:    ParseInt,
}

var AlarmLowEnabled = &SettingDefinition[bool]{
	Identifier:   "alarm_low_enabled",
	TypeString:   "bool",
	Default:      false,
	ValidateFunc: nil,
	ParseFunc:    ParseBool,
}

var AlarmLow = &SettingDefinition[int64]{
	Identifier:   "alarm_low",
	TypeString:   "int64",
	Default:      72,
	ValidateFunc: nil,
	ParseFunc:    ParseInt,
}

var AlarmLowMinutes = &SettingDefinition[int64]{
	Identifier:   "alarm_low_minutes",
	TypeString:   "int64",
	Default:      5,
	ValidateFunc: nil,
	ParseFunc:    ParseInt,
}

var AlarmLowPredictionEnabled = &SettingDefinition[bool]{
	Identifier:   "alarm_low_prediction_enabled",
	TypeString:   "bool",
	Default:      false,
	ValidateFunc: nil,
	ParseFunc:    ParseBool,
}

var AlarmHighEnabled = &SettingDefinition[bool]{
	Identifier:   "alarm_high_enabled",
	TypeString:   "bool",
	Default:      false,
	ValidateFunc: nil,
	ParseFunc:    ParseBool,
}

var AlarmHigh = &SettingDefinition[int64]{
	Identifier:   "alarm_high",
	TypeString:   "int64",
	Default:      180,
	ValidateFunc: nil,
	ParseFunc:    ParseInt,
}

var AlarmHighMinutes = &SettingDefinition[int64]{
	Identifier:   "alarm_high_minutes",
	TypeString:   "int64",
	Default:      30,
	ValidateFunc: nil,
	ParseFunc:    ParseInt,
}

var AlarmUrgentHighEnabled = &SettingDefinition[bool]{
	Identifier:   "alarm_urgent_high_enabled",
	TypeString:   "bool",
	Default:      false,
	ValidateFunc: nil,
	ParseFunc:    ParseBool,
}

var AlarmUrgentHigh = &SettingDefinition[int64]{
	Identifier:   "alarm_urgent_high",
	TypeString:   "int64",
	Default:      270,
	ValidateFunc: nil,
	ParseFunc:    ParseInt,
}

var AlarmFastRiseEnabled = &SettingDefinition[bool]{
	Identifier:   "alarm_fast_rise_enabled",
	TypeString:   "bool",
	Default:      false,
	ValidateFunc: nil,
	ParseFunc:    ParseBool,
}

var AlarmFastRise = &SettingDefinition[int64]{
	Identifier:   "alarm_fast_rise",
	TypeString:   "int64",
	Default:      18,
	ValidateFunc: nil,
	ParseFunc:    ParseInt,
}

var AlarmFastRiseMinutes = &SettingDefinition[int64]{
	Identifier:   "alarm_fast_rise_minutes",
	TypeString:   "int64",
	Default:      15,
	ValidateFunc: nil,
	ParseFunc:    ParseInt,
}

var AlarmFastDropEnabled = &SettingDefinition[bool]{
	Identifier:   "alarm_fast_drop_enabled",
	TypeString:   "bool",
	Default:      false,
	ValidateFunc: nil,
	ParseFunc:    ParseBool,
}

var AlarmFastDrop = &SettingDefinition[int64]{
	Identifier:   "alarm_fast_drop",
	TypeString:   "int64",
	Default:      18,
	ValidateFunc: nil,
	ParseFunc:    ParseInt,
}

var AlarmFastDropMinutes = &SettingDefinition[int64]{
	Identifier:   "alarm_fast_drop_minutes",
	TypeString:   "int64",
	Default:      15,
	ValidateFunc: nil,
	ParseFunc:    ParseInt,
}

var AlarmNoDataEnabled = &SettingDefinition[bool]{
	Identifier:   "alarm_no_data_enabled",
	TypeString:   "bool",
	Default:      false,
	ValidateFunc: nil,
	ParseFunc:    ParseBool,
}

var AlarmNoDataMinutes = &SettingDefinition[int64]{
	Identifier:   "alarm_no_data_minutes",
	TypeString:   "int64",
	Default:      20,
	ValidateFunc: nil,
	ParseFunc:    ParseInt,
}

var AlarmBatteryLowEnabled = &SettingDefinition[bool]{
	Identifier:   "alarm_battery_low_enabled",
	TypeString:   "bool",
	Default:      false,
	ValidateFunc: nil,
	ParseFunc:    ParseBool,
}

var AlarmBatteryLowPercent = &SettingDefinition[int64]{
	Identifier:   "alarm_battery_low_percent",
	TypeString:   "int64",
	Default:      20,
	ValidateFunc: nil,
	ParseFunc:    ParseInt,
}

var AlarmBatteryLowPwdEnabled = &SettingDefinition[bool]{
	Identifier:   "alarm_battery_low_pwd_enabled",
	TypeString:   "bool",
	Default:      false,
	ValidateFunc: nil,
	ParseFunc:    ParseBool,
}

var AlarmBatteryLowPwdPercent = &SettingDefinition[int64]{
	Identifier:   "alarm_battery_low_pwd_percent",
	TypeString:   "int64",
	Default:      20,
	ValidateFunc: nil,
	ParseFunc:    ParseInt,
}

var AlarmLowReservoirEnabled = &SettingDefinition[bool]{
	Identifier:   "alarm_low_reservoir_enabled",
	TypeString:   "bool",
	Default:      false,
	ValidateFunc: nil,
	ParseFunc:    ParseBool,
}

var AlarmLowReservoirUnits = &SettingDefinition[int64]{
	Identifier:   "alarm_low_reservoir_units",
	TypeString:   "int64",
	Default:      25,
	ValidateFunc: nil,
	ParseFunc:    ParseInt,
}

var NotificationsEnabled = &SettingDefinition[bool]{
	Identifier:   "notifications_enabled",
	TypeString:   "bool",
	Default:      false,
	ValidateFunc: nil,
	ParseFunc:    ParseBool,
}
