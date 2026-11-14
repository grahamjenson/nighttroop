package nightscoutmodels

import (
	"encoding/json"
	"time"

	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/httputils"
	"nighttroop.t1d.nz/utils/types"
)

// goverter:converter
// goverter:output:format function
// goverter:enum:unknown @ignore
// goverter:output:file ./z_treatments_converters.go
// goverter:useZeroValueOnPointerInconsistency
// goverter:extend FlexibleFloat64ToFloat64
// goverter:extend FlexibleInt64ToInt64
// goverter:extend ShortStringToString
// goverter:extend SanitizedStringToString
// goverter:extend ClientIDToString
// goverter:extend UnitsToUnits
type TreatmentsEvenetConverter interface {

	///////
	// CREATE
	///////

	// goverter:context ctx
	// goverter:map CreatedByIdentityName  | AssignCreatedByName
	// goverter:map CreatedByIdentityID | AssignCreatedByID
	// goverter:map ID                     | AssignID
	EventGroupCreate(source sitedb.EventGroup, ctx ConverterContext) sitedb.CreateEventGroupParams

	///////
	// Food Intake
	///////

	// goverter:context ctx
	// goverter:map CreatedByIdentityName  | AssignCreatedByName
	// goverter:map CreatedByIdentityID | AssignCreatedByID
	// goverter:map ID                     | AssignID
	FoodIntakeCreate(source sitedb.FoodIntake, ctx ConverterContext) sitedb.CreateFoodIntakeParams

	///////
	// Glucose Reading
	///////

	// goverter:context ctx
	// goverter:map CreatedByIdentityName  | AssignCreatedByName
	// goverter:map CreatedByIdentityID | AssignCreatedByID
	// goverter:map ID                     | AssignID
	GlucoseReadingCreate(source sitedb.GlucoseReading, ctx ConverterContext) sitedb.CreateGlucoseReadingParams

	///////
	// Insulin Dose
	///////

	// goverter:context ctx
	// goverter:map CreatedByIdentityName  | AssignCreatedByName
	// goverter:map CreatedByIdentityID | AssignCreatedByID
	// goverter:map ID                     | AssignID
	InsulinDoseCreate(source sitedb.InsulinDose, ctx ConverterContext) sitedb.CreateInsulinDoseParams

	///////
	// Notes
	///////

	// goverter:context ctx
	// goverter:map CreatedByIdentityName  | AssignCreatedByName
	// goverter:map CreatedByIdentityID | AssignCreatedByID
	// goverter:map ID                     | AssignID
	NotificationEventCreate(source sitedb.NotificationEvent, ctx ConverterContext) sitedb.CreateNotificationEventParams

	///////
	// Profile Switch
	///////

	// goverter:context ctx
	// goverter:map CreatedByIdentityName  | AssignCreatedByName
	// goverter:map CreatedByIdentityID | AssignCreatedByID
	// goverter:map ID                     | AssignID
	ProfileSwitchCreate(source sitedb.ProfileSwitch, ctx ConverterContext) sitedb.CreateProfileSwitchParams

	///////
	// Temp Basal
	///////

	// goverter:context ctx
	// goverter:map CreatedByIdentityName  | AssignCreatedByName
	// goverter:map CreatedByIdentityID | AssignCreatedByID
	// goverter:map ID                     | AssignID
	TempBasalCreate(source sitedb.TempBasal, ctx ConverterContext) sitedb.CreateTempBasalParams

	///////
	// Temp Targets
	///////

	// goverter:context ctx
	// goverter:map CreatedByIdentityName  | AssignCreatedByName
	// goverter:map CreatedByIdentityID | AssignCreatedByID
	// goverter:map ID                     | AssignID
	TempTargetCreate(source sitedb.TempTarget, ctx ConverterContext) sitedb.CreateTempTargetParams

	// goverter:context ctx
	// goverter:map CreatedByIdentityName  | AssignCreatedByName
	// goverter:map CreatedByIdentityID | AssignCreatedByID
	// goverter:map ID                     | AssignID
	ActivityCreate(source sitedb.Activity, ctx ConverterContext) sitedb.CreateActivityParams

	///////
	// EventGroup
	///////

	// goverter:ignoreMissing
	//
	// goverter:map ID Identifier
	// goverter:map GroupType EventType
	// goverter:map CreatedByIdentityName Subject
	// goverter:map . EnteredBy | EventGroupEnteredBy
	// goverter:map Uuid UUID
	// goverter:map BolusCalculatorResultJson BolusCalculatorResult
	// goverter:map CreatedByAppName App
	// goverter:map CreatedByDeviceName Device
	// goverter:map OccurredAt Date
	// goverter:map DurationMinutes Duration | MinutesToDuration
	// goverter:map ClientCreatedAt CreationDate | MillisToISODateString
	EventGroupToTreatment(source sitedb.EventGroup) *Treatment

	// goverter:context siteContext
	// goverter:update target
	//
	// goverter:ignoreMissing
	// goverter:map Mgdl Glucose    | GlucoseReadingTypeToValue
	// goverter:map Mgdl Units      | UnitsToValue
	// goverter:map MeasurementMethod GlucoseType
	UpdateGlucoseReadingToTreatment(source *sitedb.GlucoseReading, siteContext *httputils.SiteContext, target *Treatment)

	// goverter:update target
	// goverter:ignoreMissing
	//
	// goverter:map Dose Insulin
	// goverter:map DoseType Type
	// goverter:map IcfgJson ICfg | JSONToInsulinConfig
	UpdateInsulinDoseToTreatment(source *sitedb.InsulinDose, target *Treatment)

	// goverter:update target
	// goverter:ignoreMissing
	//
	// goverter:map UserEnteredAt UserEnteredAt | MillisToDateString
	UpdateFoodIntakeToTreatment(source *sitedb.FoodIntake, target *Treatment)

	// goverter:update target
	// goverter:ignoreMissing
	//
	// goverter:map Rate Absolute
	// goverter:map DurationMinutes Duration | MinutesToDuration
	// goverter:ignore Percent
	// goverter:map Multiplier TempBasalPercent | MultiplierToPercent
	UpdateTempBasalToTreatment(source *sitedb.TempBasal, target *Treatment)

	// goverter:update target
	// goverter:ignoreMissing
	//
	// goverter:map . CorrectionRange | ToCorrectionRange
	// goverter:map DurationMinutes Duration | MinutesToDuration
	// goverter:map HighTargetMgdl HighTarget
	// goverter:map LowTargetMgdl LowTarget
	UpdateTempTargetToTreatment(source *sitedb.TempTarget, target *Treatment)

	// goverter:update target
	// goverter:ignoreMissing
	//
	// goverter:map Title Name
	UpdateNotificationEventToTreatment(source *sitedb.NotificationEvent, target *Treatment)

	// goverter:context siteContext
	// goverter:update target
	// The treatment's units are its glucose's: the profile's are in its profileJson
	// goverter:ignore Units
	// goverter:ignoreMissing
	//
	// goverter:map Name Profile
	// goverter:map . ProfileJSON | ToProfileJSON
	// goverter:map DurationMinutes Duration | MinutesToDuration
	// goverter:map Multiplier Percent | MultiplierToPercentage
	// goverter:map IcfgJson ICfg | JSONToInsulinConfig
	UpdateProfileSwitchToTreatment(source *sitedb.ProfileSwitch, siteContext *httputils.SiteContext, target *Treatment)

	// goverter:update target
	// goverter:ignoreMissing
	//
	// goverter:map DurationMinutes Duration | MinutesToDuration
	UpdateActivityToTreatment(source *sitedb.Activity, target *Treatment)

	//////
	// Partial Treatment Converters
	//////

	// goverter:ignore ID DeletedAt CreatedAt CreatedByIdentityName CreatedByIdentityID UpdatedAt UpdatedByIdentityName UpdatedByIdentityID
	// goverter:map Date OccurredAt | OrNow
	// goverter:map App CreatedByAppName
	// goverter:map Device CreatedByDeviceName
	// goverter:map UserEnteredAt UserEnteredAt | DateStringToMillis
	PartialTreatmentToFoodIntake(source Treatment) sitedb.FoodIntake

	// goverter:ignore ID DeletedAt CreatedAt CreatedByIdentityName CreatedByIdentityID UpdatedAt UpdatedByIdentityName UpdatedByIdentityID
	// goverter:map Date OccurredAt | OrNow
	// goverter:map App CreatedByAppName
	// goverter:map Device CreatedByDeviceName
	//
	// goverter:ignore Mgdl Delta Trend ReadingType Noise Filtered Unfiltered Rssi IsCalibration SensorMgdlAtPairing
	// goverter:map GlucoseType MeasurementMethod
	PartialTreatmentToGlucoseReading(source Treatment) sitedb.GlucoseReading

	// goverter:ignore ID DeletedAt CreatedAt CreatedByIdentityName CreatedByIdentityID UpdatedAt UpdatedByIdentityName UpdatedByIdentityID
	// goverter:map Date OccurredAt | OrNow
	// goverter:map App CreatedByAppName
	// goverter:map Device CreatedByDeviceName
	//
	// goverter:map Insulin Dose
	// goverter:map Type DoseType
	// goverter:map ICfg IcfgJson | InsulinConfigToJSON
	PartialTreatmentToInsulinDose(source Treatment) sitedb.InsulinDose

	// goverter:ignore ID DeletedAt CreatedAt CreatedByIdentityName CreatedByIdentityID UpdatedAt UpdatedByIdentityName UpdatedByIdentityID
	// goverter:map Date OccurredAt | OrNow
	// goverter:map App CreatedByAppName
	// goverter:map Device CreatedByDeviceName
	//
	// goverter:ignore NotificationType Title Severity
	// goverter:map Notes Content
	PartialTreatmentToNotificationEvent(source Treatment) sitedb.NotificationEvent

	// goverter:ignore ID DeletedAt CreatedAt CreatedByIdentityName CreatedByIdentityID UpdatedAt UpdatedByIdentityName UpdatedByIdentityID
	// goverter:map Date OccurredAt | OrNow
	// goverter:map App CreatedByAppName
	// goverter:map Device CreatedByDeviceName
	//
	// goverter:ignore Dia Units Timezone CarbsHr Delay CarbRatiosJson InsulinSensitivitiesJson  BasalRatesJson TargetsHighJson TargetsLowJson
	// goverter:map Percent Multiplier | PercentageToMultiplier
	// goverter:map ICfg IcfgJson | InsulinConfigToJSON
	// goverter:map Duration DurationMinutes | DurationToMinutes
	PartialTreatmentToProfileSwitch(source Treatment) sitedb.ProfileSwitch

	// goverter:ignore ID DeletedAt CreatedAt CreatedByIdentityName CreatedByIdentityID UpdatedAt UpdatedByIdentityName UpdatedByIdentityID
	// goverter:map Date OccurredAt | OrNow
	// goverter:map App CreatedByAppName
	// goverter:map Device CreatedByDeviceName
	// goverter:map TempBasalPercent Multiplier | PercentToMultiplier
	// goverter:map Duration DurationMinutes | DurationToMinutes
	PartialTreatmentToTempBasal(source Treatment) sitedb.TempBasal

	// goverter:ignore ID DeletedAt CreatedAt CreatedByIdentityName CreatedByIdentityID UpdatedAt UpdatedByIdentityName UpdatedByIdentityID
	// goverter:map Date OccurredAt | OrNow
	// goverter:map App CreatedByAppName
	// goverter:map Device CreatedByDeviceName
	// goverter:map HighTarget HighTargetMgdl
	// goverter:map LowTarget LowTargetMgdl
	// goverter:map Duration DurationMinutes | DurationToMinutes
	PartialTreatmentToTempTarget(source Treatment) sitedb.TempTarget

	// goverter:ignore ID DeletedAt CreatedAt CreatedByIdentityName CreatedByIdentityID UpdatedAt UpdatedByIdentityName UpdatedByIdentityID
	// goverter:map Date OccurredAt | OrNow
	// goverter:map App CreatedByAppName
	// goverter:map Device CreatedByDeviceName
	// goverter:map Duration DurationMinutes | DurationToMinutesP
	PartialTreatmentToActivity(source Treatment) sitedb.Activity

	// goverter:ignore ID DeletedAt CreatedAt CreatedByIdentityName CreatedByIdentityID UpdatedAt UpdatedByIdentityName UpdatedByIdentityID
	// goverter:map Date OccurredAt | OrNow
	// goverter:map App CreatedByAppName
	// goverter:map Device CreatedByDeviceName
	//
	// goverter:ignore GlucoseReadingID InsulinDoseID FoodIntakeID TempBasalID TempTargetID ProfileSwitchID ActivityID NotificationEventID
	// goverter:map EventType GroupType
	// goverter:map UUID Uuid
	// goverter:map Duration DurationMinutes | DurationToMinutesP
	// goverter:map Duration DurationMs | DurationToMs
	// goverter:map BolusCalculatorResult BolusCalculatorResultJson
	// goverter:map SentJSON SentHash | SentHash
	// goverter:map CreationDate ClientCreatedAt | DateStringToMillis
	TreatmentToEventGroup(source Treatment) sitedb.EventGroup
}

