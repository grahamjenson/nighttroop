package collections

import (
	"context"
	"database/sql"
	"log/slog"

	"nighttroop.t1d.nz/data/sql/site"
	"nighttroop.t1d.nz/models/nightscout"
	"nighttroop.t1d.nz/server/nightscout_api/nsquery"
	"nighttroop.t1d.nz/server/nightscout_api/nsutils"
	"nighttroop.t1d.nz/utils"
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

func (c FoodCollection) Create(ctx context.Context, siteContext *httputils.SiteContext, data map[string]any) (*nsutils.Record, *nsutils.HTTPError) {
	foodModel := &nightscout.Food{}
	err := utils.FromMap(data, foodModel)

	if err != nil {
		slog.Error("failed to map data to food model", "error", err)
		return nil, nsutils.BadDataBodyError
	}

	createParams := nightscout.FoodToCreate(
		*foodModel,
		nightscout.CreateContext{IDPrefix: "ns", CreatedByIdentityID: siteContext.IdentityModel.ID, CreatedByIdentityName: siteContext.IdentityModel.Name},
	)

	dbModel, err := siteContext.SiteDB.Queries.CreateFoodPreset(ctx, createParams)
	if err != nil {
		slog.Error("failed to create food in database", "error", err)
		return nil, nsutils.ErrInternal
	}
	if dbModel == nil {
		slog.Error("failed to create food in database, dbModel is nil")
		return nil, nsutils.ErrInternal
	}
	return ptr(c.converter(dbModel)), nil
}

func (c FoodCollection) Get(ctx context.Context, siteContext *httputils.SiteContext, identifier string) (*nsutils.Record, *nsutils.HTTPError) {
	dbModel, err := siteContext.SiteDB.Queries.GetFoodPreset(ctx, site.GetFoodPresetParams{
		ID: identifier,
	})
	if err == sql.ErrNoRows {
		return nil, nsutils.ErrNotFound
	}
	if err != nil {
		slog.Error("failed to get food", "error", err)
		return nil, nsutils.ErrInternal
	}
	if dbModel == nil {
		return nil, nsutils.ErrNotFound
	}
	record := c.converter(dbModel)
	return &record, nil
}

func (c FoodCollection) Search(ctx context.Context, siteContext *httputils.SiteContext, query nsquery.Query) ([]nsutils.Record, *nsutils.HTTPError) {
	return nil, nsutils.ErrNotImplemented
}

func (c FoodCollection) Update(ctx context.Context, siteContext *httputils.SiteContext, identifier string, patch map[string]any) (*nsutils.Record, *nsutils.HTTPError) {
	return nil, nsutils.ErrNotImplemented
}

func (c FoodCollection) Delete(ctx context.Context, siteContext *httputils.SiteContext, identifier string, permanent bool) *nsutils.HTTPError {
	return nsutils.ErrNotImplemented
}

func (c FoodCollection) DeleteType(ctx context.Context, siteContext *httputils.SiteContext, typeString string, permanent bool) (int, *nsutils.HTTPError) {
	return 0, nsutils.ErrNotImplemented
}
