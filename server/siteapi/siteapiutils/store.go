package siteapiutils

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/server/components/site_templ"

	"nighttroop.t1d.nz/utils"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/httputils"
)

const errorHTMLXDivID = `#error`

//////
// CREATE
///////

// VerifyModel checks and normalizes a request model before any transaction is opened, so it is
// where network calls (e.g. testing credentials) belong. oldModel is nil when creating.
type VerifyInterface[T any] interface {
	VerifyModel(ctx context.Context, id *string, siteContext *httputils.SiteContext, oldModel *T, model *T) *HTTPError
}

type CreateHandlerInterface[T any] interface {
	VerifyInterface[T]
	CreateModel(ctx context.Context, tx *StoreTx, id *string, siteContext *httputils.SiteContext, model *T) (bool, *HTTPError)
	CreateHTMXResponse(model *T) templ.Component
	GetName() string
}

func CreateHandler[T any](controlDB *db.ControlDB, config *config.ServerConfig, store CreateHandlerInterface[T]) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		siteContext := httputils.GetSiteContext(r.Context())
		if siteContext == nil {
			slog.Error("site context is nil")
			ErrNotFound.Write(siteContext.AcceptType, w)
			return
		}
		if !HasPermission(store.GetName(), siteContext.IdentityModel) {
			ErrForbidden.Write(siteContext.AcceptType, w)
			return
		}

		requestModel, herr := extractModelFromRequest[T](siteContext, r)
		if herr != nil {
			herr.Write(siteContext.AcceptType, w)
			return
		}

		if herr := store.VerifyModel(r.Context(), nil, siteContext, nil, requestModel); herr != nil {
			herr.Write(siteContext.AcceptType, w)
			return
		}
		tx := NewStoreTx(r.Context(), siteContext)
		created, herr := store.CreateModel(r.Context(), tx, nil, siteContext, requestModel)
		if herr = tx.Finish(herr); herr != nil {
			herr.Write(siteContext.AcceptType, w)
			return
		}

		status := http.StatusOK
		if created {
			status = http.StatusCreated
		}
		w.WriteHeader(status)

		switch siteContext.AcceptType {
		case httputils.AcceptTypeHTMX:
			store.CreateHTMXResponse(requestModel).Render(r.Context(), w)
		case httputils.AcceptTypeHTML:
			store.CreateHTMXResponse(requestModel).Render(r.Context(), w)
		default:
			WriteModelJSON(w, status, requestModel)
		}
	}
}

///////
// UPDATE
///////

type UpdateHandlerInterface[T any] interface {
	VerifyInterface[T]
	UpdateModel(ctx context.Context, tx *StoreTx, id string, siteContext *httputils.SiteContext, oldModel *T, model *T) *HTTPError

	UpdateHTMXResponse(model *T) templ.Component

	GetModel(ctx context.Context, id string, siteContext *httputils.SiteContext) (*T, *HTTPError)
	GetName() string
}

func UpdateHandler[T any](controlDB *db.ControlDB, config *config.ServerConfig, store UpdateHandlerInterface[T]) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		siteContext := httputils.GetSiteContext(r.Context())
		if siteContext == nil {
			slog.Error("site context is nil")
			ErrNotFound.Write(siteContext.AcceptType, w)
			return
		}
		if !HasPermission(store.GetName(), siteContext.IdentityModel) {
			ErrForbidden.Write(siteContext.AcceptType, w)
			return
		}
		id := chi.URLParam(r, "id")

		oldModel, herr := store.GetModel(r.Context(), id, siteContext)
		if herr != nil {
			herr.Write(siteContext.AcceptType, w)
			return
		}

		requestModel, herr := extractModelFromRequest[T](siteContext, r)
		if herr != nil {
			herr.Write(siteContext.AcceptType, w)
			return
		}

		if herr := store.VerifyModel(r.Context(), &id, siteContext, oldModel, requestModel); herr != nil {
			herr.Write(siteContext.AcceptType, w)
			return
		}
		tx := NewStoreTx(r.Context(), siteContext)
		herr = store.UpdateModel(r.Context(), tx, id, siteContext, oldModel, requestModel)
		if herr = tx.Finish(herr); herr != nil {
			herr.Write(siteContext.AcceptType, w)
			return
		}

		status := http.StatusOK
		w.WriteHeader(status)

		switch siteContext.AcceptType {
		case httputils.AcceptTypeHTMX:
			store.UpdateHTMXResponse(requestModel).Render(r.Context(), w)
		case httputils.AcceptTypeHTML:
			store.UpdateHTMXResponse(requestModel).Render(r.Context(), w)
		default:
			WriteModelJSON(w, status, requestModel)
		}
	}
}

/////
// UPSERT HANDLER
/////

type UpsertInterface[T any] interface {
	VerifyInterface[T]
	CreateModel(ctx context.Context, tx *StoreTx, id *string, siteContext *httputils.SiteContext, model *T) (bool, *HTTPError)

	UpdateModel(ctx context.Context, tx *StoreTx, id string, siteContext *httputils.SiteContext, oldModel *T, model *T) *HTTPError

	UpdateHTMXResponse(model *T) templ.Component

	GetModel(ctx context.Context, id string, siteContext *httputils.SiteContext) (*T, *HTTPError)
	GetName() string
}

