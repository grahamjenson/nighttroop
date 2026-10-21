package sitedata

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"nighttroop.t1d.nz/utils/apiutils"

	"github.com/a-h/templ"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/models/sitemodels"
	"nighttroop.t1d.nz/server/components/site_templ/site_data_templ"
	"nighttroop.t1d.nz/server/siteapi/siteapiutils"
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
func (s *ActivitiesStore) CreateModel(ctx context.Context, tx *siteapiutils.StoreTx, id *string, siteContext *httputils.SiteContext, model *sitemodels.SiteActivity) (bool, *apiutils.HTTPError) {
	if model.ActivityType == "" {
		model.ActivityType = defaultActivityType
	}
	if herr := checkActivity(model); herr != nil {
		return false, herr
	}
	q, err := tx.Site()
	if err != nil {
		return false, apiutils.ErrInternalServerError
	}
	created, err := q.CreateActivity(ctx, sitemodels.ActivityToCreate(model, siteContext))
	if err != nil {
		return false, apiutils.WriteError(err, "activity")
	}
	*model = *sitemodels.ActivityFromDB(created)
	return true, nil
}

func (s *ActivitiesStore) CreateHTMXResponse(model *sitemodels.SiteActivity) templ.Component {
	return site_data_templ.ActivityCreated(model)
}

// ////
// UPDATE
// ////

// UpdateModel changes the fields sent; anything not sent stays as it is
func (s *ActivitiesStore) UpdateModel(ctx context.Context, tx *siteapiutils.StoreTx, id string, siteContext *httputils.SiteContext, oldModel *sitemodels.SiteActivity, model *sitemodels.SiteActivity) *apiutils.HTTPError {
	if herr := checkActivity(model); herr != nil {
		return herr
	}
	q, err := tx.Site()
	if err != nil {
		return apiutils.ErrInternalServerError
	}
	model.ID = oldModel.ID
	updated, err := q.UpdateActivity(ctx, sitemodels.ActivityToUpdate(model, siteContext))
	if err == sql.ErrNoRows {
		return apiutils.ErrNotFound
	}
	if err != nil {
		return apiutils.WriteError(err, "activity")
	}
	*model = *sitemodels.ActivityFromDB(updated)
	return nil
}

func (s *ActivitiesStore) UpdateHTMXResponse(model *sitemodels.SiteActivity) templ.Component {
	return site_data_templ.ActivityUpdated(model)
}

// ////
// Get Model
// /////
func (s *ActivitiesStore) GetModel(ctx context.Context, tx *siteapiutils.StoreTx, id string, siteContext *httputils.SiteContext) (*sitemodels.SiteActivity, *apiutils.HTTPError) {
	q, err := tx.Site()
	if err != nil {
		return nil, apiutils.ErrInternalServerError
	}
	activity, err := q.GetActivity(ctx, sitedb.GetActivityParams{ID: id})
	return apiutils.Found(activity, err, "activity", sitemodels.ActivityFromDB)
}

// GetHTMXResponse is the activity's details, shown in a modal
func (s *ActivitiesStore) GetHTMXResponse(model *sitemodels.SiteActivity) templ.Component {
	return site_data_templ.ActivityModal(model)
}

// ///
// Search Models
// ///

func (s *ActivitiesStore) SearchModels(ctx context.Context, tx *siteapiutils.StoreTx, query *apiutils.Query, siteContext *httputils.SiteContext) ([]*sitemodels.SiteActivity, *apiutils.HTTPError) {
	q, err := tx.Site()
	if err != nil {
		return nil, apiutils.ErrInternalServerError
	}
	activities, err := q.AllActivities(ctx, sitequery.QueryToAllActivities(query))
	return apiutils.Searched(activities, err, "activities", sitemodels.ActivityFromDB)
}

func (s *ActivitiesStore) SearchHTMXResponse(models []*sitemodels.SiteActivity, query *apiutils.Query) templ.Component {
	return site_data_templ.ActivitiesTable(models, query.Skip, query.Limit)
}

// ///
// Delete
// ///

// DeleteModel soft deletes, like the Nightscout API, so syncing clients see the delete
func (s *ActivitiesStore) DeleteModel(ctx context.Context, tx *siteapiutils.StoreTx, model *sitemodels.SiteActivity, siteContext *httputils.SiteContext) *apiutils.HTTPError {
	q, err := tx.Site()
	if err != nil {
		return apiutils.ErrInternalServerError
	}
	_, err = q.SoftDeleteActivity(ctx, sitedb.SoftDeleteActivityParams{
		UpdatedByIdentityName: &siteContext.IdentityModel.Name,
		ID:                    model.ID,
	})
	if err != nil {
		slog.Error("Error deleting activity", "error", err)
		return apiutils.ErrInternalServerError
	}
	return nil
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
