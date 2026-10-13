package siteapi

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
	"nighttroop.t1d.nz/data/db"

	"nighttroop.t1d.nz/utils"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/httputils"
)

//////
// CREATE
///////

type CreateHandlerInterface[T any] interface {
	CreateModel(ctx context.Context, siteContext *httputils.SiteContext, model *T) (bool, *HTTPError)
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

		var doc map[string]any

		switch siteContext.ContentType {
		case httputils.ContentTypeJSON:
			if err := ReadJSON(r, &doc); err != nil || len(doc) == 0 {
				slog.Error("failed to read JSON body", "error", err)
				ErrBadBody.Write(siteContext.AcceptType, w)
				return
			}
		case httputils.ContentTypeForm:

			if err := r.ParseForm(); err != nil || len(r.Form) == 0 {
				slog.Debug("failed to parse form data", "error", err)
				ErrBadForm.Write(siteContext.AcceptType, w)
				return
			}
			doc = FormToMap(r.Form)
		default:
			ErrUnknownContent.Write(siteContext.AcceptType, w)
			return
		}

		var model *T
		err := utils.FromMap(doc, model)
		if err != nil {
			slog.Error("failed to convert doc to model", "error", err)
			ErrBadBody.Write(siteContext.AcceptType, w)
			return
		}

		created, herr := store.CreateModel(r.Context(), siteContext, model)
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
			store.CreateHTMXResponse(model).Render(r.Context(), w)
		case httputils.AcceptTypeHTML:
			store.CreateHTMXResponse(model).Render(r.Context(), w)
		default:
			WriteModelJSON(w, status, model)
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

		var doc map[string]any

		fmt.Println("Updating model with ID:", id, "siteContext:", siteContext, "ContentType:", siteContext.ContentType, "model:", oldModel)

		switch siteContext.ContentType {
		case httputils.ContentTypeJSON:
			if err := ReadJSON(r, &doc); err != nil || len(doc) == 0 {
				slog.Error("failed to read JSON body", "error", err)
				ErrBadBody.Write(siteContext.AcceptType, w)
				return
			}
		case httputils.ContentTypeForm:

			if err := r.ParseForm(); err != nil || len(r.Form) == 0 {
				slog.Debug("failed to parse form data", "error", err)
				ErrBadForm.Write(siteContext.AcceptType, w)
				return
			}
			doc = FormToMap(r.Form)
		default:
			ErrUnknownContent.Write(siteContext.AcceptType, w)
			return
		}

		var model *T
		err := utils.FromMap(doc, model)
		if err != nil {
			slog.Error("failed to convert doc to model", "error", err)
			ErrBadBody.Write(siteContext.AcceptType, w)
			return
		}

		herr = store.UpdateModel(r.Context(), id, siteContext, oldModel, model)
		if herr != nil {
			herr.Write(siteContext.AcceptType, w)
			return
		}

		status := http.StatusOK
		w.WriteHeader(status)

		switch siteContext.AcceptType {
		case httputils.AcceptTypeHTMX:
			store.UpdateHTMXResponse(model).Render(r.Context(), w)
		case httputils.AcceptTypeHTML:
			store.UpdateHTMXResponse(model).Render(r.Context(), w)
		default:
			WriteModelJSON(w, status, model)
		}
	}
}

/////
// CREATE UPDATE HANDLER
/////

type CreateUpdateHandlerInterface[T any] interface {
	UpdateModel(ctx context.Context, id string, siteContext *httputils.SiteContext, oldModel *T, model *T) *HTTPError

	UpdateHTMXResponse(model *T) templ.Component

	GetModel(ctx context.Context, id string, siteContext *httputils.SiteContext) (*T, *HTTPError)
	GetName() string
}

func CreateUpdateHandler[T any](controlDB *db.ControlDB, config *config.ServerConfig, store UpdateHandlerInterface[T]) http.HandlerFunc {
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
			fmt.Println("FASHUOHFASOUF")
			herr.Write(siteContext.AcceptType, w)
			return
		}

		var doc map[string]any

		fmt.Println("Updating model with ID:", id, "siteContext:", siteContext, "ContentType:", siteContext.ContentType, "model:", oldModel)

		switch siteContext.ContentType {
		case httputils.ContentTypeJSON:
			if err := ReadJSON(r, &doc); err != nil || len(doc) == 0 {
				slog.Error("failed to read JSON body", "error", err)
				ErrBadBody.Write(siteContext.AcceptType, w)
				return
			}
		case httputils.ContentTypeForm:

			if err := r.ParseForm(); err != nil || len(r.Form) == 0 {
				slog.Debug("failed to parse form data", "error", err)
				ErrBadForm.Write(siteContext.AcceptType, w)
				return
			}
			doc = FormToMap(r.Form)
		default:
			ErrUnknownContent.Write(siteContext.AcceptType, w)
			return
		}

		var model *T
		err := utils.FromMap(doc, model)
		if err != nil {
			slog.Error("failed to convert doc to model", "error", err)
			ErrBadBody.Write(siteContext.AcceptType, w)
			return
		}

		herr = store.UpdateModel(r.Context(), id, siteContext, oldModel, model)
		if herr != nil {
			herr.Write(siteContext.AcceptType, w)
			return
		}

		status := http.StatusOK
		w.WriteHeader(status)

		switch siteContext.AcceptType {
		case httputils.AcceptTypeHTMX:
			store.UpdateHTMXResponse(model).Render(r.Context(), w)
		case httputils.AcceptTypeHTML:
			store.UpdateHTMXResponse(model).Render(r.Context(), w)
		default:
			WriteModelJSON(w, status, model)
		}
	}
}
