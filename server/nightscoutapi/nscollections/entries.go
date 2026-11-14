package nscollections

import (
	"cmp"
	"context"
	"database/sql"
	"log/slog"
	"time"

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
	if err := validateEntryDate(nsModel); err != nil {
		return err
	}
	return validateEntryType(nsModel, dbModel)
}

// validateEntryDate refuses a reading dated well ahead of now, beyond a phone clock's error, as followers, Loop's and
// Trio's Nightscout CGM among them, show the newest reading as the current one
func validateEntryDate(nsModel nightscoutmodels.Entry) *nsutils.HTTPError {
	if date := nsModel.GetDate(); date != nil && *date > time.Now().Add(consts.MaxFutureReading).UnixMilli() {
		return nsutils.ErrFutureReading
	}
	return nil
}

// Rows are its reading, whose glucose is out of range when far outside real use, checked in mg/dL as it is stored
func (c EntriesCollection) Rows(nsModel nightscoutmodels.Entry) []sitedb.TableRow {
	return []sitedb.TableRow{{Table: "glucose_readings", Row: nightscoutmodels.EntryToCreate(nsModel, nightscoutmodels.ConverterContext{})}}
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

// FallbackKey is a reading's date and type, whichever device sent it and whatever its _id, as Nightscout upserts an
// entry on its date and type: xDrip+ sends a careportal's BG Check back as its own meter reading, and Trio, with
// Nightscout as its CGM, uploads each reading it downloads under its own id and device
func (c EntriesCollection) FallbackKey() FallbackKey[nightscoutmodels.Entry, sitedb.GlucoseReading] {
	return FallbackKey[nightscoutmodels.Entry, sitedb.GlucoseReading]{
		search: func(q *sitedb.Queries, ctx context.Context, nsModel nightscoutmodels.Entry, limit int64) ([]*sitedb.GlucoseReading, error) {
			readingType := nightscoutmodels.SetReadingType(nsModel)
			return q.SearchGlucoseReadings(ctx, sitedb.SearchGlucoseReadingsParams{DateEQ: nsModel.Date.Int64p(), TypeEq: &readingType, Limit: limit})
		},
		fields: []KeyField[nightscoutmodels.Entry, sitedb.GlucoseReading]{
			sameDate[nightscoutmodels.Entry, sitedb.GlucoseReading](),
			{"type", func(sent nightscoutmodels.Entry, stored *sitedb.GlucoseReading) bool {
				return nightscoutmodels.SetReadingType(sent) == stored.ReadingType
			}},
		},
		whateverItsIdentifier: true,
	}
}

// Refreshes a reading sent again of its type, as Nightscout upserts a reading on its date and type: xDrip4iOS smooths
// and rebuilds readings by sending them again without their _id
func (c EntriesCollection) Refreshes(nsModel nightscoutmodels.Entry, stored *sitedb.GlucoseReading) bool {
	return nightscoutmodels.SetReadingType(nsModel) == stored.ReadingType
}

// SameUploader is whether the identity and device that stored the reading sent it again, so a looper's own correction
// refreshes it. A copy from another identity or device, like an xDrip+ follower's echo of AAPS's reading, only fills
// what it is missing. The device is only what the app says; the identity is the token's, which no other app can send.
func (c EntriesCollection) SameUploader(nsModel nightscoutmodels.Entry, stored *sitedb.GlucoseReading, cc nightscoutmodels.ConverterContext) bool {
	return *cmp.Or(stored.CreatedByIdentityID, new(int64(0))) == cc.IdentityID && kept(nsModel.Device.Stringp(), stored.CreatedByDeviceName)
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

// RewriteModel keeps the reading's notes and reason, from the site, and how it was measured, which an entry does not have
func (c EntriesCollection) RewriteModel(q *sitedb.Queries, ctx context.Context, nsModel nightscoutmodels.Entry, _ map[string]any, stored *sitedb.GlucoseReading, cc nightscoutmodels.ConverterContext) (*sitedb.GlucoseReading, error) {
	params := nightscoutmodels.EntryToCreate(nsModel, cc)
	params.ID, params.Notes, params.Reason, params.MeasurementMethod = stored.ID, stored.Notes, stored.Reason, stored.MeasurementMethod
	return q.RewriteGlucoseReading(ctx, sitedb.RewriteGlucoseReadingParams(params))
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
