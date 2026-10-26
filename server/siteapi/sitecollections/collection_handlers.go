package sitecollections

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"

	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
	"nighttroop.t1d.nz/data/events/sitemodels"
	"nighttroop.t1d.nz/server/components/site_templ"
	"nighttroop.t1d.nz/server/siteapi/siteapiutils"

	"nighttroop.t1d.nz/utils/apiutils"
	"nighttroop.t1d.nz/utils/httputils"
)

type CollectionBase[R, M any] interface {
	// GetName is what the collection is called, which its permissions and errors use
	GetName() string
	// GetRoute is where the collection is in its URL, e.g. sync for /settings/sync
	GetRoute() string
	ToModel(row *R, siteContext *httputils.SiteContext) *M
	GetRow(ctx context.Context, collectionTx *siteapiutils.CollectionTx, id string, siteContext *httputils.SiteContext) (*R, error)
}

type CollectionCreate[R, M any] interface {
	CollectionBase[R, M]

	CreateRow(ctx context.Context, collectionTx *siteapiutils.CollectionTx, id *string, siteContext *httputils.SiteContext, model *M) (*R, error)
	CreateHTMXResponse(model *M) templ.Component
}

type CollectionUpdate[R, M any] interface {
	CollectionBase[R, M]

	UpdateRow(ctx context.Context, collectionTx *siteapiutils.CollectionTx, siteContext *httputils.SiteContext, oldModel *M, model *M) (*R, error)
	UpdateHTMXResponse(model *M) templ.Component
}

// CollectionUpsert is a collection whose update creates what is not there yet, as each type of sync source has one
type CollectionUpsert[R, M any] interface {
	CollectionCreate[R, M]
	CollectionUpdate[R, M]
	Upserts()
}

type CollectionGet[R, M any] interface {
	CollectionBase[R, M]
	GetHTMXResponse(model *M) templ.Component
}

type CollectionSearch[R, M any] interface {
	CollectionBase[R, M]

	SearchRows(ctx context.Context, collectionTx *siteapiutils.CollectionTx, query *apiutils.Query, siteContext *httputils.SiteContext) ([]*R, error)
	SearchHTMXResponse(models []*M, query *apiutils.Query) templ.Component
}

type CollectionDelete[R, M any] interface {
	CollectionBase[R, M]

	DeleteRow(ctx context.Context, collectionTx *siteapiutils.CollectionTx, id string, siteContext *httputils.SiteContext) error
	DeleteHTMXResponse(model *M) templ.Component
}

// //
// HANDLERS
// //
func CreateHandler[R, M any](collection CollectionCreate[R, M]) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		siteContext := SiteContextFor(w, r, collection.GetName(), siteapiutils.ActionCreate)
		if siteContext == nil {
			return
		}

		requestModel, herr := apiutils.ExtractModelFromRequest[M](siteContext.ContentType, r)
		if herr != nil {
			writeError(w, r, siteContext.AcceptType, herr)
			return
		}

		tx := siteapiutils.NewCollectionTx(r.Context(), siteContext)
		defer tx.Rollback()
		herr = create(r.Context(), tx, collection, nil, siteContext, requestModel)
		if herr = tx.Finish(herr); herr != nil {
			writeError(w, r, siteContext.AcceptType, herr)
			return
		}

		WriteResponse(w, r, siteContext, http.StatusCreated, requestModel, collection.CreateHTMXResponse(requestModel), false)
	}
}

///////
// UPDATE
///////

func UpdateHandler[R, M any](collection CollectionUpdate[R, M]) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		siteContext := SiteContextFor(w, r, collection.GetName(), siteapiutils.ActionUpdate)
		if siteContext == nil {
			return
		}
		id := chi.URLParam(r, "id")

		requestModel, herr := apiutils.ExtractModelFromRequest[M](siteContext.ContentType, r)
		if herr != nil {
			writeError(w, r, siteContext.AcceptType, herr)
			return
		}

		tx := siteapiutils.NewCollectionTx(r.Context(), siteContext)
		defer tx.Rollback()
		oldModel, _, herr := get(r.Context(), tx, collection, id, siteContext)
		if herr == nil && oldModel == nil {
			herr = apiutils.ErrNotFound
		}
		if herr != nil {
			writeError(w, r, siteContext.AcceptType, herr)
			return
		}

		herr = update(r.Context(), tx, collection, siteContext, oldModel, requestModel)
		if herr = tx.Finish(herr); herr != nil {
			writeError(w, r, siteContext.AcceptType, herr)
			return
		}

		WriteResponse(w, r, siteContext, http.StatusOK, requestModel, collection.UpdateHTMXResponse(requestModel), false)
	}
}

