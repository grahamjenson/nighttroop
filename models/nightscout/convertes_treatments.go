package nightscout

import (
	"nighttroop.t1d.nz/data/sql/site"
)

// goverter:converter
// goverter:output:format function
// goverter:output:file ./z_treatments_converters.go
// goverter:useZeroValueOnPointerInconsistency
type TreatmentsEvenetConverter interface {

	///////
	// CREATE
	///////

	// goverter:context ctx
	// goverter:map CreatedByIdentityName  | AssignCreatedByName
	EventGroupCreate(source site.EventGroup, ctx ConverterContext) site.CreateEventGroupParams

	///////
	// Food Intake
	///////

	// goverter:context ctx
	// goverter:map CreatedByIdentityName  | AssignCreatedByName
	FoodIntakeCreate(source site.FoodIntake, ctx ConverterContext) site.CreateFoodIntakeParams

	///////
	// Glucose Reading
	///////

	// goverter:context ctx
	// goverter:map CreatedByIdentityName  | AssignCreatedByName
	GlucoseReadingCreate(source site.GlucoseReading, ctx ConverterContext) site.CreateGlucoseReadingParams

	///////
	// Insulin Dose
	///////

	// goverter:context ctx
	// goverter:map CreatedByIdentityName  | AssignCreatedByName
	InsulinDoseCreate(source site.InsulinDose, ctx ConverterContext) site.CreateInsulinDoseParams

	///////
	// Notes
	///////

	// goverter:context ctx
	// goverter:map CreatedByIdentityName  | AssignCreatedByName
	NoteCreate(source site.Note, ctx ConverterContext) site.CreateNoteParams

	///////
	// Profile Switch
	///////

	// goverter:context ctx
	// goverter:map CreatedByIdentityName  | AssignCreatedByName
	ProfileSwitchCreate(source site.ProfileSwitch, ctx ConverterContext) site.CreateProfileSwitchParams

	///////
	// Temp Basal
	///////

	// goverter:context ctx
	// goverter:map CreatedByIdentityName  | AssignCreatedByName
	TempBasalCreate(source site.TempBasal, ctx ConverterContext) site.CreateTempBasalParams

	///////
	// Temp Targets
	///////

	// goverter:context ctx
	// goverter:map CreatedByIdentityName  | AssignCreatedByName
	TempTargetCreate(source site.TempTarget, ctx ConverterContext) site.CreateTempTargetParams

	// goverter:context ctx
	// goverter:map CreatedByIdentityName  | AssignCreatedByName
	ComponentEventCreate(source site.ComponentEvent, ctx ConverterContext) site.CreateComponentEventParams

	// goverter:context ctx
	// goverter:map CreatedByIdentityName  | AssignCreatedByName
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
	NoteUpdate(source site.Note, ctx ConverterContext) site.UpdateNoteParams

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
	//
	// goverter:ignore PreBolus SplitNow SplitExt Profile Units PumpType ProfileJSON Timeshift
	// goverter:map EventGroup.GroupType EventType
	//
	// goverter:map GlucoseReading.Reading Glucose
	// goverter:map GlucoseReading.ReadingType GlucoseType
	//
	// goverter:map InsulinDose.Dose Insulin
	//
	// goverter:map FoodIntake.Category Category
	// goverter:map FoodIntake.Subcategory Subcategory
	// goverter:map FoodIntake.Portions Portions
	// goverter:map FoodIntake.Carbs Carbs
	// goverter:map FoodIntake.Protein Protein
	// goverter:map FoodIntake.Fat Fat
	// goverter:map FoodIntake.Fibre Fibre
	// goverter:map FoodIntake.Energy Energy
	//
	// goverter:map TempBasal.DurationMinutes Duration
	// goverter:map TempBasal.Percent Rate
	// goverter:map TempBasal.Absolute Absolute
	//
	// goverter:map TempTarget.DurationMinutes Duration
	// goverter:map TempTarget.HighTarget TargetTop
	// goverter:map TempTarget.LowTarget TargetBottom
	//
	// goverter:map ProfileSwitch.Name Name
	// goverter:map ProfileSwitch.Name Profile
	// goverter:map ProfileSwitch.Name OriginalProfileName
	// goverter:map ProfileSwitch.Percent Percentage
	//
	// goverter:map Activity.ActivityType ActivityType
	// goverter:map Activity.Steps Steps
	// goverter:map Activity.Calories Calories
	// goverter:map Activity.DurationMinutes Duration
	// goverter:map Activity.HeartRate HeartRate
	//
	// goverter:map Note.Content Notes
	// goverter:map Note.Reason Reason
	DBModelToTreatment(source *site.EventGroupsBundle) Treatment

	////////////////////////////////////////
	// TREATMENT TO CREATE/UPDATE
	////////////////////////////////////////

	// goverter:map Date OccurredAt | OrNow
	// goverter:map App CreatedByAppName
	// goverter:map Device CreatedByDeviceName
	// goverter:map ID                | AssignID
	//
	// goverter:ignore DeletedAt CreatedAt UpdatedAt CreatedByIdentityName UpdatedByIdentityName GlucoseReadingID InsulinDoseID FoodIntakeID TempBasalID TempTargetID ProfileSwitchID ComponentEventID ActivityID NoteID
	// goverter:map Date OccurredAt | OrNow
	// goverter:map EventType GroupType
	TreatmentToEventGroup(source Treatment) site.EventGroup

	///////
	// Food Intake
	///////

	// goverter:map Date OccurredAt | OrNow
	// goverter:map App CreatedByAppName
	// goverter:map Device CreatedByDeviceName
	// goverter:map ID                | AssignID
	//
	// goverter:ignore DeletedAt CreatedAt UpdatedAt CreatedByIdentityName UpdatedByIdentityName
	// goverter:map Date OccurredAt | OrNow
	TreatmentToFoodIntake(source Treatment) site.FoodIntake

	///////
	// Glucose Reading
	///////

	// goverter:map Date OccurredAt | OrNow
	// goverter:map App CreatedByAppName
	// goverter:map Device CreatedByDeviceName
	// goverter:map ID                | AssignID
	//
	// goverter:ignore DeletedAt CreatedAt UpdatedAt CreatedByIdentityName UpdatedByIdentityName
	// goverter:map Glucose Reading
	// goverter:map GlucoseType ReadingType
	TreatmentToGlucoseReading(source Treatment) site.GlucoseReading

	///////
	// Insulin Dose
	///////

	// goverter:map Date OccurredAt | OrNow
	// goverter:map App CreatedByAppName
	// goverter:map Device CreatedByDeviceName
	// goverter:map ID                | AssignID
	//
	// goverter:ignore DeletedAt CreatedAt UpdatedAt CreatedByIdentityName UpdatedByIdentityName
	// goverter:ignore DoseType
	// goverter:map Insulin Dose
	TreatmentToInsulinDose(source Treatment) site.InsulinDose

	///////
	// Notes
	///////

	// goverter:map Date OccurredAt | OrNow
	// goverter:map App CreatedByAppName
	// goverter:map Device CreatedByDeviceName
	// goverter:map ID                | AssignID
	//
	// goverter:ignore DeletedAt CreatedAt UpdatedAt CreatedByIdentityName UpdatedByIdentityName
	// goverter:map Notes Content
	// goverter:map Reason Reason
	TreatmentToNote(source Treatment) site.Note

	///////
	// Profile Switch
	///////

	// goverter:map Date OccurredAt | OrNow
	// goverter:map App CreatedByAppName
	// goverter:map Device CreatedByDeviceName
	// goverter:map ID                | AssignID
	//
	// goverter:ignore DeletedAt CreatedAt UpdatedAt CreatedByIdentityName UpdatedByIdentityName CarbRatiosJson InsulinSensitivitiesJson BasalRatesJson TargetsHighJson TargetsLowJson Dia Timezone
	// goverter:map Duration DurationMinutes
	// goverter:map Percentage Percent
	TreatmentToProfileSwitch(source Treatment) site.ProfileSwitch

	///////
	// Temp Basal
	///////

	// goverter:map Date OccurredAt | OrNow
	// goverter:map App CreatedByAppName
	// goverter:map Device CreatedByDeviceName
	// goverter:map ID                | AssignID
	//
	// goverter:ignore DeletedAt CreatedAt UpdatedAt CreatedByIdentityName UpdatedByIdentityName
	// goverter:map Duration DurationMinutes
	// goverter:map Percentage Percent
	TreatmentToTempBasal(source Treatment) site.TempBasal

	///////
	// Temp Targets
	///////

	// goverter:map Date OccurredAt | OrNow
	// goverter:map App CreatedByAppName
	// goverter:map Device CreatedByDeviceName
	// goverter:map ID                | AssignID
	//
	// goverter:ignore DeletedAt CreatedAt UpdatedAt CreatedByIdentityName UpdatedByIdentityName
	// goverter:map Duration DurationMinutes
	// goverter:map TargetTop HighTarget
	// goverter:map TargetBottom LowTarget
	TreatmentToTempTarget(source Treatment) site.TempTarget

	// goverter:map Date OccurredAt | OrNow
	// goverter:map App CreatedByAppName
	// goverter:map Device CreatedByDeviceName
	// goverter:map ID                | AssignID
	//
	// goverter:ignore DeletedAt CreatedAt UpdatedAt CreatedByIdentityName UpdatedByIdentityName
	TreatmentToComponentEvent(source Treatment) site.ComponentEvent

	// goverter:map Date OccurredAt | OrNow
	// goverter:map App CreatedByAppName
	// goverter:map Device CreatedByDeviceName
	// goverter:map ID                | AssignID
	//
	// goverter:ignore DeletedAt CreatedAt UpdatedAt CreatedByIdentityName UpdatedByIdentityName
	// goverter:map Duration DurationMinutes
	TreatmentToActivity(source Treatment) site.Activity

	//////
	// Update Override
	//////

	// goverter:update target
	UpdateToEventGroup(source site.EventGroup, target *site.EventGroup)

	// goverter:update target
	UpdateToFoodIntake(source site.FoodIntake, target *site.FoodIntake)

	// goverter:update target
	UpdateToGlucoseReading(source site.GlucoseReading, target *site.GlucoseReading)

	// goverter:update target
	UpdateToInsulinDose(source site.InsulinDose, target *site.InsulinDose)

	// goverter:update target
	UpdateToNote(source site.Note, target *site.Note)

	// goverter:update target
	UpdateToProfileSwitch(source site.ProfileSwitch, target *site.ProfileSwitch)

	// goverter:update target
	UpdateToTempBasal(source site.TempBasal, target *site.TempBasal)

	// goverter:update target
	UpdateToTempTarget(source site.TempTarget, target *site.TempTarget)

	// goverter:update target
	UpdateToComponentEvent(source site.ComponentEvent, target *site.ComponentEvent)

	// goverter:update target
	UpdateToActivity(source site.Activity, target *site.Activity)
}
