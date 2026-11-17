
package controldb

import (
	"nighttroop.t1d.nz/data/types"
	"nighttroop.t1d.nz/utils/consts"
)


// GetID returns the ID value.
func (m ControlIdentity) GetID() int64 {
	return m.ID
}

// SetID sets the ID value and returns the updated ControlIdentity.
func (m *ControlIdentity) SetID(v int64) {
	m.ID = v
}

// GetName returns the Name value.
func (m ControlIdentity) GetName() string {
	return m.Name
}

// GetCreatedAt returns the CreatedAt value.
func (m ControlIdentity) GetCreatedAt() int64 {
	return m.CreatedAt
}

// SetCreatedAt sets the CreatedAt value and returns the updated ControlIdentity.
func (m *ControlIdentity) SetCreatedAt(v int64) {
	m.CreatedAt = v
}

// GetCreatedByIdentityName returns the CreatedByIdentityName value.
func (m ControlIdentity) GetCreatedByIdentityName() string {
	return m.CreatedByIdentityName
}

// SetCreatedByIdentityName sets the CreatedByIdentityName value and returns the updated ControlIdentity.
func (m *ControlIdentity) SetCreatedByIdentityName(v string) {
	m.CreatedByIdentityName = v
}

// GetDeletedAt returns the DeletedAt value.
func (m ControlIdentity) GetDeletedAt() *int64 {
	return m.DeletedAt
}

// SetDeletedAt sets the DeletedAt value and returns the updated ControlIdentity.
func (m *ControlIdentity) SetDeletedAt(v *int64) {
	m.DeletedAt = v
}

// GetCredentialsChangedAt returns the CredentialsChangedAt value.
func (m ControlIdentity) GetCredentialsChangedAt() int64 {
	return m.CredentialsChangedAt
}

// GetID returns the ID value.
func (m Passkey) GetID() int64 {
	return m.ID
}

// SetID sets the ID value and returns the updated Passkey.
func (m *Passkey) SetID(v int64) {
	m.ID = v
}

// GetIdentityID returns the IdentityID value.
func (m Passkey) GetIdentityID() *int64 {
	return m.IdentityID
}

// GetCreatedAt returns the CreatedAt value.
func (m Passkey) GetCreatedAt() int64 {
	return m.CreatedAt
}

// SetCreatedAt sets the CreatedAt value and returns the updated Passkey.
func (m *Passkey) SetCreatedAt(v int64) {
	m.CreatedAt = v
}

// GetID returns the ID value.
func (m PushCredential) GetID() int64 {
	return m.ID
}

// SetID sets the ID value and returns the updated PushCredential.
func (m *PushCredential) SetID(v int64) {
	m.ID = v
}

// GetName returns the Name value.
func (m PushCredential) GetName() string {
	return m.Name
}

// GetCreatedAt returns the CreatedAt value.
func (m PushCredential) GetCreatedAt() int64 {
	return m.CreatedAt
}

// SetCreatedAt sets the CreatedAt value and returns the updated PushCredential.
func (m *PushCredential) SetCreatedAt(v int64) {
	m.CreatedAt = v
}

// GetUpdatedAt returns the UpdatedAt value.
func (m PushCredential) GetUpdatedAt() int64 {
	return m.UpdatedAt
}

// SetUpdatedAt sets the UpdatedAt value and returns the updated PushCredential.
func (m *PushCredential) SetUpdatedAt(v int64) {
	m.UpdatedAt = v
}

// GetCreatedByIdentityName returns the CreatedByIdentityName value.
func (m PushCredential) GetCreatedByIdentityName() string {
	return m.CreatedByIdentityName
}

// SetCreatedByIdentityName sets the CreatedByIdentityName value and returns the updated PushCredential.
func (m *PushCredential) SetCreatedByIdentityName(v string) {
	m.CreatedByIdentityName = v
}

// GetID returns the ID value.
func (m PushRegistration) GetID() int64 {
	return m.ID
}

// SetID sets the ID value and returns the updated PushRegistration.
func (m *PushRegistration) SetID(v int64) {
	m.ID = v
}

// GetIdentityID returns the IdentityID value.
func (m PushRegistration) GetIdentityID() int64 {
	return m.IdentityID
}

// GetApp returns the App value.
func (m PushRegistration) GetApp() consts.PushApp {
	return m.App
}

// GetCreatedAt returns the CreatedAt value.
func (m PushRegistration) GetCreatedAt() int64 {
	return m.CreatedAt
}

// SetCreatedAt sets the CreatedAt value and returns the updated PushRegistration.
func (m *PushRegistration) SetCreatedAt(v int64) {
	m.CreatedAt = v
}

// GetUpdatedAt returns the UpdatedAt value.
func (m PushRegistration) GetUpdatedAt() int64 {
	return m.UpdatedAt
}

// SetUpdatedAt sets the UpdatedAt value and returns the updated PushRegistration.
func (m *PushRegistration) SetUpdatedAt(v int64) {
	m.UpdatedAt = v
}

// GetID returns the ID value.
func (m ServerSetting) GetID() int64 {
	return m.ID
}

// SetID sets the ID value and returns the updated ServerSetting.
func (m *ServerSetting) SetID(v int64) {
	m.ID = v
}

// GetUpdatedAt returns the UpdatedAt value.
func (m ServerSetting) GetUpdatedAt() *int64 {
	return m.UpdatedAt
}

// SetUpdatedAt sets the UpdatedAt value and returns the updated ServerSetting.
func (m *ServerSetting) SetUpdatedAt(v *int64) {
	m.UpdatedAt = v
}

// GetID returns the ID value.
func (m Site) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated Site.
func (m *Site) SetID(v string) {
	m.ID = v
}

// GetCreatedAt returns the CreatedAt value.
func (m Site) GetCreatedAt() int64 {
	return m.CreatedAt
}

// SetCreatedAt sets the CreatedAt value and returns the updated Site.
func (m *Site) SetCreatedAt(v int64) {
	m.CreatedAt = v
}

// GetDeletedAt returns the DeletedAt value.
func (m Site) GetDeletedAt() *int64 {
	return m.DeletedAt
}

// SetDeletedAt sets the DeletedAt value and returns the updated Site.
func (m *Site) SetDeletedAt(v *int64) {
	m.DeletedAt = v
}

// GetUnits returns the Units value.
func (m Site) GetUnits() *consts.Units {
	return m.Units
}

// SetUnits sets the Units value and returns the updated Site.
func (m *Site) SetUnits(v *consts.Units) {
	m.Units = v
}

// GetUnitsMgdlLabel returns the UnitsMgdlLabel value.
func (m Site) GetUnitsMgdlLabel() *string {
	return m.UnitsMgdlLabel
}

// SetUnitsMgdlLabel sets the UnitsMgdlLabel value and returns the updated Site.
func (m *Site) SetUnitsMgdlLabel(v *string) {
	m.UnitsMgdlLabel = v
}

// GetUnitsMmolLabel returns the UnitsMmolLabel value.
func (m Site) GetUnitsMmolLabel() *string {
	return m.UnitsMmolLabel
}

// SetUnitsMmolLabel sets the UnitsMmolLabel value and returns the updated Site.
func (m *Site) SetUnitsMmolLabel(v *string) {
	m.UnitsMmolLabel = v
}

// GetTimezone returns the Timezone value.
func (m Site) GetTimezone() *string {
	return m.Timezone
}

// SetTimezone sets the Timezone value and returns the updated Site.
func (m *Site) SetTimezone(v *string) {
	m.Timezone = v
}

// GetLanguage returns the Language value.
func (m Site) GetLanguage() *string {
	return m.Language
}

// SetLanguage sets the Language value and returns the updated Site.
func (m *Site) SetLanguage(v *string) {
	m.Language = v
}

// GetDateFormat returns the DateFormat value.
func (m Site) GetDateFormat() *string {
	return m.DateFormat
}

// SetDateFormat sets the DateFormat value and returns the updated Site.
func (m *Site) SetDateFormat(v *string) {
	m.DateFormat = v
}

// GetTimeFormat returns the TimeFormat value.
func (m Site) GetTimeFormat() *string {
	return m.TimeFormat
}

// SetTimeFormat sets the TimeFormat value and returns the updated Site.
func (m *Site) SetTimeFormat(v *string) {
	m.TimeFormat = v
}

// GetWeekStartDay returns the WeekStartDay value.
func (m Site) GetWeekStartDay() *string {
	return m.WeekStartDay
}

// SetWeekStartDay sets the WeekStartDay value and returns the updated Site.
func (m *Site) SetWeekStartDay(v *string) {
	m.WeekStartDay = v
}

// GetNumberFormat returns the NumberFormat value.
func (m Site) GetNumberFormat() *string {
	return m.NumberFormat
}

// SetNumberFormat sets the NumberFormat value and returns the updated Site.
func (m *Site) SetNumberFormat(v *string) {
	m.NumberFormat = v
}

// GetBolusDiaHours returns the BolusDiaHours value.
func (m Site) GetBolusDiaHours() *float64 {
	return m.BolusDiaHours
}

// SetBolusDiaHours sets the BolusDiaHours value and returns the updated Site.
func (m *Site) SetBolusDiaHours(v *float64) {
	m.BolusDiaHours = v
}

// GetBolusOnsetMinutes returns the BolusOnsetMinutes value.
func (m Site) GetBolusOnsetMinutes() *float64 {
	return m.BolusOnsetMinutes
}

// SetBolusOnsetMinutes sets the BolusOnsetMinutes value and returns the updated Site.
func (m *Site) SetBolusOnsetMinutes(v *float64) {
	m.BolusOnsetMinutes = v
}

// GetBolusPeakMinutes returns the BolusPeakMinutes value.
func (m Site) GetBolusPeakMinutes() *float64 {
	return m.BolusPeakMinutes
}

// SetBolusPeakMinutes sets the BolusPeakMinutes value and returns the updated Site.
func (m *Site) SetBolusPeakMinutes(v *float64) {
	m.BolusPeakMinutes = v
}

// GetBasalDiaHours returns the BasalDiaHours value.
func (m Site) GetBasalDiaHours() *float64 {
	return m.BasalDiaHours
}

// SetBasalDiaHours sets the BasalDiaHours value and returns the updated Site.
func (m *Site) SetBasalDiaHours(v *float64) {
	m.BasalDiaHours = v
}

// GetBasalOnsetMinutes returns the BasalOnsetMinutes value.
func (m Site) GetBasalOnsetMinutes() *float64 {
	return m.BasalOnsetMinutes
}

// SetBasalOnsetMinutes sets the BasalOnsetMinutes value and returns the updated Site.
func (m *Site) SetBasalOnsetMinutes(v *float64) {
	m.BasalOnsetMinutes = v
}

// GetBasalPeakMinutes returns the BasalPeakMinutes value.
func (m Site) GetBasalPeakMinutes() *float64 {
	return m.BasalPeakMinutes
}

// SetBasalPeakMinutes sets the BasalPeakMinutes value and returns the updated Site.
func (m *Site) SetBasalPeakMinutes(v *float64) {
	m.BasalPeakMinutes = v
}

// GetDefaultInsulinSensitivity returns the DefaultInsulinSensitivity value.
func (m Site) GetDefaultInsulinSensitivity() *float64 {
	return m.DefaultInsulinSensitivity
}

// SetDefaultInsulinSensitivity sets the DefaultInsulinSensitivity value and returns the updated Site.
func (m *Site) SetDefaultInsulinSensitivity(v *float64) {
	m.DefaultInsulinSensitivity = v
}

// GetDefaultCarbRatio returns the DefaultCarbRatio value.
func (m Site) GetDefaultCarbRatio() *float64 {
	return m.DefaultCarbRatio
}

// SetDefaultCarbRatio sets the DefaultCarbRatio value and returns the updated Site.
func (m *Site) SetDefaultCarbRatio(v *float64) {
	m.DefaultCarbRatio = v
}

// GetDefaultBasalRate returns the DefaultBasalRate value.
func (m Site) GetDefaultBasalRate() *float64 {
	return m.DefaultBasalRate
}

// SetDefaultBasalRate sets the DefaultBasalRate value and returns the updated Site.
func (m *Site) SetDefaultBasalRate(v *float64) {
	m.DefaultBasalRate = v
}

// GetCarbMaxAbsorptionHours returns the CarbMaxAbsorptionHours value.
func (m Site) GetCarbMaxAbsorptionHours() *float64 {
	return m.CarbMaxAbsorptionHours
}

// SetCarbMaxAbsorptionHours sets the CarbMaxAbsorptionHours value and returns the updated Site.
func (m *Site) SetCarbMaxAbsorptionHours(v *float64) {
	m.CarbMaxAbsorptionHours = v
}

// GetCarbMin5mImpact returns the CarbMin5mImpact value.
func (m Site) GetCarbMin5mImpact() *float64 {
	return m.CarbMin5mImpact
}

// SetCarbMin5mImpact sets the CarbMin5mImpact value and returns the updated Site.
func (m *Site) SetCarbMin5mImpact(v *float64) {
	m.CarbMin5mImpact = v
}

// GetAlarmUrgentLowEnabled returns the AlarmUrgentLowEnabled value.
func (m Site) GetAlarmUrgentLowEnabled() *bool {
	return m.AlarmUrgentLowEnabled
}

// SetAlarmUrgentLowEnabled sets the AlarmUrgentLowEnabled value and returns the updated Site.
func (m *Site) SetAlarmUrgentLowEnabled(v *bool) {
	m.AlarmUrgentLowEnabled = v
}

// GetRangeUrgentLowMgdl returns the RangeUrgentLowMgdl value.
func (m Site) GetRangeUrgentLowMgdl() *int64 {
	return m.RangeUrgentLowMgdl
}

// SetRangeUrgentLowMgdl sets the RangeUrgentLowMgdl value and returns the updated Site.
func (m *Site) SetRangeUrgentLowMgdl(v *int64) {
	m.RangeUrgentLowMgdl = v
}

