package sitedata

import (
	"cmp"
	"context"

	"nighttroop.t1d.nz/utils/apiutils"

	"github.com/a-h/templ"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/models/sitemodels"
	"nighttroop.t1d.nz/server/components/site_templ/site_data_templ"
	"nighttroop.t1d.nz/server/siteapi/siteapiutils"
	"nighttroop.t1d.nz/utils/httputils"
)

// maxFoodGrams is far more of any one nutrient than a meal has; as much again below 0 takes back food entered twice
const maxFoodGrams = 1000

type FoodIntakesCollection struct{}

// ////
// Shared
// ////
func (s *FoodIntakesCollection) GetName() string {
	return "food"
}

// GetRoute is where the collection is in its URL
func (s *FoodIntakesCollection) GetRoute() string {
	return "food"
}

// ////
// CREATE
// ////
func (s *FoodIntakesCollection) CreateRow(ctx context.Context, collectionTx *siteapiutils.CollectionTx, id *string, siteContext *httputils.SiteContext, model *sitemodels.SiteFoodIntake) (*sitedb.FoodIntake, error) {
	if model.Carbs == nil {
		return nil, apiutils.ErrMissingRequiredFieldWithHint("carbs")
	}
	if herr := checkFoodIntake(model); herr != nil {
		return nil, herr
	}
	return siteapiutils.SiteQuery(ctx, collectionTx, (*sitedb.Queries).CreateFoodIntake, sitemodels.FoodIntakeToCreate(model, siteContext))
}

// NewHTMXResponse is the modal to add a food intake in
func (s *FoodIntakesCollection) NewHTMXResponse() templ.Component {
	return site_data_templ.FoodIntakeNew()
}

func (s *FoodIntakesCollection) CreateHTMXResponse(model *sitemodels.SiteFoodIntake) templ.Component {
	return site_data_templ.FoodIntakeCreated(model)
}

// ////
// UPDATE
// ////

// UpdateRow changes a food intake's name, nutrients, notes or reason; anything not sent stays as it is
func (s *FoodIntakesCollection) UpdateRow(ctx context.Context, collectionTx *siteapiutils.CollectionTx, siteContext *httputils.SiteContext, oldModel *sitemodels.SiteFoodIntake, model *sitemodels.SiteFoodIntake) (*sitedb.FoodIntake, error) {
	if herr := checkFoodIntake(model); herr != nil {
		return nil, herr
	}
	model.ID = oldModel.ID
	return siteapiutils.SiteQuery(ctx, collectionTx, (*sitedb.Queries).UpdateFoodIntake, sitemodels.FoodIntakeToUpdate(model, siteContext))
}

func (s *FoodIntakesCollection) UpdateHTMXResponse(model *sitemodels.SiteFoodIntake) templ.Component {
	return site_data_templ.FoodIntakeUpdated(model)
}

// ////
// Get Model
// /////
func (s *FoodIntakesCollection) GetRow(ctx context.Context, collectionTx *siteapiutils.CollectionTx, id string, siteContext *httputils.SiteContext) (*sitedb.FoodIntake, error) {
	return siteapiutils.SiteQuery(ctx, collectionTx, (*sitedb.Queries).GetFoodIntake, sitedb.GetFoodIntakeParams{ID: id})
}

// ToModel is the site model of a row
func (s *FoodIntakesCollection) ToModel(row *sitedb.FoodIntake, siteContext *httputils.SiteContext) *sitemodels.SiteFoodIntake {
	return sitemodels.FoodIntakeFromDB(row)
}

// GetHTMXResponse is the food intake's details, shown in a modal
func (s *FoodIntakesCollection) GetHTMXResponse(model *sitemodels.SiteFoodIntake) templ.Component {
	return site_data_templ.FoodIntakeModal(model)
}

// ///
// Search Models
// ///

func (s *FoodIntakesCollection) SearchRows(ctx context.Context, collectionTx *siteapiutils.CollectionTx, query *apiutils.Query, siteContext *httputils.SiteContext) ([]*sitedb.FoodIntake, error) {
	return siteapiutils.SiteQuery(ctx, collectionTx, (*sitedb.Queries).AllFoodIntakes, siteapiutils.QueryToAllFoodIntakes(query))
}

// SearchTotals are how many the search pages through, and when the earliest and latest occurred
func (s *FoodIntakesCollection) SearchTotals(ctx context.Context, collectionTx *siteapiutils.CollectionTx, query *apiutils.Query, siteContext *httputils.SiteContext) (*apiutils.Totals, error) {
	return siteapiutils.SearchTotals(ctx, collectionTx, (*sitedb.Queries).FoodIntakeTotals)
}

func (s *FoodIntakesCollection) SearchHTMXResponse(models []*sitemodels.SiteFoodIntake, query *apiutils.Query) templ.Component {
	return site_data_templ.FoodIntakesTable(models, query.Skip, query.Limit)
}

// ///
// Delete
// ///

// DeleteRow soft deletes, like the Nightscout API, so syncing clients see the delete
func (s *FoodIntakesCollection) DeleteRow(ctx context.Context, collectionTx *siteapiutils.CollectionTx, id string, siteContext *httputils.SiteContext) error {
	return siteapiutils.SoftDelete(ctx, collectionTx, id, siteContext, (*sitedb.Queries).SoftDeleteFoodIntake)
}

func (s *FoodIntakesCollection) DeleteHTMXResponse(model *sitemodels.SiteFoodIntake) templ.Component {
	return site_data_templ.FoodIntakeDeleted()
}

///////
// UTIL METHODS
////////

// checkFoodIntake keeps any carbs, fat, protein or fibre sent within what a meal could have, or take back
func checkFoodIntake(model *sitemodels.SiteFoodIntake) *apiutils.HTTPError {
	return cmp.Or(
		apiutils.CheckRange("carbs, fat, protein and fibre", model.Carbs, -maxFoodGrams, maxFoodGrams, "g"),
		apiutils.CheckRange("carbs, fat, protein and fibre", model.Fat, -maxFoodGrams, maxFoodGrams, "g"),
		apiutils.CheckRange("carbs, fat, protein and fibre", model.Protein, -maxFoodGrams, maxFoodGrams, "g"),
		apiutils.CheckRange("carbs, fat, protein and fibre", model.Fibre, -maxFoodGrams, maxFoodGrams, "g"),
	)
}
