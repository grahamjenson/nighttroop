package v1

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"nighttroop.t1d.nz/models/nightscout"
	"nighttroop.t1d.nz/server/nightscout_api/nsutils"
)

func Search(dep nsutils.Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		collection, _, err := nsutils.GetCollectionAndEnsurePermission(dep, r, nsutils.CollectionActionRead)
		if err != nil {
			nsutils.WriteError(w, err)
			return
		}

		query, err := nsutils.ParseV1(r.URL.Query())
		if err != nil {
			nsutils.WriteError(w, err)
			return
		}
		identifier_or_type := chi.URLParam(r, "identifier_or_type")
		if identifier_or_type != "" {
			// If the URL has an identifier or type segment, treat it as a filter on the default id or type field.
			if collection.IsTypeString(identifier_or_type) {
				query.TypeEq = ptr(nightscout.ShortString(identifier_or_type))
			} else {
				query.IDEq = ptr(nightscout.ShortString(identifier_or_type))
			}
		}

		records, err := dep.Store.Search(r.Context(), collection, query)
		if err != nil {
			nsutils.WriteError(w, err)
			return
		}
		nsutils.WriteRecordsV1(w, records, query.Fields.ToArray())
	}
}

func Create(dep nsutils.Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		collection, _, err := nsutils.GetCollectionAndEnsurePermission(dep, r, nsutils.CollectionActionCreate)
		if err != nil {
			nsutils.WriteError(w, err)
			return
		}

		var body any
		if err := nsutils.ReadJSON(r, &body); err != nil {
			nsutils.WriteError(w, nsutils.BadDataBodyError)
			return
		}
		var created []nsutils.Record

		// Deal with one record being sent or an array being sent
		arrayOfModels := []any{}
		switch typed := body.(type) {
		case map[string]any:
			arrayOfModels = append(arrayOfModels, typed)
		case []any:
			arrayOfModels = typed
		default:
			nsutils.WriteError(w, nsutils.BadDataBodyError)
			return
		}

		// Loop over the array of models sent
		for _, item := range arrayOfModels {
			doc, ok := item.(map[string]any)

			if !ok {
				nsutils.WriteError(w, nsutils.BadDataBodyError)
				return
			}

			record, err := dep.Store.Create(r.Context(), collection, doc)
			if err != nil {
				nsutils.WriteError(w, err)
				return
			}
			created = append(created, record)
		}

		nsutils.WriteRecordsV1(w, created, nil)
	}
}

func Current(dep nsutils.Dependencies) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		collection, _, err := nsutils.GetCollectionAndEnsurePermission(dep, r, nsutils.CollectionActionRead)
		if err != nil {
			nsutils.WriteError(w, err)
			return
		}

		query, err := nsutils.ParseV1(r.URL.Query())
		if err != nil {
			nsutils.WriteError(w, err)
			return
		}
		// Force limit 1 and sort decending by date
		query.Limit = ptr(nightscout.FlexibleInt64(1))
		query.SortDesc = nil
		query.Sort = nil

		records, err := dep.Store.Search(r.Context(), collection, query)
		if err != nil {
			nsutils.WriteError(w, err)
			return
		}
		nsutils.WriteRecordsV1(w, records, query.Fields.ToArray())
	}

}

func Delete(dep nsutils.Dependencies) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		collection, _, err := nsutils.GetCollectionAndEnsurePermission(dep, r, nsutils.CollectionActionDelete)
		if err != nil {
			nsutils.WriteError(w, err)
			return
		}
		permanent := r.URL.Query().Get("permanent") == "true"

		identifier_or_type := chi.URLParam(r, "identifier_or_type")
		if identifier_or_type != "" {
			if collection.IsTypeString(identifier_or_type) {
				dep.Store.DeleteType(r.Context(), collection, identifier_or_type, permanent)
				return
			} else {
				err = dep.Store.Delete(r.Context(), collection, identifier_or_type, permanent)
				if err != nil {
					nsutils.WriteError(w, err)
					return
				}
				w.WriteHeader(http.StatusNoContent)
				return
			}
		}

	}
}
