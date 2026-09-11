package collections

import (
	"context"
	"database/sql"
	"log/slog"

	"nighttroop.t1d.nz/data/sql/site"
	"nighttroop.t1d.nz/models/nightscout"
	"nighttroop.t1d.nz/server/nightscout_api/nsquery"
	"nighttroop.t1d.nz/server/nightscout_api/nsutils"
	"nighttroop.t1d.nz/utils/httputils"
)

type EntriesCollection struct {
	converter func(c *site.GlucoseReading, siteContext *httputils.SiteContext, raw bool) nsutils.Record
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

func (c EntriesCollection) Convert(dr *site.GlucoseReading, siteContext *httputils.SiteContext, raw bool) nsutils.Record {
	return c.converter(dr, siteContext, raw)
}

// Cannot change properties

func (c EntriesCollection) Validate(nsModel nightscout.Entry, dbModel *site.GlucoseReading) *nsutils.HTTPError {
	if dbModel == nil {
		return nil
	}

	if nsModel.Type != nil && nsModel.Type.String() != dbModel.ReadingType {
		slog.Error("cannot change reading_type attribute on glucose reading", "existingReadingType", dbModel.ReadingType, "newReadingType", nsModel.Type.String())
		return nsutils.ErrCannotChangeType
	}

	return nil

}

func (c EntriesCollection) ValidatePatch(nsModel nightscout.Entry, dbModel *site.GlucoseReading) *nsutils.HTTPError {
	return c.Validate(nsModel, dbModel)
}

func (c EntriesCollection) CreateDBFunc(q *site.Queries, ctx context.Context, nsModel nightscout.Entry, cc nightscout.ConverterContext, overrideIdentifier *string) (*site.GlucoseReading, error) {
	createParams := nightscout.EntryToCreate(nsModel, cc)
	if overrideIdentifier != nil {
		createParams.ID = *overrideIdentifier
	}
	return q.CreateGlucoseReading(ctx, createParams)
}

func (c EntriesCollection) GetDuplicateDBFunc(q *site.Queries, ctx context.Context, nsModel nightscout.Entry) (*site.GlucoseReading, error) {
	glucoseReadings, rerr := q.SearchGlucoseReadings(ctx, site.SearchGlucoseReadingsParams{
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

func (c EntriesCollection) GetDBFunc(q *site.Queries, ctx context.Context, identifier string, includeDeleted bool) (*site.GlucoseReading, error) {
	return q.GetGlucoseReading(ctx, site.GetGlucoseReadingParams{ID: identifier, IncludeDeleted: includeDeleted})
}

func (c EntriesCollection) HardDeleteDBFunc(q *site.Queries, ctx context.Context, identifier string) (sql.Result, error) {
	return q.HardDeleteGlucoseReading(ctx, site.HardDeleteGlucoseReadingParams{ID: identifier})
}

func (c EntriesCollection) SoftDeleteDBFunc(q *site.Queries, ctx context.Context, identifier string) (sql.Result, error) {
	return q.SoftDeleteGlucoseReading(ctx, site.SoftDeleteGlucoseReadingParams{ID: identifier})
}

func (c EntriesCollection) HardDeleteTypeDBFunc(q *site.Queries, ctx context.Context, typeString string) (sql.Result, error) {
	return q.HardDeleteGlucoseReadingsByType(ctx, site.HardDeleteGlucoseReadingsByTypeParams{ReadingType: typeString})
}

func (c EntriesCollection) SoftDeleteTypeDBFunc(q *site.Queries, ctx context.Context, typeString string) (sql.Result, error) {
	return q.SoftDeleteGlucoseReadingsByType(ctx, site.SoftDeleteGlucoseReadingsByTypeParams{ReadingType: typeString})
}

func (c EntriesCollection) PatchDBFunc(q *site.Queries, ctx context.Context, nsModel nightscout.Entry, identifier string, dbModel *site.GlucoseReading, uc nightscout.ConverterContext) (*site.GlucoseReading, error) {
	updateParams := nightscout.EntryToUpdate(nsModel, uc)
	updateParams.ID = identifier

	return q.UpdateGlucoseReading(ctx, updateParams)
}

/////
// END OF FUNCTIONS USED BY META COLLECTIONS
/////

func (c EntriesCollection) Create(ctx context.Context, siteContext *httputils.SiteContext, data map[string]any, raw bool) (*nsutils.Record, bool, *nsutils.HTTPError) {
	return MetaCreate(c, ctx, siteContext, data, raw)
}

func (c EntriesCollection) Get(ctx context.Context, siteContext *httputils.SiteContext, identifier string, query nsquery.Query) (*nsutils.Record, *nsutils.HTTPError) {
	return MetaGet(c, ctx, siteContext, identifier, query.Raw, query.IncludeDeleted)
}

func (c EntriesCollection) Delete(ctx context.Context, siteContext *httputils.SiteContext, identifier string, permanent bool) *nsutils.HTTPError {
	return MetaDelete(c, ctx, siteContext, identifier, permanent)
}

func (c EntriesCollection) DeleteType(ctx context.Context, siteContext *httputils.SiteContext, typeString string, permanent bool) (int, *nsutils.HTTPError) {
	return MetaDeleteType(c, ctx, siteContext, typeString, permanent)
}

func (c EntriesCollection) Patch(ctx context.Context, siteContext *httputils.SiteContext, identifier string, patch map[string]any, raw bool) (*nsutils.Record, *nsutils.HTTPError) {
	return MetaPatch(c, ctx, siteContext, identifier, patch, raw)
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

	return nsutils.CreateRecordSlice(entries, c.Convert, siteContext, query.Raw), nil
}

// Replace will delete then create the record again, BUT it will not
func (c EntriesCollection) Replace(ctx context.Context, siteContext *httputils.SiteContext, identifier string, patch map[string]any, canCreate bool, raw bool) (*nsutils.Record, bool, *nsutils.HTTPError) {
	return MetaReplace(c, ctx, siteContext, identifier, patch, canCreate, raw)
}
