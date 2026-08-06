package nightscout

import (
	"time"

	"nighttroop.t1d.nz/data/sql/site"
	"nighttroop.t1d.nz/utils"
)

type CreateContext struct {
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
	// goverter:map Date OccurredAt | OrNow
	// goverter:map App CreatedByAppName
	// goverter:map Device CreatedByDeviceName
	// goverter:map ID                | AssignID
	// goverter:map CreatedByIdentityName  | AssignCreatedByName
	//
	// goverter:map Duration DurationMinutes
	ActivityToCreate(source Activity, ctx CreateContext) site.CreateActivityParams

	///////
	// ENTRIES TO GLUCOSE READINGS
	///////

	// goverter:context ctx
	// goverter:map Date OccurredAt | OrNow
	// goverter:map App CreatedByAppName
	// goverter:map Device CreatedByDeviceName
	// goverter:map ID                         | AssignID
	// goverter:map CreatedByIdentityName      | AssignCreatedByName
	//
	// goverter:map Type ReadingType
	EntryToCreate(source Entry, ctx CreateContext) site.CreateGlucoseReadingParams

	///////
	// FOOD TO FOOD PRESET
	///////

	// goverter:context ctx
	// goverter:map App CreatedByAppName
	// goverter:map Device CreatedByDeviceName
	// goverter:map ID                | AssignID
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
	// goverter:map ID                | AssignID
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
	// goverter:map Date OccurredAt | OrNow
	// goverter:map App CreatedByAppName
	// goverter:map Device CreatedByDeviceName
	// goverter:map ID                | AssignID
	// goverter:map CreatedByIdentityName  | AssignCreatedByName
	//
	// goverter:ignore  GlucoseReadingID InsulinDoseID FoodIntakeID TempBasalID TempTargetID ProfileSwitchID DeviceStateID LoopStateID PumpStateID CgmStateID ActivityID NoteID
	// goverter:map EventType GroupType
	TreatmentToCreateEventGroup(source Treatment, ctx CreateContext) site.CreateEventGroupParams

	///////
	// Food Intake
	///////

	// goverter:context ctx
	// goverter:map Date OccurredAt | OrNow
	// goverter:map App CreatedByAppName
	// goverter:map Device CreatedByDeviceName
	// goverter:map ID                | AssignID
	// goverter:map CreatedByIdentityName  | AssignCreatedByName
	//
	// goverter:ignore Name FoodType
	// goverter:map Energy Calories
	TreatmentToCreateFoodIntake(source Treatment, ctx CreateContext) site.CreateFoodIntakeParams

	///////
	// Glucose Reading
	///////

	// goverter:context ctx
	// goverter:map Date OccurredAt | OrNow
	// goverter:map App CreatedByAppName
	// goverter:map Device CreatedByDeviceName
	// goverter:map ID                | AssignID
	// goverter:map CreatedByIdentityName  | AssignCreatedByName
	//
	// goverter:map Glucose Reading
	// goverter:map GlucoseType ReadingType
	TreatmentToCreateGlucoseReading(source Treatment, ctx CreateContext) site.CreateGlucoseReadingParams

	///////
	// Insulin Dose
	///////

	// goverter:context ctx
	// goverter:map Date OccurredAt | OrNow
	// goverter:map App CreatedByAppName
	// goverter:map Device CreatedByDeviceName
	// goverter:map ID                | AssignID
	// goverter:map CreatedByIdentityName  | AssignCreatedByName
	//
	// goverter:ignore DoseType
	// goverter:map Insulin Dose
	TreatmentToCreateInsulinDose(source Treatment, ctx CreateContext) site.CreateInsulinDoseParams

	///////
	// Notes
	///////

	// goverter:context ctx
	// goverter:map Date OccurredAt | OrNow
	// goverter:map App CreatedByAppName
	// goverter:map Device CreatedByDeviceName
	// goverter:map ID                | AssignID
	// goverter:map CreatedByIdentityName  | AssignCreatedByName
	//
	// goverter:map Notes Content
	// goverter:map Reason Reason
	TreatmentToCreateNote(source Treatment, ctx CreateContext) site.CreateNoteParams

	///////
	// Profile
	///////

	// goverter:context ctx
	// goverter:map App CreatedByAppName
	// goverter:map Device CreatedByDeviceName
	// goverter:map ID                | AssignID
	// goverter:map CreatedByIdentityName  | AssignCreatedByName
	//
	// goverter:ignore CarbRatiosJson InsulinSensitivitiesJson BasalRatesJson TargetsJson
	TreatmentToCreateProfile(source Treatment, ctx CreateContext) site.CreateProfilePresetParams

	///////
	// Profile Switch
	///////

	// goverter:context ctx
	// goverter:map Date OccurredAt | OrNow
	// goverter:map App CreatedByAppName
	// goverter:map Device CreatedByDeviceName
	// goverter:map ID                | AssignID
	// goverter:map CreatedByIdentityName  | AssignCreatedByName
	//
	// goverter:ignore CarbRatiosJson InsulinSensitivitiesJson BasalRatesJson TargetsJson
	// goverter:map Duration DurationMinutes
	// goverter:map Percentage Percent
	TreatmentToCreateProfileSwitch(source Treatment, ctx CreateContext) site.CreateProfileSwitchParams

	///////
	// Temp Basal
	///////

	// goverter:context ctx
	// goverter:map Date OccurredAt | OrNow
	// goverter:map App CreatedByAppName
	// goverter:map Device CreatedByDeviceName
	// goverter:map ID                | AssignID
	// goverter:map CreatedByIdentityName  | AssignCreatedByName
	//
	// goverter:map Duration DurationMinutes
	// goverter:map Percentage Percent
	TreatmentToCreateTempBasal(source Treatment, ctx CreateContext) site.CreateTempBasalParams

	///////
	// Temp Targets
	///////

	// goverter:context ctx
	// goverter:map Date OccurredAt | OrNow
	// goverter:map App CreatedByAppName
	// goverter:map Device CreatedByDeviceName
	// goverter:map ID                | AssignID
	// goverter:map CreatedByIdentityName  | AssignCreatedByName
	//
	// goverter:map Duration DurationMinutes
	// goverter:map TargetTop HighTarget
	// goverter:map TargetBottom LowTarget
	TreatmentToCreateTempTarget(source Treatment, ctx CreateContext) site.CreateTempTargetParams

	// goverter:context ctx
	// goverter:map Date OccurredAt | OrNow
	// goverter:map App CreatedByAppName
	// goverter:map Device CreatedByDeviceName
	// goverter:map ID                | AssignID
	// goverter:map CreatedByIdentityName  | AssignCreatedByName
	//
	// goverter:ignore DeviceState DeviceTimestamp DeviceBatteryPercent DeviceChargingState
	TreatmentToCreateDeviceState(source Treatment, ctx CreateContext) site.CreateDeviceStateParams

	// goverter:context ctx
	// goverter:map Date OccurredAt | OrNow
	// goverter:map App CreatedByAppName
	// goverter:map Device CreatedByDeviceName
	// goverter:map ID                | AssignID
	// goverter:map CreatedByIdentityName  | AssignCreatedByName
	//
	// goverter:ignore LoopState LoopTimestamp Iob IobBasal IobTreatment Cob
	TreatmentToCreateLoopState(source Treatment, ctx CreateContext) site.CreateLoopStateParams

	// goverter:context ctx
	// goverter:map Date OccurredAt | OrNow
	// goverter:map App CreatedByAppName
	// goverter:map Device CreatedByDeviceName
	// goverter:map ID                | AssignID
	// goverter:map CreatedByIdentityName  | AssignCreatedByName
	//
	// goverter:ignore PumpState PumpTimestamp PumpInsertionTime PumpReservoir PumpBatteryPercent PumpChargingState
	TreatmentToCreatePumpState(source Treatment, ctx CreateContext) site.CreatePumpStateParams

	// goverter:context ctx
	// goverter:map Date OccurredAt | OrNow
	// goverter:map App CreatedByAppName
	// goverter:map Device CreatedByDeviceName
	// goverter:map ID                | AssignID
	// goverter:map CreatedByIdentityName  | AssignCreatedByName
	//
	// goverter:ignore CgmState CgmTimestamp CgmInsertionTime
	TreatmentToCreateCgmState(source Treatment, ctx CreateContext) site.CreateCGMStateParams

	// goverter:context ctx
	// goverter:map Date OccurredAt | OrNow
	// goverter:map App CreatedByAppName
	// goverter:map Device CreatedByDeviceName
	// goverter:map ID                | AssignID
	// goverter:map CreatedByIdentityName  | AssignCreatedByName
	//
	// goverter:ignore
	// goverter:map Duration DurationMinutes
	TreatmentToCreateActivity(source Treatment, ctx CreateContext) site.CreateActivityParams

	//
	//
	//

	////////////////////////////////////////
	// DEVICE STATUS TO CREATE/UPDATE
	////////////////////////////////////////

	// goverter:context ctx
	// goverter:map Date OccurredAt | OrNow
	// goverter:map App CreatedByAppName
	// goverter:map Device CreatedByDeviceName
	// goverter:map ID                | AssignID
	// goverter:map CreatedByIdentityName  | AssignCreatedByName
	//
	// goverter:ignore GroupType GlucoseReadingID InsulinDoseID FoodIntakeID TempBasalID TempTargetID ProfileSwitchID DeviceStateID LoopStateID PumpStateID CgmStateID ActivityID NoteID
	DeviceStatusToCreateEventGroup(source DeviceStatus, ctx CreateContext) site.CreateEventGroupParams

	// goverter:context ctx
	// goverter:map Date OccurredAt | OrNow
	// goverter:map App CreatedByAppName
	// goverter:map Device CreatedByDeviceName
	// goverter:map ID                | AssignID
	// goverter:map CreatedByIdentityName  | AssignCreatedByName
	//
	// goverter:ignore DeviceState DeviceChargingState
	// goverter:map Date DeviceTimestamp
	// goverter:map UploaderBattery DeviceBatteryPercent
	DeviceStatusToCreateDevice(source DeviceStatus, ctx CreateContext) site.CreateDeviceStateParams

	// goverter:context ctx
	// goverter:map Date OccurredAt | OrNow
	// goverter:map App CreatedByAppName
	// goverter:map Device CreatedByDeviceName
	// goverter:map ID                | AssignID
	// goverter:map CreatedByIdentityName  | AssignCreatedByName
	//
	// goverter:ignore LoopState LoopTimestamp IobTreatment
	// goverter:map MergedLoopStatus.IOB.IOB Iob
	// goverter:map MergedLoopStatus.IOB.BasalIOB IobBasal
	// goverter:map MergedLoopStatus.COB.COB Cob
	DeviceStatusToCreateLoop(source DeviceStatus, ctx CreateContext) site.CreateLoopStateParams

	// goverter:context ctx
	// goverter:map Date OccurredAt | OrNow
	// goverter:map App CreatedByAppName
	// goverter:map Device CreatedByDeviceName
	// goverter:map ID                | AssignID
	// goverter:map CreatedByIdentityName  | AssignCreatedByName
	//
	// goverter:ignore PumpState PumpChargingState PumpTimestamp PumpInsertionTime PumpBatteryPercent
	// goverter:map Pump.Reservoir PumpReservoir
	DeviceStatusToCreatePump(source DeviceStatus, ctx CreateContext) site.CreatePumpStateParams
}

// //
// Custom Converter Methods
// //

func AssignID() string {
	return utils.NextID()
}

// goverter:context ctx
func AssignCreatedByName(ctx CreateContext) string {
	return ctx.CreatedByIdentityName
}

func ToDateInt(createdAt int64) *DateInt {
	x := DateInt(createdAt)
	return &x
}

func OrNow(d *DateInt) int64 {
	if d != nil {
		return int64(*d)
	}
	return time.Now().UnixMilli()
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

func FlexibleFloat64ToFloat64(source FlexibleFloat64) float64 {
	return float64(source)
}

func FlexibleInt64ToInt64(source FlexibleInt64) int64 {
	return int64(source)
}
func ShortStringToString(source ShortString) string {
	return string(source)
}
