package collections

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	"nighttroop.t1d.nz/data/sql/user"
	"nighttroop.t1d.nz/models/nightscout"
	"nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/nsutils"
	"nighttroop.t1d.nz/utils"
	"nighttroop.t1d.nz/utils/httputils"
)

type EntriesCollection struct{}

func (c EntriesCollection) Create(ctx context.Context, data map[string]any) (*nsutils.Record, *nsutils.HTTPError) {
	userContext := httputils.GetUserContext(ctx)
	entryModel := &nightscout.Entry{}

	err := utils.FromMap(data, entryModel)

	if err != nil {
		slog.Error("failed to map data to entry model", "error", err)
		return nil, nsutils.ErrInternal
	}

	createParams := nightscout.EntryToCreate(*entryModel, nightscout.CreateContext{IDPrefix: "ns", CreatedByIdentityID: userContext.IdentityModel.ID, CreatedByIdentityName: userContext.IdentityModel.Name})

	userContext.UserDB.Queries.CreateGlucoseReading(ctx, createParams)

	return &nsutils.Record{
		Collection:  "entries",
		ID:          createParams.ID,
		Data:        data,
		SrvModified: time.Now().UnixMilli(),
	}, nil
}

func (c EntriesCollection) Get(ctx context.Context, identifier string) (*nsutils.Record, *nsutils.HTTPError) {
	userContext := httputils.GetUserContext(ctx)

	glucoseReading, err := userContext.UserDB.Queries.GetGlucoseReading(ctx, identifier)
	if err == sql.ErrNoRows {
		return nil, nsutils.ErrNotFound
	}
	if err != nil {
		slog.Error("failed to get glucose reading by ID", "error", err)
		return nil, nsutils.ErrInternal
	}
	if glucoseReading == nil {
		return nil, nsutils.ErrNotFound
	}
	record := GlucoseReadingToRecord(glucoseReading)
	return &record, nil
}

func (c EntriesCollection) Search(ctx context.Context, query *nsutils.Query) ([]nsutils.Record, *nsutils.HTTPError) {
	userContext := httputils.GetUserContext(ctx)
	entries, err := userContext.UserDB.Queries.AllGlucoseReadings(ctx, int64(query.Limit))
	if err != nil {
		slog.Error("failed to query glucose readings from database", "error", err)
		return nil, nsutils.ErrInternal
	}

	return CreateRecordSlice(entries, GlucoseReadingToRecord), nil
}

func (c EntriesCollection) Update(ctx context.Context, identifier string, patch map[string]any) (*nsutils.Record, *nsutils.HTTPError) {
	userContext := httputils.GetUserContext(ctx)
	entryModel := &nightscout.Entry{}
	err := utils.FromMap(patch, entryModel)
	if err != nil {
		slog.Error("failed to map data to entry model", "error", err)
		return nil, nsutils.ErrInternal
	}

	updateParams := nightscout.EntryToUpdate(*entryModel)
	glucoseReading, err := userContext.UserDB.Queries.UpdateGlucoseReading(ctx, updateParams)

	record := GlucoseReadingToRecord(glucoseReading)
	return &record, nil
}

func (c EntriesCollection) Delete(ctx context.Context, identifier string, permanent bool) *nsutils.HTTPError {
	userContext := httputils.GetUserContext(ctx)
	if permanent {
		res, err := userContext.UserDB.Queries.DeletePermanentGlucoseReading(ctx, identifier)
		if err != nil {
			slog.Error("failed to permanently delete glucose reading by ID", "error", err)
			return nsutils.ErrInternal
		}
		rowsAffected, err := res.RowsAffected()
		if err != nil {
			slog.Error("failed to get rows affected for glucose reading delete", "error", err)
			return nsutils.ErrInternal
		}
		if rowsAffected == 0 {
			return nsutils.ErrNotFound
		}
		slog.Info("Deleted glucose reading", "eid", identifier, "rowsAffected", rowsAffected)

	} else {
		res, err := userContext.UserDB.Queries.DeleteGlucoseReading(ctx, identifier)
		if err != nil {
			slog.Error("failed to soft delete glucose reading by ID", "error", err)
			return nsutils.ErrInternal
		}
		rowsAffected, err := res.RowsAffected()
		if err != nil {
			slog.Error("failed to get rows affected for glucose reading delete", "error", err)
			return nsutils.ErrInternal
		}
		if rowsAffected == 0 {
			return nsutils.ErrNotFound
		}
		slog.Info("Soft Deleted glucose reading", "eid", identifier, "rowsAffected", rowsAffected)

	}
	return nil
}

func (c EntriesCollection) DeleteType(ctx context.Context, typeString string, permanent bool) (int, *nsutils.HTTPError) {
	userContext := httputils.GetUserContext(ctx)

	if permanent {
		res, err := userContext.UserDB.Queries.DeletePermanentGlucoseReadingsByType(ctx, typeString)
		if err != nil {
			slog.Error("failed to permanently delete glucose readings by type", "error", err)
			return 0, nsutils.ErrInternal
		}
		rowsAffected, err := res.RowsAffected()
		if err != nil {
			slog.Error("failed to get rows affected for glucose readings delete", "error", err)
			return 0, nsutils.ErrInternal
		}
		return int(rowsAffected), nil
	} else {
		res, err := userContext.UserDB.Queries.DeleteGlucoseReadingsByType(ctx, typeString)
		if err != nil {
			slog.Error("failed to soft delete glucose readings by IDs", "error", err)
			return 0, nsutils.ErrInternal
		}
		rowsAffected, err := res.RowsAffected()
		if err != nil {
			slog.Error("failed to get rows affected for glucose readings delete", "error", err)
			return 0, nsutils.ErrInternal
		}
		return int(rowsAffected), nil
	}
}

func GlucoseReadingToRecord(sgv *user.GlucoseReading) nsutils.Record {
	if sgv == nil {
		return nsutils.Record{}
	}

	entryModel := nightscout.DBModelToEntry(*sgv)

	entryMap, err := utils.ToMap(&entryModel)
	if err != nil {
		return nsutils.Record{}
	}

	return nsutils.Record{
		Collection:  "entries",
		ID:          *entryModel.Identifier,
		Data:        entryMap,
		SrvCreated:  entryModel.SrvCreated.Int64(),
		SrvModified: entryModel.SrvModified.Int64(),
		Subject:     "",
		IsValid:     true,
		DeletedAt:   nil,
	}
}
