package nscollections

import (
	"context"
	"database/sql"
	"log/slog"

	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/models/nightscoutmodels"
	"nighttroop.t1d.nz/server/nightscoutapi/nsquery"
	"nighttroop.t1d.nz/server/nightscoutapi/nsutils"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/httputils"
)

type EntriesCollection struct {
	converter func(c *sitedb.GlucoseReading, siteContext *httputils.SiteContext, raw bool) nsutils.Record
}

func NewEntriesCollection() CollectionStore {
	return NewCollection[nightscoutmodels.Entry, sitedb.GlucoseReading](EntriesCollection{
		converter: nsutils.DBModelToRecord(nsutils.Entries, nightscoutmodels.DBModelToEntry),
	})
}

func (c EntriesCollection) Name() string {
	return "entries"
}

func (c EntriesCollection) Convert(dr *sitedb.GlucoseReading, siteContext *httputils.SiteContext, raw bool) nsutils.Record {
	return c.converter(dr, siteContext, raw)
}

// Cannot change properties

func (c EntriesCollection) Validate(nsModel nightscoutmodels.Entry, dbModel *sitedb.GlucoseReading) *nsutils.HTTPError {
	if err := validateEntryMgdl(nsModel); err != nil {
		return err
	}
	return validateEntryType(nsModel, dbModel)
}

// A patch without sgv or mbg leaves the reading as it is
func (c EntriesCollection) ValidatePatch(nsModel nightscoutmodels.Entry, dbModel *sitedb.GlucoseReading) *nsutils.HTTPError {
	if nsModel.SGV != nil || nsModel.MBG != nil {
		if err := validateEntryMgdl(nsModel); err != nil {
			return err
		}
	}
	return validateEntryType(nsModel, dbModel)
}

// Refuse glucose far outside real use, checked in mg/dL as it is stored
func validateEntryMgdl(nsModel nightscoutmodels.Entry) *nsutils.HTTPError {
	if nsModel.Mgdl == nil || *nsModel.Mgdl <= 0 || *nsModel.Mgdl > consts.MaxGlucoseMgdl {
		return nsutils.BadGlucoseError
	}
	return nil
}

func validateEntryType(nsModel nightscoutmodels.Entry, dbModel *sitedb.GlucoseReading) *nsutils.HTTPError {
	if dbModel == nil {
		return nil
	}

	if nsModel.Type != nil && nsModel.Type.String() != dbModel.ReadingType {
		slog.Error("cannot change reading_type attribute on glucose reading", "existingReadingType", dbModel.ReadingType, "newReadingType", nsModel.Type.String())
		return nsutils.ErrCannotChangeType
	}

	return nil

}

func (c EntriesCollection) CreateModel(q *sitedb.Queries, ctx context.Context, nsModel nightscoutmodels.Entry, cc nightscoutmodels.ConverterContext, overrideIdentifier *string) (*sitedb.GlucoseReading, error) {
	createParams := nightscoutmodels.EntryToCreate(nsModel, cc)
	if overrideIdentifier != nil {
		createParams.ID = *overrideIdentifier
	}
	return q.CreateGlucoseReading(ctx, createParams)
}

func (c EntriesCollection) GetDuplicate(q *sitedb.Queries, ctx context.Context, nsModel nightscoutmodels.Entry) (*sitedb.GlucoseReading, error) {
	// Without a date every filter below is empty and would match the newest row
	if nsModel.Date == nil {
		return nil, nil
	}
	glucoseReadings, rerr := q.SearchGlucoseReadings(ctx, sitedb.SearchGlucoseReadingsParams{
		DateEQ:   nsModel.Date.Int64p(),
		DeviceEq: nsModel.Device.Stringp(),
		Limit:    1,
	})
	if rerr != nil {
		return nil, rerr
	}
	if len(glucoseReadings) > 0 {
		return glucoseReadings[0], nil
	}
	return nil, nil
}

func (c EntriesCollection) GetModel(q *sitedb.Queries, ctx context.Context, identifier string, includeDeleted bool) (*sitedb.GlucoseReading, error) {
	return q.GetGlucoseReading(ctx, sitedb.GetGlucoseReadingParams{ID: identifier, IncludeDeleted: includeDeleted})
}

func (c EntriesCollection) HardDelete(q *sitedb.Queries, ctx context.Context, identifier string) (sql.Result, error) {
	return q.HardDeleteGlucoseReading(ctx, sitedb.HardDeleteGlucoseReadingParams{ID: identifier})
}

func (c EntriesCollection) SoftDelete(q *sitedb.Queries, ctx context.Context, identifier string) (sql.Result, error) {
	return q.SoftDeleteGlucoseReading(ctx, sitedb.SoftDeleteGlucoseReadingParams{ID: identifier})
}

func (c EntriesCollection) HardDeleteType(q *sitedb.Queries, ctx context.Context, typeString string) (sql.Result, error) {
	return q.HardDeleteGlucoseReadingsByType(ctx, sitedb.HardDeleteGlucoseReadingsByTypeParams{ReadingType: typeString})
}

func (c EntriesCollection) SoftDeleteType(q *sitedb.Queries, ctx context.Context, typeString string) (sql.Result, error) {
	return q.SoftDeleteGlucoseReadingsByType(ctx, sitedb.SoftDeleteGlucoseReadingsByTypeParams{ReadingType: typeString})
}

func (c EntriesCollection) UpdateModel(q *sitedb.Queries, ctx context.Context, nsModel nightscoutmodels.Entry, identifier string, dbModel *sitedb.GlucoseReading, uc nightscoutmodels.ConverterContext) (*sitedb.GlucoseReading, error) {
	updateParams := nightscoutmodels.EntryToUpdate(nsModel, uc)
	updateParams.ID = identifier

	return q.UpdateGlucoseReading(ctx, updateParams)
}

func (c EntriesCollection) SearchModels(q *sitedb.Queries, ctx context.Context, query nsquery.Query) ([]*sitedb.GlucoseReading, error) {
	if !query.HasFilters {
		return q.AllGlucoseReadings(ctx, nsquery.GlucoseAll(query))
	}
	return q.SearchGlucoseReadings(ctx, nsquery.GlucoseSearch(query))
}
