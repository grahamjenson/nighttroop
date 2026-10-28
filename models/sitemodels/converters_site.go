package sitemodels

import (
	"time"

	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/utils"
	"nighttroop.t1d.nz/utils/authclients"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/httputils"
	"nighttroop.t1d.nz/utils/security"
	"nighttroop.t1d.nz/utils/settings"
)

// goverter:converter
// goverter:enum:unknown @ignore
// goverter:output:format function
// goverter:output:file ./z_site_converters.go
// goverter:useZeroValueOnPointerInconsistency
type SiteConverter interface {

	///////
	// SYNC SOURCE
	///////

	// goverter:map Password | SetPasswordGarbage
	// goverter:map UrlOrUsernameEnc UrlOrUsername | Decrypt
	// goverter:ignore RecentEvents RecentErrors
	// goverter:context security
	SyncSourceFromDB(source *controldb.SiteSyncSource, security *security.SecurityConfig) (*SyncSource, error)

	// goverter:map Password PasswordEnc
	// goverter:map UrlOrUsername UrlOrUsernameEnc
	SyncSourceToUpdate(source *SyncSource) controldb.UpdateSyncSourceParams

	// goverter:map Password PasswordEnc
	// goverter:map UrlOrUsername UrlOrUsernameEnc
	SyncSourceToCreate(source *SyncSource) controldb.CreateSyncSourceParams

	///////
	// SITE IDENTITY
	///////

	// goverter:context siteContext
	// goverter:map ID CurrentIdentity | IsCurrentIdentity
	// goverter:map TotpSecretEnc HasTotpSecret | HasTotpSecret
	// goverter:map ID RecentClients | RecentClients
	// goverter:ignore Type Password TotpSecret NSToken SiteURL
	IdentityFromDB(source *controldb.SiteIdentity, siteContext *httputils.SiteContext) *UserIdentity

	// goverter:context siteContext
	// goverter:map SiteID | ContextSiteID
	// goverter:map UpdatedByIdentityID | ContextIdentityID
	// goverter:map UpdatedByIdentityName | ContextIdentityName
	IdentityToUpdate(source *UserIdentity, siteContext *httputils.SiteContext) controldb.UpdateSiteIdentityEnabledParams

	// goverter:context siteContext
	// goverter:map SiteID | ContextSiteID
	IdentityToDelete(source *UserIdentity, siteContext *httputils.SiteContext) controldb.DeleteSiteIdentityParams

	///////
	// GLUCOSE READINGS
	///////

	// goverter:context siteContext
	// goverter:map Mgdl Reading | MgdlToSettingsUnits
	// goverter:map Units OriginalUnits
	// goverter:map Units | ContextUnits
	GlucoseReadingFromDB(source *sitedb.GlucoseReading, siteContext *httputils.SiteContext) *SiteGlucoseReading

	// goverter:context siteContext
	// goverter:map ID | NewSiteDataID
	// goverter:map Reading Mgdl | SettingsUnitsToMgdl
	// goverter:map Units | ContextUnits
	// goverter:map ReadingType | GlucoseReadingTypeOrMeter
	// goverter:map OccurredAt | OccurredAtOrNow
	// goverter:map CreatedByIdentityID | ContextIdentityID
	// goverter:map CreatedByIdentityName | ContextIdentityNameValue
	GlucoseReadingToCreate(source *SiteGlucoseReading, siteContext *httputils.SiteContext) sitedb.CreateGlucoseReadingParams

	// goverter:context siteContext
	// goverter:map Reading Mgdl | SettingsUnitsToMgdlOrNil
	// goverter:map UpdatedByIdentityID | ContextIdentityID
	// goverter:map UpdatedByIdentityName | ContextIdentityName
	// goverter:ignore RawJsonData
	GlucoseReadingToUpdate(source *SiteGlucoseReading, siteContext *httputils.SiteContext) sitedb.UpdateGlucoseReadingParams

	///////
	// TEMP TARGETS
	///////

	// goverter:context siteContext
	// goverter:map HighTargetMgdl HighTarget | MgdlToSettingsUnits
	// goverter:map LowTargetMgdl LowTarget | MgdlToSettingsUnits
	// goverter:map . Units | ContextUnits
	TempTargetFromDB(source *sitedb.TempTarget, siteContext *httputils.SiteContext) *SiteTempTarget

	// goverter:context siteContext
	// goverter:map ID | NewSiteDataID
	// goverter:map HighTarget HighTargetMgdl | SettingsUnitsToMgdl
	// goverter:map LowTarget LowTargetMgdl | SettingsUnitsToMgdl
	// goverter:map OccurredAt | OccurredAtOrNow
	// goverter:map CreatedByIdentityID | ContextIdentityID
	// goverter:map CreatedByIdentityName | ContextIdentityNameValue
	TempTargetToCreate(source *SiteTempTarget, siteContext *httputils.SiteContext) sitedb.CreateTempTargetParams

	// goverter:context siteContext
	// goverter:map DurationMinutes | Int64OrNil
	// goverter:map HighTarget HighTargetMgdl | SettingsUnitsToMgdlOrNil
	// goverter:map LowTarget LowTargetMgdl | SettingsUnitsToMgdlOrNil
	// goverter:map UpdatedByIdentityID | ContextIdentityID
	// goverter:map UpdatedByIdentityName | ContextIdentityName
	// goverter:ignore RawJsonData
	TempTargetToUpdate(source *SiteTempTarget, siteContext *httputils.SiteContext) sitedb.UpdateTempTargetParams

	///////
	// TEMP BASALS
	///////

	TempBasalFromDB(source *sitedb.TempBasal) *SiteTempBasal

	// goverter:context siteContext
	// goverter:map ID | NewSiteDataID
	// goverter:map OccurredAt | OccurredAtOrNow
	// goverter:map CreatedByIdentityID | ContextIdentityID
	// goverter:map CreatedByIdentityName | ContextIdentityNameValue
	TempBasalToCreate(source *SiteTempBasal, siteContext *httputils.SiteContext) sitedb.CreateTempBasalParams

	// goverter:context siteContext
	// goverter:map DurationMinutes | Int64OrNil
	// goverter:map UpdatedByIdentityID | ContextIdentityID
	// goverter:map UpdatedByIdentityName | ContextIdentityName
	// goverter:ignore RawJsonData
	TempBasalToUpdate(source *SiteTempBasal, siteContext *httputils.SiteContext) sitedb.UpdateTempBasalParams

	///////
	// ACTIVITIES
	///////

	ActivityFromDB(source *sitedb.Activity) *SiteActivity

	// goverter:context siteContext
	// goverter:map ID | NewSiteDataID
	// goverter:map OccurredAt | OccurredAtOrNow
	// goverter:map CreatedByIdentityID | ContextIdentityID
	// goverter:map CreatedByIdentityName | ContextIdentityNameValue
	ActivityToCreate(source *SiteActivity, siteContext *httputils.SiteContext) sitedb.CreateActivityParams

	// goverter:context siteContext
	// goverter:map ActivityType | StringOrNil
	// goverter:map UpdatedByIdentityID | ContextIdentityID
	// goverter:map UpdatedByIdentityName | ContextIdentityName
	// goverter:ignore RawJsonData
	ActivityToUpdate(source *SiteActivity, siteContext *httputils.SiteContext) sitedb.UpdateActivityParams

	///////
	// COMPONENT EVENTS
	///////

	ComponentEventFromDB(source *sitedb.ComponentEvent) *SiteComponentEvent

	// goverter:context siteContext
	// goverter:map ID | NewSiteDataID
	// goverter:map OccurredAt | OccurredAtOrNow
	// goverter:map CreatedByIdentityID | ContextIdentityID
	// goverter:map CreatedByIdentityName | ContextIdentityNameValue
	ComponentEventToCreate(source *SiteComponentEvent, siteContext *httputils.SiteContext) sitedb.CreateComponentEventParams

	// goverter:context siteContext
	// goverter:map EventType | StringOrNil
	// goverter:map UpdatedByIdentityID | ContextIdentityID
	// goverter:map UpdatedByIdentityName | ContextIdentityName
	// goverter:ignore RawJsonData
	ComponentEventToUpdate(source *SiteComponentEvent, siteContext *httputils.SiteContext) sitedb.UpdateComponentEventParams

	///////
	// NOTIFICATION EVENTS
	///////

	NotificationEventFromDB(source *sitedb.NotificationEvent) *SiteNotificationEvent

	// goverter:context siteContext
	// goverter:map ID | NewSiteDataID
	// goverter:map OccurredAt | OccurredAtOrNow
	// goverter:map CreatedByIdentityID | ContextIdentityID
	// goverter:map CreatedByIdentityName | ContextIdentityNameValue
	NotificationEventToCreate(source *SiteNotificationEvent, siteContext *httputils.SiteContext) sitedb.CreateNotificationEventParams

	// goverter:context siteContext
	// goverter:map NotificationType | StringOrNil
	// goverter:map Title | StringOrNil
	// goverter:map Severity | Int64OrNil
	// goverter:map UpdatedByIdentityID | ContextIdentityID
	// goverter:map UpdatedByIdentityName | ContextIdentityName
	// goverter:ignore RawJsonData
	NotificationEventToUpdate(source *SiteNotificationEvent, siteContext *httputils.SiteContext) sitedb.UpdateNotificationEventParams

	///////
	// PROFILE SWITCHES
	///////

	ProfileSwitchFromDB(source *sitedb.ProfileSwitch) *SiteProfileSwitch

	// goverter:context siteContext
	// goverter:map ID | NewSiteDataID
	// goverter:map OccurredAt | OccurredAtOrNow
	// goverter:map CreatedByIdentityID | ContextIdentityID
	// goverter:map CreatedByIdentityName | ContextIdentityNameValue
	// goverter:ignore Dia Units Timezone CarbsHr Delay CarbRatiosJson InsulinSensitivitiesJson BasalRatesJson TargetsHighJson TargetsLowJson
	ProfileSwitchToCreate(source *SiteProfileSwitch, siteContext *httputils.SiteContext) sitedb.CreateProfileSwitchParams

	// goverter:context siteContext
	// goverter:map Name | StringOrNil
	// goverter:map UpdatedByIdentityID | ContextIdentityID
	// goverter:map UpdatedByIdentityName | ContextIdentityName
	// goverter:ignore RawJsonData
	ProfileSwitchToUpdate(source *SiteProfileSwitch, siteContext *httputils.SiteContext) sitedb.UpdateProfileSwitchParams

	///////
	// EVENT GROUPS
	///////

	// goverter:context siteContext
	// goverter:autoMap EventGroup
	// goverter:map EventGroup.ProfileSwitchID ProfileSwitchID
	// goverter:map EventGroup.ComponentEventID ComponentEventID
	// goverter:map EventGroup.ActivityID ActivityID
	// goverter:map EventGroup.NotificationEventID NotificationEventID
	EventGroupFromDB(source *sitedb.EventGroupsBundle, siteContext *httputils.SiteContext) *SiteEventGroup

	///////
	// INSULIN DOSES
	///////

	InsulinDoseFromDB(source *sitedb.InsulinDose) *SiteInsulinDose

	// goverter:context siteContext
	// goverter:map ID | NewSiteDataID
	// goverter:map OccurredAt | OccurredAtOrNow
	// goverter:map CreatedByIdentityID | ContextIdentityID
	// goverter:map CreatedByIdentityName | ContextIdentityNameValue
	InsulinDoseToCreate(source *SiteInsulinDose, siteContext *httputils.SiteContext) sitedb.CreateInsulinDoseParams

	// goverter:context siteContext
	// goverter:map Dose | Float64OrNil
	// goverter:map UpdatedByIdentityID | ContextIdentityID
	// goverter:map UpdatedByIdentityName | ContextIdentityName
	// goverter:ignore RawJsonData
	InsulinDoseToUpdate(source *SiteInsulinDose, siteContext *httputils.SiteContext) sitedb.UpdateInsulinDoseParams

	///////
	// FOOD INTAKES
	///////

	FoodIntakeFromDB(source *sitedb.FoodIntake) *SiteFoodIntake

	// goverter:context siteContext
	// goverter:map ID | NewSiteDataID
	// goverter:map OccurredAt | OccurredAtOrNow
	// goverter:map CreatedByIdentityID | ContextIdentityID
	// goverter:map CreatedByIdentityName | ContextIdentityNameValue
	FoodIntakeToCreate(source *SiteFoodIntake, siteContext *httputils.SiteContext) sitedb.CreateFoodIntakeParams

	// goverter:context siteContext
	// goverter:map UpdatedByIdentityID | ContextIdentityID
	// goverter:map UpdatedByIdentityName | ContextIdentityName
	// goverter:ignore RawJsonData
	FoodIntakeToUpdate(source *SiteFoodIntake, siteContext *httputils.SiteContext) sitedb.UpdateFoodIntakeParams

	///////
	// FOOD PRESETS
	///////

	FoodPresetFromDB(source *sitedb.FoodPreset) *SiteFoodPreset

	// goverter:context siteContext
	// goverter:map ID | NewSiteDataID
	// goverter:map CreatedByIdentityID | ContextIdentityID
	// goverter:map CreatedByIdentityName | ContextIdentityNameValue
	// goverter:ignore RawJsonData
	FoodPresetToCreate(source *SiteFoodPreset, siteContext *httputils.SiteContext) sitedb.CreateFoodPresetParams

	// goverter:context siteContext
	// goverter:map UpdatedByIdentityID | ContextIdentityID
	// goverter:map UpdatedByIdentityName | ContextIdentityName
	// goverter:ignore RawJsonData
	FoodPresetToUpdate(source *SiteFoodPreset, siteContext *httputils.SiteContext) sitedb.UpdateFoodPresetParams

	///////
	// PROFILE PRESETS
	///////

	// goverter:context siteContext
	// goverter:map CarbRatiosJson CarbRatios | ScheduleFromJSON
	// goverter:map BasalRatesJson BasalRates | ScheduleFromJSON
	// goverter:map . Sensitivities | ProfileSensitivities
	// goverter:map TargetsLowJson LowTargets | GlucoseScheduleFromJSON
	// goverter:map TargetsHighJson HighTargets | GlucoseScheduleFromJSON
	// goverter:map . Units | ContextUnits
	ProfilePresetFromDB(source *sitedb.ProfilePreset, siteContext *httputils.SiteContext) *SiteProfilePreset

	// goverter:context siteContext
	// goverter:map UpdatedByIdentityID | ContextIdentityID
	// goverter:map UpdatedByIdentityName | ContextIdentityName
	// goverter:ignore RawJsonData
	ProfilePresetToUpdate(source *SiteProfilePreset, siteContext *httputils.SiteContext) sitedb.UpdateProfilePresetParams

	///////
	// DEVICE STATUSES
	///////

	DeviceStatusFromDB(source *sitedb.DeviceReport) *SiteDeviceStatus

	///////
	// SETTINGS
	///////

	// goverter:map GetUnits Units
	// goverter:map GetTimezone Timezone
	// goverter:map GetLanguage Language
	// goverter:map GetDateFormat DateFormat
	// goverter:map GetTimeFormat TimeFormat
	// goverter:map GetWeekStartDay WeekStartDay
	// goverter:ignore Inherited Site
	SettingsFromMap(source *settings.SettingMap) *Settings
}

