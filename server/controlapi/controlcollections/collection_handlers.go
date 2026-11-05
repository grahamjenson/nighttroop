package controlcollections

import (
	"context"
	"database/sql"
	"errors"
	"net/http"

	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
	"nighttroop.t1d.nz/server/components/control_templ"
	"nighttroop.t1d.nz/utils/apiutils"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/httputils"
)

// The control plane's collections work like the site's, on control and jobs DB rows, for a signed in control identity

type CollectionBase[R, M any] interface {
	// GetName is what the collection is called, which its errors use
	GetName() string
	// GetRoute is where the collection is in its URL, e.g. job-stats for /job-stats
	GetRoute() string
	ToModel(row *R) *M
	GetRow(ctx context.Context, id string, controlContext *httputils.ControlContext) (*R, error)
}

type CollectionGet[R, M any] interface {
	CollectionBase[R, M]
	GetHTMXResponse(model *M) templ.Component
}

type CollectionSearch[R, M any] interface {
	CollectionBase[R, M]

	SearchRows(ctx context.Context, query *apiutils.Query, controlContext *httputils.ControlContext) ([]*R, error)
	SearchHTMXResponse(models []*M, query *apiutils.Query) templ.Component
}

// CollectionUpdate is a collection admins can change; read only control identities cannot
type CollectionUpdate[R, M any] interface {
	CollectionBase[R, M]

	UpdateRow(ctx context.Context, controlContext *httputils.ControlContext, oldModel *M, model *M) (*R, error)
	UpdateHTMXResponse(model *M) templ.Component
}

// //
// HANDLERS
// //

func SearchHandler[R, M any](collection CollectionSearch[R, M]) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		controlContext, acceptType := httputils.GetControlContext(r.Context()), httputils.GetAcceptType(r)

		query, herr := apiutils.QueryFromParams(r.URL.Query())
		if herr != nil {
			herr.Write(acceptType, w)
			return
		}
		if query == nil {
			query = &apiutils.Query{}
		}
		rows, err := collection.SearchRows(r.Context(), query, controlContext)
		if herr := apiutils.ErrorFor(err, collection.GetName()); herr != nil {
			herr.Write(acceptType, w)
			return
		}
		models := make([]*M, 0, len(rows))
		for _, row := range rows {
			models = append(models, collection.ToModel(row))
		}
		WriteResponse(w, r, acceptType, http.StatusOK, models, collection.SearchHTMXResponse(models, query))
	}
}

func GetHandler[R, M any](collection CollectionGet[R, M]) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		controlContext, acceptType := httputils.GetControlContext(r.Context()), httputils.GetAcceptType(r)

		model, herr := get(r.Context(), collection, chi.URLParam(r, "id"), controlContext)
		if herr == nil && model == nil {
			herr = apiutils.ErrNotFound
		}
		if herr != nil {
			herr.Write(acceptType, w)
			return
		}
		WriteResponse(w, r, acceptType, http.StatusOK, model, collection.GetHTMXResponse(model))
	}
}

func UpdateHandler[R, M any](collection CollectionUpdate[R, M]) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		controlContext, acceptType := httputils.GetControlContext(r.Context()), httputils.GetAcceptType(r)
		if controlContext.IdentityModel.Role != consts.ControlRoleAdmin {
			apiutils.ErrForbidden.Write(acceptType, w)
			return
		}

		requestModel, herr := apiutils.ExtractModelFromRequest[M](httputils.GetContentType(r), r)
		if herr != nil {
			herr.Write(acceptType, w)
			return
		}
		oldModel, herr := get(r.Context(), collection, chi.URLParam(r, "id"), controlContext)
		if herr == nil && oldModel == nil {
			herr = apiutils.ErrNotFound
		}
		if herr == nil {
			row, err := collection.UpdateRow(r.Context(), controlContext, oldModel, requestModel)
			if herr = apiutils.ErrorFor(err, collection.GetName()); herr == nil {
				*requestModel = *collection.ToModel(row)
			}
		}
		if herr != nil {
			herr.Write(acceptType, w)
			return
		}
		WriteResponse(w, r, acceptType, http.StatusOK, requestModel, collection.UpdateHTMXResponse(requestModel))
	}
}

// ////
// UTILS
// ////

// get is the model, or nil when there is none
func get[R, M any](ctx context.Context, collection CollectionBase[R, M], id string, controlContext *httputils.ControlContext) (*M, *apiutils.HTTPError) {
	row, err := collection.GetRow(ctx, id, controlContext)
	if errors.Is(err, sql.ErrNoRows) || (err == nil && row == nil) {
		return nil, nil
	}
	if herr := apiutils.ErrorFor(err, collection.GetName()); herr != nil {
		return nil, herr
	}
	return collection.ToModel(row), nil
}

// WriteResponse sends HTMX the component, a browser the component in a full page, and anything else the model as JSON
func WriteResponse(w http.ResponseWriter, r *http.Request, acceptType httputils.AcceptType, status int, model any, component templ.Component) {
	apiutils.WriteStatus(w, acceptType, status)
	switch acceptType {
	case httputils.AcceptTypeHTMX:
		component.Render(r.Context(), w)
	case httputils.AcceptTypeHTML:
		control_templ.FullPage(component).Render(r.Context(), w)
	default:
		apiutils.WriteModelJSON(w, status, model)
	}
}