// GetAlarmLowEnabled returns the AlarmLowEnabled value.
func (m Site) GetAlarmLowEnabled() *bool {
	return m.AlarmLowEnabled
}

// SetAlarmLowEnabled sets the AlarmLowEnabled value and returns the updated Site.
func (m *Site) SetAlarmLowEnabled(v *bool) {
	m.AlarmLowEnabled = v
}

// GetRangeLowMgdl returns the RangeLowMgdl value.
func (m Site) GetRangeLowMgdl() *int64 {
	return m.RangeLowMgdl
}

// SetRangeLowMgdl sets the RangeLowMgdl value and returns the updated Site.
func (m *Site) SetRangeLowMgdl(v *int64) {
	m.RangeLowMgdl = v
}

// GetAlarmLowMinutes returns the AlarmLowMinutes value.
func (m Site) GetAlarmLowMinutes() *int64 {
	return m.AlarmLowMinutes
}

// SetAlarmLowMinutes sets the AlarmLowMinutes value and returns the updated Site.
func (m *Site) SetAlarmLowMinutes(v *int64) {
	m.AlarmLowMinutes = v
}

// GetAlarmLowPredictionEnabled returns the AlarmLowPredictionEnabled value.
func (m Site) GetAlarmLowPredictionEnabled() *bool {
	return m.AlarmLowPredictionEnabled
}

// SetAlarmLowPredictionEnabled sets the AlarmLowPredictionEnabled value and returns the updated Site.
func (m *Site) SetAlarmLowPredictionEnabled(v *bool) {
	m.AlarmLowPredictionEnabled = v
}

// GetAlarmLowPredictionMinutes returns the AlarmLowPredictionMinutes value.
func (m Site) GetAlarmLowPredictionMinutes() *int64 {
	return m.AlarmLowPredictionMinutes
}

// SetAlarmLowPredictionMinutes sets the AlarmLowPredictionMinutes value and returns the updated Site.
func (m *Site) SetAlarmLowPredictionMinutes(v *int64) {
	m.AlarmLowPredictionMinutes = v
}

// GetAlarmHighEnabled returns the AlarmHighEnabled value.
func (m Site) GetAlarmHighEnabled() *bool {
	return m.AlarmHighEnabled
}

// SetAlarmHighEnabled sets the AlarmHighEnabled value and returns the updated Site.
func (m *Site) SetAlarmHighEnabled(v *bool) {
	m.AlarmHighEnabled = v
}

// GetRangeHighMgdl returns the RangeHighMgdl value.
func (m Site) GetRangeHighMgdl() *int64 {
	return m.RangeHighMgdl
}

// SetRangeHighMgdl sets the RangeHighMgdl value and returns the updated Site.
func (m *Site) SetRangeHighMgdl(v *int64) {
	m.RangeHighMgdl = v
}

// GetAlarmHighMinutes returns the AlarmHighMinutes value.
func (m Site) GetAlarmHighMinutes() *int64 {
	return m.AlarmHighMinutes
}

// SetAlarmHighMinutes sets the AlarmHighMinutes value and returns the updated Site.
func (m *Site) SetAlarmHighMinutes(v *int64) {
	m.AlarmHighMinutes = v
}

// GetAlarmUrgentHighEnabled returns the AlarmUrgentHighEnabled value.
func (m Site) GetAlarmUrgentHighEnabled() *bool {
	return m.AlarmUrgentHighEnabled
}

// SetAlarmUrgentHighEnabled sets the AlarmUrgentHighEnabled value and returns the updated Site.
func (m *Site) SetAlarmUrgentHighEnabled(v *bool) {
	m.AlarmUrgentHighEnabled = v
}

// GetRangeUrgentHighMgdl returns the RangeUrgentHighMgdl value.
func (m Site) GetRangeUrgentHighMgdl() *int64 {
	return m.RangeUrgentHighMgdl
}

// SetRangeUrgentHighMgdl sets the RangeUrgentHighMgdl value and returns the updated Site.
func (m *Site) SetRangeUrgentHighMgdl(v *int64) {
	m.RangeUrgentHighMgdl = v
}

// GetAlarmFastRiseEnabled returns the AlarmFastRiseEnabled value.
func (m Site) GetAlarmFastRiseEnabled() *bool {
	return m.AlarmFastRiseEnabled
}

// SetAlarmFastRiseEnabled sets the AlarmFastRiseEnabled value and returns the updated Site.
func (m *Site) SetAlarmFastRiseEnabled(v *bool) {
	m.AlarmFastRiseEnabled = v
}

// GetAlarmFastRiseMgdl returns the AlarmFastRiseMgdl value.
func (m Site) GetAlarmFastRiseMgdl() *int64 {
	return m.AlarmFastRiseMgdl
}

// SetAlarmFastRiseMgdl sets the AlarmFastRiseMgdl value and returns the updated Site.
func (m *Site) SetAlarmFastRiseMgdl(v *int64) {
	m.AlarmFastRiseMgdl = v
}

// GetAlarmFastRiseMinutes returns the AlarmFastRiseMinutes value.
func (m Site) GetAlarmFastRiseMinutes() *int64 {
	return m.AlarmFastRiseMinutes
}

// SetAlarmFastRiseMinutes sets the AlarmFastRiseMinutes value and returns the updated Site.
func (m *Site) SetAlarmFastRiseMinutes(v *int64) {
	m.AlarmFastRiseMinutes = v
}

// GetAlarmFastDropEnabled returns the AlarmFastDropEnabled value.
func (m Site) GetAlarmFastDropEnabled() *bool {
	return m.AlarmFastDropEnabled
}

// SetAlarmFastDropEnabled sets the AlarmFastDropEnabled value and returns the updated Site.
func (m *Site) SetAlarmFastDropEnabled(v *bool) {
	m.AlarmFastDropEnabled = v
}

// GetAlarmFastDropMgdl returns the AlarmFastDropMgdl value.
func (m Site) GetAlarmFastDropMgdl() *int64 {
	return m.AlarmFastDropMgdl
}

// SetAlarmFastDropMgdl sets the AlarmFastDropMgdl value and returns the updated Site.
func (m *Site) SetAlarmFastDropMgdl(v *int64) {
	m.AlarmFastDropMgdl = v
}

// GetAlarmFastDropMinutes returns the AlarmFastDropMinutes value.
func (m Site) GetAlarmFastDropMinutes() *int64 {
	return m.AlarmFastDropMinutes
}

// SetAlarmFastDropMinutes sets the AlarmFastDropMinutes value and returns the updated Site.
func (m *Site) SetAlarmFastDropMinutes(v *int64) {
	m.AlarmFastDropMinutes = v
}

// GetAlarmNoDataEnabled returns the AlarmNoDataEnabled value.
func (m Site) GetAlarmNoDataEnabled() *bool {
	return m.AlarmNoDataEnabled
}

// SetAlarmNoDataEnabled sets the AlarmNoDataEnabled value and returns the updated Site.
func (m *Site) SetAlarmNoDataEnabled(v *bool) {
	m.AlarmNoDataEnabled = v
}

// GetAlarmNoDataMinutes returns the AlarmNoDataMinutes value.
func (m Site) GetAlarmNoDataMinutes() *int64 {
	return m.AlarmNoDataMinutes
}

// SetAlarmNoDataMinutes sets the AlarmNoDataMinutes value and returns the updated Site.
func (m *Site) SetAlarmNoDataMinutes(v *int64) {
	m.AlarmNoDataMinutes = v
}

// GetAlarmDeviceBatteryLowEnabled returns the AlarmDeviceBatteryLowEnabled value.
func (m Site) GetAlarmDeviceBatteryLowEnabled() *bool {
	return m.AlarmDeviceBatteryLowEnabled
}

// SetAlarmDeviceBatteryLowEnabled sets the AlarmDeviceBatteryLowEnabled value and returns the updated Site.
func (m *Site) SetAlarmDeviceBatteryLowEnabled(v *bool) {
	m.AlarmDeviceBatteryLowEnabled = v
}

// GetAlarmDeviceBatteryLowPercent returns the AlarmDeviceBatteryLowPercent value.
func (m Site) GetAlarmDeviceBatteryLowPercent() *int64 {
	return m.AlarmDeviceBatteryLowPercent
}

// SetAlarmDeviceBatteryLowPercent sets the AlarmDeviceBatteryLowPercent value and returns the updated Site.
func (m *Site) SetAlarmDeviceBatteryLowPercent(v *int64) {
	m.AlarmDeviceBatteryLowPercent = v
}

// GetAlarmDeviceBatteryLowMinutes returns the AlarmDeviceBatteryLowMinutes value.
func (m Site) GetAlarmDeviceBatteryLowMinutes() *int64 {
	return m.AlarmDeviceBatteryLowMinutes
}

// SetAlarmDeviceBatteryLowMinutes sets the AlarmDeviceBatteryLowMinutes value and returns the updated Site.
func (m *Site) SetAlarmDeviceBatteryLowMinutes(v *int64) {
	m.AlarmDeviceBatteryLowMinutes = v
}

// GetAlarmDeviceBatteryLowDevicesJson returns the AlarmDeviceBatteryLowDevicesJson value.
func (m Site) GetAlarmDeviceBatteryLowDevicesJson() *types.DeviceNames {
	return m.AlarmDeviceBatteryLowDevicesJson
}

// SetAlarmDeviceBatteryLowDevicesJson sets the AlarmDeviceBatteryLowDevicesJson value and returns the updated Site.
func (m *Site) SetAlarmDeviceBatteryLowDevicesJson(v *types.DeviceNames) {
	m.AlarmDeviceBatteryLowDevicesJson = v
}

// GetAlarmLowReservoirEnabled returns the AlarmLowReservoirEnabled value.
func (m Site) GetAlarmLowReservoirEnabled() *bool {
	return m.AlarmLowReservoirEnabled
}

// SetAlarmLowReservoirEnabled sets the AlarmLowReservoirEnabled value and returns the updated Site.
func (m *Site) SetAlarmLowReservoirEnabled(v *bool) {
	m.AlarmLowReservoirEnabled = v
}

// GetAlarmLowReservoirUnits returns the AlarmLowReservoirUnits value.
func (m Site) GetAlarmLowReservoirUnits() *int64 {
	return m.AlarmLowReservoirUnits
}

// SetAlarmLowReservoirUnits sets the AlarmLowReservoirUnits value and returns the updated Site.
func (m *Site) SetAlarmLowReservoirUnits(v *int64) {
	m.AlarmLowReservoirUnits = v
}

// GetNotificationsEnabled returns the NotificationsEnabled value.
func (m Site) GetNotificationsEnabled() *bool {
	return m.NotificationsEnabled
}

// SetNotificationsEnabled sets the NotificationsEnabled value and returns the updated Site.
func (m *Site) SetNotificationsEnabled(v *bool) {
	m.NotificationsEnabled = v
}

// GetType returns the Type value.
func (m SiteConnector) GetType() consts.ConnectorType {
	return m.Type
}

// GetCreatedAt returns the CreatedAt value.
func (m SiteConnector) GetCreatedAt() int64 {
	return m.CreatedAt
}

// SetCreatedAt sets the CreatedAt value and returns the updated SiteConnector.
func (m *SiteConnector) SetCreatedAt(v int64) {
	m.CreatedAt = v
}

// GetUpdatedAt returns the UpdatedAt value.
func (m SiteConnector) GetUpdatedAt() int64 {
	return m.UpdatedAt
}

// SetUpdatedAt sets the UpdatedAt value and returns the updated SiteConnector.
func (m *SiteConnector) SetUpdatedAt(v int64) {
	m.UpdatedAt = v
}

// GetID returns the ID value.
func (m SiteIdentity) GetID() int64 {
	return m.ID
}

// SetID sets the ID value and returns the updated SiteIdentity.
func (m *SiteIdentity) SetID(v int64) {
	m.ID = v
}

// GetName returns the Name value.
func (m SiteIdentity) GetName() string {
	return m.Name
}

// GetCreatedAt returns the CreatedAt value.
func (m SiteIdentity) GetCreatedAt() int64 {
	return m.CreatedAt
}

// SetCreatedAt sets the CreatedAt value and returns the updated SiteIdentity.
func (m *SiteIdentity) SetCreatedAt(v int64) {
	m.CreatedAt = v
}

// GetCreatedByIdentityName returns the CreatedByIdentityName value.
func (m SiteIdentity) GetCreatedByIdentityName() string {
	return m.CreatedByIdentityName
}

// SetCreatedByIdentityName sets the CreatedByIdentityName value and returns the updated SiteIdentity.
func (m *SiteIdentity) SetCreatedByIdentityName(v string) {
	m.CreatedByIdentityName = v
}

// GetCredentialsChangedAt returns the CredentialsChangedAt value.
func (m SiteIdentity) GetCredentialsChangedAt() int64 {
	return m.CredentialsChangedAt
}

// GetUnits returns the Units value.
func (m SiteIdentity) GetUnits() *consts.Units {
	return m.Units
}

// SetUnits sets the Units value and returns the updated SiteIdentity.
func (m *SiteIdentity) SetUnits(v *consts.Units) {
	m.Units = v
}

// GetUnitsMgdlLabel returns the UnitsMgdlLabel value.
func (m SiteIdentity) GetUnitsMgdlLabel() *string {
	return m.UnitsMgdlLabel
}

// SetUnitsMgdlLabel sets the UnitsMgdlLabel value and returns the updated SiteIdentity.
func (m *SiteIdentity) SetUnitsMgdlLabel(v *string) {
	m.UnitsMgdlLabel = v
}

// GetUnitsMmolLabel returns the UnitsMmolLabel value.
func (m SiteIdentity) GetUnitsMmolLabel() *string {
	return m.UnitsMmolLabel
}

// SetUnitsMmolLabel sets the UnitsMmolLabel value and returns the updated SiteIdentity.
func (m *SiteIdentity) SetUnitsMmolLabel(v *string) {
	m.UnitsMmolLabel = v
}

// GetTimezone returns the Timezone value.
func (m SiteIdentity) GetTimezone() *string {
	return m.Timezone
}

