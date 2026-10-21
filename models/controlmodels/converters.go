package controlmodels

import (
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/data/sql/jobsdb"
	"nighttroop.t1d.nz/utils/apiutils/apiquery"
)

//go:generate goverter gen .

// goverter:converter
// goverter:output:format function
// goverter:output:file ./z_control_converters.go
type ControlConverter interface {
	SiteFromDB(source *controldb.Site) *Site

	// goverter:useZeroValueOnPointerInconsistency
	QueryToSearchSites(source *apiquery.Query) controldb.SearchSitesParams

	JobStatsDailyFromDB(source *jobsdb.JobStatsDaily) *JobStatsDaily

	// goverter:useZeroValueOnPointerInconsistency
	QueryToSearchJobStatsDaily(source *apiquery.Query) jobsdb.SearchJobStatsDailyParams

	JobFromDB(source *jobsdb.Job) *Job

	// goverter:useZeroValueOnPointerInconsistency
	QueryToSearchJobs(source *apiquery.Query) jobsdb.SearchJobsParams
}
