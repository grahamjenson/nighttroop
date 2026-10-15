package siteapi

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
	"nighttroop.t1d.nz/data/db"

	"nighttroop.t1d.nz/utils"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/httputils"
)

const errorHTMLXDivID = `#error`

//////
// CREATE
///////

type CreateHandlerInterface[T any] interface {
	CreateModel(ctx context.Context, id *string, siteContext *httputils.SiteContext, model *T) (bool, *HTTPError)
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

		created, herr := store.CreateModel(r.Context(), nil, siteContext, requestModel)
		if herr != nil {
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
	UpdateModel(ctx context.Context, id string, siteContext *httputils.SiteContext, oldModel *T, model *T) *HTTPError

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

		herr = store.UpdateModel(r.Context(), id, siteContext, oldModel, requestModel)
		if herr != nil {
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
// CREATE UPDATE HANDLER
/////

type CreateUpdateHandlerInterface[T any] interface {
	CreateModel(ctx context.Context, id *string, siteContext *httputils.SiteContext, model *T) (bool, *HTTPError)

	UpdateModel(ctx context.Context, id string, siteContext *httputils.SiteContext, oldModel *T, model *T) *HTTPError

	UpdateHTMXResponse(model *T) templ.Component

	GetModel(ctx context.Context, id string, siteContext *httputils.SiteContext) (*T, *HTTPError)
	GetName() string
}

func CreateUpdateHandler[T any](controlDB *db.ControlDB, config *config.ServerConfig, store CreateUpdateHandlerInterface[T]) http.HandlerFunc {
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

		if oldModel == nil {
			_, herr = store.CreateModel(r.Context(), &id, siteContext, requestModel)
			if herr != nil {
				herr.Write(siteContext.AcceptType, w)
				return
			}
		} else {

			herr = store.UpdateModel(r.Context(), id, siteContext, oldModel, requestModel)
			if herr != nil {
				herr.Write(siteContext.AcceptType, w)
				return
			}
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
