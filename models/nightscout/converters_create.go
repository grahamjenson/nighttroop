package nightscout

import (
	"time"

	"nighttroop.t1d.nz/data/sql/site"
	"nighttroop.t1d.nz/utils"
)

type ConverterContext struct {
	IdentityID   int64
	IdentityName string
}

// goverter:converter
// goverter:output:format function
// goverter:output:file ./z_create_converters.go
// goverter:useZeroValueOnPointerInconsistency
// goverter:extend FlexibleFloat64ToFloat64
// goverter:extend FlexibleInt64ToInt64
// goverter:extend ShortStringToString
// goverter:extend SanitizedStringToString
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
	ActivityToCreate(source Activity, ctx ConverterContext) site.CreateActivityParams

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
	// goverter:ignore Notes Reason MeasurementMethod
	// goverter:map . ReadingType | SetReadingType
	// goverter:map Trend Delta
	// goverter:map RSSI Rssi
	EntryToCreate(source Entry, ctx ConverterContext) site.CreateGlucoseReadingParams

	///////
	// FOOD TO FOOD PRESET
	///////

	// goverter:context ctx
	// goverter:map App CreatedByAppName
	// goverter:map Device CreatedByDeviceName
	// goverter:map ID                | AssignID
	// goverter:map CreatedByIdentityName  | AssignCreatedByName
	//
	FoodToCreate(source Food, ctx ConverterContext) site.CreateFoodPresetParams

	///////
	// ProfileStore To Profile Preset
	///////

	// goverter:context ctx
	// goverter:map App CreatedByAppName
	// goverter:map Device CreatedByDeviceName
	// goverter:map ID                | AssignID
	// goverter:map CreatedByIdentityName  | AssignCreatedByName
	//
	// goverter:ignore Name CarbRatiosJson InsulinSensitivitiesJson BasalRatesJson TargetsHighJson TargetsLowJson Dia Timezone Units IsDefault ProfileStoreID LoopSettingsJson
	// goverter:map Date OccurredAt | OrNow
	PartialProfileStoreToCreate(source ProfileStore, ctx ConverterContext) site.CreateProfilePresetParams

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
	// goverter:ignore LoopName LoopPredictionsDefault LoopPredictionsZt LoopPredictionsStartTime IsDeviceCharging DeviceBatteryPercent PumpState PumpBatteryPercent PumpChargingState PumpReservoir  LoopIob LoopIobBasal LoopIobTreatment LoopCob
	DeviceStatusToCreateDeviceReport(source DeviceStatus, ctx ConverterContext) site.CreateDeviceReportParams
}

// //
// Custom Converter Methods
// //

func AssignID() string {
	return utils.NextID()
}

// goverter:context ctx
func AssignCreatedByName(ctx ConverterContext) string {
	return ctx.IdentityName
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
func SanitizedStringToString(source SanitizedString) string {
	return string(source)
}
func SetReadingType(source Entry) string {
	if source.Type != nil {
		return source.Type.String()
	}
	if source.SGV != nil {
		return "sgv"
	}
	if source.MBG != nil {
		return "mbg"
	}
	return "sgv"
}
