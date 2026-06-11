package v3

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/nsutils"
)

// Version reports the application and API version metadata for API v3.
func Version(dep nsutils.Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		nsutils.WriteJSON(w, http.StatusOK, map[string]any{
			"status": http.StatusOK,
			"result": map[string]any{
				"version":    "3.0.4",
				"apiVersion": "3.0.4",
				"srvDate":    time.Now().UnixMilli(),
			},
		})
	}
}

// Status returns a lightweight authenticated health payload for API v3.
func Status(_ nsutils.Dependencies) func(http.ResponseWriter, *http.Request, *nsutils.Identity) {
	return func(w http.ResponseWriter, r *http.Request, identity *nsutils.Identity) {
		nsutils.WriteJSON(w, http.StatusOK, map[string]any{
			"status": http.StatusOK,
			"result": map[string]any{
				"srvDate": time.Now().UnixMilli(),
			},
		})
	}
}

// Test is a minimal authenticated endpoint used for API v3 connectivity checks.
func Test(_ nsutils.Dependencies) func(http.ResponseWriter, *http.Request, *nsutils.Identity) {
	return func(w http.ResponseWriter, r *http.Request, identity *nsutils.Identity) {
		nsutils.WriteJSON(w, http.StatusOK, map[string]any{"status": http.StatusOK, "result": "ok"})
	}
}

// LastModified returns the newest modification timestamp for each known
// collection.
func LastModified(dep nsutils.Dependencies) func(http.ResponseWriter, *http.Request, *nsutils.Identity) {
	return func(w http.ResponseWriter, r *http.Request, identity *nsutils.Identity) {
		collections := []nsutils.Collection{nsutils.Entries, nsutils.Treatments, nsutils.Devicestatus, nsutils.Profile, nsutils.Food, nsutils.Settings}
		lastModifiedMap := make(map[nsutils.Collection]int64)
		for _, collection := range collections {
			record, err := dep.Store.Search(r.Context(), collection, nsutils.Query{Limit: 1, SortDesc: true, SortField: "srv_modified"})
			if err != nil || record == nil || len(record) == 0 {
				continue
			}
			lastModifiedMap[collection] = record[0].SrvModified
		}

		nsutils.WriteJSON(w, http.StatusOK, map[string]any{
			"status": http.StatusOK,
			"result": map[string]any{
				"srvDate":     time.Now().UnixMilli(),
				"collections": lastModifiedMap,
			},
		})
	}
}

// Search implements collection search for API v3.
func Search(dep nsutils.Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		collection, err := nsutils.ParseCollection(chi.URLParam(r, "collection"))
		if err != nil {
			nsutils.WriteJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid collection"})
			return
		}
		permission := fmt.Sprintf("api:%s:read", collection)
		dep.Auth.Require(permission, false, func(w http.ResponseWriter, r *http.Request, identity *nsutils.Identity) {
			parsed, err := nsutils.ParseV3(r.URL.Query())
			if err != nil {
				nsutils.WriteJSON(w, http.StatusBadRequest, map[string]any{"status": http.StatusBadRequest, "message": err.Error()})
				return
			}

			records, err := dep.Store.Search(r.Context(), collection, parsed)
			if err != nil {
				nsutils.WriteJSON(w, http.StatusInternalServerError, map[string]any{"status": http.StatusInternalServerError, "message": err.Error()})
				return
			}
			result := make([]map[string]any, 0, len(records))
			for _, record := range records {
				result = append(result, selectV3Fields(record, parsed.Fields))
			}
			nsutils.WriteJSON(w, http.StatusOK, map[string]any{"status": http.StatusOK, "result": result})
		}).ServeHTTP(w, r)
	}
}

