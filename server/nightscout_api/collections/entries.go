package collections

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"nighttroop.t1d.nz/data/sql/user"
	"nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/di"
	"nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/model"
	"nighttroop.t1d.nz/utils"
	"nighttroop.t1d.nz/utils/httputils"
)

type EntriesCollection struct{}

func (c EntriesCollection) Create(ctx context.Context, data map[string]any) (*model.Record, bool, error) {
	userContext := httputils.GetUserContext(ctx)
	var entryModel Entry
	raw_json, err := MapToStruct(data, &entryModel)
	if err != nil {
		return nil, false, fmt.Errorf("failed to map data to entry model: %w", err)
	}
	occurredAt := time.Now().UnixMilli()
	if entryModel.Date != nil {
		occurredAt = *entryModel.Date
	}
	if entryModel.SGV == nil {
		return nil, false, fmt.Errorf("SGV value is required for entries collection")
	}
	eid := utils.NextEID("ns")
	userContext.UserDB.Queries.CreateSGV(ctx, user.CreateSGVParams{
		Eid:                   eid,
		OccurredAt:            occurredAt,
		Value:                 float64(*entryModel.SGV),
		RawJsonData:           &raw_json,
		CreatedByIdentityID:   userContext.IdentityModel.ID,
		CreatedByIdentityName: userContext.IdentityModel.Name,
	})

	return &model.Record{
		Collection:  "entries",
		ID:          eid,
		Data:        data,
		SrvModified: time.Now().UnixMilli(),
	}, true, nil
}

func (c EntriesCollection) Get(ctx context.Context, identifier string) (*model.Record, error) {
	userContext := httputils.GetUserContext(ctx)

	sgv, err := userContext.UserDB.Queries.GetSGVByEID(ctx, identifier)
	if err == sql.ErrNoRows {
		return nil, di.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get SGV by ID: %w", err)
	}
	if sgv == nil {
		return nil, di.ErrNotFound
	}
	record := SGVToRecord(sgv)
	return &record, nil
}

func (c EntriesCollection) Search(ctx context.Context, query di.Query) ([]model.Record, error) {
	userContext := httputils.GetUserContext(ctx)
	entries, err := userContext.UserDB.Queries.AllSGVs(ctx)
	if err != nil {
		return nil, err
	}

	return CreateRecordSlice(entries, SGVToRecord), nil
}

func (c EntriesCollection) Update(ctx context.Context, collection, identifier string, patch map[string]any) (*model.Record, error) {
	userContext := httputils.GetUserContext(ctx)
	var entryModel Entry
	raw_json, err := MapToStruct(patch, &entryModel)
	if err != nil {
		return nil, fmt.Errorf("failed to map data to entry model: %w", err)
	}

	var value *float64
	if entryModel.SGV != nil {
		value = (*float64)(entryModel.SGV)
	}
	sgv, err := userContext.UserDB.Queries.UpdateSGVByEID(ctx, user.UpdateSGVByEIDParams{
		Eid:                   identifier,
		OccurredAt:            entryModel.Date,
		Value:                 value,
		RawJsonData:           &raw_json,
		CreatedByIdentityID:   ptr(userContext.IdentityModel.ID),
		CreatedByIdentityName: ptr(userContext.IdentityModel.Name),
	})

	record := SGVToRecord(sgv)
	return &record, nil
}

func (c EntriesCollection) Delete(ctx context.Context, collection, identifier string, permanent bool) error {
	return nil
}

func (c EntriesCollection) DeleteMatching(ctx context.Context, collection string, query di.Query, permanent bool) (int, error) {
	return 0, nil
}

////
// Models
////

type Entry struct {
	Identifier *string          `json:"identifier,omitempty"`
	Type       *string          `json:"type,omitempty"`
	SGV        *FlexibleFloat64 `json:"sgv,omitempty"`
	MBG        *FlexibleFloat64 `json:"mbg,omitempty"`
	Date       *int64           `json:"date,omitempty"`
	DateString *string          `json:"dateString,omitempty"`
	Device     *string          `json:"device,omitempty"`
	Direction  *string          `json:"direction,omitempty"`
	Payload    map[string]any   `json:"payload,omitempty"`
}

////
// Mappers
////

func SGVToRecord(sgv *user.Sgv) model.Record {
	if sgv == nil {
		return model.Record{}
	}

	entryModel := Entry{
		Identifier: ptr(sgv.Eid),
		Type:       ptr("sgv"),
		SGV:        ptr(FlexibleFloat64(sgv.Value)),
		Date:       ptr(sgv.OccurredAt),
		DateString: ptr(dateString(sgv.OccurredAt)),
		Device:     ptr(fmt.Sprintf("%d", sgv.CreatedByIdentityID)),
	}

	return model.Record{
		Collection:  "entries",
		ID:          sgv.Eid,
		Data:        StructToMap(entryModel),
		SrvCreated:  sgv.OccurredAt,
		SrvModified: sgv.OccurredAt,
		Subject:     "",
		IsValid:     true,
		DeletedAt:   nil,
	}
}
