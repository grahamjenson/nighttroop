package sitecollections

import (
	"cmp"
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
	"nighttroop.t1d.nz/data/events"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/models/sitemodels"
	"nighttroop.t1d.nz/server/components/site_templ"
	"nighttroop.t1d.nz/server/components/site_templ/site_reports_templ"
	"nighttroop.t1d.nz/server/components/site_templ/site_summary_templ"
	"nighttroop.t1d.nz/server/siteapi/sitecollections/sitedata"
	"nighttroop.t1d.nz/server/siteapi/siteapiutils"

	"nighttroop.t1d.nz/utils/apiutils"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/httputils"
)

type CollectionBase[R, M any] interface {
	// GetName is what the collection is called, which its permissions and errors use
	GetName() string
	// GetRoute is where the collection is in its URL, e.g. sync for /settings/connectors
	GetRoute() string
	ToModel(row *R, siteContext *httputils.SiteContext) *M
	GetRow(ctx context.Context, collectionTx *siteapiutils.CollectionTx, id string, siteContext *httputils.SiteContext) (*R, error)
}

type CollectionCreate[R, M any] interface {
	CollectionBase[R, M]

	CreateRow(ctx context.Context, collectionTx *siteapiutils.CollectionTx, id *string, siteContext *httputils.SiteContext, model *M) (*R, error)
	CreateHTMXResponse(model *M) templ.Component
}

// CollectionNew is a collection added to in a modal of all its fields
type CollectionNew[R, M any] interface {
	CollectionCreate[R, M]
	NewHTMXResponse() templ.Component
}

type CollectionUpdate[R, M any] interface {
	CollectionBase[R, M]

	UpdateRow(ctx context.Context, collectionTx *siteapiutils.CollectionTx, siteContext *httputils.SiteContext, oldModel *M, model *M) (*R, error)
	UpdateHTMXResponse(model *M) templ.Component
}

// CollectionUpsert is a collection whose update creates what is not there yet, as each type of connector has one
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

// NewHandler is the modal to add a record in, for whoever can add one
func NewHandler[R, M any](collection CollectionNew[R, M]) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		siteContext := SiteContextFor(w, r, collection.GetName(), siteapiutils.ActionCreate)
		if siteContext == nil {
			return
		}
		WriteResponse(w, r, siteContext, http.StatusOK, nil, collection.NewHTMXResponse(), true)
	}
}

///////
// UPDATE
///////

// UpdateHandler updates the model with the URL's id
func UpdateHandler[R, M any](collection CollectionUpdate[R, M]) http.HandlerFunc {
	return updateHandler(collection, nil, siteapiutils.ActionUpdate)
}

// UpsertHandler creates the model with the URL's id, or updates it when there is one
func UpsertHandler[R, M any](collection CollectionUpsert[R, M]) http.HandlerFunc {
	return updateHandler(collection, collection, siteapiutils.ActionCreate, siteapiutils.ActionUpdate)
}

