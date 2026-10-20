package controlapiutils

import (
	"context"
	"net/http"

	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/server/components/control_templ"
	"nighttroop.t1d.nz/server/siteapi/siteapiutils"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/httputils"
)

// The control API's stores work like the site API's, against the control DB, for a signed in control identity

/////
// READ
/////

type SearchInterface[T any] interface {
	SearchModels(ctx context.Context, query *siteapiutils.Query, controlContext *httputils.ControlContext) ([]*T, *siteapiutils.HTTPError)
	SearchHTMXResponse(models []*T, query *siteapiutils.Query) templ.Component
}

func SearchHandler[T any](controlDB *db.ControlDB, config *config.ServerConfig, store SearchInterface[T]) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		controlContext, acceptType := httputils.GetControlContext(r.Context()), httputils.GetAcceptType(r)

		query, herr := siteapiutils.QueryFromParams(r.URL.Query())
		if herr == nil {
			var models []*T
			if models, herr = store.SearchModels(r.Context(), query, controlContext); herr == nil {
				writeResponse(w, r, acceptType, http.StatusOK, models, store.SearchHTMXResponse(models, query))
				return
			}
		}
		herr.Write(acceptType, w)
	}
}

type GetInterface[T any] interface {
	GetModel(ctx context.Context, id string, controlContext *httputils.ControlContext) (*T, *siteapiutils.HTTPError)
	GetHTMXResponse(model *T) templ.Component
}

func GetHandler[T any](controlDB *db.ControlDB, config *config.ServerConfig, store GetInterface[T]) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		controlContext, acceptType := httputils.GetControlContext(r.Context()), httputils.GetAcceptType(r)

		model, herr := getModel(r, store, controlContext)
		if herr != nil {
			herr.Write(acceptType, w)
			return
		}
		writeResponse(w, r, acceptType, http.StatusOK, model, store.GetHTMXResponse(model))
	}
}

///////
// UPDATE
///////

type UpdateInterface[T any] interface {
	GetModel(ctx context.Context, id string, controlContext *httputils.ControlContext) (*T, *siteapiutils.HTTPError)
	UpdateModel(ctx context.Context, controlContext *httputils.ControlContext, oldModel *T, model *T) *siteapiutils.HTTPError
	UpdateHTMXResponse(model *T) templ.Component
}

func UpdateHandler[T any](controlDB *db.ControlDB, config *config.ServerConfig, store UpdateInterface[T]) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		controlContext, acceptType := httputils.GetControlContext(r.Context()), httputils.GetAcceptType(r)

		requestModel, herr := siteapiutils.ExtractModelFromRequest[T](httputils.GetContentType(r), r)
		if herr != nil {
			herr.Write(acceptType, w)
			return
		}
		oldModel, herr := getModel(r, store, controlContext)
		if herr == nil {
			herr = store.UpdateModel(r.Context(), controlContext, oldModel, requestModel)
		}
		if herr != nil {
			herr.Write(acceptType, w)
			return
		}
		writeResponse(w, r, acceptType, http.StatusOK, requestModel, store.UpdateHTMXResponse(requestModel))
	}
}

// ////
// UTILS
// ////

func getModel[T any](r *http.Request, store interface {
	GetModel(ctx context.Context, id string, controlContext *httputils.ControlContext) (*T, *siteapiutils.HTTPError)
}, controlContext *httputils.ControlContext) (*T, *siteapiutils.HTTPError) {
	model, herr := store.GetModel(r.Context(), chi.URLParam(r, "id"), controlContext)
	if herr == nil && model == nil {
		herr = siteapiutils.ErrNotFound
	}
	return model, herr
}

// writeResponse sends HTMX the component, a browser the component in a full page, and anything else the model as JSON
func writeResponse(w http.ResponseWriter, r *http.Request, acceptType httputils.AcceptType, status int, model any, component templ.Component) {
	siteapiutils.WriteStatus(w, acceptType, status)
	switch acceptType {
	case httputils.AcceptTypeHTMX:
		component.Render(r.Context(), w)
	case httputils.AcceptTypeHTML:
		control_templ.FullPage(component).Render(r.Context(), w)
	default:
		siteapiutils.WriteModelJSON(w, status, model)
	}
}
