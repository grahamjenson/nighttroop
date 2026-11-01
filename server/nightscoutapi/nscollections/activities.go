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
	named[sitedb.Activity]
}

func NewActivityCollection() CollectionStore {
	return NewCollection[nightscoutmodels.Activity, sitedb.Activity](ActivityCollection{
		named[sitedb.Activity]{nsutils.Activity, nsutils.DBModelToRecord(nsutils.Activity, nightscoutmodels.DBModelToActivity)},
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

func (c ActivityCollection) ValidatePatch(nsModel nightscoutmodels.Activity, dbModel *sitedb.Activity) *nsutils.HTTPError {
	return c.Validate(nsModel, dbModel)
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

func (c ActivityCollection) GetDuplicate(q *sitedb.Queries, ctx context.Context, nsModel nightscoutmodels.Activity) (*sitedb.Activity, error) {
	return nil, nil
}

func (c ActivityCollection) HardDelete(q *sitedb.Queries, ctx context.Context, identifier string) (sql.Result, error) {
	return q.HardDeleteActivity(ctx, sitedb.HardDeleteActivityParams{ID: identifier})
}

func (c ActivityCollection) SoftDelete(q *sitedb.Queries, ctx context.Context, identifier string) (sql.Result, error) {
	return q.SoftDeleteActivity(ctx, sitedb.SoftDeleteActivityParams{ID: identifier})
}

func (c ActivityCollection) HardDeleteType(q *sitedb.Queries, ctx context.Context, typeString string) (sql.Result, error) {
	return q.HardDeleteActivitiesByType(ctx, sitedb.HardDeleteActivitiesByTypeParams{ActivityType: typeString})
}

func (c ActivityCollection) SoftDeleteType(q *sitedb.Queries, ctx context.Context, typeString string) (sql.Result, error) {
	return q.SoftDeleteActivitiesByType(ctx, sitedb.SoftDeleteActivitiesByTypeParams{ActivityType: typeString})
}

func (c ActivityCollection) UpdateModel(q *sitedb.Queries, ctx context.Context, nsModel nightscoutmodels.Activity, identifier string, dbModel *sitedb.Activity, uc nightscoutmodels.ConverterContext) (*sitedb.Activity, error) {
	updateParams := nightscoutmodels.ActivityToUpdate(nsModel, uc)
	updateParams.ID = identifier

	return q.UpdateActivity(ctx, updateParams)
}

func (c ActivityCollection) SearchModels(q *sitedb.Queries, ctx context.Context, query nsquery.Query) ([]*sitedb.Activity, error) {
	if !query.HasFilters {
		return q.AllActivities(ctx, nsquery.ActivityAll(query))
	}
	return q.SearchActivities(ctx, nsquery.ActivitySearch(query))
}
