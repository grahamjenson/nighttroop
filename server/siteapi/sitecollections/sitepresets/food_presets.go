package sitepresets

import (
	"cmp"
	"context"

	"github.com/a-h/templ"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/models/sitemodels"
	"nighttroop.t1d.nz/server/components/site_templ/site_presets_templ"
	"nighttroop.t1d.nz/server/siteapi/siteapiutils"
	"nighttroop.t1d.nz/utils/apiutils"
	"nighttroop.t1d.nz/utils/httputils"
)

// maxPresetGrams is far more of any one nutrient than a portion has
const maxPresetGrams = 1000

// maxAbsorptionMinutes is a day, far longer than any food takes
const maxAbsorptionMinutes = 24 * 60

type FoodPresetsCollection struct{}

func (s *FoodPresetsCollection) GetName() string {
	return "food_presets"
}

// GetRoute is where the collection is in its URL, /presets/food
func (s *FoodPresetsCollection) GetRoute() string {
	return "food"
}

func (s *FoodPresetsCollection) CreateRow(ctx context.Context, collectionTx *siteapiutils.CollectionTx, id *string, siteContext *httputils.SiteContext, model *sitemodels.SiteFoodPreset) (*sitedb.FoodPreset, error) {
	if model.Name == nil || *model.Name == "" {
		return nil, apiutils.ErrMissingRequiredFieldWithHint("name")
	}
	if herr := checkFoodPreset(model); herr != nil {
		return nil, herr
	}
	q, err := collectionTx.Site()
	if err != nil {
		return nil, err
	}
	return q.CreateFoodPreset(ctx, sitemodels.FoodPresetToCreate(model, siteContext))
}

func (s *FoodPresetsCollection) CreateHTMXResponse(model *sitemodels.SiteFoodPreset) templ.Component {
	return site_presets_templ.FoodPresetCreated(model)
}

// UpdateRow changes what is sent of a food preset; anything not sent stays as it is
func (s *FoodPresetsCollection) UpdateRow(ctx context.Context, collectionTx *siteapiutils.CollectionTx, siteContext *httputils.SiteContext, oldModel *sitemodels.SiteFoodPreset, model *sitemodels.SiteFoodPreset) (*sitedb.FoodPreset, error) {
	if model.Name != nil && *model.Name == "" {
		return nil, apiutils.ErrMissingRequiredFieldWithHint("name")
	}
	if herr := checkFoodPreset(model); herr != nil {
		return nil, herr
	}
	model.ID = oldModel.ID
	q, err := collectionTx.Site()
	if err != nil {
		return nil, err
	}
	return q.UpdateFoodPreset(ctx, sitemodels.FoodPresetToUpdate(model, siteContext))
}

func (s *FoodPresetsCollection) UpdateHTMXResponse(model *sitemodels.SiteFoodPreset) templ.Component {
	return site_presets_templ.FoodPresetUpdated(model)
}

func (s *FoodPresetsCollection) GetRow(ctx context.Context, collectionTx *siteapiutils.CollectionTx, id string, siteContext *httputils.SiteContext) (*sitedb.FoodPreset, error) {
	q, err := collectionTx.Site()
	if err != nil {
		return nil, err
	}
	return q.GetFoodPreset(ctx, sitedb.GetFoodPresetParams{ID: id})
}

// ToModel is the site model of a row
func (s *FoodPresetsCollection) ToModel(row *sitedb.FoodPreset, siteContext *httputils.SiteContext) *sitemodels.SiteFoodPreset {
	return sitemodels.FoodPresetFromDB(row)
}

// GetHTMXResponse is the food preset's details, shown in a modal
func (s *FoodPresetsCollection) GetHTMXResponse(model *sitemodels.SiteFoodPreset) templ.Component {
	return site_presets_templ.FoodPresetModal(model)
}

func (s *FoodPresetsCollection) SearchRows(ctx context.Context, collectionTx *siteapiutils.CollectionTx, query *apiutils.Query, siteContext *httputils.SiteContext) ([]*sitedb.FoodPreset, error) {
	q, err := collectionTx.Site()
	if err != nil {
		return nil, err
	}
	return q.AllFoodPresets(ctx, siteapiutils.QueryToAllFoodPresets(query))
}

func (s *FoodPresetsCollection) SearchHTMXResponse(models []*sitemodels.SiteFoodPreset, query *apiutils.Query) templ.Component {
	return site_presets_templ.FoodPresetsTable(models, query.Skip, query.Limit)
}

// DeleteRow hard deletes, as the Nightscout API does presets, which are not events with a history to keep
func (s *FoodPresetsCollection) DeleteRow(ctx context.Context, collectionTx *siteapiutils.CollectionTx, id string, siteContext *httputils.SiteContext) error {
	q, err := collectionTx.Site()
	if err != nil {
		return err
	}
	_, err = q.HardDeleteFoodPreset(ctx, sitedb.HardDeleteFoodPresetParams{ID: id})
	return err
}

func (s *FoodPresetsCollection) DeleteHTMXResponse(model *sitemodels.SiteFoodPreset) templ.Component {
	return site_presets_templ.FoodPresetDeleted()
}

// checkFoodPreset keeps any nutrients and absorption time sent within what a portion of food could have
func checkFoodPreset(model *sitemodels.SiteFoodPreset) *apiutils.HTTPError {
	return cmp.Or(
		apiutils.CheckRange("carbs, fat, protein and fibre", model.Carbs, 0, maxPresetGrams, "g"),
		apiutils.CheckRange("carbs, fat, protein and fibre", model.Fat, 0, maxPresetGrams, "g"),
		apiutils.CheckRange("carbs, fat, protein and fibre", model.Protein, 0, maxPresetGrams, "g"),
		apiutils.CheckRange("carbs, fat, protein and fibre", model.Fibre, 0, maxPresetGrams, "g"),
		apiutils.CheckRange("absorption time", model.AbsorptionTimeMinutes, 0, maxAbsorptionMinutes, "minutes"),
	)
}
