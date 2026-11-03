package nightscoutmodels

import (
	"encoding/json"

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
	ComponentEventCreate(source sitedb.ComponentEvent, ctx ConverterContext) sitedb.CreateComponentEventParams

	// goverter:context ctx
	// goverter:map CreatedByIdentityName  | AssignCreatedByName
	// goverter:map CreatedByIdentityID | AssignCreatedByID
	// goverter:map ID                     | AssignID
	ActivityCreate(source sitedb.Activity, ctx ConverterContext) sitedb.CreateActivityParams

	///////
	// Update
	///////

	// goverter:context ctx
	// goverter:map UpdatedByIdentityName | AssignUpdatedByName
	// goverter:map UpdatedByIdentityID | AssignUpdatedByID
	EventGroupUpdate(source sitedb.EventGroup, ctx ConverterContext) sitedb.UpdateEventGroupParams

	///////
	// Food Intake
	///////

	// goverter:context ctx
	// goverter:map UpdatedByIdentityName | AssignUpdatedByName
	// goverter:map UpdatedByIdentityID | AssignUpdatedByID
	FoodIntakeUpdate(source sitedb.FoodIntake, ctx ConverterContext) sitedb.UpdateFoodIntakeParams

	///////
	// Glucose Reading
	///////

	// goverter:context ctx
	// goverter:map UpdatedByIdentityName | AssignUpdatedByName
	// goverter:map UpdatedByIdentityID | AssignUpdatedByID
	GlucoseReadingUpdate(source sitedb.GlucoseReading, ctx ConverterContext) sitedb.UpdateGlucoseReadingParams

	///////
	// Insulin Dose
	///////

	// goverter:context ctx
	// goverter:map UpdatedByIdentityName | AssignUpdatedByName
	// goverter:map UpdatedByIdentityID | AssignUpdatedByID
	InsulinDoseUpdate(source sitedb.InsulinDose, ctx ConverterContext) sitedb.UpdateInsulinDoseParams

	///////
	// Notes
	///////

	// goverter:context ctx
	// goverter:map UpdatedByIdentityName | AssignUpdatedByName
	// goverter:map UpdatedByIdentityID | AssignUpdatedByID
	NotificationEventUpdate(source sitedb.NotificationEvent, ctx ConverterContext) sitedb.UpdateNotificationEventParams

	///////
	// Profile
	///////

	// goverter:context ctx
	// goverter:map UpdatedByIdentityName | AssignUpdatedByName
	// goverter:map UpdatedByIdentityID | AssignUpdatedByID
	ProfileUpdate(source sitedb.ProfilePreset, ctx ConverterContext) sitedb.UpdateProfilePresetParams

	///////
	// Profile Switch
	///////

	// goverter:context ctx
	// goverter:map UpdatedByIdentityName | AssignUpdatedByName
	// goverter:map UpdatedByIdentityID | AssignUpdatedByID
	ProfileSwitchUpdate(source sitedb.ProfileSwitch, ctx ConverterContext) sitedb.UpdateProfileSwitchParams

	///////
	// Temp Basal
	///////

	// goverter:context ctx
	// goverter:map UpdatedByIdentityName | AssignUpdatedByName
	// goverter:map UpdatedByIdentityID | AssignUpdatedByID
	TempBasalUpdate(source sitedb.TempBasal, ctx ConverterContext) sitedb.UpdateTempBasalParams

	///////
	// Temp Targets
	///////

	// goverter:context ctx
	// goverter:map UpdatedByIdentityName | AssignUpdatedByName
	// goverter:map UpdatedByIdentityID | AssignUpdatedByID
	TempTargetUpdate(source sitedb.TempTarget, ctx ConverterContext) sitedb.UpdateTempTargetParams

	// goverter:context ctx
	// goverter:map UpdatedByIdentityName | AssignUpdatedByName
	// goverter:map UpdatedByIdentityID | AssignUpdatedByID
	ComponentEventUpdate(source sitedb.ComponentEvent, ctx ConverterContext) sitedb.UpdateComponentEventParams

	// goverter:context ctx
	// goverter:map UpdatedByIdentityName | AssignUpdatedByName
	// goverter:map UpdatedByIdentityID | AssignUpdatedByID
	ActivityUpdate(source sitedb.Activity, ctx ConverterContext) sitedb.UpdateActivityParams

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
	// goverter:map CreatedByAppName App
	// goverter:map CreatedByDeviceName Device
	// goverter:map OccurredAt Date
	// goverter:map . CreatedAt | ModelDateStringFromDateIntAndUTCOffset
	EventGroupToTreatment(source sitedb.EventGroup) *Treatment

	// goverter:context siteContext
	// goverter:update target
	// goverter:ignore CreatedAt
	//
	// goverter:ignoreMissing
	// goverter:map Mgdl Glucose    | GlucoseReadingTypeToValue
	// goverter:map Mgdl Units      | UnitsToValue
	// goverter:map MeasurementMethod GlucoseType
	UpdateGlucoseReadingToTreatment(source *sitedb.GlucoseReading, siteContext *httputils.SiteContext, target *Treatment)

	// goverter:update target
	// goverter:ignore CreatedAt
	// goverter:ignoreMissing
	//
	// goverter:map Dose Insulin
	// goverter:map DoseType Type
	UpdateInsulinDoseToTreatment(source *sitedb.InsulinDose, target *Treatment)

	// goverter:update target
	// goverter:ignore CreatedAt
	// goverter:ignoreMissing
	//
	UpdateFoodIntakeToTreatment(source *sitedb.FoodIntake, target *Treatment)

	// goverter:update target
	// goverter:ignore CreatedAt
	// goverter:ignoreMissing
	//
	// goverter:map Absolute Rate
	// goverter:map DurationMinutes DurationInMilliseconds | MinutesToMilliseconds
	// goverter:ignore Percent
	// goverter:map Percent TempBasalPercent
	UpdateTempBasalToTreatment(source *sitedb.TempBasal, target *Treatment)

	// goverter:update target
	// goverter:ignore CreatedAt
	// goverter:ignoreMissing
	//
	// goverter:map . CorrectionRange | ToCorrectionRange
	// goverter:map DurationMinutes DurationInMilliseconds | MinutesToMilliseconds
	// goverter:map HighTargetMgdl HighTarget
	// goverter:map LowTargetMgdl LowTarget
	UpdateTempTargetToTreatment(source *sitedb.TempTarget, target *Treatment)

	// goverter:update target
	// goverter:ignore CreatedAt
	// goverter:ignoreMissing
	//
	// goverter:map Title Name
	UpdateNotificationEventToTreatment(source *sitedb.NotificationEvent, target *Treatment)

	// goverter:context siteContext
	// goverter:update target
	// goverter:ignore CreatedAt
	// goverter:ignoreMissing
	//
	// goverter:map Name Profile
	// goverter:map . ProfileJSON | ToProfileJSON
	// goverter:map DurationMinutes DurationInMilliseconds | MinutesToMilliseconds
	UpdateProfileSwitchToTreatment(source *sitedb.ProfileSwitch, siteContext *httputils.SiteContext, target *Treatment)

	// goverter:update target
	// goverter:ignore CreatedAt
	// goverter:ignoreMissing
	//
	UpdateComponentEventToTreatment(source *sitedb.ComponentEvent, target *Treatment)

	// goverter:update target
	// goverter:ignore CreatedAt
	// goverter:ignoreMissing
	//
	// goverter:map DurationMinutes DurationInMilliseconds | MinutesToMilliseconds
	UpdateActivityToTreatment(source *sitedb.Activity, target *Treatment)

	//////
	// Update Override
	//////

	// goverter:update target
	// goverter:ignore ID
	UpdateToEventGroup(source sitedb.EventGroup, target *sitedb.EventGroup)

	// goverter:update target
	// goverter:ignore ID
	UpdateToFoodIntake(source *sitedb.FoodIntake, target *sitedb.FoodIntake)

	// goverter:update target
	// goverter:ignore ID
	UpdateToGlucoseReading(source *sitedb.GlucoseReading, target *sitedb.GlucoseReading)

	// goverter:update target
	// goverter:ignore ID
	UpdateToInsulinDose(source *sitedb.InsulinDose, target *sitedb.InsulinDose)

	// goverter:update target
	// goverter:ignore ID
	UpdateToNotificationEvent(source *sitedb.NotificationEvent, target *sitedb.NotificationEvent)

	// goverter:update target
	// goverter:ignore ID
	UpdateToProfileSwitch(source *sitedb.ProfileSwitch, target *sitedb.ProfileSwitch)

	// goverter:update target
	// goverter:ignore ID
	UpdateToTempBasal(source *sitedb.TempBasal, target *sitedb.TempBasal)

	// goverter:update target
	// goverter:ignore ID
	UpdateToTempTarget(source *sitedb.TempTarget, target *sitedb.TempTarget)

	// goverter:update target
	// goverter:ignore ID
	UpdateToComponentEvent(source *sitedb.ComponentEvent, target *sitedb.ComponentEvent)

	// goverter:update target
	// goverter:ignore ID
	UpdateToActivity(source *sitedb.Activity, target *sitedb.Activity)

	//////
	// Partial Treatment Converters
	//////

	// goverter:ignore ID DeletedAt CreatedAt CreatedByIdentityName CreatedByIdentityID UpdatedAt UpdatedByIdentityName UpdatedByIdentityID
	// goverter:map Date OccurredAt | OrNow
	// goverter:map App CreatedByAppName
	// goverter:map Device CreatedByDeviceName
	PartialTreatmentToFoodIntake(source Treatment) sitedb.FoodIntake

	// goverter:ignore ID DeletedAt CreatedAt CreatedByIdentityName CreatedByIdentityID UpdatedAt UpdatedByIdentityName UpdatedByIdentityID
	// goverter:map Date OccurredAt | OrNow
	// goverter:map App CreatedByAppName
	// goverter:map Device CreatedByDeviceName
	//
	// goverter:ignore Mgdl Delta Direction Noise Filtered Unfiltered Rssi ReadingType
	// goverter:map GlucoseType MeasurementMethod
	PartialTreatmentToGlucoseReading(source Treatment) sitedb.GlucoseReading

	// goverter:ignore ID DeletedAt CreatedAt CreatedByIdentityName CreatedByIdentityID UpdatedAt UpdatedByIdentityName UpdatedByIdentityID
	// goverter:map Date OccurredAt | OrNow
	// goverter:map App CreatedByAppName
	// goverter:map Device CreatedByDeviceName
	//
	// goverter:map Insulin Dose
	// goverter:map Type DoseType
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
	PartialTreatmentToProfileSwitch(source Treatment) sitedb.ProfileSwitch

	// goverter:ignore ID DeletedAt CreatedAt CreatedByIdentityName CreatedByIdentityID UpdatedAt UpdatedByIdentityName UpdatedByIdentityID
	// goverter:map Date OccurredAt | OrNow
	// goverter:map App CreatedByAppName
	// goverter:map Device CreatedByDeviceName
	// goverter:map TempBasalPercent Percent
	PartialTreatmentToTempBasal(source Treatment) sitedb.TempBasal

	// goverter:ignore ID DeletedAt CreatedAt CreatedByIdentityName CreatedByIdentityID UpdatedAt UpdatedByIdentityName UpdatedByIdentityID
	// goverter:map Date OccurredAt | OrNow
	// goverter:map App CreatedByAppName
	// goverter:map Device CreatedByDeviceName
	// goverter:map HighTarget HighTargetMgdl
	// goverter:map LowTarget LowTargetMgdl
	PartialTreatmentToTempTarget(source Treatment) sitedb.TempTarget

	// goverter:ignore ID DeletedAt CreatedAt CreatedByIdentityName CreatedByIdentityID UpdatedAt UpdatedByIdentityName UpdatedByIdentityID
	// goverter:map Date OccurredAt | OrNow
	// goverter:map App CreatedByAppName
	// goverter:map Device CreatedByDeviceName
	PartialTreatmentToComponentEvent(source Treatment) sitedb.ComponentEvent

	// goverter:ignore ID DeletedAt CreatedAt CreatedByIdentityName CreatedByIdentityID UpdatedAt UpdatedByIdentityName UpdatedByIdentityID
	// goverter:map Date OccurredAt | OrNow
	// goverter:map App CreatedByAppName
	// goverter:map Device CreatedByDeviceName
	PartialTreatmentToActivity(source Treatment) sitedb.Activity

	// goverter:ignore ID DeletedAt CreatedAt CreatedByIdentityName CreatedByIdentityID UpdatedAt UpdatedByIdentityName UpdatedByIdentityID
	// goverter:map Date OccurredAt | OrNow
	// goverter:map App CreatedByAppName
	// goverter:map Device CreatedByDeviceName
	//
	// goverter:ignore GlucoseReadingID InsulinDoseID FoodIntakeID TempBasalID TempTargetID ProfileSwitchID ComponentEventID ActivityID NotificationEventID
	// goverter:map EventType GroupType
	// goverter:map UUID Uuid
	TreatmentToEventGroup(source Treatment) sitedb.EventGroup
}

func MinutesToMilliseconds(durationMinutes any) *types.FlexibleInt64 {
	switch v := durationMinutes.(type) {
	case int64:
		value := types.FlexibleInt64(v * 60 * 1000)
		return &value
	case *int64:
		if v == nil {
			return nil
		}
		value := types.FlexibleInt64(*v * 60 * 1000)
		return &value
	}

	return nil
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
}, siteContext *httputils.SiteContext) *types.LongJSONString {
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
		profile.Sens = JSONToScheduleValueSlice(source.GetInsulinSensitivitiesJson())
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
// xdrip, Trio Trio, iAPS remote), else the identity that uploaded it
func EventGroupEnteredBy(source sitedb.EventGroup) *types.ShortString {
	if source.EnteredBy != nil {
		return (*types.ShortString)(source.EnteredBy)
	}
	return (*types.ShortString)(&source.CreatedByIdentityName)
}
