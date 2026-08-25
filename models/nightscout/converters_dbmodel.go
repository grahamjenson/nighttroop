package nightscout

import (
	"encoding/json"

	"nighttroop.t1d.nz/data/sql/site"
)

// goverter:converter
// goverter:output:format function
// goverter:output:file ./z_dbmodel_converters.go
// goverter:useZeroValueOnPointerInconsistency
type DBModelConverter interface {

	///////
	// ACTIVITIES
	///////

	// goverter:map . CreatedAt | ModelDateStringFromDateIntAndUTCOffset
	// goverter:map CreatedByIdentityName Subject
	// goverter:map CreatedByAppName App
	// goverter:map CreatedByDeviceName Device
	// goverter:map ID Identifier
	// goverter:map OccurredAt Date
	//
	DBModelToActivity(source *site.Activity) Activity

	///////
	// ENTRIES TO GLUCOSE READINGS
	///////

	// goverter:ignore SysTime Units
	// goverter:map CreatedByIdentityName Subject
	// goverter:map CreatedByAppName App
	// goverter:map CreatedByDeviceName Device
	// goverter:map ID Identifier
	// goverter:map . SGV | SGVGlucoseReadingTypeToValue
	// goverter:map . MBG | MBGGlucoseReadingTypeToValue
	// goverter:map OccurredAt Date
	// goverter:map OccurredAt Mills
	// goverter:map . DateString | ModelDateStringFromDateIntAndUTCOffset
	// goverter:map ReadingType Type
	// goverter:map Delta Trend
	// goverter:map Rssi RSSI
	DBModelToEntry(source *site.GlucoseReading) Entry

	///////
	// DEVICE REPORT TO DEVICE STATUS
	///////

	// goverter:ignore Uploader Loop Pump OpenAPS
	// goverter:map CreatedByIdentityName Subject
	// goverter:map CreatedByAppName App
	// goverter:map CreatedByDeviceName Device
	// goverter:map ID Identifier
	// goverter:map OccurredAt Date
	// goverter:map . CreatedAt | ModelDateStringFromDateIntAndUTCOffset
	//
	// goverter:ignore IsCharging Uploader UploaderBattery Pump OpenAPS Loop OpenAPSRaw LoopRaw ConfigurationRaw Configuration
	DBModelToDeviceStatus(source *site.DeviceReport) DeviceStatus

	///////
	// FOOD TO FOOD PRESET
	///////

	// goverter:map CreatedByIdentityName Subject
	// goverter:map CreatedByAppName App
	// goverter:map CreatedByDeviceName Device
	// goverter:map ID Identifier
	//
	// goverter:ignore Unit GI HideAfterUse Hidden Position
	// goverter:map CreatedAt Date
	DBModelToFood(source *site.FoodPreset) Food

	///////
	// Profile To Profile Preset
	///////

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
	DBModelToProfileStore(source *site.ProfilePreset) ProfileStore

	// goverter:map CarbRatiosJson           CarbRatio  | JSONToScheduleValueSlice
	// goverter:map InsulinSensitivitiesJson Sens       | JSONToScheduleValueSlice
	// goverter:map BasalRatesJson           Basal      | JSONToScheduleValueSlice
	// goverter:map TargetsHighJson          TargetHigh | JSONToScheduleValueSlice
	// goverter:map TargetsLowJson           TargetLow  | JSONToScheduleValueSlice
	DBModelToProfile(source *site.ProfilePreset) Profile
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
