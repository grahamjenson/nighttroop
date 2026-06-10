package collections

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
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
	userContext.UserDB.Queries.CreateGlucoseReading(ctx, user.CreateGlucoseReadingParams{
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

	glucoseReading, err := userContext.UserDB.Queries.GetGlucoseReadingByEID(ctx, identifier)
	if err == sql.ErrNoRows {
		return nil, di.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get glucose reading by ID: %w", err)
	}
	if glucoseReading == nil {
		return nil, di.ErrNotFound
	}
	record := GlucoseReadingToRecord(glucoseReading)
	return &record, nil
}

func (c EntriesCollection) Search(ctx context.Context, query di.Query) ([]model.Record, error) {
	userContext := httputils.GetUserContext(ctx)
	entries, err := userContext.UserDB.Queries.AllGlucoseReadings(ctx, int64(query.Limit))
	if err != nil {
		return nil, err
	}

	return CreateRecordSlice(entries, GlucoseReadingToRecord), nil
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
	glucoseReading, err := userContext.UserDB.Queries.UpdateGlucoseReadingByEID(ctx, user.UpdateGlucoseReadingByEIDParams{
		Eid:                   identifier,
		OccurredAt:            entryModel.Date,
		Value:                 value,
		RawJsonData:           &raw_json,
		CreatedByIdentityID:   ptr(userContext.IdentityModel.ID),
		CreatedByIdentityName: ptr(userContext.IdentityModel.Name),
	})

	record := GlucoseReadingToRecord(glucoseReading)
	return &record, nil
}

func (c EntriesCollection) Delete(ctx context.Context, collection, identifier string, permanent bool) error {
	userContext := httputils.GetUserContext(ctx)
	if permanent {
		res, err := userContext.UserDB.Queries.DeletePermanentGlucoseReadingByEID(ctx, identifier)
		if err != nil {
			return fmt.Errorf("failed to permanently delete glucose reading by ID: %w", err)
		}
		rowsAffected, err := res.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to get rows affected for glucose reading delete: %w", err)
		}
		if rowsAffected == 0 {
			return di.ErrNotFound
		}
		slog.Info("Deleted glucose reading", "eid", identifier, "rowsAffected", rowsAffected)

	} else {
		res, err := userContext.UserDB.Queries.DeleteGlucoseReadingByEID(ctx, identifier)
		if err != nil {
			return fmt.Errorf("failed to soft delete glucose reading by ID: %w", err)
		}
		rowsAffected, err := res.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to get rows affected for glucose reading delete: %w", err)
		}
		if rowsAffected == 0 {
			return di.ErrNotFound
		}
		slog.Info("Soft Deleted glucose reading", "eid", identifier, "rowsAffected", rowsAffected)

	}
	return nil
}

func (c EntriesCollection) DeleteMatching(ctx context.Context, collection string, query di.Query, permanent bool) (int, error) {
	userContext := httputils.GetUserContext(ctx)
	records, err := c.Search(ctx, query)
	if err != nil {
		return 0, err
	}
	if len(records) == 0 {
		return 0, nil
	}
	recordIDs := make([]string, len(records))
	for i, record := range records {
		recordIDs[i] = record.Identifier()
	}

	if permanent {
		res, err := userContext.UserDB.Queries.DeletePermanentGlucoseReadingsByEIDs(ctx, recordIDs)
		if err != nil {
			return 0, fmt.Errorf("failed to permanently delete glucose readings by IDs: %w", err)
		}
		rowsAffected, err := res.RowsAffected()
		if err != nil {
			return 0, fmt.Errorf("failed to get rows affected for glucose readings delete: %w", err)
		}
		return int(rowsAffected), nil
	} else {
		res, err := userContext.UserDB.Queries.DeleteGlucoseReadingsByEIDs(ctx, recordIDs)
		if err != nil {
			return 0, fmt.Errorf("failed to soft delete glucose readings by IDs: %w", err)
		}
		rowsAffected, err := res.RowsAffected()
		if err != nil {
			return 0, fmt.Errorf("failed to get rows affected for glucose readings delete: %w", err)
		}
		return int(rowsAffected), nil
	}
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

func GlucoseReadingToRecord(sgv *user.GlucoseReading) model.Record {
	if sgv == nil {
		return model.Record{}
	}

	entryModel := Entry{
		Identifier: ptr(sgv.Eid),
		Type:       ptr(sgv.Type),
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
