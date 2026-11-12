package nsv1

import (
	"log/slog"
	"net/http"

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

		// Loop over the array of models sent
		_, isBatch := body.([]any)
		leftOut := false
		for _, item := range arrayOfModels {
			doc, ok := item.(map[string]any)

			if !ok {
				nsutils.WriteError(w, nsutils.BadDataBodyError)
				return
			}
			// xDrip+ and xDrip4iOS send their sensor calibrations (type cal) in the same batch as their readings.
			// They are not kept, so the readings still are.
			if collection == nsutils.Entries && doc["type"] == "cal" {
				continue
			}

			// As in Nightscout, a document sent with an _id replaces the one with it, as Loop sends an override again
			// with its _id when it changes
			var record nsutils.Document
			var err *nsutils.HTTPError
			if id, ok := doc["_id"].(string); ok && nsutils.ValidateIdentifierIsntBad(id) == nil {
				record, _, err = store.Replace(r.Context(), collection, id, doc, true)
			} else {
				record, _, err = store.Create(r.Context(), collection, doc)
			}
			// A reading out of range in a batch is left out so the rest are stored, as xDrip+ sends a failed batch
			// again with every new reading added, and Nightscout stores them all
			if err == nsutils.BadGlucoseError && isBatch {
				slog.InfoContext(r.Context(), "left out an entry with glucose out of range from a batch")
				leftOut = true
				continue
			}
			if err != nil {
				nsutils.WriteError(w, err)
				return
			}
			created = append(created, record)
		}
		// A batch with only readings out of range is refused, so the client is told
		if leftOut && len(created) == 0 {
			nsutils.WriteError(w, nsutils.BadGlucoseError)
			return
		}

		nsutils.WriteRecordsV1(w, created, nil)
	}
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

		updated, _, herr := store.Replace(r.Context(), collection, identifier, body, canCreate)
		if herr != nil {
			nsutils.WriteError(w, herr)
			return
		}
		nsutils.WriteRecordV1(w, updated, nil)
	}
}
