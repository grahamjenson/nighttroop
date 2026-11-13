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

type ActivityCollection struct {
	named[nightscoutmodels.Activity, sitedb.Activity]
}

func NewActivityCollection() CollectionStore {
	return NewCollection[nightscoutmodels.Activity, sitedb.Activity](ActivityCollection{
		named[nightscoutmodels.Activity, sitedb.Activity]{nsutils.Activity, nightscoutmodels.DBModelToActivity},
	})
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

func (c ActivityCollection) CreateModel(q *sitedb.Queries, ctx context.Context, nsModel nightscoutmodels.Activity, cc nightscoutmodels.ConverterContext, overrideIdentifier *string) (*sitedb.Activity, error) {
	createParams := nightscoutmodels.ActivityToCreate(nsModel, cc)
	if overrideIdentifier != nil {
		createParams.ID = *overrideIdentifier
	}
	return q.CreateActivity(ctx, createParams)
}

func (c ActivityCollection) GetModel(q *sitedb.Queries, ctx context.Context, identifier string, includeDeleted bool) (*sitedb.Activity, error) {
	return q.GetActivity(ctx, sitedb.GetActivityParams{ID: identifier, IncludeDeleted: includeDeleted})
}

func (c ActivityCollection) Delete(q *sitedb.Queries, ctx context.Context, identifier string, permanent bool) (sql.Result, error) {
	if permanent {
		return q.HardDeleteActivity(ctx, sitedb.HardDeleteActivityParams{ID: identifier})
	}
	return q.SoftDeleteActivity(ctx, sitedb.SoftDeleteActivityParams{ID: identifier})
}

func (c ActivityCollection) DeleteType(q *sitedb.Queries, ctx context.Context, typeString string, permanent bool) (sql.Result, error) {
	if permanent {
		return q.HardDeleteActivitiesByType(ctx, sitedb.HardDeleteActivitiesByTypeParams{ActivityType: typeString})
	}
	return q.SoftDeleteActivitiesByType(ctx, sitedb.SoftDeleteActivitiesByTypeParams{ActivityType: typeString})
}

func (c ActivityCollection) RewriteModel(q *sitedb.Queries, ctx context.Context, nsModel nightscoutmodels.Activity, _ map[string]any, stored *sitedb.Activity, cc nightscoutmodels.ConverterContext) (*sitedb.Activity, error) {
	params := nightscoutmodels.ActivityToCreate(nsModel, cc)
	params.ID = stored.ID
	return q.RewriteActivity(ctx, sitedb.RewriteActivityParams(params))
}

func (c ActivityCollection) AscendingIDs(q *sitedb.Queries, ctx context.Context, query nsquery.Query) ([]string, error) {
	return q.ActivityAscendingIDs(ctx, nsquery.ActivityAscending(query))
}

func (c ActivityCollection) HistoryIDs(q *sitedb.Queries, ctx context.Context, since, limit int64) ([]string, error) {
	return q.ActivityHistoryIDs(ctx, sitedb.ActivityHistoryIDsParams{Since: since, Limit: limit})
}

func (c ActivityCollection) SearchModels(q *sitedb.Queries, ctx context.Context, query nsquery.Query) ([]*sitedb.Activity, error) {
	if !query.HasFilters {
		return q.AllActivities(ctx, nsquery.ActivityAll(query))
	}
	return q.SearchActivities(ctx, nsquery.ActivitySearch(query))
}
