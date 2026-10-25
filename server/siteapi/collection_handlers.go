package siteapi

import (
	"log/slog"
	"net/http"

	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/server/components/site_templ"
	"nighttroop.t1d.nz/utils/apiutils"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/httputils"
)

func CreateHandler[R, M, Q any](controlDB *db.ControlDB, config *config.ServerConfig, store *Rows[R, M, Q]) http.HandlerFunc {
	creator := storeCan[RowCreator[R, M, Q]](store, "create")
	return func(w http.ResponseWriter, r *http.Request) {
		siteContext := SiteContextFor(w, r, store.GetName(), ActionCreate)
		if siteContext == nil {
			return
		}

		requestModel, herr := apiutils.ExtractModelFromRequest[M](siteContext.ContentType, r)
		if herr != nil {
			herr.Write(siteContext.AcceptType, w)
			return
		}

		tx := NewCollectionTx(r.Context(), siteContext)
		defer tx.Rollback()
		herr = store.create(r.Context(), tx, creator, nil, siteContext, requestModel)
		if herr = tx.Finish(herr); herr != nil {
			herr.Write(siteContext.AcceptType, w)
			return
		}

		WriteResponse(w, r, siteContext, http.StatusCreated, requestModel, creator.CreateHTMXResponse(requestModel), false)
	}
}

///////
// UPDATE
///////

func UpdateHandler[R, M, Q any](controlDB *db.ControlDB, config *config.ServerConfig, store *Rows[R, M, Q]) http.HandlerFunc {
	updater := storeCan[RowUpdater[R, M, Q]](store, "update")
	return func(w http.ResponseWriter, r *http.Request) {
		siteContext := SiteContextFor(w, r, store.GetName(), ActionUpdate)
		if siteContext == nil {
			return
		}
		id := chi.URLParam(r, "id")

		requestModel, herr := apiutils.ExtractModelFromRequest[M](siteContext.ContentType, r)
		if herr != nil {
			herr.Write(siteContext.AcceptType, w)
			return
		}

		tx := NewCollectionTx(r.Context(), siteContext)
		defer tx.Rollback()
		oldModel, _, herr := store.get(r.Context(), tx, id, siteContext)
		if herr == nil && oldModel == nil {
			herr = apiutils.ErrNotFound
		}
		if herr != nil {
			herr.Write(siteContext.AcceptType, w)
			return
		}

		herr = store.update(r.Context(), tx, updater, siteContext, oldModel, requestModel)
		if herr = tx.Finish(herr); herr != nil {
			herr.Write(siteContext.AcceptType, w)
			return
		}

		WriteResponse(w, r, siteContext, http.StatusOK, requestModel, updater.UpdateHTMXResponse(requestModel), false)
	}
}

/////
// UPSERT HANDLER
/////

// UpsertHandler creates the model with the URL's id, or updates it when there is one
func UpsertHandler[R, M, Q any](controlDB *db.ControlDB, config *config.ServerConfig, store *Rows[R, M, Q]) http.HandlerFunc {
	creator := storeCan[RowCreator[R, M, Q]](store, "create")
	updater := storeCan[RowUpdater[R, M, Q]](store, "update")
	return func(w http.ResponseWriter, r *http.Request) {
		siteContext := SiteContextFor(w, r, store.GetName(), ActionCreate, ActionUpdate)
		if siteContext == nil {
			return
		}
		id := chi.URLParam(r, "id")

		requestModel, herr := apiutils.ExtractModelFromRequest[M](siteContext.ContentType, r)
		if herr != nil {
			herr.Write(siteContext.AcceptType, w)
			return
		}

		tx := NewCollectionTx(r.Context(), siteContext)
		defer tx.Rollback()
		oldModel, _, herr := store.get(r.Context(), tx, id, siteContext)
		if herr != nil {
			herr.Write(siteContext.AcceptType, w)
			return
		}

		if oldModel == nil {
			herr = store.create(r.Context(), tx, creator, &id, siteContext, requestModel)
		} else {
			herr = store.update(r.Context(), tx, updater, siteContext, oldModel, requestModel)
		}
		if herr = tx.Finish(herr); herr != nil {
			herr.Write(siteContext.AcceptType, w)
			return
		}

		WriteResponse(w, r, siteContext, http.StatusOK, requestModel, updater.UpdateHTMXResponse(requestModel), false)
	}
}

