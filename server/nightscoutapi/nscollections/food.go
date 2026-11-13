package nscollections

import (
	"context"
	"database/sql"
	"log/slog"

	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/models/nightscoutmodels"
	"nighttroop.t1d.nz/server/nightscoutapi/nsquery"
	"nighttroop.t1d.nz/server/nightscoutapi/nsutils"
)

type FoodCollection struct {
	named[nightscoutmodels.Food, sitedb.FoodPreset]
}

func NewFoodCollection() CollectionStore {
	return NewCollection[nightscoutmodels.Food, sitedb.FoodPreset](FoodCollection{
		named[nightscoutmodels.Food, sitedb.FoodPreset]{nsutils.Food, nightscoutmodels.DBModelToFood},
	})
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

func (c FoodCollection) Rows(nsModel nightscoutmodels.Food, _ *sitedb.FoodPreset) []sitedb.TableRow {
	return []sitedb.TableRow{{Table: "food_presets", Row: nightscoutmodels.FoodToCreate(nsModel, nightscoutmodels.ConverterContext{})}}
}

func (c FoodCollection) CreateModel(q *sitedb.Queries, ctx context.Context, nsModel nightscoutmodels.Food, cc nightscoutmodels.ConverterContext, overrideIdentifier *string) (*sitedb.FoodPreset, error) {
	createParams := nightscoutmodels.FoodToCreate(nsModel, cc)
	if overrideIdentifier != nil {
		createParams.ID = *overrideIdentifier
	}
	return q.CreateFoodPreset(ctx, createParams)

}

func (c FoodCollection) GetModel(q *sitedb.Queries, ctx context.Context, identifier string, includeDeleted bool) (*sitedb.FoodPreset, error) {
	return q.GetFoodPreset(ctx, sitedb.GetFoodPresetParams{ID: identifier})
}

// Delete always deletes a preset for good, as presets are not soft deleted
func (c FoodCollection) Delete(q *sitedb.Queries, ctx context.Context, identifier string, _ bool) (sql.Result, error) {
	return q.HardDeleteFoodPreset(ctx, sitedb.HardDeleteFoodPresetParams{ID: identifier})
}

// DeleteType always deletes presets for good, as presets are not soft deleted
func (c FoodCollection) DeleteType(q *sitedb.Queries, ctx context.Context, typeOF string, _ bool) (sql.Result, error) {
	return q.HardDeleteFoodPresetsByType(ctx, sitedb.HardDeleteFoodPresetsByTypeParams{FoodType: &typeOF})
}

func (c FoodCollection) UpdateModel(q *sitedb.Queries, ctx context.Context, nsModel nightscoutmodels.Food, identifier string, dbModel *sitedb.FoodPreset, uc nightscoutmodels.ConverterContext) (*sitedb.FoodPreset, error) {
	updateParams := nightscoutmodels.FoodToUpdate(nsModel, uc)
	updateParams.ID = identifier

	return q.UpdateFoodPreset(ctx, updateParams)
}

func (c FoodCollection) AscendingIDs(q *sitedb.Queries, ctx context.Context, query nsquery.Query) ([]string, error) {
	return q.FoodPresetAscendingIDs(ctx, nsquery.FoodPresetAscending(query))
}

func (c FoodCollection) HistoryIDs(q *sitedb.Queries, ctx context.Context, since, limit int64) ([]string, error) {
	return q.FoodPresetHistoryIDs(ctx, sitedb.FoodPresetHistoryIDsParams{Since: since, Limit: limit})
}

func (c FoodCollection) SearchModels(q *sitedb.Queries, ctx context.Context, query nsquery.Query) ([]*sitedb.FoodPreset, error) {
	if !query.HasFilters {
		return q.AllFoodPresets(ctx, nsquery.FoodPresetAll(query))
	}
	return q.SearchFoodPresets(ctx, nsquery.FoodPresetSearch(query))
}
