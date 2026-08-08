package collections

import (
	"context"

	"nighttroop.t1d.nz/data/sql/site"
	"nighttroop.t1d.nz/models/nightscout"
	"nighttroop.t1d.nz/server/nightscout_api/nsquery"
	"nighttroop.t1d.nz/server/nightscout_api/nsutils"
	"nighttroop.t1d.nz/utils/httputils"
)

type ActivityCollection struct {
	converter func(c *site.Activity) nsutils.Record
}

func NewActivityCollection() ActivityCollection {
	return ActivityCollection{
		converter: nsutils.DBModelToRecord(nsutils.Activity, nightscout.DBModelToActivity),
	}
}

// ///
// FUNCTIONS USED BY META COLLECTIONS
// ///

func (c ActivityCollection) Name() string {
	return "activity"
}

func (c ActivityCollection) Convert(dr *site.Activity) nsutils.Record {
	return c.converter(dr)
}

func (c ActivityCollection) CreateDBFunc(siteContext *httputils.SiteContext, ctx context.Context, nsModel nightscout.Activity, cc nightscout.CreateContext) (*site.Activity, error) {
	return siteContext.SiteDB.Queries.CreateActivity(ctx, nightscout.ActivityToCreate(nsModel, cc))
}

func (c ActivityCollection) UpdateSBFunc(siteContext *httputils.SiteContext, ctx context.Context, updateParams site.UpdateActivityParams) (*site.Activity, error) {
	return siteContext.SiteDB.Queries.UpdateActivity(ctx, updateParams)
}

func (c ActivityCollection) GetDBFunc(siteContext *httputils.SiteContext, ctx context.Context, identifier string) (*site.Activity, error) {
	return siteContext.SiteDB.Queries.GetActivity(ctx, site.GetActivityParams{ID: identifier})
}

// ///
// END OF FUNCTIONS USED BY META COLLECTIONS
// ///

func (c ActivityCollection) Create(ctx context.Context, siteContext *httputils.SiteContext, data map[string]any) (*nsutils.Record, *nsutils.HTTPError) {
	return MetaCreate(c, ctx, siteContext, data)
}

func (c ActivityCollection) Get(ctx context.Context, siteContext *httputils.SiteContext, identifier string) (*nsutils.Record, *nsutils.HTTPError) {
	return MetaGet(c, ctx, siteContext, identifier)
}

func (c ActivityCollection) Search(ctx context.Context, siteContext *httputils.SiteContext, query nsquery.Query) ([]nsutils.Record, *nsutils.HTTPError) {
	// TODO
	return nil, nsutils.ErrNotImplemented
}

func (c ActivityCollection) Delete(ctx context.Context, siteContext *httputils.SiteContext, identifier string, permanent bool) *nsutils.HTTPError {
	return nsutils.ErrNotImplemented
}

func (c ActivityCollection) DeleteType(ctx context.Context, siteContext *httputils.SiteContext, typeString string, permanent bool) (int, *nsutils.HTTPError) {
	return 0, nsutils.ErrNotImplemented
}

func (c ActivityCollection) Replace(ctx context.Context, siteContext *httputils.SiteContext, identifier string, patch map[string]any, canCreate bool) (*nsutils.Record, bool, *nsutils.HTTPError) {
	return nil, false, nsutils.ErrNotImplemented
}

func (c ActivityCollection) Patch(ctx context.Context, siteContext *httputils.SiteContext, identifier string, patch map[string]any) (*nsutils.Record, *nsutils.HTTPError) {
	return nil, nsutils.ErrNotImplemented
}
