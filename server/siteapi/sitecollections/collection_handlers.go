package sitecollections

import (
	"cmp"
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"strings"
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

// CollectionTotals is a searched collection that counts every row its search matches, over all its pages
type CollectionTotals interface {
	SearchTotals(ctx context.Context, collectionTx *siteapiutils.CollectionTx, query *apiutils.Query, siteContext *httputils.SiteContext) (*apiutils.Totals, error)
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
		oldModel, oldRow, herr := get(r.Context(), tx, collection, id, siteContext)
		switch {
		case herr != nil:
		case oldModel != nil:
			herr = update(r.Context(), tx, collection, siteContext, oldModel, oldRow, requestModel)
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
		var totals *apiutils.Totals
		if counted, ok := collection.(CollectionTotals); ok && herr == nil {
			var err error
			totals, err = counted.SearchTotals(r.Context(), tx, query, siteContext)
			herr = apiutils.ErrorFor(err, collection.GetName())
		}
		if herr = tx.Finish(herr); herr != nil {
			writeError(w, r, siteContext.AcceptType, herr)
			return
		}

		r = r.WithContext(apiutils.WithTotals(r.Context(), totals))
		result := &apiutils.SearchResult[M]{Items: models, Skip: query.Skip, Limit: query.Limit, Totals: totals}
		WriteResponse(w, r, siteContext, http.StatusOK, result, collection.SearchHTMXResponse(models, query), true)
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

// WriteResponse sends HTMX the component, a browser the component (in the site's page when fullPage), and anything else
// the model as JSON. A page, loaded in full or opened in the main content, is in its hub's tabs.
func WriteResponse(w http.ResponseWriter, r *http.Request, siteContext *httputils.SiteContext, status int, model any, component templ.Component, fullPage bool) {
	apiutils.WriteStatus(w, siteContext.AcceptType, status)
	switch siteContext.AcceptType {
	case httputils.AcceptTypeHTMX:
		if fullPage && replacesPage(r) {
			component = site_templ.InHub(r.URL.Path, component, true)
		}
		component.Render(r.Context(), w)
	case httputils.AcceptTypeHTML:
		if fullPage {
			component = site_templ.FullPage(r.URL.Path, component)
		}
		component.Render(r.Context(), w)
	default:
		apiutils.WriteModelJSON(w, status, model)
	}
}

// replacesPage is whether an HTMX request swaps the whole page in the main content, as the header's links, the hubs'
// tabs and the pagers do, rather than a part of it, such as a table, a form's answer, a modal or a live update
func replacesPage(r *http.Request) bool {
	target := r.Header.Get("HX-Target")
	return target == "main-content" || strings.HasSuffix(target, "#main-content")
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

func update[R, M any](ctx context.Context, tx *siteapiutils.CollectionTx, collection CollectionUpdate[R, M], siteContext *httputils.SiteContext, oldModel *M, oldRow *R, model *M) *apiutils.HTTPError {

	row, err := collection.UpdateRow(ctx, tx, siteContext, oldModel, model)
	if herr := apiutils.ErrorFor(err, collection.GetName()); herr != nil {
		return herr
	}
	tx.ReportChange(events.Updated, row, oldRow)
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
	tx.ReportChange(action, row, nil)
}

// GettingStartedHandler is the guide to setting up the site, a page for people, so it has no JSON
// PeriodReportHandler is the report page of the day, week, fortnight or month its id names, as week-2026-01-31, or of
// the 7 days to yesterday without one, made from the period, its days and the period before it, to compare with, nil
// when it is out of range; a page for people, so it has no JSON
func PeriodReportHandler(page func(p *sitemodels.SiteSummaryPeriod, days []*sitemodels.SiteSummaryDaily, previous *sitemodels.SiteSummaryPeriod) templ.Component) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		siteContext := SiteContextFor(w, r, (&sitedata.SummaryPeriodsCollection{}).GetName(), siteapiutils.ActionGet)
		if siteContext == nil {
			return
		}
		if siteContext.AcceptType != httputils.AcceptTypeHTMX && siteContext.AcceptType != httputils.AcceptTypeHTML {
			writeError(w, r, siteContext.AcceptType, apiutils.ErrNotFound)
			return
		}
		now := time.Now()
		id := cmp.Or(chi.URLParam(r, "id"), "week-"+sitedata.LocalDate(now, siteContext.GetSiteTimezone(), -1))
		report, herr := periodReport(r, siteContext, id, now)
		if herr != nil {
			writeError(w, r, siteContext.AcceptType, herr)
			return
		}
		WriteResponse(w, r, siteContext, http.StatusOK, nil, page(report.Period, report.Days, report.Previous), true)
	}
}

// periodReport is the period its id names, its days and the period before it, to compare with, nil when it has no data
func periodReport(r *http.Request, siteContext *httputils.SiteContext, id string, now time.Time) (*site_reports_templ.PeriodReport, *apiutils.HTTPError) {
	row, rows, err := sitedata.SummaryReport(r.Context(), siteContext.SiteDB.Queries(), siteContext, id, now)
	if notFound := (*apiutils.HTTPError)(nil); errors.As(err, &notFound) {
		return nil, notFound
	}
	days, herr := apiutils.Searched(rows, err, "summary_periods", apiutils.WithContext(sitemodels.SummaryDailyFromDB, siteContext))
	if herr != nil {
		return nil, herr
	}
	report := &site_reports_templ.PeriodReport{Period: sitemodels.SummaryPeriodFromDB(row, siteContext), Days: days}
	before := row.Period + "-" + time.UnixMilli(row.StartAt).In(siteContext.GetSiteTimezone()).AddDate(0, 0, -1).Format(time.DateOnly)
	// A period before with no data has nothing to compare with, though some of its numbers, as its time CGM active,
	// are 0 rather than missing
	if previousRow, _, err := sitedata.SummaryReport(r.Context(), siteContext.SiteDB.Queries(), siteContext, before, now); err == nil && previousRow.Days > 0 {
		report.Previous = sitemodels.SummaryPeriodFromDB(previousRow, siteContext)
	}
	return report, nil
}

// PrintReportHandler is the printable report of this month so far and the 3 months before it, each its overview and
// AGP on pages of their own; a page for people, so it has no JSON. Each builds four months, so a site has only a few a
// minute.
func PrintReportHandler(w http.ResponseWriter, r *http.Request) {
	siteContext := SiteContextFor(w, r, (&sitedata.SummaryPeriodsCollection{}).GetName(), siteapiutils.ActionGet)
	if siteContext == nil {
		return
	}
	if siteContext.AcceptType != httputils.AcceptTypeHTML {
		writeError(w, r, siteContext.AcceptType, apiutils.ErrNotFound)
		return
	}
	if allowed, wait := httputils.WithinRateLimitAll(siteContext.Config.RateLimiters.SiteReportPrintLimiter(siteContext.SiteID)); !allowed {
		writeError(w, r, siteContext.AcceptType, apiutils.ErrRateLimitExceeded(wait))
		return
	}
	now := time.Now()
	local := now.In(siteContext.GetSiteTimezone())
	first := time.Date(local.Year(), local.Month(), 1, 0, 0, 0, 0, local.Location())
	var months []*site_reports_templ.PeriodReport
	for i := range site_reports_templ.PrintMonths {
		report, herr := periodReport(r, siteContext, "month-"+first.AddDate(0, -i, 0).Format(time.DateOnly), now)
		if herr != nil {
			writeError(w, r, siteContext.AcceptType, herr)
			return
		}
		months = append(months, report)
	}
	WriteResponse(w, r, siteContext, http.StatusOK, nil, site_reports_templ.Printable(now, months), false)
}

// TDDReportHandler is the TDD report of the month ending on the day its end query names, today so far without one; a
// page for people, so it has no JSON
func TDDReportHandler(w http.ResponseWriter, r *http.Request) {
	siteContext := SiteContextFor(w, r, (&sitedata.SummaryDailyCollection{}).GetName(), siteapiutils.ActionSearch)
	if siteContext == nil {
		return
	}
	if siteContext.AcceptType != httputils.AcceptTypeHTMX && siteContext.AcceptType != httputils.AcceptTypeHTML {
		writeError(w, r, siteContext.AcceptType, apiutils.ErrNotFound)
		return
	}
	now := time.Now()
	last, ok := site_reports_templ.TDDEnd(r.URL.Query().Get("end"), now, siteContext.GetSiteTimezone())
	if !ok {
		writeError(w, r, siteContext.AcceptType, &apiutils.HTTPError{Status: http.StatusBadRequest, Message: "end is a day of the past year, up to today, as 2026-01-31"})
		return
	}
	first := last.AddDate(0, 0, 1-site_reports_templ.TDDDays)
	rows, err := sitedata.SummaryDaysSoFar(r.Context(), siteContext.SiteDB.Queries(), siteContext, first.Format(time.DateOnly), last.Format(time.DateOnly), now)
	days, herr := apiutils.Searched(rows, err, "summary_daily", apiutils.WithContext(sitemodels.SummaryDailyFromDB, siteContext))
	if herr != nil {
		writeError(w, r, siteContext.AcceptType, herr)
		return
	}
	WriteResponse(w, r, siteContext, http.StatusOK, nil, site_reports_templ.TDDReport(last, days), true)
}

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
	first, ok := site_reports_templ.SummariesMonth(r.URL.Query().Get("month"), time.Now(), siteContext.GetSiteTimezone())
	if !ok {
		writeError(w, r, siteContext.AcceptType, &apiutils.HTTPError{Status: http.StatusBadRequest, Message: "month is this month or one of the past year's, as 2026-01"})
		return
	}
	summaries := map[string][]*sitemodels.SiteSummaryPeriod{}
	for period, dates := range site_reports_templ.SummariesDates(first) {
		rows, err := siteContext.SiteDB.Queries().SearchSummaryPeriods(r.Context(), sitedb.SearchSummaryPeriodsParams{Period: period,
			DateGTE: &dates[0], DateLTE: &dates[1], Limit: 31})
		models, herr := apiutils.Searched(rows, err, "summary_periods", apiutils.WithContext(sitemodels.SummaryPeriodFromDB, siteContext))
		if herr != nil {
			writeError(w, r, siteContext.AcceptType, herr)
			return
		}
		summaries[period] = models
	}
	admin := siteContext.IdentityModel.Role == consts.IdentityRoleAdmin
	WriteResponse(w, r, siteContext, http.StatusOK, nil, site_reports_templ.SummariesReport(first, summaries, admin), true)
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
