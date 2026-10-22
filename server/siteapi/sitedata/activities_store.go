package sitedata

import (
	"context"
	"fmt"

	"nighttroop.t1d.nz/utils/apiutils"

	"github.com/a-h/templ"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/models/sitemodels"
	"nighttroop.t1d.nz/server/components/site_templ/site_data_templ"
	"nighttroop.t1d.nz/server/siteapi/sitequery"
	"nighttroop.t1d.nz/utils/httputils"
)

// defaultActivityType is the type of an activity sent without one
const defaultActivityType = "exercise"

type ActivitiesStore struct{}

// ////
// Shared
// ////
func (s *ActivitiesStore) GetName() string {
	return "activities"
}

// ////
// CREATE
// ////
func (s *ActivitiesStore) CreateRow(ctx context.Context, q *sitedb.Queries, id *string, siteContext *httputils.SiteContext, model *sitemodels.SiteActivity) (*sitedb.Activity, error) {
	if model.ActivityType == "" {
		model.ActivityType = defaultActivityType
	}
	if herr := checkActivity(model); herr != nil {
		return nil, herr
	}
	return q.CreateActivity(ctx, sitemodels.ActivityToCreate(model, siteContext))
}

func (s *ActivitiesStore) CreateHTMXResponse(model *sitemodels.SiteActivity) templ.Component {
	return site_data_templ.ActivityCreated(model)
}

// ////
// UPDATE
// ////

// UpdateRow changes the fields sent; anything not sent stays as it is
func (s *ActivitiesStore) UpdateRow(ctx context.Context, q *sitedb.Queries, siteContext *httputils.SiteContext, oldModel *sitemodels.SiteActivity, model *sitemodels.SiteActivity) (*sitedb.Activity, error) {
	if herr := checkActivity(model); herr != nil {
		return nil, herr
	}
	model.ID = oldModel.ID
	return q.UpdateActivity(ctx, sitemodels.ActivityToUpdate(model, siteContext))
}

func (s *ActivitiesStore) UpdateHTMXResponse(model *sitemodels.SiteActivity) templ.Component {
	return site_data_templ.ActivityUpdated(model)
}

// ////
// Get Model
// /////
func (s *ActivitiesStore) GetRow(ctx context.Context, q *sitedb.Queries, id string, siteContext *httputils.SiteContext) (*sitedb.Activity, error) {
	return q.GetActivity(ctx, sitedb.GetActivityParams{ID: id})
}

// ToModel is the site model of a row
func (s *ActivitiesStore) ToModel(row *sitedb.Activity, siteContext *httputils.SiteContext) *sitemodels.SiteActivity {
	return sitemodels.ActivityFromDB(row)
}

// GetHTMXResponse is the activity's details, shown in a modal
func (s *ActivitiesStore) GetHTMXResponse(model *sitemodels.SiteActivity) templ.Component {
	return site_data_templ.ActivityModal(model)
}

// ///
// Search Models
// ///

func (s *ActivitiesStore) SearchRows(ctx context.Context, q *sitedb.Queries, query *apiutils.Query, siteContext *httputils.SiteContext) ([]*sitedb.Activity, error) {
	return q.AllActivities(ctx, sitequery.QueryToAllActivities(query))
}

func (s *ActivitiesStore) SearchHTMXResponse(models []*sitemodels.SiteActivity, query *apiutils.Query) templ.Component {
	return site_data_templ.ActivitiesTable(models, query.Skip, query.Limit)
}

// ///
// Delete
// ///

// DeleteRow soft deletes, like the Nightscout API, so syncing clients see the delete
func (s *ActivitiesStore) DeleteRow(ctx context.Context, q *sitedb.Queries, id string, siteContext *httputils.SiteContext) error {
	_, err := q.SoftDeleteActivity(ctx, sitedb.SoftDeleteActivityParams{
		UpdatedByIdentityName: &siteContext.IdentityModel.Name,
		ID:                    id,
	})
	return err
}

func (s *ActivitiesStore) DeleteHTMXResponse(model *sitemodels.SiteActivity) templ.Component {
	return site_data_templ.ActivityDeleted()
}

///////
// UTIL METHODS
////////

// activityLimits keep each number sent within what a day's activity could have
var activityLimits = []struct {
	name     string
	value    func(*sitemodels.SiteActivity) *int64
	min, max int64
}{
	{"duration", func(a *sitemodels.SiteActivity) *int64 { return a.DurationMinutes }, 1, 24 * 60},
	{"calories", func(a *sitemodels.SiteActivity) *int64 { return a.Calories }, 0, 20_000},
	{"steps", func(a *sitemodels.SiteActivity) *int64 { return a.Steps }, 0, 200_000},
	{"heart rate", func(a *sitemodels.SiteActivity) *int64 { return a.HeartRate }, 20, 250},
}

func checkActivity(model *sitemodels.SiteActivity) *apiutils.HTTPError {
	for _, limit := range activityLimits {
		if v := limit.value(model); v != nil && (*v < limit.min || *v > limit.max) {
			return apiutils.ErrConstraintCheckFailedWithHint(fmt.Sprintf("%s must be from %d to %d", limit.name, limit.min, limit.max))
		}
	}
	return nil
}
