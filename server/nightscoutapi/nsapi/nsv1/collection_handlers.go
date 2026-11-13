package nsv1

import (
	"cmp"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"nighttroop.t1d.nz/server/nightscoutapi/nsutils"
	"nighttroop.t1d.nz/utils/apiutils"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/httputils"
)

func Search(store nsutils.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		collection, query, ok := nsutils.ReadCollectionAndQuery(w, r)
		if !ok {
			return
		}
		// With no count, a profile search is Nightscout's default count of the newest stores, as if it were sent
		if collection == nsutils.Profile && !query.CountGiven {
			query.Limit, query.CountGiven = consts.ProfilesDefaultCount, true
		}
		identifierOrType := chi.URLParam(r, "identifier_or_type")
		if identifierOrType != "" {
			// If the URL has an identifier or type segment, treat it as a filter on the default id or type field.
			if collection.IsTypeString(identifierOrType) {
				query.TypeEq = new(identifierOrType)
				// HasFilters was worked out when the query was parsed, so it must be set for a filter added after
				query.HasFilters = true
			} else {
				query.IDEq = new(identifierOrType)
			}
		}
		records := []nsutils.Document{}
		// A filter Nighttroop does not know is searched, which gives nothing
		if query.IDEq != nil && !query.UnknownFilter {
			record, err := store.Get(r.Context(), collection, *query.IDEq, query)
			// API v1 has no gone: a deleted record is not found
			if err != nil && err.Status == http.StatusGone {
				err = nsutils.ErrNotFound
			}
			// A find[_id] that matches nothing is an empty search, as in Nightscout; xDrip4iOS reads it as deleted
			if err != nil && err.Status == http.StatusNotFound && identifierOrType == "" {
				nsutils.WriteRecordsV1(w, records, query.Fields)
				return
			}
			if err != nil {
				nsutils.WriteError(w, err)
				return
			}
			records = append(records, record)
		} else {
			var err *nsutils.HTTPError
			records, err = store.Search(r.Context(), collection, query)
			if err != nil {
				nsutils.WriteError(w, err)
				return
			}
		}

		nsutils.WriteRecordsV1(w, records, query.Fields)
	}
}

// maxLeftOut is how many documents with a problem, other than a reading out of range, a batch can leave out before it
// is refused
const maxLeftOut = 10

