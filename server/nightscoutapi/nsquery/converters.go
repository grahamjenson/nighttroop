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
	// goverter:map DateIn | QueryStringArrayToString
	// goverter:map TypeIn | QueryStringArrayToString
	// goverter:map TypeNin | QueryStringArrayToString
	// goverter:map IDIn | QueryStringArrayToString
	// goverter:map EnteredByNin | QueryStringArrayToString
	// goverter:map Fields | QueryStringArrayToArray
	// goverter:map . HasFilters | GetHasFilters
	// goverter:map . UnknownFilter | GetUnknownFilter
	// goverter:map . Skip | GetSkipInt64
	// goverter:map . Limit | GetLimitInt64
	// goverter:map . CountGiven | GetCountGiven
	// goverter:map . CountAsked | GetCountAskedInt64
	// goverter:ignore V3
	QueryValueToQuery(source QueryValues) Query

	//////
	// GLUCOSE READINGS
	//////

	// goverter:ignore Before BeforeID
	GlucoseAll(source Query) sitedb.AllGlucoseReadingsParams
	GlucoseSearch(source Query) sitedb.SearchGlucoseReadingsParams
	GlucoseAscending(source Query) sitedb.GlucoseReadingAscendingIDsParams

	//////
	// Activities
	//////

	// goverter:ignore Before BeforeID
	ActivityAll(source Query) sitedb.AllActivitiesParams
	ActivitySearch(source Query) sitedb.SearchActivitiesParams
	ActivityAscending(source Query) sitedb.ActivityAscendingIDsParams

	//////
	// DEVICE REPORTS
	//////

	DeviceReportAll(source Query) sitedb.AllDeviceReportsParams
	DeviceReportSearch(source Query) sitedb.SearchDeviceReportsParams
	DeviceReportAscending(source Query) sitedb.DeviceReportAscendingIDsParams

	//////
	// FoodPresets
	//////

	FoodPresetAll(source Query) sitedb.AllFoodPresetsParams
	FoodPresetSearch(source Query) sitedb.SearchFoodPresetsParams
	FoodPresetAscending(source Query) sitedb.FoodPresetAscendingIDsParams
	//////
	// FoodIntake
	//////

	// goverter:ignore Before BeforeID
	FoodIntakeAll(source Query) sitedb.AllFoodIntakesParams
	FoodIntakeSearch(source Query) sitedb.SearchFoodIntakesParams

	///////
	// Event Group
	///////

	// goverter:ignore Before BeforeID
	EventGroupAll(source Query) sitedb.AllEventGroupsParams
	EventGroupSearch(source Query) sitedb.SearchEventGroupsParams
	EventGroupAscending(source Query) sitedb.EventGroupAscendingIDsParams

	///////
	// Profiles Presets
	////////

	ProfilePresetsAll(source Query) sitedb.AllProfilePresetsParams
	ProfileStoreSearch(source Query) sitedb.SearchProfileStoresParams
	ProfileStoreAscending(source Query) sitedb.ProfileStoreAscendingIDsParams
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

func GetUnknownFilter(source QueryValues) bool {
	return source.HasUnknownFilter()
}

// GetSkipInt64 defaults a missing or negative skip to 0, and caps it at apiquery.MaxQuerySkip, as the site API's is, which
// Nightguard's report and statistics, paging by skip up to 32 pages of a search's most, fit in
func GetSkipInt64(source QueryValues) int64 {
	if source.Skip == nil || *source.Skip < 0 {
		return 0
	}
	return min(source.Skip.Int64(), apiquery.MaxQuerySkip)
}

// GetLimitInt64 defaults a missing or bad limit or count to the most a search returns, and caps it there, as the site's
// own API does
func GetLimitInt64(source QueryValues) int64 {
	if source.Limit == nil || *source.Limit <= 0 || *source.Limit > apiquery.MaxQueryLimit {
		return apiquery.MaxQueryLimit
	}
	return source.Limit.Int64()
}

// GetCountAskedInt64 is the limit or count sent, before it is capped, or 0 when none or a bad one was
func GetCountAskedInt64(source QueryValues) int64 {
	if source.Limit == nil || *source.Limit <= 0 {
		return 0
	}
	return source.Limit.Int64()
}

func GetCountGiven(source QueryValues) bool {
	return source.Limit != nil
}
