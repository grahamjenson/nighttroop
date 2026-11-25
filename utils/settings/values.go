package settings

import (
	"reflect"
	"strings"
	"time"

	"nighttroop.t1d.nz/data/types"
	"nighttroop.t1d.nz/utils/consts"
)

// Version is the version of Values that is written. It changes only when what a stored value means changes, with
// code here that reads the older versions: a new field does not change it.
const Version = 1

// Values is every setting with its value. It is what the code reads a setting from, e.g. Values.Range.LowMgdl, and
// its JSON is what a level's settings_json column holds, with only the settings that level has set.
//
// Rows of every release are read by this struct's JSON names, so it only grows: a field is added, and never renamed,
// removed, moved or given another type or JSON name. TestValuesOnlyGrow fails when one is. AGENTS.md has the rule.
type Values struct {
	Version int `json:"version"`

	Units        consts.Units `json:"units"`
	Timezone     string       `json:"timezone"`
	Language     string       `json:"language"`
	DateFormat   string       `json:"date_format"`
	TimeFormat   string       `json:"time_format"`
	WeekStartDay string       `json:"week_start_day"`
	NumberFormat string       `json:"number_format"`

	Bolus         Insulin       `json:"bolus"`
	Basal         Insulin       `json:"basal"`
	Defaults      Defaults      `json:"defaults"`
	Carbs         Carbs         `json:"carbs"`
	Range         GlucoseRange  `json:"range"`
	Alarms        Alarms        `json:"alarms"`
	Notifications Notifications `json:"notifications"`
}

// Insulin is how an insulin acts: how long, and when it starts and peaks
type Insulin struct {
	DiaHours     float64 `json:"dia_hours"`
	OnsetMinutes float64 `json:"onset_minutes"`
	PeakMinutes  float64 `json:"peak_minutes"`
}

// Defaults are the profile values used when a site has no profile
type Defaults struct {
	InsulinSensitivity float64 `json:"insulin_sensitivity"`
	CarbRatio          float64 `json:"carb_ratio"`
	BasalRate          float64 `json:"basal_rate"`
}

// Carbs is how carbs are absorbed
type Carbs struct {
	MaxAbsorptionHours float64 `json:"max_absorption_hours"`
	Min5mImpact        float64 `json:"min_5m_impact"`
}

// GlucoseRange is the glucose range in mg/dL, which the graphs, the summaries and the glucose alarms share
type GlucoseRange struct {
	UrgentLowMgdl  int64 `json:"urgent_low_mgdl"`
	LowMgdl        int64 `json:"low_mgdl"`
	HighMgdl       int64 `json:"high_mgdl"`
	UrgentHighMgdl int64 `json:"urgent_high_mgdl"`
}

// Alarms is each alarm: whether it is on, and what it waits for
type Alarms struct {
	UrgentLow        Alarm          `json:"urgent_low"`
	Low              TimedAlarm     `json:"low"`
	LowPrediction    TimedAlarm     `json:"low_prediction"`
	High             TimedAlarm     `json:"high"`
	UrgentHigh       Alarm          `json:"urgent_high"`
	FastRise         ChangeAlarm    `json:"fast_rise"`
	FastDrop         ChangeAlarm    `json:"fast_drop"`
	NoData           TimedAlarm     `json:"no_data"`
	DeviceBatteryLow BatteryAlarm   `json:"device_battery_low"`
	LowReservoir     ReservoirAlarm `json:"low_reservoir"`
}

// Alarm is an alarm that is on or off
type Alarm struct {
	Enabled bool `json:"enabled"`
}

// TimedAlarm is an alarm that waits for, or looks ahead, some minutes
type TimedAlarm struct {
	Enabled bool  `json:"enabled"`
	Minutes int64 `json:"minutes"`
}

// ChangeAlarm is an alarm for glucose changing by so many mg/dL over some minutes
type ChangeAlarm struct {
	Enabled bool  `json:"enabled"`
	Mgdl    int64 `json:"mgdl"`
	Minutes int64 `json:"minutes"`
}

// BatteryAlarm is the alarm for the listed devices' batteries, and for one going quiet
type BatteryAlarm struct {
	Enabled bool              `json:"enabled"`
	Percent int64             `json:"percent"`
	Minutes int64             `json:"minutes"`
	Devices types.DeviceNames `json:"devices"`
}

