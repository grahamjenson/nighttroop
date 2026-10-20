package siteapiutils

import (
	"context"
	"log/slog"
	"net/http"

	"nighttroop.t1d.nz/utils/apiutils"

	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/server/components/site_templ"

	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/httputils"
)

//////
// CREATE
///////

type CreateHandlerInterface[T any] interface {
	CreateModel(ctx context.Context, tx *StoreTx, id *string, siteContext *httputils.SiteContext, model *T) (bool, *apiutils.HTTPError)
	CreateHTMXResponse(model *T) templ.Component
	GetName() string
}

func CreateHandler[T any](controlDB *db.ControlDB, config *config.ServerConfig, store CreateHandlerInterface[T]) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		siteContext := siteContextFor(w, r, store.GetName(), ActionCreate)
		if siteContext == nil {
			return
		}

		requestModel, herr := apiutils.ExtractModelFromRequest[T](siteContext.ContentType, r)
		if herr != nil {
			herr.Write(siteContext.AcceptType, w)
			return
		}

		tx := NewStoreTx(r.Context(), siteContext)
		defer tx.Rollback()
		created, herr := store.CreateModel(r.Context(), tx, nil, siteContext, requestModel)
		if herr = tx.Finish(herr); herr != nil {
			herr.Write(siteContext.AcceptType, w)
			return
		}

		status := http.StatusOK
		if created {
			status = http.StatusCreated
		}
		writeResponse(w, r, siteContext, status, requestModel, store.CreateHTMXResponse(requestModel), false)
	}
}

///////
// UPDATE
///////

type UpdateHandlerInterface[T any] interface {
	UpdateModel(ctx context.Context, tx *StoreTx, id string, siteContext *httputils.SiteContext, oldModel *T, model *T) *apiutils.HTTPError

	UpdateHTMXResponse(model *T) templ.Component

	GetModel(ctx context.Context, tx *StoreTx, id string, siteContext *httputils.SiteContext) (*T, *apiutils.HTTPError)
	GetName() string
}

func UpdateHandler[T any](controlDB *db.ControlDB, config *config.ServerConfig, store UpdateHandlerInterface[T]) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		siteContext := siteContextFor(w, r, store.GetName(), ActionUpdate)
		if siteContext == nil {
			return
		}
		id := chi.URLParam(r, "id")

		requestModel, herr := apiutils.ExtractModelFromRequest[T](siteContext.ContentType, r)
		if herr != nil {
			herr.Write(siteContext.AcceptType, w)
			return
		}

		tx := NewStoreTx(r.Context(), siteContext)
		defer tx.Rollback()
		oldModel, herr := store.GetModel(r.Context(), tx, id, siteContext)
		if herr == nil && oldModel == nil {
			herr = apiutils.ErrNotFound
		}
		if herr != nil {
			herr.Write(siteContext.AcceptType, w)
			return
		}

		herr = store.UpdateModel(r.Context(), tx, id, siteContext, oldModel, requestModel)
		if herr = tx.Finish(herr); herr != nil {
			herr.Write(siteContext.AcceptType, w)
			return
		}

		writeResponse(w, r, siteContext, http.StatusOK, requestModel, store.UpdateHTMXResponse(requestModel), false)
	}
}

/////
// UPSERT HANDLER
/////

type UpsertInterface[T any] interface {
	CreateModel(ctx context.Context, tx *StoreTx, id *string, siteContext *httputils.SiteContext, model *T) (bool, *apiutils.HTTPError)

	UpdateModel(ctx context.Context, tx *StoreTx, id string, siteContext *httputils.SiteContext, oldModel *T, model *T) *apiutils.HTTPError

	UpdateHTMXResponse(model *T) templ.Component

	GetModel(ctx context.Context, tx *StoreTx, id string, siteContext *httputils.SiteContext) (*T, *apiutils.HTTPError)
	GetName() string
}

func UpsertHandler[T any](controlDB *db.ControlDB, config *config.ServerConfig, store UpsertInterface[T]) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		siteContext := siteContextFor(w, r, store.GetName(), ActionCreate, ActionUpdate)
		if siteContext == nil {
			return
		}
		id := chi.URLParam(r, "id")

		requestModel, herr := apiutils.ExtractModelFromRequest[T](siteContext.ContentType, r)
		if herr != nil {
			herr.Write(siteContext.AcceptType, w)
			return
		}

		tx := NewStoreTx(r.Context(), siteContext)
		defer tx.Rollback()
		oldModel, herr := store.GetModel(r.Context(), tx, id, siteContext)
		if herr != nil {
			herr.Write(siteContext.AcceptType, w)
			return
		}

		if oldModel == nil {
			_, herr = store.CreateModel(r.Context(), tx, &id, siteContext, requestModel)
		} else {
			herr = store.UpdateModel(r.Context(), tx, id, siteContext, oldModel, requestModel)
		}
		if herr = tx.Finish(herr); herr != nil {
			herr.Write(siteContext.AcceptType, w)
			return
		}

		writeResponse(w, r, siteContext, http.StatusOK, requestModel, store.UpdateHTMXResponse(requestModel), false)
	}
}

/////
// READ
/////

