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
	// goverter:map . CreatedAt | ModelDateStringFromDateIntAndUTCOffset
	// goverter:map CreatedByIdentityName Subject
	// goverter:map CreatedByAppName App
	// goverter:map CreatedByDeviceName Device
	// goverter:map ID Identifier
	// goverter:map OccurredAt Date
	//
	DBModelToActivity(source *sitedb.Activity, siteContext *httputils.SiteContext) Activity

	///////
	// ENTRIES TO GLUCOSE READINGS
	///////

	// goverter:context siteContext
	// goverter:ignore SysTime
	// goverter:map CreatedByIdentityName Subject
	// goverter:map CreatedByAppName App
	// goverter:map CreatedByDeviceName Device
	// goverter:map ID Identifier
	// goverter:map Mgdl SGV   | SGVGlucoseReadingTypeToValue
	// goverter:map Mgdl MBG   | MBGGlucoseReadingTypeToValue
	// goverter:map Mgdl Units | UnitsToValue
	// goverter:map OccurredAt Date
	// goverter:map OccurredAt Mills
	// goverter:map . DateString | ModelDateStringFromDateIntAndUTCOffset
	// goverter:map ReadingType Type
	// goverter:map Delta Trend
	// goverter:map Rssi RSSI
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
	// goverter:map . CreatedAt | ModelDateStringFromDateIntAndUTCOffset
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
	// goverter:ignore Unit GI HideAfterUse Hidden Position
	// goverter:map CreatedAt Date
	DBModelToFood(source *sitedb.FoodPreset, siteContext *httputils.SiteContext) Food

	///////
	// Profile To Profile Preset
	///////

	// goverter:context siteContext
	// goverter:map . StartDate | ModelDateStringFromDateIntAndUTCOffset
	// goverter:map CreatedByIdentityName Subject
	// goverter:map CreatedByIdentityName EnteredBy
	// goverter:map CreatedByAppName App
	// goverter:map CreatedByDeviceName Device
	// goverter:map ID Identifier
	// goverter:map OccurredAt Date
	//
	// goverter:ignore Store
	// goverter:map OccurredAt Mills
	// goverter:map Name DefaultProfile
	// goverter:map LoopSettingsJson LoopSettings | JSONToLoopSettings
	// goverter:map RemoteSettingsJson RemoteSettings | JSONToRemoteSettings
	DBModelToProfileStore(source *sitedb.ProfilePreset, siteContext *httputils.SiteContext) ProfileStore

	// goverter:context siteContext
	// goverter:map CarbRatiosJson           CarbRatio  | JSONToScheduleValueSlice
	// goverter:map InsulinSensitivitiesJson Sens       | JSONToScheduleValueSlice
	// goverter:map BasalRatesJson           Basal      | JSONToScheduleValueSlice
	// goverter:map TargetsHighJson          TargetHigh | JSONToScheduleValueSliceWithUnitConversionFromMGDL
	// goverter:map TargetsLowJson           TargetLow  | JSONToScheduleValueSliceWithUnitConversionFromMGDL
	DBModelToProfile(source *sitedb.ProfilePreset, siteContext *httputils.SiteContext) Profile
}

// //
// Custom Converter Methods
// //

func ModelDateStringFromDateIntAndUTCOffset(m interface{}) *types.DateString {
	if m == nil {
		return nil
	}

	switch v := m.(type) {
	case interface {
		GetOccurredAt() int64
		GetUtcOffsetMinutes() *int64
	}:
		ds := types.DateStringFromInt64AndUTCOffset(v.GetOccurredAt(), v.GetUtcOffsetMinutes())
		return &ds
	case interface {
		GetStart() int64
		GetUtcOffsetMinutes() *int64
	}:
		ds := types.DateStringFromInt64AndUTCOffset(v.GetStart(), v.GetUtcOffsetMinutes())
		return &ds
	}

	return nil
}

// goverter:context siteContext
func SGVGlucoseReadingTypeToValue(mgdl float64, siteContext *httputils.SiteContext) *types.FlexibleFloat64 {
	if true { // SGV is assumed for this function
		_, value := GetUnitsAndValue(mgdl, siteContext)
		return value
	}
	return nil

}

// goverter:context siteContext
func MBGGlucoseReadingTypeToValue(mgdl float64, siteContext *httputils.SiteContext) *types.FlexibleFloat64 {
	if true { // MBG is assumed for this function
		_, value := GetUnitsAndValue(mgdl, siteContext)
		return value
	}
	return nil

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