// SetTimezone sets the Timezone value and returns the updated SiteIdentity.
func (m *SiteIdentity) SetTimezone(v *string) {
	m.Timezone = v
}

// GetLanguage returns the Language value.
func (m SiteIdentity) GetLanguage() *string {
	return m.Language
}

// SetLanguage sets the Language value and returns the updated SiteIdentity.
func (m *SiteIdentity) SetLanguage(v *string) {
	m.Language = v
}

// GetDateFormat returns the DateFormat value.
func (m SiteIdentity) GetDateFormat() *string {
	return m.DateFormat
}

// SetDateFormat sets the DateFormat value and returns the updated SiteIdentity.
func (m *SiteIdentity) SetDateFormat(v *string) {
	m.DateFormat = v
}

// GetTimeFormat returns the TimeFormat value.
func (m SiteIdentity) GetTimeFormat() *string {
	return m.TimeFormat
}

// SetTimeFormat sets the TimeFormat value and returns the updated SiteIdentity.
func (m *SiteIdentity) SetTimeFormat(v *string) {
	m.TimeFormat = v
}

// GetWeekStartDay returns the WeekStartDay value.
func (m SiteIdentity) GetWeekStartDay() *string {
	return m.WeekStartDay
}

// SetWeekStartDay sets the WeekStartDay value and returns the updated SiteIdentity.
func (m *SiteIdentity) SetWeekStartDay(v *string) {
	m.WeekStartDay = v
}

// GetNumberFormat returns the NumberFormat value.
func (m SiteIdentity) GetNumberFormat() *string {
	return m.NumberFormat
}

// SetNumberFormat sets the NumberFormat value and returns the updated SiteIdentity.
func (m *SiteIdentity) SetNumberFormat(v *string) {
	m.NumberFormat = v
}

// GetBolusDiaHours returns the BolusDiaHours value.
func (m SiteIdentity) GetBolusDiaHours() *float64 {
	return m.BolusDiaHours
}

// SetBolusDiaHours sets the BolusDiaHours value and returns the updated SiteIdentity.
func (m *SiteIdentity) SetBolusDiaHours(v *float64) {
	m.BolusDiaHours = v
}

// GetBolusOnsetMinutes returns the BolusOnsetMinutes value.
func (m SiteIdentity) GetBolusOnsetMinutes() *float64 {
	return m.BolusOnsetMinutes
}

// SetBolusOnsetMinutes sets the BolusOnsetMinutes value and returns the updated SiteIdentity.
func (m *SiteIdentity) SetBolusOnsetMinutes(v *float64) {
	m.BolusOnsetMinutes = v
}

// GetBolusPeakMinutes returns the BolusPeakMinutes value.
func (m SiteIdentity) GetBolusPeakMinutes() *float64 {
	return m.BolusPeakMinutes
}

// SetBolusPeakMinutes sets the BolusPeakMinutes value and returns the updated SiteIdentity.
func (m *SiteIdentity) SetBolusPeakMinutes(v *float64) {
	m.BolusPeakMinutes = v
}

// GetBasalDiaHours returns the BasalDiaHours value.
func (m SiteIdentity) GetBasalDiaHours() *float64 {
	return m.BasalDiaHours
}

// SetBasalDiaHours sets the BasalDiaHours value and returns the updated SiteIdentity.
func (m *SiteIdentity) SetBasalDiaHours(v *float64) {
	m.BasalDiaHours = v
}

// GetBasalOnsetMinutes returns the BasalOnsetMinutes value.
func (m SiteIdentity) GetBasalOnsetMinutes() *float64 {
	return m.BasalOnsetMinutes
}

// SetBasalOnsetMinutes sets the BasalOnsetMinutes value and returns the updated SiteIdentity.
func (m *SiteIdentity) SetBasalOnsetMinutes(v *float64) {
	m.BasalOnsetMinutes = v
}

// GetBasalPeakMinutes returns the BasalPeakMinutes value.
func (m SiteIdentity) GetBasalPeakMinutes() *float64 {
	return m.BasalPeakMinutes
}

// SetBasalPeakMinutes sets the BasalPeakMinutes value and returns the updated SiteIdentity.
func (m *SiteIdentity) SetBasalPeakMinutes(v *float64) {
	m.BasalPeakMinutes = v
}

// GetDefaultInsulinSensitivity returns the DefaultInsulinSensitivity value.
func (m SiteIdentity) GetDefaultInsulinSensitivity() *float64 {
	return m.DefaultInsulinSensitivity
}

// SetDefaultInsulinSensitivity sets the DefaultInsulinSensitivity value and returns the updated SiteIdentity.
func (m *SiteIdentity) SetDefaultInsulinSensitivity(v *float64) {
	m.DefaultInsulinSensitivity = v
}

// GetDefaultCarbRatio returns the DefaultCarbRatio value.
func (m SiteIdentity) GetDefaultCarbRatio() *float64 {
	return m.DefaultCarbRatio
}

// SetDefaultCarbRatio sets the DefaultCarbRatio value and returns the updated SiteIdentity.
func (m *SiteIdentity) SetDefaultCarbRatio(v *float64) {
	m.DefaultCarbRatio = v
}

// GetDefaultBasalRate returns the DefaultBasalRate value.
func (m SiteIdentity) GetDefaultBasalRate() *float64 {
	return m.DefaultBasalRate
}

// SetDefaultBasalRate sets the DefaultBasalRate value and returns the updated SiteIdentity.
func (m *SiteIdentity) SetDefaultBasalRate(v *float64) {
	m.DefaultBasalRate = v
}

// GetCarbMaxAbsorptionHours returns the CarbMaxAbsorptionHours value.
func (m SiteIdentity) GetCarbMaxAbsorptionHours() *float64 {
	return m.CarbMaxAbsorptionHours
}

// SetCarbMaxAbsorptionHours sets the CarbMaxAbsorptionHours value and returns the updated SiteIdentity.
func (m *SiteIdentity) SetCarbMaxAbsorptionHours(v *float64) {
	m.CarbMaxAbsorptionHours = v
}

// GetCarbMin5mImpact returns the CarbMin5mImpact value.
func (m SiteIdentity) GetCarbMin5mImpact() *float64 {
	return m.CarbMin5mImpact
}

// SetCarbMin5mImpact sets the CarbMin5mImpact value and returns the updated SiteIdentity.
func (m *SiteIdentity) SetCarbMin5mImpact(v *float64) {
	m.CarbMin5mImpact = v
}

// GetAlarmUrgentLowEnabled returns the AlarmUrgentLowEnabled value.
func (m SiteIdentity) GetAlarmUrgentLowEnabled() *bool {
	return m.AlarmUrgentLowEnabled
}

// SetAlarmUrgentLowEnabled sets the AlarmUrgentLowEnabled value and returns the updated SiteIdentity.
func (m *SiteIdentity) SetAlarmUrgentLowEnabled(v *bool) {
	m.AlarmUrgentLowEnabled = v
}

// GetRangeUrgentLowMgdl returns the RangeUrgentLowMgdl value.
func (m SiteIdentity) GetRangeUrgentLowMgdl() *int64 {
	return m.RangeUrgentLowMgdl
}

// SetRangeUrgentLowMgdl sets the RangeUrgentLowMgdl value and returns the updated SiteIdentity.
func (m *SiteIdentity) SetRangeUrgentLowMgdl(v *int64) {
	m.RangeUrgentLowMgdl = v
}

// GetAlarmLowEnabled returns the AlarmLowEnabled value.
func (m SiteIdentity) GetAlarmLowEnabled() *bool {
	return m.AlarmLowEnabled
}

// SetAlarmLowEnabled sets the AlarmLowEnabled value and returns the updated SiteIdentity.
func (m *SiteIdentity) SetAlarmLowEnabled(v *bool) {
	m.AlarmLowEnabled = v
}

// GetRangeLowMgdl returns the RangeLowMgdl value.
func (m SiteIdentity) GetRangeLowMgdl() *int64 {
	return m.RangeLowMgdl
}

// SetRangeLowMgdl sets the RangeLowMgdl value and returns the updated SiteIdentity.
func (m *SiteIdentity) SetRangeLowMgdl(v *int64) {
	m.RangeLowMgdl = v
}

// GetAlarmLowMinutes returns the AlarmLowMinutes value.
func (m SiteIdentity) GetAlarmLowMinutes() *int64 {
	return m.AlarmLowMinutes
}

// SetAlarmLowMinutes sets the AlarmLowMinutes value and returns the updated SiteIdentity.
func (m *SiteIdentity) SetAlarmLowMinutes(v *int64) {
	m.AlarmLowMinutes = v
}

// GetAlarmLowPredictionEnabled returns the AlarmLowPredictionEnabled value.
func (m SiteIdentity) GetAlarmLowPredictionEnabled() *bool {
	return m.AlarmLowPredictionEnabled
}

// SetAlarmLowPredictionEnabled sets the AlarmLowPredictionEnabled value and returns the updated SiteIdentity.
func (m *SiteIdentity) SetAlarmLowPredictionEnabled(v *bool) {
	m.AlarmLowPredictionEnabled = v
}

// GetAlarmLowPredictionMinutes returns the AlarmLowPredictionMinutes value.
func (m SiteIdentity) GetAlarmLowPredictionMinutes() *int64 {
	return m.AlarmLowPredictionMinutes
}

// SetAlarmLowPredictionMinutes sets the AlarmLowPredictionMinutes value and returns the updated SiteIdentity.
func (m *SiteIdentity) SetAlarmLowPredictionMinutes(v *int64) {
	m.AlarmLowPredictionMinutes = v
}

// GetAlarmHighEnabled returns the AlarmHighEnabled value.
func (m SiteIdentity) GetAlarmHighEnabled() *bool {
	return m.AlarmHighEnabled
}

// SetAlarmHighEnabled sets the AlarmHighEnabled value and returns the updated SiteIdentity.
func (m *SiteIdentity) SetAlarmHighEnabled(v *bool) {
	m.AlarmHighEnabled = v
}

// GetRangeHighMgdl returns the RangeHighMgdl value.
func (m SiteIdentity) GetRangeHighMgdl() *int64 {
	return m.RangeHighMgdl
}

// SetRangeHighMgdl sets the RangeHighMgdl value and returns the updated SiteIdentity.
func (m *SiteIdentity) SetRangeHighMgdl(v *int64) {
	m.RangeHighMgdl = v
}

// GetAlarmHighMinutes returns the AlarmHighMinutes value.
func (m SiteIdentity) GetAlarmHighMinutes() *int64 {
	return m.AlarmHighMinutes
}

// SetAlarmHighMinutes sets the AlarmHighMinutes value and returns the updated SiteIdentity.
func (m *SiteIdentity) SetAlarmHighMinutes(v *int64) {
	m.AlarmHighMinutes = v
}

// GetAlarmUrgentHighEnabled returns the AlarmUrgentHighEnabled value.
func (m SiteIdentity) GetAlarmUrgentHighEnabled() *bool {
	return m.AlarmUrgentHighEnabled
}

// SetAlarmUrgentHighEnabled sets the AlarmUrgentHighEnabled value and returns the updated SiteIdentity.
func (m *SiteIdentity) SetAlarmUrgentHighEnabled(v *bool) {
	m.AlarmUrgentHighEnabled = v
}

// GetRangeUrgentHighMgdl returns the RangeUrgentHighMgdl value.
func (m SiteIdentity) GetRangeUrgentHighMgdl() *int64 {
	return m.RangeUrgentHighMgdl
}

// SetRangeUrgentHighMgdl sets the RangeUrgentHighMgdl value and returns the updated SiteIdentity.
func (m *SiteIdentity) SetRangeUrgentHighMgdl(v *int64) {
	m.RangeUrgentHighMgdl = v
}

// GetAlarmFastRiseEnabled returns the AlarmFastRiseEnabled value.
func (m SiteIdentity) GetAlarmFastRiseEnabled() *bool {
	return m.AlarmFastRiseEnabled
}

// SetAlarmFastRiseEnabled sets the AlarmFastRiseEnabled value and returns the updated SiteIdentity.
func (m *SiteIdentity) SetAlarmFastRiseEnabled(v *bool) {
	m.AlarmFastRiseEnabled = v
}

// GetAlarmFastRiseMgdl returns the AlarmFastRiseMgdl value.
func (m SiteIdentity) GetAlarmFastRiseMgdl() *int64 {
	return m.AlarmFastRiseMgdl
}

// SetAlarmFastRiseMgdl sets the AlarmFastRiseMgdl value and returns the updated SiteIdentity.
func (m *SiteIdentity) SetAlarmFastRiseMgdl(v *int64) {
	m.AlarmFastRiseMgdl = v
}

// GetAlarmFastRiseMinutes returns the AlarmFastRiseMinutes value.
func (m SiteIdentity) GetAlarmFastRiseMinutes() *int64 {
	return m.AlarmFastRiseMinutes
}

// SetAlarmFastRiseMinutes sets the AlarmFastRiseMinutes value and returns the updated SiteIdentity.
func (m *SiteIdentity) SetAlarmFastRiseMinutes(v *int64) {
	m.AlarmFastRiseMinutes = v
}

// GetAlarmFastDropEnabled returns the AlarmFastDropEnabled value.
func (m SiteIdentity) GetAlarmFastDropEnabled() *bool {
	return m.AlarmFastDropEnabled
}

// SetAlarmFastDropEnabled sets the AlarmFastDropEnabled value and returns the updated SiteIdentity.
func (m *SiteIdentity) SetAlarmFastDropEnabled(v *bool) {
	m.AlarmFastDropEnabled = v
}

// GetAlarmFastDropMgdl returns the AlarmFastDropMgdl value.
func (m SiteIdentity) GetAlarmFastDropMgdl() *int64 {
	return m.AlarmFastDropMgdl
}

// SetAlarmFastDropMgdl sets the AlarmFastDropMgdl value and returns the updated SiteIdentity.
func (m *SiteIdentity) SetAlarmFastDropMgdl(v *int64) {
	m.AlarmFastDropMgdl = v
}

