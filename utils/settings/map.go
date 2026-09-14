package settings

import (
	"time"

	"nighttroop.t1d.nz/utils/consts"
)

type SettingAnyType struct {
	Identifier string
	TypeString string
	Value      any
}

type Setting[T any] struct {
	definition *SettingDefinition[T]
	Value      T
	CreatedAt  int64
}

func (s *Setting[T]) ToSettingAnyType() *SettingAnyType {
	if s == nil {
		return nil
	}
	return &SettingAnyType{
		Identifier: s.definition.Identifier,
		TypeString: s.definition.TypeString,
		Value:      s.Value,
	}
}

func NewSetting[T any](definition *SettingDefinition[T], value *T) *Setting[T] {
	// Setting value to nil is just deleting the setting
	if value == nil {
		return nil
	}

	setting := &Setting[T]{
		definition: definition,
		Value:      *value,
		CreatedAt:  time.Now().Unix(),
	}

	return setting
}

func NewDefaultSetting[T any](definition *SettingDefinition[T]) *Setting[T] {

	setting := &Setting[T]{
		definition: definition,
		Value:      definition.Default,
		CreatedAt:  time.Now().Unix(),
	}

	return setting
}

func NewSettingAndParseValidate[T any](definition *SettingDefinition[T], value any) (*Setting[T], error) {
	if value == nil {
		return nil, nil
	}
	parsedValue, err := definition.ParseValidate(value)
	if err != nil {
		return nil, err
	}
	return NewSetting(definition, parsedValue), nil
}

type SettingMap struct {
	Units                     *Setting[consts.Units]
	Timezone                  *Setting[string]
	Language                  *Setting[string]
	DateFormat                *Setting[string]
	TimeFormat                *Setting[string]
	WeekStartDay              *Setting[string]
	BolusDiaHours             *Setting[float64]
	BolusOnsetMinutes         *Setting[float64]
	BolusPeakMinutes          *Setting[float64]
	BasalDiaHours             *Setting[float64]
	BasalOnsetMinutes         *Setting[float64]
	BasalPeakMinutes          *Setting[float64]
	DefaultInsulinSensitivity *Setting[float64]
	DefaultCarbRatio          *Setting[float64]
	DefaultBasalRate          *Setting[float64]
	CarbMaxAbsorptionHours    *Setting[float64]
	CarbMin5mImpact           *Setting[float64]
	AlarmUrgentLowEnabled     *Setting[bool]
	AlarmUrgentLow            *Setting[int64]
	AlarmLowEnabled           *Setting[bool]
	AlarmLow                  *Setting[int64]
	AlarmLowMinutes           *Setting[int64]
	AlarmLowPredictionEnabled *Setting[bool]
	AlarmHighEnabled          *Setting[bool]
	AlarmHigh                 *Setting[int64]
	AlarmHighMinutes          *Setting[int64]
	AlarmUrgentHighEnabled    *Setting[bool]
	AlarmUrgentHigh           *Setting[int64]
	AlarmFastRiseEnabled      *Setting[bool]
	AlarmFastRise             *Setting[int64]
	AlarmFastRiseMinutes      *Setting[int64]
	AlarmFastDropEnabled      *Setting[bool]
	AlarmFastDrop             *Setting[int64]
	AlarmFastDropMinutes      *Setting[int64]
	AlarmNoDataEnabled        *Setting[bool]
	AlarmNoDataMinutes        *Setting[int64]
	AlarmBatteryLowEnabled    *Setting[bool]
	AlarmBatteryLowPercent    *Setting[int64]
	AlarmBatteryLowPwdEnabled *Setting[bool]
	AlarmBatteryLowPwdPercent *Setting[int64]
	AlarmLowReservoirEnabled  *Setting[bool]
	AlarmLowReservoirUnits    *Setting[int64]
	NotificationsEnabled      *Setting[bool]
}

