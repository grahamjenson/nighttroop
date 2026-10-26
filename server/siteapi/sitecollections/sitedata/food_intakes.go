package sitedata

import (
	"context"
	"fmt"

	"nighttroop.t1d.nz/utils/apiutils"

	"github.com/a-h/templ"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/models/sitemodels"
	"nighttroop.t1d.nz/server/components/site_templ/site_data_templ"
	"nighttroop.t1d.nz/server/siteapi/siteapiutils"
	"nighttroop.t1d.nz/utils/httputils"
)

// maxFoodGrams is far more of any one nutrient than a meal has
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
	q, err := collectionTx.Site()
	if err != nil {
		return nil, err
	}
	return q.CreateFoodIntake(ctx, sitemodels.FoodIntakeToCreate(model, siteContext))
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
	q, err := collectionTx.Site()
	if err != nil {
		return nil, err
	}
	return q.UpdateFoodIntake(ctx, sitemodels.FoodIntakeToUpdate(model, siteContext))
}

func (s *FoodIntakesCollection) UpdateHTMXResponse(model *sitemodels.SiteFoodIntake) templ.Component {
	return site_data_templ.FoodIntakeUpdated(model)
}

// ////
// Get Model
// /////
func (s *FoodIntakesCollection) GetRow(ctx context.Context, collectionTx *siteapiutils.CollectionTx, id string, siteContext *httputils.SiteContext) (*sitedb.FoodIntake, error) {
	q, err := collectionTx.Site()
	if err != nil {
		return nil, err
	}
	return q.GetFoodIntake(ctx, sitedb.GetFoodIntakeParams{ID: id})
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
	q, err := collectionTx.Site()
	if err != nil {
		return nil, err
	}
	return q.AllFoodIntakes(ctx, siteapiutils.QueryToAllFoodIntakes(query))
}

func (s *FoodIntakesCollection) SearchHTMXResponse(models []*sitemodels.SiteFoodIntake, query *apiutils.Query) templ.Component {
	return site_data_templ.FoodIntakesTable(models, query.Skip, query.Limit)
}

// ///
// Delete
// ///

// DeleteRow soft deletes, like the Nightscout API, so syncing clients see the delete
func (s *FoodIntakesCollection) DeleteRow(ctx context.Context, collectionTx *siteapiutils.CollectionTx, id string, siteContext *httputils.SiteContext) error {
	q, err := collectionTx.Site()
	if err != nil {
		return err
	}
	_, err = q.SoftDeleteFoodIntake(ctx, sitedb.SoftDeleteFoodIntakeParams{
		UpdatedByIdentityName: &siteContext.IdentityModel.Name,
		ID:                    id,
	})
	return err
}

func (s *FoodIntakesCollection) DeleteHTMXResponse(model *sitemodels.SiteFoodIntake) templ.Component {
	return site_data_templ.FoodIntakeDeleted()
}

///////
// UTIL METHODS
////////

// checkFoodIntake keeps any carbs, fat, protein or fibre sent within what a meal could have
func checkFoodIntake(model *sitemodels.SiteFoodIntake) *apiutils.HTTPError {
	for _, grams := range []*float64{model.Carbs, model.Fat, model.Protein, model.Fibre} {
		if grams != nil && (*grams < 0 || *grams > maxFoodGrams) {
			return apiutils.ErrConstraintCheckFailedWithHint(fmt.Sprintf("carbs, fat, protein and fibre must be from 0 to %d g", maxFoodGrams))
		}
	}
	return nil
}
