package nsquery

import (
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/utils/apiutils/apiquery"
)

//go:generate go tool goverter gen .

// goverter:converter
// goverter:output:format function
// goverter:output:file ./z_converter.go
type ToSearchParams interface {

	// goverter:useZeroValueOnPointerInconsistency
	// goverter:map TypeIn | QueryStringArrayToString
	// goverter:map TypeNin | QueryStringArrayToString
	// goverter:map IDIn | QueryStringArrayToString
	// goverter:map Fields | QueryStringArrayToArray
	// goverter:map . HasFilters | GetHasFilters
	// goverter:map . Skip | GetSkipInt64
	// goverter:map . Limit | GetLimitInt64
	QueryValueToQuery(source QueryValues) Query

	//////
	// GLUCOSE READINGS
	//////

	GlucoseAll(source Query) sitedb.AllGlucoseReadingsParams
	GlucoseSearch(source Query) sitedb.SearchGlucoseReadingsParams

	//////
	// Activities
	//////

	ActivityAll(source Query) sitedb.AllActivitiesParams
	ActivitySearch(source Query) sitedb.SearchActivitiesParams

	//////
	// DEVICE REPORTS
	//////

	DeviceReportAll(source Query) sitedb.AllDeviceReportsParams
	DeviceReportSearch(source Query) sitedb.SearchDeviceReportsParams

	//////
	// FoodPresets
	//////

	FoodPresetAll(source Query) sitedb.AllFoodPresetsParams
	FoodPresetSearch(source Query) sitedb.SearchFoodPresetsParams
	//////
	// FoodIntake
	//////

	FoodIntakeAll(source Query) sitedb.AllFoodIntakesParams
	FoodIntakeSearch(source Query) sitedb.SearchFoodIntakesParams

	///////
	// Event Group
	///////

	EventGroupAll(source Query) sitedb.AllEventGroupsParams
	EventGroupSearch(source Query) sitedb.SearchEventGroupsParams

	///////
	// Profiles Presets
	////////

	ProfilePresetsAll(source Query) sitedb.AllProfilePresetsParams
	ProfileStoreSearch(source Query) sitedb.SearchProfileStoresParams
}

func QueryStringArrayToString(source *QueryStringArray) *string {
	return source.Stringp()
}

func QueryStringArrayToArray(source *QueryStringArray) []string {
	return source.ToArray()
}

func GetHasFilters(source QueryValues) bool {
	return source.HasAnyFilters()
}

func GetSkipInt64(source QueryValues) int64 {
	if source.Skip == nil {
		return 0
	}
	return source.Skip.Int64()
}

// GetLimitInt64 defaults a missing or bad limit or count to the most a search returns, and caps it there, as the site's
// own API does
func GetLimitInt64(source QueryValues) int64 {
	if source.Limit == nil || *source.Limit <= 0 || *source.Limit > apiquery.MaxQueryLimit {
		return apiquery.MaxQueryLimit
	}
	return source.Limit.Int64()
}

func GetSortOrderInt64(source Query) int64 {
	return source.SortOrder.ToIntVal()
}
