package controlcollections

import (
	"context"

	"github.com/a-h/templ"
	"nighttroop.t1d.nz/data/sql/jobsdb"
	"nighttroop.t1d.nz/models/controlmodels"
	"nighttroop.t1d.nz/server/components/control_templ/control_job_stats_templ"
	"nighttroop.t1d.nz/utils/apiutils"
	"nighttroop.t1d.nz/utils/httputils"
)

// JobStatsCollection lets control identities see each day's totals for every queue's jobs
type JobStatsCollection struct{}

func (s *JobStatsCollection) GetName() string {
	return "job stats"
}

// GetRoute is where the collection is in its URL, /job-stats
func (s *JobStatsCollection) GetRoute() string {
	return "job-stats"
}

func (s *JobStatsCollection) ToModel(row *jobsdb.JobStatsDaily) *controlmodels.JobStatsDaily {
	return controlmodels.JobStatsDailyFromDB(row)
}

func (s *JobStatsCollection) SearchRows(ctx context.Context, query *apiutils.Query, controlContext *httputils.ControlContext) ([]*jobsdb.JobStatsDaily, error) {
	return controlContext.ControlDB.Jobs.Queries.SearchJobStatsDaily(ctx, controlmodels.QueryToSearchJobStatsDaily(query))
}

func (s *JobStatsCollection) SearchHTMXResponse(models []*controlmodels.JobStatsDaily, query *apiutils.Query) templ.Component {
	return control_job_stats_templ.JobStatsTable(models, query.Skip, query.Limit)
}

func (s *JobStatsCollection) GetRow(ctx context.Context, id string, controlContext *httputils.ControlContext) (*jobsdb.JobStatsDaily, error) {
	return controlContext.ControlDB.Jobs.Queries.GetJobStatsDaily(ctx, jobsdb.GetJobStatsDailyParams{DayQueueJob: id})
}

// GetHTMXResponse is the day's totals and errors, shown in a modal
func (s *JobStatsCollection) GetHTMXResponse(model *controlmodels.JobStatsDaily) templ.Component {
	return control_job_stats_templ.JobStatsModal(model)
}
