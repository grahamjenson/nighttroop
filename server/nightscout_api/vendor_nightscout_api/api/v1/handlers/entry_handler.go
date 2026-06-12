package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/nsutils"
)

// EntriesCurrent returns the newest entry.
func EntriesCurrent(dep nsutils.Dependencies) func(http.ResponseWriter, *http.Request, *nsutils.Identity) {
	return func(w http.ResponseWriter, r *http.Request, identity *nsutils.Identity) {
		query := nsutils.DefaultQuery()
		query.Limit = 1
		records, err := dep.Store.Search(r.Context(), "entries", query)
		if err != nil {
			nsutils.WriteJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
			return
		}
		nsutils.WriteRecordsV1(w, records, nil)
	}
}

// EntriesList lists entries using the Nightscout v1 query format.
func EntriesList(dep nsutils.Dependencies) func(http.ResponseWriter, *http.Request, *nsutils.Identity) {
	return func(w http.ResponseWriter, r *http.Request, identity *nsutils.Identity) {
		query := nsutils.ParseV1(r.URL.Query(), "date")
		records, err := dep.Store.Search(r.Context(), "entries", query)
		if err != nil {
			nsutils.WriteJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
			return
		}
		nsutils.WriteRecordsV1(w, records, nil)
	}
}

// EntriesSpec resolves entry sub-routes such as current, sgv, mbg, or a
// specific identifier.
func EntriesSpec(dep nsutils.Dependencies) func(http.ResponseWriter, *http.Request, *nsutils.Identity) {
	return func(w http.ResponseWriter, r *http.Request, identity *nsutils.Identity) {
		spec := chi.URLParam(r, "spec")
		switch spec {
		case "current":
			EntriesCurrent(dep)(w, r, identity)
			return
		case "sgv", "mbg":
			query := nsutils.ParseV1(r.URL.Query(), "date")
			query.Filters = append(query.Filters, nsutils.Filter{Field: "type", Op: "eq", Value: spec})
			records, err := dep.Store.Search(r.Context(), "entries", query)
			if err != nil {
				nsutils.WriteJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
				return
			}
			nsutils.WriteRecordsV1(w, records, nil)
			return
		default:
			record, err := dep.Store.Get(r.Context(), "entries", spec)
			if err != nil {
				status, body := nsutils.RequireRecord(err)
				nsutils.WriteJSON(w, status, body)
				return
			}
			nsutils.WriteRecordsV1(w, []nsutils.Record{record}, nil)
		}
	}
}

// EntriesCreate creates entries or returns a preview when persist is false.
func EntriesCreate(dep nsutils.Dependencies, persist bool) func(http.ResponseWriter, *http.Request, *nsutils.Identity) {
	return func(w http.ResponseWriter, r *http.Request, identity *nsutils.Identity) {
		var body any
		if err := nsutils.ReadJSON(r, &body); err != nil {
			nsutils.WriteJSON(w, http.StatusBadRequest, map[string]any{"status": http.StatusBadRequest})
			return
		}
		preview := make([]map[string]any, 0)
		persisted := make([]nsutils.Record, 0)
		apply := func(doc map[string]any) error {
			doc, err := nsutils.Entries.DefaultNormalizeValidateCreateDoc(doc)
			if err != nil {
				return err
			}
			if persist {
				record, err := dep.Store.Create(r.Context(), "entries", doc, identity.Name)
				if err != nil {
					return err
				}
				persisted = append(persisted, record)
			} else {
				// Return id as empty, because it is not a real record
				doc["_id"] = ""
				doc["identifier"] = ""
				preview = append(preview, doc)
			}
			return nil
		}
		switch typed := body.(type) {
		case map[string]any:
			if err := apply(typed); err != nil {
				nsutils.WriteJSON(w, http.StatusBadRequest, map[string]any{"status": http.StatusBadRequest, "message": err.Error()})
				return
			}
		case []any:
			for _, item := range typed {
				doc, ok := item.(map[string]any)
				if !ok {
					nsutils.WriteJSON(w, http.StatusBadRequest, map[string]any{"status": http.StatusBadRequest})
					return
				}
				if err := apply(doc); err != nil {
					nsutils.WriteJSON(w, http.StatusBadRequest, map[string]any{"status": http.StatusBadRequest, "message": err.Error()})
					return
				}
			}
		default:
			nsutils.WriteJSON(w, http.StatusBadRequest, map[string]any{"status": http.StatusBadRequest})
			return
		}

		if persist {
			nsutils.WriteRecordsV1(w, persisted, nil)
		} else {
			nsutils.WriteJSON(w, http.StatusCreated, preview)
		}
	}
}

// EntriesDelete deletes entries matching the supplied v1 query.
func EntriesDelete(dep nsutils.Dependencies) func(http.ResponseWriter, *http.Request, *nsutils.Identity) {
	return func(w http.ResponseWriter, r *http.Request, identity *nsutils.Identity) {
		query := nsutils.ParseV1(r.URL.Query(), "date")
		if spec := chi.URLParam(r, "spec"); spec != "" && spec != "json" {
			query.Filters = append(query.Filters, nsutils.Filter{Field: "type", Op: "eq", Value: spec})
		}
		deleted, err := dep.Store.DeleteMatching(r.Context(), "entries", query, true, identity.Name)
		if err != nil {
			nsutils.WriteJSON(w, http.StatusInternalServerError, map[string]any{"status": http.StatusInternalServerError, "message": err.Error()})
			return
		}
		nsutils.WriteJSON(w, http.StatusOK, map[string]any{"deleted": deleted})
	}
}
