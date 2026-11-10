
package controldb

import (
	"nighttroop.t1d.nz/data/types"
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
func (m Passkey) GetID() int64 {
	return m.ID
}

// SetID sets the ID value and returns the updated Passkey.
func (m *Passkey) SetID(v int64) {
	m.ID = v
}

// GetSiteID returns the SiteID value.
func (m Passkey) GetSiteID() *string {
	return m.SiteID
}

// SetSiteID sets the SiteID value and returns the updated Passkey.
func (m *Passkey) SetSiteID(v *string) {
	m.SiteID = v
}

// GetIdentityID returns the IdentityID value.
func (m Passkey) GetIdentityID() *int64 {
	return m.IdentityID
}

// SetIdentityID sets the IdentityID value and returns the updated Passkey.
func (m *Passkey) SetIdentityID(v *int64) {
	m.IdentityID = v
}

// GetControlIdentityID returns the ControlIdentityID value.
func (m Passkey) GetControlIdentityID() *int64 {
	return m.ControlIdentityID
}

// SetControlIdentityID sets the ControlIdentityID value and returns the updated Passkey.
func (m *Passkey) SetControlIdentityID(v *int64) {
	m.ControlIdentityID = v
}

// GetCredentialID returns the CredentialID value.
func (m Passkey) GetCredentialID() string {
	return m.CredentialID
}

// SetCredentialID sets the CredentialID value and returns the updated Passkey.
func (m *Passkey) SetCredentialID(v string) {
	m.CredentialID = v
}

// GetCredentialJson returns the CredentialJson value.
func (m Passkey) GetCredentialJson() string {
	return m.CredentialJson
}

// SetCredentialJson sets the CredentialJson value and returns the updated Passkey.
func (m *Passkey) SetCredentialJson(v string) {
	m.CredentialJson = v
}

// GetCreatedAt returns the CreatedAt value.
func (m Passkey) GetCreatedAt() int64 {
	return m.CreatedAt
}

// SetCreatedAt sets the CreatedAt value and returns the updated Passkey.
func (m *Passkey) SetCreatedAt(v int64) {
	m.CreatedAt = v
}

// GetLastUsedAt returns the LastUsedAt value.
func (m Passkey) GetLastUsedAt() *int64 {
	return m.LastUsedAt
}

// SetLastUsedAt sets the LastUsedAt value and returns the updated Passkey.
func (m *Passkey) SetLastUsedAt(v *int64) {
	m.LastUsedAt = v
}

// GetID returns the ID value.
func (m PushCredential) GetID() int64 {
	return m.ID
}

// SetID sets the ID value and returns the updated PushCredential.
func (m *PushCredential) SetID(v int64) {
	m.ID = v
}

// GetSiteID returns the SiteID value.
func (m PushCredential) GetSiteID() *string {
	return m.SiteID
}

// SetSiteID sets the SiteID value and returns the updated PushCredential.
func (m *PushCredential) SetSiteID(v *string) {
	m.SiteID = v
}

// GetName returns the Name value.
func (m PushCredential) GetName() string {
	return m.Name
}

// SetName sets the Name value and returns the updated PushCredential.
func (m *PushCredential) SetName(v string) {
	m.Name = v
}

// GetProvider returns the Provider value.
func (m PushCredential) GetProvider() consts.PushProvider {
	return m.Provider
}

// SetProvider sets the Provider value and returns the updated PushCredential.
func (m *PushCredential) SetProvider(v consts.PushProvider) {
	m.Provider = v
}

// GetTeamOrProjectID returns the TeamOrProjectID value.
func (m PushCredential) GetTeamOrProjectID() string {
	return m.TeamOrProjectID
}

// SetTeamOrProjectID sets the TeamOrProjectID value and returns the updated PushCredential.
func (m *PushCredential) SetTeamOrProjectID(v string) {
	m.TeamOrProjectID = v
}

// GetKeyID returns the KeyID value.
func (m PushCredential) GetKeyID() *string {
	return m.KeyID
}

// SetKeyID sets the KeyID value and returns the updated PushCredential.
func (m *PushCredential) SetKeyID(v *string) {
	m.KeyID = v
}

// GetSecretEnc returns the SecretEnc value.
func (m PushCredential) GetSecretEnc() *string {
	return m.SecretEnc
}

// SetSecretEnc sets the SecretEnc value and returns the updated PushCredential.
func (m *PushCredential) SetSecretEnc(v *string) {
	m.SecretEnc = v
}

// GetAppsJson returns the AppsJson value.
func (m PushCredential) GetAppsJson() string {
	return m.AppsJson
}

// SetAppsJson sets the AppsJson value and returns the updated PushCredential.
func (m *PushCredential) SetAppsJson(v string) {
	m.AppsJson = v
}

// GetBundleIdsJson returns the BundleIdsJson value.
func (m PushCredential) GetBundleIdsJson() *string {
	return m.BundleIdsJson
}

// SetBundleIdsJson sets the BundleIdsJson value and returns the updated PushCredential.
func (m *PushCredential) SetBundleIdsJson(v *string) {
	m.BundleIdsJson = v
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

// GetDeletedAt returns the DeletedAt value.
func (m PushCredential) GetDeletedAt() *int64 {
	return m.DeletedAt
}

// SetDeletedAt sets the DeletedAt value and returns the updated PushCredential.
func (m *PushCredential) SetDeletedAt(v *int64) {
	m.DeletedAt = v
}

// GetID returns the ID value.
func (m PushRegistration) GetID() int64 {
	return m.ID
}

// SetID sets the ID value and returns the updated PushRegistration.
func (m *PushRegistration) SetID(v int64) {
	m.ID = v
}

// GetSiteID returns the SiteID value.
func (m PushRegistration) GetSiteID() string {
	return m.SiteID
}

// SetSiteID sets the SiteID value and returns the updated PushRegistration.
func (m *PushRegistration) SetSiteID(v string) {
	m.SiteID = v
}

// GetIdentityID returns the IdentityID value.
func (m PushRegistration) GetIdentityID() int64 {
	return m.IdentityID
}

// SetIdentityID sets the IdentityID value and returns the updated PushRegistration.
func (m *PushRegistration) SetIdentityID(v int64) {
	m.IdentityID = v
}

// GetApp returns the App value.
func (m PushRegistration) GetApp() consts.PushApp {
	return m.App
}

// SetApp sets the App value and returns the updated PushRegistration.
func (m *PushRegistration) SetApp(v consts.PushApp) {
	m.App = v
}

// GetProvider returns the Provider value.
func (m PushRegistration) GetProvider() consts.PushProvider {
	return m.Provider
}

// SetProvider sets the Provider value and returns the updated PushRegistration.
func (m *PushRegistration) SetProvider(v consts.PushProvider) {
	m.Provider = v
}

// GetBundleID returns the BundleID value.
func (m PushRegistration) GetBundleID() string {
	return m.BundleID
}

// SetBundleID sets the BundleID value and returns the updated PushRegistration.
func (m *PushRegistration) SetBundleID(v string) {
	m.BundleID = v
}

// GetProduction returns the Production value.
func (m PushRegistration) GetProduction() bool {
	return m.Production
}

// SetProduction sets the Production value and returns the updated PushRegistration.
func (m *PushRegistration) SetProduction(v bool) {
	m.Production = v
}

// GetTokenEnc returns the TokenEnc value.
func (m PushRegistration) GetTokenEnc() string {
	return m.TokenEnc
}

// SetTokenEnc sets the TokenEnc value and returns the updated PushRegistration.
func (m *PushRegistration) SetTokenEnc(v string) {
	m.TokenEnc = v
}

// GetTokenHash returns the TokenHash value.
func (m PushRegistration) GetTokenHash() string {
	return m.TokenHash
}

// SetTokenHash sets the TokenHash value and returns the updated PushRegistration.
func (m *PushRegistration) SetTokenHash(v string) {
	m.TokenHash = v
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

// GetLastPushedAt returns the LastPushedAt value.
func (m PushRegistration) GetLastPushedAt() *int64 {
	return m.LastPushedAt
}

// SetLastPushedAt sets the LastPushedAt value and returns the updated PushRegistration.
func (m *PushRegistration) SetLastPushedAt(v *int64) {
	m.LastPushedAt = v
}

// GetLastError returns the LastError value.
func (m PushRegistration) GetLastError() *string {
	return m.LastError
}

// SetLastError sets the LastError value and returns the updated PushRegistration.
func (m *PushRegistration) SetLastError(v *string) {
	m.LastError = v
}

// GetID returns the ID value.
func (m ServerSetting) GetID() int64 {
	return m.ID
}

// SetID sets the ID value and returns the updated ServerSetting.
func (m *ServerSetting) SetID(v int64) {
	m.ID = v
}

// GetPublicSiteCreation returns the PublicSiteCreation value.
func (m ServerSetting) GetPublicSiteCreation() bool {
	return m.PublicSiteCreation
}

// SetPublicSiteCreation sets the PublicSiteCreation value and returns the updated ServerSetting.
func (m *ServerSetting) SetPublicSiteCreation(v bool) {
	m.PublicSiteCreation = v
}

// GetRateLimitsJson returns the RateLimitsJson value.
func (m ServerSetting) GetRateLimitsJson() string {
	return m.RateLimitsJson
}

// SetRateLimitsJson sets the RateLimitsJson value and returns the updated ServerSetting.
func (m *ServerSetting) SetRateLimitsJson(v string) {
	m.RateLimitsJson = v
}

// GetUpdatedAt returns the UpdatedAt value.
func (m ServerSetting) GetUpdatedAt() *int64 {
	return m.UpdatedAt
}

// SetUpdatedAt sets the UpdatedAt value and returns the updated ServerSetting.
func (m *ServerSetting) SetUpdatedAt(v *int64) {
	m.UpdatedAt = v
}

// GetUpdatedByIdentityName returns the UpdatedByIdentityName value.
func (m ServerSetting) GetUpdatedByIdentityName() *string {
	return m.UpdatedByIdentityName
}

// SetUpdatedByIdentityName sets the UpdatedByIdentityName value and returns the updated ServerSetting.
func (m *ServerSetting) SetUpdatedByIdentityName(v *string) {
	m.UpdatedByIdentityName = v
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

// GetEstimatedDbSizeKb returns the EstimatedDbSizeKb value.
func (m Site) GetEstimatedDbSizeKb() *int64 {
	return m.EstimatedDbSizeKb
}

// SetEstimatedDbSizeKb sets the EstimatedDbSizeKb value and returns the updated Site.
func (m *Site) SetEstimatedDbSizeKb(v *int64) {
	m.EstimatedDbSizeKb = v
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

// GetSiteID returns the SiteID value.
func (m SiteConnector) GetSiteID() string {
	return m.SiteID
}

// SetSiteID sets the SiteID value and returns the updated SiteConnector.
func (m *SiteConnector) SetSiteID(v string) {
	m.SiteID = v
}

// GetType returns the Type value.
func (m SiteConnector) GetType() consts.ConnectorType {
	return m.Type
}

// SetType sets the Type value and returns the updated SiteConnector.
func (m *SiteConnector) SetType(v consts.ConnectorType) {
	m.Type = v
}

// GetUrlOrUsernameEnc returns the UrlOrUsernameEnc value.
func (m SiteConnector) GetUrlOrUsernameEnc() *string {
	return m.UrlOrUsernameEnc
}

// SetUrlOrUsernameEnc sets the UrlOrUsernameEnc value and returns the updated SiteConnector.
func (m *SiteConnector) SetUrlOrUsernameEnc(v *string) {
	m.UrlOrUsernameEnc = v
}

// GetPasswordEnc returns the PasswordEnc value.
func (m SiteConnector) GetPasswordEnc() *string {
	return m.PasswordEnc
}

// SetPasswordEnc sets the PasswordEnc value and returns the updated SiteConnector.
func (m *SiteConnector) SetPasswordEnc(v *string) {
	m.PasswordEnc = v
}

// GetRegion returns the Region value.
func (m SiteConnector) GetRegion() *string {
	return m.Region
}

// SetRegion sets the Region value and returns the updated SiteConnector.
func (m *SiteConnector) SetRegion(v *string) {
	m.Region = v
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

// GetEnabled returns the Enabled value.
func (m SiteConnector) GetEnabled() bool {
	return m.Enabled
}

// SetEnabled sets the Enabled value and returns the updated SiteConnector.
func (m *SiteConnector) SetEnabled(v bool) {
	m.Enabled = v
}

// GetFirstSuccessfulSyncTime returns the FirstSuccessfulSyncTime value.
func (m SiteConnector) GetFirstSuccessfulSyncTime() *int64 {
	return m.FirstSuccessfulSyncTime
}

// SetFirstSuccessfulSyncTime sets the FirstSuccessfulSyncTime value and returns the updated SiteConnector.
func (m *SiteConnector) SetFirstSuccessfulSyncTime(v *int64) {
	m.FirstSuccessfulSyncTime = v
}

// GetBackfillStartedAt returns the BackfillStartedAt value.
func (m SiteConnector) GetBackfillStartedAt() *int64 {
	return m.BackfillStartedAt
}

// SetBackfillStartedAt sets the BackfillStartedAt value and returns the updated SiteConnector.
func (m *SiteConnector) SetBackfillStartedAt(v *int64) {
	m.BackfillStartedAt = v
}

// GetLastBackfillFrom returns the LastBackfillFrom value.
func (m SiteConnector) GetLastBackfillFrom() *int64 {
	return m.LastBackfillFrom
}

// SetLastBackfillFrom sets the LastBackfillFrom value and returns the updated SiteConnector.
func (m *SiteConnector) SetLastBackfillFrom(v *int64) {
	m.LastBackfillFrom = v
}

// GetLastBackfillTo returns the LastBackfillTo value.
func (m SiteConnector) GetLastBackfillTo() *int64 {
	return m.LastBackfillTo
}

// SetLastBackfillTo sets the LastBackfillTo value and returns the updated SiteConnector.
func (m *SiteConnector) SetLastBackfillTo(v *int64) {
	m.LastBackfillTo = v
}

// GetLastBackfillFinishedAt returns the LastBackfillFinishedAt value.
func (m SiteConnector) GetLastBackfillFinishedAt() *int64 {
	return m.LastBackfillFinishedAt
}

// SetLastBackfillFinishedAt sets the LastBackfillFinishedAt value and returns the updated SiteConnector.
func (m *SiteConnector) SetLastBackfillFinishedAt(v *int64) {
	m.LastBackfillFinishedAt = v
}

// GetLastBackfillDuration returns the LastBackfillDuration value.
func (m SiteConnector) GetLastBackfillDuration() *int64 {
	return m.LastBackfillDuration
}

// SetLastBackfillDuration sets the LastBackfillDuration value and returns the updated SiteConnector.
func (m *SiteConnector) SetLastBackfillDuration(v *int64) {
	m.LastBackfillDuration = v
}

// GetLastBackfillRecords returns the LastBackfillRecords value.
func (m SiteConnector) GetLastBackfillRecords() *int64 {
	return m.LastBackfillRecords
}

// SetLastBackfillRecords sets the LastBackfillRecords value and returns the updated SiteConnector.
func (m *SiteConnector) SetLastBackfillRecords(v *int64) {
	m.LastBackfillRecords = v
}

// GetLastBackfillError returns the LastBackfillError value.
func (m SiteConnector) GetLastBackfillError() *string {
	return m.LastBackfillError
}

// SetLastBackfillError sets the LastBackfillError value and returns the updated SiteConnector.
func (m *SiteConnector) SetLastBackfillError(v *string) {
	m.LastBackfillError = v
}

// GetTotalBackfills returns the TotalBackfills value.
func (m SiteConnector) GetTotalBackfills() int64 {
	return m.TotalBackfills
}

// SetTotalBackfills sets the TotalBackfills value and returns the updated SiteConnector.
func (m *SiteConnector) SetTotalBackfills(v int64) {
	m.TotalBackfills = v
}

// GetSyncIntervalSecondsReason returns the SyncIntervalSecondsReason value.
func (m SiteConnector) GetSyncIntervalSecondsReason() string {
	return m.SyncIntervalSecondsReason
}

// SetSyncIntervalSecondsReason sets the SyncIntervalSecondsReason value and returns the updated SiteConnector.
func (m *SiteConnector) SetSyncIntervalSecondsReason(v string) {
	m.SyncIntervalSecondsReason = v
}

// GetSyncIntervalSeconds returns the SyncIntervalSeconds value.
func (m SiteConnector) GetSyncIntervalSeconds() int64 {
	return m.SyncIntervalSeconds
}

// SetSyncIntervalSeconds sets the SyncIntervalSeconds value and returns the updated SiteConnector.
func (m *SiteConnector) SetSyncIntervalSeconds(v int64) {
	m.SyncIntervalSeconds = v
}

// GetLastAttemptedSync returns the LastAttemptedSync value.
func (m SiteConnector) GetLastAttemptedSync() *int64 {
	return m.LastAttemptedSync
}

// SetLastAttemptedSync sets the LastAttemptedSync value and returns the updated SiteConnector.
func (m *SiteConnector) SetLastAttemptedSync(v *int64) {
	m.LastAttemptedSync = v
}

// GetLastSuccessfulSync returns the LastSuccessfulSync value.
func (m SiteConnector) GetLastSuccessfulSync() *int64 {
	return m.LastSuccessfulSync
}

// SetLastSuccessfulSync sets the LastSuccessfulSync value and returns the updated SiteConnector.
func (m *SiteConnector) SetLastSuccessfulSync(v *int64) {
	m.LastSuccessfulSync = v
}

// GetLastFailedSync returns the LastFailedSync value.
func (m SiteConnector) GetLastFailedSync() *int64 {
	return m.LastFailedSync
}

// SetLastFailedSync sets the LastFailedSync value and returns the updated SiteConnector.
func (m *SiteConnector) SetLastFailedSync(v *int64) {
	m.LastFailedSync = v
}

// GetTotalSyncs returns the TotalSyncs value.
func (m SiteConnector) GetTotalSyncs() *int64 {
	return m.TotalSyncs
}

// SetTotalSyncs sets the TotalSyncs value and returns the updated SiteConnector.
func (m *SiteConnector) SetTotalSyncs(v *int64) {
	m.TotalSyncs = v
}

// GetTotalSuccessfulSyncs returns the TotalSuccessfulSyncs value.
func (m SiteConnector) GetTotalSuccessfulSyncs() *int64 {
	return m.TotalSuccessfulSyncs
}

// SetTotalSuccessfulSyncs sets the TotalSuccessfulSyncs value and returns the updated SiteConnector.
func (m *SiteConnector) SetTotalSuccessfulSyncs(v *int64) {
	m.TotalSuccessfulSyncs = v
}

// GetTotalFailedSyncs returns the TotalFailedSyncs value.
func (m SiteConnector) GetTotalFailedSyncs() *int64 {
	return m.TotalFailedSyncs
}

// SetTotalFailedSyncs sets the TotalFailedSyncs value and returns the updated SiteConnector.
func (m *SiteConnector) SetTotalFailedSyncs(v *int64) {
	m.TotalFailedSyncs = v
}

// GetTotalRecords returns the TotalRecords value.
func (m SiteConnector) GetTotalRecords() int64 {
	return m.TotalRecords
}

// SetTotalRecords sets the TotalRecords value and returns the updated SiteConnector.
func (m *SiteConnector) SetTotalRecords(v int64) {
	m.TotalRecords = v
}

// GetLastSyncRecords returns the LastSyncRecords value.
func (m SiteConnector) GetLastSyncRecords() *int64 {
	return m.LastSyncRecords
}

// SetLastSyncRecords sets the LastSyncRecords value and returns the updated SiteConnector.
func (m *SiteConnector) SetLastSyncRecords(v *int64) {
	m.LastSyncRecords = v
}

// GetConsecutiveSuccessfulSyncs returns the ConsecutiveSuccessfulSyncs value.
func (m SiteConnector) GetConsecutiveSuccessfulSyncs() *int64 {
	return m.ConsecutiveSuccessfulSyncs
}

// SetConsecutiveSuccessfulSyncs sets the ConsecutiveSuccessfulSyncs value and returns the updated SiteConnector.
func (m *SiteConnector) SetConsecutiveSuccessfulSyncs(v *int64) {
	m.ConsecutiveSuccessfulSyncs = v
}

// GetConsecutiveFailedSyncs returns the ConsecutiveFailedSyncs value.
func (m SiteConnector) GetConsecutiveFailedSyncs() *int64 {
	return m.ConsecutiveFailedSyncs
}

// SetConsecutiveFailedSyncs sets the ConsecutiveFailedSyncs value and returns the updated SiteConnector.
func (m *SiteConnector) SetConsecutiveFailedSyncs(v *int64) {
	m.ConsecutiveFailedSyncs = v
}

// GetPreviousSyncErrors returns the PreviousSyncErrors value.
func (m SiteConnector) GetPreviousSyncErrors() errors.ErrorList {
	return m.PreviousSyncErrors
}

// SetPreviousSyncErrors sets the PreviousSyncErrors value and returns the updated SiteConnector.
func (m *SiteConnector) SetPreviousSyncErrors(v errors.ErrorList) {
	m.PreviousSyncErrors = v
}

// GetTotalEma returns the TotalEma value.
func (m SiteConnector) GetTotalEma() *float64 {
	return m.TotalEma
}

// SetTotalEma sets the TotalEma value and returns the updated SiteConnector.
func (m *SiteConnector) SetTotalEma(v *float64) {
	m.TotalEma = v
}

// GetSuccessEma returns the SuccessEma value.
func (m SiteConnector) GetSuccessEma() *float64 {
	return m.SuccessEma
}

// SetSuccessEma sets the SuccessEma value and returns the updated SiteConnector.
func (m *SiteConnector) SetSuccessEma(v *float64) {
	m.SuccessEma = v
}

// GetFailureEma returns the FailureEma value.
func (m SiteConnector) GetFailureEma() *float64 {
	return m.FailureEma
}

// SetFailureEma sets the FailureEma value and returns the updated SiteConnector.
func (m *SiteConnector) SetFailureEma(v *float64) {
	m.FailureEma = v
}

// GetLastSyncDuration returns the LastSyncDuration value.
func (m SiteConnector) GetLastSyncDuration() *int64 {
	return m.LastSyncDuration
}

// SetLastSyncDuration sets the LastSyncDuration value and returns the updated SiteConnector.
func (m *SiteConnector) SetLastSyncDuration(v *int64) {
	m.LastSyncDuration = v
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

// GetClient returns the Client value.
func (m SiteIdentity) GetClient() *string {
	return m.Client
}

// SetClient sets the Client value and returns the updated SiteIdentity.
func (m *SiteIdentity) SetClient(v *string) {
	m.Client = v
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

// GetEnabled returns the Enabled value.
func (m SiteIdentity) GetEnabled() bool {
	return m.Enabled
}

// SetEnabled sets the Enabled value and returns the updated SiteIdentity.
func (m *SiteIdentity) SetEnabled(v bool) {
	m.Enabled = v
}

// GetCredentialsChangedAt returns the CredentialsChangedAt value.
func (m SiteIdentity) GetCredentialsChangedAt() int64 {
	return m.CredentialsChangedAt
}

// SetCredentialsChangedAt sets the CredentialsChangedAt value and returns the updated SiteIdentity.
func (m *SiteIdentity) SetCredentialsChangedAt(v int64) {
	m.CredentialsChangedAt = v
}

// GetNightscoutTokenDigest returns the NightscoutTokenDigest value.
func (m SiteIdentity) GetNightscoutTokenDigest() *string {
	return m.NightscoutTokenDigest
}

// SetNightscoutTokenDigest sets the NightscoutTokenDigest value and returns the updated SiteIdentity.
func (m *SiteIdentity) SetNightscoutTokenDigest(v *string) {
	m.NightscoutTokenDigest = v
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

// SetName sets the Name value and returns the updated GetControlIdentityParams.
func (m *GetControlIdentityParams) SetName(v string) {
	m.Name = v
}

// GetID returns the ID value.
func (m GetControlIdentityByIDParams) GetID() int64 {
	return m.ID
}

// SetID sets the ID value and returns the updated GetControlIdentityByIDParams.
func (m *GetControlIdentityByIDParams) SetID(v int64) {
	m.ID = v
}

// GetPasswordHash returns the PasswordHash value.
func (m UpdateControlIdentityPasswordParams) GetPasswordHash() string {
	return m.PasswordHash
}

// SetPasswordHash sets the PasswordHash value and returns the updated UpdateControlIdentityPasswordParams.
func (m *UpdateControlIdentityPasswordParams) SetPasswordHash(v string) {
	m.PasswordHash = v
}

// GetUpdatedByIdentityID returns the UpdatedByIdentityID value.
func (m UpdateControlIdentityPasswordParams) GetUpdatedByIdentityID() *int64 {
	return m.UpdatedByIdentityID
}

// SetUpdatedByIdentityID sets the UpdatedByIdentityID value and returns the updated UpdateControlIdentityPasswordParams.
func (m *UpdateControlIdentityPasswordParams) SetUpdatedByIdentityID(v *int64) {
	m.UpdatedByIdentityID = v
}

// GetUpdatedByIdentityName returns the UpdatedByIdentityName value.
func (m UpdateControlIdentityPasswordParams) GetUpdatedByIdentityName() *string {
	return m.UpdatedByIdentityName
}

// SetUpdatedByIdentityName sets the UpdatedByIdentityName value and returns the updated UpdateControlIdentityPasswordParams.
func (m *UpdateControlIdentityPasswordParams) SetUpdatedByIdentityName(v *string) {
	m.UpdatedByIdentityName = v
}

// GetID returns the ID value.
func (m UpdateControlIdentityPasswordParams) GetID() int64 {
	return m.ID
}

// SetID sets the ID value and returns the updated UpdateControlIdentityPasswordParams.
func (m *UpdateControlIdentityPasswordParams) SetID(v int64) {
	m.ID = v
}

// GetRole returns the Role value.
func (m UpdateControlIdentityRoleParams) GetRole() consts.ControlRole {
	return m.Role
}

// SetRole sets the Role value and returns the updated UpdateControlIdentityRoleParams.
func (m *UpdateControlIdentityRoleParams) SetRole(v consts.ControlRole) {
	m.Role = v
}

// GetUpdatedByIdentityID returns the UpdatedByIdentityID value.
func (m UpdateControlIdentityRoleParams) GetUpdatedByIdentityID() *int64 {
	return m.UpdatedByIdentityID
}

// SetUpdatedByIdentityID sets the UpdatedByIdentityID value and returns the updated UpdateControlIdentityRoleParams.
func (m *UpdateControlIdentityRoleParams) SetUpdatedByIdentityID(v *int64) {
	m.UpdatedByIdentityID = v
}

// GetUpdatedByIdentityName returns the UpdatedByIdentityName value.
func (m UpdateControlIdentityRoleParams) GetUpdatedByIdentityName() *string {
	return m.UpdatedByIdentityName
}

// SetUpdatedByIdentityName sets the UpdatedByIdentityName value and returns the updated UpdateControlIdentityRoleParams.
func (m *UpdateControlIdentityRoleParams) SetUpdatedByIdentityName(v *string) {
	m.UpdatedByIdentityName = v
}

// GetID returns the ID value.
func (m UpdateControlIdentityRoleParams) GetID() int64 {
	return m.ID
}

// SetID sets the ID value and returns the updated UpdateControlIdentityRoleParams.
func (m *UpdateControlIdentityRoleParams) SetID(v int64) {
	m.ID = v
}

// GetTotpSecretEnc returns the TotpSecretEnc value.
func (m UpdateControlIdentityTOTPParams) GetTotpSecretEnc() string {
	return m.TotpSecretEnc
}

// SetTotpSecretEnc sets the TotpSecretEnc value and returns the updated UpdateControlIdentityTOTPParams.
func (m *UpdateControlIdentityTOTPParams) SetTotpSecretEnc(v string) {
	m.TotpSecretEnc = v
}

// GetUpdatedByIdentityID returns the UpdatedByIdentityID value.
func (m UpdateControlIdentityTOTPParams) GetUpdatedByIdentityID() *int64 {
	return m.UpdatedByIdentityID
}

// SetUpdatedByIdentityID sets the UpdatedByIdentityID value and returns the updated UpdateControlIdentityTOTPParams.
func (m *UpdateControlIdentityTOTPParams) SetUpdatedByIdentityID(v *int64) {
	m.UpdatedByIdentityID = v
}

// GetUpdatedByIdentityName returns the UpdatedByIdentityName value.
func (m UpdateControlIdentityTOTPParams) GetUpdatedByIdentityName() *string {
	return m.UpdatedByIdentityName
}

// SetUpdatedByIdentityName sets the UpdatedByIdentityName value and returns the updated UpdateControlIdentityTOTPParams.
func (m *UpdateControlIdentityTOTPParams) SetUpdatedByIdentityName(v *string) {
	m.UpdatedByIdentityName = v
}

// GetID returns the ID value.
func (m UpdateControlIdentityTOTPParams) GetID() int64 {
	return m.ID
}

// SetID sets the ID value and returns the updated UpdateControlIdentityTOTPParams.
func (m *UpdateControlIdentityTOTPParams) SetID(v int64) {
	m.ID = v
}

// GetControlIdentityID returns the ControlIdentityID value.
func (m CreateControlPasskeyParams) GetControlIdentityID() int64 {
	return m.ControlIdentityID
}

// SetControlIdentityID sets the ControlIdentityID value and returns the updated CreateControlPasskeyParams.
func (m *CreateControlPasskeyParams) SetControlIdentityID(v int64) {
	m.ControlIdentityID = v
}

// GetCredentialID returns the CredentialID value.
func (m CreateControlPasskeyParams) GetCredentialID() string {
	return m.CredentialID
}

// SetCredentialID sets the CredentialID value and returns the updated CreateControlPasskeyParams.
func (m *CreateControlPasskeyParams) SetCredentialID(v string) {
	m.CredentialID = v
}

// GetCredentialJson returns the CredentialJson value.
func (m CreateControlPasskeyParams) GetCredentialJson() string {
	return m.CredentialJson
}

// SetCredentialJson sets the CredentialJson value and returns the updated CreateControlPasskeyParams.
func (m *CreateControlPasskeyParams) SetCredentialJson(v string) {
	m.CredentialJson = v
}

// GetSiteID returns the SiteID value.
func (m CreateSiteIdentityPasskeyParams) GetSiteID() string {
	return m.SiteID
}

// SetSiteID sets the SiteID value and returns the updated CreateSiteIdentityPasskeyParams.
func (m *CreateSiteIdentityPasskeyParams) SetSiteID(v string) {
	m.SiteID = v
}

// GetIdentityID returns the IdentityID value.
func (m CreateSiteIdentityPasskeyParams) GetIdentityID() int64 {
	return m.IdentityID
}

// SetIdentityID sets the IdentityID value and returns the updated CreateSiteIdentityPasskeyParams.
func (m *CreateSiteIdentityPasskeyParams) SetIdentityID(v int64) {
	m.IdentityID = v
}

// GetCredentialID returns the CredentialID value.
func (m CreateSiteIdentityPasskeyParams) GetCredentialID() string {
	return m.CredentialID
}

// SetCredentialID sets the CredentialID value and returns the updated CreateSiteIdentityPasskeyParams.
func (m *CreateSiteIdentityPasskeyParams) SetCredentialID(v string) {
	m.CredentialID = v
}

// GetCredentialJson returns the CredentialJson value.
func (m CreateSiteIdentityPasskeyParams) GetCredentialJson() string {
	return m.CredentialJson
}

// SetCredentialJson sets the CredentialJson value and returns the updated CreateSiteIdentityPasskeyParams.
func (m *CreateSiteIdentityPasskeyParams) SetCredentialJson(v string) {
	m.CredentialJson = v
}

// GetControlIdentityID returns the ControlIdentityID value.
func (m DeleteControlIdentityPasskeysParams) GetControlIdentityID() int64 {
	return m.ControlIdentityID
}

// SetControlIdentityID sets the ControlIdentityID value and returns the updated DeleteControlIdentityPasskeysParams.
func (m *DeleteControlIdentityPasskeysParams) SetControlIdentityID(v int64) {
	m.ControlIdentityID = v
}

// GetControlIdentityID returns the ControlIdentityID value.
func (m DeleteControlPasskeyParams) GetControlIdentityID() int64 {
	return m.ControlIdentityID
}

// SetControlIdentityID sets the ControlIdentityID value and returns the updated DeleteControlPasskeyParams.
func (m *DeleteControlPasskeyParams) SetControlIdentityID(v int64) {
	m.ControlIdentityID = v
}

// GetID returns the ID value.
func (m DeleteControlPasskeyParams) GetID() int64 {
	return m.ID
}

// SetID sets the ID value and returns the updated DeleteControlPasskeyParams.
func (m *DeleteControlPasskeyParams) SetID(v int64) {
	m.ID = v
}

// GetSiteID returns the SiteID value.
func (m DeleteSiteAdminPasskeysParams) GetSiteID() string {
	return m.SiteID
}

// SetSiteID sets the SiteID value and returns the updated DeleteSiteAdminPasskeysParams.
func (m *DeleteSiteAdminPasskeysParams) SetSiteID(v string) {
	m.SiteID = v
}

// GetName returns the Name value.
func (m DeleteSiteAdminPasskeysParams) GetName() string {
	return m.Name
}

// SetName sets the Name value and returns the updated DeleteSiteAdminPasskeysParams.
func (m *DeleteSiteAdminPasskeysParams) SetName(v string) {
	m.Name = v
}

// GetSiteID returns the SiteID value.
func (m DeleteSiteIdentityPasskeyParams) GetSiteID() string {
	return m.SiteID
}

// SetSiteID sets the SiteID value and returns the updated DeleteSiteIdentityPasskeyParams.
func (m *DeleteSiteIdentityPasskeyParams) SetSiteID(v string) {
	m.SiteID = v
}

// GetIdentityID returns the IdentityID value.
func (m DeleteSiteIdentityPasskeyParams) GetIdentityID() int64 {
	return m.IdentityID
}

// SetIdentityID sets the IdentityID value and returns the updated DeleteSiteIdentityPasskeyParams.
func (m *DeleteSiteIdentityPasskeyParams) SetIdentityID(v int64) {
	m.IdentityID = v
}

// GetID returns the ID value.
func (m DeleteSiteIdentityPasskeyParams) GetID() int64 {
	return m.ID
}

// SetID sets the ID value and returns the updated DeleteSiteIdentityPasskeyParams.
func (m *DeleteSiteIdentityPasskeyParams) SetID(v int64) {
	m.ID = v
}

// GetSiteID returns the SiteID value.
func (m DeleteSiteIdentityPasskeysParams) GetSiteID() string {
	return m.SiteID
}

// SetSiteID sets the SiteID value and returns the updated DeleteSiteIdentityPasskeysParams.
func (m *DeleteSiteIdentityPasskeysParams) SetSiteID(v string) {
	m.SiteID = v
}

// GetIdentityID returns the IdentityID value.
func (m DeleteSiteIdentityPasskeysParams) GetIdentityID() int64 {
	return m.IdentityID
}

// SetIdentityID sets the IdentityID value and returns the updated DeleteSiteIdentityPasskeysParams.
func (m *DeleteSiteIdentityPasskeysParams) SetIdentityID(v int64) {
	m.IdentityID = v
}

// GetCredentialID returns the CredentialID value.
func (m GetControlPasskeyParams) GetCredentialID() string {
	return m.CredentialID
}

// SetCredentialID sets the CredentialID value and returns the updated GetControlPasskeyParams.
func (m *GetControlPasskeyParams) SetCredentialID(v string) {
	m.CredentialID = v
}

// GetSiteID returns the SiteID value.
func (m GetSiteIdentityPasskeyParams) GetSiteID() string {
	return m.SiteID
}

// SetSiteID sets the SiteID value and returns the updated GetSiteIdentityPasskeyParams.
func (m *GetSiteIdentityPasskeyParams) SetSiteID(v string) {
	m.SiteID = v
}

// GetCredentialID returns the CredentialID value.
func (m GetSiteIdentityPasskeyParams) GetCredentialID() string {
	return m.CredentialID
}

// SetCredentialID sets the CredentialID value and returns the updated GetSiteIdentityPasskeyParams.
func (m *GetSiteIdentityPasskeyParams) SetCredentialID(v string) {
	m.CredentialID = v
}

// GetControlIdentityID returns the ControlIdentityID value.
func (m ListControlPasskeysParams) GetControlIdentityID() int64 {
	return m.ControlIdentityID
}

// SetControlIdentityID sets the ControlIdentityID value and returns the updated ListControlPasskeysParams.
func (m *ListControlPasskeysParams) SetControlIdentityID(v int64) {
	m.ControlIdentityID = v
}

// GetSiteID returns the SiteID value.
func (m ListSiteIdentityPasskeysParams) GetSiteID() string {
	return m.SiteID
}

// SetSiteID sets the SiteID value and returns the updated ListSiteIdentityPasskeysParams.
func (m *ListSiteIdentityPasskeysParams) SetSiteID(v string) {
	m.SiteID = v
}

// GetIdentityID returns the IdentityID value.
func (m ListSiteIdentityPasskeysParams) GetIdentityID() int64 {
	return m.IdentityID
}

// SetIdentityID sets the IdentityID value and returns the updated ListSiteIdentityPasskeysParams.
func (m *ListSiteIdentityPasskeysParams) SetIdentityID(v int64) {
	m.IdentityID = v
}

// GetCredentialJson returns the CredentialJson value.
func (m UpdateControlPasskeyUsedParams) GetCredentialJson() string {
	return m.CredentialJson
}

// SetCredentialJson sets the CredentialJson value and returns the updated UpdateControlPasskeyUsedParams.
func (m *UpdateControlPasskeyUsedParams) SetCredentialJson(v string) {
	m.CredentialJson = v
}

// GetCredentialID returns the CredentialID value.
func (m UpdateControlPasskeyUsedParams) GetCredentialID() string {
	return m.CredentialID
}

// SetCredentialID sets the CredentialID value and returns the updated UpdateControlPasskeyUsedParams.
func (m *UpdateControlPasskeyUsedParams) SetCredentialID(v string) {
	m.CredentialID = v
}

// GetCredentialJson returns the CredentialJson value.
func (m UpdateSiteIdentityPasskeyUsedParams) GetCredentialJson() string {
	return m.CredentialJson
}

// SetCredentialJson sets the CredentialJson value and returns the updated UpdateSiteIdentityPasskeyUsedParams.
func (m *UpdateSiteIdentityPasskeyUsedParams) SetCredentialJson(v string) {
	m.CredentialJson = v
}

// GetSiteID returns the SiteID value.
func (m UpdateSiteIdentityPasskeyUsedParams) GetSiteID() string {
	return m.SiteID
}

// SetSiteID sets the SiteID value and returns the updated UpdateSiteIdentityPasskeyUsedParams.
func (m *UpdateSiteIdentityPasskeyUsedParams) SetSiteID(v string) {
	m.SiteID = v
}

// GetCredentialID returns the CredentialID value.
func (m UpdateSiteIdentityPasskeyUsedParams) GetCredentialID() string {
	return m.CredentialID
}

// SetCredentialID sets the CredentialID value and returns the updated UpdateSiteIdentityPasskeyUsedParams.
func (m *UpdateSiteIdentityPasskeyUsedParams) SetCredentialID(v string) {
	m.CredentialID = v
}

// GetSiteID returns the SiteID value.
func (m CountPushCredentialsParams) GetSiteID() *string {
	return m.SiteID
}

// SetSiteID sets the SiteID value and returns the updated CountPushCredentialsParams.
func (m *CountPushCredentialsParams) SetSiteID(v *string) {
	m.SiteID = v
}

// GetSiteID returns the SiteID value.
func (m CountPushRegistrationsParams) GetSiteID() string {
	return m.SiteID
}

// SetSiteID sets the SiteID value and returns the updated CountPushRegistrationsParams.
func (m *CountPushRegistrationsParams) SetSiteID(v string) {
	m.SiteID = v
}

// GetIdentityID returns the IdentityID value.
func (m CountPushRegistrationsParams) GetIdentityID() int64 {
	return m.IdentityID
}

// SetIdentityID sets the IdentityID value and returns the updated CountPushRegistrationsParams.
func (m *CountPushRegistrationsParams) SetIdentityID(v int64) {
	m.IdentityID = v
}

// GetTokenHash returns the TokenHash value.
func (m CountPushRegistrationsParams) GetTokenHash() string {
	return m.TokenHash
}

// SetTokenHash sets the TokenHash value and returns the updated CountPushRegistrationsParams.
func (m *CountPushRegistrationsParams) SetTokenHash(v string) {
	m.TokenHash = v
}

// GetSiteID returns the SiteID value.
func (m CreatePushCredentialParams) GetSiteID() *string {
	return m.SiteID
}

// SetSiteID sets the SiteID value and returns the updated CreatePushCredentialParams.
func (m *CreatePushCredentialParams) SetSiteID(v *string) {
	m.SiteID = v
}

// GetName returns the Name value.
func (m CreatePushCredentialParams) GetName() string {
	return m.Name
}

// SetName sets the Name value and returns the updated CreatePushCredentialParams.
func (m *CreatePushCredentialParams) SetName(v string) {
	m.Name = v
}

// GetProvider returns the Provider value.
func (m CreatePushCredentialParams) GetProvider() consts.PushProvider {
	return m.Provider
}

// SetProvider sets the Provider value and returns the updated CreatePushCredentialParams.
func (m *CreatePushCredentialParams) SetProvider(v consts.PushProvider) {
	m.Provider = v
}

// GetTeamOrProjectID returns the TeamOrProjectID value.
func (m CreatePushCredentialParams) GetTeamOrProjectID() string {
	return m.TeamOrProjectID
}

// SetTeamOrProjectID sets the TeamOrProjectID value and returns the updated CreatePushCredentialParams.
func (m *CreatePushCredentialParams) SetTeamOrProjectID(v string) {
	m.TeamOrProjectID = v
}

// GetKeyID returns the KeyID value.
func (m CreatePushCredentialParams) GetKeyID() *string {
	return m.KeyID
}

// SetKeyID sets the KeyID value and returns the updated CreatePushCredentialParams.
func (m *CreatePushCredentialParams) SetKeyID(v *string) {
	m.KeyID = v
}

// GetSecretEnc returns the SecretEnc value.
func (m CreatePushCredentialParams) GetSecretEnc() *string {
	return m.SecretEnc
}

// SetSecretEnc sets the SecretEnc value and returns the updated CreatePushCredentialParams.
func (m *CreatePushCredentialParams) SetSecretEnc(v *string) {
	m.SecretEnc = v
}

// GetAppsJson returns the AppsJson value.
func (m CreatePushCredentialParams) GetAppsJson() string {
	return m.AppsJson
}

// SetAppsJson sets the AppsJson value and returns the updated CreatePushCredentialParams.
func (m *CreatePushCredentialParams) SetAppsJson(v string) {
	m.AppsJson = v
}

// GetBundleIdsJson returns the BundleIdsJson value.
func (m CreatePushCredentialParams) GetBundleIdsJson() *string {
	return m.BundleIdsJson
}

// SetBundleIdsJson sets the BundleIdsJson value and returns the updated CreatePushCredentialParams.
func (m *CreatePushCredentialParams) SetBundleIdsJson(v *string) {
	m.BundleIdsJson = v
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

// GetSiteID returns the SiteID value.
func (m DeletePushCredentialParams) GetSiteID() *string {
	return m.SiteID
}

// SetSiteID sets the SiteID value and returns the updated DeletePushCredentialParams.
func (m *DeletePushCredentialParams) SetSiteID(v *string) {
	m.SiteID = v
}

// GetID returns the ID value.
func (m DeletePushRegistrationParams) GetID() int64 {
	return m.ID
}

// SetID sets the ID value and returns the updated DeletePushRegistrationParams.
func (m *DeletePushRegistrationParams) SetID(v int64) {
	m.ID = v
}

// GetSiteID returns the SiteID value.
func (m DeletePushRegistrationParams) GetSiteID() string {
	return m.SiteID
}

// SetSiteID sets the SiteID value and returns the updated DeletePushRegistrationParams.
func (m *DeletePushRegistrationParams) SetSiteID(v string) {
	m.SiteID = v
}

// GetIdentityID returns the IdentityID value.
func (m DeletePushRegistrationParams) GetIdentityID() *int64 {
	return m.IdentityID
}

// SetIdentityID sets the IdentityID value and returns the updated DeletePushRegistrationParams.
func (m *DeletePushRegistrationParams) SetIdentityID(v *int64) {
	m.IdentityID = v
}

// GetSiteID returns the SiteID value.
func (m DeletePushRegistrationByTokenHashParams) GetSiteID() string {
	return m.SiteID
}

// SetSiteID sets the SiteID value and returns the updated DeletePushRegistrationByTokenHashParams.
func (m *DeletePushRegistrationByTokenHashParams) SetSiteID(v string) {
	m.SiteID = v
}

// GetTokenHash returns the TokenHash value.
func (m DeletePushRegistrationByTokenHashParams) GetTokenHash() string {
	return m.TokenHash
}

// SetTokenHash sets the TokenHash value and returns the updated DeletePushRegistrationByTokenHashParams.
func (m *DeletePushRegistrationByTokenHashParams) SetTokenHash(v string) {
	m.TokenHash = v
}

// GetSiteID returns the SiteID value.
func (m DeleteSitePushCredentialsParams) GetSiteID() *string {
	return m.SiteID
}

// SetSiteID sets the SiteID value and returns the updated DeleteSitePushCredentialsParams.
func (m *DeleteSitePushCredentialsParams) SetSiteID(v *string) {
	m.SiteID = v
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

// GetSiteID returns the SiteID value.
func (m GetPushCredentialParams) GetSiteID() *string {
	return m.SiteID
}

// SetSiteID sets the SiteID value and returns the updated GetPushCredentialParams.
func (m *GetPushCredentialParams) SetSiteID(v *string) {
	m.SiteID = v
}

// GetID returns the ID value.
func (m GetPushRegistrationParams) GetID() int64 {
	return m.ID
}

// SetID sets the ID value and returns the updated GetPushRegistrationParams.
func (m *GetPushRegistrationParams) SetID(v int64) {
	m.ID = v
}

// GetSiteID returns the SiteID value.
func (m GetPushRegistrationParams) GetSiteID() string {
	return m.SiteID
}

// SetSiteID sets the SiteID value and returns the updated GetPushRegistrationParams.
func (m *GetPushRegistrationParams) SetSiteID(v string) {
	m.SiteID = v
}

// GetIdentityID returns the IdentityID value.
func (m GetPushRegistrationParams) GetIdentityID() *int64 {
	return m.IdentityID
}

// SetIdentityID sets the IdentityID value and returns the updated GetPushRegistrationParams.
func (m *GetPushRegistrationParams) SetIdentityID(v *int64) {
	m.IdentityID = v
}

// GetSiteID returns the SiteID value.
func (m PruneDeletedPushCredentialsParams) GetSiteID() *string {
	return m.SiteID
}

// SetSiteID sets the SiteID value and returns the updated PruneDeletedPushCredentialsParams.
func (m *PruneDeletedPushCredentialsParams) SetSiteID(v *string) {
	m.SiteID = v
}

// GetKeep returns the Keep value.
func (m PruneDeletedPushCredentialsParams) GetKeep() int64 {
	return m.Keep
}

// SetKeep sets the Keep value and returns the updated PruneDeletedPushCredentialsParams.
func (m *PruneDeletedPushCredentialsParams) SetKeep(v int64) {
	m.Keep = v
}

// GetSiteID returns the SiteID value.
func (m PushCredentialForParams) GetSiteID() *string {
	return m.SiteID
}

// SetSiteID sets the SiteID value and returns the updated PushCredentialForParams.
func (m *PushCredentialForParams) SetSiteID(v *string) {
	m.SiteID = v
}

// GetProvider returns the Provider value.
func (m PushCredentialForParams) GetProvider() consts.PushProvider {
	return m.Provider
}

// SetProvider sets the Provider value and returns the updated PushCredentialForParams.
func (m *PushCredentialForParams) SetProvider(v consts.PushProvider) {
	m.Provider = v
}

// GetApp returns the App value.
func (m PushCredentialForParams) GetApp() string {
	return m.App
}

// SetApp sets the App value and returns the updated PushCredentialForParams.
func (m *PushCredentialForParams) SetApp(v string) {
	m.App = v
}

// GetBundleID returns the BundleID value.
func (m PushCredentialForParams) GetBundleID() string {
	return m.BundleID
}

// SetBundleID sets the BundleID value and returns the updated PushCredentialForParams.
func (m *PushCredentialForParams) SetBundleID(v string) {
	m.BundleID = v
}

// GetLastError returns the LastError value.
func (m RecordPushParams) GetLastError() *string {
	return m.LastError
}

// SetLastError sets the LastError value and returns the updated RecordPushParams.
func (m *RecordPushParams) SetLastError(v *string) {
	m.LastError = v
}

// GetID returns the ID value.
func (m RecordPushParams) GetID() int64 {
	return m.ID
}

// SetID sets the ID value and returns the updated RecordPushParams.
func (m *RecordPushParams) SetID(v int64) {
	m.ID = v
}

// GetSiteID returns the SiteID value.
func (m RecordPushParams) GetSiteID() string {
	return m.SiteID
}

// SetSiteID sets the SiteID value and returns the updated RecordPushParams.
func (m *RecordPushParams) SetSiteID(v string) {
	m.SiteID = v
}

// GetTokenHash returns the TokenHash value.
func (m RecordPushParams) GetTokenHash() string {
	return m.TokenHash
}

// SetTokenHash sets the TokenHash value and returns the updated RecordPushParams.
func (m *RecordPushParams) SetTokenHash(v string) {
	m.TokenHash = v
}

// GetSkip returns the Skip value.
func (m SearchAllPushCredentialsParams) GetSkip() int64 {
	return m.Skip
}

// SetSkip sets the Skip value and returns the updated SearchAllPushCredentialsParams.
func (m *SearchAllPushCredentialsParams) SetSkip(v int64) {
	m.Skip = v
}

// GetLimit returns the Limit value.
func (m SearchAllPushCredentialsParams) GetLimit() int64 {
	return m.Limit
}

// SetLimit sets the Limit value and returns the updated SearchAllPushCredentialsParams.
func (m *SearchAllPushCredentialsParams) SetLimit(v int64) {
	m.Limit = v
}

// GetSiteID returns the SiteID value.
func (m SearchAllPushRegistrationsParams) GetSiteID() *string {
	return m.SiteID
}

// SetSiteID sets the SiteID value and returns the updated SearchAllPushRegistrationsParams.
func (m *SearchAllPushRegistrationsParams) SetSiteID(v *string) {
	m.SiteID = v
}

// GetLimit returns the Limit value.
func (m SearchAllPushRegistrationsParams) GetLimit() int64 {
	return m.Limit
}

// SetLimit sets the Limit value and returns the updated SearchAllPushRegistrationsParams.
func (m *SearchAllPushRegistrationsParams) SetLimit(v int64) {
	m.Limit = v
}

// GetSiteID returns the SiteID value.
func (m SearchPushCredentialsParams) GetSiteID() *string {
	return m.SiteID
}

// SetSiteID sets the SiteID value and returns the updated SearchPushCredentialsParams.
func (m *SearchPushCredentialsParams) SetSiteID(v *string) {
	m.SiteID = v
}

// GetSkip returns the Skip value.
func (m SearchPushCredentialsParams) GetSkip() int64 {
	return m.Skip
}

// SetSkip sets the Skip value and returns the updated SearchPushCredentialsParams.
func (m *SearchPushCredentialsParams) SetSkip(v int64) {
	m.Skip = v
}

// GetLimit returns the Limit value.
func (m SearchPushCredentialsParams) GetLimit() int64 {
	return m.Limit
}

// SetLimit sets the Limit value and returns the updated SearchPushCredentialsParams.
func (m *SearchPushCredentialsParams) SetLimit(v int64) {
	m.Limit = v
}

// GetSiteID returns the SiteID value.
func (m SearchPushRegistrationsParams) GetSiteID() string {
	return m.SiteID
}

// SetSiteID sets the SiteID value and returns the updated SearchPushRegistrationsParams.
func (m *SearchPushRegistrationsParams) SetSiteID(v string) {
	m.SiteID = v
}

// GetIdentityID returns the IdentityID value.
func (m SearchPushRegistrationsParams) GetIdentityID() *int64 {
	return m.IdentityID
}

// SetIdentityID sets the IdentityID value and returns the updated SearchPushRegistrationsParams.
func (m *SearchPushRegistrationsParams) SetIdentityID(v *int64) {
	m.IdentityID = v
}

// GetSkip returns the Skip value.
func (m SearchPushRegistrationsParams) GetSkip() int64 {
	return m.Skip
}

// SetSkip sets the Skip value and returns the updated SearchPushRegistrationsParams.
func (m *SearchPushRegistrationsParams) SetSkip(v int64) {
	m.Skip = v
}

// GetLimit returns the Limit value.
func (m SearchPushRegistrationsParams) GetLimit() int64 {
	return m.Limit
}

// SetLimit sets the Limit value and returns the updated SearchPushRegistrationsParams.
func (m *SearchPushRegistrationsParams) SetLimit(v int64) {
	m.Limit = v
}

// GetSiteID returns the SiteID value.
func (m SearchPushRegistrationsWithIdentityParams) GetSiteID() string {
	return m.SiteID
}

// SetSiteID sets the SiteID value and returns the updated SearchPushRegistrationsWithIdentityParams.
func (m *SearchPushRegistrationsWithIdentityParams) SetSiteID(v string) {
	m.SiteID = v
}

// GetIdentityID returns the IdentityID value.
func (m SearchPushRegistrationsWithIdentityParams) GetIdentityID() *int64 {
	return m.IdentityID
}

// SetIdentityID sets the IdentityID value and returns the updated SearchPushRegistrationsWithIdentityParams.
func (m *SearchPushRegistrationsWithIdentityParams) SetIdentityID(v *int64) {
	m.IdentityID = v
}

// GetLimit returns the Limit value.
func (m SearchPushRegistrationsWithIdentityParams) GetLimit() int64 {
	return m.Limit
}

// SetLimit sets the Limit value and returns the updated SearchPushRegistrationsWithIdentityParams.
func (m *SearchPushRegistrationsWithIdentityParams) SetLimit(v int64) {
	m.Limit = v
}

// GetSiteID returns the SiteID value.
func (m UpsertPushRegistrationParams) GetSiteID() string {
	return m.SiteID
}

// SetSiteID sets the SiteID value and returns the updated UpsertPushRegistrationParams.
func (m *UpsertPushRegistrationParams) SetSiteID(v string) {
	m.SiteID = v
}

// GetIdentityID returns the IdentityID value.
func (m UpsertPushRegistrationParams) GetIdentityID() int64 {
	return m.IdentityID
}

// SetIdentityID sets the IdentityID value and returns the updated UpsertPushRegistrationParams.
func (m *UpsertPushRegistrationParams) SetIdentityID(v int64) {
	m.IdentityID = v
}

// GetApp returns the App value.
func (m UpsertPushRegistrationParams) GetApp() consts.PushApp {
	return m.App
}

// SetApp sets the App value and returns the updated UpsertPushRegistrationParams.
func (m *UpsertPushRegistrationParams) SetApp(v consts.PushApp) {
	m.App = v
}

// GetProvider returns the Provider value.
func (m UpsertPushRegistrationParams) GetProvider() consts.PushProvider {
	return m.Provider
}

// SetProvider sets the Provider value and returns the updated UpsertPushRegistrationParams.
func (m *UpsertPushRegistrationParams) SetProvider(v consts.PushProvider) {
	m.Provider = v
}

// GetBundleID returns the BundleID value.
func (m UpsertPushRegistrationParams) GetBundleID() string {
	return m.BundleID
}

// SetBundleID sets the BundleID value and returns the updated UpsertPushRegistrationParams.
func (m *UpsertPushRegistrationParams) SetBundleID(v string) {
	m.BundleID = v
}

// GetProduction returns the Production value.
func (m UpsertPushRegistrationParams) GetProduction() bool {
	return m.Production
}

// SetProduction sets the Production value and returns the updated UpsertPushRegistrationParams.
func (m *UpsertPushRegistrationParams) SetProduction(v bool) {
	m.Production = v
}

// GetTokenEnc returns the TokenEnc value.
func (m UpsertPushRegistrationParams) GetTokenEnc() string {
	return m.TokenEnc
}

// SetTokenEnc sets the TokenEnc value and returns the updated UpsertPushRegistrationParams.
func (m *UpsertPushRegistrationParams) SetTokenEnc(v string) {
	m.TokenEnc = v
}

// GetTokenHash returns the TokenHash value.
func (m UpsertPushRegistrationParams) GetTokenHash() string {
	return m.TokenHash
}

// SetTokenHash sets the TokenHash value and returns the updated UpsertPushRegistrationParams.
func (m *UpsertPushRegistrationParams) SetTokenHash(v string) {
	m.TokenHash = v
}

// GetPublicSiteCreation returns the PublicSiteCreation value.
func (m UpdatePublicSiteCreationParams) GetPublicSiteCreation() bool {
	return m.PublicSiteCreation
}

// SetPublicSiteCreation sets the PublicSiteCreation value and returns the updated UpdatePublicSiteCreationParams.
func (m *UpdatePublicSiteCreationParams) SetPublicSiteCreation(v bool) {
	m.PublicSiteCreation = v
}

// GetUpdatedByIdentityName returns the UpdatedByIdentityName value.
func (m UpdatePublicSiteCreationParams) GetUpdatedByIdentityName() *string {
	return m.UpdatedByIdentityName
}

// SetUpdatedByIdentityName sets the UpdatedByIdentityName value and returns the updated UpdatePublicSiteCreationParams.
func (m *UpdatePublicSiteCreationParams) SetUpdatedByIdentityName(v *string) {
	m.UpdatedByIdentityName = v
}

// GetRateLimitsJson returns the RateLimitsJson value.
func (m UpdateRateLimitsParams) GetRateLimitsJson() string {
	return m.RateLimitsJson
}

// SetRateLimitsJson sets the RateLimitsJson value and returns the updated UpdateRateLimitsParams.
func (m *UpdateRateLimitsParams) SetRateLimitsJson(v string) {
	m.RateLimitsJson = v
}

// GetUpdatedByIdentityName returns the UpdatedByIdentityName value.
func (m UpdateRateLimitsParams) GetUpdatedByIdentityName() *string {
	return m.UpdatedByIdentityName
}

// SetUpdatedByIdentityName sets the UpdatedByIdentityName value and returns the updated UpdateRateLimitsParams.
func (m *UpdateRateLimitsParams) SetUpdatedByIdentityName(v *string) {
	m.UpdatedByIdentityName = v
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

// GetSiteID returns the SiteID value.
func (m CreateConnectorParams) GetSiteID() string {
	return m.SiteID
}

// SetSiteID sets the SiteID value and returns the updated CreateConnectorParams.
func (m *CreateConnectorParams) SetSiteID(v string) {
	m.SiteID = v
}

// GetType returns the Type value.
func (m CreateConnectorParams) GetType() consts.ConnectorType {
	return m.Type
}

// SetType sets the Type value and returns the updated CreateConnectorParams.
func (m *CreateConnectorParams) SetType(v consts.ConnectorType) {
	m.Type = v
}

// GetUrlOrUsernameEnc returns the UrlOrUsernameEnc value.
func (m CreateConnectorParams) GetUrlOrUsernameEnc() *string {
	return m.UrlOrUsernameEnc
}

// SetUrlOrUsernameEnc sets the UrlOrUsernameEnc value and returns the updated CreateConnectorParams.
func (m *CreateConnectorParams) SetUrlOrUsernameEnc(v *string) {
	m.UrlOrUsernameEnc = v
}

// GetPasswordEnc returns the PasswordEnc value.
func (m CreateConnectorParams) GetPasswordEnc() *string {
	return m.PasswordEnc
}

// SetPasswordEnc sets the PasswordEnc value and returns the updated CreateConnectorParams.
func (m *CreateConnectorParams) SetPasswordEnc(v *string) {
	m.PasswordEnc = v
}

// GetRegion returns the Region value.
func (m CreateConnectorParams) GetRegion() *string {
	return m.Region
}

// SetRegion sets the Region value and returns the updated CreateConnectorParams.
func (m *CreateConnectorParams) SetRegion(v *string) {
	m.Region = v
}

// GetEnabled returns the Enabled value.
func (m CreateConnectorParams) GetEnabled() bool {
	return m.Enabled
}

// SetEnabled sets the Enabled value and returns the updated CreateConnectorParams.
func (m *CreateConnectorParams) SetEnabled(v bool) {
	m.Enabled = v
}

// GetType returns the Type value.
func (m DeleteConnectorParams) GetType() consts.ConnectorType {
	return m.Type
}

// SetType sets the Type value and returns the updated DeleteConnectorParams.
func (m *DeleteConnectorParams) SetType(v consts.ConnectorType) {
	m.Type = v
}

// GetSiteID returns the SiteID value.
func (m DeleteConnectorParams) GetSiteID() string {
	return m.SiteID
}

// SetSiteID sets the SiteID value and returns the updated DeleteConnectorParams.
func (m *DeleteConnectorParams) SetSiteID(v string) {
	m.SiteID = v
}

// GetDurationMs returns the DurationMs value.
func (m FinishConnectorBackfillParams) GetDurationMs() *int64 {
	return m.DurationMs
}

// SetDurationMs sets the DurationMs value and returns the updated FinishConnectorBackfillParams.
func (m *FinishConnectorBackfillParams) SetDurationMs(v *int64) {
	m.DurationMs = v
}

// GetRecords returns the Records value.
func (m FinishConnectorBackfillParams) GetRecords() *int64 {
	return m.Records
}

// SetRecords sets the Records value and returns the updated FinishConnectorBackfillParams.
func (m *FinishConnectorBackfillParams) SetRecords(v *int64) {
	m.Records = v
}

// GetError returns the Error value.
func (m FinishConnectorBackfillParams) GetError() *string {
	return m.Error
}

// SetError sets the Error value and returns the updated FinishConnectorBackfillParams.
func (m *FinishConnectorBackfillParams) SetError(v *string) {
	m.Error = v
}

// GetType returns the Type value.
func (m FinishConnectorBackfillParams) GetType() consts.ConnectorType {
	return m.Type
}

// SetType sets the Type value and returns the updated FinishConnectorBackfillParams.
func (m *FinishConnectorBackfillParams) SetType(v consts.ConnectorType) {
	m.Type = v
}

// GetSiteID returns the SiteID value.
func (m FinishConnectorBackfillParams) GetSiteID() string {
	return m.SiteID
}

// SetSiteID sets the SiteID value and returns the updated FinishConnectorBackfillParams.
func (m *FinishConnectorBackfillParams) SetSiteID(v string) {
	m.SiteID = v
}

// GetSiteID returns the SiteID value.
func (m GetAllConnectorsForSiteParams) GetSiteID() string {
	return m.SiteID
}

// SetSiteID sets the SiteID value and returns the updated GetAllConnectorsForSiteParams.
func (m *GetAllConnectorsForSiteParams) SetSiteID(v string) {
	m.SiteID = v
}

// GetType returns the Type value.
func (m GetConnectorParams) GetType() consts.ConnectorType {
	return m.Type
}

// SetType sets the Type value and returns the updated GetConnectorParams.
func (m *GetConnectorParams) SetType(v consts.ConnectorType) {
	m.Type = v
}

// GetSiteID returns the SiteID value.
func (m GetConnectorParams) GetSiteID() string {
	return m.SiteID
}

// SetSiteID sets the SiteID value and returns the updated GetConnectorParams.
func (m *GetConnectorParams) SetSiteID(v string) {
	m.SiteID = v
}

// GetSkip returns the Skip value.
func (m SearchConnectorsParams) GetSkip() int64 {
	return m.Skip
}

// SetSkip sets the Skip value and returns the updated SearchConnectorsParams.
func (m *SearchConnectorsParams) SetSkip(v int64) {
	m.Skip = v
}

// GetLimit returns the Limit value.
func (m SearchConnectorsParams) GetLimit() int64 {
	return m.Limit
}

// SetLimit sets the Limit value and returns the updated SearchConnectorsParams.
func (m *SearchConnectorsParams) SetLimit(v int64) {
	m.Limit = v
}

// GetFrom returns the From value.
func (m StartConnectorBackfillParams) GetFrom() *int64 {
	return m.From
}

// SetFrom sets the From value and returns the updated StartConnectorBackfillParams.
func (m *StartConnectorBackfillParams) SetFrom(v *int64) {
	m.From = v
}

// GetTo returns the To value.
func (m StartConnectorBackfillParams) GetTo() *int64 {
	return m.To
}

// SetTo sets the To value and returns the updated StartConnectorBackfillParams.
func (m *StartConnectorBackfillParams) SetTo(v *int64) {
	m.To = v
}

// GetType returns the Type value.
func (m StartConnectorBackfillParams) GetType() consts.ConnectorType {
	return m.Type
}

// SetType sets the Type value and returns the updated StartConnectorBackfillParams.
func (m *StartConnectorBackfillParams) SetType(v consts.ConnectorType) {
	m.Type = v
}

// GetSiteID returns the SiteID value.
func (m StartConnectorBackfillParams) GetSiteID() string {
	return m.SiteID
}

// SetSiteID sets the SiteID value and returns the updated StartConnectorBackfillParams.
func (m *StartConnectorBackfillParams) SetSiteID(v string) {
	m.SiteID = v
}

// GetPasswordEnc returns the PasswordEnc value.
func (m UpdateConnectorParams) GetPasswordEnc() *string {
	return m.PasswordEnc
}

// SetPasswordEnc sets the PasswordEnc value and returns the updated UpdateConnectorParams.
func (m *UpdateConnectorParams) SetPasswordEnc(v *string) {
	m.PasswordEnc = v
}

// GetEnabled returns the Enabled value.
func (m UpdateConnectorParams) GetEnabled() bool {
	return m.Enabled
}

// SetEnabled sets the Enabled value and returns the updated UpdateConnectorParams.
func (m *UpdateConnectorParams) SetEnabled(v bool) {
	m.Enabled = v
}

// GetUrlOrUsernameEnc returns the UrlOrUsernameEnc value.
func (m UpdateConnectorParams) GetUrlOrUsernameEnc() *string {
	return m.UrlOrUsernameEnc
}

// SetUrlOrUsernameEnc sets the UrlOrUsernameEnc value and returns the updated UpdateConnectorParams.
func (m *UpdateConnectorParams) SetUrlOrUsernameEnc(v *string) {
	m.UrlOrUsernameEnc = v
}

// GetRegion returns the Region value.
func (m UpdateConnectorParams) GetRegion() *string {
	return m.Region
}

// SetRegion sets the Region value and returns the updated UpdateConnectorParams.
func (m *UpdateConnectorParams) SetRegion(v *string) {
	m.Region = v
}

// GetType returns the Type value.
func (m UpdateConnectorParams) GetType() consts.ConnectorType {
	return m.Type
}

// SetType sets the Type value and returns the updated UpdateConnectorParams.
func (m *UpdateConnectorParams) SetType(v consts.ConnectorType) {
	m.Type = v
}

// GetSiteID returns the SiteID value.
func (m UpdateConnectorParams) GetSiteID() string {
	return m.SiteID
}

// SetSiteID sets the SiteID value and returns the updated UpdateConnectorParams.
func (m *UpdateConnectorParams) SetSiteID(v string) {
	m.SiteID = v
}

// GetType returns the Type value.
func (m UpdateConnectorAttemptParams) GetType() consts.ConnectorType {
	return m.Type
}

// SetType sets the Type value and returns the updated UpdateConnectorAttemptParams.
func (m *UpdateConnectorAttemptParams) SetType(v consts.ConnectorType) {
	m.Type = v
}

// GetSiteID returns the SiteID value.
func (m UpdateConnectorAttemptParams) GetSiteID() string {
	return m.SiteID
}

// SetSiteID sets the SiteID value and returns the updated UpdateConnectorAttemptParams.
func (m *UpdateConnectorAttemptParams) SetSiteID(v string) {
	m.SiteID = v
}

// GetErrorMessages returns the ErrorMessages value.
func (m UpdateConnectorReportsFailureParams) GetErrorMessages() errors.ErrorList {
	return m.ErrorMessages
}

// SetErrorMessages sets the ErrorMessages value and returns the updated UpdateConnectorReportsFailureParams.
func (m *UpdateConnectorReportsFailureParams) SetErrorMessages(v errors.ErrorList) {
	m.ErrorMessages = v
}

// GetElapsedMs returns the ElapsedMs value.
func (m UpdateConnectorReportsFailureParams) GetElapsedMs() *int64 {
	return m.ElapsedMs
}

// SetElapsedMs sets the ElapsedMs value and returns the updated UpdateConnectorReportsFailureParams.
func (m *UpdateConnectorReportsFailureParams) SetElapsedMs(v *int64) {
	m.ElapsedMs = v
}

// GetType returns the Type value.
func (m UpdateConnectorReportsFailureParams) GetType() consts.ConnectorType {
	return m.Type
}

// SetType sets the Type value and returns the updated UpdateConnectorReportsFailureParams.
func (m *UpdateConnectorReportsFailureParams) SetType(v consts.ConnectorType) {
	m.Type = v
}

// GetSiteID returns the SiteID value.
func (m UpdateConnectorReportsFailureParams) GetSiteID() string {
	return m.SiteID
}

// SetSiteID sets the SiteID value and returns the updated UpdateConnectorReportsFailureParams.
func (m *UpdateConnectorReportsFailureParams) SetSiteID(v string) {
	m.SiteID = v
}

// GetRecords returns the Records value.
func (m UpdateConnectorReportsSuccessParams) GetRecords() *int64 {
	return m.Records
}

// SetRecords sets the Records value and returns the updated UpdateConnectorReportsSuccessParams.
func (m *UpdateConnectorReportsSuccessParams) SetRecords(v *int64) {
	m.Records = v
}

// GetElapsedMs returns the ElapsedMs value.
func (m UpdateConnectorReportsSuccessParams) GetElapsedMs() *int64 {
	return m.ElapsedMs
}

// SetElapsedMs sets the ElapsedMs value and returns the updated UpdateConnectorReportsSuccessParams.
func (m *UpdateConnectorReportsSuccessParams) SetElapsedMs(v *int64) {
	m.ElapsedMs = v
}

// GetType returns the Type value.
func (m UpdateConnectorReportsSuccessParams) GetType() consts.ConnectorType {
	return m.Type
}

// SetType sets the Type value and returns the updated UpdateConnectorReportsSuccessParams.
func (m *UpdateConnectorReportsSuccessParams) SetType(v consts.ConnectorType) {
	m.Type = v
}

// GetSiteID returns the SiteID value.
func (m UpdateConnectorReportsSuccessParams) GetSiteID() string {
	return m.SiteID
}

// SetSiteID sets the SiteID value and returns the updated UpdateConnectorReportsSuccessParams.
func (m *UpdateConnectorReportsSuccessParams) SetSiteID(v string) {
	m.SiteID = v
}

// GetSiteID returns the SiteID value.
func (m AllSiteIdentitiesParams) GetSiteID() string {
	return m.SiteID
}

// SetSiteID sets the SiteID value and returns the updated AllSiteIdentitiesParams.
func (m *AllSiteIdentitiesParams) SetSiteID(v string) {
	m.SiteID = v
}

// GetSince returns the Since value.
func (m CountActiveIdentitiesParams) GetSince() *int64 {
	return m.Since
}

// SetSince sets the Since value and returns the updated CountActiveIdentitiesParams.
func (m *CountActiveIdentitiesParams) SetSince(v *int64) {
	m.Since = v
}

// GetSiteID returns the SiteID value.
func (m CountSiteIdentitiesParams) GetSiteID() string {
	return m.SiteID
}

// SetSiteID sets the SiteID value and returns the updated CountSiteIdentitiesParams.
func (m *CountSiteIdentitiesParams) SetSiteID(v string) {
	m.SiteID = v
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

// GetClient returns the Client value.
func (m CreateSiteIdentityParams) GetClient() *string {
	return m.Client
}

// SetClient sets the Client value and returns the updated CreateSiteIdentityParams.
func (m *CreateSiteIdentityParams) SetClient(v *string) {
	m.Client = v
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

// GetNightscoutTokenDigest returns the NightscoutTokenDigest value.
func (m CreateSiteIdentityParams) GetNightscoutTokenDigest() *string {
	return m.NightscoutTokenDigest
}

// SetNightscoutTokenDigest sets the NightscoutTokenDigest value and returns the updated CreateSiteIdentityParams.
func (m *CreateSiteIdentityParams) SetNightscoutTokenDigest(v *string) {
	m.NightscoutTokenDigest = v
}

// GetSiteID returns the SiteID value.
func (m DeleteAllSiteIdentitiesParams) GetSiteID() string {
	return m.SiteID
}

// SetSiteID sets the SiteID value and returns the updated DeleteAllSiteIdentitiesParams.
func (m *DeleteAllSiteIdentitiesParams) SetSiteID(v string) {
	m.SiteID = v
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

// GetSiteID returns the SiteID value.
func (m GetSiteIdentityByTokenDigestParams) GetSiteID() string {
	return m.SiteID
}

// SetSiteID sets the SiteID value and returns the updated GetSiteIdentityByTokenDigestParams.
func (m *GetSiteIdentityByTokenDigestParams) SetSiteID(v string) {
	m.SiteID = v
}

// GetDigest returns the Digest value.
func (m GetSiteIdentityByTokenDigestParams) GetDigest() *string {
	return m.Digest
}

// SetDigest sets the Digest value and returns the updated GetSiteIdentityByTokenDigestParams.
func (m *GetSiteIdentityByTokenDigestParams) SetDigest(v *string) {
	m.Digest = v
}

// GetPasswordHash returns the PasswordHash value.
func (m ResetSiteAdminCredentialsParams) GetPasswordHash() string {
	return m.PasswordHash
}

// SetPasswordHash sets the PasswordHash value and returns the updated ResetSiteAdminCredentialsParams.
func (m *ResetSiteAdminCredentialsParams) SetPasswordHash(v string) {
	m.PasswordHash = v
}

// GetTotpSecretEnc returns the TotpSecretEnc value.
func (m ResetSiteAdminCredentialsParams) GetTotpSecretEnc() string {
	return m.TotpSecretEnc
}

// SetTotpSecretEnc sets the TotpSecretEnc value and returns the updated ResetSiteAdminCredentialsParams.
func (m *ResetSiteAdminCredentialsParams) SetTotpSecretEnc(v string) {
	m.TotpSecretEnc = v
}

// GetSiteID returns the SiteID value.
func (m ResetSiteAdminCredentialsParams) GetSiteID() string {
	return m.SiteID
}

// SetSiteID sets the SiteID value and returns the updated ResetSiteAdminCredentialsParams.
func (m *ResetSiteAdminCredentialsParams) SetSiteID(v string) {
	m.SiteID = v
}

// GetName returns the Name value.
func (m ResetSiteAdminCredentialsParams) GetName() string {
	return m.Name
}

// SetName sets the Name value and returns the updated ResetSiteAdminCredentialsParams.
func (m *ResetSiteAdminCredentialsParams) SetName(v string) {
	m.Name = v
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

// GetPasswordHash returns the PasswordHash value.
func (m UpdateSiteIdentityPasswordParams) GetPasswordHash() string {
	return m.PasswordHash
}

// SetPasswordHash sets the PasswordHash value and returns the updated UpdateSiteIdentityPasswordParams.
func (m *UpdateSiteIdentityPasswordParams) SetPasswordHash(v string) {
	m.PasswordHash = v
}

// GetNightscoutTokenDigest returns the NightscoutTokenDigest value.
func (m UpdateSiteIdentityPasswordParams) GetNightscoutTokenDigest() *string {
	return m.NightscoutTokenDigest
}

// SetNightscoutTokenDigest sets the NightscoutTokenDigest value and returns the updated UpdateSiteIdentityPasswordParams.
func (m *UpdateSiteIdentityPasswordParams) SetNightscoutTokenDigest(v *string) {
	m.NightscoutTokenDigest = v
}

// GetUpdatedByIdentityName returns the UpdatedByIdentityName value.
func (m UpdateSiteIdentityPasswordParams) GetUpdatedByIdentityName() *string {
	return m.UpdatedByIdentityName
}

// SetUpdatedByIdentityName sets the UpdatedByIdentityName value and returns the updated UpdateSiteIdentityPasswordParams.
func (m *UpdateSiteIdentityPasswordParams) SetUpdatedByIdentityName(v *string) {
	m.UpdatedByIdentityName = v
}

// GetSiteID returns the SiteID value.
func (m UpdateSiteIdentityPasswordParams) GetSiteID() string {
	return m.SiteID
}

// SetSiteID sets the SiteID value and returns the updated UpdateSiteIdentityPasswordParams.
func (m *UpdateSiteIdentityPasswordParams) SetSiteID(v string) {
	m.SiteID = v
}

// GetID returns the ID value.
func (m UpdateSiteIdentityPasswordParams) GetID() int64 {
	return m.ID
}

// SetID sets the ID value and returns the updated UpdateSiteIdentityPasswordParams.
func (m *UpdateSiteIdentityPasswordParams) SetID(v int64) {
	m.ID = v
}

// GetTotpSecretEnc returns the TotpSecretEnc value.
func (m UpdateSiteIdentityTOTPParams) GetTotpSecretEnc() string {
	return m.TotpSecretEnc
}

// SetTotpSecretEnc sets the TotpSecretEnc value and returns the updated UpdateSiteIdentityTOTPParams.
func (m *UpdateSiteIdentityTOTPParams) SetTotpSecretEnc(v string) {
	m.TotpSecretEnc = v
}

// GetUpdatedByIdentityName returns the UpdatedByIdentityName value.
func (m UpdateSiteIdentityTOTPParams) GetUpdatedByIdentityName() *string {
	return m.UpdatedByIdentityName
}

// SetUpdatedByIdentityName sets the UpdatedByIdentityName value and returns the updated UpdateSiteIdentityTOTPParams.
func (m *UpdateSiteIdentityTOTPParams) SetUpdatedByIdentityName(v *string) {
	m.UpdatedByIdentityName = v
}

// GetSiteID returns the SiteID value.
func (m UpdateSiteIdentityTOTPParams) GetSiteID() string {
	return m.SiteID
}

// SetSiteID sets the SiteID value and returns the updated UpdateSiteIdentityTOTPParams.
func (m *UpdateSiteIdentityTOTPParams) SetSiteID(v string) {
	m.SiteID = v
}

// GetID returns the ID value.
func (m UpdateSiteIdentityTOTPParams) GetID() int64 {
	return m.ID
}

// SetID sets the ID value and returns the updated UpdateSiteIdentityTOTPParams.
func (m *UpdateSiteIdentityTOTPParams) SetID(v int64) {
	m.ID = v
}

// GetJsonEach returns the JsonEach value.
func (m ArchiveSitesParams) GetJsonEach() interface{} {
	return m.JsonEach
}

// SetJsonEach sets the JsonEach value and returns the updated ArchiveSitesParams.
func (m *ArchiveSitesParams) SetJsonEach(v interface{}) {
	m.JsonEach = v
}

// GetSince returns the Since value.
func (m CountActiveSitesParams) GetSince() *int64 {
	return m.Since
}

// SetSince sets the Since value and returns the updated CountActiveSitesParams.
func (m *CountActiveSitesParams) SetSince(v *int64) {
	m.Since = v
}

// GetSince returns the Since value.
func (m CountSitesCreatedParams) GetSince() int64 {
	return m.Since
}

// SetSince sets the Since value and returns the updated CountSitesCreatedParams.
func (m *CountSitesCreatedParams) SetSince(v int64) {
	m.Since = v
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

// GetJsonEach returns the JsonEach value.
func (m DeleteSitesParams) GetJsonEach() interface{} {
	return m.JsonEach
}

// SetJsonEach sets the JsonEach value and returns the updated DeleteSitesParams.
func (m *DeleteSitesParams) SetJsonEach(v interface{}) {
	m.JsonEach = v
}

// GetID returns the ID value.
func (m DisableSiteParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated DisableSiteParams.
func (m *DisableSiteParams) SetID(v string) {
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

// GetBatch returns the Batch value.
func (m GetBatchSitesToBackUpParams) GetBatch() string {
	return m.Batch
}

// SetBatch sets the Batch value and returns the updated GetBatchSitesToBackUpParams.
func (m *GetBatchSitesToBackUpParams) SetBatch(v string) {
	m.Batch = v
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

// GetID returns the ID value.
func (m HydrateSiteParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated HydrateSiteParams.
func (m *HydrateSiteParams) SetID(v string) {
	m.ID = v
}

// GetNewHash returns the NewHash value.
func (m RotateSiteBackupKeyParams) GetNewHash() string {
	return m.NewHash
}

// SetNewHash sets the NewHash value and returns the updated RotateSiteBackupKeyParams.
func (m *RotateSiteBackupKeyParams) SetNewHash(v string) {
	m.NewHash = v
}

// GetID returns the ID value.
func (m RotateSiteBackupKeyParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated RotateSiteBackupKeyParams.
func (m *RotateSiteBackupKeyParams) SetID(v string) {
	m.ID = v
}

// GetOldHash returns the OldHash value.
func (m RotateSiteBackupKeyParams) GetOldHash() string {
	return m.OldHash
}

// SetOldHash sets the OldHash value and returns the updated RotateSiteBackupKeyParams.
func (m *RotateSiteBackupKeyParams) SetOldHash(v string) {
	m.OldHash = v
}

// GetSkip returns the Skip value.
func (m SearchArchivedSitesParams) GetSkip() int64 {
	return m.Skip
}

// SetSkip sets the Skip value and returns the updated SearchArchivedSitesParams.
func (m *SearchArchivedSitesParams) SetSkip(v int64) {
	m.Skip = v
}

// GetLimit returns the Limit value.
func (m SearchArchivedSitesParams) GetLimit() int64 {
	return m.Limit
}

// SetLimit sets the Limit value and returns the updated SearchArchivedSitesParams.
func (m *SearchArchivedSitesParams) SetLimit(v int64) {
	m.Limit = v
}

// GetSkip returns the Skip value.
func (m SearchSitesParams) GetSkip() int64 {
	return m.Skip
}

// SetSkip sets the Skip value and returns the updated SearchSitesParams.
func (m *SearchSitesParams) SetSkip(v int64) {
	m.Skip = v
}

// GetLimit returns the Limit value.
func (m SearchSitesParams) GetLimit() int64 {
	return m.Limit
}

// SetLimit sets the Limit value and returns the updated SearchSitesParams.
func (m *SearchSitesParams) SetLimit(v int64) {
	m.Limit = v
}

// GetSkip returns the Skip value.
func (m SearchSitesByDBSizeParams) GetSkip() int64 {
	return m.Skip
}

// SetSkip sets the Skip value and returns the updated SearchSitesByDBSizeParams.
func (m *SearchSitesByDBSizeParams) SetSkip(v int64) {
	m.Skip = v
}

// GetLimit returns the Limit value.
func (m SearchSitesByDBSizeParams) GetLimit() int64 {
	return m.Limit
}

// SetLimit sets the Limit value and returns the updated SearchSitesByDBSizeParams.
func (m *SearchSitesByDBSizeParams) SetLimit(v int64) {
	m.Limit = v
}

// GetSkip returns the Skip value.
func (m SearchSitesByLastWriteParams) GetSkip() int64 {
	return m.Skip
}

// SetSkip sets the Skip value and returns the updated SearchSitesByLastWriteParams.
func (m *SearchSitesByLastWriteParams) SetSkip(v int64) {
	m.Skip = v
}

// GetLimit returns the Limit value.
func (m SearchSitesByLastWriteParams) GetLimit() int64 {
	return m.Limit
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

// GetEstimatedDbSizeKb returns the EstimatedDbSizeKb value.
func (m UpdateSiteDBSizeParams) GetEstimatedDbSizeKb() *int64 {
	return m.EstimatedDbSizeKb
}

// SetEstimatedDbSizeKb sets the EstimatedDbSizeKb value and returns the updated UpdateSiteDBSizeParams.
func (m *UpdateSiteDBSizeParams) SetEstimatedDbSizeKb(v *int64) {
	m.EstimatedDbSizeKb = v
}

// GetID returns the ID value.
func (m UpdateSiteDBSizeParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated UpdateSiteDBSizeParams.
func (m *UpdateSiteDBSizeParams) SetID(v string) {
	m.ID = v
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

// GetEstimatedDbSizeKb returns the EstimatedDbSizeKb value.
func (m UpdateSiteDateFlagsParams) GetEstimatedDbSizeKb() *int64 {
	return m.EstimatedDbSizeKb
}

// SetEstimatedDbSizeKb sets the EstimatedDbSizeKb value and returns the updated UpdateSiteDateFlagsParams.
func (m *UpdateSiteDateFlagsParams) SetEstimatedDbSizeKb(v *int64) {
	m.EstimatedDbSizeKb = v
}

// GetID returns the ID value.
func (m UpdateSiteDateFlagsParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated UpdateSiteDateFlagsParams.
func (m *UpdateSiteDateFlagsParams) SetID(v string) {
	m.ID = v
}

