package nightscout

import (
	"nighttroop.t1d.nz/data/sql/site"
	"nighttroop.t1d.nz/utils"
)

type CreateContext struct {
	IDPrefix              string
	CreatedByIdentityID   int64
	CreatedByIdentityName string
}

// goverter:converter
// goverter:output:format function
// goverter:output:file ./z_create_converters.go
// goverter:useZeroValueOnPointerInconsistency
// goverter:extend FlexibleFloat64ToFloat64
// goverter:extend FlexibleInt64ToInt64
// goverter:extend ShortStringToString
type CreateMapConverter interface {

	///////
	// ACTIVITIES
	///////

	// goverter:context ctx
	// goverter:map Date OccurredAt
	// goverter:map App CreatedByAppName
	// goverter:map Device CreatedByDeviceName
	// goverter:map Date ID                | AssignID
	// goverter:map CreatedByIdentityName  | AssignCreatedByName
	//
	// goverter:map Duration DurationMinutes
	ActivityToCreate(source Activity, ctx CreateContext) site.CreateActivityParams

	///////
	// ENTRIES TO GLUCOSE READINGS
	///////

	// goverter:context ctx
	// goverter:map Date OccurredAt
	// goverter:map App CreatedByAppName
	// goverter:map Device CreatedByDeviceName
	// goverter:map Date ID                | AssignID
	// goverter:map CreatedByIdentityName  | AssignCreatedByName
	//
	// goverter:map Type ReadingType
	EntryToCreate(source Entry, ctx CreateContext) site.CreateGlucoseReadingParams

	///////
	// FOOD TO FOOD PRESET
	///////

	// goverter:context ctx
	// goverter:map App CreatedByAppName
	// goverter:map Device CreatedByDeviceName
	// goverter:map Date ID                | AssignID
	// goverter:map CreatedByIdentityName  | AssignCreatedByName
	//
	// goverter:map Food FoodType
	// goverter:map Energy Calories
	FoodToCreate(source Food, ctx CreateContext) site.CreateFoodPresetParams

	///////
	// Profile To Profile Preset
	///////

	// goverter:context ctx
	// goverter:map App CreatedByAppName
	// goverter:map Device CreatedByDeviceName
	// goverter:map Date ID                | AssignID
	// goverter:map CreatedByIdentityName  | AssignCreatedByName
	//
	// goverter:map Duration DurationMinutes
	// ProfileToCreate(source Profile, ctx CreateContext) []site.CreateProfilePresetParams

	//
	//
	//

	////////////////////////////////////////
	// TREATMENT TO CREATE/UPDATE
	////////////////////////////////////////

	// goverter:context ctx
	// goverter:map Date OccurredAt
	// goverter:map App CreatedByAppName
	// goverter:map Device CreatedByDeviceName
	// goverter:map Date ID                | AssignID
	// goverter:map CreatedByIdentityName  | AssignCreatedByName
	//
	// goverter:ignore GroupType GlucoseReadingID InsulinDoseID FoodIntakeID TempBasalID TempTargetID ProfileSwitchID DeviceStateID LoopStateID PumpStateID CgmStateID ActivityID NoteID
	TreatmentToCreateEventGroup(source Treatment, ctx CreateContext) site.CreateEventGroupParams

	///////
	// Food Intake
	///////

	// goverter:context ctx
	// goverter:map Date OccurredAt
	// goverter:map App CreatedByAppName
	// goverter:map Device CreatedByDeviceName
	// goverter:map Date ID                | AssignID
	// goverter:map CreatedByIdentityName  | AssignCreatedByName
	//
	// goverter:ignore Name FoodType
	// goverter:map Energy Calories
	TreatmentToCreateFoodIntake(source Treatment, ctx CreateContext) site.CreateFoodIntakeParams

	///////
	// Glucose Reading
	///////

	// goverter:context ctx
	// goverter:map Date OccurredAt
	// goverter:map App CreatedByAppName
	// goverter:map Device CreatedByDeviceName
	// goverter:map Date ID                | AssignID
	// goverter:map CreatedByIdentityName  | AssignCreatedByName
	//
	// goverter:map Glucose Reading
	// goverter:map GlucoseType ReadingType
	TreatmentToCreateGlucoseReading(source Treatment, ctx CreateContext) site.CreateGlucoseReadingParams

	///////
	// Insulin Dose
	///////

	// goverter:context ctx
	// goverter:map Date OccurredAt
	// goverter:map App CreatedByAppName
	// goverter:map Device CreatedByDeviceName
	// goverter:map Date ID                | AssignID
	// goverter:map CreatedByIdentityName  | AssignCreatedByName
	//
	// goverter:ignore DoseType
	// goverter:map Insulin Dose
	TreatmentToCreateInsulinDose(source Treatment, ctx CreateContext) site.CreateInsulinDoseParams

	///////
	// Notes
	///////

	// goverter:context ctx
	// goverter:map Date OccurredAt
	// goverter:map App CreatedByAppName
	// goverter:map Device CreatedByDeviceName
	// goverter:map Date ID                | AssignID
	// goverter:map CreatedByIdentityName  | AssignCreatedByName
	//
	// goverter:map Notes Content
	TreatmentToCreateNote(source Treatment, ctx CreateContext) site.CreateNoteParams

	///////
	// Profile
	///////

	// // goverter:context ctx
	// // goverter:map Date OccurredAt
	// // goverter:map App CreatedByAppName
	// // goverter:map Device CreatedByDeviceName
	// // goverter:map Date ID                | AssignID
	// // goverter:map CreatedByIdentityName  | AssignCreatedByName
	// //
	// // goverter:ignore Name
	// TreatmentToCreateProfile(source Treatment, ctx CreateContext) site.CreateProfilePresetParams

	///////
	// Profile Switch
	///////

	// goverter:context ctx
	// goverter:map Date OccurredAt
	// goverter:map App CreatedByAppName
	// goverter:map Device CreatedByDeviceName
	// goverter:map Date ID                | AssignID
	// goverter:map CreatedByIdentityName  | AssignCreatedByName
	//
	// goverter:map Duration DurationMinutes
	// TreatmentToCreateProfileSwitch(source Treatment, ctx CreateContext) site.CreateProfileSwitchParams

	///////
	// Temp Basal
	///////

	// goverter:context ctx
	// goverter:map Date OccurredAt
	// goverter:map App CreatedByAppName
	// goverter:map Device CreatedByDeviceName
	// goverter:map Date ID                | AssignID
	// goverter:map CreatedByIdentityName  | AssignCreatedByName
	//
	// goverter:map Duration DurationMinutes
	// goverter:map Percentage Percent
	TreatmentToCreateTempBasal(source Treatment, ctx CreateContext) site.CreateTempBasalParams

	///////
	// Temp Targets
	///////

	// goverter:context ctx
	// goverter:map Date OccurredAt
	// goverter:map App CreatedByAppName
	// goverter:map Device CreatedByDeviceName
	// goverter:map Date ID                | AssignID
	// goverter:map CreatedByIdentityName  | AssignCreatedByName
	//
	// goverter:map Duration DurationMinutes
	// goverter:map TargetTop HighTarget
	// goverter:map TargetBottom LowTarget
	TreatmentToCreateTempTarget(source Treatment, ctx CreateContext) site.CreateTempTargetParams

	//
	//
	//

	////////////////////////////////////////
	// DEVICE STATUS TO CREATE/UPDATE
	////////////////////////////////////////

	// goverter:context ctx
	// goverter:map Date OccurredAt
	// goverter:map App CreatedByAppName
	// goverter:map Device CreatedByDeviceName
	// goverter:map Date ID                | AssignID
	// goverter:map CreatedByIdentityName  | AssignCreatedByName
	//
	// goverter:ignore GroupType GlucoseReadingID InsulinDoseID FoodIntakeID TempBasalID TempTargetID ProfileSwitchID DeviceStateID LoopStateID PumpStateID CgmStateID ActivityID NoteID
	DeviceStatusToCreateEventGroup(source DeviceStatus, ctx CreateContext) site.CreateEventGroupParams

	// goverter:context ctx
	// goverter:map Date OccurredAt
	// goverter:map App CreatedByAppName
	// goverter:map Device CreatedByDeviceName
	// goverter:map Date ID                | AssignID
	// goverter:map CreatedByIdentityName  | AssignCreatedByName
	//
	// goverter:ignore DeviceState DeviceChargingState
	// goverter:map Date DeviceTimestamp
	// goverter:map UploaderBattery DeviceBatteryPercent
	DeviceStatusToCreateDevice(source DeviceStatus, ctx CreateContext) site.CreateDeviceStateParams

	// goverter:context ctx
	// goverter:map Date OccurredAt
	// goverter:map App CreatedByAppName
	// goverter:map Device CreatedByDeviceName
	// goverter:map Date ID                | AssignID
	// goverter:map CreatedByIdentityName  | AssignCreatedByName
	//
	// goverter:ignore LoopState LoopTimestamp IobTreatment
	// goverter:map MergedLoopStatus.IOB.IOB Iob
	// goverter:map MergedLoopStatus.IOB.BasalIOB IobBasal
	// goverter:map MergedLoopStatus.COB.COB Cob
	DeviceStatusToCreateLoop(source DeviceStatus, ctx CreateContext) site.CreateLoopStateParams

	// goverter:context ctx
	// goverter:map Date OccurredAt
	// goverter:map App CreatedByAppName
	// goverter:map Device CreatedByDeviceName
	// goverter:map Date ID                | AssignID
	// goverter:map CreatedByIdentityName  | AssignCreatedByName
	//
	// goverter:ignore PumpState PumpChargingState PumpTimestamp PumpInsertionTime PumpBatteryPercent
	// goverter:map Pump.Reservoir PumpReservoir
	DeviceStatusToCreatePump(source DeviceStatus, ctx CreateContext) site.CreatePumpStateParams
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

func ToDateInt(createdAt int64) *DateInt {
	x := DateInt(createdAt)
	return &x
}

// ////
// GlucoseReading
// ////

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

func FlexibleFloat64ToFloat64(source FlexibleFloat64) float64 {
	return float64(source)
}

func FlexibleInt64ToInt64(source FlexibleInt64) int64 {
	return int64(source)
}
func ShortStringToString(source ShortString) string {
	return string(source)
}
