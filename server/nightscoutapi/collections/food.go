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

func NewFoodCollection() FoodCollection {
	return FoodCollection{
		converter: nsutils.DBModelToRecord(nsutils.Food, nightscoutmodels.DBModelToFood),
	}
}

// ///
// FUNCTIONS USED BY META COLLECTIONS
// ///

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

func (c FoodCollection) CreateDBFunc(q *sitedb.Queries, ctx context.Context, nsModel nightscoutmodels.Food, cc nightscoutmodels.ConverterContext, overrideIdentifier *string) (*sitedb.FoodPreset, error) {
	createParams := nightscoutmodels.FoodToCreate(nsModel, cc)
	if overrideIdentifier != nil {
		createParams.ID = *overrideIdentifier
	}
	return q.CreateFoodPreset(ctx, createParams)

}

func (c FoodCollection) GetDuplicateDBFunc(q *sitedb.Queries, ctx context.Context, nsModel nightscoutmodels.Food) (*sitedb.FoodPreset, error) {
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

func (c FoodCollection) GetDBFunc(q *sitedb.Queries, ctx context.Context, identifier string, includeDeleted bool) (*sitedb.FoodPreset, error) {
	return q.GetFoodPreset(ctx, sitedb.GetFoodPresetParams{ID: identifier})
}

func (c FoodCollection) HardDeleteDBFunc(q *sitedb.Queries, ctx context.Context, identifier string) (sql.Result, error) {
	return q.HardDeleteFoodPreset(ctx, sitedb.HardDeleteFoodPresetParams{ID: identifier})
}

// SoftDeleteDBFunc Only Hard Deletes For Presets
func (c FoodCollection) SoftDeleteDBFunc(q *sitedb.Queries, ctx context.Context, identifier string) (sql.Result, error) {
	return q.HardDeleteFoodPreset(ctx, sitedb.HardDeleteFoodPresetParams{ID: identifier})
}

func (c FoodCollection) HardDeleteTypeDBFunc(q *sitedb.Queries, ctx context.Context, typeOF string) (sql.Result, error) {
	return q.HardDeleteFoodPresetsByType(ctx, sitedb.HardDeleteFoodPresetsByTypeParams{FoodType: &typeOF})
}

// SoftDeleteDBFunc Only Hard Deletes For Presets
func (c FoodCollection) SoftDeleteTypeDBFunc(q *sitedb.Queries, ctx context.Context, typeOF string) (sql.Result, error) {
	return q.HardDeleteFoodPresetsByType(ctx, sitedb.HardDeleteFoodPresetsByTypeParams{FoodType: &typeOF})
}

func (c FoodCollection) PatchDBFunc(q *sitedb.Queries, ctx context.Context, nsModel nightscoutmodels.Food, identifier string, dbModel *sitedb.FoodPreset, uc nightscoutmodels.ConverterContext) (*sitedb.FoodPreset, error) {
	updateParams := nightscoutmodels.FoodToUpdate(nsModel, uc)
	updateParams.ID = identifier

	return q.UpdateFoodPreset(ctx, updateParams)
}

// ///
// END OF FUNCTIONS USED BY META COLLECTIONS
// ///

func (c FoodCollection) Create(ctx context.Context, siteContext *httputils.SiteContext, data map[string]any, raw bool) (*nsutils.Record, bool, *nsutils.HTTPError) {
	return MetaCreate(c, ctx, siteContext, data, raw)
}

func (c FoodCollection) Get(ctx context.Context, siteContext *httputils.SiteContext, identifier string, query nsquery.Query) (*nsutils.Record, *nsutils.HTTPError) {
	return MetaGet(c, ctx, siteContext, identifier, query.Raw, query.IncludeDeleted)
}

func (c FoodCollection) Patch(ctx context.Context, siteContext *httputils.SiteContext, identifier string, patch map[string]any, raw bool) (*nsutils.Record, *nsutils.HTTPError) {
	return MetaPatch(c, ctx, siteContext, identifier, patch, raw)
}

func (c FoodCollection) Delete(ctx context.Context, siteContext *httputils.SiteContext, identifier string, permanent bool) *nsutils.HTTPError {
	return MetaDelete(c, ctx, siteContext, identifier, permanent)
}

func (c FoodCollection) DeleteType(ctx context.Context, siteContext *httputils.SiteContext, typeString string, permanent bool) (int, *nsutils.HTTPError) {
	return MetaDeleteType(c, ctx, siteContext, typeString, permanent)
}

func (c FoodCollection) Search(ctx context.Context, siteContext *httputils.SiteContext, query nsquery.Query) ([]nsutils.Record, *nsutils.HTTPError) {
	var foodPresets []*sitedb.FoodPreset
	var rerr error

	// There are three options,
	if !query.HasFilters {
		foodPresets, rerr = siteContext.SiteDB.Queries.AllFoodPresets(ctx, nsquery.FoodPresetAll(query))
	} else {
		foodPresets, rerr = siteContext.SiteDB.Queries.SearchFoodPresets(ctx, nsquery.FoodPresetSearch(query))
	}

	if rerr != nil {
		slog.Error("failed to query food presets from database", "error", rerr)
		return nil, nsutils.ErrInternal
	}

	return nsutils.CreateRecordSlice(foodPresets, c.converter, siteContext, query.Raw), nil
}

func (c FoodCollection) Replace(ctx context.Context, siteContext *httputils.SiteContext, identifier string, patch map[string]any, canCreate bool, raw bool) (*nsutils.Record, bool, *nsutils.HTTPError) {
	return MetaReplace(c, ctx, siteContext, identifier, patch, canCreate, raw)
}
