package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/di"
	"nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/httpx"
	"nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/model"
	"nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/query"

	"nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/util"
)

// EntriesCurrent returns the newest entry.
func EntriesCurrent(dep di.Dependencies) func(http.ResponseWriter, *http.Request, *di.Identity) {
	return func(w http.ResponseWriter, r *http.Request, identity *di.Identity) {
		query := di.DefaultQuery()
		query.Limit = 1
		records, err := dep.Store.Search(r.Context(), "entries", query)
		if err != nil {
			httpx.WriteJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
			return
		}
		util.WriteRecords(w, records, nil)
	}
}

// EntriesList lists entries using the Nightscout v1 query format.
func EntriesList(dep di.Dependencies) func(http.ResponseWriter, *http.Request, *di.Identity) {
	return func(w http.ResponseWriter, r *http.Request, identity *di.Identity) {
		query := query.ParseV1(r.URL.Query(), "date")
		records, err := dep.Store.Search(r.Context(), "entries", query)
		if err != nil {
			httpx.WriteJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
			return
		}
		util.WriteRecords(w, records, nil)
	}
}

// EntriesSpec resolves entry sub-routes such as current, sgv, mbg, or a
// specific identifier.
func EntriesSpec(dep di.Dependencies) func(http.ResponseWriter, *http.Request, *di.Identity) {
	return func(w http.ResponseWriter, r *http.Request, identity *di.Identity) {
		spec := chi.URLParam(r, "spec")
		switch spec {
		case "current":
			EntriesCurrent(dep)(w, r, identity)
			return
		case "sgv", "mbg":
			query := query.ParseV1(r.URL.Query(), "date")
			query.Filters = append(query.Filters, di.Filter{Field: "type", Op: "eq", Value: spec})
			records, err := dep.Store.Search(r.Context(), "entries", query)
			if err != nil {
				httpx.WriteJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
				return
			}
			util.WriteRecords(w, records, nil)
			return
		default:
			record, err := dep.Store.Get(r.Context(), "entries", spec)
			if err == nil && record == nil {
				httpx.WriteJSON(w, http.StatusInternalServerError, map[string]any{"status": http.StatusInternalServerError})
				return
			}
			if err != nil {
				status, body := httpx.RequireRecord(err)
				httpx.WriteJSON(w, status, body)
				return
			}
			util.WriteRecords(w, []model.Record{*record}, nil)
		}
	}
}

// EntriesCreate creates entries or returns a preview when persist is false.
func EntriesCreate(dep di.Dependencies, persist bool) func(http.ResponseWriter, *http.Request, *di.Identity) {
	return func(w http.ResponseWriter, r *http.Request, identity *di.Identity) {
		var body any
		if err := httpx.ReadJSON(r, &body); err != nil {
			httpx.WriteJSON(w, http.StatusBadRequest, map[string]any{"status": http.StatusBadRequest})
			return
		}
		preview := make([]map[string]any, 0)
		apply := func(doc map[string]any) error {
			doc = model.CloneMap(doc)
			if dateString, ok := model.StringField(doc, "dateString"); ok && dateString != "" {
				normalized, offset, err := model.ToUTCString(dateString)
				if err == nil {
					doc["dateString"] = normalized
					doc["utcOffset"] = offset
				}
			}
			if persist {
				record, _, err := dep.Store.Create(r.Context(), "entries", doc, identity.Name)
				if err == nil {
					preview = append(preview, record.ToMap(false))
				}
				return err
			}
			if _, ok := model.StringField(doc, "_id"); !ok {
				doc["_id"] = di.GenerateIdentifier()
			}
			if _, ok := model.StringField(doc, "identifier"); !ok {
				doc["identifier"] = doc["_id"]
			}
			preview = append(preview, doc)
			return nil
		}
		switch typed := body.(type) {
		case map[string]any:
			if err := apply(typed); err != nil {
				httpx.WriteJSON(w, http.StatusBadRequest, map[string]any{"status": http.StatusBadRequest, "message": err.Error()})
				return
			}
		case []any:
			for _, item := range typed {
				doc, ok := item.(map[string]any)
				if !ok {
					httpx.WriteJSON(w, http.StatusBadRequest, map[string]any{"status": http.StatusBadRequest})
					return
				}
				if err := apply(doc); err != nil {
					httpx.WriteJSON(w, http.StatusBadRequest, map[string]any{"status": http.StatusBadRequest, "message": err.Error()})
					return
				}
			}
		default:
			httpx.WriteJSON(w, http.StatusBadRequest, map[string]any{"status": http.StatusBadRequest})
			return
		}
		status := http.StatusOK
		if !persist {
			status = http.StatusCreated
		}
		httpx.WriteJSON(w, status, preview)
	}
}

// EntriesDelete deletes entries matching the supplied v1 query.
func EntriesDelete(dep di.Dependencies) func(http.ResponseWriter, *http.Request, *di.Identity) {
	return func(w http.ResponseWriter, r *http.Request, identity *di.Identity) {
		query := query.ParseV1(r.URL.Query(), "date")
		if spec := chi.URLParam(r, "spec"); spec != "" && spec != "json" {
			query.Filters = append(query.Filters, di.Filter{Field: "type", Op: "eq", Value: spec})
		}
		deleted, err := dep.Store.DeleteMatching(r.Context(), "entries", query, true, identity.Name)
		if err != nil {
			httpx.WriteJSON(w, http.StatusInternalServerError, map[string]any{"status": http.StatusInternalServerError, "message": err.Error()})
			return
		}
		httpx.WriteJSON(w, http.StatusOK, map[string]any{"deleted": deleted})
	}
}