// MinutesToDuration is a duration of the minutes an event keeps, nil when it has none
func MinutesToDuration(durationMinutes any) *types.Duration {
	switch v := durationMinutes.(type) {
	case int64:
		return types.DurationOfMinutes(&v)
	case *int64:
		return types.DurationOfMinutes(v)
	}
	return nil
}

// DurationToMinutes are the whole minutes of a duration an event keeps, 0 when there is none
func DurationToMinutes(source *types.Duration) int64 {
	return source.MinutesOrZero()
}

// DurationToMinutesP are the whole minutes of a duration, nil when there is none
func DurationToMinutesP(source *types.Duration) *int64 {
	return source.MinutesP()
}

// DurationToMs are the milliseconds a duration keeps, nil when they are its minutes
func DurationToMs(source *types.Duration) *int64 {
	return source.MsP()
}

// goverter:context siteContext
func ToProfileJSON(source interface {
	GetDia() *float64
	GetTimezone() *string
	GetUnits() *consts.Units
	GetCarbRatiosJson() *string
	GetInsulinSensitivitiesJson() *string
	GetBasalRatesJson() *string
	GetTargetsLowJson() *string
	GetTargetsHighJson() *string
	GetIcfgJson() *string
}, siteContext *httputils.SiteContext) *types.LongJSONString {
	// A switch sent with no profileJson, as the careportal and T1DCarer send one, names a profile the app has, so it
	// gets none back: AAPS uses its own profile of that name only when there is no profileJson, and drops the switch
	// when the profileJson has no schedules
	if source.GetCarbRatiosJson() == nil && source.GetInsulinSensitivitiesJson() == nil && source.GetBasalRatesJson() == nil &&
		source.GetTargetsLowJson() == nil && source.GetTargetsHighJson() == nil {
		return nil
	}
	profile := Profile{}

	if source.GetDia() != nil {
		profile.Dia = new(types.FlexibleFloat64(*source.GetDia()))
	}
	if source.GetTimezone() != nil {
		profile.Timezone = new(types.Timezone(*source.GetTimezone()))
	}

	profile.Units = siteContext.GetSettings().GetUnits()

	if source.GetCarbRatiosJson() != nil {
		profile.CarbRatio = JSONToScheduleValueSlice(source.GetCarbRatiosJson())
	}
	if source.GetInsulinSensitivitiesJson() != nil {
		profile.Sens = SensitivitiesFromJSON(source.GetInsulinSensitivitiesJson(), source.GetUnits(), siteContext)
	}
	if source.GetBasalRatesJson() != nil {
		profile.Basal = JSONToScheduleValueSlice(source.GetBasalRatesJson())
	}
	if source.GetTargetsLowJson() != nil {
		profile.TargetLow = JSONToScheduleValueSliceWithUnitConversionFromMGDL(source.GetTargetsLowJson(), siteContext)
	}
	if source.GetTargetsHighJson() != nil {
		profile.TargetHigh = JSONToScheduleValueSliceWithUnitConversionFromMGDL(source.GetTargetsHighJson(), siteContext)
	}

	profileBytes, err := json.Marshal(profile)
	if err != nil {
		return nil
	}
	// The insulin the switch was made with goes in as AAPS's iCfg, as an AAPSClient follower reads it from there
	// (pureProfileFromJson), after the rest so the profile's own fields keep their order
	if icfg := source.GetIcfgJson(); icfg != nil && len(profileBytes) > 2 {
		profileBytes = append(append(profileBytes[:len(profileBytes)-1], `,"iCfg":`...), *icfg+"}"...)
	}
	profileJSON := types.LongJSONString(string(profileBytes))
	return &profileJSON
}