func UpsertHandler[T any](controlDB *db.ControlDB, config *config.ServerConfig, store UpsertInterface[T]) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		siteContext := httputils.GetSiteContext(r.Context())
		if siteContext == nil {
			slog.Error("site context is nil")
			ErrNotFound.Write(siteContext.AcceptType, w)
			return
		}
		if !HasPermission(store.GetName(), siteContext.IdentityModel) {
			ErrForbidden.Write(siteContext.AcceptType, w)
			return
		}
		id := chi.URLParam(r, "id")

		oldModel, herr := store.GetModel(r.Context(), id, siteContext)
		if herr != nil {
			herr.Write(siteContext.AcceptType, w)
			return
		}

		requestModel, herr := extractModelFromRequest[T](siteContext, r)
		if herr != nil {
			herr.Write(siteContext.AcceptType, w)
			return
		}

		if herr := store.VerifyModel(r.Context(), &id, siteContext, oldModel, requestModel); herr != nil {
			herr.Write(siteContext.AcceptType, w)
			return
		}

		tx := NewStoreTx(r.Context(), siteContext)
		if oldModel == nil {
			_, herr = store.CreateModel(r.Context(), tx, &id, siteContext, requestModel)
		} else {
			herr = store.UpdateModel(r.Context(), tx, id, siteContext, oldModel, requestModel)
		}
		if herr = tx.Finish(herr); herr != nil {
			herr.Write(siteContext.AcceptType, w)
			return
		}

		status := http.StatusOK
		w.WriteHeader(status)

		switch siteContext.AcceptType {
		case httputils.AcceptTypeHTMX:
			store.UpdateHTMXResponse(requestModel).Render(r.Context(), w)
		case httputils.AcceptTypeHTML:
			store.UpdateHTMXResponse(requestModel).Render(r.Context(), w)
		default:
			WriteModelJSON(w, status, requestModel)
		}
	}
}

/////
// READ
/////

type SearchInterface[T any] interface {
	SearchModels(ctx context.Context, query *Query, siteContext *httputils.SiteContext) ([]*T, *HTTPError)
	SearchHTMXResponse(models []*T) templ.Component
	GetName() string
}

func SearchHandler[T any](controlDB *db.ControlDB, config *config.ServerConfig, store SearchInterface[T]) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		siteContext := httputils.GetSiteContext(r.Context())
		if siteContext == nil {
			slog.Error("site context is nil")
			ErrNotFound.Write(siteContext.AcceptType, w)
			return
		}
		if !HasPermission(store.GetName(), siteContext.IdentityModel) {
			ErrForbidden.Write(siteContext.AcceptType, w)
			return
		}

		query := QueryFromParams(r.URL.Query())
		models, herr := store.SearchModels(r.Context(), query, siteContext)
		if herr != nil {
			herr.Write(siteContext.AcceptType, w)
			return
		}

		status := http.StatusOK
		w.WriteHeader(status)

		switch siteContext.AcceptType {
		case httputils.AcceptTypeHTMX:
			store.SearchHTMXResponse(models).Render(r.Context(), w)
		case httputils.AcceptTypeHTML:
			site_templ.FullPage(store.SearchHTMXResponse(models)).Render(r.Context(), w)
		default:
			WriteModelJSON(w, status, models)
		}
	}
}

// ////
// DELETE
// ////

type DeleteInterface[T any] interface {
	GetModel(ctx context.Context, id string, siteContext *httputils.SiteContext) (*T, *HTTPError)
	DeleteModel(ctx context.Context, tx *StoreTx, model *T, siteContext *httputils.SiteContext) *HTTPError
	DeleteHTMXResponse(model *T) templ.Component
	GetName() string
}

func DeleteHandler[T any](controlDB *db.ControlDB, config *config.ServerConfig, store DeleteInterface[T]) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		siteContext := httputils.GetSiteContext(r.Context())
		if siteContext == nil {
			slog.Error("site context is nil")
			ErrNotFound.Write(siteContext.AcceptType, w)
			return
		}
		if !HasPermission(store.GetName(), siteContext.IdentityModel) {
			ErrForbidden.Write(siteContext.AcceptType, w)
			return
		}

		id := chi.URLParam(r, "id")
		if id == "" {
			ErrBadRequest.Write(siteContext.AcceptType, w)
			return
		}

		model, err := store.GetModel(r.Context(), id, siteContext)
		if err != nil {
			ErrNotFound.Write(siteContext.AcceptType, w)
			return
		}

		tx := NewStoreTx(r.Context(), siteContext)
		herr := tx.Finish(store.DeleteModel(r.Context(), tx, model, siteContext))
		if herr != nil {
			herr.Write(siteContext.AcceptType, w)
			return
		}

		status := http.StatusOK
		w.WriteHeader(status)

		switch siteContext.AcceptType {
		case httputils.AcceptTypeHTMX:
			store.DeleteHTMXResponse(model).Render(r.Context(), w)
		case httputils.AcceptTypeHTML:
			// return to the page they came from
			http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
		default:
			w.Write([]byte(`{"status":"deleted"}`))
		}
	}
}

// ////
// UTILS
// ////
func extractModelFromRequest[T any](siteContext *httputils.SiteContext, r *http.Request) (*T, *HTTPError) {
	var doc map[string]any
	switch siteContext.ContentType {
	case httputils.ContentTypeJSON:
		if err := ReadJSON(r, &doc); err != nil || len(doc) == 0 {
			slog.Error("failed to read JSON body", "error", err)
			return nil, ErrBadBody

		}
	case httputils.ContentTypeForm:

		if err := r.ParseForm(); err != nil || len(r.Form) == 0 {
			slog.Debug("failed to parse form data", "error", err)
			return nil, ErrBadForm
		}
		doc = FormToMap[T](r.Form)
	default:
		return nil, ErrUnknownContent
	}

	model := new(T)
	err := utils.FromMap(doc, model)
	if err != nil {
		slog.Error("failed to convert doc to model", "error", err)
		return nil, ErrBadBody
	}
	return model, nil
}
