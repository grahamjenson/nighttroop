package nightscout

import (
	"encoding/json"

	"nighttroop.t1d.nz/data/sql/site"
)

// goverter:converter
// goverter:output:format function
// goverter:output:file ./z_treatments_converters.go
// goverter:useZeroValueOnPointerInconsistency
// goverter:extend FlexibleFloat64ToFloat64
// goverter:extend FlexibleInt64ToInt64
// goverter:extend ShortStringToString
// goverter:extend SanitizedStringToString
type TreatmentsEvenetConverter interface {

	///////
	// CREATE
	///////

	// goverter:context ctx
	// goverter:map CreatedByIdentityName  | AssignCreatedByName
	// goverter:map ID                     | AssignID
	EventGroupCreate(source site.EventGroup, ctx ConverterContext) site.CreateEventGroupParams

	///////
	// Food Intake
	///////

	// goverter:context ctx
	// goverter:map CreatedByIdentityName  | AssignCreatedByName
	// goverter:map ID                     | AssignID
	FoodIntakeCreate(source site.FoodIntake, ctx ConverterContext) site.CreateFoodIntakeParams

	///////
	// Glucose Reading
	///////

	// goverter:context ctx
	// goverter:map CreatedByIdentityName  | AssignCreatedByName
	// goverter:map ID                     | AssignID
	GlucoseReadingCreate(source site.GlucoseReading, ctx ConverterContext) site.CreateGlucoseReadingParams

	///////
	// Insulin Dose
	///////

	// goverter:context ctx
	// goverter:map CreatedByIdentityName  | AssignCreatedByName
	// goverter:map ID                     | AssignID
	InsulinDoseCreate(source site.InsulinDose, ctx ConverterContext) site.CreateInsulinDoseParams

	///////
	// Notes
	///////

	// goverter:context ctx
	// goverter:map CreatedByIdentityName  | AssignCreatedByName
	// goverter:map ID                     | AssignID
	NotificationEventCreate(source site.NotificationEvent, ctx ConverterContext) site.CreateNotificationEventParams

	///////
	// Profile Switch
	///////

	// goverter:context ctx
	// goverter:map CreatedByIdentityName  | AssignCreatedByName
	// goverter:map ID                     | AssignID
	ProfileSwitchCreate(source site.ProfileSwitch, ctx ConverterContext) site.CreateProfileSwitchParams

	///////
	// Temp Basal
	///////

	// goverter:context ctx
	// goverter:map CreatedByIdentityName  | AssignCreatedByName
	// goverter:map ID                     | AssignID
	TempBasalCreate(source site.TempBasal, ctx ConverterContext) site.CreateTempBasalParams

	///////
	// Temp Targets
	///////

	// goverter:context ctx
	// goverter:map CreatedByIdentityName  | AssignCreatedByName
	// goverter:map ID                     | AssignID
	TempTargetCreate(source site.TempTarget, ctx ConverterContext) site.CreateTempTargetParams

	// goverter:context ctx
	// goverter:map CreatedByIdentityName  | AssignCreatedByName
	// goverter:map ID                     | AssignID
	ComponentEventCreate(source site.ComponentEvent, ctx ConverterContext) site.CreateComponentEventParams

	// goverter:context ctx
	// goverter:map CreatedByIdentityName  | AssignCreatedByName
	// goverter:map ID                     | AssignID
	ActivityCreate(source site.Activity, ctx ConverterContext) site.CreateActivityParams

	///////
	// Update
	///////

	// goverter:context ctx
	// goverter:map UpdatedByIdentityName | AssignUpdatedByName
	EventGroupUpdate(source site.EventGroup, ctx ConverterContext) site.UpdateEventGroupParams

	///////
	// Food Intake
	///////

	// goverter:context ctx
	// goverter:map UpdatedByIdentityName | AssignUpdatedByName
	FoodIntakeUpdate(source site.FoodIntake, ctx ConverterContext) site.UpdateFoodIntakeParams

	///////
	// Glucose Reading
	///////

	// goverter:context ctx
	// goverter:map UpdatedByIdentityName | AssignUpdatedByName
	GlucoseReadingUpdate(source site.GlucoseReading, ctx ConverterContext) site.UpdateGlucoseReadingParams

	///////
	// Insulin Dose
	///////

	// goverter:context ctx
	// goverter:map UpdatedByIdentityName | AssignUpdatedByName
	InsulinDoseUpdate(source site.InsulinDose, ctx ConverterContext) site.UpdateInsulinDoseParams

	///////
	// Notes
	///////

	// goverter:context ctx
	// goverter:map UpdatedByIdentityName | AssignUpdatedByName
	NotificationEventUpdate(source site.NotificationEvent, ctx ConverterContext) site.UpdateNotificationEventParams

	///////
	// Profile
	///////

	// goverter:context ctx
	// goverter:map UpdatedByIdentityName | AssignUpdatedByName
	ProfileUpdate(source site.ProfilePreset, ctx ConverterContext) site.UpdateProfilePresetParams

	///////
	// Profile Switch
	///////

	// goverter:context ctx
	// goverter:map UpdatedByIdentityName | AssignUpdatedByName
	ProfileSwitchUpdate(source site.ProfileSwitch, ctx ConverterContext) site.UpdateProfileSwitchParams

	///////
	// Temp Basal
	///////

	// goverter:context ctx
	// goverter:map UpdatedByIdentityName | AssignUpdatedByName
	TempBasalUpdate(source site.TempBasal, ctx ConverterContext) site.UpdateTempBasalParams

	///////
	// Temp Targets
	///////

	// goverter:context ctx
	// goverter:map UpdatedByIdentityName | AssignUpdatedByName
	TempTargetUpdate(source site.TempTarget, ctx ConverterContext) site.UpdateTempTargetParams

	// goverter:context ctx
	// goverter:map UpdatedByIdentityName | AssignUpdatedByName
	ComponentEventUpdate(source site.ComponentEvent, ctx ConverterContext) site.UpdateComponentEventParams

	// goverter:context ctx
	// goverter:map UpdatedByIdentityName | AssignUpdatedByName
	ActivityUpdate(source site.Activity, ctx ConverterContext) site.UpdateActivityParams

	///////
	// EventGroup
	///////

	// goverter:ignoreMissing
	//
	// goverter:map ID Identifier
	// goverter:map GroupType EventType
	// goverter:map CreatedByIdentityName Subject
	// goverter:map CreatedByIdentityName EnteredBy
	// goverter:map CreatedByAppName App
	// goverter:map CreatedByDeviceName Device
	// goverter:map OccurredAt Date
	// goverter:map . CreatedAt | ModelDateStringFromDateIntAndUTCOffset
	EventGroupToTreatment(source site.EventGroup) *Treatment

	// goverter:update target
	// goverter:ignore CreatedAt
	//
	// goverter:ignoreMissing
	// goverter:map Reading Glucose
	// goverter:map ReadingType GlucoseType
	UpdateGlucoseReadingToTreatment(source *site.GlucoseReading, target *Treatment)

	// goverter:update target
	// goverter:ignore CreatedAt
	// goverter:ignoreMissing
	//
	// goverter:map Dose Insulin
	// goverter:map DoseType Type
	UpdateInsulinDoseToTreatment(source *site.InsulinDose, target *Treatment)

	// goverter:update target
	// goverter:ignore CreatedAt
	// goverter:ignoreMissing
	//
	UpdateFoodIntakeToTreatment(source *site.FoodIntake, target *Treatment)

	// goverter:update target
	// goverter:ignore CreatedAt
	// goverter:ignoreMissing
	//
	// goverter:map Absolute Rate
	// goverter:map DurationMinutes DurationInMilliseconds | MinutesToMilliseconds
	UpdateTempBasalToTreatment(source *site.TempBasal, target *Treatment)

	// goverter:update target
	// goverter:ignore CreatedAt
	// goverter:ignoreMissing
	//
	// goverter:map . CorrectionRange | ToCorrectionRange
	// goverter:map DurationMinutes DurationInMilliseconds | MinutesToMilliseconds
	UpdateTempTargetToTreatment(source *site.TempTarget, target *Treatment)

	// goverter:update target
	// goverter:ignore CreatedAt
	// goverter:ignoreMissing
	//
	UpdateNotificationEventToTreatment(source *site.NotificationEvent, target *Treatment)

	// goverter:update target
	// goverter:ignore CreatedAt
	// goverter:ignoreMissing
	//
	// goverter:map Name Profile
	// goverter:map . ProfileJSON | ToProfileJSON
	// goverter:map DurationMinutes DurationInMilliseconds | MinutesToMilliseconds
	UpdateProfileSwitchToTreatment(source *site.ProfileSwitch, target *Treatment)

	// goverter:update target
	// goverter:ignore CreatedAt
	// goverter:ignoreMissing
	//
	UpdateComponentEventToTreatment(source *site.ComponentEvent, target *Treatment)

	// goverter:update target
	// goverter:ignore CreatedAt
	// goverter:ignoreMissing
	//
	// goverter:map DurationMinutes DurationInMilliseconds | MinutesToMilliseconds
	UpdateActivityToTreatment(source *site.Activity, target *Treatment)

	//////
	// Update Override
	//////

	// goverter:update target
	// goverter:ignore ID
	UpdateToEventGroup(source site.EventGroup, target *site.EventGroup)

	// goverter:update target
	// goverter:ignore ID
	UpdateToFoodIntake(source *site.FoodIntake, target *site.FoodIntake)

	// goverter:update target
	// goverter:ignore ID
	UpdateToGlucoseReading(source *site.GlucoseReading, target *site.GlucoseReading)

	// goverter:update target
	// goverter:ignore ID
	UpdateToInsulinDose(source *site.InsulinDose, target *site.InsulinDose)

	// goverter:update target
	// goverter:ignore ID
	UpdateToNotificationEvent(source *site.NotificationEvent, target *site.NotificationEvent)

	// goverter:update target
	// goverter:ignore ID
	UpdateToProfileSwitch(source *site.ProfileSwitch, target *site.ProfileSwitch)

	// goverter:update target
	// goverter:ignore ID
	UpdateToTempBasal(source *site.TempBasal, target *site.TempBasal)

	// goverter:update target
	// goverter:ignore ID
	UpdateToTempTarget(source *site.TempTarget, target *site.TempTarget)

	// goverter:update target
	// goverter:ignore ID
	UpdateToComponentEvent(source *site.ComponentEvent, target *site.ComponentEvent)

	// goverter:update target
	// goverter:ignore ID
	UpdateToActivity(source *site.Activity, target *site.Activity)

	//////
	// Partial Treatment Converters
	//////

	// goverter:ignore ID DeletedAt CreatedAt CreatedByIdentityName UpdatedAt UpdatedByIdentityName
	// goverter:map Date OccurredAt | OrNow
	// goverter:map App CreatedByAppName
	// goverter:map Device CreatedByDeviceName
	PartialTreatmentToFoodIntake(source Treatment) site.FoodIntake

	// goverter:ignore ID DeletedAt CreatedAt CreatedByIdentityName UpdatedAt UpdatedByIdentityName
	// goverter:map Date OccurredAt | OrNow
	// goverter:map App CreatedByAppName
	// goverter:map Device CreatedByDeviceName
	//
	// goverter:ignore Delta Direction Noise Filtered Unfiltered Rssi ReadingType
	// goverter:map Glucose Reading
	PartialTreatmentToGlucoseReading(source Treatment) site.GlucoseReading

	// goverter:ignore ID DeletedAt CreatedAt CreatedByIdentityName UpdatedAt UpdatedByIdentityName
	// goverter:map Date OccurredAt | OrNow
	// goverter:map App CreatedByAppName
	// goverter:map Device CreatedByDeviceName
	//
	// goverter:map Insulin Dose
	// goverter:map Type DoseType
	PartialTreatmentToInsulinDose(source Treatment) site.InsulinDose

	// goverter:ignore ID DeletedAt CreatedAt CreatedByIdentityName UpdatedAt UpdatedByIdentityName
	// goverter:map Date OccurredAt | OrNow
	// goverter:map App CreatedByAppName
	// goverter:map Device CreatedByDeviceName
	//
	// goverter:ignore NotificationType Title Severity
	// goverter:map Notes Content
	PartialTreatmentToNotificationEvent(source Treatment) site.NotificationEvent

	// goverter:ignore ID DeletedAt CreatedAt CreatedByIdentityName UpdatedAt UpdatedByIdentityName
	// goverter:map Date OccurredAt | OrNow
	// goverter:map App CreatedByAppName
	// goverter:map Device CreatedByDeviceName
	//
	// goverter:ignore Dia Units Timezone CarbRatiosJson InsulinSensitivitiesJson  BasalRatesJson TargetsHighJson TargetsLowJson
	PartialTreatmentToProfileSwitch(source Treatment) site.ProfileSwitch

	// goverter:ignore ID DeletedAt CreatedAt CreatedByIdentityName UpdatedAt UpdatedByIdentityName
	// goverter:map Date OccurredAt | OrNow
	// goverter:map App CreatedByAppName
	// goverter:map Device CreatedByDeviceName
	PartialTreatmentToTempBasal(source Treatment) site.TempBasal

	// goverter:ignore ID DeletedAt CreatedAt CreatedByIdentityName UpdatedAt UpdatedByIdentityName
	// goverter:map Date OccurredAt | OrNow
	// goverter:map App CreatedByAppName
	// goverter:map Device CreatedByDeviceName
	PartialTreatmentToTempTarget(source Treatment) site.TempTarget

	// goverter:ignore ID DeletedAt CreatedAt CreatedByIdentityName UpdatedAt UpdatedByIdentityName
	// goverter:map Date OccurredAt | OrNow
	// goverter:map App CreatedByAppName
	// goverter:map Device CreatedByDeviceName
	PartialTreatmentToComponentEvent(source Treatment) site.ComponentEvent

	// goverter:ignore ID DeletedAt CreatedAt CreatedByIdentityName UpdatedAt UpdatedByIdentityName
	// goverter:map Date OccurredAt | OrNow
	// goverter:map App CreatedByAppName
	// goverter:map Device CreatedByDeviceName
	PartialTreatmentToActivity(source Treatment) site.Activity

	// goverter:ignore ID DeletedAt CreatedAt CreatedByIdentityName UpdatedAt UpdatedByIdentityName
	// goverter:map Date OccurredAt | OrNow
	// goverter:map App CreatedByAppName
	// goverter:map Device CreatedByDeviceName
	//
	// goverter:ignore GlucoseReadingID InsulinDoseID FoodIntakeID TempBasalID TempTargetID ProfileSwitchID ComponentEventID ActivityID NotificationEventID
	// goverter:map EventType GroupType
	TreatmentToEventGroup(source Treatment) site.EventGroup
}