/////
// UPSERT HANDLER
/////

// UpsertHandler creates the model with the URL's id, or updates it when there is one

func UpsertHandler[R, M any](collection CollectionUpsert[R, M]) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		siteContext := SiteContextFor(w, r, collection.GetName(), siteapiutils.ActionCreate, siteapiutils.ActionUpdate)
		if siteContext == nil {
			return
		}
		id := chi.URLParam(r, "id")

		requestModel, herr := apiutils.ExtractModelFromRequest[M](siteContext.ContentType, r)
		if herr != nil {
			writeError(w, r, siteContext.AcceptType, herr)
			return
		}

		tx := siteapiutils.NewCollectionTx(r.Context(), siteContext)
		defer tx.Rollback()
		oldModel, _, herr := get(r.Context(), tx, collection, id, siteContext)
		if herr != nil {
			writeError(w, r, siteContext.AcceptType, herr)
			return
		}

		if oldModel == nil {
			herr = create(r.Context(), tx, collection, &id, siteContext, requestModel)
		} else {
			herr = update(r.Context(), tx, collection, siteContext, oldModel, requestModel)
		}
		if herr = tx.Finish(herr); herr != nil {
			writeError(w, r, siteContext.AcceptType, herr)
			return
		}

		WriteResponse(w, r, siteContext, http.StatusOK, requestModel, collection.UpdateHTMXResponse(requestModel), false)
	}
}

/////
// READ
/////

func SearchHandler[R, M any](collection CollectionSearch[R, M]) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		siteContext := SiteContextFor(w, r, collection.GetName(), siteapiutils.ActionSearch)
		if siteContext == nil {
			return
		}

		query, herr := apiutils.QueryFromParams(r.URL.Query())
		if herr != nil {
			writeError(w, r, siteContext.AcceptType, herr)
			return
		}
		if query == nil {
			query = &apiutils.Query{}
		}
		tx := siteapiutils.NewCollectionTx(r.Context(), siteContext)
		defer tx.Rollback()
		models, herr := search(r.Context(), tx, collection, query, siteContext)
		if herr = tx.Finish(herr); herr != nil {
			writeError(w, r, siteContext.AcceptType, herr)
			return
		}

		WriteResponse(w, r, siteContext, http.StatusOK, models, collection.SearchHTMXResponse(models, query), true)
	}
}

func GetHandler[R, M any](collection CollectionGet[R, M]) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		siteContext := SiteContextFor(w, r, collection.GetName(), siteapiutils.ActionGet)
		if siteContext == nil {
			return
		}

		id := chi.URLParam(r, "id")
		tx := siteapiutils.NewCollectionTx(r.Context(), siteContext)
		defer tx.Rollback()
		model, _, herr := get(r.Context(), tx, collection, id, siteContext)
		if herr == nil && model == nil {
			herr = apiutils.ErrNotFound
		}
		if herr = tx.Finish(herr); herr != nil {
			writeError(w, r, siteContext.AcceptType, herr)
			return
		}

		WriteResponse(w, r, siteContext, http.StatusOK, model, collection.GetHTMXResponse(model), true)
	}
}

// ////
// DELETE
// ////

func DeleteHandler[R, M any](collection CollectionDelete[R, M]) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		siteContext := SiteContextFor(w, r, collection.GetName(), siteapiutils.ActionDelete)
		if siteContext == nil {
			return
		}

		id := chi.URLParam(r, "id")
		if id == "" {
			writeError(w, r, siteContext.AcceptType, apiutils.ErrBadRequest)
			return
		}

		tx := siteapiutils.NewCollectionTx(r.Context(), siteContext)
		defer tx.Rollback()
		model, row, herr := get(r.Context(), tx, collection, id, siteContext)
		if herr != nil || model == nil {
			writeError(w, r, siteContext.AcceptType, apiutils.ErrNotFound)
			return
		}

		herr = tx.Finish(delete(r.Context(), siteContext, tx, collection, id, row))
		if herr != nil {
			writeError(w, r, siteContext.AcceptType, herr)
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
			collection.DeleteHTMXResponse(model).Render(r.Context(), w)
		default:
			w.Write([]byte(`{"status":"deleted"}`))
		}
	}
}

// ////
// UTILS
// ////