func ToCorrectionRange(source interface {
	GetLowTargetMgdl() float64
	GetHighTargetMgdl() float64
}) []*types.FlexibleFloat64 {

	correctionRange := []*types.FlexibleFloat64{
		new(types.FlexibleFloat64(source.GetLowTargetMgdl())),
		new(types.FlexibleFloat64(source.GetHighTargetMgdl())),
	}
	return correctionRange
}

// EventGroupEnteredBy is who the app said entered a treatment, as clients look for their own by it (xDrip+ one starting
// xdrip, Trio Trio, iAPS remote), else the app that sent it, else the identity that uploaded it
func EventGroupEnteredBy(source sitedb.EventGroup) *types.ShortString {
	if source.EnteredBy != nil {
		return (*types.ShortString)(source.EnteredBy)
	}
	// One sent without enteredBy is by its app, as AAPS names it, before the device that uploaded it
	if source.CreatedByAppName != nil && *source.CreatedByAppName != "" {
		return (*types.ShortString)(source.CreatedByAppName)
	}
	return (*types.ShortString)(&source.CreatedByIdentityName)
}

// InsulinConfigToJSON is the insulin config as it is stored, with only its fields
func InsulinConfigToJSON(source *InsulinConfig) *string {
	if source == nil {
		return nil
	}
	raw, _ := json.Marshal(source)
	return new(string(raw))
}

// JSONToInsulinConfig is the stored insulin config, nil when there is none
func JSONToInsulinConfig(source *string) *InsulinConfig {
	var config InsulinConfig
	if source == nil || json.Unmarshal([]byte(*source), &config) != nil {
		return nil
	}
	return &config
}

// DateStringToMillis is a date in milliseconds, nil when there is none
func DateStringToMillis(source *types.DateString) *int64 {
	if source == nil {
		return nil
	}
	return source.UnixMillip()
}

// MillisToISODateString is a date in milliseconds as toISOString writes it, as iAPS sent it, nil when there is none
func MillisToISODateString(source *int64) *types.DateString {
	if source == nil {
		return nil
	}
	return new(types.DateString(types.ISOString(time.UnixMilli(*source))))
}

// MillisToDateString is a date in milliseconds as a UTC date string, nil when there is none
func MillisToDateString(source *int64) *types.DateString {
	if source == nil {
		return nil
	}
	return new(types.DateStringFromInt64AndUTCOffset(*source, nil))
}
