package sitedata

import (
	"context"

	"nighttroop.t1d.nz/utils/apiutils"

	"github.com/a-h/templ"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/models/sitemodels"
	"nighttroop.t1d.nz/server/components/site_templ/site_data_templ"
	"nighttroop.t1d.nz/server/siteapi/siteapiutils"
	"nighttroop.t1d.nz/utils/httputils"
)

// defaultActivityType is the type of an activity sent without one
const defaultActivityType = "exercise"

type ActivitiesCollection struct{}

// ////
// Shared
// ////
func (s *ActivitiesCollection) GetName() string {
	return "activities"
}

// GetRoute is where the collection is in its URL
func (s *ActivitiesCollection) GetRoute() string {
	return "activities"
}

// ////
// CREATE
// ////
func (s *ActivitiesCollection) CreateRow(ctx context.Context, collectionTx *siteapiutils.CollectionTx, id *string, siteContext *httputils.SiteContext, model *sitemodels.SiteActivity) (*sitedb.Activity, error) {
	if model.ActivityType == "" {
		model.ActivityType = defaultActivityType
	}
	q, err := collectionTx.Site()
	if err != nil {
		return nil, err
	}

	return q.CreateActivity(ctx, sitemodels.ActivityToCreate(model, siteContext))
}

// NewHTMXResponse is the modal to add an activity in
func (s *ActivitiesCollection) NewHTMXResponse() templ.Component {
	return site_data_templ.ActivityNew()
}

func (s *ActivitiesCollection) CreateHTMXResponse(model *sitemodels.SiteActivity) templ.Component {
	return site_data_templ.ActivityCreated(model)
}

// ////
// UPDATE
// ////

// UpdateRow changes the fields sent; anything not sent stays as it is
func (s *ActivitiesCollection) UpdateRow(ctx context.Context, collectionTx *siteapiutils.CollectionTx, siteContext *httputils.SiteContext, oldModel *sitemodels.SiteActivity, model *sitemodels.SiteActivity) (*sitedb.Activity, error) {
	model.ID = oldModel.ID
	return siteapiutils.SiteQuery(ctx, collectionTx, (*sitedb.Queries).UpdateActivity, sitemodels.ActivityToUpdate(model, siteContext))
}

func (s *ActivitiesCollection) UpdateHTMXResponse(model *sitemodels.SiteActivity) templ.Component {
	return site_data_templ.ActivityUpdated(model)
}

// ////
// Get Model
// /////
func (s *ActivitiesCollection) GetRow(ctx context.Context, collectionTx *siteapiutils.CollectionTx, id string, siteContext *httputils.SiteContext) (*sitedb.Activity, error) {
	return siteapiutils.SiteQuery(ctx, collectionTx, (*sitedb.Queries).GetActivity, sitedb.GetActivityParams{ID: id})
}

// ToModel is the site model of a row
func (s *ActivitiesCollection) ToModel(row *sitedb.Activity, siteContext *httputils.SiteContext) *sitemodels.SiteActivity {
	return sitemodels.ActivityFromDB(row)
}

// GetHTMXResponse is the activity's details, shown in a modal
func (s *ActivitiesCollection) GetHTMXResponse(model *sitemodels.SiteActivity) templ.Component {
	return site_data_templ.ActivityModal(model)
}

// ///
// Search Models
// ///

func (s *ActivitiesCollection) SearchRows(ctx context.Context, collectionTx *siteapiutils.CollectionTx, query *apiutils.Query, siteContext *httputils.SiteContext) ([]*sitedb.Activity, error) {
	return siteapiutils.SiteQuery(ctx, collectionTx, (*sitedb.Queries).AllActivities, siteapiutils.QueryToAllActivities(query))
}

// SearchTotals are how many the search pages through, and when the earliest and latest occurred
func (s *ActivitiesCollection) SearchTotals(ctx context.Context, collectionTx *siteapiutils.CollectionTx, query *apiutils.Query, siteContext *httputils.SiteContext) (*apiutils.Totals, error) {
	return siteapiutils.SearchTotals(ctx, collectionTx, (*sitedb.Queries).ActivityTotals)
}

func (s *ActivitiesCollection) SearchHTMXResponse(models []*sitemodels.SiteActivity, query *apiutils.Query) templ.Component {
	return site_data_templ.ActivitiesTable(models, query.Skip, query.Limit, query.Before)
}

// ///
// Delete
// ///

// DeleteRow soft deletes, like the Nightscout API, so syncing clients see the delete
func (s *ActivitiesCollection) DeleteRow(ctx context.Context, collectionTx *siteapiutils.CollectionTx, id string, siteContext *httputils.SiteContext) error {
	return siteapiutils.SoftDelete(ctx, collectionTx, id, siteContext, (*sitedb.Queries).SoftDeleteActivity)
}

func (s *ActivitiesCollection) DeleteHTMXResponse(model *sitemodels.SiteActivity) templ.Component {
	return site_data_templ.ActivityDeleted()
}
