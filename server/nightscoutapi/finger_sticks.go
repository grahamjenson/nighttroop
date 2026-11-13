package nightscoutapi

import (
	"cmp"
	"context"
	"log/slog"
	"math"
	"slices"
	"strings"

	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/server/nightscoutapi/nscollections"
	"nighttroop.t1d.nz/server/nightscoutapi/nsquery"
	"nighttroop.t1d.nz/server/nightscoutapi/nsutils"
	"nighttroop.t1d.nz/utils/httputils"
)

// bgCheck is a meter reading as a read-only BG Check, as API v3 gives it, and false for any other reading
func (nts *NightTroopAPIStore) bgCheck(reading *sitedb.GlucoseReading, siteContext *httputils.SiteContext) (nsutils.Document, bool) {
	entries, _ := nts.stores[nsutils.Entries].(interface {
		Record(model any, siteContext *httputils.SiteContext) (nsutils.Document, bool)
	})
	if doc, ok := entries.Record(reading, siteContext); ok {
		return nscollections.BGCheck(doc, siteContext)
	}
	return nil, false
}

// bgChecks are the readings as BG Checks
func (nts *NightTroopAPIStore) bgChecks(readings []*sitedb.GlucoseReading, siteContext *httputils.SiteContext) []nsutils.Document {
	docs := make([]nsutils.Document, 0, len(readings))
	for _, reading := range readings {
		if doc, ok := nts.bgCheck(reading, siteContext); ok {
			docs = append(docs, doc)
		}
	}
	return docs
}

// withFingerSticks is an API v3 search of treatments by date, with the meter readings in no treatment then as BG Checks
func (nts *NightTroopAPIStore) withFingerSticks(ctx context.Context, siteContext *httputils.SiteContext, query nsquery.Query, search func(nsquery.Query) ([]nsutils.Document, *nsutils.HTTPError)) ([]nsutils.Document, *nsutils.HTTPError) {
	if !query.V3 || query.Limit <= 0 || !query.DatesOnly() {
		return search(query)
	}
	// Both searched for a page from the first, as the page asked for is of both together
	wide := query
	wide.Skip, wide.Limit = 0, query.Skip+query.Limit
	docs, httpErr := search(wide)
	if httpErr != nil {
		return nil, httpErr
	}
	from := max(*cmp.Or(query.DateGTE, new(int64(0))), *cmp.Or(query.DateGT, new(int64(-1)))+1, *cmp.Or(query.DateEQ, new(int64(0))))
	to := min(*cmp.Or(query.DateLTE, new(int64(math.MaxInt64))), *cmp.Or(query.DateLT, new(int64(math.MaxInt64)))-1, *cmp.Or(query.DateEQ, new(int64(math.MaxInt64))))
	ascending := query.SortOrder != nil && *query.SortOrder == nsquery.SortOccurredAtAsc
	q := siteContext.SiteDB.Queries()
	var readings []*sitedb.GlucoseReading
	var err error
	if ascending {
		readings, err = q.FingerSticksAscending(ctx, sitedb.FingerSticksAscendingParams{FromMs: from, ToMs: to, Limit: wide.Limit})
	} else {
		readings, err = q.FingerSticks(ctx, sitedb.FingerSticksParams{FromMs: from, ToMs: to, Limit: wide.Limit})
	}
	if err != nil {
		slog.ErrorContext(ctx, "failed to search the finger sticks", "error", err)
		return nil, nsutils.ErrInternal
	}
	docs = append(docs, nts.bgChecks(readings, siteContext)...)
	slices.SortStableFunc(docs, func(a, b nsutils.Document) int {
		if ascending {
			return cmp.Compare(a.OccurredAt(), b.OccurredAt())
		}
		return cmp.Compare(b.OccurredAt(), a.OccurredAt())
	})
	return docs[min(int(query.Skip), len(docs)):min(int(query.Skip+query.Limit), len(docs))], nil
}

// historyWithFingerSticks is API v3's history of treatments with the meter readings in no treatment that changed, as BG
// Checks, in the order history gives them
func (nts *NightTroopAPIStore) historyWithFingerSticks(ctx context.Context, siteContext *httputils.SiteContext, since int64, query nsquery.Query, docs []nsutils.Document) ([]nsutils.Document, *nsutils.HTTPError) {
	if !query.V3 || query.HasFilters || query.Limit <= 0 {
		return docs, nil
	}
	readings, err := siteContext.SiteDB.Queries().FingerSticksChanged(ctx, sitedb.FingerSticksChangedParams{Since: since, Limit: query.Limit})
	if err != nil {
		slog.ErrorContext(ctx, "failed to read the finger sticks' history", "error", err)
		return nil, nsutils.ErrInternal
	}
	docs = append(docs, nts.bgChecks(readings, siteContext)...)
	slices.SortStableFunc(docs, func(a, b nsutils.Document) int {
		return cmp.Or(cmp.Compare(a.Modified(), b.Modified()), cmp.Compare(a.OccurredAt(), b.OccurredAt()), strings.Compare(a.Identifier(), b.Identifier()))
	})
	return docs[:min(int(query.Limit), len(docs))], nil
}

// fingerStick is the meter reading in no treatment a BG Check's identifier is of, so a client can read or delete it
func (nts *NightTroopAPIStore) fingerStick(ctx context.Context, siteContext *httputils.SiteContext, collection nsutils.Collection, identifier string) (*sitedb.GlucoseReading, bool) {
	readingID, ok := strings.CutSuffix(identifier, nscollections.FingerSuffix)
	if !ok || collection != nsutils.Treatments {
		return nil, false
	}
	q := siteContext.SiteDB.Queries()
	reading, err := q.GetGlucoseReading(ctx, sitedb.GetGlucoseReadingParams{ID: readingID, IncludeDeleted: true})
	if err != nil || !nscollections.FingerStick(reading) {
		return nil, false
	}
	treated, err := q.GlucoseReadingInTreatment(ctx, sitedb.GlucoseReadingInTreatmentParams{ID: &readingID})
	return reading, err == nil && !treated
}
