package treatments

import (
	"context"
	"log/slog"

	"nighttroop.t1d.nz/models/nightscout"
	"nighttroop.t1d.nz/server/nightscout_api/nsutils"
	"nighttroop.t1d.nz/utils/httputils"
)

type CarbsCorrection struct{}

func (c CarbsCorrection) Create(ctx context.Context, siteContext *httputils.SiteContext, treatment nightscout.Treatment) (*nsutils.Record, *nsutils.HTTPError) {
	dbModel, err := siteContext.SiteDB.Queries.CreateFoodIntake(ctx,
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

func (c CarbsCorrection) Update(ctx context.Context, siteContext *httputils.SiteContext, identifier string, patch map[string]any) (*nsutils.Record, *nsutils.HTTPError) {
	return nil, nsutils.ErrNotImplemented
}
