package collections

import (
	"context"
	"database/sql"
	"log/slog"

	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/models/nightscoutmodels"
	"nighttroop.t1d.nz/server/nightscout_api/nsquery"
	"nighttroop.t1d.nz/server/nightscout_api/nsutils"
	"nighttroop.t1d.nz/utils/httputils"
)

type ActivityCollection struct {
	converter func(c *sitedb.Activity, siteContext *httputils.SiteContext, raw bool) nsutils.Record
}

func NewActivityCollection() ActivityCollection {
	return ActivityCollection{
		converter: nsutils.DBModelToRecord(nsutils.Activity, nightscoutmodels.DBModelToActivity),
	}
}

// ///
// FUNCTIONS USED BY META COLLECTIONS
// ///

func (c ActivityCollection) Name() string {
	return "activity"
}

func (c ActivityCollection) Convert(dr *sitedb.Activity, siteContext *httputils.SiteContext, raw bool) nsutils.Record {
	return c.converter(dr, siteContext, raw)
}

func (c ActivityCollection) Validate(nsModel nightscoutmodels.Activity, dbModel *sitedb.Activity) *nsutils.HTTPError {
	if dbModel == nil {
		return nil
	}

	if nsModel.ActivityType != nil && nsModel.ActivityType.String() != dbModel.ActivityType {
		slog.Error("cannot change activity_type attribute on activity", "existingActivityType", dbModel.ActivityType, "newActivityType", nsModel.ActivityType.String())
		return nsutils.ErrCannotChangeType
	}

	return nil
}

func (c ActivityCollection) ValidatePatch(nsModel nightscoutmodels.Activity, dbModel *sitedb.Activity) *nsutils.HTTPError {
	return c.Validate(nsModel, dbModel)
}

func (c ActivityCollection) CreateDBFunc(q *sitedb.Queries, ctx context.Context, nsModel nightscoutmodels.Activity, cc nightscoutmodels.ConverterContext, overrideIdentifier *string) (*sitedb.Activity, error) {
	createParams := nightscoutmodels.ActivityToCreate(nsModel, cc)
	if overrideIdentifier != nil {
		createParams.ID = *overrideIdentifier
	}
	return q.CreateActivity(ctx, createParams)
}

func (c ActivityCollection) GetDBFunc(q *sitedb.Queries, ctx context.Context, identifier string, includeDeleted bool) (*sitedb.Activity, error) {
	return q.GetActivity(ctx, sitedb.GetActivityParams{ID: identifier, IncludeDeleted: includeDeleted})
}

func (c ActivityCollection) GetDuplicateDBFunc(q *sitedb.Queries, ctx context.Context, nsModel nightscoutmodels.Activity) (*sitedb.Activity, error) {
	return nil, nil
}

func (c ActivityCollection) HardDeleteDBFunc(q *sitedb.Queries, ctx context.Context, identifier string) (sql.Result, error) {
	return q.HardDeleteActivity(ctx, sitedb.HardDeleteActivityParams{ID: identifier})
}

func (c ActivityCollection) SoftDeleteDBFunc(q *sitedb.Queries, ctx context.Context, identifier string) (sql.Result, error) {
	return q.SoftDeleteActivity(ctx, sitedb.SoftDeleteActivityParams{ID: identifier})
}

func (c ActivityCollection) HardDeleteTypeDBFunc(q *sitedb.Queries, ctx context.Context, typeString string) (sql.Result, error) {
	return q.HardDeleteActivitiesByType(ctx, sitedb.HardDeleteActivitiesByTypeParams{ActivityType: typeString})
}

func (c ActivityCollection) SoftDeleteTypeDBFunc(q *sitedb.Queries, ctx context.Context, typeString string) (sql.Result, error) {
	return q.SoftDeleteActivitiesByType(ctx, sitedb.SoftDeleteActivitiesByTypeParams{ActivityType: typeString})
}

func (c ActivityCollection) PatchDBFunc(q *sitedb.Queries, ctx context.Context, nsModel nightscoutmodels.Activity, identifier string, dbModel *sitedb.Activity, uc nightscoutmodels.ConverterContext) (*sitedb.Activity, error) {
	updateParams := nightscoutmodels.ActivityToUpdate(nsModel, uc)
	updateParams.ID = identifier

	return q.UpdateActivity(ctx, updateParams)
}

// ///
// END OF FUNCTIONS USED BY META COLLECTIONS
// ///

func (c ActivityCollection) Create(ctx context.Context, siteContext *httputils.SiteContext, data map[string]any, raw bool) (*nsutils.Record, bool, *nsutils.HTTPError) {
	return MetaCreate(c, ctx, siteContext, data, raw)
}

func (c ActivityCollection) Get(ctx context.Context, siteContext *httputils.SiteContext, identifier string, query nsquery.Query) (*nsutils.Record, *nsutils.HTTPError) {
	return MetaGet(c, ctx, siteContext, identifier, query.Raw, query.IncludeDeleted)
}

func (c ActivityCollection) Search(ctx context.Context, siteContext *httputils.SiteContext, query nsquery.Query) ([]nsutils.Record, *nsutils.HTTPError) {
	var activities []*sitedb.Activity
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

	return nsutils.CreateRecordSlice(activities, c.converter, siteContext, query.Raw), nil
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
