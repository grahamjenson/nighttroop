
package controldb

import (
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
func (m NewestPushRegistrationForAppParams) GetApp() string {
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

// GetID returns the ID value.
func (m PatchSiteIdentitySettingsParams) GetID() int64 {
	return m.ID
}

// SetID sets the ID value and returns the updated PatchSiteIdentitySettingsParams.
func (m *PatchSiteIdentitySettingsParams) SetID(v int64) {
	m.ID = v
}

// GetID returns the ID value.
func (m PatchSiteSettingsParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated PatchSiteSettingsParams.
func (m *PatchSiteSettingsParams) SetID(v string) {
	m.ID = v
}

// GetID returns the ID value.
func (m UpdateSiteIdentitySettingsParams) GetID() int64 {
	return m.ID
}

// SetID sets the ID value and returns the updated UpdateSiteIdentitySettingsParams.
func (m *UpdateSiteIdentitySettingsParams) SetID(v int64) {
	m.ID = v
}

// GetID returns the ID value.
func (m UpdateSiteSettingsParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated UpdateSiteSettingsParams.
func (m *UpdateSiteSettingsParams) SetID(v string) {
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
func (m DeleteSiteIdentityNamedNotAdminParams) GetName() string {
	return m.Name
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
func (m EnableSiteParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated EnableSiteParams.
func (m *EnableSiteParams) SetID(v string) {
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

// SetLimit sets the Limit value and returns the updated SearchSitesParams.
func (m *SearchSitesParams) SetLimit(v int64) {
	m.Limit = v
}

// SetLimit sets the Limit value and returns the updated SearchSitesByDBSizeParams.
func (m *SearchSitesByDBSizeParams) SetLimit(v int64) {
	m.Limit = v
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

