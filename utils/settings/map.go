package settings

import (
	"nighttroop.t1d.nz/data/types"
	"nighttroop.t1d.nz/utils/consts"
)

type SettingMap struct {
	Units                            *Setting[consts.Units]
	Timezone                         *Setting[string]
	Language                         *Setting[string]
	DateFormat                       *Setting[string]
	TimeFormat                       *Setting[string]
	WeekStartDay                     *Setting[string]
	BolusDiaHours                    *Setting[float64]
	BolusOnsetMinutes                *Setting[float64]
	BolusPeakMinutes                 *Setting[float64]
	BasalDiaHours                    *Setting[float64]
	BasalOnsetMinutes                *Setting[float64]
	BasalPeakMinutes                 *Setting[float64]
	DefaultInsulinSensitivity        *Setting[float64]
	DefaultCarbRatio                 *Setting[float64]
	DefaultBasalRate                 *Setting[float64]
	CarbMaxAbsorptionHours           *Setting[float64]
	CarbMin5mImpact                  *Setting[float64]
	AlarmUrgentLowEnabled            *Setting[bool]
	AlarmUrgentLow                   *Setting[int64]
	AlarmLowEnabled                  *Setting[bool]
	AlarmLow                         *Setting[int64]
	AlarmLowMinutes                  *Setting[int64]
	AlarmLowPredictionEnabled        *Setting[bool]
	AlarmHighEnabled                 *Setting[bool]
	AlarmHigh                        *Setting[int64]
	AlarmHighMinutes                 *Setting[int64]
	AlarmUrgentHighEnabled           *Setting[bool]
	AlarmUrgentHigh                  *Setting[int64]
	AlarmFastRiseEnabled             *Setting[bool]
	AlarmFastRise                    *Setting[int64]
	AlarmFastRiseMinutes             *Setting[int64]
	AlarmFastDropEnabled             *Setting[bool]
	AlarmFastDrop                    *Setting[int64]
	AlarmFastDropMinutes             *Setting[int64]
	AlarmNoDataEnabled               *Setting[bool]
	AlarmNoDataMinutes               *Setting[int64]
	AlarmDeviceBatteryLowEnabled     *Setting[bool]
	AlarmDeviceBatteryLowPercent     *Setting[int64]
	AlarmDeviceBatteryLowMinutes     *Setting[int64]
	AlarmDeviceBatteryLowDevicesJson *Setting[types.DeviceNames]
	AlarmLowReservoirEnabled         *Setting[bool]
	AlarmLowReservoirUnits           *Setting[int64]
	NotificationsEnabled             *Setting[bool]
}

var DefaultSettingsMap = &SettingMap{
	Units:                            NewDefaultSetting(Units),
	Timezone:                         NewDefaultSetting(Timezone),
	Language:                         NewDefaultSetting(Language),
	DateFormat:                       NewDefaultSetting(DateFormat),
	TimeFormat:                       NewDefaultSetting(TimeFormat),
	WeekStartDay:                     NewDefaultSetting(WeekStartDay),
	BolusDiaHours:                    NewDefaultSetting(BolusDIAHours),
	BolusOnsetMinutes:                NewDefaultSetting(BolusOnsetMinutes),
	BolusPeakMinutes:                 NewDefaultSetting(BolusPeakMinutes),
	BasalDiaHours:                    NewDefaultSetting(BasalDIAHours),
	BasalOnsetMinutes:                NewDefaultSetting(BasalOnsetMinutes),
	BasalPeakMinutes:                 NewDefaultSetting(BasalPeakMinutes),
	DefaultInsulinSensitivity:        NewDefaultSetting(DefaultInsulinSensitivity),
	DefaultCarbRatio:                 NewDefaultSetting(DefaultCarbRatio),
	DefaultBasalRate:                 NewDefaultSetting(DefaultBasalRate),
	CarbMaxAbsorptionHours:           NewDefaultSetting(CarbMaxAbsorptionHours),
	CarbMin5mImpact:                  NewDefaultSetting(CarbMin5mImpact),
	AlarmUrgentLowEnabled:            NewDefaultSetting(AlarmUrgentLowEnabled),
	AlarmUrgentLow:                   NewDefaultSetting(AlarmUrgentLow),
	AlarmLowEnabled:                  NewDefaultSetting(AlarmLowEnabled),
	AlarmLow:                         NewDefaultSetting(AlarmLow),
	AlarmLowMinutes:                  NewDefaultSetting(AlarmLowMinutes),
	AlarmLowPredictionEnabled:        NewDefaultSetting(AlarmLowPredictionEnabled),
	AlarmHighEnabled:                 NewDefaultSetting(AlarmHighEnabled),
	AlarmHigh:                        NewDefaultSetting(AlarmHigh),
	AlarmHighMinutes:                 NewDefaultSetting(AlarmHighMinutes),
	AlarmUrgentHighEnabled:           NewDefaultSetting(AlarmUrgentHighEnabled),
	AlarmUrgentHigh:                  NewDefaultSetting(AlarmUrgentHigh),
	AlarmFastRiseEnabled:             NewDefaultSetting(AlarmFastRiseEnabled),
	AlarmFastRise:                    NewDefaultSetting(AlarmFastRise),
	AlarmFastRiseMinutes:             NewDefaultSetting(AlarmFastRiseMinutes),
	AlarmFastDropEnabled:             NewDefaultSetting(AlarmFastDropEnabled),
	AlarmFastDrop:                    NewDefaultSetting(AlarmFastDrop),
	AlarmFastDropMinutes:             NewDefaultSetting(AlarmFastDropMinutes),
	AlarmNoDataEnabled:               NewDefaultSetting(AlarmNoDataEnabled),
	AlarmNoDataMinutes:               NewDefaultSetting(AlarmNoDataMinutes),
	AlarmDeviceBatteryLowEnabled:     NewDefaultSetting(AlarmDeviceBatteryLowEnabled),
	AlarmDeviceBatteryLowPercent:     NewDefaultSetting(AlarmDeviceBatteryLowPercent),
	AlarmDeviceBatteryLowMinutes:     NewDefaultSetting(AlarmDeviceBatteryLowMinutes),
	AlarmDeviceBatteryLowDevicesJson: NewDefaultSetting(AlarmDeviceBatteryLowDevicesJson),
	AlarmLowReservoirEnabled:         NewDefaultSetting(AlarmLowReservoirEnabled),
	AlarmLowReservoirUnits:           NewDefaultSetting(AlarmLowReservoirUnits),
	NotificationsEnabled:             NewDefaultSetting(NotificationsEnabled),
}