/////
// READ
/////

func SearchHandler[R, M, Q any](controlDB *db.ControlDB, config *config.ServerConfig, store *Rows[R, M, Q]) http.HandlerFunc {
	searcher := storeCan[RowSearcher[R, M, Q]](store, "search")
	return func(w http.ResponseWriter, r *http.Request) {
		siteContext := SiteContextFor(w, r, store.GetName(), ActionSearch)
		if siteContext == nil {
			return
		}

		query, herr := apiutils.QueryFromParams(r.URL.Query())
		if herr != nil {
			herr.Write(siteContext.AcceptType, w)
			return
		}
		tx := NewCollectionTx(r.Context(), siteContext)
		defer tx.Rollback()
		models, herr := store.search(r.Context(), tx, searcher, query, siteContext)
		if herr = tx.Finish(herr); herr != nil {
			herr.Write(siteContext.AcceptType, w)
			return
		}

		WriteResponse(w, r, siteContext, http.StatusOK, models, searcher.SearchHTMXResponse(models, query), true)
	}
}

func GetHandler[R, M, Q any](controlDB *db.ControlDB, config *config.ServerConfig, store *Rows[R, M, Q]) http.HandlerFunc {
	getter := storeCan[RowGetter[M]](store, "get")
	return func(w http.ResponseWriter, r *http.Request) {
		siteContext := SiteContextFor(w, r, store.GetName(), ActionGet)
		if siteContext == nil {
			return
		}

		id := chi.URLParam(r, "id")
		tx := NewCollectionTx(r.Context(), siteContext)
		defer tx.Rollback()
		model, _, herr := store.get(r.Context(), tx, id, siteContext)
		if herr == nil && model == nil {
			herr = apiutils.ErrNotFound
		}
		if herr = tx.Finish(herr); herr != nil {
			herr.Write(siteContext.AcceptType, w)
			return
		}

		WriteResponse(w, r, siteContext, http.StatusOK, model, getter.GetHTMXResponse(model), true)
	}
}

// ////
// DELETE
// ////

func DeleteHandler[R, M, Q any](controlDB *db.ControlDB, config *config.ServerConfig, store *Rows[R, M, Q]) http.HandlerFunc {
	deleter := storeCan[RowDeleter[R, M, Q]](store, "delete")
	return func(w http.ResponseWriter, r *http.Request) {
		siteContext := SiteContextFor(w, r, store.GetName(), ActionDelete)
		if siteContext == nil {
			return
		}

		id := chi.URLParam(r, "id")
		if id == "" {
			apiutils.ErrBadRequest.Write(siteContext.AcceptType, w)
			return
		}

		tx := NewCollectionTx(r.Context(), siteContext)
		defer tx.Rollback()
		model, row, herr := store.get(r.Context(), tx, id, siteContext)
		if herr != nil || model == nil {
			apiutils.ErrNotFound.Write(siteContext.AcceptType, w)
			return
		}

		herr = tx.Finish(store.delete(r.Context(), tx, deleter, id, row, siteContext))
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
			deleter.DeleteHTMXResponse(model).Render(r.Context(), w)
		default:
			w.Write([]byte(`{"status":"deleted"}`))
		}
	}
}

// ////
// UTILS
// ////

// SiteContextFor is the request's site context when its identity can do every action on the store; otherwise it writes the error and is nil
func SiteContextFor(w http.ResponseWriter, r *http.Request, collectionName string, actions ...Action) *httputils.SiteContext {
	siteContext := httputils.GetSiteContext(r.Context())
	if siteContext == nil {
		slog.Error("site context is nil")
		apiutils.ErrNotFound.Write(httputils.GetAcceptType(r), w)
		return nil
	}
	for _, action := range actions {
		if !HasPermission(collectionName, action, siteContext.IdentityModel) {
			apiutils.ErrForbidden.Write(siteContext.AcceptType, w)
			return nil
		}
	}
	return siteContext
}

// WriteResponse sends HTMX the component, a browser the component (in the site's page when fullPage), and anything else the model as JSON
func WriteResponse(w http.ResponseWriter, r *http.Request, siteContext *httputils.SiteContext, status int, model any, component templ.Component, fullPage bool) {
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
