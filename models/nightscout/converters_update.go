package nightscout

import (
	"nighttroop.t1d.nz/data/sql/site"
)

// goverter:converter
// goverter:output:format function
// goverter:output:file ./z_update_converters.go
// goverter:useZeroValueOnPointerInconsistency
// goverter:extend FlexibleFloat64ToFloat64
// goverter:extend FlexibleInt64ToInt64
// goverter:extend ShortStringToString
// goverter:extend SanitizedStringToString
type UpdateMapConverter interface {

	///////
	// ACTIVITIES
	///////

	// goverter:context ctx
	// goverter:map Identifier ID
	// goverter:map UpdatedByIdentityName | AssignUpdatedByName
	//
	// goverter:ignore Reason
	ActivityToUpdate(source Activity, ctx ConverterContext) site.UpdateActivityParams

	///////
	// ENTRIES TO GLUCOSE READINGS
	///////

	// goverter:context ctx
	// goverter:map Identifier ID
	// goverter:map UpdatedByIdentityName | AssignUpdatedByName
	//
	// goverter:ignore Notes Reason
	EntryToUpdate(source Entry, ctx ConverterContext) site.UpdateGlucoseReadingParams

	///////
	// FOOD TO FOOD PRESET
	///////

	// goverter:context ctx
	// goverter:map Identifier ID
	// goverter:map UpdatedByIdentityName | AssignUpdatedByName
	//
	// goverter:ignore
	FoodToUpdate(source Food, ctx ConverterContext) site.UpdateFoodPresetParams

	//
	//
	//

	////////////////////////////////////////
	// DEVICE STATUS TO CREATE/UPDATE
	////////////////////////////////////////

	// goverter:context ctx
	// goverter:map Identifier ID
	// goverter:map UpdatedByIdentityName | AssignUpdatedByName
	//
	// goverter:ignore
	DeviceStatusToUpdateDeviceReport(source DeviceStatus, ctx ConverterContext) site.UpdateDeviceReportParams
}

// //
// Custom Converter Methods
// //

// goverter:context ctx
func AssignUpdatedByName(ctx ConverterContext) *string {
	return ptr(ctx.IdentityName)
}
