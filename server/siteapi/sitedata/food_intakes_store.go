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

// maxFoodGrams is far more of any one nutrient than a meal has
const maxFoodGrams = 1000

type FoodIntakesStore struct{}

// ////
// Shared
// ////
func (s *FoodIntakesStore) GetName() string {
	return "food_intakes"
}

// ////
// CREATE
// ////
func (s *FoodIntakesStore) CreateModel(ctx context.Context, tx *siteapiutils.StoreTx, id *string, siteContext *httputils.SiteContext, model *sitemodels.SiteFoodIntake) (bool, *apiutils.HTTPError) {
	if model.Carbs == nil {
		return false, apiutils.ErrMissingRequiredFieldWithHint("carbs")
	}
	if herr := checkFoodIntake(model); herr != nil {
		return false, herr
	}
	q, err := tx.Site()
	if err != nil {
		return false, apiutils.ErrInternalServerError
	}
	created, err := q.CreateFoodIntake(ctx, sitemodels.FoodIntakeToCreate(model, siteContext))
	if err != nil {
		return false, apiutils.WriteError(err, "food intake")
	}
	*model = *sitemodels.FoodIntakeFromDB(created)
	return true, nil
}

func (s *FoodIntakesStore) CreateHTMXResponse(model *sitemodels.SiteFoodIntake) templ.Component {
	return site_data_templ.FoodIntakeCreated(model)
}

// ////
// UPDATE
// ////

// UpdateModel changes a food intake's name, nutrients, notes or reason; anything not sent stays as it is
func (s *FoodIntakesStore) UpdateModel(ctx context.Context, tx *siteapiutils.StoreTx, id string, siteContext *httputils.SiteContext, oldModel *sitemodels.SiteFoodIntake, model *sitemodels.SiteFoodIntake) *apiutils.HTTPError {
	if herr := checkFoodIntake(model); herr != nil {
		return herr
	}
	q, err := tx.Site()
	if err != nil {
		return apiutils.ErrInternalServerError
	}
	model.ID = oldModel.ID
	updated, err := q.UpdateFoodIntake(ctx, sitemodels.FoodIntakeToUpdate(model, siteContext))
	if err == sql.ErrNoRows {
		return apiutils.ErrNotFound
	}
	if err != nil {
		return apiutils.WriteError(err, "food intake")
	}
	*model = *sitemodels.FoodIntakeFromDB(updated)
	return nil
}

func (s *FoodIntakesStore) UpdateHTMXResponse(model *sitemodels.SiteFoodIntake) templ.Component {
	return site_data_templ.FoodIntakeUpdated(model)
}

// ////
// Get Model
// /////
func (s *FoodIntakesStore) GetModel(ctx context.Context, tx *siteapiutils.StoreTx, id string, siteContext *httputils.SiteContext) (*sitemodels.SiteFoodIntake, *apiutils.HTTPError) {
	q, err := tx.Site()
	if err != nil {
		return nil, apiutils.ErrInternalServerError
	}
	foodIntake, err := q.GetFoodIntake(ctx, sitedb.GetFoodIntakeParams{ID: id})
	if err == sql.ErrNoRows || (err == nil && foodIntake == nil) {
		return nil, nil
	}
	if err != nil {
		slog.Error("Error fetching food intake", "error", err)
		return nil, apiutils.ErrInternalServerError
	}
	return sitemodels.FoodIntakeFromDB(foodIntake), nil
}

// GetHTMXResponse is the food intake's details, shown in a modal
func (s *FoodIntakesStore) GetHTMXResponse(model *sitemodels.SiteFoodIntake) templ.Component {
	return site_data_templ.FoodIntakeModal(model)
}

// ///
// Search Models
// ///

func (s *FoodIntakesStore) SearchModels(ctx context.Context, tx *siteapiutils.StoreTx, query *apiutils.Query, siteContext *httputils.SiteContext) ([]*sitemodels.SiteFoodIntake, *apiutils.HTTPError) {
	q, err := tx.Site()
	if err != nil {
		return nil, apiutils.ErrInternalServerError
	}
	foodIntakes, err := q.AllFoodIntakes(ctx, sitequery.QueryToAllFoodIntakes(query))
	if err != nil {
		slog.Error("Search food intakes", "error", err)
		return nil, apiutils.ErrInternalServerError
	}
	models := []*sitemodels.SiteFoodIntake{}
	for _, foodIntake := range foodIntakes {
		models = append(models, sitemodels.FoodIntakeFromDB(foodIntake))
	}
	return models, nil
}

func (s *FoodIntakesStore) SearchHTMXResponse(models []*sitemodels.SiteFoodIntake, query *apiutils.Query) templ.Component {
	return site_data_templ.FoodIntakesTable(models, query.Skip, query.Limit)
}

// ///
// Delete
// ///

// DeleteModel soft deletes, like the Nightscout API, so syncing clients see the delete
func (s *FoodIntakesStore) DeleteModel(ctx context.Context, tx *siteapiutils.StoreTx, model *sitemodels.SiteFoodIntake, siteContext *httputils.SiteContext) *apiutils.HTTPError {
	q, err := tx.Site()
	if err != nil {
		return apiutils.ErrInternalServerError
	}
	_, err = q.SoftDeleteFoodIntake(ctx, sitedb.SoftDeleteFoodIntakeParams{
		UpdatedByIdentityName: &siteContext.IdentityModel.Name,
		ID:                    model.ID,
	})
	if err != nil {
		slog.Error("Error deleting food intake", "error", err)
		return apiutils.ErrInternalServerError
	}
	return nil
}

func (s *FoodIntakesStore) DeleteHTMXResponse(model *sitemodels.SiteFoodIntake) templ.Component {
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