var DefaultSettingsMap = &SettingMap{
	Units:                     NewDefaultSetting(Units),
	Timezone:                  NewDefaultSetting(Timezone),
	Language:                  NewDefaultSetting(Language),
	DateFormat:                NewDefaultSetting(DateFormat),
	TimeFormat:                NewDefaultSetting(TimeFormat),
	WeekStartDay:              NewDefaultSetting(WeekStartDay),
	BolusDiaHours:             NewDefaultSetting(BolusDIAHours),
	BolusOnsetMinutes:         NewDefaultSetting(BolusOnsetMinutes),
	BolusPeakMinutes:          NewDefaultSetting(BolusPeakMinutes),
	BasalDiaHours:             NewDefaultSetting(BasalDIAHours),
	BasalOnsetMinutes:         NewDefaultSetting(BasalOnsetMinutes),
	BasalPeakMinutes:          NewDefaultSetting(BasalPeakMinutes),
	DefaultInsulinSensitivity: NewDefaultSetting(DefaultInsulinSensitivity),
	DefaultCarbRatio:          NewDefaultSetting(DefaultCarbRatio),
	DefaultBasalRate:          NewDefaultSetting(DefaultBasalRate),
	CarbMaxAbsorptionHours:    NewDefaultSetting(CarbMaxAbsorptionHours),
	CarbMin5mImpact:           NewDefaultSetting(CarbMin5mImpact),
	AlarmUrgentLowEnabled:     NewDefaultSetting(AlarmUrgentLowEnabled),
	AlarmUrgentLow:            NewDefaultSetting(AlarmUrgentLow),
	AlarmLowEnabled:           NewDefaultSetting(AlarmLowEnabled),
	AlarmLow:                  NewDefaultSetting(AlarmLow),
	AlarmLowMinutes:           NewDefaultSetting(AlarmLowMinutes),
	AlarmLowPredictionEnabled: NewDefaultSetting(AlarmLowPredictionEnabled),
	AlarmHighEnabled:          NewDefaultSetting(AlarmHighEnabled),
	AlarmHigh:                 NewDefaultSetting(AlarmHigh),
	AlarmHighMinutes:          NewDefaultSetting(AlarmHighMinutes),
	AlarmUrgentHighEnabled:    NewDefaultSetting(AlarmUrgentHighEnabled),
	AlarmUrgentHigh:           NewDefaultSetting(AlarmUrgentHigh),
	AlarmFastRiseEnabled:      NewDefaultSetting(AlarmFastRiseEnabled),
	AlarmFastRise:             NewDefaultSetting(AlarmFastRise),
	AlarmFastRiseMinutes:      NewDefaultSetting(AlarmFastRiseMinutes),
	AlarmFastDropEnabled:      NewDefaultSetting(AlarmFastDropEnabled),
	AlarmFastDrop:             NewDefaultSetting(AlarmFastDrop),
	AlarmFastDropMinutes:      NewDefaultSetting(AlarmFastDropMinutes),
	AlarmNoDataEnabled:        NewDefaultSetting(AlarmNoDataEnabled),
	AlarmNoDataMinutes:        NewDefaultSetting(AlarmNoDataMinutes),
	AlarmBatteryLowEnabled:    NewDefaultSetting(AlarmBatteryLowEnabled),
	AlarmBatteryLowPercent:    NewDefaultSetting(AlarmBatteryLowPercent),
	AlarmBatteryLowPwdEnabled: NewDefaultSetting(AlarmBatteryLowPwdEnabled),
	AlarmBatteryLowPwdPercent: NewDefaultSetting(AlarmBatteryLowPwdPercent),
	AlarmLowReservoirEnabled:  NewDefaultSetting(AlarmLowReservoirEnabled),
	AlarmLowReservoirUnits:    NewDefaultSetting(AlarmLowReservoirUnits),
	NotificationsEnabled:      NewDefaultSetting(NotificationsEnabled),
}