// New Setting Does not validate Check
func NewSettingsMap(g SettingsGetter, createdAt *int64, subject *string) *SettingMap {
	s := &SettingMap{
		Units:                            NewSetting(Units, g.GetUnits(), createdAt, subject),
		Timezone:                         NewSetting(Timezone, g.GetTimezone(), createdAt, subject),
		Language:                         NewSetting(Language, g.GetLanguage(), createdAt, subject),
		DateFormat:                       NewSetting(DateFormat, g.GetDateFormat(), createdAt, subject),
		TimeFormat:                       NewSetting(TimeFormat, g.GetTimeFormat(), createdAt, subject),
		WeekStartDay:                     NewSetting(WeekStartDay, g.GetWeekStartDay(), createdAt, subject),
		BolusDiaHours:                    NewSetting(BolusDIAHours, g.GetBolusDiaHours(), createdAt, subject),
		BolusOnsetMinutes:                NewSetting(BolusOnsetMinutes, g.GetBolusOnsetMinutes(), createdAt, subject),
		BolusPeakMinutes:                 NewSetting(BolusPeakMinutes, g.GetBolusPeakMinutes(), createdAt, subject),
		BasalDiaHours:                    NewSetting(BasalDIAHours, g.GetBasalDiaHours(), createdAt, subject),
		BasalOnsetMinutes:                NewSetting(BasalOnsetMinutes, g.GetBasalOnsetMinutes(), createdAt, subject),
		BasalPeakMinutes:                 NewSetting(BasalPeakMinutes, g.GetBasalPeakMinutes(), createdAt, subject),
		DefaultInsulinSensitivity:        NewSetting(DefaultInsulinSensitivity, g.GetDefaultInsulinSensitivity(), createdAt, subject),
		DefaultCarbRatio:                 NewSetting(DefaultCarbRatio, g.GetDefaultCarbRatio(), createdAt, subject),
		DefaultBasalRate:                 NewSetting(DefaultBasalRate, g.GetDefaultBasalRate(), createdAt, subject),
		CarbMaxAbsorptionHours:           NewSetting(CarbMaxAbsorptionHours, g.GetCarbMaxAbsorptionHours(), createdAt, subject),
		CarbMin5mImpact:                  NewSetting(CarbMin5mImpact, g.GetCarbMin5mImpact(), createdAt, subject),
		AlarmUrgentLowEnabled:            NewSetting(AlarmUrgentLowEnabled, g.GetAlarmUrgentLowEnabled(), createdAt, subject),
		AlarmUrgentLow:                   NewSetting(AlarmUrgentLow, g.GetAlarmUrgentLow(), createdAt, subject),
		AlarmLowEnabled:                  NewSetting(AlarmLowEnabled, g.GetAlarmLowEnabled(), createdAt, subject),
		AlarmLow:                         NewSetting(AlarmLow, g.GetAlarmLow(), createdAt, subject),
		AlarmLowMinutes:                  NewSetting(AlarmLowMinutes, g.GetAlarmLowMinutes(), createdAt, subject),
		AlarmLowPredictionEnabled:        NewSetting(AlarmLowPredictionEnabled, g.GetAlarmLowPredictionEnabled(), createdAt, subject),
		AlarmHighEnabled:                 NewSetting(AlarmHighEnabled, g.GetAlarmHighEnabled(), createdAt, subject),
		AlarmHigh:                        NewSetting(AlarmHigh, g.GetAlarmHigh(), createdAt, subject),
		AlarmHighMinutes:                 NewSetting(AlarmHighMinutes, g.GetAlarmHighMinutes(), createdAt, subject),
		AlarmUrgentHighEnabled:           NewSetting(AlarmUrgentHighEnabled, g.GetAlarmUrgentHighEnabled(), createdAt, subject),
		AlarmUrgentHigh:                  NewSetting(AlarmUrgentHigh, g.GetAlarmUrgentHigh(), createdAt, subject),
		AlarmFastRiseEnabled:             NewSetting(AlarmFastRiseEnabled, g.GetAlarmFastRiseEnabled(), createdAt, subject),
		AlarmFastRise:                    NewSetting(AlarmFastRise, g.GetAlarmFastRise(), createdAt, subject),
		AlarmFastRiseMinutes:             NewSetting(AlarmFastRiseMinutes, g.GetAlarmFastRiseMinutes(), createdAt, subject),
		AlarmFastDropEnabled:             NewSetting(AlarmFastDropEnabled, g.GetAlarmFastDropEnabled(), createdAt, subject),
		AlarmFastDrop:                    NewSetting(AlarmFastDrop, g.GetAlarmFastDrop(), createdAt, subject),
		AlarmFastDropMinutes:             NewSetting(AlarmFastDropMinutes, g.GetAlarmFastDropMinutes(), createdAt, subject),
		AlarmNoDataEnabled:               NewSetting(AlarmNoDataEnabled, g.GetAlarmNoDataEnabled(), createdAt, subject),
		AlarmNoDataMinutes:               NewSetting(AlarmNoDataMinutes, g.GetAlarmNoDataMinutes(), createdAt, subject),
		AlarmDeviceBatteryLowEnabled:     NewSetting(AlarmDeviceBatteryLowEnabled, g.GetAlarmDeviceBatteryLowEnabled(), createdAt, subject),
		AlarmDeviceBatteryLowPercent:     NewSetting(AlarmDeviceBatteryLowPercent, g.GetAlarmDeviceBatteryLowPercent(), createdAt, subject),
		AlarmDeviceBatteryLowMinutes:     NewSetting(AlarmDeviceBatteryLowMinutes, g.GetAlarmDeviceBatteryLowMinutes(), createdAt, subject),
		AlarmDeviceBatteryLowDevicesJson: NewSetting(AlarmDeviceBatteryLowDevicesJson, g.GetAlarmDeviceBatteryLowDevicesJson(), createdAt, subject),
		AlarmLowReservoirEnabled:         NewSetting(AlarmLowReservoirEnabled, g.GetAlarmLowReservoirEnabled(), createdAt, subject),
		AlarmLowReservoirUnits:           NewSetting(AlarmLowReservoirUnits, g.GetAlarmLowReservoirUnits(), createdAt, subject),
		NotificationsEnabled:             NewSetting(NotificationsEnabled, g.GetNotificationsEnabled(), createdAt, subject),
	}

	return s
}

