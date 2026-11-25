package alarms

import "nighttroop.t1d.nz/utils/settings"

// FromSettings is the alarms turned on in the settings in effect for an identity, in the order they are checked
func FromSettings(v settings.SettingsValues) []Alarm {
	on, glucose := v.Alarms, v.Range
	var alarms []Alarm
	if on.UrgentLow.Enabled {
		alarms = append(alarms, UrgentLowAlarm{Mgdl: float64(glucose.UrgentLowMgdl)})
	}
	if on.Low.Enabled {
		alarms = append(alarms, LowAlarm{Mgdl: float64(glucose.LowMgdl), Minutes: on.Low.Minutes})
	}
	if on.LowPrediction.Enabled {
		alarms = append(alarms, PredictedLowAlarm{Mgdl: float64(glucose.LowMgdl), Minutes: on.LowPrediction.Minutes})
	}
	if on.UrgentHigh.Enabled {
		alarms = append(alarms, UrgentHighAlarm{Mgdl: float64(glucose.UrgentHighMgdl)})
	}
	if on.High.Enabled {
		alarms = append(alarms, HighAlarm{Mgdl: float64(glucose.HighMgdl), Minutes: on.High.Minutes})
	}
	if on.FastRise.Enabled {
		alarms = append(alarms, FastRiseAlarm{Mgdl: float64(on.FastRise.Mgdl), Minutes: on.FastRise.Minutes})
	}
	if on.FastDrop.Enabled {
		alarms = append(alarms, FastDropAlarm{Mgdl: float64(on.FastDrop.Mgdl), Minutes: on.FastDrop.Minutes})
	}
	if on.NoData.Enabled {
		alarms = append(alarms, NoDataAlarm{Minutes: on.NoData.Minutes})
	}
	// With no devices listed there is nothing for it to check
	if battery := on.DeviceBatteryLow; battery.Enabled && len(battery.Devices) > 0 {
		alarms = append(alarms, DeviceBatteryAlarm{Percent: float64(battery.Percent), Minutes: battery.Minutes, Devices: battery.Devices})
	}
	return alarms
}

// Checked is what a check found for an identity, which the checker sends as an alarm event when it notifies or the
// identity's alerting alarms change
type Checked struct {
	IdentityID    int64
	Notifications []Notification
	Alerting      []Type
	// WasAlerting are those that were alerting before the check, as the checker last sent them
	WasAlerting []Type
}