// SiteContextFor is the request's site context when its identity can do every action on the store; otherwise it writes the error and is nil
func SiteContextFor(w http.ResponseWriter, r *http.Request, collectionName string, actions ...siteapiutils.Action) *httputils.SiteContext {
	siteContext := httputils.GetSiteContext(r.Context())
	if siteContext == nil {
		slog.Error("site context is nil")
		writeError(w, r, httputils.GetAcceptType(r), apiutils.ErrNotFound)
		return nil
	}
	for _, action := range actions {
		if !siteapiutils.HasPermission(collectionName, siteapiutils.Action(action), siteContext.IdentityModel) {
			writeError(w, r, siteContext.AcceptType, apiutils.ErrForbidden)
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

// writeError sends HTMX the error as a toast, which forms add to #toaster, and anything else the error as it writes itself
func writeError(w http.ResponseWriter, r *http.Request, acceptType httputils.AcceptType, herr *apiutils.HTTPError) {
	if acceptType != httputils.AcceptTypeHTMX {
		herr.Write(acceptType, w)
		return
	}
	apiutils.WriteStatus(w, acceptType, herr.Status)
	site_templ.SiteError(herr.Message).Render(r.Context(), w)
}

// get is the model, or nil when there is none
func get[R, M any](ctx context.Context, tx *siteapiutils.CollectionTx, collection CollectionBase[R, M], id string, siteContext *httputils.SiteContext) (*M, *R, *apiutils.HTTPError) {

	row, err := collection.GetRow(ctx, tx, id, siteContext)
	if errors.Is(err, sql.ErrNoRows) || (err == nil && row == nil) {
		return nil, nil, nil
	}
	if herr := answer(err, collection.GetName()); herr != nil {
		return nil, nil, herr
	}
	return collection.ToModel(row, siteContext), row, nil
}

func search[R, M any](ctx context.Context, tx *siteapiutils.CollectionTx, collection CollectionSearch[R, M], query *apiutils.Query, siteContext *httputils.SiteContext) ([]*M, *apiutils.HTTPError) {
	rows, err := collection.SearchRows(ctx, tx, query, siteContext)
	if herr := answer(err, collection.GetName()); herr != nil {
		return nil, herr
	}
	models := make([]*M, 0, len(rows))
	for _, row := range rows {
		models = append(models, collection.ToModel(row, siteContext))
	}
	return models, nil
}

func create[R, M any](ctx context.Context, tx *siteapiutils.CollectionTx, collection CollectionCreate[R, M], id *string, siteContext *httputils.SiteContext, model *M) *apiutils.HTTPError {

	row, err := collection.CreateRow(ctx, tx, id, siteContext, model)
	if herr := answer(err, collection.GetName()); herr != nil {
		return herr
	}
	reportChange(tx, sitemodels.Created, row)
	*model = *collection.ToModel(row, siteContext)
	return nil
}

func update[R, M any](ctx context.Context, tx *siteapiutils.CollectionTx, collection CollectionUpdate[R, M], siteContext *httputils.SiteContext, oldModel *M, model *M) *apiutils.HTTPError {

	row, err := collection.UpdateRow(ctx, tx, siteContext, oldModel, model)
	if herr := answer(err, collection.GetName()); herr != nil {
		return herr
	}
	reportChange(tx, sitemodels.Updated, row)
	*model = *collection.ToModel(row, siteContext)
	return nil
}

// delete deletes the row read as the model, so the change can say what was deleted
func delete[R, M any](ctx context.Context, siteContext *httputils.SiteContext, tx *siteapiutils.CollectionTx, collection CollectionDelete[R, M], id string, row *R) *apiutils.HTTPError {

	if herr := answer(collection.DeleteRow(ctx, tx, id, siteContext), collection.GetName()); herr != nil {
		return herr
	}
	reportChange(tx, sitemodels.Deleted, row)
	return nil
}

// reportChange reports a change to the site's own rows, to be published once the transaction commits
func reportChange[R any](tx *siteapiutils.CollectionTx, action sitemodels.Action, row *R) {
	// TODO filter for the relevant site or collection before reporting the change
	tx.ReportChange(action, row)
}

// answer is the response to a store's error: its own, not found for a missing row, or the database's
func answer(err error, what string) *apiutils.HTTPError {
	var herr *apiutils.HTTPError
	switch {
	case err == nil:
		return nil
	case errors.As(err, &herr):
		return herr
	case errors.Is(err, sql.ErrNoRows):
		return apiutils.ErrNotFound
	}
	return apiutils.WriteError(err, what)
}
