package nsquery

import "nighttroop.t1d.nz/data/sql/user"

//go:generate goverter gen .

// goverter:converter
// goverter:output:format function
// goverter:output:file ./z_converter.go
type ToSearchParams interface {

	// goverter:useZeroValueOnPointerInconsistency
	// goverter:ignore TypeNin IDIn Fields
	// goverter:map TypeIn | QueryStringArrayToString
	// goverter:map TypeNin | QueryStringArrayToString
	// goverter:map IDIn | QueryStringArrayToString
	// goverter:map Fields | QueryStringArrayToArray
	// goverter:map . HasFilters | GetHasFilters
	QueryValueToQuery(source *QueryValues) *Query

	//////
	// GLUCOSE READINGS
	//////

	GlucoseSearchAll(source *Query) *user.AllGlucoseReadingsParams
	GlucoseSearchDefaultOrder(source *Query) *user.SearchWithDefaultSortGlucoseReadingsParams

	// goverter:map . SortOrder | GetSortOrderInt64
	GlucoseSearchNonDefaultOrder(source *Query) *user.SearchWithNonDefaultSortGlucoseReadingsParams
}

func QueryStringArrayToString(source *QueryStringArray) *string {
	return source.Stringp()
}

func QueryStringArrayToArray(source *QueryStringArray) []string {
	return source.ToArray()
}

func GetHasFilters(source *QueryValues) bool {
	return source.HasAnyFilters()
}

func GetSortOrderInt64(source *Query) int64 {
	return source.SortOrder.ToIntVal()
}
