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

type TempBasalTypeCollection struct {
	converter func(c *site.TempBasal) nsutils.Record
}

func (c TempBasalTypeCollection) converterToRecord(tempBasal *site.TempBasal) nsutils.Record {
	if c.converter == nil {
		c.converter = nsutils.DBModelToRecord(nsutils.Treatments, nightscout.TempBasalDBModelToTreatment)
	}
	return c.converter(tempBasal)
}

func (c TempBasalTypeCollection) GetPrefix() string {
	return "tb"
}

func (c TempBasalTypeCollection) Create(ctx context.Context, siteContext *httputils.SiteContext, treatment nightscout.Treatment) (*nsutils.Record, *nsutils.HTTPError) {
	dbModel, err := siteContext.SiteDB.Queries.CreateTempBasal(ctx,
		nightscout.TreatmentToCreateTempBasal(treatment, nightscout.CreateContext{
			IDPrefix:              c.GetPrefix(),
			CreatedByIdentityID:   siteContext.IdentityModel.ID,
			CreatedByIdentityName: siteContext.IdentityModel.Name,
		},
		),
	)
	if err != nil {
		slog.Error("failed to create temp basal in database", "error", err)
		return nil, nsutils.ErrInternal
	}
	return ptr(c.converterToRecord(dbModel)), nil
}

func (c TempBasalTypeCollection) Get(ctx context.Context, siteContext *httputils.SiteContext, identifier string) (*nsutils.Record, *nsutils.HTTPError) {
	return nil, nsutils.ErrNotImplemented
}

func (c TempBasalTypeCollection) Search(ctx context.Context, siteContext *httputils.SiteContext, query nsquery.Query) ([]nsutils.Record, *nsutils.HTTPError) {
	return nil, nsutils.ErrNotImplemented
}

func (c TempBasalTypeCollection) Update(ctx context.Context, siteContext *httputils.SiteContext, identifier string, patch map[string]any) (*nsutils.Record, *nsutils.HTTPError) {
	return nil, nsutils.ErrNotImplemented
}

func (c TempBasalTypeCollection) Delete(ctx context.Context, siteContext *httputils.SiteContext, identifier string, permanent bool) *nsutils.HTTPError {
	return nsutils.ErrNotImplemented
}

func (c TempBasalTypeCollection) DeleteAll(ctx context.Context, siteContext *httputils.SiteContext, permanent bool) (int, *nsutils.HTTPError) {
	return 0, nsutils.ErrNotImplemented
}
