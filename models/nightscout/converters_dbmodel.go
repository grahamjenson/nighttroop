package nightscout

import (
	"encoding/json"

	"nighttroop.t1d.nz/data/sql/site"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/httputils"
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
	DBModelToActivity(source *site.Activity, siteContext *httputils.SiteContext) Activity

	///////
	// ENTRIES TO GLUCOSE READINGS
	///////

	// goverter:context siteContext
	// goverter:ignore SysTime
	// goverter:map CreatedByIdentityName Subject
	// goverter:map CreatedByAppName App
	// goverter:map CreatedByDeviceName Device
	// goverter:map ID Identifier
	// goverter:map . SGV   | SGVGlucoseReadingTypeToValue
	// goverter:map . MBG   | MBGGlucoseReadingTypeToValue
	// goverter:map . Units | UnitsToValue
	// goverter:map OccurredAt Date
	// goverter:map OccurredAt Mills
	// goverter:map . DateString | ModelDateStringFromDateIntAndUTCOffset
	// goverter:map ReadingType Type
	// goverter:map Delta Trend
	// goverter:map Rssi RSSI
	DBModelToEntry(source *site.GlucoseReading, siteContext *httputils.SiteContext) Entry

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
	// goverter:ignore IsCharging Uploader UploaderBattery Pump OpenAPS Loop OpenAPSRaw LoopRaw ConfigurationRaw Configuration
	DBModelToDeviceStatus(source *site.DeviceReport, siteContext *httputils.SiteContext) DeviceStatus

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
	DBModelToFood(source *site.FoodPreset, siteContext *httputils.SiteContext) Food

	///////
	// Profile To Profile Preset
	///////

	// goverter:context siteContext
	// goverter:map . StartDate | ModelDateStringFromDateIntAndUTCOffset
	// goverter:map CreatedByIdentityName Subject
	// goverter:map CreatedByAppName App
	// goverter:map CreatedByDeviceName Device
	// goverter:map ID Identifier
	// goverter:map OccurredAt Date
	//
	// goverter:ignore Store
	// goverter:map OccurredAt Mills
	// goverter:map Name DefaultProfile
	// goverter:map LoopSettingsJson LoopSettings | JSONToLoopSettings
	DBModelToProfileStore(source *site.ProfilePreset, siteContext *httputils.SiteContext) ProfileStore

	// goverter:context siteContext
	// goverter:map CarbRatiosJson           CarbRatio  | JSONToScheduleValueSlice
	// goverter:map InsulinSensitivitiesJson Sens       | JSONToScheduleValueSlice
	// goverter:map BasalRatesJson           Basal      | JSONToScheduleValueSlice
	// goverter:map TargetsHighJson          TargetHigh | JSONToScheduleValueSlice
	// goverter:map TargetsLowJson           TargetLow  | JSONToScheduleValueSlice
	DBModelToProfile(source *site.ProfilePreset, siteContext *httputils.SiteContext) Profile
}

// //
// Custom Converter Methods
// //

func ModelDateStringFromDateIntAndUTCOffset(m interface{}) *DateString {
	if m == nil {
		return nil
	}

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
	}

	return nil
}

// goverter:context siteContext
func SGVGlucoseReadingTypeToValue(gr site.GlucoseReading, siteContext *httputils.SiteContext) *FlexibleFloat64 {
	if gr.ReadingType == "sgv" {
		_, value := GetUnitsAndValue(gr, siteContext)
		return value
	}
	return nil

}

// goverter:context siteContext
func MBGGlucoseReadingTypeToValue(gr site.GlucoseReading, siteContext *httputils.SiteContext) *FlexibleFloat64 {
	if gr.ReadingType == "mbg" {
		_, value := GetUnitsAndValue(gr, siteContext)
		return value
	}
	return nil

}

// goverter:context siteContext
func UnitsToValue(gr site.GlucoseReading, siteContext *httputils.SiteContext) *consts.Units {
	// TODO stuff
	if gr.Units == nil {
		units, _ := GetUnitsAndValue(gr, siteContext)
		return units
	}

	return gr.Units
}

func GetUnitsAndValue(gr site.GlucoseReading, siteContext *httputils.SiteContext) (*consts.Units, *FlexibleFloat64) {
	mgdl := gr.Mgdl
	value := FlexibleFloat64(mgdl)
	return ptr(consts.UnitsMGDL), &value
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