// GetAlarmFastDropMinutes returns the AlarmFastDropMinutes value.
func (m SiteIdentity) GetAlarmFastDropMinutes() *int64 {
	return m.AlarmFastDropMinutes
}

// SetAlarmFastDropMinutes sets the AlarmFastDropMinutes value and returns the updated SiteIdentity.
func (m *SiteIdentity) SetAlarmFastDropMinutes(v *int64) {
	m.AlarmFastDropMinutes = v
}

// GetAlarmNoDataEnabled returns the AlarmNoDataEnabled value.
func (m SiteIdentity) GetAlarmNoDataEnabled() *bool {
	return m.AlarmNoDataEnabled
}

// SetAlarmNoDataEnabled sets the AlarmNoDataEnabled value and returns the updated SiteIdentity.
func (m *SiteIdentity) SetAlarmNoDataEnabled(v *bool) {
	m.AlarmNoDataEnabled = v
}

// GetAlarmNoDataMinutes returns the AlarmNoDataMinutes value.
func (m SiteIdentity) GetAlarmNoDataMinutes() *int64 {
	return m.AlarmNoDataMinutes
}

// SetAlarmNoDataMinutes sets the AlarmNoDataMinutes value and returns the updated SiteIdentity.
func (m *SiteIdentity) SetAlarmNoDataMinutes(v *int64) {
	m.AlarmNoDataMinutes = v
}

// GetAlarmDeviceBatteryLowEnabled returns the AlarmDeviceBatteryLowEnabled value.
func (m SiteIdentity) GetAlarmDeviceBatteryLowEnabled() *bool {
	return m.AlarmDeviceBatteryLowEnabled
}

// SetAlarmDeviceBatteryLowEnabled sets the AlarmDeviceBatteryLowEnabled value and returns the updated SiteIdentity.
func (m *SiteIdentity) SetAlarmDeviceBatteryLowEnabled(v *bool) {
	m.AlarmDeviceBatteryLowEnabled = v
}

// GetAlarmDeviceBatteryLowPercent returns the AlarmDeviceBatteryLowPercent value.
func (m SiteIdentity) GetAlarmDeviceBatteryLowPercent() *int64 {
	return m.AlarmDeviceBatteryLowPercent
}

// SetAlarmDeviceBatteryLowPercent sets the AlarmDeviceBatteryLowPercent value and returns the updated SiteIdentity.
func (m *SiteIdentity) SetAlarmDeviceBatteryLowPercent(v *int64) {
	m.AlarmDeviceBatteryLowPercent = v
}

// GetAlarmDeviceBatteryLowMinutes returns the AlarmDeviceBatteryLowMinutes value.
func (m SiteIdentity) GetAlarmDeviceBatteryLowMinutes() *int64 {
	return m.AlarmDeviceBatteryLowMinutes
}

// SetAlarmDeviceBatteryLowMinutes sets the AlarmDeviceBatteryLowMinutes value and returns the updated SiteIdentity.
func (m *SiteIdentity) SetAlarmDeviceBatteryLowMinutes(v *int64) {
	m.AlarmDeviceBatteryLowMinutes = v
}

// GetAlarmDeviceBatteryLowDevicesJson returns the AlarmDeviceBatteryLowDevicesJson value.
func (m SiteIdentity) GetAlarmDeviceBatteryLowDevicesJson() *types.DeviceNames {
	return m.AlarmDeviceBatteryLowDevicesJson
}

// SetAlarmDeviceBatteryLowDevicesJson sets the AlarmDeviceBatteryLowDevicesJson value and returns the updated SiteIdentity.
func (m *SiteIdentity) SetAlarmDeviceBatteryLowDevicesJson(v *types.DeviceNames) {
	m.AlarmDeviceBatteryLowDevicesJson = v
}

// GetAlarmLowReservoirEnabled returns the AlarmLowReservoirEnabled value.
func (m SiteIdentity) GetAlarmLowReservoirEnabled() *bool {
	return m.AlarmLowReservoirEnabled
}

// SetAlarmLowReservoirEnabled sets the AlarmLowReservoirEnabled value and returns the updated SiteIdentity.
func (m *SiteIdentity) SetAlarmLowReservoirEnabled(v *bool) {
	m.AlarmLowReservoirEnabled = v
}

// GetAlarmLowReservoirUnits returns the AlarmLowReservoirUnits value.
func (m SiteIdentity) GetAlarmLowReservoirUnits() *int64 {
	return m.AlarmLowReservoirUnits
}

// SetAlarmLowReservoirUnits sets the AlarmLowReservoirUnits value and returns the updated SiteIdentity.
func (m *SiteIdentity) SetAlarmLowReservoirUnits(v *int64) {
	m.AlarmLowReservoirUnits = v
}

// GetNotificationsEnabled returns the NotificationsEnabled value.
func (m SiteIdentity) GetNotificationsEnabled() *bool {
	return m.NotificationsEnabled
}

// SetNotificationsEnabled sets the NotificationsEnabled value and returns the updated SiteIdentity.
func (m *SiteIdentity) SetNotificationsEnabled(v *bool) {
	m.NotificationsEnabled = v
}

// GetName returns the Name value.
func (m SqliteSequence) GetName() interface{} {
	return m.Name
}

// GetName returns the Name value.
func (m CreateControlIdentityParams) GetName() string {
	return m.Name
}

// GetCreatedByIdentityName returns the CreatedByIdentityName value.
func (m CreateControlIdentityParams) GetCreatedByIdentityName() string {
	return m.CreatedByIdentityName
}

// SetCreatedByIdentityName sets the CreatedByIdentityName value and returns the updated CreateControlIdentityParams.
func (m *CreateControlIdentityParams) SetCreatedByIdentityName(v string) {
	m.CreatedByIdentityName = v
}

// GetID returns the ID value.
func (m DeleteControlIdentityParams) GetID() int64 {
	return m.ID
}

// SetID sets the ID value and returns the updated DeleteControlIdentityParams.
func (m *DeleteControlIdentityParams) SetID(v int64) {
	m.ID = v
}

// GetName returns the Name value.
func (m GetControlIdentityParams) GetName() string {
	return m.Name
}

// GetID returns the ID value.
func (m GetControlIdentityByIDParams) GetID() int64 {
	return m.ID
}

// SetID sets the ID value and returns the updated GetControlIdentityByIDParams.
func (m *GetControlIdentityByIDParams) SetID(v int64) {
	m.ID = v
}

// GetID returns the ID value.
func (m UpdateControlIdentityPasswordParams) GetID() int64 {
	return m.ID
}

// SetID sets the ID value and returns the updated UpdateControlIdentityPasswordParams.
func (m *UpdateControlIdentityPasswordParams) SetID(v int64) {
	m.ID = v
}

// GetID returns the ID value.
func (m UpdateControlIdentityRoleParams) GetID() int64 {
	return m.ID
}

// SetID sets the ID value and returns the updated UpdateControlIdentityRoleParams.
func (m *UpdateControlIdentityRoleParams) SetID(v int64) {
	m.ID = v
}

// GetID returns the ID value.
func (m UpdateControlIdentityTOTPParams) GetID() int64 {
	return m.ID
}

// SetID sets the ID value and returns the updated UpdateControlIdentityTOTPParams.
func (m *UpdateControlIdentityTOTPParams) SetID(v int64) {
	m.ID = v
}

// GetIdentityID returns the IdentityID value.
func (m CreateSiteIdentityPasskeyParams) GetIdentityID() int64 {
	return m.IdentityID
}

// GetID returns the ID value.
func (m DeleteControlPasskeyParams) GetID() int64 {
	return m.ID
}

// SetID sets the ID value and returns the updated DeleteControlPasskeyParams.
func (m *DeleteControlPasskeyParams) SetID(v int64) {
	m.ID = v
}

// GetName returns the Name value.
func (m DeleteSiteAdminPasskeysParams) GetName() string {
	return m.Name
}

// GetIdentityID returns the IdentityID value.
func (m DeleteSiteIdentityPasskeyParams) GetIdentityID() int64 {
	return m.IdentityID
}

// GetID returns the ID value.
func (m DeleteSiteIdentityPasskeyParams) GetID() int64 {
	return m.ID
}

// SetID sets the ID value and returns the updated DeleteSiteIdentityPasskeyParams.
func (m *DeleteSiteIdentityPasskeyParams) SetID(v int64) {
	m.ID = v
}

// GetIdentityID returns the IdentityID value.
func (m DeleteSiteIdentityPasskeysParams) GetIdentityID() int64 {
	return m.IdentityID
}

// GetIdentityID returns the IdentityID value.
func (m ListSiteIdentityPasskeysParams) GetIdentityID() int64 {
	return m.IdentityID
}

// GetIdentityID returns the IdentityID value.
func (m CountPushRegistrationsParams) GetIdentityID() int64 {
	return m.IdentityID
}

// GetName returns the Name value.
func (m CreatePushCredentialParams) GetName() string {
	return m.Name
}

// GetCreatedByIdentityName returns the CreatedByIdentityName value.
func (m CreatePushCredentialParams) GetCreatedByIdentityName() string {
	return m.CreatedByIdentityName
}

// SetCreatedByIdentityName sets the CreatedByIdentityName value and returns the updated CreatePushCredentialParams.
func (m *CreatePushCredentialParams) SetCreatedByIdentityName(v string) {
	m.CreatedByIdentityName = v
}

// GetID returns the ID value.
func (m DeletePushCredentialParams) GetID() int64 {
	return m.ID
}

// SetID sets the ID value and returns the updated DeletePushCredentialParams.
func (m *DeletePushCredentialParams) SetID(v int64) {
	m.ID = v
}

// GetID returns the ID value.
func (m DeletePushRegistrationParams) GetID() int64 {
	return m.ID
}

// SetID sets the ID value and returns the updated DeletePushRegistrationParams.
func (m *DeletePushRegistrationParams) SetID(v int64) {
	m.ID = v
}

// GetIdentityID returns the IdentityID value.
func (m DeletePushRegistrationParams) GetIdentityID() *int64 {
	return m.IdentityID
}

// GetID returns the ID value.
func (m GetAnyPushCredentialParams) GetID() int64 {
	return m.ID
}

// SetID sets the ID value and returns the updated GetAnyPushCredentialParams.
func (m *GetAnyPushCredentialParams) SetID(v int64) {
	m.ID = v
}

// GetID returns the ID value.
func (m GetAnySitePushRegistrationParams) GetID() int64 {
	return m.ID
}

// SetID sets the ID value and returns the updated GetAnySitePushRegistrationParams.
func (m *GetAnySitePushRegistrationParams) SetID(v int64) {
	m.ID = v
}

// GetID returns the ID value.
func (m GetPushCredentialParams) GetID() int64 {
	return m.ID
}

// SetID sets the ID value and returns the updated GetPushCredentialParams.
func (m *GetPushCredentialParams) SetID(v int64) {
	m.ID = v
}

// GetID returns the ID value.
func (m GetPushRegistrationParams) GetID() int64 {
	return m.ID
}

// SetID sets the ID value and returns the updated GetPushRegistrationParams.
func (m *GetPushRegistrationParams) SetID(v int64) {
	m.ID = v
}

// GetIdentityID returns the IdentityID value.
func (m GetPushRegistrationParams) GetIdentityID() *int64 {
	return m.IdentityID
}

// GetApp returns the App value.
func (m NewestPushRegistrationForAppParams) GetApp() consts.PushApp {
	return m.App
}

// GetApp returns the App value.
func (m PushCredentialForParams) GetApp() string {
	return m.App
}

// GetID returns the ID value.
func (m RecordPushParams) GetID() int64 {
	return m.ID
}

// SetID sets the ID value and returns the updated RecordPushParams.
func (m *RecordPushParams) SetID(v int64) {
	m.ID = v
}

// SetLimit sets the Limit value and returns the updated SearchAllPushCredentialsParams.
func (m *SearchAllPushCredentialsParams) SetLimit(v int64) {
	m.Limit = v
}

// SetLimit sets the Limit value and returns the updated SearchAllPushRegistrationsParams.
func (m *SearchAllPushRegistrationsParams) SetLimit(v int64) {
	m.Limit = v
}

// SetLimit sets the Limit value and returns the updated SearchPushCredentialsParams.
func (m *SearchPushCredentialsParams) SetLimit(v int64) {
	m.Limit = v
}

// GetIdentityID returns the IdentityID value.
func (m SearchPushRegistrationsParams) GetIdentityID() *int64 {
	return m.IdentityID
}

// SetLimit sets the Limit value and returns the updated SearchPushRegistrationsParams.
func (m *SearchPushRegistrationsParams) SetLimit(v int64) {
	m.Limit = v
}

// GetIdentityID returns the IdentityID value.
func (m SearchPushRegistrationsWithIdentityParams) GetIdentityID() *int64 {
	return m.IdentityID
}

// SetLimit sets the Limit value and returns the updated SearchPushRegistrationsWithIdentityParams.
func (m *SearchPushRegistrationsWithIdentityParams) SetLimit(v int64) {
	m.Limit = v
}

// GetIdentityID returns the IdentityID value.
func (m UpsertPushRegistrationParams) GetIdentityID() int64 {
	return m.IdentityID
}

// GetApp returns the App value.
func (m UpsertPushRegistrationParams) GetApp() consts.PushApp {
	return m.App
}

// GetUnitsMgdlLabel returns the UnitsMgdlLabel value.
func (m LearnSiteIdentityUnitsLabelsParams) GetUnitsMgdlLabel() *string {
	return m.UnitsMgdlLabel
}

// SetUnitsMgdlLabel sets the UnitsMgdlLabel value and returns the updated LearnSiteIdentityUnitsLabelsParams.
func (m *LearnSiteIdentityUnitsLabelsParams) SetUnitsMgdlLabel(v *string) {
	m.UnitsMgdlLabel = v
}

// GetUnitsMmolLabel returns the UnitsMmolLabel value.
func (m LearnSiteIdentityUnitsLabelsParams) GetUnitsMmolLabel() *string {
	return m.UnitsMmolLabel
}

