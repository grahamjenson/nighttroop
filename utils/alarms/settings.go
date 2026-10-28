package alarms

import "nighttroop.t1d.nz/utils/settings"

// FromSettings is the alarms turned on in an identity's settings, each setting the identity's, then the site's, then
// the default
func FromSettings(m *settings.CombinedSettingsMap) []Alarm {
	mgdl := func(get func(settings.SettingMap) *int64) float64 { return float64(resolve(m, get)) }
	var alarms []Alarm
	if resolve(m, settings.SettingMap.GetAlarmUrgentLowEnabled) {
		alarms = append(alarms, UrgentLowAlarm{Mgdl: mgdl(settings.SettingMap.GetAlarmUrgentLow)})
	}
	if resolve(m, settings.SettingMap.GetAlarmLowEnabled) {
		alarms = append(alarms, LowAlarm{Mgdl: mgdl(settings.SettingMap.GetAlarmLow), Minutes: resolve(m, settings.SettingMap.GetAlarmLowMinutes)})
	}
	if resolve(m, settings.SettingMap.GetAlarmUrgentHighEnabled) {
		alarms = append(alarms, UrgentHighAlarm{Mgdl: mgdl(settings.SettingMap.GetAlarmUrgentHigh)})
	}
	if resolve(m, settings.SettingMap.GetAlarmHighEnabled) {
		alarms = append(alarms, HighAlarm{Mgdl: mgdl(settings.SettingMap.GetAlarmHigh), Minutes: resolve(m, settings.SettingMap.GetAlarmHighMinutes)})
	}
	if resolve(m, settings.SettingMap.GetAlarmNoDataEnabled) {
		alarms = append(alarms, NoDataAlarm{Minutes: resolve(m, settings.SettingMap.GetAlarmNoDataMinutes)})
	}
	return alarms
}

// resolve is a setting's value from the first of the identity's, the site's and the default settings that has it
func resolve[T any](m *settings.CombinedSettingsMap, get func(settings.SettingMap) *T) T {
	for _, level := range []*settings.SettingMap{m.IdentitySettingsMap, m.SiteSettingsMap, m.DefaultSettingsMap} {
		if value := get(*level); value != nil {
			return *value
		}
	}
	var zero T
	return zero
}

// Checked is what a check found for an identity, which the checker sends as an alarm event when it notifies or the
// identity's active or alerting alarms change
type Checked struct {
	IdentityID    int64
	Notifications []Notification
	Active        []Type
	Alerting      []Type
}

// Snoozed is an identity's snooze of some kinds of alarm until a time in milliseconds, which the checker applies
// when it is sent as an alarm_snoozed event
type Snoozed struct {
	IdentityID int64
	Types      []Type
	Until      int64
}
