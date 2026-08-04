package nightscout

import (
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
	// goverter:map CreatedAt SrvCreated
	// goverter:map UpdatedAt SrvModified
	//
	// goverter:map DurationMinutes Duration
	DBModelToActivity(source *site.Activity) Activity

	///////
	// ENTRIES TO GLUCOSE READINGS
	///////

	// goverter:ignore Direction Trend Units Filtered Unfiltered RSSI Noise SysTime
	// goverter:map CreatedByIdentityName Subject
	// goverter:map CreatedByAppName App
	// goverter:map CreatedByDeviceName Device
	// goverter:map ID Identifier
	// goverter:map . SGV | SGVGlucoseReadingTypeToValue
	// goverter:map . MBG | MBGGlucoseReadingTypeToValue
	// goverter:map OccurredAt Date
	// goverter:map OccurredAt Mills
	// goverter:map CreatedAt SrvCreated
	// goverter:map UpdatedAt SrvModified
	// goverter:map . DateString | ModelDateStringFromDateIntAndUTCOffset
	// goverter:map ReadingType Type
	DBModelToEntry(source *site.GlucoseReading) Entry

	///////
	// FOOD TO FOOD PRESET
	///////

	// goverter:map CreatedByIdentityName Subject
	// goverter:map CreatedByAppName App
	// goverter:map CreatedByDeviceName Device
	// goverter:map ID Identifier
	// goverter:map CreatedAt SrvCreated
	// goverter:map UpdatedAt SrvModified
	//
	// goverter:ignore Food Category Subcategory Portion Unit GI Portions HideAfterUse Hidden Position Date
	// goverter:map Calories Energy
	DBModelToFood(source *site.FoodPreset) Food

	///////
	// Profile To Profile Preset
	///////

	// // goverter:map . CreatedAt | ModelDateStringFromDateIntAndUTCOffset
	// // goverter:map CreatedByIdentityName Subject
	// // goverter:map CreatedByAppName App
	// // goverter:map CreatedByDeviceName Device
	// // goverter:map ID Identifier
	// // goverter:map OccurredAt Date
	// // goverter:map CreatedAt SrvCreated
	// // goverter:map UpdatedAt SrvModified
	// //
	// // goverter:ignore DefaultProfile Store StartDate Mills
	// DBModelToProfile(source *site.ProfilePreset) Profile

	//
	//
	//

	////////////////////////////////////////
	// TREATMENT TO CREATE/UPDATE
	////////////////////////////////////////

	///////
	// EventGroup
	///////

	// goverter:map . CreatedAt | ModelDateStringFromDateIntAndUTCOffset
	// goverter:map EventGroup.UtcOffsetMinutes UtcOffsetMinutes
	// goverter:map EventGroup.RawJsonData RawJsonData
	// goverter:map EventGroup.CreatedByIdentityName Subject
	// goverter:map EventGroup.CreatedByAppName App
	// goverter:map EventGroup.CreatedByDeviceName Device
	// goverter:map EventGroup.ID Identifier
	// goverter:map EventGroup.OccurredAt Date
	// goverter:map EventGroup.CreatedAt SrvCreated
	// goverter:map EventGroup.UpdatedAt SrvModified
	//
	// goverter:ignore Name ActivityType Steps Calories HeartRate Carbs Insulin Protein Fat Fibre Energy PreBolus SplitNow SplitExt Duration Notes EnteredBy Profile Percentage Rate Absolute Reason TargetTop TargetBottom Units
	// goverter:map EventGroup.GroupType EventType
	// goverter:map GlucoseReading.Reading Glucose
	// goverter:map GlucoseReading.ReadingType GlucoseType
	DBModelToTreatment(source *site.GetEventGroupRow) Treatment

	////////////////////////////////////////
	// DEVICE STATUS TO CREATE/UPDATE
	////////////////////////////////////////

	// goverter:map . CreatedAt | ModelDateStringFromDateIntAndUTCOffset
	// goverter:map EventGroup.UtcOffsetMinutes UtcOffsetMinutes
	// goverter:map EventGroup.RawJsonData RawJsonData
	// goverter:map EventGroup.CreatedByIdentityName Subject
	// goverter:map EventGroup.CreatedByAppName App
	// goverter:map EventGroup.CreatedByDeviceName Device
	// goverter:map EventGroup.ID Identifier
	// goverter:map EventGroup.OccurredAt Date
	// goverter:map EventGroup.CreatedAt SrvCreated
	// goverter:map EventGroup.UpdatedAt SrvModified
	//
	// goverter:ignore Uploader Loop OpenAPS Pump UploaderBattery MergedLoopStatus
	DBModelToDeviceStatus(source *site.GetEventGroupRow) DeviceStatus
}

// //
// Custom Converter Methods
// //

func ModelDateStringFromDateIntAndUTCOffset(m interface{}) *DateString {
	if m == nil {
		return nil
	}

	// TODO this should be mapping to interfaces now we have getters
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
	default:
		return nil
	}

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
