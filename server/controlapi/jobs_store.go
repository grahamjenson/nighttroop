package controlapi

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

// JobsStore lets control identities see the background jobs not yet summarised into the job stats
type JobsStore struct{}

func (s *JobsStore) SearchModels(ctx context.Context, query *apiutils.Query, controlContext *httputils.ControlContext) ([]*controlmodels.Job, *apiutils.HTTPError) {
	jobs, err := controlContext.ControlDB.Jobs.Queries.SearchJobs(ctx, controlmodels.QueryToSearchJobs(query))
	return apiutils.Searched(jobs, err, "jobs", controlmodels.JobFromDB)
}

func (s *JobsStore) SearchHTMXResponse(models []*controlmodels.Job, query *apiutils.Query) templ.Component {
	return control_jobs_templ.JobsTable(models, query.Skip, query.Limit)
}

// GetModel finds a job by its number; anything else is not a job
func (s *JobsStore) GetModel(ctx context.Context, id string, controlContext *httputils.ControlContext) (*controlmodels.Job, *apiutils.HTTPError) {
	jobID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return nil, nil
	}
	job, err := controlContext.ControlDB.Jobs.Queries.FindJob(ctx, jobsdb.FindJobParams{ID: jobID})
	return apiutils.Found(job, err, "job", controlmodels.JobFromDB)
}

// GetHTMXResponse is the job's details and errors, shown in a modal
func (s *JobsStore) GetHTMXResponse(model *controlmodels.Job) templ.Component {
	return control_jobs_templ.JobModal(model)
}