// updateHandler updates the model with the URL's id; when there is none, creator creates it, or without one it is not found
func updateHandler[R, M any](collection CollectionUpdate[R, M], creator CollectionCreate[R, M], actions ...siteapiutils.Action) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		siteContext := SiteContextFor(w, r, collection.GetName(), actions...)
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
		switch {
		case herr != nil:
		case oldModel != nil:
			herr = update(r.Context(), tx, collection, siteContext, oldModel, requestModel)
		case creator != nil:
			herr = create(r.Context(), tx, creator, &id, siteContext, requestModel)
		default:
			herr = apiutils.ErrNotFound
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
		if herr == nil && model == nil {
			herr = apiutils.ErrNotFound
		}
		if herr != nil {
			writeError(w, r, siteContext.AcceptType, herr)
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
		slog.ErrorContext(r.Context(), "site context is nil")
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

// writeError sends HTMX the error as a toast, which forms add to #toaster, and anything else the error as it writes itself.
// A toast shows only what was written for users: a server error, or one in the server's own words, says it is on the
// Errors page, which keeps its words.
func writeError(w http.ResponseWriter, r *http.Request, acceptType httputils.AcceptType, herr *apiutils.HTTPError) {
	if acceptType != httputils.AcceptTypeHTMX {
		herr.Write(acceptType, w)
		return
	}
	httputils.SetErrorReason(r.Context(), herr.Message)
	apiutils.WriteStatus(w, acceptType, herr.Status)
	if herr.Status >= http.StatusInternalServerError || herr.Internal {
		siteContext := httputils.GetSiteContext(r.Context())
		admin := siteContext != nil && siteContext.IdentityModel != nil && siteContext.IdentityModel.Role == consts.IdentityRoleAdmin
		site_templ.SiteErrorRecorded(admin).Render(r.Context(), w)
		return
	}
	site_templ.SiteError(cmp.Or(herr.Hint, herr.Message)).Render(r.Context(), w)
}

// get is the model, or nil when there is none
func get[R, M any](ctx context.Context, tx *siteapiutils.CollectionTx, collection CollectionBase[R, M], id string, siteContext *httputils.SiteContext) (*M, *R, *apiutils.HTTPError) {

	row, err := collection.GetRow(ctx, tx, id, siteContext)
	if errors.Is(err, sql.ErrNoRows) || (err == nil && row == nil) {
		return nil, nil, nil
	}
	if herr := apiutils.ErrorFor(err, collection.GetName()); herr != nil {
		return nil, nil, herr
	}
	return collection.ToModel(row, siteContext), row, nil
}

func search[R, M any](ctx context.Context, tx *siteapiutils.CollectionTx, collection CollectionSearch[R, M], query *apiutils.Query, siteContext *httputils.SiteContext) ([]*M, *apiutils.HTTPError) {
	rows, err := collection.SearchRows(ctx, tx, query, siteContext)
	if herr := apiutils.ErrorFor(err, collection.GetName()); herr != nil {
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
	if herr := apiutils.ErrorFor(err, collection.GetName()); herr != nil {
		return herr
	}
	reportChange(tx, events.Created, row)
	*model = *collection.ToModel(row, siteContext)
	return nil
}

func update[R, M any](ctx context.Context, tx *siteapiutils.CollectionTx, collection CollectionUpdate[R, M], siteContext *httputils.SiteContext, oldModel *M, model *M) *apiutils.HTTPError {

	row, err := collection.UpdateRow(ctx, tx, siteContext, oldModel, model)
	if herr := apiutils.ErrorFor(err, collection.GetName()); herr != nil {
		return herr
	}
	reportChange(tx, events.Updated, row)
	*model = *collection.ToModel(row, siteContext)
	return nil
}

// delete deletes the row read as the model, so the change can say what was deleted
func delete[R, M any](ctx context.Context, siteContext *httputils.SiteContext, tx *siteapiutils.CollectionTx, collection CollectionDelete[R, M], id string, row *R) *apiutils.HTTPError {

	if herr := apiutils.ErrorFor(collection.DeleteRow(ctx, tx, id, siteContext), collection.GetName()); herr != nil {
		return herr
	}
	reportChange(tx, events.Deleted, row)
	return nil
}

// reportChange reports a change to the site's own rows, to be published once the transaction commits, only to the site
func reportChange[R any](tx *siteapiutils.CollectionTx, action events.Action, row *R) {
	tx.ReportChange(action, row)
}

// GettingStartedHandler is the guide to setting up the site, a page for people, so it has no JSON
// ReportsHandler is get's reports page of the week with yesterday in it, in the site's timezone
func ReportsHandler(get http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if siteContext := httputils.GetSiteContext(r.Context()); siteContext != nil {
			chi.RouteContext(r.Context()).URLParams.Add("id", "week-"+sitedata.LocalDate(time.Now(), siteContext.GetSiteTimezone(), -1))
		}
		get(w, r)
	}
}

// TDDReportHandler is the TDD report of the week, fortnight or month before today, by its range query, the week
// without one; a page for people, so it has no JSON
func TDDReportHandler(w http.ResponseWriter, r *http.Request) {
	siteContext := SiteContextFor(w, r, (&sitedata.SummaryDailyCollection{}).GetName(), siteapiutils.ActionSearch)
	if siteContext == nil {
		return
	}
	if siteContext.AcceptType != httputils.AcceptTypeHTMX && siteContext.AcceptType != httputils.AcceptTypeHTML {
		writeError(w, r, siteContext.AcceptType, apiutils.ErrNotFound)
		return
	}
	rangeName := r.URL.Query().Get("range")
	if rangeName == "" {
		rangeName = "week"
	}
	count := site_reports_templ.TDDRangeDays(rangeName)
	if count == 0 {
		writeError(w, r, siteContext.AcceptType, &apiutils.HTTPError{Status: http.StatusBadRequest, Message: "range is week, fortnight or month"})
		return
	}
	rows, err := sitedata.PastSummaryDays(r.Context(), siteContext.SiteDB.Queries(), siteContext.GetSiteTimezone(), time.Now(), count)
	days, herr := apiutils.Searched(rows, err, "summary_daily", apiutils.WithContext(sitemodels.SummaryDailyFromDB, siteContext))
	if herr != nil {
		writeError(w, r, siteContext.AcceptType, herr)
		return
	}
	WriteResponse(w, r, siteContext, http.StatusOK, nil, site_reports_templ.TDDReport(rangeName, days), true)
}

// summaryLimits are how many of each period's kept summaries, newest first, the summaries report shows
var summaryLimits = map[string]int64{"day": 31, "week": 26, "fortnight": 13, "month": 12}

// SummariesReportHandler is the report of the site's kept summaries as tables of their numbers, a page for people, so
// it has no JSON
func SummariesReportHandler(w http.ResponseWriter, r *http.Request) {
	siteContext := SiteContextFor(w, r, (&sitedata.SummaryPeriodsCollection{}).GetName(), siteapiutils.ActionSearch)
	if siteContext == nil {
		return
	}
	if siteContext.AcceptType != httputils.AcceptTypeHTMX && siteContext.AcceptType != httputils.AcceptTypeHTML {
		writeError(w, r, siteContext.AcceptType, apiutils.ErrNotFound)
		return
	}
	summaries := map[string][]*sitemodels.SiteSummaryPeriod{}
	for period, limit := range summaryLimits {
		rows, err := siteContext.SiteDB.Queries().SearchSummaryPeriods(r.Context(), sitedb.SearchSummaryPeriodsParams{Period: period, Limit: limit})
		models, herr := apiutils.Searched(rows, err, "summary_periods", apiutils.WithContext(sitemodels.SummaryPeriodFromDB, siteContext))
		if herr != nil {
			writeError(w, r, siteContext.AcceptType, herr)
			return
		}
		summaries[period] = models
	}
	admin := siteContext.IdentityModel.Role == consts.IdentityRoleAdmin
	WriteResponse(w, r, siteContext, http.StatusOK, nil, site_reports_templ.SummariesReport(summaries, admin), true)
}

// RegenerateSummariesHandler asks for the site's summaries to be built again from its data, which a background job
// does, as a year of them takes a while. Only its admins can, and asking again before it runs asks once.
func RegenerateSummariesHandler(w http.ResponseWriter, r *http.Request) {
	siteContext := SiteContextFor(w, r, (&sitedata.SummaryPeriodsCollection{}).GetName(), siteapiutils.ActionSearch)
	if siteContext == nil {
		return
	}
	if siteContext.IdentityModel.Role != consts.IdentityRoleAdmin {
		writeError(w, r, siteContext.AcceptType, apiutils.ErrForbidden)
		return
	}
	siteContext.ControlDB.Logs.SummariesRegenerationRequestedEvent(r.Context(), siteContext.SiteID, siteContext.IdentityModel.Name)
	WriteResponse(w, r, siteContext, http.StatusAccepted, map[string]bool{"regenerating": true}, site_reports_templ.SummariesRegenerating(), false)
}

func GettingStartedHandler(w http.ResponseWriter, r *http.Request) {
	siteContext := httputils.GetSiteContext(r.Context())
	if siteContext.AcceptType != httputils.AcceptTypeHTMX && siteContext.AcceptType != httputils.AcceptTypeHTML {
		writeError(w, r, siteContext.AcceptType, apiutils.ErrNotFound)
		return
	}
	WriteResponse(w, r, siteContext, http.StatusOK, nil, site_summary_templ.GettingStartedPage(), true)
}