// New Setting Does not validate Check
func NewSettingsMap(g SettingsGetter) *SettingMap {
	s := &SettingMap{
		Units:                     NewSetting(Units, g.GetUnits()),
		Timezone:                  NewSetting(Timezone, g.GetTimezone()),
		Language:                  NewSetting(Language, g.GetLanguage()),
		DateFormat:                NewSetting(DateFormat, g.GetDateFormat()),
		TimeFormat:                NewSetting(TimeFormat, g.GetTimeFormat()),
		WeekStartDay:              NewSetting(WeekStartDay, g.GetWeekStartDay()),
		BolusDiaHours:             NewSetting(BolusDIAHours, g.GetBolusDiaHours()),
		BolusOnsetMinutes:         NewSetting(BolusOnsetMinutes, g.GetBolusOnsetMinutes()),
		BolusPeakMinutes:          NewSetting(BolusPeakMinutes, g.GetBolusPeakMinutes()),
		BasalDiaHours:             NewSetting(BasalDIAHours, g.GetBasalDiaHours()),
		BasalOnsetMinutes:         NewSetting(BasalOnsetMinutes, g.GetBasalOnsetMinutes()),
		BasalPeakMinutes:          NewSetting(BasalPeakMinutes, g.GetBasalPeakMinutes()),
		DefaultInsulinSensitivity: NewSetting(DefaultInsulinSensitivity, g.GetDefaultInsulinSensitivity()),
		DefaultCarbRatio:          NewSetting(DefaultCarbRatio, g.GetDefaultCarbRatio()),
		DefaultBasalRate:          NewSetting(DefaultBasalRate, g.GetDefaultBasalRate()),
		CarbMaxAbsorptionHours:    NewSetting(CarbMaxAbsorptionHours, g.GetCarbMaxAbsorptionHours()),
		CarbMin5mImpact:           NewSetting(CarbMin5mImpact, g.GetCarbMin5mImpact()),
		AlarmUrgentLowEnabled:     NewSetting(AlarmUrgentLowEnabled, g.GetAlarmUrgentLowEnabled()),
		AlarmUrgentLow:            NewSetting(AlarmUrgentLow, g.GetAlarmUrgentLow()),
		AlarmLowEnabled:           NewSetting(AlarmLowEnabled, g.GetAlarmLowEnabled()),
		AlarmLow:                  NewSetting(AlarmLow, g.GetAlarmLow()),
		AlarmLowMinutes:           NewSetting(AlarmLowMinutes, g.GetAlarmLowMinutes()),
		AlarmLowPredictionEnabled: NewSetting(AlarmLowPredictionEnabled, g.GetAlarmLowPredictionEnabled()),
		AlarmHighEnabled:          NewSetting(AlarmHighEnabled, g.GetAlarmHighEnabled()),
		AlarmHigh:                 NewSetting(AlarmHigh, g.GetAlarmHigh()),
		AlarmHighMinutes:          NewSetting(AlarmHighMinutes, g.GetAlarmHighMinutes()),
		AlarmUrgentHighEnabled:    NewSetting(AlarmUrgentHighEnabled, g.GetAlarmUrgentHighEnabled()),
		AlarmUrgentHigh:           NewSetting(AlarmUrgentHigh, g.GetAlarmUrgentHigh()),
		AlarmFastRiseEnabled:      NewSetting(AlarmFastRiseEnabled, g.GetAlarmFastRiseEnabled()),
		AlarmFastRise:             NewSetting(AlarmFastRise, g.GetAlarmFastRise()),
		AlarmFastRiseMinutes:      NewSetting(AlarmFastRiseMinutes, g.GetAlarmFastRiseMinutes()),
		AlarmFastDropEnabled:      NewSetting(AlarmFastDropEnabled, g.GetAlarmFastDropEnabled()),
		AlarmFastDrop:             NewSetting(AlarmFastDrop, g.GetAlarmFastDrop()),
		AlarmFastDropMinutes:      NewSetting(AlarmFastDropMinutes, g.GetAlarmFastDropMinutes()),
		AlarmNoDataEnabled:        NewSetting(AlarmNoDataEnabled, g.GetAlarmNoDataEnabled()),
		AlarmNoDataMinutes:        NewSetting(AlarmNoDataMinutes, g.GetAlarmNoDataMinutes()),
		AlarmBatteryLowEnabled:    NewSetting(AlarmBatteryLowEnabled, g.GetAlarmBatteryLowEnabled()),
		AlarmBatteryLowPercent:    NewSetting(AlarmBatteryLowPercent, g.GetAlarmBatteryLowPercent()),
		AlarmBatteryLowPwdEnabled: NewSetting(AlarmBatteryLowPwdEnabled, g.GetAlarmBatteryLowPwdEnabled()),
		AlarmBatteryLowPwdPercent: NewSetting(AlarmBatteryLowPwdPercent, g.GetAlarmBatteryLowPwdPercent()),
		AlarmLowReservoirEnabled:  NewSetting(AlarmLowReservoirEnabled, g.GetAlarmLowReservoirEnabled()),
		AlarmLowReservoirUnits:    NewSetting(AlarmLowReservoirUnits, g.GetAlarmLowReservoirUnits()),
		NotificationsEnabled:      NewSetting(NotificationsEnabled, g.GetNotificationsEnabled()),
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
	m.Units = NewSetting(Units, v)
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
	m.Timezone = NewSetting(Timezone, v)
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

	m.Language = NewSetting(Language, v)

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

	m.DateFormat = NewSetting(DateFormat, v)

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

	m.TimeFormat = NewSetting(TimeFormat, v)

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

	m.WeekStartDay = NewSetting(WeekStartDay, v)

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

	m.BolusDiaHours = NewSetting(BolusDIAHours, v)

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
	m.BolusOnsetMinutes = NewSetting(BolusOnsetMinutes, v)
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

	m.BolusPeakMinutes = NewSetting(BolusPeakMinutes, v)

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

	m.BasalDiaHours = NewSetting(BasalDIAHours, v)

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

	m.BasalOnsetMinutes = NewSetting(BasalOnsetMinutes, v)

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

	m.BasalPeakMinutes = NewSetting(BasalPeakMinutes, v)

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

	m.DefaultInsulinSensitivity = NewSetting(DefaultInsulinSensitivity, v)

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

	m.DefaultCarbRatio = NewSetting(DefaultCarbRatio, v)

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

	m.DefaultBasalRate = NewSetting(DefaultBasalRate, v)

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

	m.CarbMaxAbsorptionHours = NewSetting(CarbMaxAbsorptionHours, v)

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

	m.CarbMin5mImpact = NewSetting(CarbMin5mImpact, v)

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

	m.AlarmUrgentLowEnabled = NewSetting(AlarmUrgentLowEnabled, v)

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

	m.AlarmUrgentLow = NewSetting(AlarmUrgentLow, v)

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

	m.AlarmLowEnabled = NewSetting(AlarmLowEnabled, v)

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

	m.AlarmLow = NewSetting(AlarmLow, v)

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

	m.AlarmLowMinutes = NewSetting(AlarmLowMinutes, v)

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

	m.AlarmLowPredictionEnabled = NewSetting(AlarmLowPredictionEnabled, v)

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

	m.AlarmHighEnabled = NewSetting(AlarmHighEnabled, v)

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

	m.AlarmHigh = NewSetting(AlarmHigh, v)

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

	m.AlarmHighMinutes = NewSetting(AlarmHighMinutes, v)

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

	m.AlarmUrgentHighEnabled = NewSetting(AlarmUrgentHighEnabled, v)

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

	m.AlarmUrgentHigh = NewSetting(AlarmUrgentHigh, v)

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

	m.AlarmFastRiseEnabled = NewSetting(AlarmFastRiseEnabled, v)

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

	m.AlarmFastRise = NewSetting(AlarmFastRise, v)

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

	m.AlarmFastRiseMinutes = NewSetting(AlarmFastRiseMinutes, v)

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

	m.AlarmFastDropEnabled = NewSetting(AlarmFastDropEnabled, v)

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

	m.AlarmFastDrop = NewSetting(AlarmFastDrop, v)

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

	m.AlarmFastDropMinutes = NewSetting(AlarmFastDropMinutes, v)

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

	m.AlarmNoDataEnabled = NewSetting(AlarmNoDataEnabled, v)

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

	m.AlarmNoDataMinutes = NewSetting(AlarmNoDataMinutes, v)

}

// GetAlarmBatteryLowEnabled returns the AlarmBatteryLowEnabled value.
func (m SettingMap) GetAlarmBatteryLowEnabled() *bool {
	if m.AlarmBatteryLowEnabled == nil {
		return nil
	}
	return ptr(m.AlarmBatteryLowEnabled.Value)
}

// SetAlarmBatteryLowEnabled sets the AlarmBatteryLowEnabled value and returns the updated SettingMap.
func (m *SettingMap) SetAlarmBatteryLowEnabled(v *bool) {

	m.AlarmBatteryLowEnabled = NewSetting(AlarmBatteryLowEnabled, v)

}

// GetAlarmBatteryLowPercent returns the AlarmBatteryLowPercent value.
func (m SettingMap) GetAlarmBatteryLowPercent() *int64 {
	if m.AlarmBatteryLowPercent == nil {
		return nil
	}
	return ptr(m.AlarmBatteryLowPercent.Value)
}

// SetAlarmBatteryLowPercent sets the AlarmBatteryLowPercent value and returns the updated SettingMap.
func (m *SettingMap) SetAlarmBatteryLowPercent(v *int64) {

	m.AlarmBatteryLowPercent = NewSetting(AlarmBatteryLowPercent, v)

}

// GetAlarmBatteryLowPwdEnabled returns the AlarmBatteryLowPwdEnabled value.
func (m SettingMap) GetAlarmBatteryLowPwdEnabled() *bool {
	if m.AlarmBatteryLowPwdEnabled == nil {
		return nil
	}
	return ptr(m.AlarmBatteryLowPwdEnabled.Value)
}

// SetAlarmBatteryLowPwdEnabled sets the AlarmBatteryLowPwdEnabled value and returns the updated SettingMap.
func (m *SettingMap) SetAlarmBatteryLowPwdEnabled(v *bool) {

	m.AlarmBatteryLowPwdEnabled = NewSetting(AlarmBatteryLowPwdEnabled, v)

}

// GetAlarmBatteryLowPwdPercent returns the AlarmBatteryLowPwdPercent value.
func (m SettingMap) GetAlarmBatteryLowPwdPercent() *int64 {
	if m.AlarmBatteryLowPwdPercent == nil {
		return nil
	}
	return ptr(m.AlarmBatteryLowPwdPercent.Value)
}

// SetAlarmBatteryLowPwdPercent sets the AlarmBatteryLowPwdPercent value and returns the updated SettingMap.
func (m *SettingMap) SetAlarmBatteryLowPwdPercent(v *int64) {

	m.AlarmBatteryLowPwdPercent = NewSetting(AlarmBatteryLowPwdPercent, v)

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

	m.AlarmLowReservoirEnabled = NewSetting(AlarmLowReservoirEnabled, v)

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

	m.AlarmLowReservoirUnits = NewSetting(AlarmLowReservoirUnits, v)

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

	m.NotificationsEnabled = NewSetting(NotificationsEnabled, v)

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
	case "alarm_battery_low_enabled":
		return m.AlarmBatteryLowEnabled.ToSettingAnyType()
	case "alarm_battery_low_percent":
		return m.AlarmBatteryLowPercent.ToSettingAnyType()
	case "alarm_battery_low_pwd_enabled":
		return m.AlarmBatteryLowPwdEnabled.ToSettingAnyType()
	case "alarm_battery_low_pwd_percent":
		return m.AlarmBatteryLowPwdPercent.ToSettingAnyType()
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

func (m *SettingMap) SetSettingByName(name string, v any) error {
	var err error
	switch name {
	case "units":
		m.Units, err = NewSettingAndParseValidate(Units, v)
		return err
	case "timezone":
		m.Timezone, err = NewSettingAndParseValidate(Timezone, v)
		return err
	case "language":
		m.Language, err = NewSettingAndParseValidate(Language, v)
		return err
	case "date_format":
		m.DateFormat, err = NewSettingAndParseValidate(DateFormat, v)
		return err
	case "time_format":
		m.TimeFormat, err = NewSettingAndParseValidate(TimeFormat, v)
		return err
	case "week_start_day":
		m.WeekStartDay, err = NewSettingAndParseValidate(WeekStartDay, v)
		return err
	case "bolus_dia_hours":
		m.BolusDiaHours, err = NewSettingAndParseValidate(BolusDIAHours, v)
		return err
	case "bolus_onset_minutes":
		m.BolusOnsetMinutes, err = NewSettingAndParseValidate(BolusOnsetMinutes, v)
		return err
	case "bolus_peak_minutes":
		m.BolusPeakMinutes, err = NewSettingAndParseValidate(BolusPeakMinutes, v)
		return err
	case "basal_dia_hours":
		m.BasalDiaHours, err = NewSettingAndParseValidate(BasalDIAHours, v)
		return err
	case "basal_onset_minutes":
		m.BasalOnsetMinutes, err = NewSettingAndParseValidate(BasalOnsetMinutes, v)
		return err
	case "basal_peak_minutes":
		m.BasalPeakMinutes, err = NewSettingAndParseValidate(BasalPeakMinutes, v)
		return err
	case "default_insulin_sensitivity":
		m.DefaultInsulinSensitivity, err = NewSettingAndParseValidate(DefaultInsulinSensitivity, v)
		return err
	case "default_carb_ratio":
		m.DefaultCarbRatio, err = NewSettingAndParseValidate(DefaultCarbRatio, v)
		return err
	case "default_basal_rate":
		m.DefaultBasalRate, err = NewSettingAndParseValidate(DefaultBasalRate, v)
		return err
	case "carb_max_absorption_hours":
		m.CarbMaxAbsorptionHours, err = NewSettingAndParseValidate(CarbMaxAbsorptionHours, v)
		return err
	case "carb_min_5m_impact":
		m.CarbMin5mImpact, err = NewSettingAndParseValidate(CarbMin5mImpact, v)
		return err
	case "alarm_urgent_low_enabled":
		m.AlarmUrgentLowEnabled, err = NewSettingAndParseValidate(AlarmUrgentLowEnabled, v)
		return err
	case "alarm_urgent_low":
		m.AlarmUrgentLow, err = NewSettingAndParseValidate(AlarmUrgentLow, v)
		return err
	case "alarm_low_enabled":
		m.AlarmLowEnabled, err = NewSettingAndParseValidate(AlarmLowEnabled, v)
		return err
	case "alarm_low":
		m.AlarmLow, err = NewSettingAndParseValidate(AlarmLow, v)
		return err
	case "alarm_low_minutes":
		m.AlarmLowMinutes, err = NewSettingAndParseValidate(AlarmLowMinutes, v)
		return err
	case "alarm_low_prediction_enabled":
		m.AlarmLowPredictionEnabled, err = NewSettingAndParseValidate(AlarmLowPredictionEnabled, v)
		return err
	case "alarm_high_enabled":
		m.AlarmHighEnabled, err = NewSettingAndParseValidate(AlarmHighEnabled, v)
		return err
	case "alarm_high":
		m.AlarmHigh, err = NewSettingAndParseValidate(AlarmHigh, v)
		return err
	case "alarm_high_minutes":
		m.AlarmHighMinutes, err = NewSettingAndParseValidate(AlarmHighMinutes, v)
		return err
	case "alarm_urgent_high_enabled":
		m.AlarmUrgentHighEnabled, err = NewSettingAndParseValidate(AlarmUrgentHighEnabled, v)
		return err
	case "alarm_urgent_high":
		m.AlarmUrgentHigh, err = NewSettingAndParseValidate(AlarmUrgentHigh, v)
		return err
	case "alarm_fast_rise_enabled":
		m.AlarmFastRiseEnabled, err = NewSettingAndParseValidate(AlarmFastRiseEnabled, v)
		return err
	case "alarm_fast_rise":
		m.AlarmFastRise, err = NewSettingAndParseValidate(AlarmFastRise, v)
		return err
	case "alarm_fast_rise_minutes":
		m.AlarmFastRiseMinutes, err = NewSettingAndParseValidate(AlarmFastRiseMinutes, v)
		return err
	case "alarm_fast_drop_enabled":
		m.AlarmFastDropEnabled, err = NewSettingAndParseValidate(AlarmFastDropEnabled, v)
		return err
	case "alarm_fast_drop":
		m.AlarmFastDrop, err = NewSettingAndParseValidate(AlarmFastDrop, v)
		return err
	case "alarm_fast_drop_minutes":
		m.AlarmFastDropMinutes, err = NewSettingAndParseValidate(AlarmFastDropMinutes, v)
		return err
	case "alarm_no_data_enabled":
		m.AlarmNoDataEnabled, err = NewSettingAndParseValidate(AlarmNoDataEnabled, v)
		return err
	case "alarm_no_data_minutes":
		m.AlarmNoDataMinutes, err = NewSettingAndParseValidate(AlarmNoDataMinutes, v)
		return err
	case "alarm_battery_low_enabled":
		m.AlarmBatteryLowEnabled, err = NewSettingAndParseValidate(AlarmBatteryLowEnabled, v)
		return err
	case "alarm_battery_low_percent":
		m.AlarmBatteryLowPercent, err = NewSettingAndParseValidate(AlarmBatteryLowPercent, v)
		return err
	case "alarm_battery_low_pwd_enabled":
		m.AlarmBatteryLowPwdEnabled, err = NewSettingAndParseValidate(AlarmBatteryLowPwdEnabled, v)
		return err
	case "alarm_battery_low_pwd_percent":
		m.AlarmBatteryLowPwdPercent, err = NewSettingAndParseValidate(AlarmBatteryLowPwdPercent, v)
		return err
	case "alarm_low_reservoir_enabled":
		m.AlarmLowReservoirEnabled, err = NewSettingAndParseValidate(AlarmLowReservoirEnabled, v)
		return err
	case "alarm_low_reservoir_units":
		m.AlarmLowReservoirUnits, err = NewSettingAndParseValidate(AlarmLowReservoirUnits, v)
		return err
	case "notifications_enabled":
		m.NotificationsEnabled, err = NewSettingAndParseValidate(NotificationsEnabled, v)
		return err
	default:
		return nil
	}

}
