package collections

import (
	"context"
	"fmt"
	"time"

	"nighttroop.t1d.nz/data/sql/user"
	"nighttroop.t1d.nz/nightscout"
	"nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/nsutils"
	"nighttroop.t1d.nz/utils"
	"nighttroop.t1d.nz/utils/httputils"
)

type TreatmentCollection struct{}

func (c TreatmentCollection) Create(ctx context.Context, data map[string]any) (*nsutils.Record, error) {
	userContext := httputils.GetUserContext(ctx)

	treatmentModel := &nightscout.Treatment{}
	err := treatmentModel.FromMap(data)
	if err != nil {
		return nil, fmt.Errorf("failed to map data to treatment model: %w", err)
	}
	externalID := utils.NextEID("ns")
	created := false
	switch *treatmentModel.EventType {
	case nightscout.CarbCorrection:
		err = userContext.UserDB.Queries.CreateCarb(ctx, user.CreateCarbParams{
			Eid:                   externalID,
			Amount:                treatmentModel.Carbs.Float64(),
			RawJsonData:           &treatmentModel.RawJSON,
			CreatedByIdentityName: userContext.IdentityModel.Name,
			CreatedByIdentityID:   userContext.IdentityModel.ID,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create carb in database: %w", err)
		}
		created = true
	}
	if !created {
		return nil, fmt.Errorf("unsupported treatment event type: %s", *treatmentModel.EventType)
	}
	return &nsutils.Record{
		Collection:  "treatments",
		ID:          externalID,
		Data:        data,
		SrvModified: time.Now().UnixMilli(),
	}, nil
}

func (c TreatmentCollection) Get(ctx context.Context, identifier string) (*nsutils.Record, error) {
	return nil, nil
}

func (c TreatmentCollection) Search(ctx context.Context, query *nsutils.Query) ([]nsutils.Record, error) {
	userContext := httputils.GetUserContext(ctx)
	records := []nsutils.Record{}
	carbs, err := userContext.UserDB.Queries.AllCarbs(ctx)
	if err != nil {
		return nil, err
	}
	records = append(records, CreateRecordSlice(carbs, CarbsToRecord)...)

	return records, nil
}

func (c TreatmentCollection) Update(ctx context.Context, identifier string, patch map[string]any) (*nsutils.Record, error) {
	return nil, nil
}

func (c TreatmentCollection) Delete(ctx context.Context, identifier string, permanent bool) error {
	return nil
}

func (c TreatmentCollection) DeleteMatching(ctx context.Context, query *nsutils.Query, permanent bool) (int, error) {
	return 0, nil
}

func CarbsToRecord(carbs *user.Carb) nsutils.Record {
	if carbs == nil {
		return nsutils.Record{}
	}

	treatmentModel := &nightscout.Treatment{}
	err := treatmentModel.FromCarbDBModel(carbs)
	if err != nil {
		return nsutils.Record{}
	}

	treatmentMap, err := treatmentModel.ToMap()
	if err != nil {
		return nsutils.Record{}
	}

	return nsutils.Record{
		Collection:  "treatments",
		ID:          *treatmentModel.Identifier,
		Data:        treatmentMap,
		SrvCreated:  treatmentModel.Date.Int64(),
		SrvModified: treatmentModel.Date.Int64(),
		Subject:     "",
		IsValid:     true,
		DeletedAt:   nil,
	}
}
