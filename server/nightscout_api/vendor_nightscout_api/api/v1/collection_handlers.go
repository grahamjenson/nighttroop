package v1

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/nsutils"
)

func Search(dep nsutils.Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		collection, _, err := nsutils.GetCollectionAndEnsurePermission(dep, r, nsutils.CollectionActionRead)
		if err != nil {
			nsutils.WriteError(w, err)
			return
		}

		query := nsutils.ParseV1(r.URL.Query(), collection.DefaultDateField())

		identifier_or_type := chi.URLParam(r, "identifier_or_type")
		if identifier_or_type != "" {
			// If the URL has an identifier or type segment, treat it as a filter on the default id or type field.
			if collection.IsTypeString(identifier_or_type) {
				query.Filters = append(query.Filters, nsutils.Filter{Field: "type", Op: "eq", Value: identifier_or_type})
			} else {
				query.Filters = append(query.Filters, nsutils.Filter{Field: "_id", Op: "eq", Value: identifier_or_type})
			}
		}

		records, err := dep.Store.Search(r.Context(), collection, query)
		if err != nil {
			nsutils.WriteError(w, err)
			return
		}
		nsutils.WriteRecordsV1(w, records, query.Fields)
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
		switch typed := body.(type) {
		case map[string]any:
			doc := typed
			doc, err := collection.DefaultNormalizeValidateCreateDoc(doc)
			if err != nil {
				nsutils.WriteError(w, err)
				return
			}
			record, err := dep.Store.Create(r.Context(), collection, doc)
			if err != nil {
				nsutils.WriteError(w, err)
				return
			}
			created = append(created, record)
		case []any:
			for _, item := range typed {
				doc, ok := item.(map[string]any)
				if !ok {
					nsutils.WriteError(w, nsutils.BadDataBodyError)
					return
				}
				doc, err := collection.DefaultNormalizeValidateCreateDoc(doc)
				if err != nil {
					nsutils.WriteError(w, err)
					return
				}
				record, err := dep.Store.Create(r.Context(), collection, doc)
				if err != nil {
					nsutils.WriteError(w, err)
					return
				}
				created = append(created, record)
			}
		default:
			nsutils.WriteError(w, nsutils.BadDataBodyError)
			return
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

		query := nsutils.ParseV1(r.URL.Query(), collection.DefaultDateField())
		// Force limit 1 and sort decending by date
		query.Limit = 1
		query.SortField = collection.DefaultDateField()
		query.SortDesc = true

		records, err := dep.Store.Search(r.Context(), collection, query)
		if err != nil {
			nsutils.WriteError(w, err)
			return
		}
		nsutils.WriteRecordsV1(w, records, query.Fields)
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
