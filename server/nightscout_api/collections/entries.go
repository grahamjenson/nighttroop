package collections

import (
	"context"
	"database/sql"
	"log/slog"

	"nighttroop.t1d.nz/data/sql/user"
	"nighttroop.t1d.nz/models/nightscout"
	"nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/nsutils"
	"nighttroop.t1d.nz/utils"
	"nighttroop.t1d.nz/utils/httputils"
)

type EntriesCollection struct {
	converter func(c *user.GlucoseReading) nsutils.Record
	creator   func(ctx context.Context, userContext *httputils.UserContext, data map[string]any) (*nsutils.Record, *nsutils.HTTPError)
}

func NewEntriesCollection() EntriesCollection {
	return EntriesCollection{
		converter: DBModelToRecord(nsutils.Entries, nightscout.DBModelToEntry),
	}
}

func (c EntriesCollection) Create(ctx context.Context, userContext *httputils.UserContext, data map[string]any) (*nsutils.Record, *nsutils.HTTPError) {
	entryModel := &nightscout.Entry{}
	err := utils.FromMap(data, entryModel)

	if err != nil {
		slog.Error("failed to map data to entry model", "error", err)
		return nil, nsutils.BadDataBodyError
	}

	createParams := nightscout.EntryToCreate(
		*entryModel,
		nightscout.CreateContext{IDPrefix: "ns", CreatedByIdentityID: userContext.IdentityModel.ID, CreatedByIdentityName: userContext.IdentityModel.Name},
	)

	dbModel, err := userContext.UserDB.Queries.CreateGlucoseReading(ctx, createParams)
	if err != nil {
		slog.Error("failed to create glucose reading in database", "error", err)
		return nil, nsutils.ErrInternal
	}
	if dbModel == nil {
		slog.Error("failed to create glucose reading in database, dbModel is nil")
		return nil, nsutils.ErrInternal
	}
	return ptr(c.converter(dbModel)), nil
}

func (c EntriesCollection) Get(ctx context.Context, userContext *httputils.UserContext, identifier string) (*nsutils.Record, *nsutils.HTTPError) {
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
	record := c.converter(glucoseReading)
	return &record, nil
}

func (c EntriesCollection) Search(ctx context.Context, userContext *httputils.UserContext, query *nsutils.Query) ([]nsutils.Record, *nsutils.HTTPError) {
	entries, err := userContext.UserDB.Queries.AllGlucoseReadings(ctx, int64(query.Limit.Int64()))
	if err != nil {
		slog.Error("failed to query glucose readings from database", "error", err)
		return nil, nsutils.ErrInternal
	}

	return CreateRecordSlice(entries, c.converter), nil
}

func (c EntriesCollection) Update(ctx context.Context, userContext *httputils.UserContext, identifier string, patch map[string]any) (*nsutils.Record, *nsutils.HTTPError) {

	entryModel := &nightscout.Entry{}
	err := utils.FromMap(patch, entryModel)
	if err != nil {
		slog.Error("failed to map data to entry model", "error", err)
		return nil, nsutils.ErrInternal
	}

	updateParams := nightscout.EntryToUpdate(*entryModel)
	glucoseReading, err := userContext.UserDB.Queries.UpdateGlucoseReading(ctx, updateParams)
	if err == sql.ErrNoRows {
		return nil, nsutils.ErrNotFound
	}
	if err != nil {
		slog.Error("failed to update glucose reading in database", "error", err)
		return nil, nsutils.ErrInternal
	}
	if glucoseReading == nil {
		return nil, nsutils.ErrNotFound
	}
	record := c.converter(glucoseReading)
	return &record, nil
}

func (c EntriesCollection) Delete(ctx context.Context, userContext *httputils.UserContext, identifier string, permanent bool) *nsutils.HTTPError {

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

func (c EntriesCollection) DeleteType(ctx context.Context, userContext *httputils.UserContext, typeString string, permanent bool) (int, *nsutils.HTTPError) {

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
