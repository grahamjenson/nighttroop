package nsquery

import "nighttroop.t1d.nz/data/sql/site"

//go:generate goverter gen .

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

	GlucoseAll(source Query) site.AllGlucoseReadingsParams
	GlucoseSearch(source Query) site.SearchGlucoseReadingsParams

	//////
	// DEVICE REPORTS
	//////

	DeviceReportAll(source Query) site.AllDeviceReportsParams
	DeviceReportSearch(source Query) site.SearchDeviceReportsParams

	//////
	// FoodIntake
	//////

	FoodIntakeAll(source Query) site.AllFoodIntakesParams
	FoodIntakeSearch(source Query) site.SearchFoodIntakesParams

	///////
	// Event Group
	///////

	EventGroupAll(source Query) site.AllEventGroupsParams
	EventGroupSearch(source Query) site.SearchEventGroupsParams
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

func GetLimitInt64(source QueryValues) int64 {
	if source.Limit == nil || *source.Limit <= 0 || *source.Limit > 1000 {
		return 1000
	}
	return source.Limit.Int64()
}

func GetSortOrderInt64(source Query) int64 {
	return source.SortOrder.ToIntVal()
}
