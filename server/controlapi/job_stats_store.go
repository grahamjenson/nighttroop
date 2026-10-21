package controlapi

import (
	"context"

	"github.com/a-h/templ"
	"nighttroop.t1d.nz/data/sql/jobsdb"
	"nighttroop.t1d.nz/models/controlmodels"
	"nighttroop.t1d.nz/server/components/control_templ/control_job_stats_templ"
	"nighttroop.t1d.nz/utils/apiutils"
	"nighttroop.t1d.nz/utils/httputils"
)

// JobStatsStore lets control identities see each day's totals for every queue's jobs
type JobStatsStore struct{}

func (s *JobStatsStore) SearchModels(ctx context.Context, query *apiutils.Query, controlContext *httputils.ControlContext) ([]*controlmodels.JobStatsDaily, *apiutils.HTTPError) {
	stats, err := controlContext.ControlDB.Jobs.Queries.SearchJobStatsDaily(ctx, controlmodels.QueryToSearchJobStatsDaily(query))
	return apiutils.Searched(stats, err, "job stats", controlmodels.JobStatsDailyFromDB)
}

func (s *JobStatsStore) SearchHTMXResponse(models []*controlmodels.JobStatsDaily, query *apiutils.Query) templ.Component {
	return control_job_stats_templ.JobStatsTable(models, query.Skip, query.Limit)
}

func (s *JobStatsStore) GetModel(ctx context.Context, id string, controlContext *httputils.ControlContext) (*controlmodels.JobStatsDaily, *apiutils.HTTPError) {
	stats, err := controlContext.ControlDB.Jobs.Queries.GetJobStatsDaily(ctx, jobsdb.GetJobStatsDailyParams{DayQueueJob: id})
	return apiutils.Found(stats, err, "job stats", controlmodels.JobStatsDailyFromDB)
}

// GetHTMXResponse is the day's totals and errors, shown in a modal
func (s *JobStatsStore) GetHTMXResponse(model *controlmodels.JobStatsDaily) templ.Component {
	return control_job_stats_templ.JobStatsModal(model)
}
