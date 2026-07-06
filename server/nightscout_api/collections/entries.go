package collections

import (
	"context"
	"database/sql"
	"log/slog"

	user "nighttroop.t1d.nz/data/sql/site"
	"nighttroop.t1d.nz/models/nightscout"
	"nighttroop.t1d.nz/server/nightscout_api/nsquery"
	"nighttroop.t1d.nz/server/nightscout_api/nsutils"
	"nighttroop.t1d.nz/utils"
	"nighttroop.t1d.nz/utils/httputils"
)

type EntriesCollection struct {
	converter func(c *user.GlucoseReading) nsutils.Record
	creator   func(ctx context.Context, siteContext *httputils.SiteContext, data map[string]any) (*nsutils.Record, *nsutils.HTTPError)
}

func NewEntriesCollection() EntriesCollection {
	return EntriesCollection{
		converter: DBModelToRecord(nsutils.Entries, nightscout.DBModelToEntry),
	}
}

func (c EntriesCollection) Create(ctx context.Context, siteContext *httputils.SiteContext, data map[string]any) (*nsutils.Record, *nsutils.HTTPError) {
	entryModel := &nightscout.Entry{}
	err := utils.FromMap(data, entryModel)

	if err != nil {
		slog.Error("failed to map data to entry model", "error", err)
		return nil, nsutils.BadDataBodyError
	}

	createParams := nightscout.EntryToCreate(
		*entryModel,
		nightscout.CreateContext{IDPrefix: "ns", CreatedByIdentityID: siteContext.IdentityModel.ID, CreatedByIdentityName: siteContext.IdentityModel.Name},
	)

	dbModel, err := siteContext.SiteDB.Queries.CreateGlucoseReading(ctx, createParams)
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

func (c EntriesCollection) Get(ctx context.Context, siteContext *httputils.SiteContext, identifier string) (*nsutils.Record, *nsutils.HTTPError) {
	glucoseReading, err := siteContext.SiteDB.Queries.GetGlucoseReading(ctx, user.GetGlucoseReadingParams{ID: identifier})
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

func (c EntriesCollection) Search(ctx context.Context, siteContext *httputils.SiteContext, query nsquery.Query) ([]nsutils.Record, *nsutils.HTTPError) {
	// Declare the support orders
	so := query.SortOrder.Support(nsquery.SortOccurredAtDesc, nsquery.SortOccurredAtAsc, nsquery.SortCreatedAtDesc, nsquery.SortCreatedAtAsc, nsquery.SortUpdatedAtDesc, nsquery.SortUpdatedAtAsc)
	// check against default sort order
	defaultSortOrder := so == nsquery.SortOccurredAtDesc

	var entries []*user.GlucoseReading
	var rerr error

	// There are three options,
	if defaultSortOrder && !query.HasFilters {
		entries, rerr = siteContext.SiteDB.Queries.AllGlucoseReadings(ctx, nsquery.GlucoseSearchAll(query))
	} else if defaultSortOrder && query.HasFilters {
		entries, rerr = siteContext.SiteDB.Queries.SearchWithDefaultSortGlucoseReadings(ctx, nsquery.GlucoseSearchDefaultOrder(query))
	} else {
		entries, rerr = siteContext.SiteDB.Queries.SearchWithNonDefaultSortGlucoseReadings(ctx, nsquery.GlucoseSearchNonDefaultOrder(query))
	}

	if rerr != nil {
		slog.Error("failed to query glucose readings from database", "error", rerr)
		return nil, nsutils.ErrInternal
	}

	return CreateRecordSlice(entries, c.converter), nil
}

func (c EntriesCollection) Update(ctx context.Context, siteContext *httputils.SiteContext, identifier string, patch map[string]any) (*nsutils.Record, *nsutils.HTTPError) {

	entryModel := &nightscout.Entry{}
	err := utils.FromMap(patch, entryModel)
	if err != nil {
		slog.Error("failed to map data to entry model", "error", err)
		return nil, nsutils.ErrInternal
	}

	updateParams := nightscout.EntryToUpdate(
		*entryModel,
		nightscout.UpdateContext{UpdatedByIdentityID: siteContext.IdentityModel.ID,
			UpdatedByIdentityName: siteContext.IdentityModel.Name,
		},
	)
	glucoseReading, err := siteContext.SiteDB.Queries.UpdateGlucoseReading(ctx, updateParams)
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

func (c EntriesCollection) Delete(ctx context.Context, siteContext *httputils.SiteContext, identifier string, permanent bool) *nsutils.HTTPError {

	if permanent {
		res, err := siteContext.SiteDB.Queries.DeletePermanentGlucoseReading(ctx, user.DeletePermanentGlucoseReadingParams{ID: identifier})
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
		res, err := siteContext.SiteDB.Queries.DeleteGlucoseReading(ctx, user.DeleteGlucoseReadingParams{ID: identifier})
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

func (c EntriesCollection) DeleteType(ctx context.Context, siteContext *httputils.SiteContext, typeString string, permanent bool) (int, *nsutils.HTTPError) {

	if permanent {
		res, err := siteContext.SiteDB.Queries.DeletePermanentGlucoseReadingsByType(ctx, user.DeletePermanentGlucoseReadingsByTypeParams{Type: typeString})
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
		res, err := siteContext.SiteDB.Queries.DeleteGlucoseReadingsByType(ctx, user.DeleteGlucoseReadingsByTypeParams{Type: typeString})
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
