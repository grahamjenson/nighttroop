package nightscout

import (
	"nighttroop.t1d.nz/data/sql/site"
)

type UpdateContext struct {
	UpdatedByIdentityID   int64
	UpdatedByIdentityName string
}

// goverter:converter
// goverter:output:format function
// goverter:output:file ./z_update_converters.go
// goverter:useZeroValueOnPointerInconsistency
// goverter:extend FlexibleFloat64ToFloat64
// goverter:extend FlexibleInt64ToInt64
// goverter:extend ShortStringToString
type UpdateMapConverter interface {

	///////
	// ACTIVITIES
	///////

	// goverter:context ctx
	// goverter:map Identifier ID
	// goverter:map Date OccurredAt
	// goverter:map UpdatedByIdentityName | AssignUpdatedByName
	//
	// goverter:map Duration DurationMinutes
	ActivityToUpdate(source Activity, ctx UpdateContext) site.UpdateActivityParams

	///////
	// ENTRIES TO GLUCOSE READINGS
	///////

	// goverter:context ctx
	// goverter:map Identifier ID
	// goverter:map Date OccurredAt
	// goverter:map UpdatedByIdentityName | AssignUpdatedByName
	//
	// goverter:ignore
	// goverter:map Type ReadingType
	EntryToUpdate(source Entry, ctx UpdateContext) site.UpdateGlucoseReadingParams

	///////
	// FOOD TO FOOD PRESET
	///////

	// goverter:context ctx
	// goverter:map Identifier ID
	// goverter:map UpdatedByIdentityName | AssignUpdatedByName
	//
	// goverter:ignore
	// goverter:map Food Name
	// goverter:map Category FoodType
	// goverter:map Energy	Calories
	FoodToUpdate(source Food, ctx UpdateContext) site.UpdateFoodPresetParams

	///////
	// Profile To Profile Preset
	///////

	// goverter:context ctx
	// goverter:map Identifier ID
	// goverter:map Date OccurredAt
	// goverter:map UpdatedByIdentityName | AssignUpdatedByName
	//
	// goverter:map Duration DurationMinutes
	// ProfileToUpdate(source Profile, ctx UpdateContext) site.UpdateProfilePresetParams

	//
	//
	//

	////////////////////////////////////////
	// TREATMENT TO CREATE/UPDATE
	////////////////////////////////////////

	///////
	// Food Intake
	///////

	// goverter:context ctx
	// goverter:map Identifier ID
	// goverter:map Date OccurredAt
	// goverter:map UpdatedByIdentityName | AssignUpdatedByName
	//
	// goverter:ignore Name FoodType
	// goverter:map Energy Calories
	TreatmentToUpdateFoodIntake(source Treatment, ctx UpdateContext) site.UpdateFoodIntakeParams

	///////
	// Glucose Reading
	///////

	// goverter:context ctx
	// goverter:map Identifier ID
	// goverter:map Date OccurredAt
	// goverter:map UpdatedByIdentityName | AssignUpdatedByName
	//
	// goverter:map Glucose Reading
	// goverter:map GlucoseType ReadingType
	TreatmentToUpdateGlucoseReading(source Treatment, ctx UpdateContext) site.UpdateGlucoseReadingParams

	///////
	// Insulin Dose
	///////

	// goverter:context ctx
	// goverter:map Identifier ID
	// goverter:map Date OccurredAt
	// goverter:map UpdatedByIdentityName | AssignUpdatedByName
	//
	// goverter:ignore DoseType
	// goverter:map Insulin Dose
	TreatmentToUpdateInsulinDose(source Treatment, ctx UpdateContext) site.UpdateInsulinDoseParams

	///////
	// Notes
	///////

	// goverter:context ctx
	// goverter:map  Identifier ID
	// goverter:map  Date OccurredAt
	// goverter:map  Notes Content
	// goverter:map UpdatedByIdentityName | AssignUpdatedByName
	//
	// goverter:map Notes Content
	TreatmentToUpdateNote(source Treatment, ctx UpdateContext) site.UpdateNoteParams

	///////
	// Profile
	///////

	// // goverter:context ctx
	// // goverter:map Identifier ID
	// // goverter:map Date OccurredAt
	// // goverter:map UpdatedByIdentityName | AssignUpdatedByName
	// //
	// // goverter:map Duration DurationMinutes
	// TreatmentToUpdateProfile(source Treatment, ctx UpdateContext) site.UpdateProfilePresetParams

	///////
	// Profile Switch
	///////

	// // goverter:context ctx
	// // goverter:map Identifier ID
	// // goverter:map Date OccurredAt
	// // goverter:map UpdatedByIdentityName | AssignUpdatedByName
	// //
	// // goverter:map Duration DurationMinutes
	// TreatmentToUpdateProfileSwitch(source Treatment, ctx UpdateContext) site.UpdateProfileSwitchParams

	///////
	// Temp Basal
	///////

	// goverter:context ctx
	// goverter:map Identifier ID
	// goverter:map Date OccurredAt
	// goverter:map UpdatedByIdentityName | AssignUpdatedByName
	//
	// goverter:map Duration DurationMinutes
	// goverter:map Percentage Percent
	TreatmentToUpdateTempBasal(source Treatment, ctx UpdateContext) site.UpdateTempBasalParams

	///////
	// Temp Targets
	///////

	// goverter:context ctx
	// goverter:map Identifier ID
	// goverter:map Date OccurredAt
	// goverter:map UpdatedByIdentityName | AssignUpdatedByName
	//
	// goverter:map Duration DurationMinutes
	// goverter:map TargetTop HighTarget
	// goverter:map TargetBottom LowTarget
	TreatmentToUpdateTempTarget(source Treatment, ctx UpdateContext) site.UpdateTempTargetParams

	//
	//
	//

	////////////////////////////////////////
	// DEVICE STATUS TO CREATE/UPDATE
	////////////////////////////////////////

	// goverter:context ctx
	// goverter:map Identifier ID
	// goverter:map Date OccurredAt
	// goverter:map UpdatedByIdentityName | AssignUpdatedByName
	//
	// goverter:ignore DeviceState DeviceChargingState DeviceTimestamp DeviceBatteryPercent
	DeviceStatusToUpdateDevice(source DeviceStatus, ctx UpdateContext) site.UpdateDeviceStateParams

	// goverter:context ctx
	// goverter:map Identifier ID
	// goverter:map Date OccurredAt
	// goverter:map UpdatedByIdentityName | AssignUpdatedByName
	//
	// goverter:ignore LoopState LoopTimestamp Iob IobBasal IobTreatment Cob
	DeviceStatusToUpdateLoop(source DeviceStatus, ctx UpdateContext) site.UpdateLoopStateParams

	// goverter:context ctx
	// goverter:map Identifier ID
	// goverter:map Date OccurredAt
	// goverter:map UpdatedByIdentityName | AssignUpdatedByName
	//
	// goverter:ignore PumpState PumpChargingState PumpTimestamp PumpInsertionTime PumpBatteryPercent PumpReservoir
	DeviceStatusToUpdatePump(source DeviceStatus, ctx UpdateContext) site.UpdatePumpStateParams
}

// //
// Custom Converter Methods
// //

// goverter:context ctx
func AssignUpdatedByName(ctx UpdateContext) *string {
	return ptr(ctx.UpdatedByIdentityName)
}
