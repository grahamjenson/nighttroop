package collections

import (
	"context"
	"database/sql"
	"log/slog"

	"nighttroop.t1d.nz/data/sql/user"
	"nighttroop.t1d.nz/models/nightscout"
	"nighttroop.t1d.nz/server/nightscout_api/nsquery"
	"nighttroop.t1d.nz/server/nightscout_api/nsutils"
	"nighttroop.t1d.nz/utils"
	"nighttroop.t1d.nz/utils/httputils"
)

type ActivityCollection struct {
	converter func(c *user.Activity) nsutils.Record
}

func NewActivityCollection() ActivityCollection {
	return ActivityCollection{
		converter: DBModelToRecord(nsutils.Activity, nightscout.DBModelToActivity),
	}
}

func (c ActivityCollection) Create(ctx context.Context, userContext *httputils.UserContext, data map[string]any) (*nsutils.Record, *nsutils.HTTPError) {
	activityModel := &nightscout.Activity{}
	err := utils.FromMap(data, activityModel)

	if err != nil {
		slog.Error("failed to map data to activity model", "error", err)
		return nil, nsutils.BadDataBodyError
	}

	createParams := nightscout.ActivityToCreate(
		*activityModel,
		nightscout.CreateContext{IDPrefix: "ns", CreatedByIdentityID: userContext.IdentityModel.ID, CreatedByIdentityName: userContext.IdentityModel.Name},
	)

	dbModel, err := userContext.SiteDB.Queries.CreateActivity(ctx, createParams)
	if err != nil {
		slog.Error("failed to create activity in database", "error", err)
		return nil, nsutils.ErrInternal
	}
	if dbModel == nil {
		slog.Error("failed to create activity in database, dbModel is nil")
		return nil, nsutils.ErrInternal
	}
	return ptr(c.converter(dbModel)), nil
}

func (c ActivityCollection) Get(ctx context.Context, userContext *httputils.UserContext, identifier string) (*nsutils.Record, *nsutils.HTTPError) {
	dbModel, err := userContext.SiteDB.Queries.GetActivity(ctx, user.GetActivityParams{
		ID: identifier,
	})
	if err == sql.ErrNoRows {
		return nil, nsutils.ErrNotFound
	}
	if err != nil {
		slog.Error("failed to get activity", "error", err)
		return nil, nsutils.ErrInternal
	}
	if dbModel == nil {
		return nil, nsutils.ErrNotFound
	}
	record := c.converter(dbModel)
	return &record, nil
}

func (c ActivityCollection) Search(ctx context.Context, userContext *httputils.UserContext, query nsquery.Query) ([]nsutils.Record, *nsutils.HTTPError) {
	// TODO
	return nil, nsutils.ErrNotImplemented
}

func (c ActivityCollection) Update(ctx context.Context, userContext *httputils.UserContext, identifier string, patch map[string]any) (*nsutils.Record, *nsutils.HTTPError) {
	return nil, nsutils.ErrNotImplemented
}

func (c ActivityCollection) Delete(ctx context.Context, userContext *httputils.UserContext, identifier string, permanent bool) *nsutils.HTTPError {
	return nsutils.ErrNotImplemented
}

func (c ActivityCollection) DeleteType(ctx context.Context, userContext *httputils.UserContext, typeString string, permanent bool) (int, *nsutils.HTTPError) {
	return 0, nsutils.ErrNotImplemented
}