func Create(store nsutils.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		collection, _, err := nsutils.GetCollectionAndEnsurePermission(r, nsutils.CollectionActionCreate)
		if err != nil {
			nsutils.WriteError(w, err)
			return
		}

		var body any
		if err := apiutils.ReadJSON(r, &body); err != nil {
			nsutils.WriteError(w, nsutils.BadDataBodyError)
			return
		}

		var created []nsutils.Document

		// Deal with one record being sent or an array being sent
		arrayOfModels := []any{}
		switch typed := body.(type) {
		case map[string]any:
			arrayOfModels = append(arrayOfModels, typed)
		case []any:
			arrayOfModels = typed
		default:
			slog.InfoContext(r.Context(), "Collection Create JSON type error", "type", typed)
			nsutils.WriteError(w, nsutils.BadDataBodyError)
			return
		}

		// Each record costs a duplicate check, a transaction and a fan-out, so cap a batch before storing any of it, and
		// have each MaxBatchRecords past the first cost a write, so a large batch is limited as a run of small ones
		if len(arrayOfModels) > consts.MaxV1BatchRecords {
			nsutils.WriteError(w, nsutils.ErrTooManyRecords)
			return
		}
		if !httputils.ExtraWritesWithinRateLimit(r, (len(arrayOfModels)-1)/consts.MaxBatchRecords) {
			nsutils.WriteError(w, nsutils.ErrTooManyBatchWrites)
			return
		}

		// Loop over the array of models sent. A batch of any collection leaves out a document it cannot store and stores
		// the rest, as Nightscout stores them all: xDrip+ sends a failed entries batch again with every new reading added,
		// and its device status and treatments only after it is taken, and Loop, Trio and xDrip4iOS send treatment and
		// device status batches
		_, isBatch := body.([]any)
		var left leftOut
		for _, item := range arrayOfModels {
			doc, ok := item.(map[string]any)
			// xDrip+ and xDrip4iOS send their sensor calibrations (type cal) in the same batch as their readings.
			// They are not kept, so the readings still are.
			if collection == nsutils.Entries && doc["type"] == "cal" {
				continue
			}

			// As in Nightscout, a document sent with an _id replaces the one with it, as Loop sends an override again
			// with its _id when it changes
			var record nsutils.Document
			err := nsutils.BadDataBodyError
			if id, isID := doc["_id"].(string); ok && isID && nsutils.ValidateIdentifierIsntBad(id) == nil {
				record, _, err = store.Replace(movableDate(r, collection), collection, id, doc, true)
			} else if ok {
				record, _, err = store.Create(r.Context(), collection, doc)
			}
			if isBatch && left.leaves(err) {
				continue
			}
			if err != nil {
				nsutils.WriteError(w, err)
				return
			}
			created = append(created, record)
		}
		if left.outOfRange+left.bad > 0 {
			slog.InfoContext(r.Context(), "left out documents from a batch", "collection", collection, "out_of_range", left.outOfRange, "bad", left.bad, "stored", len(created))
		}
		// A batch with nothing stored and a document with another problem is refused with it, so the client is told
		if len(created) == 0 && left.first != nil {
			nsutils.WriteError(w, left.first)
			return
		}
		// One stored with a document left out for another problem shows on the site's client errors, so the admin can
		// see what the client gets wrong, though the client is not told. Readings out of range alone are not, as a sensor
		// can send them for hours.
		if left.bad > 0 {
			httputils.SetLeftOut(r.Context(), left.summary(collection, len(created)))
		}

		nsutils.WriteRecordsV1(w, created, nil)
	}
}

// leftOut counts the documents a batch left out by why, in the order each why was first seen
type leftOut struct {
	reasons         []string
	counts          map[string]int
	outOfRange, bad int
	// first is the first problem other than a reading out of range, which refuses a batch with nothing stored
	first *nsutils.HTTPError
}

// leaves is whether a batch leaves out a document for its problem, err, and stores the rest. Only a 400, from the
// document's conversion, validation or a database CHECK, is the document's problem. Any other status, such as 401, 403,
// 412, 429 or 5xx, is the request's or the server's, so it stops the batch, as does an 11th document with a problem. A
// reading out of range is always left out, as a sensor sends 0 for hours, but each other problem can log an error, so
// only a few are.
func (l *leftOut) leaves(err *nsutils.HTTPError) bool {
	switch {
	case err == nil || err.Status != http.StatusBadRequest:
		return false
	case err == nsutils.BadGlucoseError:
		l.outOfRange++
		l.add("glucose out of range")
	case l.bad < maxLeftOut:
		l.bad, l.first = l.bad+1, cmp.Or(l.first, err)
		l.add(err.Message)
	default:
		return false
	}
	return true
}

func (l *leftOut) add(reason string) {
	if l.counts == nil {
		l.counts = map[string]int{}
	}
	if l.counts[reason] == 0 {
		l.reasons = append(l.reasons, reason)
	}
	l.counts[reason]++
}

// summary says how many documents of the collection were stored and left out, then how many were left out for each
// why, so the counts are kept when a long why is cut short
func (l *leftOut) summary(collection nsutils.Collection, stored int) string {
	total, reasons := 0, make([]string, len(l.reasons))
	for i, reason := range l.reasons {
		total += l.counts[reason]
		reasons[i] = fmt.Sprintf("%d %s", l.counts[reason], reason)
	}
	return fmt.Sprintf("Stored %d %s and left out %d: %s", stored, collection, total, strings.Join(reasons, "; "))
}

