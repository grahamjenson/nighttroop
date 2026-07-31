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
	// goverter:map Reading Value
	DBModelToEntry(source *site.GlucoseReading) Entry

	///////
	// FOOD TO FOOD PRESET
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
	DBModelToFood(source *site.FoodPreset) Food

	///////
	// Profile To Profile Preset
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
	DBModelToProfile(source *site.ProfilePreset) Profile

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
	// goverter:map CreatedByIdentityName Subject
	// goverter:map CreatedByAppName App
	// goverter:map CreatedByDeviceName Device
	// goverter:map ID Identifier
	// goverter:map OccurredAt Date
	// goverter:map CreatedAt SrvCreated
	// goverter:map UpdatedAt SrvModified
	//
	// goverter:map DurationMinutes Duration
	DBModelToTreatment(source *site.SearchEventGroupsRow) Treatment

	////////////////////////////////////////
	// DEVICE STATUS TO CREATE/UPDATE
	////////////////////////////////////////

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
	DBModelToDeviceStatus(source *site.SearchEventGroupsRow) DeviceStatus

	////////
	// SETTINGS
	////////

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
	DBModelToSettings(source *site.SearchEventGroupsRow) []Setting
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
