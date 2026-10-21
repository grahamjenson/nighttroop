package sitequery

import (
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/utils/apiutils/apiquery"
)

//go:generate goverter gen .

// goverter:converter
// goverter:output:format function
// goverter:output:file ./z_converter.go
type QueryConverter interface {
	// goverter:useZeroValueOnPointerInconsistency
	// goverter:ignore IncludeDeleted
	QueryToAllGlucoseReadings(source *apiquery.Query) sitedb.AllGlucoseReadingsParams

	// goverter:useZeroValueOnPointerInconsistency
	// goverter:ignore IncludeDeleted
	QueryToAllTempTargets(source *apiquery.Query) sitedb.AllTempTargetsParams

	// goverter:useZeroValueOnPointerInconsistency
	// goverter:ignore IncludeDeleted
	QueryToAllTempBasals(source *apiquery.Query) sitedb.AllTempBasalsParams

	// goverter:useZeroValueOnPointerInconsistency
	// goverter:ignore IncludeDeleted
	QueryToAllInsulinDoses(source *apiquery.Query) sitedb.AllInsulinDosesParams

	// goverter:useZeroValueOnPointerInconsistency
	// goverter:ignore IncludeDeleted
	QueryToAllFoodIntakes(source *apiquery.Query) sitedb.AllFoodIntakesParams

	// goverter:useZeroValueOnPointerInconsistency
	// goverter:ignore IncludeDeleted
	QueryToAllComponentEvents(source *apiquery.Query) sitedb.AllComponentEventsParams

	// goverter:useZeroValueOnPointerInconsistency
	QueryToDeviceReportsForDevice(source *apiquery.Query) sitedb.DeviceReportsForDeviceParams
}
