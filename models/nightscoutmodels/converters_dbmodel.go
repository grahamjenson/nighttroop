package nightscoutmodels

import (
	"encoding/json"

	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/httputils"
	"nighttroop.t1d.nz/utils/types"
)

// goverter:converter
// goverter:output:format function
// goverter:enum:unknown @ignore
// goverter:output:file ./z_dbmodel_converters.go
// goverter:useZeroValueOnPointerInconsistency
type DBModelConverter interface {

	///////
	// ACTIVITIES
	///////

	// goverter:context siteContext
	// goverter:map CreatedByIdentityName Subject
	// goverter:map CreatedByAppName App
	// goverter:map CreatedByDeviceName Device
	// goverter:map ID Identifier
	// goverter:map OccurredAt Date
	DBModelToActivity(source *sitedb.Activity, siteContext *httputils.SiteContext) Activity

	///////
	// ENTRIES TO GLUCOSE READINGS
	///////

	// goverter:context siteContext
	// goverter:map CreatedByIdentityName Subject
	// goverter:map CreatedByAppName App
	// goverter:map CreatedByDeviceName Device
	// goverter:map ID Identifier
	// goverter:map . SGV   | SGVGlucoseReadingTypeToValue
	// goverter:map . MBG   | MBGGlucoseReadingTypeToValue
	// goverter:map Mgdl Units | UnitsToValue
	// goverter:map OccurredAt Date
	// goverter:map ReadingType Type
	// goverter:map Rssi RSSI
	// goverter:map Trend Direction | TrendDirection
	// goverter:map Trend | EntryTrend
	// goverter:map Noise | WholeNoise
	DBModelToEntry(source *sitedb.GlucoseReading, siteContext *httputils.SiteContext) Entry

	///////
	// DEVICE REPORT TO DEVICE STATUS
	///////

	// goverter:context siteContext
	// goverter:ignore Uploader Loop Pump OpenAPS
	// goverter:map CreatedByIdentityName Subject
	// goverter:map CreatedByAppName App
	// goverter:map CreatedByDeviceName Device
	// goverter:map ID Identifier
	// goverter:map OccurredAt Date
	//
	// goverter:ignore IsCharging Uploader UploaderBattery Pump OpenAPS Loop OpenAPSRaw LoopRaw ConfigurationRaw Configuration OverrideRaw
	DBModelToDeviceStatus(source *sitedb.DeviceReport, siteContext *httputils.SiteContext) DeviceStatus

	///////
	// FOOD TO FOOD PRESET
	///////

	// goverter:context siteContext
	// goverter:map CreatedByIdentityName Subject
	// goverter:map CreatedByAppName App
	// goverter:map CreatedByDeviceName Device
	// goverter:map ID Identifier
	//
	// goverter:map Gi GI
	// goverter:map CreatedAt Date
	// goverter:map . Type | FoodDocumentType
	DBModelToFood(source *sitedb.FoodPreset, siteContext *httputils.SiteContext) Food

	///////
	// Profile To Profile Preset
	///////

	// goverter:context siteContext
	// goverter:map CreatedByIdentityName Subject
	// goverter:map . EnteredBy | ProfilePresetEnteredBy
	// goverter:map CreatedByAppName App
	// goverter:map CreatedByDeviceName Device
	// goverter:map ID Identifier
	// goverter:map OccurredAt Date
	//
	// goverter:ignore Store
	// goverter:map Name DefaultProfile
	// goverter:ignore LoopSettings RemoteSettings
	DBModelToProfileStore(source *sitedb.ProfilePreset, siteContext *httputils.SiteContext) ProfileStore

	// goverter:context siteContext
	// goverter:map CarbRatiosJson           CarbRatio  | JSONToScheduleValueSlice
	// goverter:map .                        Sens       | ProfilePresetSensitivities
	// goverter:map BasalRatesJson           Basal      | JSONToScheduleValueSlice
	// goverter:map TargetsHighJson          TargetHigh | JSONToScheduleValueSliceWithUnitConversionFromMGDL
	// goverter:map TargetsLowJson           TargetLow  | JSONToScheduleValueSliceWithUnitConversionFromMGDL
	DBModelToProfile(source *sitedb.ProfilePreset, siteContext *httputils.SiteContext) Profile
}

// //
// Custom Converter Methods
// //

// SGVGlucoseReadingTypeToValue is the sgv of any reading but a meter's. As in Nightscout an entry has one of sgv and
// mbg, as followers like Nightguard take any entry with an sgv as a CGM reading
// goverter:context siteContext
func SGVGlucoseReadingTypeToValue(source *sitedb.GlucoseReading, siteContext *httputils.SiteContext) *types.FlexibleFloat64 {
	if source.ReadingType == string(MBGType) {
		return nil
	}
	_, value := GetUnitsAndValue(source.Mgdl, siteContext)
	return value
}