// GetUnits returns the Units value.
func (m SettingMap) GetUnits() *consts.Units {
	if m.Units == nil {
		return nil
	}
	return ptr(m.Units.Value)
}

// SetUnits sets the Units value and returns the updated SettingMap.
func (m *SettingMap) SetUnits(v *consts.Units) {
	m.Units = NewSetting(Units, v, nil, nil)
}

// GetTimezone returns the Timezone value.
func (m SettingMap) GetTimezone() *string {
	if m.Timezone == nil {
		return nil
	}
	return ptr(m.Timezone.Value)
}

// SetTimezone sets the Timezone value and returns the updated SettingMap.
func (m *SettingMap) SetTimezone(v *string) {
	m.Timezone = NewSetting(Timezone, v, nil, nil)
}

// GetLanguage returns the Language value.
func (m SettingMap) GetLanguage() *string {
	if m.Language == nil {
		return nil
	}
	return ptr(m.Language.Value)
}

// SetLanguage sets the Language value and returns the updated SettingMap.
func (m *SettingMap) SetLanguage(v *string) {

	m.Language = NewSetting(Language, v, nil, nil)

}

// GetDateFormat returns the DateFormat value.
func (m SettingMap) GetDateFormat() *string {
	if m.DateFormat == nil {
		return nil
	}
	return ptr(m.DateFormat.Value)
}

// SetDateFormat sets the DateFormat value and returns the updated SettingMap.
func (m *SettingMap) SetDateFormat(v *string) {

	m.DateFormat = NewSetting(DateFormat, v, nil, nil)

}

// GetTimeFormat returns the TimeFormat value.
func (m SettingMap) GetTimeFormat() *string {
	if m.TimeFormat == nil {
		return nil
	}
	return ptr(m.TimeFormat.Value)
}

// SetTimeFormat sets the TimeFormat value and returns the updated SettingMap.
func (m *SettingMap) SetTimeFormat(v *string) {

	m.TimeFormat = NewSetting(TimeFormat, v, nil, nil)

}

// GetWeekStartDay returns the WeekStartDay value.
func (m SettingMap) GetWeekStartDay() *string {
	if m.WeekStartDay == nil {
		return nil
	}
	return ptr(m.WeekStartDay.Value)
}

// SetWeekStartDay sets the WeekStartDay value and returns the updated SettingMap.
func (m *SettingMap) SetWeekStartDay(v *string) {

	m.WeekStartDay = NewSetting(WeekStartDay, v, nil, nil)

}

// GetBolusDiaHours returns the BolusDiaHours value.
func (m SettingMap) GetBolusDiaHours() *float64 {
	if m.BolusDiaHours == nil {
		return nil
	}
	return ptr(m.BolusDiaHours.Value)
}

// SetBolusDiaHours sets the BolusDiaHours value and returns the updated SettingMap.
func (m *SettingMap) SetBolusDiaHours(v *float64) {

	m.BolusDiaHours = NewSetting(BolusDIAHours, v, nil, nil)

}

// GetBolusOnsetMinutes returns the BolusOnsetMinutes value.
func (m SettingMap) GetBolusOnsetMinutes() *float64 {
	if m.BolusOnsetMinutes == nil {
		return nil
	}
	return ptr(m.BolusOnsetMinutes.Value)
}

// SetBolusOnsetMinutes sets the BolusOnsetMinutes value and returns the updated SettingMap.
func (m *SettingMap) SetBolusOnsetMinutes(v *float64) {
	m.BolusOnsetMinutes = NewSetting(BolusOnsetMinutes, v, nil, nil)
}

// GetBolusPeakMinutes returns the BolusPeakMinutes value.
func (m SettingMap) GetBolusPeakMinutes() *float64 {
	if m.BolusPeakMinutes == nil {
		return nil
	}
	return ptr(m.BolusPeakMinutes.Value)
}

// SetBolusPeakMinutes sets the BolusPeakMinutes value and returns the updated SettingMap.
func (m *SettingMap) SetBolusPeakMinutes(v *float64) {

	m.BolusPeakMinutes = NewSetting(BolusPeakMinutes, v, nil, nil)

}

// GetBasalDiaHours returns the BasalDiaHours value.
func (m SettingMap) GetBasalDiaHours() *float64 {
	if m.BasalDiaHours == nil {
		return nil
	}
	return ptr(m.BasalDiaHours.Value)
}

// SetBasalDiaHours sets the BasalDiaHours value and returns the updated SettingMap.
func (m *SettingMap) SetBasalDiaHours(v *float64) {

	m.BasalDiaHours = NewSetting(BasalDIAHours, v, nil, nil)

}

// GetBasalOnsetMinutes returns the BasalOnsetMinutes value.
func (m SettingMap) GetBasalOnsetMinutes() *float64 {
	if m.BasalOnsetMinutes == nil {
		return nil
	}
	return ptr(m.BasalOnsetMinutes.Value)
}

// SetBasalOnsetMinutes sets the BasalOnsetMinutes value and returns the updated SettingMap.
func (m *SettingMap) SetBasalOnsetMinutes(v *float64) {

	m.BasalOnsetMinutes = NewSetting(BasalOnsetMinutes, v, nil, nil)

}

// GetBasalPeakMinutes returns the BasalPeakMinutes value.
func (m SettingMap) GetBasalPeakMinutes() *float64 {
	if m.BasalPeakMinutes == nil {
		return nil
	}
	return ptr(m.BasalPeakMinutes.Value)
}

// SetBasalPeakMinutes sets the BasalPeakMinutes value and returns the updated SettingMap.
func (m *SettingMap) SetBasalPeakMinutes(v *float64) {

	m.BasalPeakMinutes = NewSetting(BasalPeakMinutes, v, nil, nil)

}