// Create creates a document in an API v3 collection.
func Create(dep nsutils.Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		collection, err := nsutils.ParseCollection(chi.URLParam(r, "collection"))
		if err != nil {
			nsutils.WriteJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid collection"})
			return
		}
		identity, err := dep.Auth.AuthenticateExplicit(r)
		if err != nil {
			nsutils.WriteJSON(w, http.StatusUnauthorized, map[string]any{"status": http.StatusUnauthorized, "message": authErrorMessage(err)})
			return
		}
		var doc map[string]any
		if err := nsutils.ReadJSON(r, &doc); err != nil || len(doc) == 0 {
			nsutils.WriteJSON(w, http.StatusBadRequest, map[string]any{"status": http.StatusBadRequest, "message": "Bad or missing request body"})
			return
		}
		doc, err = normalizeV3Document(collection, doc)
		if err != nil {
			nsutils.WriteJSON(w, http.StatusBadRequest, map[string]any{"status": http.StatusBadRequest, "message": err.Error()})
			return
		}

		permission := fmt.Sprintf("api:%s:create", collection)

		if !dep.Auth.HasPermission(*identity, permission) {
			nsutils.WriteJSON(w, http.StatusForbidden, map[string]any{"status": http.StatusForbidden, "message": fmt.Sprintf("Missing permission %s", permission)})
			return
		}
		record, created, err := dep.Store.Create(r.Context(), collection, doc, identity.Name)
		if err != nil {
			nsutils.WriteJSON(w, http.StatusBadRequest, map[string]any{"status": http.StatusBadRequest, "message": err.Error()})
			return
		}
		status := http.StatusCreated
		body := map[string]any{
			"status":       status,
			"identifier":   record.Identifier(),
			"lastModified": record.SrvModified,
		}
		if !created {
			status = http.StatusOK
			body["status"] = status
			body["isDeduplication"] = true
		}
		nsutils.LastModifiedHeader(w, record.SrvModified)
		w.Header().Set("Location", fmt.Sprintf("/api/v3/%s/%s", collection, record.Identifier()))
		nsutils.WriteJSON(w, status, body)
	}
}

// Read returns a single document from an API v3 collection.
func Read(dep nsutils.Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		collection, err := nsutils.ParseCollection(chi.URLParam(r, "collection"))
		if err != nil {
			nsutils.WriteJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid collection"})
			return
		}

		permission := fmt.Sprintf("api:%s:read", collection)
		dep.Auth.Require(permission, false, func(w http.ResponseWriter, r *http.Request, identity *nsutils.Identity) {
			record, err := dep.Store.Get(r.Context(), collection, chi.URLParam(r, "identifier"))

			if err != nil {
				status, body := nsutils.RequireRecord(err)
				nsutils.WriteJSON(w, status, body)
				return
			}
			if modifiedSince, ok := nsutils.ParseIfModifiedSince(r); ok && time.UnixMilli(record.SrvModified).Before(modifiedSince.Add(time.Second)) {
				w.WriteHeader(http.StatusNotModified)
				return
			}
			nsutils.LastModifiedHeader(w, record.SrvModified)
			parsed, _ := nsutils.ParseV3(r.URL.Query())
			fields := parsed.Fields
			nsutils.WriteJSON(w, http.StatusOK, map[string]any{"status": http.StatusOK, "result": selectV3Fields(record, fields)})
		}).ServeHTTP(w, r)
	}
}

// Update replaces a document in an API v3 collection and supports upsert semantics when allowed by the caller's permissions.
func Update(dep nsutils.Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		collection, err := nsutils.ParseCollection(chi.URLParam(r, "collection"))
		if err != nil {
			nsutils.WriteJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid collection"})
			return
		}
		permission := fmt.Sprintf("api:%s:update", collection)
		dep.Auth.Require(permission, false, func(w http.ResponseWriter, r *http.Request, identity *nsutils.Identity) {
			var doc map[string]any
			if err := nsutils.ReadJSON(r, &doc); err != nil {
				nsutils.WriteJSON(w, http.StatusBadRequest, map[string]any{"status": http.StatusBadRequest})
				return
			}
			identifier := chi.URLParam(r, "identifier")
			// TODO cleanup folling logic make it understandable
			existing, err := dep.Store.Get(r.Context(), collection, identifier)

			if err == nsutils.ErrGone {
				// Deleted Soft
				nsutils.WriteJSON(w, http.StatusGone, map[string]any{"status": http.StatusGone})
				return
			}
			if err != nil && err != nsutils.ErrNotFound {
				status, body := nsutils.RequireRecord(err)
				nsutils.WriteJSON(w, status, body)
				return
			}

			if err == nsutils.ErrNotFound && !dep.Auth.HasPermission(*identity, fmt.Sprintf("api:%s:create", collection)) {
				nsutils.WriteJSON(w, http.StatusForbidden, map[string]any{"status": http.StatusForbidden, "message": fmt.Sprintf("Missing permission api:%s:create", collection)})
				return
			}
			if err == nil {
				if unmodifiedSince, ok := nsutils.ParseIfUnmodifiedSince(r); ok && time.UnixMilli(existing.SrvModified).After(unmodifiedSince) {
					nsutils.WriteJSON(w, http.StatusPreconditionFailed, map[string]any{"status": http.StatusPreconditionFailed})
					return
				}
				if err := validateReplaceDoc(existing, doc); err != nil {
					nsutils.WriteJSON(w, http.StatusBadRequest, map[string]any{"status": http.StatusBadRequest, "message": err.Error()})
					return
				}
			} else {
				if err := validateCreateDoc(collection, doc); err != nil {
					nsutils.WriteJSON(w, http.StatusBadRequest, map[string]any{"status": http.StatusBadRequest, "message": err.Error()})
					return
				}
			}

			created := false
			var record nsutils.Record
			// TODO this is just checking if the record actually exists
			if existing.Identifier() == "" {
				// Double check permissions here
				if !dep.Auth.HasPermission(*identity, fmt.Sprintf("api:%s:create", collection)) {
					nsutils.WriteJSON(w, http.StatusForbidden, map[string]any{"status": http.StatusForbidden, "message": fmt.Sprintf("Missing permission api:%s:create", collection)})
					return
				}
				// Create if Existing is nil
				record, created, err = dep.Store.Create(r.Context(), collection, doc, identity.Name)
			} else {
				// Update otherwise
				record, err = dep.Store.Update(r.Context(), collection, identifier, doc, identity.Name)
			}

			if err != nil {
				nsutils.WriteJSON(w, http.StatusBadRequest, map[string]any{"status": http.StatusBadRequest, "message": err.Error()})
				return
			}

			nsutils.LastModifiedHeader(w, record.SrvModified)
			status := http.StatusOK
			body := map[string]any{"status": http.StatusOK, "result": selectV3Fields(record, nil)}
			if created {
				status = http.StatusCreated
				body["status"] = status
				body["identifier"] = record.Identifier()
				body["lastModified"] = record.SrvModified
			}
			nsutils.WriteJSON(w, status, body)
		}).ServeHTTP(w, r)
	}
}

