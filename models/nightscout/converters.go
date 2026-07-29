package nightscout

import (
	"nighttroop.t1d.nz/data/sql/site"
	"nighttroop.t1d.nz/utils"
)

//go:generate goverter gen .

type CreateContext struct {
	IDPrefix              string
	CreatedByIdentityID   int64
	CreatedByIdentityName string
}

type UpdateContext struct {
	UpdatedByIdentityID   int64
	UpdatedByIdentityName string
}

// goverter:converter
// goverter:output:format function
// goverter:output:file ./z_converters.go
type MapConverter interface {

	///////
	// ACTIVITIES
	///////

	// goverter:useZeroValueOnPointerInconsistency
	// goverter:map Identifier ID
	// goverter:map Date OccurredAt
	// goverter:map Duration DurationMinutes
	// goverter:map  App CreatedByAppName
	// goverter:map  Device CreatedByDeviceName
	// goverter:context ctx
	// goverter:map  Date ID                | AssignID
	// goverter:map  CreatedByIdentityName  | AssignCreatedByName
	ActivityToCreate(source Activity, ctx CreateContext) site.CreateActivityParams

	// goverter:useZeroValueOnPointerInconsistency
	// goverter:map Identifier ID
	// goverter:map Date OccurredAt
	// goverter:map UpdatedByIdentityName | AssignUpdatedByName
	// goverter:context ctx
	ActivityToUpdate(source Activity, ctx UpdateContext) site.UpdateActivityParams

	// goverter:useZeroValueOnPointerInconsistency
	// goverter:map . CreatedAt | ModelDateStringFromDateIntAndUTCOffset
	// goverter:map CreatedByIdentityName Subject
	// goverter:map CreatedByAppName App
	// goverter:map CreatedByDeviceName Device
	// goverter:map ID Identifier
	// goverter:map OccurredAt Date
	// goverter:map DurationMinutes Duration
	// goverter:map CreatedAt SrvCreated
	// goverter:map UpdatedAt SrvModified
	DBModelToActivity(source *site.Activity) Activity

	///////
	// ENTRIES TO GLUCOSE READINGS
	///////

	// goverter:useZeroValueOnPointerInconsistency
	// goverter:context ctx
	// goverter:map  Date ID                | AssignID
	// goverter:map  Date OccurredAt
	// goverter:map  App CreatedByAppName
	// goverter:map  Device CreatedByDeviceName
	// goverter:map  CreatedByIdentityName  | AssignCreatedByName
	EntryToCreate(source Entry, ctx CreateContext) site.CreateGlucoseReadingParams

	// goverter:useZeroValueOnPointerInconsistency
	// goverter:map Identifier ID
	// goverter:map Date OccurredAt
	// goverter:map UpdatedByIdentityName | AssignUpdatedByName
	// goverter:context ctx
	EntryToUpdate(source Entry, ctx UpdateContext) site.UpdateGlucoseReadingParams

	// goverter:useZeroValueOnPointerInconsistency
	// goverter:ignore Direction Trend Units Filtered Unfiltered RSSI Noise SysTime
	// goverter:map CreatedByIdentityName Subject
	// goverter:map CreatedByAppName App
	// goverter:map CreatedByDeviceName Device
	// goverter:map ID Identifier
	// goverter:map . SGV | SGVGlucoseReadingTypeToValue
	// goverter:map . MBG | MBGGlucoseReadingTypeToValue
	// goverter:map OccurredAt Date
	// goverter:map OccurredAt Mills
	// goverter:map CreatedAt SrvCreated
	// goverter:map UpdatedAt SrvModified
	// goverter:map . DateString | ModelDateStringFromDateIntAndUTCOffset
	DBModelToEntry(source *site.GlucoseReading) Entry

	///////
	// FOOD TO FOOD PRESET
	///////

	// goverter:useZeroValueOnPointerInconsistency
	// goverter:ignore Calories
	// goverter:context ctx
	// goverter:map  App CreatedByAppName
	// goverter:map  Device CreatedByDeviceName
	// goverter:map  Date ID                | AssignID
	// goverter:map  CreatedByIdentityName  | AssignCreatedByName
	FoodToCreate(source Food, ctx CreateContext) site.CreateFoodPresetParams

	// goverter:useZeroValueOnPointerInconsistency
	// goverter:map Identifier ID
	// goverter:map Date OccurredAt
	// goverter:map UpdatedByIdentityName | AssignUpdatedByName
	// goverter:context ctx
	FoodToUpdate(source Food, ctx UpdateContext) site.UpdateFoodPresetParams

	// goverter:useZeroValueOnPointerInconsistency
	// goverter:ignore Device App Subject Food Portions HideAfterUse Hidden Position
	// goverter:ignore Category Subcategory Portion Unit Fat Protein Energy GI Date
	// goverter:map ID Identifier
	// goverter:map CreatedAt SrvCreated
	// goverter:map UpdatedAt SrvModified
	DBModelToFood(source *site.FoodPreset) Food

	///////
	// Profile To Profile Preset
	///////

	// goverter:useZeroValueOnPointerInconsistency
	// goverter:context ctx
	// goverter:ignore  Start DurationMinutes Source Percent Dia Units Timezone CarbRatiosJson InsulinSensitivitiesJson BasalRatesJson TargetsJson
	// goverter:map  App CreatedByAppName
	// goverter:map  Device CreatedByDeviceName
	// goverter:map  Date ID                | AssignID
	// goverter:map  DefaultProfile Name
	// goverter:map  CreatedByIdentityName  | AssignCreatedByName
	ProfileToCreate(source Profile, ctx CreateContext) site.CreateProfilePresetParams

	// goverter:useZeroValueOnPointerInconsistency
	// goverter:map Identifier ID
	// goverter:map Date OccurredAt
	// goverter:map UpdatedByIdentityName | AssignUpdatedByName
	// goverter:context ctx
	ProfileToUpdate(source Profile, ctx UpdateContext) site.UpdateProfilePresetParams

	// goverter:useZeroValueOnPointerInconsistency
	// goverter:ignore Device App Subject
	// goverter:ignore DefaultProfile Store
	// goverter:map ID Identifier
	// goverter:map . StartDate | ModelDateStringFromDateIntAndUTCOffset
	// goverter:map OccurredAt Mills
	// goverter:map OccurredAt Date
	// goverter:map CreatedAt SrvCreated
	// goverter:map UpdatedAt SrvModified
	DBModelToProfile(source *site.GetProfilePresetParams) Profile

	//
	//
	//

	////////////////////////////////////////
	// TREATMENT TO CREATE/UPDATE
	////////////////////////////////////////

	///////
	// EventGroup
	///////

	// goverter:useZeroValueOnPointerInconsistency
	// goverter:map ID Identifier
	// goverter:map . StartDate | ModelDateStringFromDateIntAndUTCOffset
	// goverter:map OccurredAt Mills
	// goverter:map OccurredAt Date
	// goverter:map CreatedAt SrvCreated
	// goverter:map UpdatedAt SrvModified
	DBModelToTreatment(source *site.SearchEventGroupsRow) Treatment

	///////
	// Food Intake
	///////

	// goverter:useZeroValueOnPointerInconsistency
	// goverter:context ctx
	// goverter:map  App CreatedByAppName
	// goverter:map  Device CreatedByDeviceName
	// goverter:map  Date ID                | AssignID
	// goverter:map  Date OccurredAt
	// goverter:map  CreatedByIdentityName  | AssignCreatedByName
	// goverter:map  EventType Type 	    | TreatmentTypeToDBType
	// goverter:map  Carbs Amount
	TreatmentToCreateFoodIntake(source Treatment, ctx CreateContext) site.CreateFoodIntakeParams

	// goverter:useZeroValueOnPointerInconsistency
	// goverter:map  Identifier ID
	// goverter:map  Date OccurredAt
	// goverter:map  Carbs Amount
	// goverter:map UpdatedByIdentityName | AssignUpdatedByName
	// goverter:context ctx
	TreatmentToUpdateFoodIntake(source Treatment, ctx UpdateContext) site.UpdateFoodIntakeParams

	///////
	// Glucose Reading
	///////

	// goverter:useZeroValueOnPointerInconsistency
	// goverter:context ctx
	// goverter:map  App CreatedByAppName
	// goverter:map  Device CreatedByDeviceName
	// goverter:map  Date ID                | AssignID
	// goverter:map  Date OccurredAt
	// goverter:map  CreatedByIdentityName  | AssignCreatedByName
	// goverter:map  EventType Type 	    | TreatmentTypeToDBType
	// goverter:map  Carbs Amount
	TreatmentToCreateGlucoseReading(source Treatment, ctx CreateContext) site.CreateGlucoseReadingParams

	// goverter:useZeroValueOnPointerInconsistency
	// goverter:map  Identifier ID
	// goverter:map  Date OccurredAt
	// goverter:map  Carbs Amount
	// goverter:map UpdatedByIdentityName | AssignUpdatedByName
	// goverter:context ctx
	TreatmentToUpdateGlucoseReading(source Treatment, ctx UpdateContext) site.UpdateGlucoseReadingParams

	///////
	// Insulin Dose
	///////

	// goverter:useZeroValueOnPointerInconsistency
	// goverter:context ctx
	// goverter:map  App CreatedByAppName
	// goverter:map  Device CreatedByDeviceName
	// goverter:map  Date ID                | AssignID
	// goverter:map  Date OccurredAt
	// goverter:map  CreatedByIdentityName  | AssignCreatedByName
	// goverter:map  EventType Type 	    | TreatmentTypeToDBType
	// goverter:map  Insulin Value
	TreatmentToCreateInsulinDose(source Treatment, ctx CreateContext) site.CreateInsulinDoseParams

	// goverter:useZeroValueOnPointerInconsistency
	// goverter:map  Identifier ID
	// goverter:map  Date OccurredAt
	// goverter:map  Insulin Value
	// goverter:map UpdatedByIdentityName | AssignUpdatedByName
	// goverter:context ctx
	TreatmentToUpdateInsulinDose(source Treatment, ctx UpdateContext) site.UpdateInsulinDoseParams

	///////
	// Notes
	///////
	// goverter:useZeroValueOnPointerInconsistency
	// goverter:context ctx
	// goverter:map  App CreatedByAppName
	// goverter:map  Device CreatedByDeviceName
	// goverter:map  Date ID                | AssignID
	// goverter:map  Date OccurredAt
	// goverter:map  CreatedByIdentityName  | AssignCreatedByName
	// goverter:map  Notes Content
	TreatmentToCreateNote(source Treatment, ctx CreateContext) site.CreateNoteParams

	// goverter:useZeroValueOnPointerInconsistency
	// goverter:map  Identifier ID
	// goverter:map  Date OccurredAt
	// goverter:map  Notes Content
	// goverter:map UpdatedByIdentityName | AssignUpdatedByName
	// goverter:context ctx
	TreatmentToUpdateNote(source Treatment, ctx UpdateContext) site.UpdateNoteParams

	///////
	// Profile Switch
	///////

	// goverter:useZeroValueOnPointerInconsistency
	// goverter:context ctx
	// goverter:map  App CreatedByAppName
	// goverter:map  Device CreatedByDeviceName
	// goverter:map  Date ID                | AssignID
	// goverter:map  Date OccurredAt
	// goverter:map  CreatedByIdentityName  | AssignCreatedByName
	// goverter:map  EventType Type 	    | TreatmentTypeToDBType
	// goverter:map  Insulin Value
	TreatmentToCreateProfileSwitch(source Treatment, ctx CreateContext) site.CreateProfileSwitchParams

	// goverter:useZeroValueOnPointerInconsistency
	// goverter:map  Identifier ID
	// goverter:map  Date OccurredAt
	// goverter:map  Insulin Value
	// goverter:map UpdatedByIdentityName | AssignUpdatedByName
	// goverter:context ctx
	TreatmentToUpdateProfileSwitch(source Treatment, ctx UpdateContext) site.UpdateProfileSwitchParams

	///////
	// Temp Basal
	///////

	// goverter:useZeroValueOnPointerInconsistency
	// goverter:context ctx
	// goverter:map  App CreatedByAppName
	// goverter:map  Device CreatedByDeviceName
	// goverter:map  Date ID                | AssignID
	// goverter:map  Date Start
	// goverter:map  CreatedByIdentityName  | AssignCreatedByName
	// goverter:map  . Value | TempBasalValue
	// goverter:map Duration DurationMinutes
	TreatmentToCreateTempBasal(source Treatment, ctx CreateContext) site.CreateTempBasalParams

	// goverter:useZeroValueOnPointerInconsistency
	// goverter:map  Identifier ID
	// goverter:map  Date Start
	// goverter:map  . Value | TempBasalValuePtr
	// goverter:map Duration DurationMinutes
	// goverter:map UpdatedByIdentityName | AssignUpdatedByName
	// goverter:context ctx
	TreatmentToUpdateTempBasal(source Treatment, ctx UpdateContext) site.UpdateTempBasalParams

	///////
	// Temp Targets
	///////

	// goverter:useZeroValueOnPointerInconsistency
	// goverter:context ctx
	// goverter:map  App CreatedByAppName
	// goverter:map  Device CreatedByDeviceName
	// goverter:map  Date ID                | AssignID
	// goverter:map  Date Start
	// goverter:map  CreatedByIdentityName  | AssignCreatedByName
	// goverter:map Duration DurationMinutes
	// goverter:map TargetTop HighTarget
	// goverter:map TargetBottom LowTarget
	TreatmentToCreateTempTarget(source Treatment, ctx CreateContext) site.CreateTempTargetParams

	// goverter:useZeroValueOnPointerInconsistency
	// goverter:map  Identifier ID
	// goverter:map  Date Start
	// goverter:map Duration DurationMinutes
	// goverter:map TargetTop HighTarget
	// goverter:map TargetBottom LowTarget
	// goverter:map UpdatedByIdentityName | AssignUpdatedByName
	// goverter:context ctx
	TreatmentToUpdateTempTarget(source Treatment, ctx UpdateContext) site.UpdateTempTargetParams

	///////
	// Pump States
	///////

	// goverter:useZeroValueOnPointerInconsistency
	// goverter:map  Identifier ID
	// goverter:map  Date Start
	// goverter:map Duration DurationMinutes
	// goverter:map TargetTop HighTarget
	// goverter:map TargetBottom LowTarget
	// goverter:map UpdatedByIdentityName | AssignUpdatedByName
	// goverter:context ctx
	TreatmentToCreatePumpState(source Treatment, ctx CreateContext) site.CreatePumpStateParams

	// goverter:useZeroValueOnPointerInconsistency
	// goverter:map  Identifier ID
	// goverter:map  Date Start
	// goverter:map Duration DurationMinutes
	// goverter:map TargetTop HighTarget
	// goverter:map TargetBottom LowTarget
	// goverter:map UpdatedByIdentityName | AssignUpdatedByName
	// goverter:context ctx
	TreatmentToUpdatePumpState(source Treatment, ctx UpdateContext) site.UpdatePumpStateParams

	///////
	// CGM States
	///////

	// goverter:useZeroValueOnPointerInconsistency
	// goverter:map  Identifier ID
	// goverter:map  Date Start
	// goverter:map Duration DurationMinutes
	// goverter:map TargetTop HighTarget
	// goverter:map TargetBottom LowTarget
	// goverter:map UpdatedByIdentityName | AssignUpdatedByName
	// goverter:context ctx
	TreatmentToCreateCGMState(source Treatment, ctx CreateContext) site.CreateCGMStateParams

	// goverter:useZeroValueOnPointerInconsistency
	// goverter:map  Identifier ID
	// goverter:map  Date Start
	// goverter:map Duration DurationMinutes
	// goverter:map TargetTop HighTarget
	// goverter:map TargetBottom LowTarget
	// goverter:map UpdatedByIdentityName | AssignUpdatedByName
	// goverter:context ctx
	TreatmentToUpdateCGMState(source Treatment, ctx UpdateContext) site.UpdateCGMStateParams

	//
	//
	//

	////////////////////////////////////////
	// DEVICE STATUS TO CREATE/UPDATE
	////////////////////////////////////////

	// goverter:useZeroValueOnPointerInconsistency
	// goverter:ignore InsertionTime Type Status ChargingState Timestamp Iob Cob Basaliob Treatmentiob PumpReservoir
	// goverter:map Identifier ID
	// goverter:map Date OccurredAt
	// goverter:map UploaderBattery BatteryPercentage
	// goverter:map  App CreatedByAppName
	// goverter:map  Device CreatedByDeviceName
	// goverter:context ctx
	// goverter:map  Date ID                | AssignID
	// goverter:map  CreatedByIdentityName  | AssignCreatedByName
	DeviceStatusToCreateDevice(source DeviceStatus, ctx CreateContext) site.CreateDeviceStateParams

	// goverter:useZeroValueOnPointerInconsistency
	// goverter:ignore InsertionTime Type Status ChargingState Timestamp Iob Cob Basaliob Treatmentiob PumpReservoir
	// goverter:map Identifier ID
	// goverter:map Date OccurredAt
	// goverter:map UploaderBattery BatteryPercentage
	// goverter:map  App CreatedByAppName
	// goverter:map  Device CreatedByDeviceName
	// goverter:context ctx
	// goverter:map  Date ID                | AssignID
	// goverter:map  CreatedByIdentityName  | AssignCreatedByName
	DeviceStatusToCreateLoop(source DeviceStatus, ctx CreateContext) site.CreateLoopStateParams

	// goverter:useZeroValueOnPointerInconsistency
	// goverter:ignore InsertionTime Type Status ChargingState Timestamp Iob Cob Basaliob Treatmentiob PumpReservoir
	// goverter:map Identifier ID
	// goverter:map Date OccurredAt
	// goverter:map UploaderBattery BatteryPercentage
	// goverter:map  App CreatedByAppName
	// goverter:map  Device CreatedByDeviceName
	// goverter:context ctx
	// goverter:map  Date ID                | AssignID
	// goverter:map  CreatedByIdentityName  | AssignCreatedByName
	DeviceStatusToCreateCGM(source DeviceStatus, ctx CreateContext) site.CreateCGMStateParams

	// goverter:useZeroValueOnPointerInconsistency
	// goverter:ignore InsertionTime Type Status ChargingState Timestamp Iob Cob Basaliob Treatmentiob PumpReservoir
	// goverter:map Identifier ID
	// goverter:map Date OccurredAt
	// goverter:map UploaderBattery BatteryPercentage
	// goverter:map  App CreatedByAppName
	// goverter:map  Device CreatedByDeviceName
	// goverter:context ctx
	// goverter:map  Date ID                | AssignID
	// goverter:map  CreatedByIdentityName  | AssignCreatedByName
	DeviceStatusToCreatePump(source DeviceStatus, ctx CreateContext) site.CreatePumpStateParams

	// goverter:useZeroValueOnPointerInconsistency
	// goverter:ignore Uploader Loop Pump OpenAPS
	// goverter:map ID Identifier
	// goverter:map . CreatedAt | ModelDateStringFromDateIntAndUTCOffset
	// goverter:map CreatedByIdentityName Subject
	// goverter:map CreatedByAppName App
	// goverter:map CreatedByDeviceName Device
	// goverter:map OccurredAt Date
	// goverter:map BatteryPercentage UploaderBattery
	// goverter:map CreatedAt SrvCreated
	// goverter:map UpdatedAt SrvModified
	DBModelToDeviceStatus(source *site.SearchEventGroupsRow) DeviceStatus

	////////
	// SETTINGS
	////////

	// Doing this manually for control
}

