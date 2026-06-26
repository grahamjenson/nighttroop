package collections

import (
	"context"
	"database/sql"
	"log/slog"

	"nighttroop.t1d.nz/data/sql/user"
	"nighttroop.t1d.nz/models/nightscout"
	"nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/nsutils"
	"nighttroop.t1d.nz/utils"
	"nighttroop.t1d.nz/utils/httputils"
)

type ProfileCollection struct {
	converter func(c *user.Profile) nsutils.Record
}

func NewProfileCollection() ProfileCollection {
	return ProfileCollection{
		converter: DBModelToRecord(nsutils.Profile, nightscout.DBModelToProfile),
	}
}

func (c ProfileCollection) Create(ctx context.Context, userContext *httputils.UserContext, data map[string]any) (*nsutils.Record, *nsutils.HTTPError) {
	profileModel := &nightscout.Profile{}
	err := utils.FromMap(data, profileModel)

	if err != nil {
		slog.Error("failed to map data to profile model", "error", err)
		return nil, nsutils.BadDataBodyError
	}

	createParams := nightscout.ProfileToCreate(
		*profileModel,
		nightscout.CreateContext{IDPrefix: "ns", CreatedByIdentityID: userContext.IdentityModel.ID, CreatedByIdentityName: userContext.IdentityModel.Name},
	)

	dbModel, err := userContext.UserDB.Queries.CreateProfile(ctx, createParams)
	if err != nil {
		slog.Error("failed to create profile in database", "error", err)
		return nil, nsutils.ErrInternal
	}
	if dbModel == nil {
		slog.Error("failed to create profile in database, dbModel is nil")
		return nil, nsutils.ErrInternal
	}
	return ptr(c.converter(dbModel)), nil
}

func (c ProfileCollection) Get(ctx context.Context, userContext *httputils.UserContext, identifier string) (*nsutils.Record, *nsutils.HTTPError) {
	dbModel, err := userContext.UserDB.Queries.GetProfile(ctx, identifier)
	if err == sql.ErrNoRows {
		return nil, nsutils.ErrNotFound
	}
	if err != nil {
		slog.Error("failed to get profile", "error", err)
		return nil, nsutils.ErrInternal
	}
	if dbModel == nil {
		return nil, nsutils.ErrNotFound
	}
	record := c.converter(dbModel)
	return &record, nil
}

func (c ProfileCollection) Search(ctx context.Context, userContext *httputils.UserContext, query *nsutils.Query) ([]nsutils.Record, *nsutils.HTTPError) {
	return nil, nsutils.ErrNotImplemented
}

func (c ProfileCollection) Update(ctx context.Context, userContext *httputils.UserContext, identifier string, patch map[string]any) (*nsutils.Record, *nsutils.HTTPError) {
	return nil, nsutils.ErrNotImplemented
}

func (c ProfileCollection) Delete(ctx context.Context, userContext *httputils.UserContext, identifier string, permanent bool) *nsutils.HTTPError {
	return nsutils.ErrNotImplemented
}

func (c ProfileCollection) DeleteType(ctx context.Context, userContext *httputils.UserContext, typeString string, permanent bool) (int, *nsutils.HTTPError) {
	return 0, nsutils.ErrNotImplemented
}
