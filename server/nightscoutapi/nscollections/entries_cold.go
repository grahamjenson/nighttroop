package nscollections

import (
	"context"
	"encoding/json"
	"math"
	"slices"
	"time"

	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/retention"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/server/nightscoutapi/nsquery"
	"nighttroop.t1d.nz/utils/httputils"
)

// lastMs is the last time a record can have
const lastMs = int64(253402300799999)

// coldCanMatch is whether a reading made from a cold day can be what the query searches for: a CGM reading, searched
// for by its date, its type and its glucose alone. A reading from a cold day has no device, app or time it was
// written, and its ID is made, so a search by one of those is of rows only.
func coldCanMatch(query nsquery.Query) bool {
	for _, types := range []struct {
		list *string
		in   bool
	}{{query.TypeIn, true}, {query.TypeNin, false}} {
		var listed []string
		if types.list != nil && (json.Unmarshal([]byte(*types.list), &listed) != nil || slices.Contains(listed, "sgv") != types.in) {
			return false
		}
	}
	if (query.TypeEq != nil && *query.TypeEq != "sgv") || (query.TypeNeq != nil && *query.TypeNeq == "sgv") || query.IncludeDeleted {
		return false
	}
	query.TypeEq, query.TypeNeq, query.TypeIn, query.TypeNin = nil, nil, nil, nil
	query.SgvEQ, query.SgvGTE, query.SgvGT, query.SgvLTE, query.SgvLT = nil, nil, nil, nil, nil
	return query.DatesOnly()
}

// SearchWithCold is a search for readings, the newest first, with those of the sealed days whose rows are deleted,
// and whether it was one: only where the server deletes rows, and only a search a cold day's reading can match that
// reaches back before the hot days. The rows are read first. A step of a cold day that a row is still in is not made
// again, and no cold day is read that is older than the last row of the page.
func (c EntriesCollection) SearchWithCold(q *sitedb.Queries, ctx context.Context, siteContext *httputils.SiteContext, query nsquery.Query) ([]*sitedb.GlucoseReading, bool, error) {
	cfg := siteContext.Config
	if cfg == nil || !cfg.RetentionDeleteRows || !coldCanMatch(query) {
		return nil, false, nil
	}
	orZero, orLast := func(at *int64, add int64) int64 {
		if at == nil {
			return 0
		}
		return *at + add
	}, func(at *int64, add int64) int64 {
		if at == nil {
			return lastMs
		}
		return *at + add
	}
	from := max(orZero(query.DateGTE, 0), orZero(query.DateGT, 1), orZero(query.DateEQ, 0))
	to := min(orLast(query.DateLTE, 1), orLast(query.DateLT, 0), orLast(query.DateEQ, 1), retention.HotSince(time.Now(), cfg.RetentionHotDays).UnixMilli())
	if from >= to {
		return nil, false, nil
	}
	want := int(query.Skip + query.Limit)
	rowsQuery := query
	rowsQuery.Skip, rowsQuery.Limit = 0, int64(want)
	rows, err := c.SearchModels(q, ctx, rowsQuery)
	if err != nil {
		return nil, true, err
	}
	if len(rows) == want && want > 0 {
		from = max(from, rows[want-1].OccurredAt)
	}
	step := retention.Glucose.Step.Milliseconds()
	inRows := map[int64]bool{}
	for _, row := range rows {
		inRows[row.OccurredAt/step] = inRows[row.OccurredAt/step] || row.ReadingType == "sgv"
	}
	sgv := func(bound *int64, otherwise float64) float64 {
		if bound == nil {
			return otherwise
		}
		return float64(*bound)
	}
	cold, err := db.ColdGlucoseReadings(ctx, q, from, to, want, func(at int64, mgdl float64) bool {
		return !inRows[at/step] && (query.SgvEQ == nil || mgdl == float64(*query.SgvEQ)) &&
			mgdl >= sgv(query.SgvGTE, 0) && mgdl > sgv(query.SgvGT, -1) && mgdl <= sgv(query.SgvLTE, math.MaxFloat64) && mgdl < sgv(query.SgvLT, math.MaxFloat64)
	})
	if err != nil {
		return nil, true, err
	}
	merged := append(rows, cold...)
	slices.SortStableFunc(merged, func(a, b *sitedb.GlucoseReading) int { return int(min(max(b.OccurredAt-a.OccurredAt, -1), 1)) })
	return merged[min(int(query.Skip), len(merged)):min(want, len(merged))], true, nil
}