// SetUnitsMmolLabel sets the UnitsMmolLabel value and returns the updated LearnSiteIdentityUnitsLabelsParams.
func (m *LearnSiteIdentityUnitsLabelsParams) SetUnitsMmolLabel(v *string) {
	m.UnitsMmolLabel = v
}

// GetID returns the ID value.
func (m LearnSiteIdentityUnitsLabelsParams) GetID() int64 {
	return m.ID
}

// SetID sets the ID value and returns the updated LearnSiteIdentityUnitsLabelsParams.
func (m *LearnSiteIdentityUnitsLabelsParams) SetID(v int64) {
	m.ID = v
}

// GetUnits returns the Units value.
func (m UpdateSiteIdentitySettingsParams) GetUnits() *consts.Units {
	return m.Units
}

// SetUnits sets the Units value and returns the updated UpdateSiteIdentitySettingsParams.
func (m *UpdateSiteIdentitySettingsParams) SetUnits(v *consts.Units) {
	m.Units = v
}

// GetUnitsMgdlLabel returns the UnitsMgdlLabel value.
func (m UpdateSiteIdentitySettingsParams) GetUnitsMgdlLabel() *string {
	return m.UnitsMgdlLabel
}

// SetUnitsMgdlLabel sets the UnitsMgdlLabel value and returns the updated UpdateSiteIdentitySettingsParams.
func (m *UpdateSiteIdentitySettingsParams) SetUnitsMgdlLabel(v *string) {
	m.UnitsMgdlLabel = v
}

// GetUnitsMmolLabel returns the UnitsMmolLabel value.
func (m UpdateSiteIdentitySettingsParams) GetUnitsMmolLabel() *string {
	return m.UnitsMmolLabel
}

// SetUnitsMmolLabel sets the UnitsMmolLabel value and returns the updated UpdateSiteIdentitySettingsParams.
func (m *UpdateSiteIdentitySettingsParams) SetUnitsMmolLabel(v *string) {
	m.UnitsMmolLabel = v
}

// GetTimezone returns the Timezone value.
func (m UpdateSiteIdentitySettingsParams) GetTimezone() *string {
	return m.Timezone
}

// SetTimezone sets the Timezone value and returns the updated UpdateSiteIdentitySettingsParams.
func (m *UpdateSiteIdentitySettingsParams) SetTimezone(v *string) {
	m.Timezone = v
}

// GetLanguage returns the Language value.
func (m UpdateSiteIdentitySettingsParams) GetLanguage() *string {
	return m.Language
}

// SetLanguage sets the Language value and returns the updated UpdateSiteIdentitySettingsParams.
func (m *UpdateSiteIdentitySettingsParams) SetLanguage(v *string) {
	m.Language = v
}

// GetDateFormat returns the DateFormat value.
func (m UpdateSiteIdentitySettingsParams) GetDateFormat() *string {
	return m.DateFormat
}

// SetDateFormat sets the DateFormat value and returns the updated UpdateSiteIdentitySettingsParams.
func (m *UpdateSiteIdentitySettingsParams) SetDateFormat(v *string) {
	m.DateFormat = v
}

// GetTimeFormat returns the TimeFormat value.
func (m UpdateSiteIdentitySettingsParams) GetTimeFormat() *string {
	return m.TimeFormat
}

// SetTimeFormat sets the TimeFormat value and returns the updated UpdateSiteIdentitySettingsParams.
func (m *UpdateSiteIdentitySettingsParams) SetTimeFormat(v *string) {
	m.TimeFormat = v
}

// GetWeekStartDay returns the WeekStartDay value.
func (m UpdateSiteIdentitySettingsParams) GetWeekStartDay() *string {
	return m.WeekStartDay
}

// SetWeekStartDay sets the WeekStartDay value and returns the updated UpdateSiteIdentitySettingsParams.
func (m *UpdateSiteIdentitySettingsParams) SetWeekStartDay(v *string) {
	m.WeekStartDay = v
}

// GetNumberFormat returns the NumberFormat value.
func (m UpdateSiteIdentitySettingsParams) GetNumberFormat() *string {
	return m.NumberFormat
}

// SetNumberFormat sets the NumberFormat value and returns the updated UpdateSiteIdentitySettingsParams.
func (m *UpdateSiteIdentitySettingsParams) SetNumberFormat(v *string) {
	m.NumberFormat = v
}

// GetBolusDiaHours returns the BolusDiaHours value.
func (m UpdateSiteIdentitySettingsParams) GetBolusDiaHours() *float64 {
	return m.BolusDiaHours
}

// SetBolusDiaHours sets the BolusDiaHours value and returns the updated UpdateSiteIdentitySettingsParams.
func (m *UpdateSiteIdentitySettingsParams) SetBolusDiaHours(v *float64) {
	m.BolusDiaHours = v
}

// GetBolusOnsetMinutes returns the BolusOnsetMinutes value.
func (m UpdateSiteIdentitySettingsParams) GetBolusOnsetMinutes() *float64 {
	return m.BolusOnsetMinutes
}

// SetBolusOnsetMinutes sets the BolusOnsetMinutes value and returns the updated UpdateSiteIdentitySettingsParams.
func (m *UpdateSiteIdentitySettingsParams) SetBolusOnsetMinutes(v *float64) {
	m.BolusOnsetMinutes = v
}

// GetBolusPeakMinutes returns the BolusPeakMinutes value.
func (m UpdateSiteIdentitySettingsParams) GetBolusPeakMinutes() *float64 {
	return m.BolusPeakMinutes
}

// SetBolusPeakMinutes sets the BolusPeakMinutes value and returns the updated UpdateSiteIdentitySettingsParams.
func (m *UpdateSiteIdentitySettingsParams) SetBolusPeakMinutes(v *float64) {
	m.BolusPeakMinutes = v
}

// GetBasalDiaHours returns the BasalDiaHours value.
func (m UpdateSiteIdentitySettingsParams) GetBasalDiaHours() *float64 {
	return m.BasalDiaHours
}

// SetBasalDiaHours sets the BasalDiaHours value and returns the updated UpdateSiteIdentitySettingsParams.
func (m *UpdateSiteIdentitySettingsParams) SetBasalDiaHours(v *float64) {
	m.BasalDiaHours = v
}

// GetBasalOnsetMinutes returns the BasalOnsetMinutes value.
func (m UpdateSiteIdentitySettingsParams) GetBasalOnsetMinutes() *float64 {
	return m.BasalOnsetMinutes
}

// SetBasalOnsetMinutes sets the BasalOnsetMinutes value and returns the updated UpdateSiteIdentitySettingsParams.
func (m *UpdateSiteIdentitySettingsParams) SetBasalOnsetMinutes(v *float64) {
	m.BasalOnsetMinutes = v
}

// GetBasalPeakMinutes returns the BasalPeakMinutes value.
func (m UpdateSiteIdentitySettingsParams) GetBasalPeakMinutes() *float64 {
	return m.BasalPeakMinutes
}

// SetBasalPeakMinutes sets the BasalPeakMinutes value and returns the updated UpdateSiteIdentitySettingsParams.
func (m *UpdateSiteIdentitySettingsParams) SetBasalPeakMinutes(v *float64) {
	m.BasalPeakMinutes = v
}

// GetDefaultInsulinSensitivity returns the DefaultInsulinSensitivity value.
func (m UpdateSiteIdentitySettingsParams) GetDefaultInsulinSensitivity() *float64 {
	return m.DefaultInsulinSensitivity
}

// SetDefaultInsulinSensitivity sets the DefaultInsulinSensitivity value and returns the updated UpdateSiteIdentitySettingsParams.
func (m *UpdateSiteIdentitySettingsParams) SetDefaultInsulinSensitivity(v *float64) {
	m.DefaultInsulinSensitivity = v
}

// GetDefaultCarbRatio returns the DefaultCarbRatio value.
func (m UpdateSiteIdentitySettingsParams) GetDefaultCarbRatio() *float64 {
	return m.DefaultCarbRatio
}

// SetDefaultCarbRatio sets the DefaultCarbRatio value and returns the updated UpdateSiteIdentitySettingsParams.
func (m *UpdateSiteIdentitySettingsParams) SetDefaultCarbRatio(v *float64) {
	m.DefaultCarbRatio = v
}

// GetDefaultBasalRate returns the DefaultBasalRate value.
func (m UpdateSiteIdentitySettingsParams) GetDefaultBasalRate() *float64 {
	return m.DefaultBasalRate
}

// SetDefaultBasalRate sets the DefaultBasalRate value and returns the updated UpdateSiteIdentitySettingsParams.
func (m *UpdateSiteIdentitySettingsParams) SetDefaultBasalRate(v *float64) {
	m.DefaultBasalRate = v
}

// GetCarbMaxAbsorptionHours returns the CarbMaxAbsorptionHours value.
func (m UpdateSiteIdentitySettingsParams) GetCarbMaxAbsorptionHours() *float64 {
	return m.CarbMaxAbsorptionHours
}

// SetCarbMaxAbsorptionHours sets the CarbMaxAbsorptionHours value and returns the updated UpdateSiteIdentitySettingsParams.
func (m *UpdateSiteIdentitySettingsParams) SetCarbMaxAbsorptionHours(v *float64) {
	m.CarbMaxAbsorptionHours = v
}

// GetCarbMin5mImpact returns the CarbMin5mImpact value.
func (m UpdateSiteIdentitySettingsParams) GetCarbMin5mImpact() *float64 {
	return m.CarbMin5mImpact
}

// SetCarbMin5mImpact sets the CarbMin5mImpact value and returns the updated UpdateSiteIdentitySettingsParams.
func (m *UpdateSiteIdentitySettingsParams) SetCarbMin5mImpact(v *float64) {
	m.CarbMin5mImpact = v
}

// GetAlarmUrgentLowEnabled returns the AlarmUrgentLowEnabled value.
func (m UpdateSiteIdentitySettingsParams) GetAlarmUrgentLowEnabled() *bool {
	return m.AlarmUrgentLowEnabled
}

// SetAlarmUrgentLowEnabled sets the AlarmUrgentLowEnabled value and returns the updated UpdateSiteIdentitySettingsParams.
func (m *UpdateSiteIdentitySettingsParams) SetAlarmUrgentLowEnabled(v *bool) {
	m.AlarmUrgentLowEnabled = v
}

// GetRangeUrgentLowMgdl returns the RangeUrgentLowMgdl value.
func (m UpdateSiteIdentitySettingsParams) GetRangeUrgentLowMgdl() *int64 {
	return m.RangeUrgentLowMgdl
}

// SetRangeUrgentLowMgdl sets the RangeUrgentLowMgdl value and returns the updated UpdateSiteIdentitySettingsParams.
func (m *UpdateSiteIdentitySettingsParams) SetRangeUrgentLowMgdl(v *int64) {
	m.RangeUrgentLowMgdl = v
}

// GetAlarmLowEnabled returns the AlarmLowEnabled value.
func (m UpdateSiteIdentitySettingsParams) GetAlarmLowEnabled() *bool {
	return m.AlarmLowEnabled
}

// SetAlarmLowEnabled sets the AlarmLowEnabled value and returns the updated UpdateSiteIdentitySettingsParams.
func (m *UpdateSiteIdentitySettingsParams) SetAlarmLowEnabled(v *bool) {
	m.AlarmLowEnabled = v
}

// GetRangeLowMgdl returns the RangeLowMgdl value.
func (m UpdateSiteIdentitySettingsParams) GetRangeLowMgdl() *int64 {
	return m.RangeLowMgdl
}

// SetRangeLowMgdl sets the RangeLowMgdl value and returns the updated UpdateSiteIdentitySettingsParams.
func (m *UpdateSiteIdentitySettingsParams) SetRangeLowMgdl(v *int64) {
	m.RangeLowMgdl = v
}

// GetAlarmLowMinutes returns the AlarmLowMinutes value.
func (m UpdateSiteIdentitySettingsParams) GetAlarmLowMinutes() *int64 {
	return m.AlarmLowMinutes
}

// SetAlarmLowMinutes sets the AlarmLowMinutes value and returns the updated UpdateSiteIdentitySettingsParams.
func (m *UpdateSiteIdentitySettingsParams) SetAlarmLowMinutes(v *int64) {
	m.AlarmLowMinutes = v
}

// GetAlarmLowPredictionEnabled returns the AlarmLowPredictionEnabled value.
func (m UpdateSiteIdentitySettingsParams) GetAlarmLowPredictionEnabled() *bool {
	return m.AlarmLowPredictionEnabled
}

// SetAlarmLowPredictionEnabled sets the AlarmLowPredictionEnabled value and returns the updated UpdateSiteIdentitySettingsParams.
func (m *UpdateSiteIdentitySettingsParams) SetAlarmLowPredictionEnabled(v *bool) {
	m.AlarmLowPredictionEnabled = v
}

// GetAlarmLowPredictionMinutes returns the AlarmLowPredictionMinutes value.
func (m UpdateSiteIdentitySettingsParams) GetAlarmLowPredictionMinutes() *int64 {
	return m.AlarmLowPredictionMinutes
}

// SetAlarmLowPredictionMinutes sets the AlarmLowPredictionMinutes value and returns the updated UpdateSiteIdentitySettingsParams.
func (m *UpdateSiteIdentitySettingsParams) SetAlarmLowPredictionMinutes(v *int64) {
	m.AlarmLowPredictionMinutes = v
}

// GetAlarmHighEnabled returns the AlarmHighEnabled value.
func (m UpdateSiteIdentitySettingsParams) GetAlarmHighEnabled() *bool {
	return m.AlarmHighEnabled
}

// SetAlarmHighEnabled sets the AlarmHighEnabled value and returns the updated UpdateSiteIdentitySettingsParams.
func (m *UpdateSiteIdentitySettingsParams) SetAlarmHighEnabled(v *bool) {
	m.AlarmHighEnabled = v
}

