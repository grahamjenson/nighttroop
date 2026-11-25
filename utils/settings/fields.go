package settings

import (
	"fmt"
	"reflect"
	"strings"
)

// field is one setting: its definition, and its place in a Settings and in a SettingsValues
type field struct {
	identifier, typeString string
	// path is the setting's place in the stored JSON, e.g. alarms.low.minutes, and in the two structs
	path string
	// stored and value are where the field is in a Settings and in a SettingsValues
	stored, value []int
	// parse is a value sent as the setting's own type, or why it is not valid, validate is why a value of the
	// setting's own type is not valid, and fallback is the default
	parse    func(any) (any, error)
	validate func(any) error
	fallback any
}

func settingAt[T any](definition *SettingDefinition[T], path string) field {
	return field{identifier: definition.Identifier, typeString: definition.TypeString, path: path, fallback: definition.Default,
		parse: func(value any) (any, error) {
			parsed, err := definition.ParseValidate(value)
			if err != nil {
				return nil, err
			}
			return *parsed, nil
		},
		validate: func(value any) error {
			typed, ok := value.(T)
			if !ok {
				return fmt.Errorf("%s is %T", definition.Identifier, value)
			}
			if definition.ValidateFunc == nil {
				return nil
			}
			return definition.ValidateFunc(typed)
		}}
}

// allSettings are every setting, in the order they are listed, each with where it is stored. Settings and
// SettingsValues both have a field at each place, of the setting's type (checked when the server starts, and by
// TestEverySettingHasItsField). byName is the same by identifier.
var allSettings, byName = func() ([]field, map[string]field) {
	fields := []field{
		settingAt(Units, "units"),
		settingAt(Timezone, "timezone"),
		settingAt(Language, "language"),
		settingAt(DateFormat, "date_format"),
		settingAt(TimeFormat, "time_format"),
		settingAt(WeekStartDay, "week_start_day"),
		settingAt(NumberFormat, "number_format"),
		settingAt(BolusDIAHours, "bolus.dia_hours"),
		settingAt(BolusOnsetMinutes, "bolus.onset_minutes"),
		settingAt(BolusPeakMinutes, "bolus.peak_minutes"),
		settingAt(BasalDIAHours, "basal.dia_hours"),
		settingAt(BasalOnsetMinutes, "basal.onset_minutes"),
		settingAt(BasalPeakMinutes, "basal.peak_minutes"),
		settingAt(DefaultInsulinSensitivity, "defaults.insulin_sensitivity"),
		settingAt(DefaultCarbRatio, "defaults.carb_ratio"),
		settingAt(DefaultBasalRate, "defaults.basal_rate"),
		settingAt(CarbMaxAbsorptionHours, "carbs.max_absorption_hours"),
		settingAt(CarbMin5mImpact, "carbs.min_5m_impact"),
		settingAt(AlarmUrgentLowEnabled, "alarms.urgent_low.enabled"),
		settingAt(RangeUrgentLowMgdl, "range.urgent_low_mgdl"),
		settingAt(AlarmLowEnabled, "alarms.low.enabled"),
		settingAt(RangeLowMgdl, "range.low_mgdl"),
		settingAt(AlarmLowMinutes, "alarms.low.minutes"),
		settingAt(AlarmLowPredictionEnabled, "alarms.low_prediction.enabled"),
		settingAt(AlarmLowPredictionMinutes, "alarms.low_prediction.minutes"),
		settingAt(AlarmHighEnabled, "alarms.high.enabled"),
		settingAt(RangeHighMgdl, "range.high_mgdl"),
		settingAt(AlarmHighMinutes, "alarms.high.minutes"),
		settingAt(AlarmUrgentHighEnabled, "alarms.urgent_high.enabled"),
		settingAt(RangeUrgentHighMgdl, "range.urgent_high_mgdl"),
		settingAt(AlarmFastRiseEnabled, "alarms.fast_rise.enabled"),
		settingAt(AlarmFastRiseMgdl, "alarms.fast_rise.mgdl"),
		settingAt(AlarmFastRiseMinutes, "alarms.fast_rise.minutes"),
		settingAt(AlarmFastDropEnabled, "alarms.fast_drop.enabled"),
		settingAt(AlarmFastDropMgdl, "alarms.fast_drop.mgdl"),
		settingAt(AlarmFastDropMinutes, "alarms.fast_drop.minutes"),
		settingAt(AlarmNoDataEnabled, "alarms.no_data.enabled"),
		settingAt(AlarmNoDataMinutes, "alarms.no_data.minutes"),
		settingAt(AlarmDeviceBatteryLowEnabled, "alarms.device_battery_low.enabled"),
		settingAt(AlarmDeviceBatteryLowPercent, "alarms.device_battery_low.percent"),
		settingAt(AlarmDeviceBatteryLowMinutes, "alarms.device_battery_low.minutes"),
		settingAt(AlarmDeviceBatteryLowDevicesJson, "alarms.device_battery_low.devices"),
		settingAt(AlarmLowReservoirEnabled, "alarms.low_reservoir.enabled"),
		settingAt(AlarmLowReservoirUnits, "alarms.low_reservoir.units"),
		settingAt(NotificationsEnabled, "notifications.enabled"),
	}
	stored, values, byName := map[string][]int{}, map[string][]int{}, map[string]field{}
	placesOf(reflect.TypeFor[Settings](), "", nil, stored)
	placesOf(reflect.TypeFor[SettingsValues](), "", nil, values)
	for i, f := range fields {
		at, ok := stored[f.path]
		in, also := values[f.path]
		if !ok || !also {
			panic("settings: no field at " + f.path)
		}
		// The stored field is a pointer to what the value's field is, which is what the setting's default is
		storedType, valueType := reflect.TypeFor[Settings]().FieldByIndex(at).Type, reflect.TypeFor[SettingsValues]().FieldByIndex(in).Type
		if storedType != reflect.PointerTo(valueType) || valueType != reflect.TypeOf(f.fallback) {
			panic("settings: the fields at " + f.path + " are not of its setting's type")
		}
		fields[i].stored, fields[i].value = at, in
		byName[f.identifier] = fields[i]
	}
	return fields, byName
}()

