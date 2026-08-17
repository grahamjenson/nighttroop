package collections

import (
	"context"
	"database/sql"
	"log/slog"

	"nighttroop.t1d.nz/data/sql/site"
	"nighttroop.t1d.nz/models/nightscout"
	"nighttroop.t1d.nz/server/nightscout_api/nsquery"
	"nighttroop.t1d.nz/server/nightscout_api/nsutils"
	"nighttroop.t1d.nz/utils/httputils"
)

type ActivityCollection struct {
	converter func(c *site.Activity, raw bool) nsutils.Record
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

func (c ActivityCollection) Convert(dr *site.Activity, raw bool) nsutils.Record {
	return c.converter(dr, raw)
}

func (c ActivityCollection) Validate(nsModel nightscout.Activity, dbModel *site.Activity) *nsutils.HTTPError {
	if dbModel == nil {
		return nil
	}

	if nsModel.ActivityType != nil && nsModel.ActivityType.String() != dbModel.ActivityType {
		slog.Error("cannot change activity_type attribute on activity", "existingActivityType", dbModel.ActivityType, "newActivityType", nsModel.ActivityType.String())
		return nsutils.ErrCannotChangeType
	}

	return nil
}

func (c ActivityCollection) ValidatePatch(nsModel nightscout.Activity, dbModel *site.Activity) *nsutils.HTTPError {
	return c.Validate(nsModel, dbModel)
}

func (c ActivityCollection) CreateDBFunc(q *site.Queries, ctx context.Context, nsModel nightscout.Activity, cc nightscout.ConverterContext, overrideIdentifier *string) (*site.Activity, error) {
	createParams := nightscout.ActivityToCreate(nsModel, cc)
	if overrideIdentifier != nil {
		createParams.ID = *overrideIdentifier
	}
	return q.CreateActivity(ctx, createParams)
}

func (c ActivityCollection) GetDBFunc(q *site.Queries, ctx context.Context, identifier string) (*site.Activity, error) {
	return q.GetActivity(ctx, site.GetActivityParams{ID: identifier})
}

func (c ActivityCollection) HardDeleteDBFunc(q *site.Queries, ctx context.Context, identifier string) (sql.Result, error) {
	return q.HardDeleteActivity(ctx, site.HardDeleteActivityParams{ID: identifier})
}

func (c ActivityCollection) SoftDeleteDBFunc(q *site.Queries, ctx context.Context, identifier string) (sql.Result, error) {
	return q.SoftDeleteActivity(ctx, site.SoftDeleteActivityParams{ID: identifier})
}

func (c ActivityCollection) HardDeleteTypeDBFunc(q *site.Queries, ctx context.Context, typeString string) (sql.Result, error) {
	return q.HardDeleteActivitiesByType(ctx, site.HardDeleteActivitiesByTypeParams{ActivityType: typeString})
}

func (c ActivityCollection) SoftDeleteTypeDBFunc(q *site.Queries, ctx context.Context, typeString string) (sql.Result, error) {
	return q.SoftDeleteActivitiesByType(ctx, site.SoftDeleteActivitiesByTypeParams{ActivityType: typeString})
}

func (c ActivityCollection) PatchDBFunc(q *site.Queries, ctx context.Context, nsModel nightscout.Activity, identifier string, dbModel *site.Activity, uc nightscout.ConverterContext) (*site.Activity, error) {
	updateParams := nightscout.ActivityToUpdate(nsModel, uc)
	updateParams.ID = identifier

	return q.UpdateActivity(ctx, updateParams)
}

// ///
// END OF FUNCTIONS USED BY META COLLECTIONS
// ///

func (c ActivityCollection) Create(ctx context.Context, siteContext *httputils.SiteContext, data map[string]any, raw bool) (*nsutils.Record, *nsutils.HTTPError) {
	return MetaCreate(c, ctx, siteContext, data, raw)
}

func (c ActivityCollection) Get(ctx context.Context, siteContext *httputils.SiteContext, identifier string, query nsquery.Query) (*nsutils.Record, *nsutils.HTTPError) {
	return MetaGet(c, ctx, siteContext, identifier, query.Raw)
}

func (c ActivityCollection) Search(ctx context.Context, siteContext *httputils.SiteContext, query nsquery.Query) ([]nsutils.Record, *nsutils.HTTPError) {
	var activities []*site.Activity
	var rerr error

	// There are three options,
	if !query.HasFilters {
		activities, rerr = siteContext.SiteDB.Queries.AllActivities(ctx, nsquery.ActivityAll(query))
	} else {
		activities, rerr = siteContext.SiteDB.Queries.SearchActivities(ctx, nsquery.ActivitySearch(query))
	}

	if rerr != nil {
		slog.Error("failed to query activities from database", "error", rerr)
		return nil, nsutils.ErrInternal
	}

	return nsutils.CreateRecordSlice(activities, c.converter, query.Raw), nil
}

func (c ActivityCollection) Delete(ctx context.Context, siteContext *httputils.SiteContext, identifier string, permanent bool) *nsutils.HTTPError {
	return MetaDelete(c, ctx, siteContext, identifier, permanent)
}

func (c ActivityCollection) DeleteType(ctx context.Context, siteContext *httputils.SiteContext, typeString string, permanent bool) (int, *nsutils.HTTPError) {
	return MetaDeleteType(c, ctx, siteContext, typeString, permanent)
}

func (c ActivityCollection) Replace(ctx context.Context, siteContext *httputils.SiteContext, identifier string, patch map[string]any, canCreate bool, raw bool) (*nsutils.Record, bool, *nsutils.HTTPError) {
	return MetaReplace(c, ctx, siteContext, identifier, patch, canCreate, raw)
}

func (c ActivityCollection) Patch(ctx context.Context, siteContext *httputils.SiteContext, identifier string, patch map[string]any, raw bool) (*nsutils.Record, *nsutils.HTTPError) {
	return MetaPatch(c, ctx, siteContext, identifier, patch, raw)
}
