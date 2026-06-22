package collections

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"nighttroop.t1d.nz/data/sql/user"
	"nighttroop.t1d.nz/nightscout"
	"nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/nserrors"
	"nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/nsutils"
	"nighttroop.t1d.nz/utils"
	"nighttroop.t1d.nz/utils/httputils"
)

type EntriesCollection struct{}

func (c EntriesCollection) Create(ctx context.Context, data map[string]any) (*nsutils.Record, error) {
	userContext := httputils.GetUserContext(ctx)
	entryModel := &nightscout.Entry{}
	err := entryModel.FromMap(data)
	if err != nil {
		return nil, fmt.Errorf("failed to map data to entry model: %w", err)
	}
	occurredAt := time.Now().UnixMilli()
	if entryModel.Date != nil {
		occurredAt = entryModel.Date.Int64()
	}
	if entryModel.SGV == nil {
		return nil, fmt.Errorf("SGV value is required for entries collection")
	}
	offset := int64(0)
	if entryModel.UTCOffset != nil {
		offset = int64(*entryModel.UTCOffset)
	}
	eid := utils.NextEID("ns")
	entryModel.Identifier = &eid
	entryModel.UTCOffset = ptr(nightscout.FlexibleInt64(offset))
	entryModel.Date = ptr(nightscout.DateInt(occurredAt))

	dbModel, err := entryModel.ToDBModel()
	if err != nil {
		return nil, fmt.Errorf("failed to convert entry model to DB model: %w", err)
	}
	dbModel.CreatedByIdentityID = userContext.IdentityModel.ID
	dbModel.CreatedByIdentityName = userContext.IdentityModel.Name

	userContext.UserDB.Queries.CreateGlucoseReading(ctx, dbModel.ToDBCreateParams())

	return &nsutils.Record{
		Collection:  "entries",
		ID:          eid,
		Data:        data,
		SrvModified: time.Now().UnixMilli(),
	}, nil
}

func (c EntriesCollection) Get(ctx context.Context, identifier string) (*nsutils.Record, error) {
	userContext := httputils.GetUserContext(ctx)

	glucoseReading, err := userContext.UserDB.Queries.GetGlucoseReadingByEID(ctx, identifier)
	if err == sql.ErrNoRows {
		return nil, nserrors.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get glucose reading by ID: %w", err)
	}
	if glucoseReading == nil {
		return nil, nserrors.ErrNotFound
	}
	record := GlucoseReadingToRecord(glucoseReading)
	return &record, nil
}

func (c EntriesCollection) Search(ctx context.Context, query *nsutils.Query) ([]nsutils.Record, error) {
	userContext := httputils.GetUserContext(ctx)
	entries, err := userContext.UserDB.Queries.AllGlucoseReadings(ctx, int64(query.Limit))
	if err != nil {
		return nil, err
	}

	return CreateRecordSlice(entries, GlucoseReadingToRecord), nil
}

func (c EntriesCollection) Update(ctx context.Context, identifier string, patch map[string]any) (*nsutils.Record, error) {
	userContext := httputils.GetUserContext(ctx)
	entryModel := &nightscout.Entry{}
	err := entryModel.FromMap(patch)
	if err != nil {
		return nil, fmt.Errorf("failed to map data to entry model: %w", err)
	}

	var value *float64
	if entryModel.SGV != nil {
		value = (*float64)(entryModel.SGV)
	}
	glucoseReading, err := userContext.UserDB.Queries.UpdateGlucoseReadingByEID(ctx, user.UpdateGlucoseReadingByEIDParams{
		Eid:                   identifier,
		OccurredAt:            entryModel.Date.Int64p(),
		Value:                 value,
		RawJsonData:           &entryModel.RawJSON,
		CreatedByIdentityID:   &userContext.IdentityModel.ID,
		CreatedByIdentityName: &userContext.IdentityModel.Name,
	})

	record := GlucoseReadingToRecord(glucoseReading)
	return &record, nil
}

func (c EntriesCollection) Delete(ctx context.Context, identifier string, permanent bool) error {
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
			return nserrors.ErrNotFound
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
			return nserrors.ErrNotFound
		}
		slog.Info("Soft Deleted glucose reading", "eid", identifier, "rowsAffected", rowsAffected)

	}
	return nil
}

func (c EntriesCollection) DeleteMatching(ctx context.Context, query *nsutils.Query, permanent bool) (int, error) {
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

func GlucoseReadingToRecord(sgv *user.GlucoseReading) nsutils.Record {
	if sgv == nil {
		return nsutils.Record{}
	}

	entryModel := &nightscout.Entry{}
	err := entryModel.FromDBModel(sgv)
	if err != nil {
		return nsutils.Record{}
	}

	entryMap, err := entryModel.ToMap()
	if err != nil {
		return nsutils.Record{}
	}

	return nsutils.Record{
		Collection:  "entries",
		ID:          *entryModel.Identifier,
		Data:        entryMap,
		SrvCreated:  entryModel.Date.Int64(),
		SrvModified: entryModel.Date.Int64(),
		Subject:     "",
		IsValid:     true,
		DeletedAt:   nil,
	}
}