// GetRangeHighMgdl returns the RangeHighMgdl value.
func (m UpdateSiteIdentitySettingsParams) GetRangeHighMgdl() *int64 {
	return m.RangeHighMgdl
}

// SetRangeHighMgdl sets the RangeHighMgdl value and returns the updated UpdateSiteIdentitySettingsParams.
func (m *UpdateSiteIdentitySettingsParams) SetRangeHighMgdl(v *int64) {
	m.RangeHighMgdl = v
}

// GetAlarmHighMinutes returns the AlarmHighMinutes value.
func (m UpdateSiteIdentitySettingsParams) GetAlarmHighMinutes() *int64 {
	return m.AlarmHighMinutes
}

// SetAlarmHighMinutes sets the AlarmHighMinutes value and returns the updated UpdateSiteIdentitySettingsParams.
func (m *UpdateSiteIdentitySettingsParams) SetAlarmHighMinutes(v *int64) {
	m.AlarmHighMinutes = v
}

// GetAlarmUrgentHighEnabled returns the AlarmUrgentHighEnabled value.
func (m UpdateSiteIdentitySettingsParams) GetAlarmUrgentHighEnabled() *bool {
	return m.AlarmUrgentHighEnabled
}

// SetAlarmUrgentHighEnabled sets the AlarmUrgentHighEnabled value and returns the updated UpdateSiteIdentitySettingsParams.
func (m *UpdateSiteIdentitySettingsParams) SetAlarmUrgentHighEnabled(v *bool) {
	m.AlarmUrgentHighEnabled = v
}

// GetRangeUrgentHighMgdl returns the RangeUrgentHighMgdl value.
func (m UpdateSiteIdentitySettingsParams) GetRangeUrgentHighMgdl() *int64 {
	return m.RangeUrgentHighMgdl
}

// SetRangeUrgentHighMgdl sets the RangeUrgentHighMgdl value and returns the updated UpdateSiteIdentitySettingsParams.
func (m *UpdateSiteIdentitySettingsParams) SetRangeUrgentHighMgdl(v *int64) {
	m.RangeUrgentHighMgdl = v
}

// GetAlarmFastRiseEnabled returns the AlarmFastRiseEnabled value.
func (m UpdateSiteIdentitySettingsParams) GetAlarmFastRiseEnabled() *bool {
	return m.AlarmFastRiseEnabled
}

// SetAlarmFastRiseEnabled sets the AlarmFastRiseEnabled value and returns the updated UpdateSiteIdentitySettingsParams.
func (m *UpdateSiteIdentitySettingsParams) SetAlarmFastRiseEnabled(v *bool) {
	m.AlarmFastRiseEnabled = v
}

// GetAlarmFastRiseMgdl returns the AlarmFastRiseMgdl value.
func (m UpdateSiteIdentitySettingsParams) GetAlarmFastRiseMgdl() *int64 {
	return m.AlarmFastRiseMgdl
}

// SetAlarmFastRiseMgdl sets the AlarmFastRiseMgdl value and returns the updated UpdateSiteIdentitySettingsParams.
func (m *UpdateSiteIdentitySettingsParams) SetAlarmFastRiseMgdl(v *int64) {
	m.AlarmFastRiseMgdl = v
}

// GetAlarmFastRiseMinutes returns the AlarmFastRiseMinutes value.
func (m UpdateSiteIdentitySettingsParams) GetAlarmFastRiseMinutes() *int64 {
	return m.AlarmFastRiseMinutes
}

// SetAlarmFastRiseMinutes sets the AlarmFastRiseMinutes value and returns the updated UpdateSiteIdentitySettingsParams.
func (m *UpdateSiteIdentitySettingsParams) SetAlarmFastRiseMinutes(v *int64) {
	m.AlarmFastRiseMinutes = v
}

// GetAlarmFastDropEnabled returns the AlarmFastDropEnabled value.
func (m UpdateSiteIdentitySettingsParams) GetAlarmFastDropEnabled() *bool {
	return m.AlarmFastDropEnabled
}

// SetAlarmFastDropEnabled sets the AlarmFastDropEnabled value and returns the updated UpdateSiteIdentitySettingsParams.
func (m *UpdateSiteIdentitySettingsParams) SetAlarmFastDropEnabled(v *bool) {
	m.AlarmFastDropEnabled = v
}

// GetAlarmFastDropMgdl returns the AlarmFastDropMgdl value.
func (m UpdateSiteIdentitySettingsParams) GetAlarmFastDropMgdl() *int64 {
	return m.AlarmFastDropMgdl
}

// SetAlarmFastDropMgdl sets the AlarmFastDropMgdl value and returns the updated UpdateSiteIdentitySettingsParams.
func (m *UpdateSiteIdentitySettingsParams) SetAlarmFastDropMgdl(v *int64) {
	m.AlarmFastDropMgdl = v
}

// GetAlarmFastDropMinutes returns the AlarmFastDropMinutes value.
func (m UpdateSiteIdentitySettingsParams) GetAlarmFastDropMinutes() *int64 {
	return m.AlarmFastDropMinutes
}

// SetAlarmFastDropMinutes sets the AlarmFastDropMinutes value and returns the updated UpdateSiteIdentitySettingsParams.
func (m *UpdateSiteIdentitySettingsParams) SetAlarmFastDropMinutes(v *int64) {
	m.AlarmFastDropMinutes = v
}

// GetAlarmNoDataEnabled returns the AlarmNoDataEnabled value.
func (m UpdateSiteIdentitySettingsParams) GetAlarmNoDataEnabled() *bool {
	return m.AlarmNoDataEnabled
}

// SetAlarmNoDataEnabled sets the AlarmNoDataEnabled value and returns the updated UpdateSiteIdentitySettingsParams.
func (m *UpdateSiteIdentitySettingsParams) SetAlarmNoDataEnabled(v *bool) {
	m.AlarmNoDataEnabled = v
}

// GetAlarmNoDataMinutes returns the AlarmNoDataMinutes value.
func (m UpdateSiteIdentitySettingsParams) GetAlarmNoDataMinutes() *int64 {
	return m.AlarmNoDataMinutes
}

// SetAlarmNoDataMinutes sets the AlarmNoDataMinutes value and returns the updated UpdateSiteIdentitySettingsParams.
func (m *UpdateSiteIdentitySettingsParams) SetAlarmNoDataMinutes(v *int64) {
	m.AlarmNoDataMinutes = v
}

// GetAlarmDeviceBatteryLowEnabled returns the AlarmDeviceBatteryLowEnabled value.
func (m UpdateSiteIdentitySettingsParams) GetAlarmDeviceBatteryLowEnabled() *bool {
	return m.AlarmDeviceBatteryLowEnabled
}

// SetAlarmDeviceBatteryLowEnabled sets the AlarmDeviceBatteryLowEnabled value and returns the updated UpdateSiteIdentitySettingsParams.
func (m *UpdateSiteIdentitySettingsParams) SetAlarmDeviceBatteryLowEnabled(v *bool) {
	m.AlarmDeviceBatteryLowEnabled = v
}

// GetAlarmDeviceBatteryLowPercent returns the AlarmDeviceBatteryLowPercent value.
func (m UpdateSiteIdentitySettingsParams) GetAlarmDeviceBatteryLowPercent() *int64 {
	return m.AlarmDeviceBatteryLowPercent
}

// SetAlarmDeviceBatteryLowPercent sets the AlarmDeviceBatteryLowPercent value and returns the updated UpdateSiteIdentitySettingsParams.
func (m *UpdateSiteIdentitySettingsParams) SetAlarmDeviceBatteryLowPercent(v *int64) {
	m.AlarmDeviceBatteryLowPercent = v
}

// GetAlarmDeviceBatteryLowMinutes returns the AlarmDeviceBatteryLowMinutes value.
func (m UpdateSiteIdentitySettingsParams) GetAlarmDeviceBatteryLowMinutes() *int64 {
	return m.AlarmDeviceBatteryLowMinutes
}

// SetAlarmDeviceBatteryLowMinutes sets the AlarmDeviceBatteryLowMinutes value and returns the updated UpdateSiteIdentitySettingsParams.
func (m *UpdateSiteIdentitySettingsParams) SetAlarmDeviceBatteryLowMinutes(v *int64) {
	m.AlarmDeviceBatteryLowMinutes = v
}

// GetAlarmDeviceBatteryLowDevicesJson returns the AlarmDeviceBatteryLowDevicesJson value.
func (m UpdateSiteIdentitySettingsParams) GetAlarmDeviceBatteryLowDevicesJson() *types.DeviceNames {
	return m.AlarmDeviceBatteryLowDevicesJson
}

// SetAlarmDeviceBatteryLowDevicesJson sets the AlarmDeviceBatteryLowDevicesJson value and returns the updated UpdateSiteIdentitySettingsParams.
func (m *UpdateSiteIdentitySettingsParams) SetAlarmDeviceBatteryLowDevicesJson(v *types.DeviceNames) {
	m.AlarmDeviceBatteryLowDevicesJson = v
}

// GetAlarmLowReservoirEnabled returns the AlarmLowReservoirEnabled value.
func (m UpdateSiteIdentitySettingsParams) GetAlarmLowReservoirEnabled() *bool {
	return m.AlarmLowReservoirEnabled
}

// SetAlarmLowReservoirEnabled sets the AlarmLowReservoirEnabled value and returns the updated UpdateSiteIdentitySettingsParams.
func (m *UpdateSiteIdentitySettingsParams) SetAlarmLowReservoirEnabled(v *bool) {
	m.AlarmLowReservoirEnabled = v
}

// GetAlarmLowReservoirUnits returns the AlarmLowReservoirUnits value.
func (m UpdateSiteIdentitySettingsParams) GetAlarmLowReservoirUnits() *int64 {
	return m.AlarmLowReservoirUnits
}

// SetAlarmLowReservoirUnits sets the AlarmLowReservoirUnits value and returns the updated UpdateSiteIdentitySettingsParams.
func (m *UpdateSiteIdentitySettingsParams) SetAlarmLowReservoirUnits(v *int64) {
	m.AlarmLowReservoirUnits = v
}

// GetNotificationsEnabled returns the NotificationsEnabled value.
func (m UpdateSiteIdentitySettingsParams) GetNotificationsEnabled() *bool {
	return m.NotificationsEnabled
}

// SetNotificationsEnabled sets the NotificationsEnabled value and returns the updated UpdateSiteIdentitySettingsParams.
func (m *UpdateSiteIdentitySettingsParams) SetNotificationsEnabled(v *bool) {
	m.NotificationsEnabled = v
}

// GetID returns the ID value.
func (m UpdateSiteIdentitySettingsParams) GetID() int64 {
	return m.ID
}

// SetID sets the ID value and returns the updated UpdateSiteIdentitySettingsParams.
func (m *UpdateSiteIdentitySettingsParams) SetID(v int64) {
	m.ID = v
}

// GetUnits returns the Units value.
func (m UpdateSitesSettingsParams) GetUnits() *consts.Units {
	return m.Units
}

// SetUnits sets the Units value and returns the updated UpdateSitesSettingsParams.
func (m *UpdateSitesSettingsParams) SetUnits(v *consts.Units) {
	m.Units = v
}

// GetUnitsMgdlLabel returns the UnitsMgdlLabel value.
func (m UpdateSitesSettingsParams) GetUnitsMgdlLabel() *string {
	return m.UnitsMgdlLabel
}

// SetUnitsMgdlLabel sets the UnitsMgdlLabel value and returns the updated UpdateSitesSettingsParams.
func (m *UpdateSitesSettingsParams) SetUnitsMgdlLabel(v *string) {
	m.UnitsMgdlLabel = v
}

// GetUnitsMmolLabel returns the UnitsMmolLabel value.
func (m UpdateSitesSettingsParams) GetUnitsMmolLabel() *string {
	return m.UnitsMmolLabel
}

// SetUnitsMmolLabel sets the UnitsMmolLabel value and returns the updated UpdateSitesSettingsParams.
func (m *UpdateSitesSettingsParams) SetUnitsMmolLabel(v *string) {
	m.UnitsMmolLabel = v
}

// GetTimezone returns the Timezone value.
func (m UpdateSitesSettingsParams) GetTimezone() *string {
	return m.Timezone
}

// SetTimezone sets the Timezone value and returns the updated UpdateSitesSettingsParams.
func (m *UpdateSitesSettingsParams) SetTimezone(v *string) {
	m.Timezone = v
}

// GetLanguage returns the Language value.
func (m UpdateSitesSettingsParams) GetLanguage() *string {
	return m.Language
}

// SetLanguage sets the Language value and returns the updated UpdateSitesSettingsParams.
func (m *UpdateSitesSettingsParams) SetLanguage(v *string) {
	m.Language = v
}

// GetDateFormat returns the DateFormat value.
func (m UpdateSitesSettingsParams) GetDateFormat() *string {
	return m.DateFormat
}

// SetDateFormat sets the DateFormat value and returns the updated UpdateSitesSettingsParams.
func (m *UpdateSitesSettingsParams) SetDateFormat(v *string) {
	m.DateFormat = v
}

// GetTimeFormat returns the TimeFormat value.
func (m UpdateSitesSettingsParams) GetTimeFormat() *string {
	return m.TimeFormat
}

// SetTimeFormat sets the TimeFormat value and returns the updated UpdateSitesSettingsParams.
func (m *UpdateSitesSettingsParams) SetTimeFormat(v *string) {
	m.TimeFormat = v
}

// GetWeekStartDay returns the WeekStartDay value.
func (m UpdateSitesSettingsParams) GetWeekStartDay() *string {
	return m.WeekStartDay
}

// SetWeekStartDay sets the WeekStartDay value and returns the updated UpdateSitesSettingsParams.
func (m *UpdateSitesSettingsParams) SetWeekStartDay(v *string) {
	m.WeekStartDay = v
}

// GetNumberFormat returns the NumberFormat value.
func (m UpdateSitesSettingsParams) GetNumberFormat() *string {
	return m.NumberFormat
}