// ReservoirAlarm is the alarm for a pump's reservoir running low
type ReservoirAlarm struct {
	Enabled bool  `json:"enabled"`
	Units   int64 `json:"units"`
}

// Notifications is whether the identity is sent notifications
type Notifications struct {
	Enabled bool `json:"enabled"`
}

// field is one setting's place in a Values
type field struct {
	identifier, typeString string
	// at is the field itself, a pointer to it, and path its place in the JSON, e.g. alarms.low.minutes
	at   any
	path []string
	// get is the value, set sets a value of the setting's own type, parse is a value sent or stored as the setting's
	// own type, or why it is not valid, and fallback is the default
	get      func() any
	set      func(any)
	parse    func(any) (any, error)
	fallback any
}

func fieldOf[T any](definition *SettingDefinition[T], at *T) field {
	return field{identifier: definition.Identifier, typeString: definition.TypeString, at: at, fallback: definition.Default,
		get: func() any { return *at },
		set: func(value any) { *at = value.(T) },
		parse: func(value any) (any, error) {
			parsed, err := definition.ParseValidate(value)
			if err != nil {
				return nil, err
			}
			return *parsed, nil
		}}
}

// fields are where v keeps each setting, in the order the settings are listed. A setting is in exactly one place
// (TestEverySettingHasItsField).
func (v *Values) fields() []field {
	fields := []field{
		fieldOf(Units, &v.Units),
		fieldOf(Timezone, &v.Timezone),
		fieldOf(Language, &v.Language),
		fieldOf(DateFormat, &v.DateFormat),
		fieldOf(TimeFormat, &v.TimeFormat),
		fieldOf(WeekStartDay, &v.WeekStartDay),
		fieldOf(NumberFormat, &v.NumberFormat),
		fieldOf(BolusDIAHours, &v.Bolus.DiaHours),
		fieldOf(BolusOnsetMinutes, &v.Bolus.OnsetMinutes),
		fieldOf(BolusPeakMinutes, &v.Bolus.PeakMinutes),
		fieldOf(BasalDIAHours, &v.Basal.DiaHours),
		fieldOf(BasalOnsetMinutes, &v.Basal.OnsetMinutes),
		fieldOf(BasalPeakMinutes, &v.Basal.PeakMinutes),
		fieldOf(DefaultInsulinSensitivity, &v.Defaults.InsulinSensitivity),
		fieldOf(DefaultCarbRatio, &v.Defaults.CarbRatio),
		fieldOf(DefaultBasalRate, &v.Defaults.BasalRate),
		fieldOf(CarbMaxAbsorptionHours, &v.Carbs.MaxAbsorptionHours),
		fieldOf(CarbMin5mImpact, &v.Carbs.Min5mImpact),
		fieldOf(AlarmUrgentLowEnabled, &v.Alarms.UrgentLow.Enabled),
		fieldOf(RangeUrgentLowMgdl, &v.Range.UrgentLowMgdl),
		fieldOf(AlarmLowEnabled, &v.Alarms.Low.Enabled),
		fieldOf(RangeLowMgdl, &v.Range.LowMgdl),
		fieldOf(AlarmLowMinutes, &v.Alarms.Low.Minutes),
		fieldOf(AlarmLowPredictionEnabled, &v.Alarms.LowPrediction.Enabled),
		fieldOf(AlarmLowPredictionMinutes, &v.Alarms.LowPrediction.Minutes),
		fieldOf(AlarmHighEnabled, &v.Alarms.High.Enabled),
		fieldOf(RangeHighMgdl, &v.Range.HighMgdl),
		fieldOf(AlarmHighMinutes, &v.Alarms.High.Minutes),
		fieldOf(AlarmUrgentHighEnabled, &v.Alarms.UrgentHigh.Enabled),
		fieldOf(RangeUrgentHighMgdl, &v.Range.UrgentHighMgdl),
		fieldOf(AlarmFastRiseEnabled, &v.Alarms.FastRise.Enabled),
		fieldOf(AlarmFastRiseMgdl, &v.Alarms.FastRise.Mgdl),
		fieldOf(AlarmFastRiseMinutes, &v.Alarms.FastRise.Minutes),
		fieldOf(AlarmFastDropEnabled, &v.Alarms.FastDrop.Enabled),
		fieldOf(AlarmFastDropMgdl, &v.Alarms.FastDrop.Mgdl),
		fieldOf(AlarmFastDropMinutes, &v.Alarms.FastDrop.Minutes),
		fieldOf(AlarmNoDataEnabled, &v.Alarms.NoData.Enabled),
		fieldOf(AlarmNoDataMinutes, &v.Alarms.NoData.Minutes),
		fieldOf(AlarmDeviceBatteryLowEnabled, &v.Alarms.DeviceBatteryLow.Enabled),
		fieldOf(AlarmDeviceBatteryLowPercent, &v.Alarms.DeviceBatteryLow.Percent),
		fieldOf(AlarmDeviceBatteryLowMinutes, &v.Alarms.DeviceBatteryLow.Minutes),
		fieldOf(AlarmDeviceBatteryLowDevicesJson, &v.Alarms.DeviceBatteryLow.Devices),
		fieldOf(AlarmLowReservoirEnabled, &v.Alarms.LowReservoir.Enabled),
		fieldOf(AlarmLowReservoirUnits, &v.Alarms.LowReservoir.Units),
		fieldOf(NotificationsEnabled, &v.Notifications.Enabled),
	}
	return fields
}

