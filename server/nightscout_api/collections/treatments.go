package collections

import (
	"context"
	"log/slog"
	"time"

	"nighttroop.t1d.nz/data/sql/user"
	"nighttroop.t1d.nz/models/nightscout"
	"nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/nsutils"
	"nighttroop.t1d.nz/utils"
	"nighttroop.t1d.nz/utils/httputils"
)

type TreatmentCollection struct{}

func (c TreatmentCollection) Create(ctx context.Context, data map[string]any) (*nsutils.Record, *nsutils.HTTPError) {
	userContext := httputils.GetUserContext(ctx)

	treatmentModel := nightscout.Treatment{}
	err := utils.FromMap(data, &treatmentModel)
	if err != nil {
		slog.Error("failed to map data to treatment model", "error", err)
		return nil, nsutils.ErrInternal
	}

	created := false
	switch *treatmentModel.EventType {
	case nightscout.CarbCorrection:
		err = userContext.UserDB.Queries.CreateCarb(ctx, nightscout.TreatmentToCreateCarb(treatmentModel, nightscout.CreateContext{IDPrefix: "ns", CreatedByIdentityID: userContext.IdentityModel.ID, CreatedByIdentityName: userContext.IdentityModel.Name}))
		if err != nil {
			slog.Error("failed to create carb in database", "error", err)
			return nil, nsutils.ErrInternal
		}
		created = true
	}
	if !created {
		slog.Error("unsupported treatment event type", "eventType", *treatmentModel.EventType)
		return nil, nsutils.ErrInternal
	}
	return &nsutils.Record{
		Collection:  "treatments",
		ID:          treatmentModel.Identifier.String(),
		Data:        data,
		SrvModified: time.Now().UnixMilli(),
	}, nil
}

func (c TreatmentCollection) Get(ctx context.Context, identifier string) (*nsutils.Record, *nsutils.HTTPError) {
	return nil, nil
}

func (c TreatmentCollection) Search(ctx context.Context, query *nsutils.Query) ([]nsutils.Record, *nsutils.HTTPError) {
	userContext := httputils.GetUserContext(ctx)
	records := []nsutils.Record{}
	carbs, err := userContext.UserDB.Queries.AllCarbs(ctx)
	if err != nil {
		slog.Error("failed to query carbs from database", "error", err)
		return nil, nsutils.ErrInternal
	}
	records = append(records, CreateRecordSlice(carbs, CarbsToRecord)...)

	return records, nil
}

func (c TreatmentCollection) Update(ctx context.Context, identifier string, patch map[string]any) (*nsutils.Record, *nsutils.HTTPError) {
	return nil, nil
}

func (c TreatmentCollection) Delete(ctx context.Context, identifier string, permanent bool) *nsutils.HTTPError {
	return nil
}

func (c TreatmentCollection) DeleteType(ctx context.Context, typeString string, permanent bool) (int, *nsutils.HTTPError) {
	return 0, nil
}

func CarbsToRecord(carbs *user.Carb) nsutils.Record {
	if carbs == nil {
		return nsutils.Record{}
	}

	treatmentModel := nightscout.CarbsDBModelToTreatment(*carbs)

	treatmentMap, err := utils.ToMap(&treatmentModel)
	if err != nil {
		return nsutils.Record{}
	}

	return nsutils.Record{
		Collection:  "treatments",
		ID:          treatmentModel.Identifier.String(),
		Data:        treatmentMap,
		SrvCreated:  treatmentModel.Date.Int64(),
		SrvModified: treatmentModel.Date.Int64(),
		Subject:     "",
		IsValid:     true,
		DeletedAt:   nil,
	}
}