// GetDefaultInsulinSensitivity returns the DefaultInsulinSensitivity value.
func (m SettingMap) GetDefaultInsulinSensitivity() *float64 {
	if m.DefaultInsulinSensitivity == nil {
		return nil
	}
	return ptr(m.DefaultInsulinSensitivity.Value)
}

// SetDefaultInsulinSensitivity sets the DefaultInsulinSensitivity value and returns the updated SettingMap.
func (m *SettingMap) SetDefaultInsulinSensitivity(v *float64) {

	m.DefaultInsulinSensitivity = NewSetting(DefaultInsulinSensitivity, v, nil, nil)

}

// GetDefaultCarbRatio returns the DefaultCarbRatio value.
func (m SettingMap) GetDefaultCarbRatio() *float64 {
	if m.DefaultCarbRatio == nil {
		return nil
	}
	return ptr(m.DefaultCarbRatio.Value)
}

// SetDefaultCarbRatio sets the DefaultCarbRatio value and returns the updated SettingMap.
func (m *SettingMap) SetDefaultCarbRatio(v *float64) {

	m.DefaultCarbRatio = NewSetting(DefaultCarbRatio, v, nil, nil)

}

// GetDefaultBasalRate returns the DefaultBasalRate value.
func (m SettingMap) GetDefaultBasalRate() *float64 {
	if m.DefaultBasalRate == nil {
		return nil
	}
	return ptr(m.DefaultBasalRate.Value)
}

// SetDefaultBasalRate sets the DefaultBasalRate value and returns the updated SettingMap.
func (m *SettingMap) SetDefaultBasalRate(v *float64) {

	m.DefaultBasalRate = NewSetting(DefaultBasalRate, v, nil, nil)

}

// GetCarbMaxAbsorptionHours returns the CarbMaxAbsorptionHours value.
func (m SettingMap) GetCarbMaxAbsorptionHours() *float64 {
	if m.CarbMaxAbsorptionHours == nil {
		return nil
	}
	return ptr(m.CarbMaxAbsorptionHours.Value)
}

// SetCarbMaxAbsorptionHours sets the CarbMaxAbsorptionHours value and returns the updated SettingMap.
func (m *SettingMap) SetCarbMaxAbsorptionHours(v *float64) {

	m.CarbMaxAbsorptionHours = NewSetting(CarbMaxAbsorptionHours, v, nil, nil)

}

// GetCarbMin5mImpact returns the CarbMin5mImpact value.
func (m SettingMap) GetCarbMin5mImpact() *float64 {
	if m.CarbMin5mImpact == nil {
		return nil
	}
	return ptr(m.CarbMin5mImpact.Value)
}

// SetCarbMin5mImpact sets the CarbMin5mImpact value and returns the updated SettingMap.
func (m *SettingMap) SetCarbMin5mImpact(v *float64) {

	m.CarbMin5mImpact = NewSetting(CarbMin5mImpact, v, nil, nil)

}

// GetAlarmUrgentLowEnabled returns the AlarmUrgentLowEnabled value.
func (m SettingMap) GetAlarmUrgentLowEnabled() *bool {
	if m.AlarmUrgentLowEnabled == nil {
		return nil
	}
	return ptr(m.AlarmUrgentLowEnabled.Value)
}

// SetAlarmUrgentLowEnabled sets the AlarmUrgentLowEnabled value and returns the updated SettingMap.
func (m *SettingMap) SetAlarmUrgentLowEnabled(v *bool) {

	m.AlarmUrgentLowEnabled = NewSetting(AlarmUrgentLowEnabled, v, nil, nil)

}

// GetAlarmUrgentLow returns the AlarmUrgentLow value.
func (m SettingMap) GetAlarmUrgentLow() *int64 {
	if m.AlarmUrgentLow == nil {
		return nil
	}
	return ptr(m.AlarmUrgentLow.Value)
}

// SetAlarmUrgentLow sets the AlarmUrgentLow value and returns the updated SettingMap.
func (m *SettingMap) SetAlarmUrgentLow(v *int64) {

	m.AlarmUrgentLow = NewSetting(AlarmUrgentLow, v, nil, nil)

}

// GetAlarmLowEnabled returns the AlarmLowEnabled value.
func (m SettingMap) GetAlarmLowEnabled() *bool {
	if m.AlarmLowEnabled == nil {
		return nil
	}
	return ptr(m.AlarmLowEnabled.Value)
}

// SetAlarmLowEnabled sets the AlarmLowEnabled value and returns the updated SettingMap.
func (m *SettingMap) SetAlarmLowEnabled(v *bool) {

	m.AlarmLowEnabled = NewSetting(AlarmLowEnabled, v, nil, nil)

}

// GetAlarmLow returns the AlarmLow value.
func (m SettingMap) GetAlarmLow() *int64 {
	if m.AlarmLow == nil {
		return nil
	}
	return ptr(m.AlarmLow.Value)
}

// SetAlarmLow sets the AlarmLow value and returns the updated SettingMap.
func (m *SettingMap) SetAlarmLow(v *int64) {

	m.AlarmLow = NewSetting(AlarmLow, v, nil, nil)

}

// GetAlarmLowMinutes returns the AlarmLowMinutes value.
func (m SettingMap) GetAlarmLowMinutes() *int64 {
	if m.AlarmLowMinutes == nil {
		return nil
	}
	return ptr(m.AlarmLowMinutes.Value)
}

// SetAlarmLowMinutes sets the AlarmLowMinutes value and returns the updated SettingMap.
func (m *SettingMap) SetAlarmLowMinutes(v *int64) {

	m.AlarmLowMinutes = NewSetting(AlarmLowMinutes, v, nil, nil)

}

// GetAlarmLowPredictionEnabled returns the AlarmLowPredictionEnabled value.
func (m SettingMap) GetAlarmLowPredictionEnabled() *bool {
	if m.AlarmLowPredictionEnabled == nil {
		return nil
	}
	return ptr(m.AlarmLowPredictionEnabled.Value)
}

// SetAlarmLowPredictionEnabled sets the AlarmLowPredictionEnabled value and returns the updated SettingMap.
func (m *SettingMap) SetAlarmLowPredictionEnabled(v *bool) {

	m.AlarmLowPredictionEnabled = NewSetting(AlarmLowPredictionEnabled, v, nil, nil)

}

