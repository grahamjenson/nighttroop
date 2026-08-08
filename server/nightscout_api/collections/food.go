package collections

import (
	"context"
	"database/sql"

	"nighttroop.t1d.nz/data/sql/site"
	"nighttroop.t1d.nz/models/nightscout"
	"nighttroop.t1d.nz/server/nightscout_api/nsquery"
	"nighttroop.t1d.nz/server/nightscout_api/nsutils"
	"nighttroop.t1d.nz/utils/httputils"
)

type FoodCollection struct {
	converter func(c *site.FoodPreset) nsutils.Record
}

func NewFoodCollection() FoodCollection {
	return FoodCollection{
		converter: nsutils.DBModelToRecord(nsutils.Food, nightscout.DBModelToFood),
	}
}

// ///
// FUNCTIONS USED BY META COLLECTIONS
// ///

func (c FoodCollection) Name() string {
	return "food"
}

func (c FoodCollection) Convert(dr *site.FoodPreset) nsutils.Record {
	return c.converter(dr)
}

func (c FoodCollection) CreateDBFunc(q *site.Queries, ctx context.Context, nsModel nightscout.Food, cc nightscout.CreateContext, replacing *site.FoodPreset, overrideIdentifier *string) (*site.FoodPreset, error) {
	createParams := nightscout.FoodToCreate(nsModel, cc)
	if overrideIdentifier != nil {
		createParams.ID = *overrideIdentifier
	}
	return q.CreateFoodPreset(ctx, createParams)

}

func (c FoodCollection) UpdateSBFunc(q *site.Queries, ctx context.Context, updateParams site.UpdateFoodPresetParams) (*site.FoodPreset, error) {
	return q.UpdateFoodPreset(ctx, updateParams)
}

func (c FoodCollection) GetDBFunc(q *site.Queries, ctx context.Context, identifier string) (*site.FoodPreset, error) {
	return q.GetFoodPreset(ctx, site.GetFoodPresetParams{ID: identifier})
}

func (c FoodCollection) HardDeleteDBFunc(q *site.Queries, ctx context.Context, identifier string) (sql.Result, error) {
	return q.HardDeleteFoodPreset(ctx, site.HardDeleteFoodPresetParams{ID: identifier})
}

// SoftDeleteDBFunc Only Hard Deletes For Presets
func (c FoodCollection) SoftDeleteDBFunc(q *site.Queries, ctx context.Context, identifier string) (sql.Result, error) {
	return q.HardDeleteFoodPreset(ctx, site.HardDeleteFoodPresetParams{ID: identifier})
}

func (c FoodCollection) HardDeleteTypeDBFunc(q *site.Queries, ctx context.Context, typeString string) (sql.Result, error) {
	return q.HardDeleteFoodPresetsByType(ctx, site.HardDeleteFoodPresetsByTypeParams{FoodType: &typeString})
}

// SoftDeleteTypeDBFunc Only Hard Deletes For Presets
func (c FoodCollection) SoftDeleteTypeDBFunc(q *site.Queries, ctx context.Context, typeString string) (sql.Result, error) {
	return q.HardDeleteFoodPresetsByType(ctx, site.HardDeleteFoodPresetsByTypeParams{FoodType: &typeString})
}

func (c FoodCollection) PatchDBFunc(q *site.Queries, ctx context.Context, nsModel nightscout.Food, identifier string, uc nightscout.UpdateContext) (*site.FoodPreset, error) {
	updateParams := nightscout.FoodToUpdate(nsModel, uc)
	updateParams.ID = identifier

	return q.UpdateFoodPreset(ctx, updateParams)
}

// ///
// END OF FUNCTIONS USED BY META COLLECTIONS
// ///

func (c FoodCollection) Create(ctx context.Context, siteContext *httputils.SiteContext, data map[string]any) (*nsutils.Record, *nsutils.HTTPError) {
	return MetaCreate(c, ctx, siteContext, data)
}

func (c FoodCollection) Get(ctx context.Context, siteContext *httputils.SiteContext, identifier string) (*nsutils.Record, *nsutils.HTTPError) {
	return MetaGet(c, ctx, siteContext, identifier)
}

func (c FoodCollection) Patch(ctx context.Context, siteContext *httputils.SiteContext, identifier string, patch map[string]any) (*nsutils.Record, *nsutils.HTTPError) {
	return MetaPatch(c, ctx, siteContext, identifier, patch)
}

func (c FoodCollection) Delete(ctx context.Context, siteContext *httputils.SiteContext, identifier string, permanent bool) *nsutils.HTTPError {
	return MetaDelete(c, ctx, siteContext, identifier, permanent)
}

func (c FoodCollection) DeleteType(ctx context.Context, siteContext *httputils.SiteContext, typeString string, permanent bool) (int, *nsutils.HTTPError) {
	return MetaDeleteType(c, ctx, siteContext, typeString, permanent)
}

func (c FoodCollection) Search(ctx context.Context, siteContext *httputils.SiteContext, query nsquery.Query) ([]nsutils.Record, *nsutils.HTTPError) {
	return nil, nsutils.ErrNotImplemented
}

func (c FoodCollection) Replace(ctx context.Context, siteContext *httputils.SiteContext, identifier string, patch map[string]any, canCreate bool) (*nsutils.Record, bool, *nsutils.HTTPError) {
	return nil, false, nsutils.ErrNotImplemented
}