func MinutesToMilliseconds(durationMinutes any) *FlexibleInt64 {
	switch v := durationMinutes.(type) {
	case int64:
		value := FlexibleInt64(v * 60 * 1000)
		return &value
	case *int64:
		if v == nil {
			return nil
		}
		value := FlexibleInt64(*v * 60 * 1000)
		return &value
	}

	return nil
}

func ToProfileJSON(source interface {
	GetDia() *float64
	GetTimezone() *string
	GetUnits() *string
	GetCarbRatiosJson() *string
	GetInsulinSensitivitiesJson() *string
	GetBasalRatesJson() *string
	GetTargetsLowJson() *string
	GetTargetsHighJson() *string
}) *LongJSONString {
	profile := Profile{}

	if source.GetDia() != nil {
		profile.Dia = ptr(FlexibleFloat64(*source.GetDia()))
	}
	if source.GetTimezone() != nil {
		profile.Timezone = ptr(Timezone(*source.GetTimezone()))
	}
	if source.GetUnits() != nil {
		profile.Units = ptr(Units(*source.GetUnits()))
	}
	if source.GetCarbRatiosJson() != nil {
		json.Unmarshal([]byte(*source.GetCarbRatiosJson()), &profile.CarbRatio)
	}
	if source.GetInsulinSensitivitiesJson() != nil {
		json.Unmarshal([]byte(*source.GetInsulinSensitivitiesJson()), &profile.Sens)
	}
	if source.GetBasalRatesJson() != nil {
		json.Unmarshal([]byte(*source.GetBasalRatesJson()), &profile.Basal)
	}
	if source.GetTargetsLowJson() != nil {
		json.Unmarshal([]byte(*source.GetTargetsLowJson()), &profile.TargetLow)
	}
	if source.GetTargetsHighJson() != nil {
		json.Unmarshal([]byte(*source.GetTargetsHighJson()), &profile.TargetHigh)
	}

	profileBytes, err := json.Marshal(profile)
	if err != nil {
		return nil
	}
	profileJSON := LongJSONString(string(profileBytes))
	return &profileJSON
}

func ToCorrectionRange(source interface {
	GetLowTarget() float64
	GetHighTarget() float64
}) []*FlexibleFloat64 {

	correctionRange := []*FlexibleFloat64{
		ptr(FlexibleFloat64(source.GetLowTarget())),
		ptr(FlexibleFloat64(source.GetHighTarget())),
	}
	return correctionRange
}