// SetNumberFormat sets the NumberFormat value and returns the updated UpdateSitesSettingsParams.
func (m *UpdateSitesSettingsParams) SetNumberFormat(v *string) {
	m.NumberFormat = v
}

// GetBolusDiaHours returns the BolusDiaHours value.
func (m UpdateSitesSettingsParams) GetBolusDiaHours() *float64 {
	return m.BolusDiaHours
}

// SetBolusDiaHours sets the BolusDiaHours value and returns the updated UpdateSitesSettingsParams.
func (m *UpdateSitesSettingsParams) SetBolusDiaHours(v *float64) {
	m.BolusDiaHours = v
}

// GetBolusOnsetMinutes returns the BolusOnsetMinutes value.
func (m UpdateSitesSettingsParams) GetBolusOnsetMinutes() *float64 {
	return m.BolusOnsetMinutes
}

// SetBolusOnsetMinutes sets the BolusOnsetMinutes value and returns the updated UpdateSitesSettingsParams.
func (m *UpdateSitesSettingsParams) SetBolusOnsetMinutes(v *float64) {
	m.BolusOnsetMinutes = v
}

// GetBolusPeakMinutes returns the BolusPeakMinutes value.
func (m UpdateSitesSettingsParams) GetBolusPeakMinutes() *float64 {
	return m.BolusPeakMinutes
}

// SetBolusPeakMinutes sets the BolusPeakMinutes value and returns the updated UpdateSitesSettingsParams.
func (m *UpdateSitesSettingsParams) SetBolusPeakMinutes(v *float64) {
	m.BolusPeakMinutes = v
}

// GetBasalDiaHours returns the BasalDiaHours value.
func (m UpdateSitesSettingsParams) GetBasalDiaHours() *float64 {
	return m.BasalDiaHours
}

// SetBasalDiaHours sets the BasalDiaHours value and returns the updated UpdateSitesSettingsParams.
func (m *UpdateSitesSettingsParams) SetBasalDiaHours(v *float64) {
	m.BasalDiaHours = v
}

// GetBasalOnsetMinutes returns the BasalOnsetMinutes value.
func (m UpdateSitesSettingsParams) GetBasalOnsetMinutes() *float64 {
	return m.BasalOnsetMinutes
}

// SetBasalOnsetMinutes sets the BasalOnsetMinutes value and returns the updated UpdateSitesSettingsParams.
func (m *UpdateSitesSettingsParams) SetBasalOnsetMinutes(v *float64) {
	m.BasalOnsetMinutes = v
}

// GetBasalPeakMinutes returns the BasalPeakMinutes value.
func (m UpdateSitesSettingsParams) GetBasalPeakMinutes() *float64 {
	return m.BasalPeakMinutes
}

// SetBasalPeakMinutes sets the BasalPeakMinutes value and returns the updated UpdateSitesSettingsParams.
func (m *UpdateSitesSettingsParams) SetBasalPeakMinutes(v *float64) {
	m.BasalPeakMinutes = v
}

// GetDefaultInsulinSensitivity returns the DefaultInsulinSensitivity value.
func (m UpdateSitesSettingsParams) GetDefaultInsulinSensitivity() *float64 {
	return m.DefaultInsulinSensitivity
}

// SetDefaultInsulinSensitivity sets the DefaultInsulinSensitivity value and returns the updated UpdateSitesSettingsParams.
func (m *UpdateSitesSettingsParams) SetDefaultInsulinSensitivity(v *float64) {
	m.DefaultInsulinSensitivity = v
}

// GetDefaultCarbRatio returns the DefaultCarbRatio value.
func (m UpdateSitesSettingsParams) GetDefaultCarbRatio() *float64 {
	return m.DefaultCarbRatio
}

// SetDefaultCarbRatio sets the DefaultCarbRatio value and returns the updated UpdateSitesSettingsParams.
func (m *UpdateSitesSettingsParams) SetDefaultCarbRatio(v *float64) {
	m.DefaultCarbRatio = v
}

// GetDefaultBasalRate returns the DefaultBasalRate value.
func (m UpdateSitesSettingsParams) GetDefaultBasalRate() *float64 {
	return m.DefaultBasalRate
}

// SetDefaultBasalRate sets the DefaultBasalRate value and returns the updated UpdateSitesSettingsParams.
func (m *UpdateSitesSettingsParams) SetDefaultBasalRate(v *float64) {
	m.DefaultBasalRate = v
}

// GetCarbMaxAbsorptionHours returns the CarbMaxAbsorptionHours value.
func (m UpdateSitesSettingsParams) GetCarbMaxAbsorptionHours() *float64 {
	return m.CarbMaxAbsorptionHours
}

// SetCarbMaxAbsorptionHours sets the CarbMaxAbsorptionHours value and returns the updated UpdateSitesSettingsParams.
func (m *UpdateSitesSettingsParams) SetCarbMaxAbsorptionHours(v *float64) {
	m.CarbMaxAbsorptionHours = v
}

// GetCarbMin5mImpact returns the CarbMin5mImpact value.
func (m UpdateSitesSettingsParams) GetCarbMin5mImpact() *float64 {
	return m.CarbMin5mImpact
}

// SetCarbMin5mImpact sets the CarbMin5mImpact value and returns the updated UpdateSitesSettingsParams.
func (m *UpdateSitesSettingsParams) SetCarbMin5mImpact(v *float64) {
	m.CarbMin5mImpact = v
}

// GetAlarmUrgentLowEnabled returns the AlarmUrgentLowEnabled value.
func (m UpdateSitesSettingsParams) GetAlarmUrgentLowEnabled() *bool {
	return m.AlarmUrgentLowEnabled
}

// SetAlarmUrgentLowEnabled sets the AlarmUrgentLowEnabled value and returns the updated UpdateSitesSettingsParams.
func (m *UpdateSitesSettingsParams) SetAlarmUrgentLowEnabled(v *bool) {
	m.AlarmUrgentLowEnabled = v
}

// GetRangeUrgentLowMgdl returns the RangeUrgentLowMgdl value.
func (m UpdateSitesSettingsParams) GetRangeUrgentLowMgdl() *int64 {
	return m.RangeUrgentLowMgdl
}

// SetRangeUrgentLowMgdl sets the RangeUrgentLowMgdl value and returns the updated UpdateSitesSettingsParams.
func (m *UpdateSitesSettingsParams) SetRangeUrgentLowMgdl(v *int64) {
	m.RangeUrgentLowMgdl = v
}

// GetAlarmLowEnabled returns the AlarmLowEnabled value.
func (m UpdateSitesSettingsParams) GetAlarmLowEnabled() *bool {
	return m.AlarmLowEnabled
}

// SetAlarmLowEnabled sets the AlarmLowEnabled value and returns the updated UpdateSitesSettingsParams.
func (m *UpdateSitesSettingsParams) SetAlarmLowEnabled(v *bool) {
	m.AlarmLowEnabled = v
}

// GetRangeLowMgdl returns the RangeLowMgdl value.
func (m UpdateSitesSettingsParams) GetRangeLowMgdl() *int64 {
	return m.RangeLowMgdl
}

// SetRangeLowMgdl sets the RangeLowMgdl value and returns the updated UpdateSitesSettingsParams.
func (m *UpdateSitesSettingsParams) SetRangeLowMgdl(v *int64) {
	m.RangeLowMgdl = v
}

// GetAlarmLowMinutes returns the AlarmLowMinutes value.
func (m UpdateSitesSettingsParams) GetAlarmLowMinutes() *int64 {
	return m.AlarmLowMinutes
}

// SetAlarmLowMinutes sets the AlarmLowMinutes value and returns the updated UpdateSitesSettingsParams.
func (m *UpdateSitesSettingsParams) SetAlarmLowMinutes(v *int64) {
	m.AlarmLowMinutes = v
}

// GetAlarmLowPredictionEnabled returns the AlarmLowPredictionEnabled value.
func (m UpdateSitesSettingsParams) GetAlarmLowPredictionEnabled() *bool {
	return m.AlarmLowPredictionEnabled
}

// SetAlarmLowPredictionEnabled sets the AlarmLowPredictionEnabled value and returns the updated UpdateSitesSettingsParams.
func (m *UpdateSitesSettingsParams) SetAlarmLowPredictionEnabled(v *bool) {
	m.AlarmLowPredictionEnabled = v
}

// GetAlarmLowPredictionMinutes returns the AlarmLowPredictionMinutes value.
func (m UpdateSitesSettingsParams) GetAlarmLowPredictionMinutes() *int64 {
	return m.AlarmLowPredictionMinutes
}

// SetAlarmLowPredictionMinutes sets the AlarmLowPredictionMinutes value and returns the updated UpdateSitesSettingsParams.
func (m *UpdateSitesSettingsParams) SetAlarmLowPredictionMinutes(v *int64) {
	m.AlarmLowPredictionMinutes = v
}

// GetAlarmHighEnabled returns the AlarmHighEnabled value.
func (m UpdateSitesSettingsParams) GetAlarmHighEnabled() *bool {
	return m.AlarmHighEnabled
}

// SetAlarmHighEnabled sets the AlarmHighEnabled value and returns the updated UpdateSitesSettingsParams.
func (m *UpdateSitesSettingsParams) SetAlarmHighEnabled(v *bool) {
	m.AlarmHighEnabled = v
}

// GetRangeHighMgdl returns the RangeHighMgdl value.
func (m UpdateSitesSettingsParams) GetRangeHighMgdl() *int64 {
	return m.RangeHighMgdl
}

// SetRangeHighMgdl sets the RangeHighMgdl value and returns the updated UpdateSitesSettingsParams.
func (m *UpdateSitesSettingsParams) SetRangeHighMgdl(v *int64) {
	m.RangeHighMgdl = v
}

// GetAlarmHighMinutes returns the AlarmHighMinutes value.
func (m UpdateSitesSettingsParams) GetAlarmHighMinutes() *int64 {
	return m.AlarmHighMinutes
}

// SetAlarmHighMinutes sets the AlarmHighMinutes value and returns the updated UpdateSitesSettingsParams.
func (m *UpdateSitesSettingsParams) SetAlarmHighMinutes(v *int64) {
	m.AlarmHighMinutes = v
}

// GetAlarmUrgentHighEnabled returns the AlarmUrgentHighEnabled value.
func (m UpdateSitesSettingsParams) GetAlarmUrgentHighEnabled() *bool {
	return m.AlarmUrgentHighEnabled
}

// SetAlarmUrgentHighEnabled sets the AlarmUrgentHighEnabled value and returns the updated UpdateSitesSettingsParams.
func (m *UpdateSitesSettingsParams) SetAlarmUrgentHighEnabled(v *bool) {
	m.AlarmUrgentHighEnabled = v
}

// GetRangeUrgentHighMgdl returns the RangeUrgentHighMgdl value.
func (m UpdateSitesSettingsParams) GetRangeUrgentHighMgdl() *int64 {
	return m.RangeUrgentHighMgdl
}

// SetRangeUrgentHighMgdl sets the RangeUrgentHighMgdl value and returns the updated UpdateSitesSettingsParams.
func (m *UpdateSitesSettingsParams) SetRangeUrgentHighMgdl(v *int64) {
	m.RangeUrgentHighMgdl = v
}

// GetAlarmFastRiseEnabled returns the AlarmFastRiseEnabled value.
func (m UpdateSitesSettingsParams) GetAlarmFastRiseEnabled() *bool {
	return m.AlarmFastRiseEnabled
}

// SetAlarmFastRiseEnabled sets the AlarmFastRiseEnabled value and returns the updated UpdateSitesSettingsParams.
func (m *UpdateSitesSettingsParams) SetAlarmFastRiseEnabled(v *bool) {
	m.AlarmFastRiseEnabled = v
}

// GetAlarmFastRiseMgdl returns the AlarmFastRiseMgdl value.
func (m UpdateSitesSettingsParams) GetAlarmFastRiseMgdl() *int64 {
	return m.AlarmFastRiseMgdl
}

// SetAlarmFastRiseMgdl sets the AlarmFastRiseMgdl value and returns the updated UpdateSitesSettingsParams.
func (m *UpdateSitesSettingsParams) SetAlarmFastRiseMgdl(v *int64) {
	m.AlarmFastRiseMgdl = v
}

// GetAlarmFastRiseMinutes returns the AlarmFastRiseMinutes value.
func (m UpdateSitesSettingsParams) GetAlarmFastRiseMinutes() *int64 {
	return m.AlarmFastRiseMinutes
}

// SetAlarmFastRiseMinutes sets the AlarmFastRiseMinutes value and returns the updated UpdateSitesSettingsParams.
func (m *UpdateSitesSettingsParams) SetAlarmFastRiseMinutes(v *int64) {
	m.AlarmFastRiseMinutes = v
}

// GetAlarmFastDropEnabled returns the AlarmFastDropEnabled value.
func (m UpdateSitesSettingsParams) GetAlarmFastDropEnabled() *bool {
	return m.AlarmFastDropEnabled
}

// SetAlarmFastDropEnabled sets the AlarmFastDropEnabled value and returns the updated UpdateSitesSettingsParams.
func (m *UpdateSitesSettingsParams) SetAlarmFastDropEnabled(v *bool) {
	m.AlarmFastDropEnabled = v
}

// GetAlarmFastDropMgdl returns the AlarmFastDropMgdl value.
func (m UpdateSitesSettingsParams) GetAlarmFastDropMgdl() *int64 {
	return m.AlarmFastDropMgdl
}

// SetAlarmFastDropMgdl sets the AlarmFastDropMgdl value and returns the updated UpdateSitesSettingsParams.
func (m *UpdateSitesSettingsParams) SetAlarmFastDropMgdl(v *int64) {
	m.AlarmFastDropMgdl = v
}

// GetAlarmFastDropMinutes returns the AlarmFastDropMinutes value.
func (m UpdateSitesSettingsParams) GetAlarmFastDropMinutes() *int64 {
	return m.AlarmFastDropMinutes
}