// MBGGlucoseReadingTypeToValue is the mbg of a meter reading only
// goverter:context siteContext
func MBGGlucoseReadingTypeToValue(source *sitedb.GlucoseReading, siteContext *httputils.SiteContext) *types.FlexibleFloat64 {
	if source.ReadingType != string(MBGType) {
		return nil
	}
	_, value := GetUnitsAndValue(source.Mgdl, siteContext)
	return value
}

// goverter:context siteContext
func GlucoseReadingTypeToValue(mgdl float64, siteContext *httputils.SiteContext) *types.FlexibleFloat64 {
	_, value := GetUnitsAndValue(mgdl, siteContext)
	return value
}

// goverter:context siteContext
func UnitsToValue(mgdl float64, siteContext *httputils.SiteContext) *consts.Units {
	units, _ := GetUnitsAndValue(mgdl, siteContext)
	return units
}

func GetUnitsAndValue(mgdl float64, siteContext *httputils.SiteContext) (*consts.Units, *types.FlexibleFloat64) {

	units := siteContext.GetSettings().GetUnits()
	value := units.FromMGDL(mgdl)
	return units, new(types.FlexibleFloat64(value))
}

func JSONToScheduleValueSlice(jsonStr *string) []*ScheduleValue {
	if jsonStr == nil {
		return nil
	}
	var result []*ScheduleValue
	err := json.Unmarshal([]byte(*jsonStr), &result)
	if err != nil {
		return nil
	}
	for _, sv := range result {
		if sv == nil {
			continue
		}
		if sv.TimeAsSeconds == nil && sv.Time != nil {
			sv.TimeAsSeconds = sv.Time.ToSeconds()
		}
	}
	return result
}

// goverter:context siteContext
func JSONToScheduleValueSliceWithUnitConversionFromMGDL(jsonStr *string, siteContext *httputils.SiteContext) []*ScheduleValue {
	if jsonStr == nil {
		return nil
	}
	units := siteContext.GetSettings().GetUnits()

	var result []*ScheduleValue
	err := json.Unmarshal([]byte(*jsonStr), &result)
	if err != nil {
		return nil
	}
	for _, sv := range result {
		if sv == nil {
			continue
		}

		if sv.Value != nil {
			sv.Value = new(types.FlexibleFloat64(units.FromMGDL(sv.Value.Float64())))
		}

		if sv.TimeAsSeconds == nil && sv.Time != nil {
			sv.TimeAsSeconds = sv.Time.ToSeconds()
		}
	}
	return result
}

// ProfilePresetEnteredBy is who the app said uploaded a profile, e.g. Loop, as tools pick a loop's profiles by it, else
// the app that sent it, as AAPS names it, else the identity that uploaded it, as a treatment's is
func ProfilePresetEnteredBy(source *sitedb.ProfilePreset) *types.ShortString {
	if source.EnteredBy != nil {
		return (*types.ShortString)(source.EnteredBy)
	}
	if source.CreatedByAppName != nil && *source.CreatedByAppName != "" {
		return (*types.ShortString)(source.CreatedByAppName)
	}
	return (*types.ShortString)(&source.CreatedByIdentityName)
}

// ProfilePresetSensitivities are a profile's sensitivities, stored in its own units, in the settings' units
// goverter:context siteContext
func ProfilePresetSensitivities(source *sitedb.ProfilePreset, siteContext *httputils.SiteContext) []*ScheduleValue {
	return SensitivitiesFromJSON(source.InsulinSensitivitiesJson, source.Units, siteContext)
}

// SensitivitiesFromJSON are sensitivities, stored in units, in the settings' units. Without units they are as stored,
// as a sensitivity's size cannot tell mg/dL from mmol/L.
func SensitivitiesFromJSON(jsonStr *string, units *consts.Units, siteContext *httputils.SiteContext) []*ScheduleValue {
	result := JSONToScheduleValueSlice(jsonStr)
	if units == nil {
		return result
	}
	settingsUnits := siteContext.GetSettings().GetUnits()
	for _, sv := range result {
		if sv != nil && sv.Value != nil {
			sv.Value = new(types.FlexibleFloat64(settingsUnits.FromMGDL(units.ToMGDL(sv.Value.Float64()))))
		}
	}
	return result
}

func JSONToRemoteSettings(jsonStr *string) *RemoteSettings {
	if jsonStr == nil {
		return nil
	}
	var result RemoteSettings
	if err := json.Unmarshal([]byte(*jsonStr), &result); err != nil {
		return nil
	}
	return &result
}

func JSONToLoopSettings(jsonStr *string) *LoopSettings {
	if jsonStr == nil {
		return nil
	}
	var result LoopSettings
	err := json.Unmarshal([]byte(*jsonStr), &result)
	if err != nil {
		return nil
	}
	return &result
}
