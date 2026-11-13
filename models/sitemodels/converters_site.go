package sitemodels

import (
	"encoding/json"
	"time"

	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/utils"
	"nighttroop.t1d.nz/utils/authclients"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/httputils"
	"nighttroop.t1d.nz/utils/security"
)

// goverter:converter
// goverter:enum:unknown @ignore
// goverter:output:format function
// goverter:output:file ./z_site_converters.go
// goverter:useZeroValueOnPointerInconsistency
type SiteConverter interface {

	///////
	// CONNECTOR
	///////

	// goverter:map Password | SetPasswordGarbage
	// goverter:map . UrlOrUsername | DecryptUrlOrUsername
	// goverter:ignore RecentEvents RecentErrors
	// goverter:context security
	ConnectorFromDB(source *controldb.SiteConnector, security *security.SecurityConfig) (*Connector, error)

	// goverter:map Password PasswordEnc
	// goverter:map UrlOrUsername UrlOrUsernameEnc
	// goverter:ignore UrlHostMac
	ConnectorToUpdate(source *Connector) controldb.UpdateConnectorParams

	// goverter:map Password PasswordEnc
	// goverter:map UrlOrUsername UrlOrUsernameEnc
	// goverter:ignore UrlHostMac
	ConnectorToCreate(source *Connector) controldb.CreateConnectorParams

	///////
	// SITE IDENTITY
	///////

	// goverter:context siteContext
	// goverter:map ID CurrentIdentity | IsCurrentIdentity
	// goverter:map TotpSecretEnc HasTotpSecret | HasTotpSecret
	// goverter:map ID RecentClients | RecentClients
	// goverter:ignore Type Password TotpSecret NSToken SiteURL NSURL
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
	// PUSH CREDENTIALS
	///////

	// goverter:map AppsJson Apps | JSONStrings
	// goverter:map BundleIdsJson BundleIDs | JSONStringsPtr
	// goverter:ignore Secret
	PushCredentialFromDB(source *controldb.PushCredential) *PushCredential

	// goverter:context siteContext
	// goverter:map SiteID | ContextSiteIDPtr
	// goverter:map Apps AppsJson | StringsJSON
	// goverter:map BundleIDs BundleIdsJson | StringsJSONPtr
	// goverter:map CreatedByIdentityName | ContextIdentityNameValue
	// goverter:ignore SecretEnc
	PushCredentialToCreate(source *PushCredential, siteContext *httputils.SiteContext) controldb.CreatePushCredentialParams

	///////
	// PUSH REGISTRATIONS
	///////

	// goverter:ignore Token IdentityName
	PushRegistrationFromDB(source *controldb.PushRegistration) *PushRegistration

	///////
	// GLUCOSE READINGS
	///////

	// goverter:context siteContext
	// goverter:map Mgdl Reading | MgdlToSettingsUnits
	// goverter:map Units OriginalUnits
	// goverter:map Units | ContextUnits
	// App's text is shown without its tags
	// goverter:map Notes | StripTags
	// goverter:map Reason | StripTags
	GlucoseReadingFromDB(source *sitedb.GlucoseReading, siteContext *httputils.SiteContext) *SiteGlucoseReading

	// goverter:context siteContext
	// goverter:map ID | NewSiteDataID
	// goverter:map Reading Mgdl | SettingsUnitsToMgdl
	// goverter:map Units | ContextUnits
	// goverter:map ReadingType | GlucoseReadingTypeOrMeter
	// goverter:map OccurredAt | OccurredAtOrNow
	// goverter:map CreatedByIdentityID | ContextIdentityID
	// goverter:map CreatedByIdentityName | ContextIdentityNameValue
	// The site does not set an AAPS calibration
	// goverter:ignore IsCalibration SensorMgdlAtPairing
	GlucoseReadingToCreate(source *SiteGlucoseReading, siteContext *httputils.SiteContext) sitedb.CreateGlucoseReadingParams

	// goverter:context siteContext
	// goverter:map Reading Mgdl | SettingsUnitsToMgdlOrNil
	// goverter:map UpdatedByIdentityID | ContextIdentityID
	// goverter:map UpdatedByIdentityName | ContextIdentityName
	// The site does not set an AAPS calibration
	// goverter:ignore IsCalibration SensorMgdlAtPairing
	GlucoseReadingToUpdate(source *SiteGlucoseReading, siteContext *httputils.SiteContext) sitedb.UpdateGlucoseReadingParams

	///////
	// TEMP TARGETS
	///////

	// goverter:context siteContext
	// goverter:map HighTargetMgdl HighTarget | MgdlToSettingsUnits
	// goverter:map LowTargetMgdl LowTarget | MgdlToSettingsUnits
	// goverter:map . Units | ContextUnits
	// App's text is shown without its tags
	// goverter:map Notes | StripTags
	// goverter:map Reason | StripTags
	TempTargetFromDB(source *sitedb.TempTarget, siteContext *httputils.SiteContext) *SiteTempTarget

	// goverter:context siteContext
	// goverter:map ID | NewSiteDataID
	// goverter:map HighTarget HighTargetMgdl | SettingsUnitsToMgdl
	// goverter:map LowTarget LowTargetMgdl | SettingsUnitsToMgdl
	// goverter:map OccurredAt | OccurredAtOrNow
	// goverter:map CreatedByIdentityID | ContextIdentityID
	// goverter:map CreatedByIdentityName | ContextIdentityNameValue
	// Only a loop's upload says an override has no end
	// goverter:ignore DurationType
	TempTargetToCreate(source *SiteTempTarget, siteContext *httputils.SiteContext) sitedb.CreateTempTargetParams

	// goverter:context siteContext
	// goverter:map DurationMinutes | Int64OrNil
	// goverter:map HighTarget HighTargetMgdl | SettingsUnitsToMgdlOrNil
	// goverter:map LowTarget LowTargetMgdl | SettingsUnitsToMgdlOrNil
	// goverter:map UpdatedByIdentityID | ContextIdentityID
	// goverter:map UpdatedByIdentityName | ContextIdentityName
	// goverter:ignore DurationType
	TempTargetToUpdate(source *SiteTempTarget, siteContext *httputils.SiteContext) sitedb.UpdateTempTargetParams

	///////
	// TEMP BASALS
	///////

	// App's text is shown without its tags
	// goverter:map Notes | StripTags
	// goverter:map Reason | StripTags
	TempBasalFromDB(source *sitedb.TempBasal) *SiteTempBasal

	// goverter:context siteContext
	// goverter:map ID | NewSiteDataID
	// goverter:map OccurredAt | OccurredAtOrNow
	// goverter:map CreatedByIdentityID | ContextIdentityID
	// goverter:map CreatedByIdentityName | ContextIdentityNameValue
	// The site does not set AAPS's type of temp basal, nor Loop's insulin, nor whether the loop set it
	// goverter:ignore Type InsulinType Automatic
	TempBasalToCreate(source *SiteTempBasal, siteContext *httputils.SiteContext) sitedb.CreateTempBasalParams

	// goverter:context siteContext
	// goverter:map DurationMinutes | Int64OrNil
	// goverter:map UpdatedByIdentityID | ContextIdentityID
	// goverter:map UpdatedByIdentityName | ContextIdentityName
	// goverter:ignore Type InsulinType Automatic
	TempBasalToUpdate(source *SiteTempBasal, siteContext *httputils.SiteContext) sitedb.UpdateTempBasalParams

	///////
	// ACTIVITIES
	///////

	// App's text is shown without its tags
	// goverter:map Notes | StripTags
	// goverter:map Reason | StripTags
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
	ActivityToUpdate(source *SiteActivity, siteContext *httputils.SiteContext) sitedb.UpdateActivityParams

	///////
	// NOTIFICATION EVENTS
	///////

	// App's text is shown without its tags
	// goverter:map Notes | StripTags
	// goverter:map Reason | StripTags
	// goverter:map Content | StripTags
	// goverter:map Title | StripTagsValue
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
	NotificationEventToUpdate(source *SiteNotificationEvent, siteContext *httputils.SiteContext) sitedb.UpdateNotificationEventParams

	///////
	// PROFILE SWITCHES
	///////

	// App's text is shown without its tags
	// goverter:map Notes | StripTags
	// goverter:map Reason | StripTags
	ProfileSwitchFromDB(source *sitedb.ProfileSwitch) *SiteProfileSwitch

	// goverter:context siteContext
	// goverter:map ID | NewSiteDataID
	// goverter:map OccurredAt | OccurredAtOrNow
	// goverter:map CreatedByIdentityID | ContextIdentityID
	// goverter:map CreatedByIdentityName | ContextIdentityNameValue
	// goverter:ignore Dia Units Timezone CarbsHr Delay CarbRatiosJson InsulinSensitivitiesJson BasalRatesJson TargetsHighJson TargetsLowJson
	// The site does not set what AAPS's switch was made from, nor the insulin it was made with
	// goverter:ignore OriginalProfileName OriginalCustomizedName OriginalTimeshiftMs OriginalPercentage OriginalDurationMs OriginalEnd IcfgJson
	ProfileSwitchToCreate(source *SiteProfileSwitch, siteContext *httputils.SiteContext) sitedb.CreateProfileSwitchParams

	// goverter:context siteContext
	// goverter:map Name | StringOrNil
	// goverter:map UpdatedByIdentityID | ContextIdentityID
	// goverter:map UpdatedByIdentityName | ContextIdentityName
	// goverter:ignore OriginalProfileName OriginalCustomizedName OriginalTimeshiftMs OriginalPercentage OriginalDurationMs OriginalEnd IcfgJson
	ProfileSwitchToUpdate(source *SiteProfileSwitch, siteContext *httputils.SiteContext) sitedb.UpdateProfileSwitchParams

	///////
	// EVENT GROUPS
	///////

	// goverter:context siteContext
	// goverter:autoMap EventGroup
	// goverter:map EventGroup.ProfileSwitchID ProfileSwitchID
	// goverter:map EventGroup.ActivityID ActivityID
	// goverter:map EventGroup.NotificationEventID NotificationEventID
	// goverter:map EventGroup.GlucoseReadingID GlucoseReadingID
	// goverter:map EventGroup.InsulinDoseID InsulinDoseID
	// goverter:map EventGroup.FoodIntakeID FoodIntakeID
	// goverter:map EventGroup.TempBasalID TempBasalID
	// goverter:map EventGroup.TempTargetID TempTargetID
	// App's text is shown without its tags
	// goverter:map EventGroup.Notes Notes | StripTags
	// goverter:map EventGroup.Reason Reason | StripTags
	EventGroupFromDB(source *sitedb.EventGroupsBundle, siteContext *httputils.SiteContext) *SiteEventGroup

	// goverter:context siteContext
	// goverter:map ID | NewSiteDataID
	// goverter:map OccurredAt | OccurredAtOrNow
	// goverter:map CreatedByIdentityID | ContextIdentityID
	// goverter:map CreatedByIdentityName | ContextIdentityNameValue
	// The site does not set AAPS's running mode, nor the pump, nor its insulin, nor a Bolus Wizard's calculation
	// goverter:ignore EnteredBy Uuid Mode AutoForced Reasons OriginalDurationMs DurationMs SentHash PumpID PumpSerial PumpType BolusCalculatorResultJson
	// nor Loop's ids, nor iAPS's carb entry time and id
	// goverter:ignore SyncIdentifier CreationDate FpuID
	EventGroupToCreate(source *SiteEventGroup, siteContext *httputils.SiteContext) sitedb.CreateEventGroupParams

	// goverter:context siteContext
	// goverter:map GroupType | StringOrNil
	// goverter:map OccurredAt | Int64OrNil
	// goverter:map UpdatedByIdentityID | ContextIdentityID
	// goverter:map UpdatedByIdentityName | ContextIdentityName
	EventGroupToRegroup(source *SiteEventGroup, siteContext *httputils.SiteContext) sitedb.RegroupEventGroupParams

	///////
	// INSULIN DOSES
	///////

	// App's text is shown without its tags
	// goverter:map Notes | StripTags
	// goverter:map Reason | StripTags
	InsulinDoseFromDB(source *sitedb.InsulinDose) *SiteInsulinDose

	// goverter:context siteContext
	// goverter:map ID | NewSiteDataID
	// goverter:map OccurredAt | OccurredAtOrNow
	// goverter:map CreatedByIdentityID | ContextIdentityID
	// goverter:map CreatedByIdentityName | ContextIdentityNameValue
	// Only a loop's upload says whether a dose was automatic or basal insulin, its insulin, how much is not absorbed and
	// the insulin it was given with
	// goverter:ignore Automatic IsBasalInsulin InsulinType Unabsorbed IcfgJson
	InsulinDoseToCreate(source *SiteInsulinDose, siteContext *httputils.SiteContext) sitedb.CreateInsulinDoseParams

	// goverter:context siteContext
	// goverter:map Dose | Float64OrNil
	// goverter:map UpdatedByIdentityID | ContextIdentityID
	// goverter:map UpdatedByIdentityName | ContextIdentityName
	// Only a loop's upload says whether a dose was automatic or basal insulin, its insulin, how much is not absorbed and
	// the insulin it was given with
	// goverter:ignore Automatic IsBasalInsulin InsulinType Unabsorbed IcfgJson
	InsulinDoseToUpdate(source *SiteInsulinDose, siteContext *httputils.SiteContext) sitedb.UpdateInsulinDoseParams

	///////
	// FOOD INTAKES
	///////

	// App's text is shown without its tags
	// goverter:map Notes | StripTags
	// goverter:map Reason | StripTags
	FoodIntakeFromDB(source *sitedb.FoodIntake) *SiteFoodIntake

	// goverter:context siteContext
	// goverter:map ID | NewSiteDataID
	// goverter:map OccurredAt | OccurredAtOrNow
	// goverter:map CreatedByIdentityID | ContextIdentityID
	// goverter:map CreatedByIdentityName | ContextIdentityNameValue
	// Only Loop's upload says when carbs were entered, not the site
	// goverter:ignore UserEnteredAt
	FoodIntakeToCreate(source *SiteFoodIntake, siteContext *httputils.SiteContext) sitedb.CreateFoodIntakeParams

	// goverter:context siteContext
	// goverter:map UpdatedByIdentityID | ContextIdentityID
	// goverter:map UpdatedByIdentityName | ContextIdentityName
	// goverter:ignore UserEnteredAt
	FoodIntakeToUpdate(source *SiteFoodIntake, siteContext *httputils.SiteContext) sitedb.UpdateFoodIntakeParams

	///////
	// FOOD PRESETS
	///////

	FoodPresetFromDB(source *sitedb.FoodPreset) *SiteFoodPreset

	// goverter:context siteContext
	// goverter:map ID | NewSiteDataID
	// goverter:map CreatedByIdentityID | ContextIdentityID
	// goverter:map CreatedByIdentityName | ContextIdentityNameValue
	FoodPresetToCreate(source *SiteFoodPreset, siteContext *httputils.SiteContext) sitedb.CreateFoodPresetParams

	// goverter:context siteContext
	// goverter:map UpdatedByIdentityID | ContextIdentityID
	// goverter:map UpdatedByIdentityName | ContextIdentityName
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
	ProfilePresetToUpdate(source *SiteProfilePreset, siteContext *httputils.SiteContext) sitedb.UpdateProfilePresetParams

	///////
	// DEVICE STATUSES
	///////

	DeviceStatusFromDB(source *sitedb.DeviceReport) *SiteDeviceStatus

	///////
	// SETTINGS
	///////
}

