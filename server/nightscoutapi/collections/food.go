package collections

import (
	"context"
	"database/sql"
	"log/slog"

	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/models/nightscoutmodels"
	"nighttroop.t1d.nz/server/nightscoutapi/nsquery"
	"nighttroop.t1d.nz/server/nightscoutapi/nsutils"
	"nighttroop.t1d.nz/utils/httputils"
)

type FoodCollection struct {
	converter func(c *sitedb.FoodPreset, siteContext *httputils.SiteContext, raw bool) nsutils.Record
}

func NewFoodCollection() CollectionStore {
	return NewCollection[nightscoutmodels.Food, sitedb.FoodPreset](FoodCollection{
		converter: nsutils.DBModelToRecord(nsutils.Food, nightscoutmodels.DBModelToFood),
	})
}

func (c FoodCollection) Name() string {
	return "food"
}

func (c FoodCollection) Convert(dr *sitedb.FoodPreset, siteContext *httputils.SiteContext, raw bool) nsutils.Record {
	return c.converter(dr, siteContext, raw)
}

func (c FoodCollection) Validate(nsModel nightscoutmodels.Food, dbModel *sitedb.FoodPreset) *nsutils.HTTPError {
	if dbModel == nil {
		return nil
	}

	if nsModel.Name != nil && nsModel.Name.String() != dbModel.Name {
		slog.Error("cannot change name attribute on food preset", "existingName", dbModel.Name, "newName", nsModel.Name.String())
		return nsutils.ErrCannotChangeName
	}

	return nil
}

func (c FoodCollection) ValidatePatch(nsModel nightscoutmodels.Food, dbModel *sitedb.FoodPreset) *nsutils.HTTPError {
	return c.Validate(nsModel, dbModel)
}

func (c FoodCollection) CreateModel(q *sitedb.Queries, ctx context.Context, nsModel nightscoutmodels.Food, cc nightscoutmodels.ConverterContext, overrideIdentifier *string) (*sitedb.FoodPreset, error) {
	createParams := nightscoutmodels.FoodToCreate(nsModel, cc)
	if overrideIdentifier != nil {
		createParams.ID = *overrideIdentifier
	}
	return q.CreateFoodPreset(ctx, createParams)

}

func (c FoodCollection) GetDuplicate(q *sitedb.Queries, ctx context.Context, nsModel nightscoutmodels.Food) (*sitedb.FoodPreset, error) {
	// Without a date every filter below is empty and would match the newest row
	if nsModel.Date == nil {
		return nil, nil
	}
	foodPresets, rerr := q.SearchFoodPresets(ctx, sitedb.SearchFoodPresetsParams{
		DateEQ:   nsModel.Date.Int64p(),
		DeviceEq: nsModel.Device.Stringp(),
		Limit:    1,
	})
	if rerr != nil {
		return nil, rerr
	}
	if len(foodPresets) > 0 {
		return foodPresets[0], nil
	}
	return nil, nil
}

func (c FoodCollection) GetModel(q *sitedb.Queries, ctx context.Context, identifier string, includeDeleted bool) (*sitedb.FoodPreset, error) {
	return q.GetFoodPreset(ctx, sitedb.GetFoodPresetParams{ID: identifier})
}

func (c FoodCollection) HardDelete(q *sitedb.Queries, ctx context.Context, identifier string) (sql.Result, error) {
	return q.HardDeleteFoodPreset(ctx, sitedb.HardDeleteFoodPresetParams{ID: identifier})
}

// SoftDeleteDBFunc Only Hard Deletes For Presets
func (c FoodCollection) SoftDelete(q *sitedb.Queries, ctx context.Context, identifier string) (sql.Result, error) {
	return q.HardDeleteFoodPreset(ctx, sitedb.HardDeleteFoodPresetParams{ID: identifier})
}

func (c FoodCollection) HardDeleteType(q *sitedb.Queries, ctx context.Context, typeOF string) (sql.Result, error) {
	return q.HardDeleteFoodPresetsByType(ctx, sitedb.HardDeleteFoodPresetsByTypeParams{FoodType: &typeOF})
}

// SoftDeleteDBFunc Only Hard Deletes For Presets
func (c FoodCollection) SoftDeleteType(q *sitedb.Queries, ctx context.Context, typeOF string) (sql.Result, error) {
	return q.HardDeleteFoodPresetsByType(ctx, sitedb.HardDeleteFoodPresetsByTypeParams{FoodType: &typeOF})
}

func (c FoodCollection) UpdateModel(q *sitedb.Queries, ctx context.Context, nsModel nightscoutmodels.Food, identifier string, dbModel *sitedb.FoodPreset, uc nightscoutmodels.ConverterContext) (*sitedb.FoodPreset, error) {
	updateParams := nightscoutmodels.FoodToUpdate(nsModel, uc)
	updateParams.ID = identifier

	return q.UpdateFoodPreset(ctx, updateParams)
}

func (c FoodCollection) SearchModels(q *sitedb.Queries, ctx context.Context, query nsquery.Query) ([]*sitedb.FoodPreset, error) {
	if !query.HasFilters {
		return q.AllFoodPresets(ctx, nsquery.FoodPresetAll(query))
	}
	return q.SearchFoodPresets(ctx, nsquery.FoodPresetSearch(query))
}
