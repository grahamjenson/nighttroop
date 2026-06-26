package collections

import (
	"context"
	"log/slog"

	"nighttroop.t1d.nz/data/sql/user"
	"nighttroop.t1d.nz/models/nightscout"
	"nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/nsutils"
	"nighttroop.t1d.nz/utils"
	"nighttroop.t1d.nz/utils/httputils"
)

type TreatmentCollection struct {
	carbConverter func(c *user.Carb) nsutils.Record
}

func NewTreatmentCollection() TreatmentCollection {
	return TreatmentCollection{
		carbConverter: DBModelToRecord(nsutils.Treatments, nightscout.CarbsDBModelToTreatment),
	}
}

func (c TreatmentCollection) Create(ctx context.Context, userContext *httputils.UserContext, data map[string]any) (*nsutils.Record, *nsutils.HTTPError) {

	treatmentModel := nightscout.Treatment{}
	err := utils.FromMap(data, &treatmentModel)
	if err != nil {
		slog.Error("failed to map data to treatment model", "error", err)
		return nil, nsutils.ErrInternal
	}

	switch *treatmentModel.EventType {
	case nightscout.CarbCorrection:
		dbModel, err := userContext.UserDB.Queries.CreateCarb(ctx, nightscout.TreatmentToCreateCarb(treatmentModel, nightscout.CreateContext{IDPrefix: "ns", CreatedByIdentityID: userContext.IdentityModel.ID, CreatedByIdentityName: userContext.IdentityModel.Name}))
		if err != nil {
			slog.Error("failed to create carb in database", "error", err)
			return nil, nsutils.ErrInternal
		}
		return ptr(c.carbConverter(dbModel)), nil
	}

	slog.Error("unsupported treatment event type", "eventType", *treatmentModel.EventType)
	return nil, nsutils.ErrInternal

}

func (c TreatmentCollection) Get(ctx context.Context, userContext *httputils.UserContext, identifier string) (*nsutils.Record, *nsutils.HTTPError) {
	return nil, nsutils.ErrNotImplemented
}

func (c TreatmentCollection) Search(ctx context.Context, userContext *httputils.UserContext, query *nsutils.Query) ([]nsutils.Record, *nsutils.HTTPError) {
	records := []nsutils.Record{}
	carbs, err := userContext.UserDB.Queries.AllCarbs(ctx)
	if err != nil {
		slog.Error("failed to query carbs from database", "error", err)
		return nil, nsutils.ErrInternal
	}
	records = append(records, CreateRecordSlice(carbs, c.carbConverter)...)

	return records, nil
}

func (c TreatmentCollection) Update(ctx context.Context, userContext *httputils.UserContext, identifier string, patch map[string]any) (*nsutils.Record, *nsutils.HTTPError) {
	return nil, nsutils.ErrNotImplemented
}

func (c TreatmentCollection) Delete(ctx context.Context, userContext *httputils.UserContext, identifier string, permanent bool) *nsutils.HTTPError {
	return nsutils.ErrNotImplemented
}

func (c TreatmentCollection) DeleteType(ctx context.Context, userContext *httputils.UserContext, typeString string, permanent bool) (int, *nsutils.HTTPError) {
	return 0, nsutils.ErrNotImplemented
}
