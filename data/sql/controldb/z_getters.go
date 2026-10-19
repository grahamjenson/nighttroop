
package controldb

import (
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/errors"
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

// SetName sets the Name value and returns the updated ControlIdentity.
func (m *ControlIdentity) SetName(v string) {
	m.Name = v
}

// GetPasswordHash returns the PasswordHash value.
func (m ControlIdentity) GetPasswordHash() string {
	return m.PasswordHash
}

// SetPasswordHash sets the PasswordHash value and returns the updated ControlIdentity.
func (m *ControlIdentity) SetPasswordHash(v string) {
	m.PasswordHash = v
}

// GetTotpSecretEnc returns the TotpSecretEnc value.
func (m ControlIdentity) GetTotpSecretEnc() string {
	return m.TotpSecretEnc
}

// SetTotpSecretEnc sets the TotpSecretEnc value and returns the updated ControlIdentity.
func (m *ControlIdentity) SetTotpSecretEnc(v string) {
	m.TotpSecretEnc = v
}

// GetRole returns the Role value.
func (m ControlIdentity) GetRole() consts.ControlRole {
	return m.Role
}

// SetRole sets the Role value and returns the updated ControlIdentity.
func (m *ControlIdentity) SetRole(v consts.ControlRole) {
	m.Role = v
}

// GetCreatedAt returns the CreatedAt value.
func (m ControlIdentity) GetCreatedAt() int64 {
	return m.CreatedAt
}

// SetCreatedAt sets the CreatedAt value and returns the updated ControlIdentity.
func (m *ControlIdentity) SetCreatedAt(v int64) {
	m.CreatedAt = v
}

// GetLastSeenAt returns the LastSeenAt value.
func (m ControlIdentity) GetLastSeenAt() *int64 {
	return m.LastSeenAt
}

// SetLastSeenAt sets the LastSeenAt value and returns the updated ControlIdentity.
func (m *ControlIdentity) SetLastSeenAt(v *int64) {
	m.LastSeenAt = v
}

// GetLastWriteAt returns the LastWriteAt value.
func (m ControlIdentity) GetLastWriteAt() *int64 {
	return m.LastWriteAt
}

// SetLastWriteAt sets the LastWriteAt value and returns the updated ControlIdentity.
func (m *ControlIdentity) SetLastWriteAt(v *int64) {
	m.LastWriteAt = v
}

// GetUpdatedByIdentityID returns the UpdatedByIdentityID value.
func (m ControlIdentity) GetUpdatedByIdentityID() *int64 {
	return m.UpdatedByIdentityID
}

// SetUpdatedByIdentityID sets the UpdatedByIdentityID value and returns the updated ControlIdentity.
func (m *ControlIdentity) SetUpdatedByIdentityID(v *int64) {
	m.UpdatedByIdentityID = v
}

// GetUpdatedByIdentityName returns the UpdatedByIdentityName value.
func (m ControlIdentity) GetUpdatedByIdentityName() *string {
	return m.UpdatedByIdentityName
}

// SetUpdatedByIdentityName sets the UpdatedByIdentityName value and returns the updated ControlIdentity.
func (m *ControlIdentity) SetUpdatedByIdentityName(v *string) {
	m.UpdatedByIdentityName = v
}

// GetCreatedByIdentityID returns the CreatedByIdentityID value.
func (m ControlIdentity) GetCreatedByIdentityID() int64 {
	return m.CreatedByIdentityID
}

// SetCreatedByIdentityID sets the CreatedByIdentityID value and returns the updated ControlIdentity.
func (m *ControlIdentity) SetCreatedByIdentityID(v int64) {
	m.CreatedByIdentityID = v
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

// GetID returns the ID value.
func (m DataSourceBlockList) GetID() int64 {
	return m.ID
}

// SetID sets the ID value and returns the updated DataSourceBlockList.
func (m *DataSourceBlockList) SetID(v int64) {
	m.ID = v
}

// GetCreatedAt returns the CreatedAt value.
func (m DataSourceBlockList) GetCreatedAt() int64 {
	return m.CreatedAt
}

// SetCreatedAt sets the CreatedAt value and returns the updated DataSourceBlockList.
func (m *DataSourceBlockList) SetCreatedAt(v int64) {
	m.CreatedAt = v
}

// GetUsername returns the Username value.
func (m DataSourceBlockList) GetUsername() *string {
	return m.Username
}

// SetUsername sets the Username value and returns the updated DataSourceBlockList.
func (m *DataSourceBlockList) SetUsername(v *string) {
	m.Username = v
}

// GetUrl returns the Url value.
func (m DataSourceBlockList) GetUrl() *string {
	return m.Url
}

// SetUrl sets the Url value and returns the updated DataSourceBlockList.
func (m *DataSourceBlockList) SetUrl(v *string) {
	m.Url = v
}

// GetReason returns the Reason value.
func (m DataSourceBlockList) GetReason() *string {
	return m.Reason
}

// SetReason sets the Reason value and returns the updated DataSourceBlockList.
func (m *DataSourceBlockList) SetReason(v *string) {
	m.Reason = v
}

// GetID returns the ID value.
func (m SiteIdentity) GetID() int64 {
	return m.ID
}

// SetID sets the ID value and returns the updated SiteIdentity.
func (m *SiteIdentity) SetID(v int64) {
	m.ID = v
}

// GetSiteID returns the SiteID value.
func (m SiteIdentity) GetSiteID() string {
	return m.SiteID
}

// SetSiteID sets the SiteID value and returns the updated SiteIdentity.
func (m *SiteIdentity) SetSiteID(v string) {
	m.SiteID = v
}

// GetName returns the Name value.
func (m SiteIdentity) GetName() string {
	return m.Name
}

// SetName sets the Name value and returns the updated SiteIdentity.
func (m *SiteIdentity) SetName(v string) {
	m.Name = v
}

// GetPasswordHash returns the PasswordHash value.
func (m SiteIdentity) GetPasswordHash() string {
	return m.PasswordHash
}

// SetPasswordHash sets the PasswordHash value and returns the updated SiteIdentity.
func (m *SiteIdentity) SetPasswordHash(v string) {
	m.PasswordHash = v
}

// GetTotpSecretEnc returns the TotpSecretEnc value.
func (m SiteIdentity) GetTotpSecretEnc() string {
	return m.TotpSecretEnc
}

// SetTotpSecretEnc sets the TotpSecretEnc value and returns the updated SiteIdentity.
func (m *SiteIdentity) SetTotpSecretEnc(v string) {
	m.TotpSecretEnc = v
}

// GetRole returns the Role value.
func (m SiteIdentity) GetRole() consts.IdentityRole {
	return m.Role
}

// SetRole sets the Role value and returns the updated SiteIdentity.
func (m *SiteIdentity) SetRole(v consts.IdentityRole) {
	m.Role = v
}

// GetCreatedAt returns the CreatedAt value.
func (m SiteIdentity) GetCreatedAt() int64 {
	return m.CreatedAt
}

// SetCreatedAt sets the CreatedAt value and returns the updated SiteIdentity.
func (m *SiteIdentity) SetCreatedAt(v int64) {
	m.CreatedAt = v
}

// GetLastSeenAt returns the LastSeenAt value.
func (m SiteIdentity) GetLastSeenAt() *int64 {
	return m.LastSeenAt
}

// SetLastSeenAt sets the LastSeenAt value and returns the updated SiteIdentity.
func (m *SiteIdentity) SetLastSeenAt(v *int64) {
	m.LastSeenAt = v
}

// GetLastWriteAt returns the LastWriteAt value.
func (m SiteIdentity) GetLastWriteAt() *int64 {
	return m.LastWriteAt
}

// SetLastWriteAt sets the LastWriteAt value and returns the updated SiteIdentity.
func (m *SiteIdentity) SetLastWriteAt(v *int64) {
	m.LastWriteAt = v
}

// GetUpdatedByIdentityID returns the UpdatedByIdentityID value.
func (m SiteIdentity) GetUpdatedByIdentityID() *int64 {
	return m.UpdatedByIdentityID
}

// SetUpdatedByIdentityID sets the UpdatedByIdentityID value and returns the updated SiteIdentity.
func (m *SiteIdentity) SetUpdatedByIdentityID(v *int64) {
	m.UpdatedByIdentityID = v
}

// GetUpdatedByIdentityName returns the UpdatedByIdentityName value.
func (m SiteIdentity) GetUpdatedByIdentityName() *string {
	return m.UpdatedByIdentityName
}

// SetUpdatedByIdentityName sets the UpdatedByIdentityName value and returns the updated SiteIdentity.
func (m *SiteIdentity) SetUpdatedByIdentityName(v *string) {
	m.UpdatedByIdentityName = v
}

// GetCreatedByIdentityID returns the CreatedByIdentityID value.
func (m SiteIdentity) GetCreatedByIdentityID() int64 {
	return m.CreatedByIdentityID
}

// SetCreatedByIdentityID sets the CreatedByIdentityID value and returns the updated SiteIdentity.
func (m *SiteIdentity) SetCreatedByIdentityID(v int64) {
	m.CreatedByIdentityID = v
}

// GetCreatedByIdentityName returns the CreatedByIdentityName value.
func (m SiteIdentity) GetCreatedByIdentityName() string {
	return m.CreatedByIdentityName
}

// SetCreatedByIdentityName sets the CreatedByIdentityName value and returns the updated SiteIdentity.
func (m *SiteIdentity) SetCreatedByIdentityName(v string) {
	m.CreatedByIdentityName = v
}

// GetDeletedAt returns the DeletedAt value.
func (m SiteIdentity) GetDeletedAt() *int64 {
	return m.DeletedAt
}

// SetDeletedAt sets the DeletedAt value and returns the updated SiteIdentity.
func (m *SiteIdentity) SetDeletedAt(v *int64) {
	m.DeletedAt = v
}

// GetEnabled returns the Enabled value.
func (m SiteIdentity) GetEnabled() bool {
	return m.Enabled
}

// SetEnabled sets the Enabled value and returns the updated SiteIdentity.
func (m *SiteIdentity) SetEnabled(v bool) {
	m.Enabled = v
}

// GetUnits returns the Units value.
func (m SiteIdentity) GetUnits() *consts.Units {
	return m.Units
}

// SetUnits sets the Units value and returns the updated SiteIdentity.
func (m *SiteIdentity) SetUnits(v *consts.Units) {
	m.Units = v
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

// GetAlarmUrgentLow returns the AlarmUrgentLow value.
func (m SiteIdentity) GetAlarmUrgentLow() *int64 {
	return m.AlarmUrgentLow
}

// SetAlarmUrgentLow sets the AlarmUrgentLow value and returns the updated SiteIdentity.
func (m *SiteIdentity) SetAlarmUrgentLow(v *int64) {
	m.AlarmUrgentLow = v
}

// GetAlarmLowEnabled returns the AlarmLowEnabled value.
func (m SiteIdentity) GetAlarmLowEnabled() *bool {
	return m.AlarmLowEnabled
}

// SetAlarmLowEnabled sets the AlarmLowEnabled value and returns the updated SiteIdentity.
func (m *SiteIdentity) SetAlarmLowEnabled(v *bool) {
	m.AlarmLowEnabled = v
}

// GetAlarmLow returns the AlarmLow value.
func (m SiteIdentity) GetAlarmLow() *int64 {
	return m.AlarmLow
}

// SetAlarmLow sets the AlarmLow value and returns the updated SiteIdentity.
func (m *SiteIdentity) SetAlarmLow(v *int64) {
	m.AlarmLow = v
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

// GetAlarmHighEnabled returns the AlarmHighEnabled value.
func (m SiteIdentity) GetAlarmHighEnabled() *bool {
	return m.AlarmHighEnabled
}

// SetAlarmHighEnabled sets the AlarmHighEnabled value and returns the updated SiteIdentity.
func (m *SiteIdentity) SetAlarmHighEnabled(v *bool) {
	m.AlarmHighEnabled = v
}

// GetAlarmHigh returns the AlarmHigh value.
func (m SiteIdentity) GetAlarmHigh() *int64 {
	return m.AlarmHigh
}

// SetAlarmHigh sets the AlarmHigh value and returns the updated SiteIdentity.
func (m *SiteIdentity) SetAlarmHigh(v *int64) {
	m.AlarmHigh = v
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

// GetAlarmUrgentHigh returns the AlarmUrgentHigh value.
func (m SiteIdentity) GetAlarmUrgentHigh() *int64 {
	return m.AlarmUrgentHigh
}

// SetAlarmUrgentHigh sets the AlarmUrgentHigh value and returns the updated SiteIdentity.
func (m *SiteIdentity) SetAlarmUrgentHigh(v *int64) {
	m.AlarmUrgentHigh = v
}

// GetAlarmFastRiseEnabled returns the AlarmFastRiseEnabled value.
func (m SiteIdentity) GetAlarmFastRiseEnabled() *bool {
	return m.AlarmFastRiseEnabled
}

// SetAlarmFastRiseEnabled sets the AlarmFastRiseEnabled value and returns the updated SiteIdentity.
func (m *SiteIdentity) SetAlarmFastRiseEnabled(v *bool) {
	m.AlarmFastRiseEnabled = v
}

// GetAlarmFastRise returns the AlarmFastRise value.
func (m SiteIdentity) GetAlarmFastRise() *int64 {
	return m.AlarmFastRise
}

// SetAlarmFastRise sets the AlarmFastRise value and returns the updated SiteIdentity.
func (m *SiteIdentity) SetAlarmFastRise(v *int64) {
	m.AlarmFastRise = v
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

// GetAlarmFastDrop returns the AlarmFastDrop value.
func (m SiteIdentity) GetAlarmFastDrop() *int64 {
	return m.AlarmFastDrop
}

// SetAlarmFastDrop sets the AlarmFastDrop value and returns the updated SiteIdentity.
func (m *SiteIdentity) SetAlarmFastDrop(v *int64) {
	m.AlarmFastDrop = v
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

// GetAlarmBatteryLowEnabled returns the AlarmBatteryLowEnabled value.
func (m SiteIdentity) GetAlarmBatteryLowEnabled() *bool {
	return m.AlarmBatteryLowEnabled
}

// SetAlarmBatteryLowEnabled sets the AlarmBatteryLowEnabled value and returns the updated SiteIdentity.
func (m *SiteIdentity) SetAlarmBatteryLowEnabled(v *bool) {
	m.AlarmBatteryLowEnabled = v
}

// GetAlarmBatteryLowPercent returns the AlarmBatteryLowPercent value.
func (m SiteIdentity) GetAlarmBatteryLowPercent() *int64 {
	return m.AlarmBatteryLowPercent
}

// SetAlarmBatteryLowPercent sets the AlarmBatteryLowPercent value and returns the updated SiteIdentity.
func (m *SiteIdentity) SetAlarmBatteryLowPercent(v *int64) {
	m.AlarmBatteryLowPercent = v
}

// GetAlarmBatteryLowPwdEnabled returns the AlarmBatteryLowPwdEnabled value.
func (m SiteIdentity) GetAlarmBatteryLowPwdEnabled() *bool {
	return m.AlarmBatteryLowPwdEnabled
}

// SetAlarmBatteryLowPwdEnabled sets the AlarmBatteryLowPwdEnabled value and returns the updated SiteIdentity.
func (m *SiteIdentity) SetAlarmBatteryLowPwdEnabled(v *bool) {
	m.AlarmBatteryLowPwdEnabled = v
}

// GetAlarmBatteryLowPwdPercent returns the AlarmBatteryLowPwdPercent value.
func (m SiteIdentity) GetAlarmBatteryLowPwdPercent() *int64 {
	return m.AlarmBatteryLowPwdPercent
}

// SetAlarmBatteryLowPwdPercent sets the AlarmBatteryLowPwdPercent value and returns the updated SiteIdentity.
func (m *SiteIdentity) SetAlarmBatteryLowPwdPercent(v *int64) {
	m.AlarmBatteryLowPwdPercent = v
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

// GetSiteID returns the SiteID value.
func (m SiteSyncSource) GetSiteID() string {
	return m.SiteID
}

// SetSiteID sets the SiteID value and returns the updated SiteSyncSource.
func (m *SiteSyncSource) SetSiteID(v string) {
	m.SiteID = v
}

// GetType returns the Type value.
func (m SiteSyncSource) GetType() consts.SyncSourceType {
	return m.Type
}

// SetType sets the Type value and returns the updated SiteSyncSource.
func (m *SiteSyncSource) SetType(v consts.SyncSourceType) {
	m.Type = v
}

// GetUrlOrUsernameEnc returns the UrlOrUsernameEnc value.
func (m SiteSyncSource) GetUrlOrUsernameEnc() *string {
	return m.UrlOrUsernameEnc
}

// SetUrlOrUsernameEnc sets the UrlOrUsernameEnc value and returns the updated SiteSyncSource.
func (m *SiteSyncSource) SetUrlOrUsernameEnc(v *string) {
	m.UrlOrUsernameEnc = v
}

// GetPasswordEnc returns the PasswordEnc value.
func (m SiteSyncSource) GetPasswordEnc() *string {
	return m.PasswordEnc
}

// SetPasswordEnc sets the PasswordEnc value and returns the updated SiteSyncSource.
func (m *SiteSyncSource) SetPasswordEnc(v *string) {
	m.PasswordEnc = v
}

// GetRegion returns the Region value.
func (m SiteSyncSource) GetRegion() *string {
	return m.Region
}

// SetRegion sets the Region value and returns the updated SiteSyncSource.
func (m *SiteSyncSource) SetRegion(v *string) {
	m.Region = v
}

// GetCreatedAt returns the CreatedAt value.
func (m SiteSyncSource) GetCreatedAt() int64 {
	return m.CreatedAt
}

// SetCreatedAt sets the CreatedAt value and returns the updated SiteSyncSource.
func (m *SiteSyncSource) SetCreatedAt(v int64) {
	m.CreatedAt = v
}

// GetUpdatedAt returns the UpdatedAt value.
func (m SiteSyncSource) GetUpdatedAt() int64 {
	return m.UpdatedAt
}

// SetUpdatedAt sets the UpdatedAt value and returns the updated SiteSyncSource.
func (m *SiteSyncSource) SetUpdatedAt(v int64) {
	m.UpdatedAt = v
}

// GetEnabled returns the Enabled value.
func (m SiteSyncSource) GetEnabled() bool {
	return m.Enabled
}

// SetEnabled sets the Enabled value and returns the updated SiteSyncSource.
func (m *SiteSyncSource) SetEnabled(v bool) {
	m.Enabled = v
}

// GetFirstSuccessfulSyncTime returns the FirstSuccessfulSyncTime value.
func (m SiteSyncSource) GetFirstSuccessfulSyncTime() *int64 {
	return m.FirstSuccessfulSyncTime
}

// SetFirstSuccessfulSyncTime sets the FirstSuccessfulSyncTime value and returns the updated SiteSyncSource.
func (m *SiteSyncSource) SetFirstSuccessfulSyncTime(v *int64) {
	m.FirstSuccessfulSyncTime = v
}

// GetSyncIntervalSecondsReason returns the SyncIntervalSecondsReason value.
func (m SiteSyncSource) GetSyncIntervalSecondsReason() string {
	return m.SyncIntervalSecondsReason
}

// SetSyncIntervalSecondsReason sets the SyncIntervalSecondsReason value and returns the updated SiteSyncSource.
func (m *SiteSyncSource) SetSyncIntervalSecondsReason(v string) {
	m.SyncIntervalSecondsReason = v
}

// GetSyncIntervalSeconds returns the SyncIntervalSeconds value.
func (m SiteSyncSource) GetSyncIntervalSeconds() int64 {
	return m.SyncIntervalSeconds
}

// SetSyncIntervalSeconds sets the SyncIntervalSeconds value and returns the updated SiteSyncSource.
func (m *SiteSyncSource) SetSyncIntervalSeconds(v int64) {
	m.SyncIntervalSeconds = v
}

// GetLastAttemptedSync returns the LastAttemptedSync value.
func (m SiteSyncSource) GetLastAttemptedSync() *int64 {
	return m.LastAttemptedSync
}

// SetLastAttemptedSync sets the LastAttemptedSync value and returns the updated SiteSyncSource.
func (m *SiteSyncSource) SetLastAttemptedSync(v *int64) {
	m.LastAttemptedSync = v
}

// GetLastSuccessfulSync returns the LastSuccessfulSync value.
func (m SiteSyncSource) GetLastSuccessfulSync() *int64 {
	return m.LastSuccessfulSync
}

// SetLastSuccessfulSync sets the LastSuccessfulSync value and returns the updated SiteSyncSource.
func (m *SiteSyncSource) SetLastSuccessfulSync(v *int64) {
	m.LastSuccessfulSync = v
}

// GetLastFailedSync returns the LastFailedSync value.
func (m SiteSyncSource) GetLastFailedSync() *int64 {
	return m.LastFailedSync
}

// SetLastFailedSync sets the LastFailedSync value and returns the updated SiteSyncSource.
func (m *SiteSyncSource) SetLastFailedSync(v *int64) {
	m.LastFailedSync = v
}

// GetTotalSyncs returns the TotalSyncs value.
func (m SiteSyncSource) GetTotalSyncs() *int64 {
	return m.TotalSyncs
}

// SetTotalSyncs sets the TotalSyncs value and returns the updated SiteSyncSource.
func (m *SiteSyncSource) SetTotalSyncs(v *int64) {
	m.TotalSyncs = v
}

// GetTotalSuccessfulSyncs returns the TotalSuccessfulSyncs value.
func (m SiteSyncSource) GetTotalSuccessfulSyncs() *int64 {
	return m.TotalSuccessfulSyncs
}

// SetTotalSuccessfulSyncs sets the TotalSuccessfulSyncs value and returns the updated SiteSyncSource.
func (m *SiteSyncSource) SetTotalSuccessfulSyncs(v *int64) {
	m.TotalSuccessfulSyncs = v
}

// GetTotalFailedSyncs returns the TotalFailedSyncs value.
func (m SiteSyncSource) GetTotalFailedSyncs() *int64 {
	return m.TotalFailedSyncs
}

// SetTotalFailedSyncs sets the TotalFailedSyncs value and returns the updated SiteSyncSource.
func (m *SiteSyncSource) SetTotalFailedSyncs(v *int64) {
	m.TotalFailedSyncs = v
}

// GetConsecutiveSuccessfulSyncs returns the ConsecutiveSuccessfulSyncs value.
func (m SiteSyncSource) GetConsecutiveSuccessfulSyncs() *int64 {
	return m.ConsecutiveSuccessfulSyncs
}

// SetConsecutiveSuccessfulSyncs sets the ConsecutiveSuccessfulSyncs value and returns the updated SiteSyncSource.
func (m *SiteSyncSource) SetConsecutiveSuccessfulSyncs(v *int64) {
	m.ConsecutiveSuccessfulSyncs = v
}

// GetConsecutiveFailedSyncs returns the ConsecutiveFailedSyncs value.
func (m SiteSyncSource) GetConsecutiveFailedSyncs() *int64 {
	return m.ConsecutiveFailedSyncs
}

// SetConsecutiveFailedSyncs sets the ConsecutiveFailedSyncs value and returns the updated SiteSyncSource.
func (m *SiteSyncSource) SetConsecutiveFailedSyncs(v *int64) {
	m.ConsecutiveFailedSyncs = v
}

// GetPreviousSyncErrors returns the PreviousSyncErrors value.
func (m SiteSyncSource) GetPreviousSyncErrors() errors.ErrorList {
	return m.PreviousSyncErrors
}

// SetPreviousSyncErrors sets the PreviousSyncErrors value and returns the updated SiteSyncSource.
func (m *SiteSyncSource) SetPreviousSyncErrors(v errors.ErrorList) {
	m.PreviousSyncErrors = v
}

// GetTotalEma returns the TotalEma value.
func (m SiteSyncSource) GetTotalEma() *float64 {
	return m.TotalEma
}

// SetTotalEma sets the TotalEma value and returns the updated SiteSyncSource.
func (m *SiteSyncSource) SetTotalEma(v *float64) {
	m.TotalEma = v
}

// GetSuccessEma returns the SuccessEma value.
func (m SiteSyncSource) GetSuccessEma() *float64 {
	return m.SuccessEma
}

// SetSuccessEma sets the SuccessEma value and returns the updated SiteSyncSource.
func (m *SiteSyncSource) SetSuccessEma(v *float64) {
	m.SuccessEma = v
}

// GetFailureEma returns the FailureEma value.
func (m SiteSyncSource) GetFailureEma() *float64 {
	return m.FailureEma
}

// SetFailureEma sets the FailureEma value and returns the updated SiteSyncSource.
func (m *SiteSyncSource) SetFailureEma(v *float64) {
	m.FailureEma = v
}

// GetLastSyncDuration returns the LastSyncDuration value.
func (m SiteSyncSource) GetLastSyncDuration() *int64 {
	return m.LastSyncDuration
}

// SetLastSyncDuration sets the LastSyncDuration value and returns the updated SiteSyncSource.
func (m *SiteSyncSource) SetLastSyncDuration(v *int64) {
	m.LastSyncDuration = v
}

// GetID returns the ID value.
func (m Site) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated Site.
func (m *Site) SetID(v string) {
	m.ID = v
}

// GetBatch returns the Batch value.
func (m Site) GetBatch() string {
	return m.Batch
}

// SetBatch sets the Batch value and returns the updated Site.
func (m *Site) SetBatch(v string) {
	m.Batch = v
}

// GetBackupKeyHash returns the BackupKeyHash value.
func (m Site) GetBackupKeyHash() string {
	return m.BackupKeyHash
}

// SetBackupKeyHash sets the BackupKeyHash value and returns the updated Site.
func (m *Site) SetBackupKeyHash(v string) {
	m.BackupKeyHash = v
}

// GetCreatedAt returns the CreatedAt value.
func (m Site) GetCreatedAt() int64 {
	return m.CreatedAt
}

// SetCreatedAt sets the CreatedAt value and returns the updated Site.
func (m *Site) SetCreatedAt(v int64) {
	m.CreatedAt = v
}

// GetMaxDbSizeKbLimit returns the MaxDbSizeKbLimit value.
func (m Site) GetMaxDbSizeKbLimit() *int64 {
	return m.MaxDbSizeKbLimit
}

// SetMaxDbSizeKbLimit sets the MaxDbSizeKbLimit value and returns the updated Site.
func (m *Site) SetMaxDbSizeKbLimit(v *int64) {
	m.MaxDbSizeKbLimit = v
}

// GetLastSeenAt returns the LastSeenAt value.
func (m Site) GetLastSeenAt() *int64 {
	return m.LastSeenAt
}

// SetLastSeenAt sets the LastSeenAt value and returns the updated Site.
func (m *Site) SetLastSeenAt(v *int64) {
	m.LastSeenAt = v
}

// GetLastWriteAt returns the LastWriteAt value.
func (m Site) GetLastWriteAt() *int64 {
	return m.LastWriteAt
}

// SetLastWriteAt sets the LastWriteAt value and returns the updated Site.
func (m *Site) SetLastWriteAt(v *int64) {
	m.LastWriteAt = v
}

// GetState returns the State value.
func (m Site) GetState() string {
	return m.State
}

// SetState sets the State value and returns the updated Site.
func (m *Site) SetState(v string) {
	m.State = v
}

// GetDisabledAt returns the DisabledAt value.
func (m Site) GetDisabledAt() *int64 {
	return m.DisabledAt
}

// SetDisabledAt sets the DisabledAt value and returns the updated Site.
func (m *Site) SetDisabledAt(v *int64) {
	m.DisabledAt = v
}

// GetArchivedAt returns the ArchivedAt value.
func (m Site) GetArchivedAt() *int64 {
	return m.ArchivedAt
}

// SetArchivedAt sets the ArchivedAt value and returns the updated Site.
func (m *Site) SetArchivedAt(v *int64) {
	m.ArchivedAt = v
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

// GetAlarmUrgentLow returns the AlarmUrgentLow value.
func (m Site) GetAlarmUrgentLow() *int64 {
	return m.AlarmUrgentLow
}

// SetAlarmUrgentLow sets the AlarmUrgentLow value and returns the updated Site.
func (m *Site) SetAlarmUrgentLow(v *int64) {
	m.AlarmUrgentLow = v
}

// GetAlarmLowEnabled returns the AlarmLowEnabled value.
func (m Site) GetAlarmLowEnabled() *bool {
	return m.AlarmLowEnabled
}

// SetAlarmLowEnabled sets the AlarmLowEnabled value and returns the updated Site.
func (m *Site) SetAlarmLowEnabled(v *bool) {
	m.AlarmLowEnabled = v
}

// GetAlarmLow returns the AlarmLow value.
func (m Site) GetAlarmLow() *int64 {
	return m.AlarmLow
}

// SetAlarmLow sets the AlarmLow value and returns the updated Site.
func (m *Site) SetAlarmLow(v *int64) {
	m.AlarmLow = v
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

// GetAlarmHighEnabled returns the AlarmHighEnabled value.
func (m Site) GetAlarmHighEnabled() *bool {
	return m.AlarmHighEnabled
}

// SetAlarmHighEnabled sets the AlarmHighEnabled value and returns the updated Site.
func (m *Site) SetAlarmHighEnabled(v *bool) {
	m.AlarmHighEnabled = v
}

// GetAlarmHigh returns the AlarmHigh value.
func (m Site) GetAlarmHigh() *int64 {
	return m.AlarmHigh
}

// SetAlarmHigh sets the AlarmHigh value and returns the updated Site.
func (m *Site) SetAlarmHigh(v *int64) {
	m.AlarmHigh = v
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

// GetAlarmUrgentHigh returns the AlarmUrgentHigh value.
func (m Site) GetAlarmUrgentHigh() *int64 {
	return m.AlarmUrgentHigh
}

// SetAlarmUrgentHigh sets the AlarmUrgentHigh value and returns the updated Site.
func (m *Site) SetAlarmUrgentHigh(v *int64) {
	m.AlarmUrgentHigh = v
}

// GetAlarmFastRiseEnabled returns the AlarmFastRiseEnabled value.
func (m Site) GetAlarmFastRiseEnabled() *bool {
	return m.AlarmFastRiseEnabled
}

// SetAlarmFastRiseEnabled sets the AlarmFastRiseEnabled value and returns the updated Site.
func (m *Site) SetAlarmFastRiseEnabled(v *bool) {
	m.AlarmFastRiseEnabled = v
}

// GetAlarmFastRise returns the AlarmFastRise value.
func (m Site) GetAlarmFastRise() *int64 {
	return m.AlarmFastRise
}

// SetAlarmFastRise sets the AlarmFastRise value and returns the updated Site.
func (m *Site) SetAlarmFastRise(v *int64) {
	m.AlarmFastRise = v
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

// GetAlarmFastDrop returns the AlarmFastDrop value.
func (m Site) GetAlarmFastDrop() *int64 {
	return m.AlarmFastDrop
}

// SetAlarmFastDrop sets the AlarmFastDrop value and returns the updated Site.
func (m *Site) SetAlarmFastDrop(v *int64) {
	m.AlarmFastDrop = v
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

// GetAlarmBatteryLowEnabled returns the AlarmBatteryLowEnabled value.
func (m Site) GetAlarmBatteryLowEnabled() *bool {
	return m.AlarmBatteryLowEnabled
}

// SetAlarmBatteryLowEnabled sets the AlarmBatteryLowEnabled value and returns the updated Site.
func (m *Site) SetAlarmBatteryLowEnabled(v *bool) {
	m.AlarmBatteryLowEnabled = v
}

// GetAlarmBatteryLowPercent returns the AlarmBatteryLowPercent value.
func (m Site) GetAlarmBatteryLowPercent() *int64 {
	return m.AlarmBatteryLowPercent
}

// SetAlarmBatteryLowPercent sets the AlarmBatteryLowPercent value and returns the updated Site.
func (m *Site) SetAlarmBatteryLowPercent(v *int64) {
	m.AlarmBatteryLowPercent = v
}

// GetAlarmBatteryLowPwdEnabled returns the AlarmBatteryLowPwdEnabled value.
func (m Site) GetAlarmBatteryLowPwdEnabled() *bool {
	return m.AlarmBatteryLowPwdEnabled
}

// SetAlarmBatteryLowPwdEnabled sets the AlarmBatteryLowPwdEnabled value and returns the updated Site.
func (m *Site) SetAlarmBatteryLowPwdEnabled(v *bool) {
	m.AlarmBatteryLowPwdEnabled = v
}

// GetAlarmBatteryLowPwdPercent returns the AlarmBatteryLowPwdPercent value.
func (m Site) GetAlarmBatteryLowPwdPercent() *int64 {
	return m.AlarmBatteryLowPwdPercent
}

// SetAlarmBatteryLowPwdPercent sets the AlarmBatteryLowPwdPercent value and returns the updated Site.
func (m *Site) SetAlarmBatteryLowPwdPercent(v *int64) {
	m.AlarmBatteryLowPwdPercent = v
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

// GetName returns the Name value.
func (m SqliteSequence) GetName() interface{} {
	return m.Name
}

// SetName sets the Name value and returns the updated SqliteSequence.
func (m *SqliteSequence) SetName(v interface{}) {
	m.Name = v
}

// GetSeq returns the Seq value.
func (m SqliteSequence) GetSeq() interface{} {
	return m.Seq
}

// SetSeq sets the Seq value and returns the updated SqliteSequence.
func (m *SqliteSequence) SetSeq(v interface{}) {
	m.Seq = v
}

// GetName returns the Name value.
func (m CreateControlIdentityParams) GetName() string {
	return m.Name
}

// SetName sets the Name value and returns the updated CreateControlIdentityParams.
func (m *CreateControlIdentityParams) SetName(v string) {
	m.Name = v
}

// GetPasswordHash returns the PasswordHash value.
func (m CreateControlIdentityParams) GetPasswordHash() string {
	return m.PasswordHash
}

// SetPasswordHash sets the PasswordHash value and returns the updated CreateControlIdentityParams.
func (m *CreateControlIdentityParams) SetPasswordHash(v string) {
	m.PasswordHash = v
}

// GetTotpSecretEnc returns the TotpSecretEnc value.
func (m CreateControlIdentityParams) GetTotpSecretEnc() string {
	return m.TotpSecretEnc
}

// SetTotpSecretEnc sets the TotpSecretEnc value and returns the updated CreateControlIdentityParams.
func (m *CreateControlIdentityParams) SetTotpSecretEnc(v string) {
	m.TotpSecretEnc = v
}

// GetRole returns the Role value.
func (m CreateControlIdentityParams) GetRole() consts.ControlRole {
	return m.Role
}

// SetRole sets the Role value and returns the updated CreateControlIdentityParams.
func (m *CreateControlIdentityParams) SetRole(v consts.ControlRole) {
	m.Role = v
}

// GetCreatedByIdentityID returns the CreatedByIdentityID value.
func (m CreateControlIdentityParams) GetCreatedByIdentityID() int64 {
	return m.CreatedByIdentityID
}

// SetCreatedByIdentityID sets the CreatedByIdentityID value and returns the updated CreateControlIdentityParams.
func (m *CreateControlIdentityParams) SetCreatedByIdentityID(v int64) {
	m.CreatedByIdentityID = v
}

// GetCreatedByIdentityName returns the CreatedByIdentityName value.
func (m CreateControlIdentityParams) GetCreatedByIdentityName() string {
	return m.CreatedByIdentityName
}

// SetCreatedByIdentityName sets the CreatedByIdentityName value and returns the updated CreateControlIdentityParams.
func (m *CreateControlIdentityParams) SetCreatedByIdentityName(v string) {
	m.CreatedByIdentityName = v
}

// GetName returns the Name value.
func (m GetControlIdentityParams) GetName() string {
	return m.Name
}

// SetName sets the Name value and returns the updated GetControlIdentityParams.
func (m *GetControlIdentityParams) SetName(v string) {
	m.Name = v
}

// GetUnits returns the Units value.
func (m UpdateSiteIdentitySettingsParams) GetUnits() *consts.Units {
	return m.Units
}

// SetUnits sets the Units value and returns the updated UpdateSiteIdentitySettingsParams.
func (m *UpdateSiteIdentitySettingsParams) SetUnits(v *consts.Units) {
	m.Units = v
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

// GetAlarmUrgentLow returns the AlarmUrgentLow value.
func (m UpdateSiteIdentitySettingsParams) GetAlarmUrgentLow() *int64 {
	return m.AlarmUrgentLow
}

// SetAlarmUrgentLow sets the AlarmUrgentLow value and returns the updated UpdateSiteIdentitySettingsParams.
func (m *UpdateSiteIdentitySettingsParams) SetAlarmUrgentLow(v *int64) {
	m.AlarmUrgentLow = v
}

// GetAlarmLowEnabled returns the AlarmLowEnabled value.
func (m UpdateSiteIdentitySettingsParams) GetAlarmLowEnabled() *bool {
	return m.AlarmLowEnabled
}

// SetAlarmLowEnabled sets the AlarmLowEnabled value and returns the updated UpdateSiteIdentitySettingsParams.
func (m *UpdateSiteIdentitySettingsParams) SetAlarmLowEnabled(v *bool) {
	m.AlarmLowEnabled = v
}

// GetAlarmLow returns the AlarmLow value.
func (m UpdateSiteIdentitySettingsParams) GetAlarmLow() *int64 {
	return m.AlarmLow
}

// SetAlarmLow sets the AlarmLow value and returns the updated UpdateSiteIdentitySettingsParams.
func (m *UpdateSiteIdentitySettingsParams) SetAlarmLow(v *int64) {
	m.AlarmLow = v
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

// GetAlarmHighEnabled returns the AlarmHighEnabled value.
func (m UpdateSiteIdentitySettingsParams) GetAlarmHighEnabled() *bool {
	return m.AlarmHighEnabled
}

// SetAlarmHighEnabled sets the AlarmHighEnabled value and returns the updated UpdateSiteIdentitySettingsParams.
func (m *UpdateSiteIdentitySettingsParams) SetAlarmHighEnabled(v *bool) {
	m.AlarmHighEnabled = v
}

// GetAlarmHigh returns the AlarmHigh value.
func (m UpdateSiteIdentitySettingsParams) GetAlarmHigh() *int64 {
	return m.AlarmHigh
}

// SetAlarmHigh sets the AlarmHigh value and returns the updated UpdateSiteIdentitySettingsParams.
func (m *UpdateSiteIdentitySettingsParams) SetAlarmHigh(v *int64) {
	m.AlarmHigh = v
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

// GetAlarmUrgentHigh returns the AlarmUrgentHigh value.
func (m UpdateSiteIdentitySettingsParams) GetAlarmUrgentHigh() *int64 {
	return m.AlarmUrgentHigh
}

// SetAlarmUrgentHigh sets the AlarmUrgentHigh value and returns the updated UpdateSiteIdentitySettingsParams.
func (m *UpdateSiteIdentitySettingsParams) SetAlarmUrgentHigh(v *int64) {
	m.AlarmUrgentHigh = v
}

// GetAlarmFastRiseEnabled returns the AlarmFastRiseEnabled value.
func (m UpdateSiteIdentitySettingsParams) GetAlarmFastRiseEnabled() *bool {
	return m.AlarmFastRiseEnabled
}

// SetAlarmFastRiseEnabled sets the AlarmFastRiseEnabled value and returns the updated UpdateSiteIdentitySettingsParams.
func (m *UpdateSiteIdentitySettingsParams) SetAlarmFastRiseEnabled(v *bool) {
	m.AlarmFastRiseEnabled = v
}

// GetAlarmFastRise returns the AlarmFastRise value.
func (m UpdateSiteIdentitySettingsParams) GetAlarmFastRise() *int64 {
	return m.AlarmFastRise
}

// SetAlarmFastRise sets the AlarmFastRise value and returns the updated UpdateSiteIdentitySettingsParams.
func (m *UpdateSiteIdentitySettingsParams) SetAlarmFastRise(v *int64) {
	m.AlarmFastRise = v
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

// GetAlarmFastDrop returns the AlarmFastDrop value.
func (m UpdateSiteIdentitySettingsParams) GetAlarmFastDrop() *int64 {
	return m.AlarmFastDrop
}

// SetAlarmFastDrop sets the AlarmFastDrop value and returns the updated UpdateSiteIdentitySettingsParams.
func (m *UpdateSiteIdentitySettingsParams) SetAlarmFastDrop(v *int64) {
	m.AlarmFastDrop = v
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

// GetAlarmBatteryLowEnabled returns the AlarmBatteryLowEnabled value.
func (m UpdateSiteIdentitySettingsParams) GetAlarmBatteryLowEnabled() *bool {
	return m.AlarmBatteryLowEnabled
}

// SetAlarmBatteryLowEnabled sets the AlarmBatteryLowEnabled value and returns the updated UpdateSiteIdentitySettingsParams.
func (m *UpdateSiteIdentitySettingsParams) SetAlarmBatteryLowEnabled(v *bool) {
	m.AlarmBatteryLowEnabled = v
}

// GetAlarmBatteryLowPercent returns the AlarmBatteryLowPercent value.
func (m UpdateSiteIdentitySettingsParams) GetAlarmBatteryLowPercent() *int64 {
	return m.AlarmBatteryLowPercent
}

// SetAlarmBatteryLowPercent sets the AlarmBatteryLowPercent value and returns the updated UpdateSiteIdentitySettingsParams.
func (m *UpdateSiteIdentitySettingsParams) SetAlarmBatteryLowPercent(v *int64) {
	m.AlarmBatteryLowPercent = v
}

// GetAlarmBatteryLowPwdEnabled returns the AlarmBatteryLowPwdEnabled value.
func (m UpdateSiteIdentitySettingsParams) GetAlarmBatteryLowPwdEnabled() *bool {
	return m.AlarmBatteryLowPwdEnabled
}

// SetAlarmBatteryLowPwdEnabled sets the AlarmBatteryLowPwdEnabled value and returns the updated UpdateSiteIdentitySettingsParams.
func (m *UpdateSiteIdentitySettingsParams) SetAlarmBatteryLowPwdEnabled(v *bool) {
	m.AlarmBatteryLowPwdEnabled = v
}

// GetAlarmBatteryLowPwdPercent returns the AlarmBatteryLowPwdPercent value.
func (m UpdateSiteIdentitySettingsParams) GetAlarmBatteryLowPwdPercent() *int64 {
	return m.AlarmBatteryLowPwdPercent
}

// SetAlarmBatteryLowPwdPercent sets the AlarmBatteryLowPwdPercent value and returns the updated UpdateSiteIdentitySettingsParams.
func (m *UpdateSiteIdentitySettingsParams) SetAlarmBatteryLowPwdPercent(v *int64) {
	m.AlarmBatteryLowPwdPercent = v
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

// GetSiteID returns the SiteID value.
func (m UpdateSiteIdentitySettingsParams) GetSiteID() string {
	return m.SiteID
}

// SetSiteID sets the SiteID value and returns the updated UpdateSiteIdentitySettingsParams.
func (m *UpdateSiteIdentitySettingsParams) SetSiteID(v string) {
	m.SiteID = v
}

// GetUnits returns the Units value.
func (m UpdateSitesSettingsParams) GetUnits() *consts.Units {
	return m.Units
}

// SetUnits sets the Units value and returns the updated UpdateSitesSettingsParams.
func (m *UpdateSitesSettingsParams) SetUnits(v *consts.Units) {
	m.Units = v
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

// GetAlarmUrgentLow returns the AlarmUrgentLow value.
func (m UpdateSitesSettingsParams) GetAlarmUrgentLow() *int64 {
	return m.AlarmUrgentLow
}

// SetAlarmUrgentLow sets the AlarmUrgentLow value and returns the updated UpdateSitesSettingsParams.
func (m *UpdateSitesSettingsParams) SetAlarmUrgentLow(v *int64) {
	m.AlarmUrgentLow = v
}

// GetAlarmLowEnabled returns the AlarmLowEnabled value.
func (m UpdateSitesSettingsParams) GetAlarmLowEnabled() *bool {
	return m.AlarmLowEnabled
}

// SetAlarmLowEnabled sets the AlarmLowEnabled value and returns the updated UpdateSitesSettingsParams.
func (m *UpdateSitesSettingsParams) SetAlarmLowEnabled(v *bool) {
	m.AlarmLowEnabled = v
}

// GetAlarmLow returns the AlarmLow value.
func (m UpdateSitesSettingsParams) GetAlarmLow() *int64 {
	return m.AlarmLow
}

// SetAlarmLow sets the AlarmLow value and returns the updated UpdateSitesSettingsParams.
func (m *UpdateSitesSettingsParams) SetAlarmLow(v *int64) {
	m.AlarmLow = v
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

// GetAlarmHighEnabled returns the AlarmHighEnabled value.
func (m UpdateSitesSettingsParams) GetAlarmHighEnabled() *bool {
	return m.AlarmHighEnabled
}

// SetAlarmHighEnabled sets the AlarmHighEnabled value and returns the updated UpdateSitesSettingsParams.
func (m *UpdateSitesSettingsParams) SetAlarmHighEnabled(v *bool) {
	m.AlarmHighEnabled = v
}

// GetAlarmHigh returns the AlarmHigh value.
func (m UpdateSitesSettingsParams) GetAlarmHigh() *int64 {
	return m.AlarmHigh
}

// SetAlarmHigh sets the AlarmHigh value and returns the updated UpdateSitesSettingsParams.
func (m *UpdateSitesSettingsParams) SetAlarmHigh(v *int64) {
	m.AlarmHigh = v
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

// GetAlarmUrgentHigh returns the AlarmUrgentHigh value.
func (m UpdateSitesSettingsParams) GetAlarmUrgentHigh() *int64 {
	return m.AlarmUrgentHigh
}

// SetAlarmUrgentHigh sets the AlarmUrgentHigh value and returns the updated UpdateSitesSettingsParams.
func (m *UpdateSitesSettingsParams) SetAlarmUrgentHigh(v *int64) {
	m.AlarmUrgentHigh = v
}

// GetAlarmFastRiseEnabled returns the AlarmFastRiseEnabled value.
func (m UpdateSitesSettingsParams) GetAlarmFastRiseEnabled() *bool {
	return m.AlarmFastRiseEnabled
}

// SetAlarmFastRiseEnabled sets the AlarmFastRiseEnabled value and returns the updated UpdateSitesSettingsParams.
func (m *UpdateSitesSettingsParams) SetAlarmFastRiseEnabled(v *bool) {
	m.AlarmFastRiseEnabled = v
}

// GetAlarmFastRise returns the AlarmFastRise value.
func (m UpdateSitesSettingsParams) GetAlarmFastRise() *int64 {
	return m.AlarmFastRise
}

// SetAlarmFastRise sets the AlarmFastRise value and returns the updated UpdateSitesSettingsParams.
func (m *UpdateSitesSettingsParams) SetAlarmFastRise(v *int64) {
	m.AlarmFastRise = v
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

// GetAlarmFastDrop returns the AlarmFastDrop value.
func (m UpdateSitesSettingsParams) GetAlarmFastDrop() *int64 {
	return m.AlarmFastDrop
}

// SetAlarmFastDrop sets the AlarmFastDrop value and returns the updated UpdateSitesSettingsParams.
func (m *UpdateSitesSettingsParams) SetAlarmFastDrop(v *int64) {
	m.AlarmFastDrop = v
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

// GetAlarmBatteryLowEnabled returns the AlarmBatteryLowEnabled value.
func (m UpdateSitesSettingsParams) GetAlarmBatteryLowEnabled() *bool {
	return m.AlarmBatteryLowEnabled
}

// SetAlarmBatteryLowEnabled sets the AlarmBatteryLowEnabled value and returns the updated UpdateSitesSettingsParams.
func (m *UpdateSitesSettingsParams) SetAlarmBatteryLowEnabled(v *bool) {
	m.AlarmBatteryLowEnabled = v
}

// GetAlarmBatteryLowPercent returns the AlarmBatteryLowPercent value.
func (m UpdateSitesSettingsParams) GetAlarmBatteryLowPercent() *int64 {
	return m.AlarmBatteryLowPercent
}

// SetAlarmBatteryLowPercent sets the AlarmBatteryLowPercent value and returns the updated UpdateSitesSettingsParams.
func (m *UpdateSitesSettingsParams) SetAlarmBatteryLowPercent(v *int64) {
	m.AlarmBatteryLowPercent = v
}

// GetAlarmBatteryLowPwdEnabled returns the AlarmBatteryLowPwdEnabled value.
func (m UpdateSitesSettingsParams) GetAlarmBatteryLowPwdEnabled() *bool {
	return m.AlarmBatteryLowPwdEnabled
}

// SetAlarmBatteryLowPwdEnabled sets the AlarmBatteryLowPwdEnabled value and returns the updated UpdateSitesSettingsParams.
func (m *UpdateSitesSettingsParams) SetAlarmBatteryLowPwdEnabled(v *bool) {
	m.AlarmBatteryLowPwdEnabled = v
}

// GetAlarmBatteryLowPwdPercent returns the AlarmBatteryLowPwdPercent value.
func (m UpdateSitesSettingsParams) GetAlarmBatteryLowPwdPercent() *int64 {
	return m.AlarmBatteryLowPwdPercent
}

// SetAlarmBatteryLowPwdPercent sets the AlarmBatteryLowPwdPercent value and returns the updated UpdateSitesSettingsParams.
func (m *UpdateSitesSettingsParams) SetAlarmBatteryLowPwdPercent(v *int64) {
	m.AlarmBatteryLowPwdPercent = v
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

// GetSiteID returns the SiteID value.
func (m CreateSiteIdentityParams) GetSiteID() string {
	return m.SiteID
}

// SetSiteID sets the SiteID value and returns the updated CreateSiteIdentityParams.
func (m *CreateSiteIdentityParams) SetSiteID(v string) {
	m.SiteID = v
}

// GetName returns the Name value.
func (m CreateSiteIdentityParams) GetName() string {
	return m.Name
}

// SetName sets the Name value and returns the updated CreateSiteIdentityParams.
func (m *CreateSiteIdentityParams) SetName(v string) {
	m.Name = v
}

// GetPasswordHash returns the PasswordHash value.
func (m CreateSiteIdentityParams) GetPasswordHash() string {
	return m.PasswordHash
}

// SetPasswordHash sets the PasswordHash value and returns the updated CreateSiteIdentityParams.
func (m *CreateSiteIdentityParams) SetPasswordHash(v string) {
	m.PasswordHash = v
}

// GetTotpSecretEnc returns the TotpSecretEnc value.
func (m CreateSiteIdentityParams) GetTotpSecretEnc() string {
	return m.TotpSecretEnc
}

// SetTotpSecretEnc sets the TotpSecretEnc value and returns the updated CreateSiteIdentityParams.
func (m *CreateSiteIdentityParams) SetTotpSecretEnc(v string) {
	m.TotpSecretEnc = v
}

// GetRole returns the Role value.
func (m CreateSiteIdentityParams) GetRole() consts.IdentityRole {
	return m.Role
}

// SetRole sets the Role value and returns the updated CreateSiteIdentityParams.
func (m *CreateSiteIdentityParams) SetRole(v consts.IdentityRole) {
	m.Role = v
}

// GetCreatedByIdentityID returns the CreatedByIdentityID value.
func (m CreateSiteIdentityParams) GetCreatedByIdentityID() int64 {
	return m.CreatedByIdentityID
}

// SetCreatedByIdentityID sets the CreatedByIdentityID value and returns the updated CreateSiteIdentityParams.
func (m *CreateSiteIdentityParams) SetCreatedByIdentityID(v int64) {
	m.CreatedByIdentityID = v
}

// GetCreatedByIdentityName returns the CreatedByIdentityName value.
func (m CreateSiteIdentityParams) GetCreatedByIdentityName() string {
	return m.CreatedByIdentityName
}

// SetCreatedByIdentityName sets the CreatedByIdentityName value and returns the updated CreateSiteIdentityParams.
func (m *CreateSiteIdentityParams) SetCreatedByIdentityName(v string) {
	m.CreatedByIdentityName = v
}

// GetSiteID returns the SiteID value.
func (m AllSiteIdentitiesParams) GetSiteID() string {
	return m.SiteID
}

// SetSiteID sets the SiteID value and returns the updated AllSiteIdentitiesParams.
func (m *AllSiteIdentitiesParams) SetSiteID(v string) {
	m.SiteID = v
}

// GetSiteID returns the SiteID value.
func (m GetSiteIdentityParams) GetSiteID() string {
	return m.SiteID
}

// SetSiteID sets the SiteID value and returns the updated GetSiteIdentityParams.
func (m *GetSiteIdentityParams) SetSiteID(v string) {
	m.SiteID = v
}

// GetName returns the Name value.
func (m GetSiteIdentityParams) GetName() string {
	return m.Name
}

// SetName sets the Name value and returns the updated GetSiteIdentityParams.
func (m *GetSiteIdentityParams) SetName(v string) {
	m.Name = v
}

// GetSiteID returns the SiteID value.
func (m GetSiteIdentityByIDParams) GetSiteID() string {
	return m.SiteID
}

// SetSiteID sets the SiteID value and returns the updated GetSiteIdentityByIDParams.
func (m *GetSiteIdentityByIDParams) SetSiteID(v string) {
	m.SiteID = v
}

// GetID returns the ID value.
func (m GetSiteIdentityByIDParams) GetID() int64 {
	return m.ID
}

// SetID sets the ID value and returns the updated GetSiteIdentityByIDParams.
func (m *GetSiteIdentityByIDParams) SetID(v int64) {
	m.ID = v
}

// GetLastSeenAt returns the LastSeenAt value.
func (m UpdateSiteIdentityDateFlagsParams) GetLastSeenAt() *int64 {
	return m.LastSeenAt
}

// SetLastSeenAt sets the LastSeenAt value and returns the updated UpdateSiteIdentityDateFlagsParams.
func (m *UpdateSiteIdentityDateFlagsParams) SetLastSeenAt(v *int64) {
	m.LastSeenAt = v
}

// GetLastWriteAt returns the LastWriteAt value.
func (m UpdateSiteIdentityDateFlagsParams) GetLastWriteAt() *int64 {
	return m.LastWriteAt
}

// SetLastWriteAt sets the LastWriteAt value and returns the updated UpdateSiteIdentityDateFlagsParams.
func (m *UpdateSiteIdentityDateFlagsParams) SetLastWriteAt(v *int64) {
	m.LastWriteAt = v
}

// GetID returns the ID value.
func (m UpdateSiteIdentityDateFlagsParams) GetID() int64 {
	return m.ID
}

// SetID sets the ID value and returns the updated UpdateSiteIdentityDateFlagsParams.
func (m *UpdateSiteIdentityDateFlagsParams) SetID(v int64) {
	m.ID = v
}

// GetEnabled returns the Enabled value.
func (m UpdateSiteIdentityEnabledParams) GetEnabled() bool {
	return m.Enabled
}

// SetEnabled sets the Enabled value and returns the updated UpdateSiteIdentityEnabledParams.
func (m *UpdateSiteIdentityEnabledParams) SetEnabled(v bool) {
	m.Enabled = v
}

// GetUpdatedByIdentityID returns the UpdatedByIdentityID value.
func (m UpdateSiteIdentityEnabledParams) GetUpdatedByIdentityID() *int64 {
	return m.UpdatedByIdentityID
}

// SetUpdatedByIdentityID sets the UpdatedByIdentityID value and returns the updated UpdateSiteIdentityEnabledParams.
func (m *UpdateSiteIdentityEnabledParams) SetUpdatedByIdentityID(v *int64) {
	m.UpdatedByIdentityID = v
}

// GetUpdatedByIdentityName returns the UpdatedByIdentityName value.
func (m UpdateSiteIdentityEnabledParams) GetUpdatedByIdentityName() *string {
	return m.UpdatedByIdentityName
}

// SetUpdatedByIdentityName sets the UpdatedByIdentityName value and returns the updated UpdateSiteIdentityEnabledParams.
func (m *UpdateSiteIdentityEnabledParams) SetUpdatedByIdentityName(v *string) {
	m.UpdatedByIdentityName = v
}

// GetSiteID returns the SiteID value.
func (m UpdateSiteIdentityEnabledParams) GetSiteID() string {
	return m.SiteID
}

// SetSiteID sets the SiteID value and returns the updated UpdateSiteIdentityEnabledParams.
func (m *UpdateSiteIdentityEnabledParams) SetSiteID(v string) {
	m.SiteID = v
}

// GetID returns the ID value.
func (m UpdateSiteIdentityEnabledParams) GetID() int64 {
	return m.ID
}

// SetID sets the ID value and returns the updated UpdateSiteIdentityEnabledParams.
func (m *UpdateSiteIdentityEnabledParams) SetID(v int64) {
	m.ID = v
}

// GetUpdatedByIdentityID returns the UpdatedByIdentityID value.
func (m DeleteSiteIdentityParams) GetUpdatedByIdentityID() *int64 {
	return m.UpdatedByIdentityID
}

// SetUpdatedByIdentityID sets the UpdatedByIdentityID value and returns the updated DeleteSiteIdentityParams.
func (m *DeleteSiteIdentityParams) SetUpdatedByIdentityID(v *int64) {
	m.UpdatedByIdentityID = v
}

// GetUpdatedByIdentityName returns the UpdatedByIdentityName value.
func (m DeleteSiteIdentityParams) GetUpdatedByIdentityName() *string {
	return m.UpdatedByIdentityName
}

// SetUpdatedByIdentityName sets the UpdatedByIdentityName value and returns the updated DeleteSiteIdentityParams.
func (m *DeleteSiteIdentityParams) SetUpdatedByIdentityName(v *string) {
	m.UpdatedByIdentityName = v
}

// GetSiteID returns the SiteID value.
func (m DeleteSiteIdentityParams) GetSiteID() string {
	return m.SiteID
}

// SetSiteID sets the SiteID value and returns the updated DeleteSiteIdentityParams.
func (m *DeleteSiteIdentityParams) SetSiteID(v string) {
	m.SiteID = v
}

// GetID returns the ID value.
func (m DeleteSiteIdentityParams) GetID() int64 {
	return m.ID
}

// SetID sets the ID value and returns the updated DeleteSiteIdentityParams.
func (m *DeleteSiteIdentityParams) SetID(v int64) {
	m.ID = v
}

// GetType returns the Type value.
func (m GetSyncSourceParams) GetType() consts.SyncSourceType {
	return m.Type
}

// SetType sets the Type value and returns the updated GetSyncSourceParams.
func (m *GetSyncSourceParams) SetType(v consts.SyncSourceType) {
	m.Type = v
}

// GetSiteID returns the SiteID value.
func (m GetSyncSourceParams) GetSiteID() string {
	return m.SiteID
}

// SetSiteID sets the SiteID value and returns the updated GetSyncSourceParams.
func (m *GetSyncSourceParams) SetSiteID(v string) {
	m.SiteID = v
}

// GetSiteID returns the SiteID value.
func (m GetAllSyncSourcesForSiteParams) GetSiteID() string {
	return m.SiteID
}

// SetSiteID sets the SiteID value and returns the updated GetAllSyncSourcesForSiteParams.
func (m *GetAllSyncSourcesForSiteParams) SetSiteID(v string) {
	m.SiteID = v
}

// GetSiteID returns the SiteID value.
func (m CreateSyncSourceParams) GetSiteID() string {
	return m.SiteID
}

// SetSiteID sets the SiteID value and returns the updated CreateSyncSourceParams.
func (m *CreateSyncSourceParams) SetSiteID(v string) {
	m.SiteID = v
}

// GetType returns the Type value.
func (m CreateSyncSourceParams) GetType() consts.SyncSourceType {
	return m.Type
}

// SetType sets the Type value and returns the updated CreateSyncSourceParams.
func (m *CreateSyncSourceParams) SetType(v consts.SyncSourceType) {
	m.Type = v
}

// GetUrlOrUsernameEnc returns the UrlOrUsernameEnc value.
func (m CreateSyncSourceParams) GetUrlOrUsernameEnc() *string {
	return m.UrlOrUsernameEnc
}

// SetUrlOrUsernameEnc sets the UrlOrUsernameEnc value and returns the updated CreateSyncSourceParams.
func (m *CreateSyncSourceParams) SetUrlOrUsernameEnc(v *string) {
	m.UrlOrUsernameEnc = v
}

// GetPasswordEnc returns the PasswordEnc value.
func (m CreateSyncSourceParams) GetPasswordEnc() *string {
	return m.PasswordEnc
}

// SetPasswordEnc sets the PasswordEnc value and returns the updated CreateSyncSourceParams.
func (m *CreateSyncSourceParams) SetPasswordEnc(v *string) {
	m.PasswordEnc = v
}

// GetRegion returns the Region value.
func (m CreateSyncSourceParams) GetRegion() *string {
	return m.Region
}

// SetRegion sets the Region value and returns the updated CreateSyncSourceParams.
func (m *CreateSyncSourceParams) SetRegion(v *string) {
	m.Region = v
}

// GetEnabled returns the Enabled value.
func (m CreateSyncSourceParams) GetEnabled() bool {
	return m.Enabled
}

// SetEnabled sets the Enabled value and returns the updated CreateSyncSourceParams.
func (m *CreateSyncSourceParams) SetEnabled(v bool) {
	m.Enabled = v
}

// GetPasswordEnc returns the PasswordEnc value.
func (m UpdateSyncSourceParams) GetPasswordEnc() *string {
	return m.PasswordEnc
}

// SetPasswordEnc sets the PasswordEnc value and returns the updated UpdateSyncSourceParams.
func (m *UpdateSyncSourceParams) SetPasswordEnc(v *string) {
	m.PasswordEnc = v
}

// GetEnabled returns the Enabled value.
func (m UpdateSyncSourceParams) GetEnabled() bool {
	return m.Enabled
}

// SetEnabled sets the Enabled value and returns the updated UpdateSyncSourceParams.
func (m *UpdateSyncSourceParams) SetEnabled(v bool) {
	m.Enabled = v
}

// GetUrlOrUsernameEnc returns the UrlOrUsernameEnc value.
func (m UpdateSyncSourceParams) GetUrlOrUsernameEnc() *string {
	return m.UrlOrUsernameEnc
}

// SetUrlOrUsernameEnc sets the UrlOrUsernameEnc value and returns the updated UpdateSyncSourceParams.
func (m *UpdateSyncSourceParams) SetUrlOrUsernameEnc(v *string) {
	m.UrlOrUsernameEnc = v
}

// GetRegion returns the Region value.
func (m UpdateSyncSourceParams) GetRegion() *string {
	return m.Region
}

// SetRegion sets the Region value and returns the updated UpdateSyncSourceParams.
func (m *UpdateSyncSourceParams) SetRegion(v *string) {
	m.Region = v
}

// GetType returns the Type value.
func (m UpdateSyncSourceParams) GetType() consts.SyncSourceType {
	return m.Type
}

// SetType sets the Type value and returns the updated UpdateSyncSourceParams.
func (m *UpdateSyncSourceParams) SetType(v consts.SyncSourceType) {
	m.Type = v
}

// GetSiteID returns the SiteID value.
func (m UpdateSyncSourceParams) GetSiteID() string {
	return m.SiteID
}

// SetSiteID sets the SiteID value and returns the updated UpdateSyncSourceParams.
func (m *UpdateSyncSourceParams) SetSiteID(v string) {
	m.SiteID = v
}

// GetType returns the Type value.
func (m DeleteSyncSourceParams) GetType() consts.SyncSourceType {
	return m.Type
}

// SetType sets the Type value and returns the updated DeleteSyncSourceParams.
func (m *DeleteSyncSourceParams) SetType(v consts.SyncSourceType) {
	m.Type = v
}

// GetSiteID returns the SiteID value.
func (m DeleteSyncSourceParams) GetSiteID() string {
	return m.SiteID
}

// SetSiteID sets the SiteID value and returns the updated DeleteSyncSourceParams.
func (m *DeleteSyncSourceParams) SetSiteID(v string) {
	m.SiteID = v
}

// GetType returns the Type value.
func (m UpdateSyncReportAttemptParams) GetType() consts.SyncSourceType {
	return m.Type
}

// SetType sets the Type value and returns the updated UpdateSyncReportAttemptParams.
func (m *UpdateSyncReportAttemptParams) SetType(v consts.SyncSourceType) {
	m.Type = v
}

// GetSiteID returns the SiteID value.
func (m UpdateSyncReportAttemptParams) GetSiteID() string {
	return m.SiteID
}

// SetSiteID sets the SiteID value and returns the updated UpdateSyncReportAttemptParams.
func (m *UpdateSyncReportAttemptParams) SetSiteID(v string) {
	m.SiteID = v
}

// GetElapsedMs returns the ElapsedMs value.
func (m UpdateSyncSourceReportsSuccessParams) GetElapsedMs() *int64 {
	return m.ElapsedMs
}

// SetElapsedMs sets the ElapsedMs value and returns the updated UpdateSyncSourceReportsSuccessParams.
func (m *UpdateSyncSourceReportsSuccessParams) SetElapsedMs(v *int64) {
	m.ElapsedMs = v
}

// GetType returns the Type value.
func (m UpdateSyncSourceReportsSuccessParams) GetType() consts.SyncSourceType {
	return m.Type
}

// SetType sets the Type value and returns the updated UpdateSyncSourceReportsSuccessParams.
func (m *UpdateSyncSourceReportsSuccessParams) SetType(v consts.SyncSourceType) {
	m.Type = v
}

// GetSiteID returns the SiteID value.
func (m UpdateSyncSourceReportsSuccessParams) GetSiteID() string {
	return m.SiteID
}

// SetSiteID sets the SiteID value and returns the updated UpdateSyncSourceReportsSuccessParams.
func (m *UpdateSyncSourceReportsSuccessParams) SetSiteID(v string) {
	m.SiteID = v
}

// GetErrorMessages returns the ErrorMessages value.
func (m UpdateSyncSourceReportsFailureParams) GetErrorMessages() errors.ErrorList {
	return m.ErrorMessages
}

// SetErrorMessages sets the ErrorMessages value and returns the updated UpdateSyncSourceReportsFailureParams.
func (m *UpdateSyncSourceReportsFailureParams) SetErrorMessages(v errors.ErrorList) {
	m.ErrorMessages = v
}

// GetElapsedMs returns the ElapsedMs value.
func (m UpdateSyncSourceReportsFailureParams) GetElapsedMs() *int64 {
	return m.ElapsedMs
}

// SetElapsedMs sets the ElapsedMs value and returns the updated UpdateSyncSourceReportsFailureParams.
func (m *UpdateSyncSourceReportsFailureParams) SetElapsedMs(v *int64) {
	m.ElapsedMs = v
}

// GetType returns the Type value.
func (m UpdateSyncSourceReportsFailureParams) GetType() consts.SyncSourceType {
	return m.Type
}

// SetType sets the Type value and returns the updated UpdateSyncSourceReportsFailureParams.
func (m *UpdateSyncSourceReportsFailureParams) SetType(v consts.SyncSourceType) {
	m.Type = v
}

// GetSiteID returns the SiteID value.
func (m UpdateSyncSourceReportsFailureParams) GetSiteID() string {
	return m.SiteID
}

// SetSiteID sets the SiteID value and returns the updated UpdateSyncSourceReportsFailureParams.
func (m *UpdateSyncSourceReportsFailureParams) SetSiteID(v string) {
	m.SiteID = v
}

// GetID returns the ID value.
func (m CreateSiteParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated CreateSiteParams.
func (m *CreateSiteParams) SetID(v string) {
	m.ID = v
}

// GetBackupKeyHash returns the BackupKeyHash value.
func (m CreateSiteParams) GetBackupKeyHash() string {
	return m.BackupKeyHash
}

// SetBackupKeyHash sets the BackupKeyHash value and returns the updated CreateSiteParams.
func (m *CreateSiteParams) SetBackupKeyHash(v string) {
	m.BackupKeyHash = v
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
func (m GetSiteParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated GetSiteParams.
func (m *GetSiteParams) SetID(v string) {
	m.ID = v
}

// GetLastSeenAt returns the LastSeenAt value.
func (m GetSitesToDisableParams) GetLastSeenAt() float64 {
	return m.LastSeenAt
}

// SetLastSeenAt sets the LastSeenAt value and returns the updated GetSitesToDisableParams.
func (m *GetSitesToDisableParams) SetLastSeenAt(v float64) {
	m.LastSeenAt = v
}

// GetLastWriteAt returns the LastWriteAt value.
func (m GetSitesToDisableParams) GetLastWriteAt() float64 {
	return m.LastWriteAt
}

// SetLastWriteAt sets the LastWriteAt value and returns the updated GetSitesToDisableParams.
func (m *GetSitesToDisableParams) SetLastWriteAt(v float64) {
	m.LastWriteAt = v
}

// GetLastSeenAt returns the LastSeenAt value.
func (m GetSitesToArchiveParams) GetLastSeenAt() float64 {
	return m.LastSeenAt
}

// SetLastSeenAt sets the LastSeenAt value and returns the updated GetSitesToArchiveParams.
func (m *GetSitesToArchiveParams) SetLastSeenAt(v float64) {
	m.LastSeenAt = v
}

// GetLastWriteAt returns the LastWriteAt value.
func (m GetSitesToArchiveParams) GetLastWriteAt() float64 {
	return m.LastWriteAt
}

// SetLastWriteAt sets the LastWriteAt value and returns the updated GetSitesToArchiveParams.
func (m *GetSitesToArchiveParams) SetLastWriteAt(v float64) {
	m.LastWriteAt = v
}

// GetLastSeenAt returns the LastSeenAt value.
func (m UpdateSiteDateFlagsParams) GetLastSeenAt() *int64 {
	return m.LastSeenAt
}

// SetLastSeenAt sets the LastSeenAt value and returns the updated UpdateSiteDateFlagsParams.
func (m *UpdateSiteDateFlagsParams) SetLastSeenAt(v *int64) {
	m.LastSeenAt = v
}

// GetLastWriteAt returns the LastWriteAt value.
func (m UpdateSiteDateFlagsParams) GetLastWriteAt() *int64 {
	return m.LastWriteAt
}

// SetLastWriteAt sets the LastWriteAt value and returns the updated UpdateSiteDateFlagsParams.
func (m *UpdateSiteDateFlagsParams) SetLastWriteAt(v *int64) {
	m.LastWriteAt = v
}

// GetID returns the ID value.
func (m UpdateSiteDateFlagsParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated UpdateSiteDateFlagsParams.
func (m *UpdateSiteDateFlagsParams) SetID(v string) {
	m.ID = v
}

// GetJsonEach returns the JsonEach value.
func (m DisableSitesParams) GetJsonEach() interface{} {
	return m.JsonEach
}

// SetJsonEach sets the JsonEach value and returns the updated DisableSitesParams.
func (m *DisableSitesParams) SetJsonEach(v interface{}) {
	m.JsonEach = v
}

// GetJsonEach returns the JsonEach value.
func (m ArchiveSitesParams) GetJsonEach() interface{} {
	return m.JsonEach
}

// SetJsonEach sets the JsonEach value and returns the updated ArchiveSitesParams.
func (m *ArchiveSitesParams) SetJsonEach(v interface{}) {
	m.JsonEach = v
}

// GetJsonEach returns the JsonEach value.
func (m DeleteSitesParams) GetJsonEach() interface{} {
	return m.JsonEach
}

// SetJsonEach sets the JsonEach value and returns the updated DeleteSitesParams.
func (m *DeleteSitesParams) SetJsonEach(v interface{}) {
	m.JsonEach = v
}

// GetBatch returns the Batch value.
func (m GetBatchSitesToBackUpParams) GetBatch() string {
	return m.Batch
}

// SetBatch sets the Batch value and returns the updated GetBatchSitesToBackUpParams.
func (m *GetBatchSitesToBackUpParams) SetBatch(v string) {
	m.Batch = v
}

// GetBatch returns the Batch value.
func (m GetSitesToDeleteInBatchParams) GetBatch() string {
	return m.Batch
}

// SetBatch sets the Batch value and returns the updated GetSitesToDeleteInBatchParams.
func (m *GetSitesToDeleteInBatchParams) SetBatch(v string) {
	m.Batch = v
}

// GetLastSeenAt returns the LastSeenAt value.
func (m GetSitesToDeleteInBatchParams) GetLastSeenAt() float64 {
	return m.LastSeenAt
}

// SetLastSeenAt sets the LastSeenAt value and returns the updated GetSitesToDeleteInBatchParams.
func (m *GetSitesToDeleteInBatchParams) SetLastSeenAt(v float64) {
	m.LastSeenAt = v
}