// GetAlarmHighEnabled returns the AlarmHighEnabled value.
func (m SettingMap) GetAlarmHighEnabled() *bool {
	if m.AlarmHighEnabled == nil {
		return nil
	}
	return ptr(m.AlarmHighEnabled.Value)
}

// SetAlarmHighEnabled sets the AlarmHighEnabled value and returns the updated SettingMap.
func (m *SettingMap) SetAlarmHighEnabled(v *bool) {

	m.AlarmHighEnabled = NewSetting(AlarmHighEnabled, v, nil, nil)

}

// GetAlarmHigh returns the AlarmHigh value.
func (m SettingMap) GetAlarmHigh() *int64 {
	if m.AlarmHigh == nil {
		return nil
	}
	return ptr(m.AlarmHigh.Value)
}

// SetAlarmHigh sets the AlarmHigh value and returns the updated SettingMap.
func (m *SettingMap) SetAlarmHigh(v *int64) {

	m.AlarmHigh = NewSetting(AlarmHigh, v, nil, nil)

}

// GetAlarmHighMinutes returns the AlarmHighMinutes value.
func (m SettingMap) GetAlarmHighMinutes() *int64 {
	if m.AlarmHighMinutes == nil {
		return nil
	}
	return ptr(m.AlarmHighMinutes.Value)
}

// SetAlarmHighMinutes sets the AlarmHighMinutes value and returns the updated SettingMap.
func (m *SettingMap) SetAlarmHighMinutes(v *int64) {

	m.AlarmHighMinutes = NewSetting(AlarmHighMinutes, v, nil, nil)

}

// GetAlarmUrgentHighEnabled returns the AlarmUrgentHighEnabled value.
func (m SettingMap) GetAlarmUrgentHighEnabled() *bool {
	if m.AlarmUrgentHighEnabled == nil {
		return nil
	}
	return ptr(m.AlarmUrgentHighEnabled.Value)
}

// SetAlarmUrgentHighEnabled sets the AlarmUrgentHighEnabled value and returns the updated SettingMap.
func (m *SettingMap) SetAlarmUrgentHighEnabled(v *bool) {

	m.AlarmUrgentHighEnabled = NewSetting(AlarmUrgentHighEnabled, v, nil, nil)

}

// GetAlarmUrgentHigh returns the AlarmUrgentHigh value.
func (m SettingMap) GetAlarmUrgentHigh() *int64 {
	if m.AlarmUrgentHigh == nil {
		return nil
	}
	return ptr(m.AlarmUrgentHigh.Value)
}

// SetAlarmUrgentHigh sets the AlarmUrgentHigh value and returns the updated SettingMap.
func (m *SettingMap) SetAlarmUrgentHigh(v *int64) {

	m.AlarmUrgentHigh = NewSetting(AlarmUrgentHigh, v, nil, nil)

}

// GetAlarmFastRiseEnabled returns the AlarmFastRiseEnabled value.
func (m SettingMap) GetAlarmFastRiseEnabled() *bool {
	if m.AlarmFastRiseEnabled == nil {
		return nil
	}
	return ptr(m.AlarmFastRiseEnabled.Value)
}

// SetAlarmFastRiseEnabled sets the AlarmFastRiseEnabled value and returns the updated SettingMap.
func (m *SettingMap) SetAlarmFastRiseEnabled(v *bool) {

	m.AlarmFastRiseEnabled = NewSetting(AlarmFastRiseEnabled, v, nil, nil)

}

// GetAlarmFastRise returns the AlarmFastRise value.
func (m SettingMap) GetAlarmFastRise() *int64 {
	if m.AlarmFastRise == nil {
		return nil
	}
	return ptr(m.AlarmFastRise.Value)
}

// SetAlarmFastRise sets the AlarmFastRise value and returns the updated SettingMap.
func (m *SettingMap) SetAlarmFastRise(v *int64) {

	m.AlarmFastRise = NewSetting(AlarmFastRise, v, nil, nil)

}

// GetAlarmFastRiseMinutes returns the AlarmFastRiseMinutes value.
func (m SettingMap) GetAlarmFastRiseMinutes() *int64 {
	if m.AlarmFastRiseMinutes == nil {
		return nil
	}
	return ptr(m.AlarmFastRiseMinutes.Value)
}

// SetAlarmFastRiseMinutes sets the AlarmFastRiseMinutes value and returns the updated SettingMap.
func (m *SettingMap) SetAlarmFastRiseMinutes(v *int64) {

	m.AlarmFastRiseMinutes = NewSetting(AlarmFastRiseMinutes, v, nil, nil)

}

// GetAlarmFastDropEnabled returns the AlarmFastDropEnabled value.
func (m SettingMap) GetAlarmFastDropEnabled() *bool {
	if m.AlarmFastDropEnabled == nil {
		return nil
	}
	return ptr(m.AlarmFastDropEnabled.Value)
}

// SetAlarmFastDropEnabled sets the AlarmFastDropEnabled value and returns the updated SettingMap.
func (m *SettingMap) SetAlarmFastDropEnabled(v *bool) {

	m.AlarmFastDropEnabled = NewSetting(AlarmFastDropEnabled, v, nil, nil)

}

// GetAlarmFastDrop returns the AlarmFastDrop value.
func (m SettingMap) GetAlarmFastDrop() *int64 {
	if m.AlarmFastDrop == nil {
		return nil
	}
	return ptr(m.AlarmFastDrop.Value)
}

// SetAlarmFastDrop sets the AlarmFastDrop value and returns the updated SettingMap.
func (m *SettingMap) SetAlarmFastDrop(v *int64) {

	m.AlarmFastDrop = NewSetting(AlarmFastDrop, v, nil, nil)

}

// GetAlarmFastDropMinutes returns the AlarmFastDropMinutes value.
func (m SettingMap) GetAlarmFastDropMinutes() *int64 {
	if m.AlarmFastDropMinutes == nil {
		return nil
	}
	return ptr(m.AlarmFastDropMinutes.Value)
}