// Patch applies a partial update to an API v3 document.
func Patch(dep nsutils.Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		collection, err := nsutils.ParseCollection(chi.URLParam(r, "collection"))
		if err != nil {
			nsutils.WriteJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid collection"})
			return
		}

		identifier := chi.URLParam(r, "identifier")

		permission := fmt.Sprintf("api:%s:update", collection)
		dep.Auth.Require(permission, false, func(w http.ResponseWriter, r *http.Request, identity *nsutils.Identity) {
			var doc map[string]any
			if err := nsutils.ReadJSON(r, &doc); err != nil {
				nsutils.WriteJSON(w, http.StatusBadRequest, map[string]any{"status": http.StatusBadRequest})
				return
			}
			existing, err := dep.Store.Get(r.Context(), collection, identifier)
			if err != nil {
				status, body := nsutils.RequireRecord(err)
				nsutils.WriteJSON(w, status, body)
				return
			}
			if unmodifiedSince, ok := nsutils.ParseIfUnmodifiedSince(r); ok && time.UnixMilli(existing.SrvModified).After(unmodifiedSince) {
				nsutils.WriteJSON(w, http.StatusPreconditionFailed, map[string]any{"status": http.StatusPreconditionFailed})
				return
			}
			if err := validatePatchDoc(existing, doc); err != nil {
				nsutils.WriteJSON(w, http.StatusBadRequest, map[string]any{"status": http.StatusBadRequest, "message": err.Error()})
				return
			}
			record, err := dep.Store.Update(r.Context(), collection, identifier, doc, identity.Name)

			if err != nil {
				status, body := nsutils.RequireRecord(err)
				nsutils.WriteJSON(w, status, body)
				return
			}
			nsutils.LastModifiedHeader(w, record.SrvModified)
			nsutils.WriteJSON(w, http.StatusOK, map[string]any{"status": http.StatusOK, "result": selectV3Fields(record, nil)})
		}).ServeHTTP(w, r)
	}
}

// Remove soft-deletes or permanently deletes an API v3 document.
func Remove(dep nsutils.Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		collection, err := nsutils.ParseCollection(chi.URLParam(r, "collection"))
		if err != nil {
			nsutils.WriteJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid collection"})
			return
		}
		permission := fmt.Sprintf("api:%s:delete", collection)
		dep.Auth.Require(permission, false, func(w http.ResponseWriter, r *http.Request, identity *nsutils.Identity) {
			permanent := r.URL.Query().Get("permanent") == "true"
			err := dep.Store.Delete(r.Context(), collection, chi.URLParam(r, "identifier"), permanent, identity.Name)
			if err != nil {
				status, body := nsutils.RequireRecord(err)
				nsutils.WriteJSON(w, status, body)
				return
			}
			nsutils.WriteJSON(w, http.StatusOK, map[string]any{"status": http.StatusOK})
		}).ServeHTTP(w, r)
	}
}

