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
	// goverter:map Duration DurationMinutes
	EntryToUpdate(source Entry, ctx UpdateContext) site.UpdateGlucoseReadingParams

	///////
	// FOOD TO FOOD PRESET
	///////

	// goverter:context ctx
	// goverter:map Identifier ID
	// goverter:map Date OccurredAt
	// goverter:map UpdatedByIdentityName | AssignUpdatedByName
	//
	// goverter:map Duration DurationMinutes
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
	// goverter:map Duration DurationMinutes
	TreatmentToUpdateFoodIntake(source Treatment, ctx UpdateContext) site.UpdateFoodIntakeParams

	///////
	// Glucose Reading
	///////

	// goverter:context ctx
	// goverter:map Identifier ID
	// goverter:map Date OccurredAt
	// goverter:map UpdatedByIdentityName | AssignUpdatedByName
	//
	// goverter:map Duration DurationMinutes
	TreatmentToUpdateGlucoseReading(source Treatment, ctx UpdateContext) site.UpdateGlucoseReadingParams

	///////
	// Insulin Dose
	///////

	// goverter:context ctx
	// goverter:map Identifier ID
	// goverter:map Date OccurredAt
	// goverter:map UpdatedByIdentityName | AssignUpdatedByName
	//
	// goverter:map Duration DurationMinutes
	TreatmentToUpdateInsulinDose(source Treatment, ctx UpdateContext) site.UpdateInsulinDoseParams

	///////
	// Notes
	///////

	// goverter:map  Identifier ID
	// goverter:map  Date OccurredAt
	// goverter:map  Notes Content
	// goverter:map UpdatedByIdentityName | AssignUpdatedByName
	// goverter:context ctx
	TreatmentToUpdateNote(source Treatment, ctx UpdateContext) site.UpdateNoteParams

	///////
	// Profile
	///////

	// goverter:context ctx
	// goverter:map Identifier ID
	// goverter:map Date OccurredAt
	// goverter:map UpdatedByIdentityName | AssignUpdatedByName
	//
	// goverter:map Duration DurationMinutes
	TreatmentToUpdateProfile(source Treatment, ctx UpdateContext) site.UpdateProfilePresetParams

	///////
	// Profile Switch
	///////

	// goverter:context ctx
	// goverter:map Identifier ID
	// goverter:map Date OccurredAt
	// goverter:map UpdatedByIdentityName | AssignUpdatedByName
	//
	// goverter:map Duration DurationMinutes
	TreatmentToUpdateProfileSwitch(source Treatment, ctx UpdateContext) site.UpdateProfileSwitchParams

	///////
	// Temp Basal
	///////

	// goverter:context ctx
	// goverter:map Identifier ID
	// goverter:map Date OccurredAt
	// goverter:map UpdatedByIdentityName | AssignUpdatedByName
	//
	// goverter:map Duration DurationMinutes
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
	TreatmentToUpdateTempTarget(source Treatment, ctx UpdateContext) site.UpdateTempTargetParams

	///////
	// Pump States
	///////

	// goverter:context ctx
	// goverter:map Identifier ID
	// goverter:map Date OccurredAt
	// goverter:map UpdatedByIdentityName | AssignUpdatedByName
	//
	// goverter:map Duration DurationMinutes
	TreatmentToUpdatePumpState(source Treatment, ctx UpdateContext) site.UpdatePumpStateParams

	///////
	// CGM States
	///////

	// goverter:context ctx
	// goverter:map Identifier ID
	// goverter:map Date OccurredAt
	// goverter:map UpdatedByIdentityName | AssignUpdatedByName
	//
	// goverter:map Duration DurationMinutes
	TreatmentToUpdateCGMState(source Treatment, ctx UpdateContext) site.UpdateCGMStateParams

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
	// goverter:map Duration DurationMinutes
	DeviceStatusToUpdateDevice(source DeviceStatus, ctx UpdateContext) site.UpdateDeviceStateParams

	// goverter:context ctx
	// goverter:map Identifier ID
	// goverter:map Date OccurredAt
	// goverter:map UpdatedByIdentityName | AssignUpdatedByName
	//
	// goverter:map Duration DurationMinutes
	DeviceStatusToUpdateLoop(source DeviceStatus, ctx UpdateContext) site.UpdateLoopStateParams

	// goverter:context ctx
	// goverter:map Identifier ID
	// goverter:map Date OccurredAt
	// goverter:map UpdatedByIdentityName | AssignUpdatedByName
	//
	// goverter:map Duration DurationMinutes
	DeviceStatusToUpdatePump(source DeviceStatus, ctx UpdateContext) site.UpdatePumpStateParams

	////////
	// SETTINGS
	////////

	// goverter:context ctx
	// goverter:map Identifier ID
	// goverter:map Date OccurredAt
	// goverter:map UpdatedByIdentityName | AssignUpdatedByName
	//
	// goverter:map Duration DurationMinutes
	SettingsToUpdatePump(source []Setting, ctx UpdateContext) site.UpdateSiteSettingsParams
}

// //
// Custom Converter Methods
// //

// goverter:context ctx
func AssignUpdatedByName(ctx UpdateContext) *string {
	return ptr(ctx.UpdatedByIdentityName)
}
