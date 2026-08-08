package collections

import (
	"context"

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

func (c FoodCollection) CreateDBFunc(siteContext *httputils.SiteContext, ctx context.Context, nsModel nightscout.Food, cc nightscout.CreateContext) (*site.FoodPreset, error) {
	return siteContext.SiteDB.Queries.CreateFoodPreset(ctx, nightscout.FoodToCreate(nsModel, cc))
}

func (c FoodCollection) UpdateSBFunc(siteContext *httputils.SiteContext, ctx context.Context, updateParams site.UpdateFoodPresetParams) (*site.FoodPreset, error) {
	return siteContext.SiteDB.Queries.UpdateFoodPreset(ctx, updateParams)
}

func (c FoodCollection) GetDBFunc(siteContext *httputils.SiteContext, ctx context.Context, identifier string) (*site.FoodPreset, error) {
	return siteContext.SiteDB.Queries.GetFoodPreset(ctx, site.GetFoodPresetParams{ID: identifier})
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

func (c FoodCollection) Search(ctx context.Context, siteContext *httputils.SiteContext, query nsquery.Query) ([]nsutils.Record, *nsutils.HTTPError) {
	return nil, nsutils.ErrNotImplemented
}

func (c FoodCollection) Delete(ctx context.Context, siteContext *httputils.SiteContext, identifier string, permanent bool) *nsutils.HTTPError {
	return nsutils.ErrNotImplemented
}

func (c FoodCollection) DeleteType(ctx context.Context, siteContext *httputils.SiteContext, typeString string, permanent bool) (int, *nsutils.HTTPError) {
	return 0, nsutils.ErrNotImplemented
}

func (c FoodCollection) Replace(ctx context.Context, siteContext *httputils.SiteContext, identifier string, patch map[string]any, canCreate bool) (*nsutils.Record, bool, *nsutils.HTTPError) {
	return nil, false, nsutils.ErrNotImplemented
}

func (c FoodCollection) Patch(ctx context.Context, siteContext *httputils.SiteContext, identifier string, patch map[string]any) (*nsutils.Record, *nsutils.HTTPError) {
	return nil, nsutils.ErrNotImplemented
}
