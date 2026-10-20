package sitequery

import (
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/data/sql/sitedb"
)

//go:generate goverter gen .

// goverter:converter
// goverter:output:format function
// goverter:output:file ./z_converter.go
type QueryConverter interface {
	// goverter:map . Limit | GetLimitInt64
	// goverter:map . Skip | GetSkipInt64
	QueryValuesToQuery(source QueryValues) Query

	// goverter:useZeroValueOnPointerInconsistency
	// goverter:ignore IncludeDeleted
	QueryToAllGlucoseReadings(source *Query) sitedb.AllGlucoseReadingsParams

	// goverter:useZeroValueOnPointerInconsistency
	QueryToSearchSites(source *Query) controldb.SearchSitesParams

	// goverter:useZeroValueOnPointerInconsistency
	// goverter:ignore IncludeDeleted
	QueryToAllTempTargets(source *Query) sitedb.AllTempTargetsParams

	// goverter:useZeroValueOnPointerInconsistency
	// goverter:ignore IncludeDeleted
	QueryToAllTempBasals(source *Query) sitedb.AllTempBasalsParams

	// goverter:useZeroValueOnPointerInconsistency
	// goverter:ignore IncludeDeleted
	QueryToAllInsulinDoses(source *Query) sitedb.AllInsulinDosesParams

	// goverter:useZeroValueOnPointerInconsistency
	// goverter:ignore IncludeDeleted
	QueryToAllFoodIntakes(source *Query) sitedb.AllFoodIntakesParams
}

// GetLimitInt64 defaults a missing or bad limit to MaxQueryLimit, and caps it there
func GetLimitInt64(source QueryValues) int64 {
	if source.Limit == nil || *source.Limit <= 0 || *source.Limit > MaxQueryLimit {
		return MaxQueryLimit
	}
	return source.Limit.Int64()
}

func GetSkipInt64(source QueryValues) int64 {
	if source.Skip == nil || *source.Skip < 0 {
		return 0
	}
	return source.Skip.Int64()
}
