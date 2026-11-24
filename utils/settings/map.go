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
	NumberFormat                     *Setting[string]
	AlarmUrgentLowEnabled            *Setting[bool]
	RangeUrgentLowMgdl               *Setting[int64]
	AlarmLowEnabled                  *Setting[bool]
	RangeLowMgdl                     *Setting[int64]
	AlarmLowMinutes                  *Setting[int64]
	AlarmLowPredictionEnabled        *Setting[bool]
	AlarmLowPredictionMinutes        *Setting[int64]
	AlarmHighEnabled                 *Setting[bool]
	RangeHighMgdl                    *Setting[int64]
	AlarmHighMinutes                 *Setting[int64]
	AlarmUrgentHighEnabled           *Setting[bool]
	RangeUrgentHighMgdl              *Setting[int64]
	AlarmFastRiseEnabled             *Setting[bool]
	AlarmFastRiseMgdl                *Setting[int64]
	AlarmFastRiseMinutes             *Setting[int64]
	AlarmFastDropEnabled             *Setting[bool]
	AlarmFastDropMgdl                *Setting[int64]
	AlarmFastDropMinutes             *Setting[int64]
	AlarmNoDataEnabled               *Setting[bool]
	AlarmNoDataMinutes               *Setting[int64]
	AlarmDeviceBatteryLowEnabled     *Setting[bool]
	AlarmDeviceBatteryLowPercent     *Setting[int64]
	AlarmDeviceBatteryLowMinutes     *Setting[int64]
	AlarmDeviceBatteryLowDevicesJson *Setting[types.DeviceNames]
}