// placesOf is where each field of a struct of settings is, by its place in the stored JSON: the JSON names of a
// Settings, which a SettingsValues' fields have the same Go names as
func placesOf(t reflect.Type, path string, index []int, places map[string][]int) {
	for i := range t.NumField() {
		name := jsonName(reflect.TypeFor[Settings](), path, t.Field(i))
		if name == "" {
			continue
		}
		at, in := strings.TrimPrefix(path+"."+name, "."), append(index[:len(index):len(index)], i)
		if t.Field(i).Type.Kind() == reflect.Struct {
			placesOf(t.Field(i).Type, at, in, places)
			continue
		}
		places[at] = in
	}
}

// jsonName is a field's JSON name: its own tag's, or for a SettingsValues, which has none, that of the field of the
// same Go name at the same place in a Settings
func jsonName(stored reflect.Type, path string, f reflect.StructField) string {
	if tag := f.Tag.Get("json"); tag != "" {
		name, _, _ := strings.Cut(tag, ",")
		return name
	}
	if path != "" {
		for _, part := range strings.Split(path, ".") {
			for i := range stored.NumField() {
				if name, _, _ := strings.Cut(stored.Field(i).Tag.Get("json"), ","); name == part {
					stored = stored.Field(i).Type
					break
				}
			}
		}
	}
	same, ok := stored.FieldByName(f.Name)
	if !ok {
		return ""
	}
	name, _, _ := strings.Cut(same.Tag.Get("json"), ",")
	return name
}

// own is the setting's value in s, and whether it is set
func (f field) own(s *Settings) (any, bool) {
	at := reflect.ValueOf(s).Elem().FieldByIndex(f.stored)
	if at.IsNil() {
		return nil, false
	}
	return at.Elem().Interface(), true
}

// set sets the setting in s to a value of its own type, and clear takes it out
func (f field) set(s *Settings, value any) {
	at := reflect.ValueOf(s).Elem().FieldByIndex(f.stored)
	at.Set(reflect.New(at.Type().Elem()))
	at.Elem().Set(reflect.ValueOf(value))
}

func (f field) clear(s *Settings) {
	reflect.ValueOf(s).Elem().FieldByIndex(f.stored).SetZero()
}

// setValue sets the setting in v to a value of its own type
func (f field) setValue(v *SettingsValues, value any) {
	reflect.ValueOf(v).Elem().FieldByIndex(f.value).Set(reflect.ValueOf(value))
}

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
var defaults = func() SettingsValues {
	var v SettingsValues
	for _, f := range allSettings {
		f.setValue(&v, f.fallback)
	}
	return v
}()

// DefaultValues are every setting at its default
func DefaultValues() SettingsValues {
	return defaults
}
