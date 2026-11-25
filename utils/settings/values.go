package settings

import (
	"sync"
	"time"

	"nighttroop.t1d.nz/data/types"
	"nighttroop.t1d.nz/utils/consts"
)

// SettingsValues is every setting with a value: the same fields as Settings, with no pointers, so none is ever
// "not set". The defaults are one, and so are the values in effect for a request.
type SettingsValues struct {
	Units        consts.Units
	Timezone     string
	Language     string
	DateFormat   string
	TimeFormat   string
	WeekStartDay string
	NumberFormat string

	Bolus         InsulinValues
	Basal         InsulinValues
	Defaults      DefaultsValues
	Carbs         CarbsValues
	Range         GlucoseRangeValues
	Alarms        AlarmsValues
	Notifications NotificationsValues
}

type InsulinValues struct {
	DiaHours, OnsetMinutes, PeakMinutes float64
}

type DefaultsValues struct {
	InsulinSensitivity, CarbRatio, BasalRate float64
}

type CarbsValues struct {
	MaxAbsorptionHours, Min5mImpact float64
}

type GlucoseRangeValues struct {
	UrgentLowMgdl, LowMgdl, HighMgdl, UrgentHighMgdl int64
}

type AlarmsValues struct {
	UrgentLow        AlarmValues
	Low              TimedAlarmValues
	LowPrediction    TimedAlarmValues
	High             TimedAlarmValues
	UrgentHigh       AlarmValues
	FastRise         ChangeAlarmValues
	FastDrop         ChangeAlarmValues
	NoData           TimedAlarmValues
	DeviceBatteryLow BatteryAlarmValues
	LowReservoir     ReservoirAlarmValues
}

type AlarmValues struct {
	Enabled bool
}

type TimedAlarmValues struct {
	Enabled bool
	Minutes int64
}

type ChangeAlarmValues struct {
	Enabled       bool
	Mgdl, Minutes int64
}

type BatteryAlarmValues struct {
	Enabled          bool
	Percent, Minutes int64
	Devices          types.DeviceNames
}

type ReservoirAlarmValues struct {
	Enabled bool
	Units   int64
}

type NotificationsValues struct {
	Enabled bool
}

// locations are the timezones loaded, by name, as a page shows many times in one
var locations sync.Map

// LoadLocation is the named timezone, loaded once and then kept, as time.LoadLocation reads its zone file each time.
// Only a timezone that loads is kept, so there are at most as many as the zone database has.
func LoadLocation(name string) (*time.Location, error) {
	if location, ok := locations.Load(name); ok {
		return location.(*time.Location), nil
	}
	location, err := time.LoadLocation(name)
	if err != nil {
		return nil, err
	}
	locations.Store(name, location)
	return location, nil
}

// locationOrUTC is the named timezone, or UTC if it cannot be loaded
func locationOrUTC(name string) *time.Location {
	location, err := LoadLocation(name)
	if err != nil {
		return time.UTC
	}
	return location
}