// jsonPaths is each field of a struct, by its address, with its place in the struct's JSON
func jsonPaths(v reflect.Value, path []string, paths map[any][]string) {
	for i := range v.NumField() {
		name, _, _ := strings.Cut(v.Type().Field(i).Tag.Get("json"), ",")
		at := append(path[:len(path):len(path)], name)
		if v.Field(i).Kind() == reflect.Struct {
			jsonPaths(v.Field(i), at, paths)
			continue
		}
		paths[v.Field(i).Addr().Interface()] = at
	}
}

// allSettings are every setting, in the order they are listed, each with its place in the JSON, and byName the same by
// identifier. They are worked out once: what differs between two Values is only the values.
var allSettings, byName = func() ([]field, map[string]field) {
	v := &Values{}
	fields, paths, byName := v.fields(), map[any][]string{}, map[string]field{}
	jsonPaths(reflect.ValueOf(v).Elem(), nil, paths)
	for i := range fields {
		fields[i].path = paths[fields[i].at]
		byName[fields[i].identifier] = fields[i]
	}
	return fields, byName
}()

// names are the identifiers of every setting, in the order they are listed
var names = func() []string {
	names := make([]string, len(allSettings))
	for i, f := range allSettings {
		names[i] = f.identifier
	}
	return names
}()

// IsSetting is whether there is a setting with the identifier
func IsSetting(identifier string) bool {
	_, ok := byName[identifier]
	return ok
}

// defaults are every setting at its default
var defaults = func() Values {
	v := Values{Version: Version}
	for _, f := range v.fields() {
		f.set(f.fallback)
	}
	return v
}()

// DefaultValues are every setting at its default
func DefaultValues() Values {
	v := defaults
	v.Alarms.DeviceBatteryLow.Devices = append(types.DeviceNames{}, defaults.Alarms.DeviceBatteryLow.Devices...)
	return v
}

// Location is the timezone times are shown in, or UTC if it cannot be loaded
func (v Values) Location() *time.Location {
	location, err := LoadLocation(v.Timezone)
	if err != nil {
		return time.UTC
	}
	return location
}

// DateLayout is the Go layout of the date format, e.g. 02/01/2006 for DD/MM/YYYY
func (v Values) DateLayout() string {
	return strings.NewReplacer("YYYY", "2006", "MM", "01", "DD", "02").Replace(v.DateFormat)
}

// timeLayouts are the Go layouts of the time formats, e.g. 03:04 PM for hh:mm a
var timeLayouts = strings.NewReplacer("HH", "15", "hh", "03", "mm", "04", "ss", "05", "a", "PM")

// TimeLayout is the Go layout of the time format, e.g. 03:04 PM for hh:mm a
func (v Values) TimeLayout() string {
	return timeLayouts.Replace(v.TimeFormat)
}