// SetAlarmFastDropMinutes sets the AlarmFastDropMinutes value and returns the updated SettingMap.
func (m *SettingMap) SetAlarmFastDropMinutes(v *int64) {

	m.AlarmFastDropMinutes = NewSetting(AlarmFastDropMinutes, v, nil, nil)

}

// GetAlarmNoDataEnabled returns the AlarmNoDataEnabled value.
func (m SettingMap) GetAlarmNoDataEnabled() *bool {
	if m.AlarmNoDataEnabled == nil {
		return nil
	}
	return ptr(m.AlarmNoDataEnabled.Value)
}

// SetAlarmNoDataEnabled sets the AlarmNoDataEnabled value and returns the updated SettingMap.
func (m *SettingMap) SetAlarmNoDataEnabled(v *bool) {

	m.AlarmNoDataEnabled = NewSetting(AlarmNoDataEnabled, v, nil, nil)

}

// GetAlarmNoDataMinutes returns the AlarmNoDataMinutes value.
func (m SettingMap) GetAlarmNoDataMinutes() *int64 {
	if m.AlarmNoDataMinutes == nil {
		return nil
	}
	return ptr(m.AlarmNoDataMinutes.Value)
}

// SetAlarmNoDataMinutes sets the AlarmNoDataMinutes value and returns the updated SettingMap.
func (m *SettingMap) SetAlarmNoDataMinutes(v *int64) {

	m.AlarmNoDataMinutes = NewSetting(AlarmNoDataMinutes, v, nil, nil)

}

// GetAlarmDeviceBatteryLowEnabled returns the AlarmDeviceBatteryLowEnabled value.
func (m SettingMap) GetAlarmDeviceBatteryLowEnabled() *bool {
	if m.AlarmDeviceBatteryLowEnabled == nil {
		return nil
	}
	return ptr(m.AlarmDeviceBatteryLowEnabled.Value)
}

// SetAlarmDeviceBatteryLowEnabled sets the AlarmDeviceBatteryLowEnabled value and returns the updated SettingMap.
func (m *SettingMap) SetAlarmDeviceBatteryLowEnabled(v *bool) {

	m.AlarmDeviceBatteryLowEnabled = NewSetting(AlarmDeviceBatteryLowEnabled, v, nil, nil)

}

// GetAlarmDeviceBatteryLowPercent returns the AlarmDeviceBatteryLowPercent value.
func (m SettingMap) GetAlarmDeviceBatteryLowPercent() *int64 {
	if m.AlarmDeviceBatteryLowPercent == nil {
		return nil
	}
	return ptr(m.AlarmDeviceBatteryLowPercent.Value)
}

// SetAlarmDeviceBatteryLowPercent sets the AlarmDeviceBatteryLowPercent value and returns the updated SettingMap.
func (m *SettingMap) SetAlarmDeviceBatteryLowPercent(v *int64) {

	m.AlarmDeviceBatteryLowPercent = NewSetting(AlarmDeviceBatteryLowPercent, v, nil, nil)

}

// GetAlarmDeviceBatteryLowMinutes returns the AlarmDeviceBatteryLowMinutes value.
func (m SettingMap) GetAlarmDeviceBatteryLowMinutes() *int64 {
	if m.AlarmDeviceBatteryLowMinutes == nil {
		return nil
	}
	return ptr(m.AlarmDeviceBatteryLowMinutes.Value)
}

// SetAlarmDeviceBatteryLowMinutes sets the AlarmDeviceBatteryLowMinutes value and returns the updated SettingMap.
func (m *SettingMap) SetAlarmDeviceBatteryLowMinutes(v *int64) {

	m.AlarmDeviceBatteryLowMinutes = NewSetting(AlarmDeviceBatteryLowMinutes, v, nil, nil)

}

// GetAlarmDeviceBatteryLowDevicesJson returns the AlarmDeviceBatteryLowDevicesJson value.
func (m SettingMap) GetAlarmDeviceBatteryLowDevicesJson() *types.DeviceNames {
	if m.AlarmDeviceBatteryLowDevicesJson == nil {
		return nil
	}
	return ptr(m.AlarmDeviceBatteryLowDevicesJson.Value)
}

// SetAlarmDeviceBatteryLowDevicesJson sets the AlarmDeviceBatteryLowDevicesJson value and returns the updated SettingMap.
func (m *SettingMap) SetAlarmDeviceBatteryLowDevicesJson(v *types.DeviceNames) {

	m.AlarmDeviceBatteryLowDevicesJson = NewSetting(AlarmDeviceBatteryLowDevicesJson, v, nil, nil)

}

// GetAlarmLowReservoirEnabled returns the AlarmLowReservoirEnabled value.
func (m SettingMap) GetAlarmLowReservoirEnabled() *bool {
	if m.AlarmLowReservoirEnabled == nil {
		return nil
	}
	return ptr(m.AlarmLowReservoirEnabled.Value)
}

// SetAlarmLowReservoirEnabled sets the AlarmLowReservoirEnabled value and returns the updated SettingMap.
func (m *SettingMap) SetAlarmLowReservoirEnabled(v *bool) {

	m.AlarmLowReservoirEnabled = NewSetting(AlarmLowReservoirEnabled, v, nil, nil)

}

// GetAlarmLowReservoirUnits returns the AlarmLowReservoirUnits value.
func (m SettingMap) GetAlarmLowReservoirUnits() *int64 {
	if m.AlarmLowReservoirUnits == nil {
		return nil
	}
	return ptr(m.AlarmLowReservoirUnits.Value)

}

// SetAlarmLowReservoirUnits sets the AlarmLowReservoirUnits value and returns the updated SettingMap.
func (m *SettingMap) SetAlarmLowReservoirUnits(v *int64) {

	m.AlarmLowReservoirUnits = NewSetting(AlarmLowReservoirUnits, v, nil, nil)

}

// GetNotificationsEnabled returns the NotificationsEnabled value.
func (m SettingMap) GetNotificationsEnabled() *bool {
	if m.NotificationsEnabled == nil {
		return nil
	}
	return ptr(m.NotificationsEnabled.Value)
}

// SetNotificationsEnabled sets the NotificationsEnabled value and returns the updated SettingMap.
func (m *SettingMap) SetNotificationsEnabled(v *bool) {

	m.NotificationsEnabled = NewSetting(NotificationsEnabled, v, nil, nil)

}

