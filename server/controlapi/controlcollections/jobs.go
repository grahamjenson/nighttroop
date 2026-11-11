package controlcollections

import (
	"context"
	"strconv"

	"github.com/a-h/templ"
	"nighttroop.t1d.nz/data/sql/jobsdb"
	"nighttroop.t1d.nz/models/controlmodels"
	"nighttroop.t1d.nz/server/components/control_templ/control_jobs_templ"
	"nighttroop.t1d.nz/utils/apiutils"
	"nighttroop.t1d.nz/utils/httputils"
)

// JobsCollection lets control identities see the background jobs not yet summarised into the job stats
type JobsCollection struct{}

func (s *JobsCollection) GetName() string {
	return "jobs"
}

// GetRoute is where the collection is in its URL, /jobs
func (s *JobsCollection) GetRoute() string {
	return "jobs"
}

func (s *JobsCollection) ToModel(row *jobsdb.Job) *controlmodels.Job {
	return controlmodels.JobFromDB(row)
}

func (s *JobsCollection) SearchRows(ctx context.Context, query *apiutils.Query, controlContext *httputils.ControlContext) ([]*jobsdb.Job, error) {
	return controlContext.ControlDB.Jobs.Queries.SearchJobs(ctx, controlmodels.QueryToSearchJobs(query))
}

func (s *JobsCollection) SearchHTMXResponse(models []*controlmodels.Job, query *apiutils.Query) templ.Component {
	return control_jobs_templ.JobsTable(models, query.Skip, query.Limit)
}

// GetRow finds a job by its number; anything else is not a job
func (s *JobsCollection) GetRow(ctx context.Context, id string, controlContext *httputils.ControlContext) (*jobsdb.Job, error) {
	jobID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return nil, nil
	}
	return controlContext.ControlDB.Jobs.Queries.FindJob(ctx, jobsdb.FindJobParams{ID: jobID})
}

// GetHTMXResponse is the job's details and errors, shown in a modal
func (s *JobsCollection) GetHTMXResponse(model *controlmodels.Job) templ.Component {
	return control_jobs_templ.JobModal(model)
}