func SetPasswordGarbage() *string {
	return ptr(consts.PASSWORD_GARBAGE_VALUE)
}

// goverter:context security
func Decrypt(encString *string, security *security.SecurityConfig) (*string, error) {
	if encString != nil {
		decString, err := security.Decrypt(*encString)
		if err != nil {
			return nil, err
		}
		encString = &decString
	}
	if encString != nil {
		return encString, nil
	}
	return nil, nil
}

// goverter:context siteContext
func IsCurrentIdentity(id int64, siteContext *httputils.SiteContext) bool {
	return id == siteContext.IdentityModel.ID
}

// goverter:context siteContext
func RecentClients(id int64, siteContext *httputils.SiteContext) []authclients.Client {
	if siteContext.Config == nil || siteContext.Config.RecentClients == nil {
		return nil
	}
	return siteContext.Config.RecentClients.Get(httputils.AuthClientKey(siteContext.SiteID, id))
}

// goverter:context siteContext
func ContextSiteID(siteContext *httputils.SiteContext) string {
	return siteContext.SiteID
}

// goverter:context siteContext
func ContextIdentityID(siteContext *httputils.SiteContext) *int64 {
	return &siteContext.IdentityModel.ID
}

// goverter:context siteContext
func ContextIdentityName(siteContext *httputils.SiteContext) *string {
	return &siteContext.IdentityModel.Name
}