func (m SettingMap) GetSettingByName(name string) any {
	return GetSettingByName(name, m)
}

func (m SettingMap) GetSettingAnyTypeByName(name string) *SettingAnyType {
	switch name {
	case "units":
		return m.Units.ToSettingAnyType()
	case "timezone":
		return m.Timezone.ToSettingAnyType()
	case "language":
		return m.Language.ToSettingAnyType()
	case "date_format":
		return m.DateFormat.ToSettingAnyType()
	case "time_format":
		return m.TimeFormat.ToSettingAnyType()
	case "week_start_day":
		return m.WeekStartDay.ToSettingAnyType()
	case "bolus_dia_hours":
		return m.BolusDiaHours.ToSettingAnyType()
	case "bolus_onset_minutes":
		return m.BolusOnsetMinutes.ToSettingAnyType()
	case "bolus_peak_minutes":
		return m.BolusPeakMinutes.ToSettingAnyType()
	case "basal_dia_hours":
		return m.BasalDiaHours.ToSettingAnyType()
	case "basal_onset_minutes":
		return m.BasalOnsetMinutes.ToSettingAnyType()
	case "basal_peak_minutes":
		return m.BasalPeakMinutes.ToSettingAnyType()
	case "default_insulin_sensitivity":
		return m.DefaultInsulinSensitivity.ToSettingAnyType()
	case "default_carb_ratio":
		return m.DefaultCarbRatio.ToSettingAnyType()
	case "default_basal_rate":
		return m.DefaultBasalRate.ToSettingAnyType()
	case "carb_max_absorption_hours":
		return m.CarbMaxAbsorptionHours.ToSettingAnyType()
	case "carb_min_5m_impact":
		return m.CarbMin5mImpact.ToSettingAnyType()
	case "alarm_urgent_low_enabled":
		return m.AlarmUrgentLowEnabled.ToSettingAnyType()
	case "alarm_urgent_low":
		return m.AlarmUrgentLow.ToSettingAnyType()
	case "alarm_low_enabled":
		return m.AlarmLowEnabled.ToSettingAnyType()
	case "alarm_low":
		return m.AlarmLow.ToSettingAnyType()
	case "alarm_low_minutes":
		return m.AlarmLowMinutes.ToSettingAnyType()
	case "alarm_low_prediction_enabled":
		return m.AlarmLowPredictionEnabled.ToSettingAnyType()
	case "alarm_high_enabled":
		return m.AlarmHighEnabled.ToSettingAnyType()
	case "alarm_high":
		return m.AlarmHigh.ToSettingAnyType()
	case "alarm_high_minutes":
		return m.AlarmHighMinutes.ToSettingAnyType()
	case "alarm_urgent_high_enabled":
		return m.AlarmUrgentHighEnabled.ToSettingAnyType()
	case "alarm_urgent_high":
		return m.AlarmUrgentHigh.ToSettingAnyType()
	case "alarm_fast_rise_enabled":
		return m.AlarmFastRiseEnabled.ToSettingAnyType()
	case "alarm_fast_rise":
		return m.AlarmFastRise.ToSettingAnyType()
	case "alarm_fast_rise_minutes":
		return m.AlarmFastRiseMinutes.ToSettingAnyType()
	case "alarm_fast_drop_enabled":
		return m.AlarmFastDropEnabled.ToSettingAnyType()
	case "alarm_fast_drop":
		return m.AlarmFastDrop.ToSettingAnyType()
	case "alarm_fast_drop_minutes":
		return m.AlarmFastDropMinutes.ToSettingAnyType()
	case "alarm_no_data_enabled":
		return m.AlarmNoDataEnabled.ToSettingAnyType()
	case "alarm_no_data_minutes":
		return m.AlarmNoDataMinutes.ToSettingAnyType()
	case "alarm_device_battery_low_enabled":
		return m.AlarmDeviceBatteryLowEnabled.ToSettingAnyType()
	case "alarm_device_battery_low_percent":
		return m.AlarmDeviceBatteryLowPercent.ToSettingAnyType()
	case "alarm_device_battery_low_minutes":
		return m.AlarmDeviceBatteryLowMinutes.ToSettingAnyType()
	case "alarm_device_battery_low_devices":
		return m.AlarmDeviceBatteryLowDevicesJson.ToSettingAnyType()
	case "alarm_low_reservoir_enabled":
		return m.AlarmLowReservoirEnabled.ToSettingAnyType()
	case "alarm_low_reservoir_units":
		return m.AlarmLowReservoirUnits.ToSettingAnyType()
	case "notifications_enabled":
		return m.NotificationsEnabled.ToSettingAnyType()
	default:
		return nil
	}
}