func Current(store nsutils.Store) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		collection, query, ok := nsutils.ReadCollectionAndQuery(w, r)
		if !ok {
			return
		}
		// Force limit 1 and sort decending by date
		query.Limit = 1
		query.SortOrder = nil

		records, err := store.Search(r.Context(), collection, query)
		if err != nil {
			nsutils.WriteError(w, err)
			return
		}
		// As in Nightscout, the current profile is the profile set itself, which Loop reads to be set up from it, and
		// null on a site with none
		if collection == nsutils.Profile {
			if len(records) == 0 {
				apiutils.WriteJSON(w, http.StatusOK, nil)
				return
			}
			nsutils.WriteRecordV1(w, records[0], query.Fields)
			return
		}
		nsutils.WriteRecordsV1(w, records, query.Fields)
	}

}

func Delete(store nsutils.Store) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		collection, _, err := nsutils.GetCollectionAndEnsurePermission(r, nsutils.CollectionActionDelete)
		if err != nil {
			nsutils.WriteError(w, err)
			return
		}
		permanent := r.URL.Query().Get("permanent") == "true"

		identifierOrType := chi.URLParam(r, "identifier_or_type")
		// IGNORE ANY ERROR AND JUST RETURN 0 DELETED COUNT, except for a delete by a query
		deleted := 0
		if identifierOrType == "" {
			// Only readings and treatments are deleted by a query, as xDrip4iOS and Trio do. Each find[$or]
			// alternative, as Trio deletes a reading by its id or its dateString, is checked before any is deleted
			if collection != nsutils.Entries && collection != nsutils.Treatments {
				nsutils.WriteError(w, nsutils.ErrNotImplemented)
				return
			}
			queries, err := nsutils.ParseDeleteQueries(r.URL.Query())
			for i := 0; err == nil && i < len(queries); i++ {
				var amount int
				amount, err = store.DeleteQuery(r.Context(), collection, queries[i])
				deleted += amount
			}
			if err != nil {
				nsutils.WriteError(w, err)
				return
			}
		} else if collection.IsTypeString(identifierOrType) {
			if amount, err := store.DeleteType(r.Context(), collection, identifierOrType, permanent); err == nil {
				deleted = amount
			}
		} else if err := store.Delete(r.Context(), collection, identifierOrType, permanent); err == nil {
			deleted = 1
		}
		apiutils.WriteJSON(w, http.StatusOK, map[string]any{
			"acknowledged": true,
			"deletedCount": deleted,
		})
	}
}

// Replace If not found will return 404
func Replace(store nsutils.Store) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		collection, identity, err := nsutils.GetCollectionAndEnsurePermission(r, nsutils.CollectionActionCreate)
		if err != nil {
			nsutils.WriteError(w, err)
			return
		}
		canCreate := nsutils.CollectionActionHasPermission(collection, identity, nsutils.CollectionActionCreate)

		var body map[string]any
		// A null body decodes without error to a nil map
		if err := apiutils.ReadJSON(r, &body); err != nil || body == nil {
			nsutils.WriteError(w, nsutils.BadDataBodyError)
			return
		}

		// GET THE IDENTIFIER FROM THE URL OR BODY
		identifier := chi.URLParam(r, "identifier")
		if identifier == "" {
			if uid, ok := body["_id"].(string); ok {
				identifier = uid
			}
			if identifier == "" {
				if id, ok := body["identifier"].(string); ok {
					identifier = id
				}
			}
		}

		if err := nsutils.ValidateIdentifierIsntBad(identifier); err != nil {
			nsutils.WriteError(w, err)
			return
		}

		updated, _, herr := store.Replace(movableDate(r, collection), collection, identifier, body, canCreate)
		if herr != nil {
			nsutils.WriteError(w, herr)
			return
		}
		nsutils.WriteRecordV1(w, updated, nil)
	}
}

// movableDate is the request's context, in which a treatment's replace may move its date, as Loop moves a carb entry or
// override edited to another time by sending it again with its _id, and Nightscout's v1 replaces the whole document
func movableDate(r *http.Request, collection nsutils.Collection) context.Context {
	if collection != nsutils.Treatments {
		return r.Context()
	}
	return nsutils.WithMovableDate(r.Context())
}