func HasTotpSecret(totpSecretEnc string) bool {
	return totpSecretEnc != ""
}

func ptr[T any](v T) *T {
	return &v
}

// goverter:context siteContext
func ContextIdentityNameValue(siteContext *httputils.SiteContext) string {
	return siteContext.IdentityModel.Name
}

// NewSiteDataID ignores any id sent, so data created on the site always gets a new id
func NewSiteDataID() string {
	return utils.NextID()
}

// GlucoseReadingTypeOrMeter defaults to mbg, as a reading typed in is from a meter
func GlucoseReadingTypeOrMeter(readingType string) string {
	if readingType == "" {
		return "mbg"
	}
	return readingType
}

func OccurredAtOrNow(occurredAt int64) int64 {
	if occurredAt == 0 {
		return time.Now().UnixMilli()
	}
	return occurredAt
}

// Float64OrNil leaves a value as it is when an update does not send one
func Float64OrNil(value float64) *float64 {
	if value == 0 {
		return nil
	}
	return &value
}

// Int64OrNil leaves a value as it is when an update does not send one
func Int64OrNil(value int64) *int64 {
	if value == 0 {
		return nil
	}
	return &value
}

// StringOrNil leaves a value as it is when an update does not send one
func StringOrNil(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

// ContextUnits are the units glucose is shown and typed in, from the identity's, site's or default settings
// goverter:context siteContext
func ContextUnits(siteContext *httputils.SiteContext) *consts.Units {
	return siteContext.GetSettings().GetUnits()
}

// goverter:context siteContext
func MgdlToSettingsUnits(mgdl float64, siteContext *httputils.SiteContext) float64 {
	return ContextUnits(siteContext).FromMGDL(mgdl)
}

// goverter:context siteContext
func SettingsUnitsToMgdl(value float64, siteContext *httputils.SiteContext) float64 {
	return ContextUnits(siteContext).ToMGDL(value)
}

// SettingsUnitsToMgdlOrNil leaves a glucose value as it is when an update does not send one
// goverter:context siteContext
func SettingsUnitsToMgdlOrNil(value float64, siteContext *httputils.SiteContext) *float64 {
	return Float64OrNil(SettingsUnitsToMgdl(value, siteContext))
}