// //
// Custom Converter Methods
// //

// goverter:context ctx
func AssignID(date *DateInt, ctx CreateContext) string {
	return utils.NextDateID(ctx.IDPrefix, date.Int64())
}

// goverter:context ctx
func AssignCreatedByName(ctx CreateContext) string {
	return ctx.CreatedByIdentityName
}

// goverter:context ctx
func AssignUpdatedByName(ctx UpdateContext) *string {
	return ptr(ctx.UpdatedByIdentityName)
}

func ToDateInt(createdAt int64) *DateInt {
	x := DateInt(createdAt)
	return &x
}

// ////
// GlucoseReading
// ////
func SGVGlucoseReadingTypeToValue(gr site.GlucoseReading) *FlexibleFloat64 {
	if gr.ReadingType == "sgv" {
		return ptr(FlexibleFloat64(gr.Reading))
	}
	return nil

}

func MBGGlucoseReadingTypeToValue(gr site.GlucoseReading) *FlexibleFloat64 {
	if gr.ReadingType == "mbg" {
		return ptr(FlexibleFloat64(gr.Reading))
	}
	return nil

}

func ModelDateStringFromDateIntAndUTCOffset(m interface{}) *DateString {
	if m == nil {
		return nil
	}

	// TODO this should be mapping to interfaces now we have getters
	switch v := m.(type) {
	case interface {
		GetOccurredAt() int64
		GetUtcOffsetMinutes() *int64
	}:
		ds := DateStringFromInt64AndUTCOffset(v.GetOccurredAt(), v.GetUtcOffsetMinutes())
		return &ds
	case interface {
		GetStart() int64
		GetUtcOffsetMinutes() *int64
	}:
		ds := DateStringFromInt64AndUTCOffset(v.GetStart(), v.GetUtcOffsetMinutes())
		return &ds
	default:
		return nil
	}

}

// Treatments

func TempBasalValue(t Treatment) float64 {
	v := TempBasalValuePtr(t)
	if v == nil {
		return 0
	}
	return *v
}

func TempBasalValuePtr(t Treatment) *float64 {
	if t.Rate != nil {
		x := float64(*t.Rate)
		return &x
	}
	if t.Absolute != nil {
		x := float64(*t.Absolute)
		return &x
	}
	// TODO handle t.Percent
	return nil
}

func TreatmentTypeToDBType(t *TreatmentEventType) string {
	if t == nil {
		return ""
	}

	return t.ToLocalTypeString()
}