type SearchInterface[T any] interface {
	SearchModels(ctx context.Context, tx *StoreTx, query *apiutils.Query, siteContext *httputils.SiteContext) ([]*T, *apiutils.HTTPError)
	SearchHTMXResponse(models []*T, query *apiutils.Query) templ.Component
	GetName() string
}

func SearchHandler[T any](controlDB *db.ControlDB, config *config.ServerConfig, store SearchInterface[T]) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		siteContext := siteContextFor(w, r, store.GetName(), ActionSearch)
		if siteContext == nil {
			return
		}

		query, herr := apiutils.QueryFromParams(r.URL.Query())
		if herr != nil {
			herr.Write(siteContext.AcceptType, w)
			return
		}
		tx := NewStoreTx(r.Context(), siteContext)
		defer tx.Rollback()
		models, herr := store.SearchModels(r.Context(), tx, query, siteContext)
		if herr = tx.Finish(herr); herr != nil {
			herr.Write(siteContext.AcceptType, w)
			return
		}

		writeResponse(w, r, siteContext, http.StatusOK, models, store.SearchHTMXResponse(models, query), true)
	}
}

type GetInterface[T any] interface {
	GetModel(ctx context.Context, tx *StoreTx, id string, siteContext *httputils.SiteContext) (*T, *apiutils.HTTPError)
	GetHTMXResponse(model *T) templ.Component
	GetName() string
}

func GetHandler[T any](controlDB *db.ControlDB, config *config.ServerConfig, store GetInterface[T]) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		siteContext := siteContextFor(w, r, store.GetName(), ActionGet)
		if siteContext == nil {
			return
		}

		id := chi.URLParam(r, "id")
		tx := NewStoreTx(r.Context(), siteContext)
		defer tx.Rollback()
		model, herr := store.GetModel(r.Context(), tx, id, siteContext)
		if herr == nil && model == nil {
			herr = apiutils.ErrNotFound
		}
		if herr = tx.Finish(herr); herr != nil {
			herr.Write(siteContext.AcceptType, w)
			return
		}

		writeResponse(w, r, siteContext, http.StatusOK, model, store.GetHTMXResponse(model), true)
	}
}

// ////
// DELETE
// ////

type DeleteInterface[T any] interface {
	GetModel(ctx context.Context, tx *StoreTx, id string, siteContext *httputils.SiteContext) (*T, *apiutils.HTTPError)
	DeleteModel(ctx context.Context, tx *StoreTx, model *T, siteContext *httputils.SiteContext) *apiutils.HTTPError
	DeleteHTMXResponse(model *T) templ.Component
	GetName() string
}

func DeleteHandler[T any](controlDB *db.ControlDB, config *config.ServerConfig, store DeleteInterface[T]) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		siteContext := siteContextFor(w, r, store.GetName(), ActionDelete)
		if siteContext == nil {
			return
		}

		id := chi.URLParam(r, "id")
		if id == "" {
			apiutils.ErrBadRequest.Write(siteContext.AcceptType, w)
			return
		}

		tx := NewStoreTx(r.Context(), siteContext)
		defer tx.Rollback()
		model, err := store.GetModel(r.Context(), tx, id, siteContext)
		if err != nil || model == nil {
			apiutils.ErrNotFound.Write(siteContext.AcceptType, w)
			return
		}

		herr := tx.Finish(store.DeleteModel(r.Context(), tx, model, siteContext))
		if herr != nil {
			herr.Write(siteContext.AcceptType, w)
			return
		}

		if siteContext.AcceptType == httputils.AcceptTypeHTML {
			// return to the page they came from
			http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
			return
		}

		status := http.StatusOK
		apiutils.WriteStatus(w, siteContext.AcceptType, status)

		switch siteContext.AcceptType {
		case httputils.AcceptTypeHTMX:
			store.DeleteHTMXResponse(model).Render(r.Context(), w)
		default:
			w.Write([]byte(`{"status":"deleted"}`))
		}
	}
}

// ////
// UTILS
// ////

// siteContextFor is the request's site context when its identity can do every action on the store; otherwise it writes the error and is nil
func siteContextFor(w http.ResponseWriter, r *http.Request, storeName string, actions ...Action) *httputils.SiteContext {
	siteContext := httputils.GetSiteContext(r.Context())
	if siteContext == nil {
		slog.Error("site context is nil")
		apiutils.ErrNotFound.Write(httputils.GetAcceptType(r), w)
		return nil
	}
	for _, action := range actions {
		if !HasPermission(storeName, action, siteContext.IdentityModel) {
			apiutils.ErrForbidden.Write(siteContext.AcceptType, w)
			return nil
		}
	}
	return siteContext
}

// writeResponse sends HTMX the component, a browser the component (in the site's page when fullPage), and anything else the model as JSON
func writeResponse(w http.ResponseWriter, r *http.Request, siteContext *httputils.SiteContext, status int, model any, component templ.Component, fullPage bool) {
	apiutils.WriteStatus(w, siteContext.AcceptType, status)
	switch siteContext.AcceptType {
	case httputils.AcceptTypeHTMX:
		component.Render(r.Context(), w)
	case httputils.AcceptTypeHTML:
		if fullPage {
			component = site_templ.FullPage(component)
		}
		component.Render(r.Context(), w)
	default:
		apiutils.WriteModelJSON(w, status, model)
	}
}