func (m *SettingMap) SetSettingByName(name string, v any, createdAt *int64, subject *string) error {
	var err error
	switch name {
	case "units":
		m.Units, err = NewSettingAndParseValidate(Units, v, createdAt, subject)
		return err
	case "timezone":
		m.Timezone, err = NewSettingAndParseValidate(Timezone, v, createdAt, subject)
		return err
	case "language":
		m.Language, err = NewSettingAndParseValidate(Language, v, createdAt, subject)
		return err
	case "date_format":
		m.DateFormat, err = NewSettingAndParseValidate(DateFormat, v, createdAt, subject)
		return err
	case "time_format":
		m.TimeFormat, err = NewSettingAndParseValidate(TimeFormat, v, createdAt, subject)
		return err
	case "week_start_day":
		m.WeekStartDay, err = NewSettingAndParseValidate(WeekStartDay, v, createdAt, subject)
		return err
	case "bolus_dia_hours":
		m.BolusDiaHours, err = NewSettingAndParseValidate(BolusDIAHours, v, createdAt, subject)
		return err
	case "bolus_onset_minutes":
		m.BolusOnsetMinutes, err = NewSettingAndParseValidate(BolusOnsetMinutes, v, createdAt, subject)
		return err
	case "bolus_peak_minutes":
		m.BolusPeakMinutes, err = NewSettingAndParseValidate(BolusPeakMinutes, v, createdAt, subject)
		return err
	case "basal_dia_hours":
		m.BasalDiaHours, err = NewSettingAndParseValidate(BasalDIAHours, v, createdAt, subject)
		return err
	case "basal_onset_minutes":
		m.BasalOnsetMinutes, err = NewSettingAndParseValidate(BasalOnsetMinutes, v, createdAt, subject)
		return err
	case "basal_peak_minutes":
		m.BasalPeakMinutes, err = NewSettingAndParseValidate(BasalPeakMinutes, v, createdAt, subject)
		return err
	case "default_insulin_sensitivity":
		m.DefaultInsulinSensitivity, err = NewSettingAndParseValidate(DefaultInsulinSensitivity, v, createdAt, subject)
		return err
	case "default_carb_ratio":
		m.DefaultCarbRatio, err = NewSettingAndParseValidate(DefaultCarbRatio, v, createdAt, subject)
		return err
	case "default_basal_rate":
		m.DefaultBasalRate, err = NewSettingAndParseValidate(DefaultBasalRate, v, createdAt, subject)
		return err
	case "carb_max_absorption_hours":
		m.CarbMaxAbsorptionHours, err = NewSettingAndParseValidate(CarbMaxAbsorptionHours, v, createdAt, subject)
		return err
	case "carb_min_5m_impact":
		m.CarbMin5mImpact, err = NewSettingAndParseValidate(CarbMin5mImpact, v, createdAt, subject)
		return err
	case "alarm_urgent_low_enabled":
		m.AlarmUrgentLowEnabled, err = NewSettingAndParseValidate(AlarmUrgentLowEnabled, v, createdAt, subject)
		return err
	case "alarm_urgent_low":
		m.AlarmUrgentLow, err = NewSettingAndParseValidate(AlarmUrgentLow, v, createdAt, subject)
		return err
	case "alarm_low_enabled":
		m.AlarmLowEnabled, err = NewSettingAndParseValidate(AlarmLowEnabled, v, createdAt, subject)
		return err
	case "alarm_low":
		m.AlarmLow, err = NewSettingAndParseValidate(AlarmLow, v, createdAt, subject)
		return err
	case "alarm_low_minutes":
		m.AlarmLowMinutes, err = NewSettingAndParseValidate(AlarmLowMinutes, v, createdAt, subject)
		return err
	case "alarm_low_prediction_enabled":
		m.AlarmLowPredictionEnabled, err = NewSettingAndParseValidate(AlarmLowPredictionEnabled, v, createdAt, subject)
		return err
	case "alarm_high_enabled":
		m.AlarmHighEnabled, err = NewSettingAndParseValidate(AlarmHighEnabled, v, createdAt, subject)
		return err
	case "alarm_high":
		m.AlarmHigh, err = NewSettingAndParseValidate(AlarmHigh, v, createdAt, subject)
		return err
	case "alarm_high_minutes":
		m.AlarmHighMinutes, err = NewSettingAndParseValidate(AlarmHighMinutes, v, createdAt, subject)
		return err
	case "alarm_urgent_high_enabled":
		m.AlarmUrgentHighEnabled, err = NewSettingAndParseValidate(AlarmUrgentHighEnabled, v, createdAt, subject)
		return err
	case "alarm_urgent_high":
		m.AlarmUrgentHigh, err = NewSettingAndParseValidate(AlarmUrgentHigh, v, createdAt, subject)
		return err
	case "alarm_fast_rise_enabled":
		m.AlarmFastRiseEnabled, err = NewSettingAndParseValidate(AlarmFastRiseEnabled, v, createdAt, subject)
		return err
	case "alarm_fast_rise":
		m.AlarmFastRise, err = NewSettingAndParseValidate(AlarmFastRise, v, createdAt, subject)
		return err
	case "alarm_fast_rise_minutes":
		m.AlarmFastRiseMinutes, err = NewSettingAndParseValidate(AlarmFastRiseMinutes, v, createdAt, subject)
		return err
	case "alarm_fast_drop_enabled":
		m.AlarmFastDropEnabled, err = NewSettingAndParseValidate(AlarmFastDropEnabled, v, createdAt, subject)
		return err
	case "alarm_fast_drop":
		m.AlarmFastDrop, err = NewSettingAndParseValidate(AlarmFastDrop, v, createdAt, subject)
		return err
	case "alarm_fast_drop_minutes":
		m.AlarmFastDropMinutes, err = NewSettingAndParseValidate(AlarmFastDropMinutes, v, createdAt, subject)
		return err
	case "alarm_no_data_enabled":
		m.AlarmNoDataEnabled, err = NewSettingAndParseValidate(AlarmNoDataEnabled, v, createdAt, subject)
		return err
	case "alarm_no_data_minutes":
		m.AlarmNoDataMinutes, err = NewSettingAndParseValidate(AlarmNoDataMinutes, v, createdAt, subject)
		return err
	case "alarm_device_battery_low_enabled":
		m.AlarmDeviceBatteryLowEnabled, err = NewSettingAndParseValidate(AlarmDeviceBatteryLowEnabled, v, createdAt, subject)
		return err
	case "alarm_device_battery_low_percent":
		m.AlarmDeviceBatteryLowPercent, err = NewSettingAndParseValidate(AlarmDeviceBatteryLowPercent, v, createdAt, subject)
		return err
	case "alarm_device_battery_low_minutes":
		m.AlarmDeviceBatteryLowMinutes, err = NewSettingAndParseValidate(AlarmDeviceBatteryLowMinutes, v, createdAt, subject)
		return err
	case "alarm_device_battery_low_devices":
		m.AlarmDeviceBatteryLowDevicesJson, err = NewSettingAndParseValidate(AlarmDeviceBatteryLowDevicesJson, v, createdAt, subject)
		return err
	case "alarm_low_reservoir_enabled":
		m.AlarmLowReservoirEnabled, err = NewSettingAndParseValidate(AlarmLowReservoirEnabled, v, createdAt, subject)
		return err
	case "alarm_low_reservoir_units":
		m.AlarmLowReservoirUnits, err = NewSettingAndParseValidate(AlarmLowReservoirUnits, v, createdAt, subject)
		return err
	case "notifications_enabled":
		m.NotificationsEnabled, err = NewSettingAndParseValidate(NotificationsEnabled, v, createdAt, subject)
		return err
	default:
		return nil
	}

}
