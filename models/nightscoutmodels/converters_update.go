package nightscoutmodels

import (
	"nighttroop.t1d.nz/data/sql/sitedb"
)

// goverter:converter
// goverter:output:format function
// goverter:output:file ./z_update_converters.go
// goverter:useZeroValueOnPointerInconsistency
// goverter:extend FlexibleFloat64ToFloat64
// goverter:extend FlexibleInt64ToInt64
// goverter:extend ShortStringToString
// goverter:extend SanitizedStringToString
// goverter:extend UnitsToUnits
type UpdateMapConverter interface {

	///////
	// ACTIVITIES
	///////

	// goverter:context ctx
	// goverter:map Identifier ID
	// goverter:map UpdatedByIdentityName | AssignUpdatedByName
	//
	// goverter:ignore Reason
	ActivityToUpdate(source Activity, ctx ConverterContext) sitedb.UpdateActivityParams

	///////
	// ENTRIES TO GLUCOSE READINGS
	///////

	// goverter:context ctx
	// goverter:map Identifier ID
	// goverter:map UpdatedByIdentityName | AssignUpdatedByName
	//
	// goverter:ignore Notes Reason
	EntryToUpdate(source Entry, ctx ConverterContext) sitedb.UpdateGlucoseReadingParams

	///////
	// FOOD TO FOOD PRESET
	///////

	// goverter:context ctx
	// goverter:map Identifier ID
	// goverter:map UpdatedByIdentityName | AssignUpdatedByName
	//
	// goverter:ignore
	FoodToUpdate(source Food, ctx ConverterContext) sitedb.UpdateFoodPresetParams

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
	DeviceStatusToUpdateDeviceReport(source DeviceStatus, ctx ConverterContext) sitedb.UpdateDeviceReportParams
}

// //
// Custom Converter Methods
// //

// goverter:context ctx
func AssignUpdatedByName(ctx ConverterContext) *string {
	return ptr(ctx.IdentityName)
}
