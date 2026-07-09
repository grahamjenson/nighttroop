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

type TempTargetTypeCollection struct {
	converter func(c *site.TempTarget) nsutils.Record
}

func (c TempTargetTypeCollection) converterToRecord(tempTarget *site.TempTarget) nsutils.Record {
	if c.converter == nil {
		c.converter = nsutils.DBModelToRecord(nsutils.Treatments, nightscout.TempTargetDBModelToTreatment)
	}
	return c.converter(tempTarget)
}

func (c TempTargetTypeCollection) GetPrefix() string {
	return "tt"
}

func (c TempTargetTypeCollection) Create(ctx context.Context, siteContext *httputils.SiteContext, treatment nightscout.Treatment) (*nsutils.Record, *nsutils.HTTPError) {
	dbModel, err := siteContext.SiteDB.Queries.CreateTempTarget(ctx,
		nightscout.TreatmentToCreateTempTarget(treatment, nightscout.CreateContext{
			IDPrefix:              c.GetPrefix(),
			CreatedByIdentityID:   siteContext.IdentityModel.ID,
			CreatedByIdentityName: siteContext.IdentityModel.Name,
		},
		),
	)
	if err != nil {
		slog.Error("failed to create temp target in database", "error", err)
		return nil, nsutils.ErrInternal
	}
	return ptr(c.converterToRecord(dbModel)), nil
}

func (c TempTargetTypeCollection) Get(ctx context.Context, siteContext *httputils.SiteContext, identifier string) (*nsutils.Record, *nsutils.HTTPError) {
	return nil, nsutils.ErrNotImplemented
}

func (c TempTargetTypeCollection) Search(ctx context.Context, siteContext *httputils.SiteContext, query nsquery.Query) ([]nsutils.Record, *nsutils.HTTPError) {
	return nil, nsutils.ErrNotImplemented
}

func (c TempTargetTypeCollection) Update(ctx context.Context, siteContext *httputils.SiteContext, identifier string, patch map[string]any) (*nsutils.Record, *nsutils.HTTPError) {
	return nil, nsutils.ErrNotImplemented
}

func (c TempTargetTypeCollection) Delete(ctx context.Context, siteContext *httputils.SiteContext, identifier string, permanent bool) *nsutils.HTTPError {
	return nsutils.ErrNotImplemented
}

func (c TempTargetTypeCollection) DeleteAll(ctx context.Context, siteContext *httputils.SiteContext, permanent bool) (int, *nsutils.HTTPError) {
	return 0, nsutils.ErrNotImplemented
}
