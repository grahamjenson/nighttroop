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

	// goverter:map . CreatedAt | ModelDateStringFromDateIntAndUTCOffset
	// goverter:map EventGroup.UtcOffsetMinutes UtcOffsetMinutes
	// goverter:map EventGroup.RawJsonData RawJsonData
	// goverter:map EventGroup.CreatedByIdentityName Subject
	// goverter:map EventGroup.CreatedByIdentityName EnteredBy
	// goverter:map EventGroup.CreatedByAppName App
	// goverter:map EventGroup.CreatedByDeviceName Device
	// goverter:map EventGroup.ID Identifier
	// goverter:map EventGroup.OccurredAt Date
	// goverter:map EventGroup.Notes Notes
	// goverter:map EventGroup.Reason Reason
	// goverter:map EventGroup.GroupType EventType
	//
	// goverter:ignore  SplitExt OriginalProfileName
	//
	//
	// goverter:map GlucoseReading.Reading Glucose
	// goverter:map GlucoseReading.ReadingType GlucoseType
	//
	// goverter:map InsulinDose.Dose Insulin
	// goverter:map InsulinDose.DoseType Type
	// goverter:map InsulinDose.SplitNow SplitNow
	// goverter:map InsulinDose.SplitExt SplitExt
	//
	// goverter:map FoodIntake.FoodType FoodType
	// goverter:map FoodIntake.AbsorptionTimeMinutes AbsorptionTimeMinutes
	// goverter:map FoodIntake.PreBolus PreBolus
	// goverter:map FoodIntake.Category Category
	// goverter:map FoodIntake.Subcategory Subcategory
	// goverter:map FoodIntake.Portions Portions
	// goverter:map FoodIntake.Carbs Carbs
	// goverter:map FoodIntake.Protein Protein
	// goverter:map FoodIntake.Fat Fat
	// goverter:map FoodIntake.Fibre Fibre
	// goverter:map FoodIntake.Energy Energy
	//
	// goverter:map TempBasal.Absolute Absolute
	// goverter:map TempBasal.Absolute Rate
	//
	// goverter:map TempTarget.HighTarget HighTarget
	// goverter:map TempTarget.LowTarget LowTarget
	//
	// goverter:map ProfileSwitch.Name Name
	// goverter:map ProfileSwitch.Name Profile
	// goverter:map ProfileSwitch.Timeshift Timeshift
	// goverter:map ProfileSwitch.Units Units
	// goverter:map . ProfileJSON | ToProfileJSON
	//
	// goverter:map Activity.ActivityType ActivityType
	// goverter:map Activity.Steps Steps
	// goverter:map Activity.Calories Calories
	// goverter:map Activity.HeartRate HeartRate
	//
	// goverter:map ComponentEvent.PumpID PumpID
	// goverter:map ComponentEvent.PumpType PumpType
	// goverter:map ComponentEvent.PumpSerial PumpSerial
	//
	// SHARED FIELDS
	// goverter:map . DurationInMilliseconds     | MinutesToMilliseconds
	// goverter:map . DurationMinutes            | DurationSelect
	// goverter:map . Percent                    | PercentageSelect
	DBModelToTreatment(source *site.EventGroupsBundle) Treatment

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

func MinutesToMilliseconds(source *site.EventGroupsBundle) *FlexibleInt64 {

	if source.TempBasal != nil {
		value := FlexibleInt64(source.TempBasal.DurationMinutes * 60 * 1000)
		return &value
	}
	if source.TempTarget != nil {
		value := FlexibleInt64(source.TempTarget.DurationMinutes * 60 * 1000)
		return &value

	}

	if source.ProfileSwitch != nil {
		value := FlexibleInt64(source.ProfileSwitch.DurationMinutes * 60 * 1000)
		return &value
	}

	if source.Activity != nil && source.Activity.DurationMinutes != nil {
		value := FlexibleInt64(*source.Activity.DurationMinutes * 60 * 1000)
		return &value
	}

	return nil
}

func DurationSelect(source *site.EventGroupsBundle) *FlexibleInt64 {
	if source.TempBasal != nil {
		value := FlexibleInt64(source.TempBasal.DurationMinutes)
		return &value
	}

	if source.TempTarget != nil {
		value := FlexibleInt64(source.TempTarget.DurationMinutes)
		return &value

	}

	if source.ProfileSwitch != nil {
		value := FlexibleInt64(source.ProfileSwitch.DurationMinutes)
		return &value
	}

	if source.Activity != nil && source.Activity.DurationMinutes != nil {
		value := FlexibleInt64(*source.Activity.DurationMinutes)
		return &value
	}

	return nil
}

func PercentageSelect(source *site.EventGroupsBundle) *FlexibleFloat64 {
	if source.TempBasal != nil {
		if source.TempBasal.Percent != nil {
			value := FlexibleFloat64(*source.TempBasal.Percent)
			return &value
		}
	}

	if source.ProfileSwitch != nil {
		if source.ProfileSwitch.Percent != nil {
			value := FlexibleFloat64(*source.ProfileSwitch.Percent)
			return &value
		}
	}

	return nil
}

func ToProfileJSON(source *site.EventGroupsBundle) *LongJSONString {
	if source == nil || source.ProfileSwitch == nil {
		return nil
	}
	profile := Profile{}
	profileSwitch := *source.ProfileSwitch
	if profileSwitch.Dia != nil {
		profile.Dia = ptr(FlexibleFloat64(*profileSwitch.Dia))
	}
	if profileSwitch.Timezone != nil {
		profile.Timezone = ptr(Timezone(*profileSwitch.Timezone))
	}
	if profileSwitch.Units != nil {
		profile.Units = ptr(Units(*profileSwitch.Units))
	}
	if profileSwitch.CarbRatiosJson != nil {
		json.Unmarshal([]byte(*profileSwitch.CarbRatiosJson), &profile.CarbRatio)
	}
	if profileSwitch.InsulinSensitivitiesJson != nil {
		json.Unmarshal([]byte(*profileSwitch.InsulinSensitivitiesJson), &profile.Sens)
	}
	if profileSwitch.BasalRatesJson != nil {
		json.Unmarshal([]byte(*profileSwitch.BasalRatesJson), &profile.Basal)
	}
	if profileSwitch.TargetsLowJson != nil {
		json.Unmarshal([]byte(*profileSwitch.TargetsLowJson), &profile.TargetLow)
	}
	if profileSwitch.TargetsHighJson != nil {
		json.Unmarshal([]byte(*profileSwitch.TargetsHighJson), &profile.TargetHigh)
	}

	profileBytes, err := json.Marshal(profile)
	if err != nil {
		return nil
	}
	profileJSON := LongJSONString(string(profileBytes))
	return &profileJSON
}