// SetAlarmFastDropMinutes sets the AlarmFastDropMinutes value and returns the updated UpdateSitesSettingsParams.
func (m *UpdateSitesSettingsParams) SetAlarmFastDropMinutes(v *int64) {
	m.AlarmFastDropMinutes = v
}

// GetAlarmNoDataEnabled returns the AlarmNoDataEnabled value.
func (m UpdateSitesSettingsParams) GetAlarmNoDataEnabled() *bool {
	return m.AlarmNoDataEnabled
}

// SetAlarmNoDataEnabled sets the AlarmNoDataEnabled value and returns the updated UpdateSitesSettingsParams.
func (m *UpdateSitesSettingsParams) SetAlarmNoDataEnabled(v *bool) {
	m.AlarmNoDataEnabled = v
}

// GetAlarmNoDataMinutes returns the AlarmNoDataMinutes value.
func (m UpdateSitesSettingsParams) GetAlarmNoDataMinutes() *int64 {
	return m.AlarmNoDataMinutes
}

// SetAlarmNoDataMinutes sets the AlarmNoDataMinutes value and returns the updated UpdateSitesSettingsParams.
func (m *UpdateSitesSettingsParams) SetAlarmNoDataMinutes(v *int64) {
	m.AlarmNoDataMinutes = v
}

// GetAlarmDeviceBatteryLowEnabled returns the AlarmDeviceBatteryLowEnabled value.
func (m UpdateSitesSettingsParams) GetAlarmDeviceBatteryLowEnabled() *bool {
	return m.AlarmDeviceBatteryLowEnabled
}

// SetAlarmDeviceBatteryLowEnabled sets the AlarmDeviceBatteryLowEnabled value and returns the updated UpdateSitesSettingsParams.
func (m *UpdateSitesSettingsParams) SetAlarmDeviceBatteryLowEnabled(v *bool) {
	m.AlarmDeviceBatteryLowEnabled = v
}

// GetAlarmDeviceBatteryLowPercent returns the AlarmDeviceBatteryLowPercent value.
func (m UpdateSitesSettingsParams) GetAlarmDeviceBatteryLowPercent() *int64 {
	return m.AlarmDeviceBatteryLowPercent
}

// SetAlarmDeviceBatteryLowPercent sets the AlarmDeviceBatteryLowPercent value and returns the updated UpdateSitesSettingsParams.
func (m *UpdateSitesSettingsParams) SetAlarmDeviceBatteryLowPercent(v *int64) {
	m.AlarmDeviceBatteryLowPercent = v
}

// GetAlarmDeviceBatteryLowMinutes returns the AlarmDeviceBatteryLowMinutes value.
func (m UpdateSitesSettingsParams) GetAlarmDeviceBatteryLowMinutes() *int64 {
	return m.AlarmDeviceBatteryLowMinutes
}

// SetAlarmDeviceBatteryLowMinutes sets the AlarmDeviceBatteryLowMinutes value and returns the updated UpdateSitesSettingsParams.
func (m *UpdateSitesSettingsParams) SetAlarmDeviceBatteryLowMinutes(v *int64) {
	m.AlarmDeviceBatteryLowMinutes = v
}

// GetAlarmDeviceBatteryLowDevicesJson returns the AlarmDeviceBatteryLowDevicesJson value.
func (m UpdateSitesSettingsParams) GetAlarmDeviceBatteryLowDevicesJson() *types.DeviceNames {
	return m.AlarmDeviceBatteryLowDevicesJson
}

// SetAlarmDeviceBatteryLowDevicesJson sets the AlarmDeviceBatteryLowDevicesJson value and returns the updated UpdateSitesSettingsParams.
func (m *UpdateSitesSettingsParams) SetAlarmDeviceBatteryLowDevicesJson(v *types.DeviceNames) {
	m.AlarmDeviceBatteryLowDevicesJson = v
}

// GetAlarmLowReservoirEnabled returns the AlarmLowReservoirEnabled value.
func (m UpdateSitesSettingsParams) GetAlarmLowReservoirEnabled() *bool {
	return m.AlarmLowReservoirEnabled
}

// SetAlarmLowReservoirEnabled sets the AlarmLowReservoirEnabled value and returns the updated UpdateSitesSettingsParams.
func (m *UpdateSitesSettingsParams) SetAlarmLowReservoirEnabled(v *bool) {
	m.AlarmLowReservoirEnabled = v
}

// GetAlarmLowReservoirUnits returns the AlarmLowReservoirUnits value.
func (m UpdateSitesSettingsParams) GetAlarmLowReservoirUnits() *int64 {
	return m.AlarmLowReservoirUnits
}

// SetAlarmLowReservoirUnits sets the AlarmLowReservoirUnits value and returns the updated UpdateSitesSettingsParams.
func (m *UpdateSitesSettingsParams) SetAlarmLowReservoirUnits(v *int64) {
	m.AlarmLowReservoirUnits = v
}

// GetNotificationsEnabled returns the NotificationsEnabled value.
func (m UpdateSitesSettingsParams) GetNotificationsEnabled() *bool {
	return m.NotificationsEnabled
}

// SetNotificationsEnabled sets the NotificationsEnabled value and returns the updated UpdateSitesSettingsParams.
func (m *UpdateSitesSettingsParams) SetNotificationsEnabled(v *bool) {
	m.NotificationsEnabled = v
}

// GetID returns the ID value.
func (m UpdateSitesSettingsParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated UpdateSitesSettingsParams.
func (m *UpdateSitesSettingsParams) SetID(v string) {
	m.ID = v
}

// GetType returns the Type value.
func (m CreateConnectorParams) GetType() consts.ConnectorType {
	return m.Type
}

// GetType returns the Type value.
func (m DeleteConnectorParams) GetType() consts.ConnectorType {
	return m.Type
}

// GetError returns the Error value.
func (m FinishConnectorBackfillParams) GetError() *string {
	return m.Error
}

// GetType returns the Type value.
func (m FinishConnectorBackfillParams) GetType() consts.ConnectorType {
	return m.Type
}

// GetType returns the Type value.
func (m GetConnectorParams) GetType() consts.ConnectorType {
	return m.Type
}

// SetLimit sets the Limit value and returns the updated SearchConnectorsParams.
func (m *SearchConnectorsParams) SetLimit(v int64) {
	m.Limit = v
}

// GetType returns the Type value.
func (m StartConnectorBackfillParams) GetType() consts.ConnectorType {
	return m.Type
}

// GetType returns the Type value.
func (m UpdateConnectorParams) GetType() consts.ConnectorType {
	return m.Type
}

// GetType returns the Type value.
func (m UpdateConnectorAttemptParams) GetType() consts.ConnectorType {
	return m.Type
}

// GetType returns the Type value.
func (m UpdateConnectorReportsFailureParams) GetType() consts.ConnectorType {
	return m.Type
}

// GetType returns the Type value.
func (m UpdateConnectorReportsSuccessParams) GetType() consts.ConnectorType {
	return m.Type
}

// GetName returns the Name value.
func (m CreateSiteIdentityParams) GetName() string {
	return m.Name
}

// GetCreatedByIdentityName returns the CreatedByIdentityName value.
func (m CreateSiteIdentityParams) GetCreatedByIdentityName() string {
	return m.CreatedByIdentityName
}

// SetCreatedByIdentityName sets the CreatedByIdentityName value and returns the updated CreateSiteIdentityParams.
func (m *CreateSiteIdentityParams) SetCreatedByIdentityName(v string) {
	m.CreatedByIdentityName = v
}

// GetID returns the ID value.
func (m DeleteSiteIdentityParams) GetID() int64 {
	return m.ID
}

// SetID sets the ID value and returns the updated DeleteSiteIdentityParams.
func (m *DeleteSiteIdentityParams) SetID(v int64) {
	m.ID = v
}

// GetName returns the Name value.
func (m DisableOtherSiteAdminsParams) GetName() string {
	return m.Name
}

// GetName returns the Name value.
func (m GetSiteIdentityParams) GetName() string {
	return m.Name
}

// GetID returns the ID value.
func (m GetSiteIdentityByIDParams) GetID() int64 {
	return m.ID
}

// SetID sets the ID value and returns the updated GetSiteIdentityByIDParams.
func (m *GetSiteIdentityByIDParams) SetID(v int64) {
	m.ID = v
}

// GetName returns the Name value.
func (m ResetSiteAdminCredentialsParams) GetName() string {
	return m.Name
}

// GetID returns the ID value.
func (m UpdateSiteIdentityDateFlagsParams) GetID() int64 {
	return m.ID
}

// SetID sets the ID value and returns the updated UpdateSiteIdentityDateFlagsParams.
func (m *UpdateSiteIdentityDateFlagsParams) SetID(v int64) {
	m.ID = v
}

// GetID returns the ID value.
func (m UpdateSiteIdentityEnabledParams) GetID() int64 {
	return m.ID
}

// SetID sets the ID value and returns the updated UpdateSiteIdentityEnabledParams.
func (m *UpdateSiteIdentityEnabledParams) SetID(v int64) {
	m.ID = v
}

// GetID returns the ID value.
func (m UpdateSiteIdentityPasswordParams) GetID() int64 {
	return m.ID
}

// SetID sets the ID value and returns the updated UpdateSiteIdentityPasswordParams.
func (m *UpdateSiteIdentityPasswordParams) SetID(v int64) {
	m.ID = v
}

// GetID returns the ID value.
func (m UpdateSiteIdentityTOTPParams) GetID() int64 {
	return m.ID
}

// SetID sets the ID value and returns the updated UpdateSiteIdentityTOTPParams.
func (m *UpdateSiteIdentityTOTPParams) SetID(v int64) {
	m.ID = v
}

// GetID returns the ID value.
func (m CreateSiteParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated CreateSiteParams.
func (m *CreateSiteParams) SetID(v string) {
	m.ID = v
}

// GetID returns the ID value.
func (m DisableSiteParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated DisableSiteParams.
func (m *DisableSiteParams) SetID(v string) {
	m.ID = v
}

// GetID returns the ID value.
func (m GetSiteParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated GetSiteParams.
func (m *GetSiteParams) SetID(v string) {
	m.ID = v
}

// GetID returns the ID value.
func (m GetSiteInAnyStateParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated GetSiteInAnyStateParams.
func (m *GetSiteInAnyStateParams) SetID(v string) {
	m.ID = v
}

// SetLimit sets the Limit value and returns the updated GetSitesToBackUpParams.
func (m *GetSitesToBackUpParams) SetLimit(v int64) {
	m.Limit = v
}

// GetID returns the ID value.
func (m HydrateSiteParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated HydrateSiteParams.
func (m *HydrateSiteParams) SetID(v string) {
	m.ID = v
}

// GetID returns the ID value.
func (m MarkSiteAvailableParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated MarkSiteAvailableParams.
func (m *MarkSiteAvailableParams) SetID(v string) {
	m.ID = v
}

// GetReason returns the Reason value.
func (m MarkSiteUnavailableParams) GetReason() *string {
	return m.Reason
}

// SetReason sets the Reason value and returns the updated MarkSiteUnavailableParams.
func (m *MarkSiteUnavailableParams) SetReason(v *string) {
	m.Reason = v
}

// GetID returns the ID value.
func (m MarkSiteUnavailableParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated MarkSiteUnavailableParams.
func (m *MarkSiteUnavailableParams) SetID(v string) {
	m.ID = v
}

// GetID returns the ID value.
func (m RotateSiteBackupKeyParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated RotateSiteBackupKeyParams.
func (m *RotateSiteBackupKeyParams) SetID(v string) {
	m.ID = v
}

// SetLimit sets the Limit value and returns the updated SearchArchivedSitesParams.
func (m *SearchArchivedSitesParams) SetLimit(v int64) {
	m.Limit = v
}

// GetType returns the Type value.
func (m SearchSitesParams) GetType() *string {
	return m.Type
}

// SetLimit sets the Limit value and returns the updated SearchSitesParams.
func (m *SearchSitesParams) SetLimit(v int64) {
	m.Limit = v
}

// GetType returns the Type value.
func (m SearchSitesByDBSizeParams) GetType() *string {
	return m.Type
}

// SetLimit sets the Limit value and returns the updated SearchSitesByDBSizeParams.
func (m *SearchSitesByDBSizeParams) SetLimit(v int64) {
	m.Limit = v
}

// GetType returns the Type value.
func (m SearchSitesByLastWriteParams) GetType() *string {
	return m.Type
}

// SetLimit sets the Limit value and returns the updated SearchSitesByLastWriteParams.
func (m *SearchSitesByLastWriteParams) SetLimit(v int64) {
	m.Limit = v
}

// GetID returns the ID value.
func (m SiteExistsParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated SiteExistsParams.
func (m *SiteExistsParams) SetID(v string) {
	m.ID = v
}

// GetID returns the ID value.
func (m UpdateSiteAbuseScoreParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated UpdateSiteAbuseScoreParams.
func (m *UpdateSiteAbuseScoreParams) SetID(v string) {
	m.ID = v
}

// GetID returns the ID value.
func (m UpdateSiteBackedUpParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated UpdateSiteBackedUpParams.
func (m *UpdateSiteBackedUpParams) SetID(v string) {
	m.ID = v
}

// GetID returns the ID value.
func (m UpdateSiteDBSizeParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated UpdateSiteDBSizeParams.
func (m *UpdateSiteDBSizeParams) SetID(v string) {
	m.ID = v
}

// GetID returns the ID value.
func (m UpdateSiteDateFlagsParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated UpdateSiteDateFlagsParams.
func (m *UpdateSiteDateFlagsParams) SetID(v string) {
	m.ID = v
}

// GetID returns the ID value.
func (m UpdateSiteSummariesCheckedAtParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated UpdateSiteSummariesCheckedAtParams.
func (m *UpdateSiteSummariesCheckedAtParams) SetID(v string) {
	m.ID = v
}

// GetID returns the ID value.
func (m UpdateSiteTableSizesParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated UpdateSiteTableSizesParams.
func (m *UpdateSiteTableSizesParams) SetID(v string) {
	m.ID = v
}