// History returns document revisions modified since the supplied timestamp.
func History(dep nsutils.Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		collection, err := nsutils.ParseCollection(chi.URLParam(r, "collection"))
		if err != nil {
			nsutils.WriteJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid collection"})
			return
		}
		permission := fmt.Sprintf("api:%s:read", collection)
		dep.Auth.Require(permission, false, func(w http.ResponseWriter, r *http.Request, identity *nsutils.Identity) {
			parsed, err := nsutils.ParseV3(r.URL.Query())

			since, err := strconv.ParseInt(chi.URLParam(r, "since"), 10, 64)
			if err != nil {
				nsutils.WriteJSON(w, http.StatusBadRequest, map[string]any{"status": http.StatusBadRequest})
				return
			}
			parsed.Filters = append(parsed.Filters, nsutils.Filter{Field: "srvModified", Op: "gte", Value: strconv.FormatInt(since, 10)})

			records, err := dep.Store.Search(r.Context(), collection, parsed)
			if err != nil {
				nsutils.WriteJSON(w, http.StatusInternalServerError, map[string]any{"status": http.StatusInternalServerError, "message": err.Error()})
				return
			}
			result := make([]map[string]any, 0, len(records))
			for _, record := range records {
				result = append(result, selectV3Fields(record, nil))
			}
			nsutils.WriteJSON(w, http.StatusOK, map[string]any{"status": http.StatusOK, "result": result})
		}).ServeHTTP(w, r)
	}
}

func validateCreateDoc(collection nsutils.Collection, doc map[string]any) error {

	if _, ok := nsutils.Int64Field(doc, "date"); !ok && collection != nsutils.Settings {
		return fmt.Errorf("Bad or missing date field")
	}
	if date, ok := nsutils.Int64Field(doc, "date"); ok && date <= 100000000000 {
		return fmt.Errorf("Bad or missing date field")
	}
	if utcOffset, exists := doc["utcOffset"]; exists {
		if value, err := toInt64Compat(utcOffset); err != nil || value < -1440 || value > 1440 {
			return fmt.Errorf("Bad or missing utcOffset field")
		}
	}
	if collection != nsutils.Settings {
		if app, ok := nsutils.StringField(doc, "app"); !ok || strings.TrimSpace(app) == "" {
			return fmt.Errorf("Bad or missing app field")
		}
	}
	return nil
}

func validateReplaceDoc(existing nsutils.Record, doc map[string]any) error {
	immutable := []string{"date", "utcOffset", "eventType", "device", "app", "srvCreated", "subject", "srvModified", "modifiedBy", "isValid"}
	for _, field := range immutable {
		if proposed, exists := doc[field]; exists {
			current := nsutils.PathValue(existing.ToMap(true), field)
			if field == "identifier" {
				continue
			}
			if fmt.Sprint(proposed) != fmt.Sprint(current) {
				return fmt.Errorf("Field %s cannot be modified by the client", field)
			}
		}
	}
	if identifier, ok := nsutils.StringField(doc, "identifier"); ok && identifier != "" && identifier != existing.Identifier() {
		delete(doc, "identifier")
	}
	return nil
}

func validatePatchDoc(existing nsutils.Record, doc map[string]any) error {
	immutable := []string{"identifier", "date", "utcOffset", "eventType", "device", "app", "srvCreated", "subject", "srvModified", "modifiedBy", "isValid"}
	for _, field := range immutable {
		if proposed, exists := doc[field]; exists {
			current := nsutils.PathValue(existing.ToMap(true), field)
			if fmt.Sprint(proposed) != fmt.Sprint(current) {
				return fmt.Errorf("Field %s cannot be modified by the client", field)
			}
		}
	}
	return nil
}

func normalizeV3Document(collection nsutils.Collection, doc map[string]any) (map[string]any, error) {
	clean, err := nsutils.NormalizeData(collection, doc)
	if err != nil {
		switch err.Error() {
		case "bad or missing date field":
			return nil, fmt.Errorf("Bad or missing date field")
		case "bad or missing utcOffset field":
			return nil, fmt.Errorf("Bad or missing utcOffset field")
		case "bad created_at field":
			return nil, fmt.Errorf("Bad created_at field")
		default:
			return nil, err
		}
	}

	if err := validateCreateDoc(collection, clean); err != nil {
		return nil, err
	}
	return clean, nil
}

func authErrorMessage(err error) string {
	switch {
	case err == nsutils.ErrBadToken, err == nsutils.ErrBadJWT:
		return "Bad access token or JWT"
	default:
		return "Missing or bad access token or JWT"
	}
}

func selectV3Fields(record nsutils.Record, fields []string) map[string]any {
	selected := nsutils.SelectFields(record, fields)
	delete(selected, "_id")
	return selected
}

func toInt64Compat(value any) (int64, error) {
	switch typed := value.(type) {
	case int:
		return int64(typed), nil
	case int64:
		return typed, nil
	case float64:
		return int64(typed), nil
	case json.Number:
		return typed.Int64()
	case string:
		return strconv.ParseInt(typed, 10, 64)
	default:
		return 0, fmt.Errorf("not int")
	}
}