var DefaultSettingsMap = &SettingMap{
	Units:                            NewDefaultSetting(Units),
	Timezone:                         NewDefaultSetting(Timezone),
	Language:                         NewDefaultSetting(Language),
	DateFormat:                       NewDefaultSetting(DateFormat),
	TimeFormat:                       NewDefaultSetting(TimeFormat),
	WeekStartDay:                     NewDefaultSetting(WeekStartDay),
	NumberFormat:                     NewDefaultSetting(NumberFormat),
	AlarmUrgentLowEnabled:            NewDefaultSetting(AlarmUrgentLowEnabled),
	RangeUrgentLowMgdl:               NewDefaultSetting(RangeUrgentLowMgdl),
	AlarmLowEnabled:                  NewDefaultSetting(AlarmLowEnabled),
	RangeLowMgdl:                     NewDefaultSetting(RangeLowMgdl),
	AlarmLowMinutes:                  NewDefaultSetting(AlarmLowMinutes),
	AlarmLowPredictionEnabled:        NewDefaultSetting(AlarmLowPredictionEnabled),
	AlarmLowPredictionMinutes:        NewDefaultSetting(AlarmLowPredictionMinutes),
	AlarmHighEnabled:                 NewDefaultSetting(AlarmHighEnabled),
	RangeHighMgdl:                    NewDefaultSetting(RangeHighMgdl),
	AlarmHighMinutes:                 NewDefaultSetting(AlarmHighMinutes),
	AlarmUrgentHighEnabled:           NewDefaultSetting(AlarmUrgentHighEnabled),
	RangeUrgentHighMgdl:              NewDefaultSetting(RangeUrgentHighMgdl),
	AlarmFastRiseEnabled:             NewDefaultSetting(AlarmFastRiseEnabled),
	AlarmFastRiseMgdl:                NewDefaultSetting(AlarmFastRiseMgdl),
	AlarmFastRiseMinutes:             NewDefaultSetting(AlarmFastRiseMinutes),
	AlarmFastDropEnabled:             NewDefaultSetting(AlarmFastDropEnabled),
	AlarmFastDropMgdl:                NewDefaultSetting(AlarmFastDropMgdl),
	AlarmFastDropMinutes:             NewDefaultSetting(AlarmFastDropMinutes),
	AlarmNoDataEnabled:               NewDefaultSetting(AlarmNoDataEnabled),
	AlarmNoDataMinutes:               NewDefaultSetting(AlarmNoDataMinutes),
	AlarmDeviceBatteryLowEnabled:     NewDefaultSetting(AlarmDeviceBatteryLowEnabled),
	AlarmDeviceBatteryLowPercent:     NewDefaultSetting(AlarmDeviceBatteryLowPercent),
	AlarmDeviceBatteryLowMinutes:     NewDefaultSetting(AlarmDeviceBatteryLowMinutes),
	AlarmDeviceBatteryLowDevicesJson: NewDefaultSetting(AlarmDeviceBatteryLowDevicesJson),
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

// GetNumberFormat returns the NumberFormat value.
func (m SettingMap) GetNumberFormat() *string {
	if m.NumberFormat == nil {
		return nil
	}
	return ptr(m.NumberFormat.Value)
}

// GetAlarmUrgentLowEnabled returns the AlarmUrgentLowEnabled value.
func (m SettingMap) GetAlarmUrgentLowEnabled() *bool {
	if m.AlarmUrgentLowEnabled == nil {
		return nil
	}
	return ptr(m.AlarmUrgentLowEnabled.Value)
}

// GetRangeUrgentLowMgdl returns the RangeUrgentLowMgdl value.
func (m SettingMap) GetRangeUrgentLowMgdl() *int64 {
	if m.RangeUrgentLowMgdl == nil {
		return nil
	}
	return ptr(m.RangeUrgentLowMgdl.Value)
}

// GetAlarmLowEnabled returns the AlarmLowEnabled value.
func (m SettingMap) GetAlarmLowEnabled() *bool {
	if m.AlarmLowEnabled == nil {
		return nil
	}
	return ptr(m.AlarmLowEnabled.Value)
}

// GetRangeLowMgdl returns the RangeLowMgdl value.
func (m SettingMap) GetRangeLowMgdl() *int64 {
	if m.RangeLowMgdl == nil {
		return nil
	}
	return ptr(m.RangeLowMgdl.Value)
}

// GetAlarmLowMinutes returns the AlarmLowMinutes value.
func (m SettingMap) GetAlarmLowMinutes() *int64 {
	if m.AlarmLowMinutes == nil {
		return nil
	}
	return ptr(m.AlarmLowMinutes.Value)
}

// GetAlarmLowPredictionEnabled returns the AlarmLowPredictionEnabled value.
func (m SettingMap) GetAlarmLowPredictionEnabled() *bool {
	if m.AlarmLowPredictionEnabled == nil {
		return nil
	}
	return ptr(m.AlarmLowPredictionEnabled.Value)
}

// GetAlarmLowPredictionMinutes returns the AlarmLowPredictionMinutes value.
func (m SettingMap) GetAlarmLowPredictionMinutes() *int64 {
	if m.AlarmLowPredictionMinutes == nil {
		return nil
	}
	return ptr(m.AlarmLowPredictionMinutes.Value)
}

// GetAlarmHighEnabled returns the AlarmHighEnabled value.
func (m SettingMap) GetAlarmHighEnabled() *bool {
	if m.AlarmHighEnabled == nil {
		return nil
	}
	return ptr(m.AlarmHighEnabled.Value)
}

// GetRangeHighMgdl returns the RangeHighMgdl value.
func (m SettingMap) GetRangeHighMgdl() *int64 {
	if m.RangeHighMgdl == nil {
		return nil
	}
	return ptr(m.RangeHighMgdl.Value)
}

// GetAlarmHighMinutes returns the AlarmHighMinutes value.
func (m SettingMap) GetAlarmHighMinutes() *int64 {
	if m.AlarmHighMinutes == nil {
		return nil
	}
	return ptr(m.AlarmHighMinutes.Value)
}

// GetAlarmUrgentHighEnabled returns the AlarmUrgentHighEnabled value.
func (m SettingMap) GetAlarmUrgentHighEnabled() *bool {
	if m.AlarmUrgentHighEnabled == nil {
		return nil
	}
	return ptr(m.AlarmUrgentHighEnabled.Value)
}

// GetRangeUrgentHighMgdl returns the RangeUrgentHighMgdl value.
func (m SettingMap) GetRangeUrgentHighMgdl() *int64 {
	if m.RangeUrgentHighMgdl == nil {
		return nil
	}
	return ptr(m.RangeUrgentHighMgdl.Value)
}

// GetAlarmFastRiseEnabled returns the AlarmFastRiseEnabled value.
func (m SettingMap) GetAlarmFastRiseEnabled() *bool {
	if m.AlarmFastRiseEnabled == nil {
		return nil
	}
	return ptr(m.AlarmFastRiseEnabled.Value)
}

// GetAlarmFastRiseMgdl returns the AlarmFastRiseMgdl value.
func (m SettingMap) GetAlarmFastRiseMgdl() *int64 {
	if m.AlarmFastRiseMgdl == nil {
		return nil
	}
	return ptr(m.AlarmFastRiseMgdl.Value)
}

// GetAlarmFastRiseMinutes returns the AlarmFastRiseMinutes value.
func (m SettingMap) GetAlarmFastRiseMinutes() *int64 {
	if m.AlarmFastRiseMinutes == nil {
		return nil
	}
	return ptr(m.AlarmFastRiseMinutes.Value)
}

// GetAlarmFastDropEnabled returns the AlarmFastDropEnabled value.
func (m SettingMap) GetAlarmFastDropEnabled() *bool {
	if m.AlarmFastDropEnabled == nil {
		return nil
	}
	return ptr(m.AlarmFastDropEnabled.Value)
}

// GetAlarmFastDropMgdl returns the AlarmFastDropMgdl value.
func (m SettingMap) GetAlarmFastDropMgdl() *int64 {
	if m.AlarmFastDropMgdl == nil {
		return nil
	}
	return ptr(m.AlarmFastDropMgdl.Value)
}

// GetAlarmFastDropMinutes returns the AlarmFastDropMinutes value.
func (m SettingMap) GetAlarmFastDropMinutes() *int64 {
	if m.AlarmFastDropMinutes == nil {
		return nil
	}
	return ptr(m.AlarmFastDropMinutes.Value)
}

// GetAlarmNoDataEnabled returns the AlarmNoDataEnabled value.
func (m SettingMap) GetAlarmNoDataEnabled() *bool {
	if m.AlarmNoDataEnabled == nil {
		return nil
	}
	return ptr(m.AlarmNoDataEnabled.Value)
}

// GetAlarmNoDataMinutes returns the AlarmNoDataMinutes value.
func (m SettingMap) GetAlarmNoDataMinutes() *int64 {
	if m.AlarmNoDataMinutes == nil {
		return nil
	}
	return ptr(m.AlarmNoDataMinutes.Value)
}

// GetAlarmDeviceBatteryLowEnabled returns the AlarmDeviceBatteryLowEnabled value.
func (m SettingMap) GetAlarmDeviceBatteryLowEnabled() *bool {
	if m.AlarmDeviceBatteryLowEnabled == nil {
		return nil
	}
	return ptr(m.AlarmDeviceBatteryLowEnabled.Value)
}

// GetAlarmDeviceBatteryLowPercent returns the AlarmDeviceBatteryLowPercent value.
func (m SettingMap) GetAlarmDeviceBatteryLowPercent() *int64 {
	if m.AlarmDeviceBatteryLowPercent == nil {
		return nil
	}
	return ptr(m.AlarmDeviceBatteryLowPercent.Value)
}

// GetAlarmDeviceBatteryLowMinutes returns the AlarmDeviceBatteryLowMinutes value.
func (m SettingMap) GetAlarmDeviceBatteryLowMinutes() *int64 {
	if m.AlarmDeviceBatteryLowMinutes == nil {
		return nil
	}
	return ptr(m.AlarmDeviceBatteryLowMinutes.Value)
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

func (m SettingMap) GetSettingAnyTypeByName(name string) *SettingAnyType {
	if s := m.slotFor(name); s != nil {
		return s.anyType()
	}
	return nil
}

func (m *SettingMap) SetSettingByName(name string, v any, createdAt *int64, subject *string) error {
	if s := m.slotFor(name); s != nil {
		return s.parseValidate(v, createdAt, subject)
	}
	return nil
}
