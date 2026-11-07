package nightscoutmodels

import (
	"time"

	"nighttroop.t1d.nz/data/sql/sitedb"
	datatypes "nighttroop.t1d.nz/data/types"
	"nighttroop.t1d.nz/utils"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/types"
)

type ConverterContext struct {
	IdentityID   int64
	IdentityName string
}

// goverter:converter
// goverter:enum:unknown @ignore
// goverter:output:format function
// goverter:output:file ./z_create_converters.go
// goverter:useZeroValueOnPointerInconsistency
// goverter:extend FlexibleFloat64ToFloat64
// goverter:extend FlexibleInt64ToInt64
// goverter:extend ShortStringToString
// goverter:extend SanitizedStringToString
// goverter:extend UnitsToUnits
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
	// goverter:map CreatedByIdentityID | AssignCreatedByID
	//
	ActivityToCreate(source Activity, ctx ConverterContext) sitedb.CreateActivityParams

	///////
	// ENTRIES TO GLUCOSE READINGS
	///////

	// goverter:context ctx
	// goverter:map Date OccurredAt | OrNow
	// goverter:map App CreatedByAppName
	// goverter:map Device CreatedByDeviceName
	// goverter:map ID                         | AssignID
	// goverter:map CreatedByIdentityName      | AssignCreatedByName
	// goverter:map CreatedByIdentityID | AssignCreatedByID
	//
	// goverter:ignore Notes Reason MeasurementMethod
	// goverter:map . ReadingType | SetReadingType
	// goverter:map Trend Delta
	// goverter:map . OtherAttributes | EntryOther
	EntryToCreate(source Entry, ctx ConverterContext) sitedb.CreateGlucoseReadingParams

	///////
	// FOOD TO FOOD PRESET
	///////

	// goverter:context ctx
	// goverter:map App CreatedByAppName
	// goverter:map Device CreatedByDeviceName
	// goverter:map ID                | AssignID
	// goverter:map CreatedByIdentityName  | AssignCreatedByName
	// goverter:map CreatedByIdentityID | AssignCreatedByID
	// goverter:map . OtherAttributes | FoodOther
	//
	FoodToCreate(source Food, ctx ConverterContext) sitedb.CreateFoodPresetParams

	///////
	// ProfileStore To Profile Preset
	///////

	// goverter:context ctx
	// goverter:map App CreatedByAppName
	// goverter:map Device CreatedByDeviceName
	// goverter:map ID                | AssignID
	// goverter:map CreatedByIdentityName  | AssignCreatedByName
	// goverter:map CreatedByIdentityID | AssignCreatedByID
	//
	// goverter:ignore Name CarbRatiosJson InsulinSensitivitiesJson BasalRatesJson TargetsHighJson TargetsLowJson Dia Timezone CarbsHr Delay Units IsDefault ProfileStoreID LoopSettingsJson RemoteSettingsJson
	// goverter:map Date OccurredAt | OrNow
	PartialProfileStoreToCreate(source ProfileStore, ctx ConverterContext) sitedb.CreateProfilePresetParams

	////////////////////////////////////////
	// DEVICE STATUS TO CREATE/UPDATE
	////////////////////////////////////////

	// goverter:context ctx
	// goverter:map Date OccurredAt | OrNow
	// goverter:map App CreatedByAppName
	// goverter:map Device CreatedByDeviceName
	// goverter:map ID                | AssignID
	// goverter:map CreatedByIdentityName  | AssignCreatedByName
	// goverter:map CreatedByIdentityID | AssignCreatedByID
	//
	// goverter:ignore LoopName LoopSettingsJson ConfigurationSettingsJson LoopPredictionsDefault LoopPredictionsZt LoopPredictionsStartTime IsDeviceCharging DeviceBatteryPercent PumpState PumpBatteryPercent PumpChargingState PumpReservoir PumpClock PumpManufacturer PumpModel OverrideJson LoopIob LoopIobBasal LoopIobTreatment LoopCob
	DeviceStatusToCreateDeviceReport(source DeviceStatus, ctx ConverterContext) sitedb.CreateDeviceReportParams
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

// AssignCreatedByID records which identity wrote the record, as its name can be reused once it is deleted.
// Background jobs write as an identity with no ID, which is stored as NULL.
// goverter:context ctx
func AssignCreatedByID(ctx ConverterContext) *int64 {
	if ctx.IdentityID == 0 {
		return nil
	}
	return &ctx.IdentityID
}

func ToDateInt(createdAt int64) *types.DateInt {
	x := types.DateInt(createdAt)
	return &x
}

func OrNow(d *types.DateInt) int64 {
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
	// A temp basal sent only as a percent has no rate, as that needs the profile's basal rate at the time
	return nil
}

// PercentToMultiplier is a temp basal's times the scheduled basal from Nightscout's percent change from it, so -50 is 0.5
func PercentToMultiplier(percent *types.FlexibleFloat64) *float64 {
	if percent == nil {
		return nil
	}
	return new(consts.Round(1+float64(*percent)/100, 5))
}

// MultiplierToPercent is Nightscout's percent change from the basal of a temp basal's times the scheduled basal
func MultiplierToPercent(multiplier *float64) *types.FlexibleFloat64 {
	if multiplier == nil {
		return nil
	}
	return new(types.FlexibleFloat64(consts.Round((*multiplier-1)*100, 3)))
}

func FlexibleFloat64ToFloat64(source types.FlexibleFloat64) float64 {
	return float64(source)
}

func FlexibleInt64ToInt64(source types.FlexibleInt64) int64 {
	return int64(source)
}
func ShortStringToString(source types.ShortString) string {
	return string(source)
}
func SanitizedStringToString(source types.SanitizedString) string {
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

func UnitsToUnits(source *consts.Units) *consts.Units {
	if source == nil {
		return nil
	}
	var constsUnits consts.Units
	switch *source {
	case consts.UnitsMGDL:
		constsUnits = consts.UnitsMGDL
	case consts.UnitsMMOLL:
		constsUnits = consts.UnitsMMOLL
	default: // ignored
	}
	return &constsUnits
}

// PercentageToMultiplier is a multiplier from Nightscout's percentage of a whole, so 120 is 1.2, as a profile switch's
// percentage of the profile's insulin and a combo bolus's split
func PercentageToMultiplier(percentage *types.FlexibleFloat64) *float64 {
	if percentage == nil {
		return nil
	}
	return new(consts.Round(float64(*percentage)/100, 5))
}

// MultiplierToPercentage is Nightscout's percentage of a whole from a multiplier, see PercentageToMultiplier
func MultiplierToPercentage(multiplier *float64) *types.FlexibleFloat64 {
	if multiplier == nil {
		return nil
	}
	return new(types.FlexibleFloat64(consts.Round(*multiplier*100, 3)))
}

// EntryOther is what only some apps send of an entry, kept in its reading's other attributes. It is written out, not
// generated, as a generated converter cannot be called from this package's own code, which goverter compiles first.
func EntryOther(source Entry) datatypes.GlucoseReadingOther {
	return datatypes.GlucoseReadingOther{Noise: source.Noise.Float64p(), Filtered: source.Filtered.Float64p(),
		Unfiltered: source.Unfiltered.Float64p(), Rssi: source.RSSI.Float64p(), IsCalibration: source.IsCalibration,
		SensorMgdlAtPairing: source.SensorMgdlAtPairing.Float64p()}
}

// EntryOtherJSON is what only some apps send of an entry, as JSON for an update to merge into its reading's other
// attributes
func EntryOtherJSON(source Entry) *string {
	return datatypes.OtherJSON(EntryOther(source))
}