func SetPasswordGarbage() *string {
	return new(consts.PASSWORD_GARBAGE_VALUE)
}

// DecryptUrlOrUsername is the connector's URL or username, decrypted for its site, or nil when it has none
//
// goverter:context sec
func DecryptUrlOrUsername(source *controldb.SiteConnector, sec *security.SecurityConfig) (*string, error) {
	if source.UrlOrUsernameEnc == nil {
		return nil, nil
	}
	decrypted, err := sec.DecryptFor(security.PurposeConnector, source.SiteID, *source.UrlOrUsernameEnc)
	if err != nil {
		return nil, err
	}
	return &decrypted, nil
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
func ContextSiteIDPtr(siteContext *httputils.SiteContext) *string {
	return &siteContext.SiteID
}

// JSONStrings is a JSON list of strings, empty when it cannot be read
func JSONStrings(list string) []string {
	strings := []string{}
	_ = json.Unmarshal([]byte(list), &strings)
	return strings
}

// JSONStringsPtr is JSONStrings, nil when there is no list
func JSONStringsPtr(list *string) []string {
	if list == nil {
		return nil
	}
	return JSONStrings(*list)
}

// StringsJSON is strings as a JSON list
func StringsJSON(strings []string) string {
	if strings == nil {
		strings = []string{}
	}
	b, _ := json.Marshal(strings)
	return string(b)
}

// StringsJSONPtr is StringsJSON, nil when there are none
func StringsJSONPtr(strings []string) *string {
	if len(strings) == 0 {
		return nil
	}
	return new(StringsJSON(strings))
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
