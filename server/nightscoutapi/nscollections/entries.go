package nscollections

import (
	"context"
	"database/sql"
	"log/slog"

	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/models/nightscoutmodels"
	"nighttroop.t1d.nz/server/nightscoutapi/nsquery"
	"nighttroop.t1d.nz/server/nightscoutapi/nsutils"
	"nighttroop.t1d.nz/utils/apiutils/apiquery"
	"nighttroop.t1d.nz/utils/consts"
)

type EntriesCollection struct {
	named[nightscoutmodels.Entry, sitedb.GlucoseReading]
}

func NewEntriesCollection() CollectionStore {
	return NewCollection[nightscoutmodels.Entry, sitedb.GlucoseReading](EntriesCollection{
		named[nightscoutmodels.Entry, sitedb.GlucoseReading]{nsutils.Entries, nightscoutmodels.DBModelToEntry},
	})
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
	return first(q.SearchGlucoseReadings(ctx, sitedb.SearchGlucoseReadingsParams{DateEQ: nsModel.Date.Int64p(), DeviceEq: nsModel.Device.Stringp(), Limit: 1}))
}

// Refresh gives a reading sent again the values it was sent with, as Nightscout upserts a reading on its date and type:
// xDrip4iOS smooths and rebuilds readings by sending them again without their _id. One of another type, or sent with
// the values it has, is left as it is.
func (c EntriesCollection) Refresh(q *sitedb.Queries, ctx context.Context, nsModel nightscoutmodels.Entry, stored *sitedb.GlucoseReading, cc nightscoutmodels.ConverterContext) (*sitedb.GlucoseReading, bool, error) {
	p := nightscoutmodels.EntryToUpdate(nsModel, cc)
	if nightscoutmodels.SetReadingType(nsModel) != stored.ReadingType || (kept(p.Mgdl, &stored.Mgdl) && kept(p.Delta, stored.Delta) &&
		kept(p.Direction, stored.Direction) && kept(signal(p.Noise), stored.Noise) && kept(p.Filtered, stored.Filtered) &&
		kept(p.Unfiltered, stored.Unfiltered) && kept(signal(p.Rssi), stored.Rssi)) {
		return stored, false, nil
	}
	p.ID = stored.ID
	updated, err := q.UpdateGlucoseReading(ctx, p)
	return updated, true, err
}

// kept is whether a value sent leaves the stored one as it is, as one not sent does
func kept[T comparable](sent *T, stored *T) bool {
	return sent == nil || (stored != nil && *sent == *stored)
}

// signal is a CGM's noise or rssi sent, nil for 0, which means it has none and is stored as NULL
func signal(sent *float64) *float64 {
	if sent != nil && *sent == 0 {
		return nil
	}
	return sent
}

func (c EntriesCollection) GetModel(q *sitedb.Queries, ctx context.Context, identifier string, includeDeleted bool) (*sitedb.GlucoseReading, error) {
	return q.GetGlucoseReading(ctx, sitedb.GetGlucoseReadingParams{ID: identifier, IncludeDeleted: includeDeleted})
}

func (c EntriesCollection) Delete(q *sitedb.Queries, ctx context.Context, identifier string, permanent bool) (sql.Result, error) {
	if permanent {
		return q.HardDeleteGlucoseReading(ctx, sitedb.HardDeleteGlucoseReadingParams{ID: identifier})
	}
	return q.SoftDeleteGlucoseReading(ctx, sitedb.SoftDeleteGlucoseReadingParams{ID: identifier})
}

// DeleteType deletes the entries of the type, and when permanent those already soft deleted too. A request deletes at
// most a search's worth, the newest first, and says how many: the rest are deleted by asking again.
func (c EntriesCollection) DeleteType(q *sitedb.Queries, ctx context.Context, typeString string, permanent bool) (sql.Result, error) {
	if permanent {
		return q.HardDeleteGlucoseReadingsByType(ctx, sitedb.HardDeleteGlucoseReadingsByTypeParams{ReadingType: typeString, Limit: apiquery.MaxQueryLimit})
	}
	return q.SoftDeleteGlucoseReadingsByType(ctx, sitedb.SoftDeleteGlucoseReadingsByTypeParams{ReadingType: typeString, Limit: apiquery.MaxQueryLimit})
}

func (c EntriesCollection) UpdateModel(q *sitedb.Queries, ctx context.Context, nsModel nightscoutmodels.Entry, identifier string, dbModel *sitedb.GlucoseReading, uc nightscoutmodels.ConverterContext) (*sitedb.GlucoseReading, error) {
	updateParams := nightscoutmodels.EntryToUpdate(nsModel, uc)
	updateParams.ID = identifier

	return q.UpdateGlucoseReading(ctx, updateParams)
}

func (c EntriesCollection) AscendingIDs(q *sitedb.Queries, ctx context.Context, query nsquery.Query) ([]string, error) {
	return q.GlucoseReadingAscendingIDs(ctx, nsquery.GlucoseAscending(query))
}

func (c EntriesCollection) HistoryIDs(q *sitedb.Queries, ctx context.Context, since, limit int64) ([]string, error) {
	return q.GlucoseReadingHistoryIDs(ctx, sitedb.GlucoseReadingHistoryIDsParams{Since: since, Limit: limit})
}

func (c EntriesCollection) SearchModels(q *sitedb.Queries, ctx context.Context, query nsquery.Query) ([]*sitedb.GlucoseReading, error) {
	if !query.HasFilters {
		return q.AllGlucoseReadings(ctx, nsquery.GlucoseAll(query))
	}
	return q.SearchGlucoseReadings(ctx, nsquery.GlucoseSearch(query))
}
