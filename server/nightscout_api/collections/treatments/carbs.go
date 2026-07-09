package treatments

import (
	"context"
	"log/slog"

	"nighttroop.t1d.nz/data/sql/site"
	"nighttroop.t1d.nz/models/nightscout"
	"nighttroop.t1d.nz/server/nightscout_api/nsquery"
	"nighttroop.t1d.nz/server/nightscout_api/nsutils"
	"nighttroop.t1d.nz/utils/httputils"
)

type CarbsCollection struct {
	converter func(c *site.Carb) nsutils.Record
}

func NewCarbsCollection() CarbsCollection {
	return CarbsCollection{
		converter: nsutils.DBModelToRecord(nsutils.Treatments, nightscout.CarbsDBModelToTreatment),
	}
}

func (c CarbsCollection) GetPrefix() string {
	return "cc"
}

func (c CarbsCollection) Create(ctx context.Context, siteContext *httputils.SiteContext, treatment nightscout.Treatment) (*nsutils.Record, *nsutils.HTTPError) {
	dbModel, err := siteContext.SiteDB.Queries.CreateCarb(ctx,
		nightscout.TreatmentToCreateCarb(treatment, nightscout.CreateContext{
			IDPrefix:              c.GetPrefix(),
			CreatedByIdentityID:   siteContext.IdentityModel.ID,
			CreatedByIdentityName: siteContext.IdentityModel.Name,
		},
		),
	)
	if err != nil {
		slog.Error("failed to create carb in database", "error", err)
		return nil, nsutils.ErrInternal
	}
	return ptr(c.converter(dbModel)), nil
}

func (c CarbsCollection) Get(ctx context.Context, siteContext *httputils.SiteContext, identifier string) (*nsutils.Record, *nsutils.HTTPError) {
	return nil, nsutils.ErrNotImplemented
}

func (c CarbsCollection) Search(ctx context.Context, siteContext *httputils.SiteContext, query nsquery.Query) ([]nsutils.Record, *nsutils.HTTPError) {
	var rerr error
	var carbs []*site.Carb

	if !query.HasFilters {
		carbs, rerr = siteContext.SiteDB.Queries.AllCarbs(ctx, nsquery.CarbsSearchAll(query))
	} else {
		carbs, rerr = siteContext.SiteDB.Queries.SearchWithDefaultSortCarbs(ctx, nsquery.CarbsSearchDefaultOrder(query))
	}

	if rerr != nil {
		slog.Error("failed to query carbs from database", "error", rerr)
		return nil, nsutils.ErrInternal
	}

	return nsutils.CreateRecordSlice(carbs, c.converter), nil
}

func (c CarbsCollection) Update(ctx context.Context, siteContext *httputils.SiteContext, identifier string, patch map[string]any) (*nsutils.Record, *nsutils.HTTPError) {
	return nil, nsutils.ErrNotImplemented
}

func (c CarbsCollection) Delete(ctx context.Context, siteContext *httputils.SiteContext, identifier string, permanent bool) *nsutils.HTTPError {
	return nsutils.ErrNotImplemented
}

func (c CarbsCollection) DeleteAll(ctx context.Context, siteContext *httputils.SiteContext, permanent bool) (int, *nsutils.HTTPError) {
	return 0, nsutils.ErrNotImplemented
}
