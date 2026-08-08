package collections

import (
	"context"
	"database/sql"
	"log/slog"

	"nighttroop.t1d.nz/data/sql/site"
	"nighttroop.t1d.nz/models/nightscout"
	"nighttroop.t1d.nz/server/nightscout_api/nsquery"
	"nighttroop.t1d.nz/server/nightscout_api/nsutils"
	"nighttroop.t1d.nz/utils"
	"nighttroop.t1d.nz/utils/httputils"
)

type EntriesCollection struct {
	converter func(c *site.GlucoseReading) nsutils.Record
}

func NewEntriesCollection() EntriesCollection {
	return EntriesCollection{
		converter: nsutils.DBModelToRecord(nsutils.Entries, nightscout.DBModelToEntry),
	}
}

// ///
// FUNCTIONS USED BY META COLLECTIONS
// ///

func (c EntriesCollection) Name() string {
	return "entries"
}

func (c EntriesCollection) Convert(dr *site.GlucoseReading) nsutils.Record {
	return c.converter(dr)
}

func (c EntriesCollection) CreateDBFunc(siteContext *httputils.SiteContext, ctx context.Context, nsModel nightscout.Entry, cc nightscout.CreateContext) (*site.GlucoseReading, error) {
	return siteContext.SiteDB.Queries.CreateGlucoseReading(ctx, nightscout.EntryToCreate(nsModel, cc))
}

func (c EntriesCollection) UpdateSBFunc(siteContext *httputils.SiteContext, ctx context.Context, updateParams site.UpdateGlucoseReadingParams, nsModel nightscout.Entry, uc nightscout.UpdateContext) (*site.GlucoseReading, error) {
	return siteContext.SiteDB.Queries.UpdateGlucoseReading(ctx, updateParams)
}

func (c EntriesCollection) GetDBFunc(siteContext *httputils.SiteContext, ctx context.Context, identifier string) (*site.GlucoseReading, error) {
	return siteContext.SiteDB.Queries.GetGlucoseReading(ctx, site.GetGlucoseReadingParams{ID: identifier})
}

/////
// END OF FUNCTIONS USED BY META COLLECTIONS
/////

func (c EntriesCollection) Create(ctx context.Context, siteContext *httputils.SiteContext, data map[string]any) (*nsutils.Record, *nsutils.HTTPError) {
	return MetaCreate(c, ctx, siteContext, data)
}

func (c EntriesCollection) Get(ctx context.Context, siteContext *httputils.SiteContext, identifier string) (*nsutils.Record, *nsutils.HTTPError) {
	return MetaGet(c, ctx, siteContext, identifier)
}

func (c EntriesCollection) Search(ctx context.Context, siteContext *httputils.SiteContext, query nsquery.Query) ([]nsutils.Record, *nsutils.HTTPError) {
	var entries []*site.GlucoseReading
	var rerr error

	// There are three options,
	if !query.HasFilters {
		entries, rerr = siteContext.SiteDB.Queries.AllGlucoseReadings(ctx, nsquery.GlucoseAll(query))
	} else {
		entries, rerr = siteContext.SiteDB.Queries.SearchGlucoseReadings(ctx, nsquery.GlucoseSearch(query))
	}

	if rerr != nil {
		slog.Error("failed to query glucose readings from database", "error", rerr)
		return nil, nsutils.ErrInternal
	}

	return nsutils.CreateRecordSlice(entries, c.converter), nil
}

func (c EntriesCollection) Delete(ctx context.Context, siteContext *httputils.SiteContext, identifier string, permanent bool) *nsutils.HTTPError {

	if permanent {
		res, err := siteContext.SiteDB.Queries.HardDeleteGlucoseReading(ctx, site.HardDeleteGlucoseReadingParams{ID: identifier})
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
		res, err := siteContext.SiteDB.Queries.SoftDeleteGlucoseReading(ctx, site.SoftDeleteGlucoseReadingParams{ID: identifier})
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
		res, err := siteContext.SiteDB.Queries.HardDeleteGlucoseReadingsByType(ctx, site.HardDeleteGlucoseReadingsByTypeParams{ReadingType: typeString})
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
		res, err := siteContext.SiteDB.Queries.SoftDeleteGlucoseReadingsByType(ctx, site.SoftDeleteGlucoseReadingsByTypeParams{ReadingType: typeString})
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

// Replace will delete then create the record again, BUT it will not
func (c EntriesCollection) Replace(ctx context.Context, siteContext *httputils.SiteContext, identifier string, patch map[string]any, canCreate bool) (*nsutils.Record, bool, *nsutils.HTTPError) {
	entryModel := &nightscout.Entry{}
	err := utils.FromMap(patch, entryModel)
	if err != nil {
		slog.Error("failed to map data to entry model", "error", err)
		return nil, false, nsutils.ErrInternal
	}

	getParams := site.GetGlucoseReadingParams{ID: identifier}
	deleteParams := site.HardDeleteGlucoseReadingParams{ID: identifier}
	createParams := nightscout.EntryToCreate(
		*entryModel,
		nightscout.CreateContext{CreatedByIdentityID: siteContext.IdentityModel.ID, CreatedByIdentityName: siteContext.IdentityModel.Name},
	)
	creatingNewRecord := false
	db := siteContext.SiteDB.SQLiteDB
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, false, nsutils.ErrInternal
	}
	defer tx.Rollback()

	// GET
	q := siteContext.SiteDB.Queries.WithTx(tx)
	dbModel, err := q.GetGlucoseReading(ctx, getParams)
	if err == sql.ErrNoRows {
		// Check no Record exists
		if !canCreate {
			return nil, false, nsutils.ErrNotFound
		}
		creatingNewRecord = true
	} else if err != nil {
		slog.Error("failed to get glucose reading by ID", "error", err)
		return nil, false, nsutils.ErrInternal
	}

	// VALIDATE ATTRIBUTES
	if !creatingNewRecord && dbModel != nil {
		if dbModel.ReadingType != createParams.ReadingType {
			slog.Error("cannot change reading_type attribute on glucose reading", "existingReadingType", dbModel.ReadingType, "newReadingType", createParams.ReadingType)
			return nil, false, nsutils.ErrCannotChangeReadingType
		}
	}

	createParams.ID = identifier

	// DELETE record if we are not creating a new record
	if !creatingNewRecord {
		_, err = q.HardDeleteGlucoseReading(ctx, deleteParams)
		if err != nil {
			slog.Error("failed to hard delete glucose reading in database", "error", err)
			return nil, false, nsutils.ErrInternal
		}
	}

	// CREATE
	dbCreateModel, err := q.CreateGlucoseReading(ctx, createParams)
	if err != nil {
		slog.Error("failed to create glucose reading in database", "error", err)
		return nil, false, nsutils.ErrInternal
	}

	if dbCreateModel == nil {
		slog.Error("failed to create glucose reading in database, dbCreateModel is nil")
		return nil, false, nsutils.ErrInternal
	}

	if err := tx.Commit(); err != nil {
		slog.Error("failed to commit create glucose reading in database", "error", err, "entryModel", entryModel)
		return nil, false, nsutils.ErrInternal
	}

	return ptr(c.converter(dbCreateModel)), creatingNewRecord, nil
}

func (c EntriesCollection) Patch(ctx context.Context, siteContext *httputils.SiteContext, identifier string, patch map[string]any) (*nsutils.Record, *nsutils.HTTPError) {
	entryModel := &nightscout.Entry{}
	err := utils.FromMap(patch, entryModel)
	if err != nil {
		slog.Error("failed to map data to entry model", "error", err)
		return nil, nsutils.ErrInternal
	}

	updateParams := nightscout.EntryToUpdate(
		*entryModel,
		nightscout.UpdateContext{UpdatedByIdentityID: siteContext.IdentityModel.ID, UpdatedByIdentityName: siteContext.IdentityModel.Name},
	)
	updateParams.ID = identifier

	ssss, _ := utils.ToMap[site.UpdateGlucoseReadingParams](&updateParams)
	dbUpdateModel, err := siteContext.SiteDB.Queries.UpdateGlucoseReading(ctx, updateParams)
	if err != nil {
		slog.Error("failed to update glucose reading in database", "error", err, "updateParams", ssss)
		return nil, nsutils.ErrInternal
	}
	if dbUpdateModel == nil {
		slog.Error("failed to update glucose reading in database, dbUpdateModel is nil")
		return nil, nsutils.ErrInternal
	}

	return ptr(c.converter(dbUpdateModel)), nil
}
