package sitecollections

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/models/sitemodels"
	"nighttroop.t1d.nz/server/components/site_templ/site_settings_templ"
	"nighttroop.t1d.nz/server/siteapi/sitecollections/sitesettings"
	"nighttroop.t1d.nz/utils/apiutils"
	"nighttroop.t1d.nz/utils/apiutils/apiquery"
	"nighttroop.t1d.nz/utils/can"
	"nighttroop.t1d.nz/utils/httputils"
)

// coldDaysShown is how many cold days a page lists unless limit says
const coldDaysShown = 100

// coldDaysAdmin is the context of an admin's request to look at the site's cold days, or nil after refusing it. They
// are only looked at: a cold day is written by sealing its day's rows (see docs/retention.md), so there is no create,
// change or delete.
func coldDaysAdmin(w http.ResponseWriter, r *http.Request) *httputils.SiteContext {
	return SiteContextFor(w, r, (&sitesettings.SettingsCollection{Site: true}).GetName(), can.ActionGet)
}

// ColdDaysHandler is the site's cold days, the newest first, limit of them after skip, with how many there are
func ColdDaysHandler(w http.ResponseWriter, r *http.Request) {
	siteContext := coldDaysAdmin(w, r)
	if siteContext == nil {
		return
	}
	number := func(name string, otherwise, most int64) int64 {
		if n, err := strconv.ParseInt(r.URL.Query().Get(name), 10, 64); err == nil && n >= 0 {
			return min(n, most)
		}
		return otherwise
	}
	limit, skip, q := max(number("limit", coldDaysShown, apiquery.MaxQueryLimit), 1), number("skip", 0, apiquery.MaxQuerySkip), siteContext.SiteDB.Queries()
	rows, err := q.ColdDaysPage(r.Context(), sitedb.ColdDaysPageParams{Limit: limit, Skip: skip})
	days, herr := apiutils.Searched(rows, err, "cold days", func(row *sitedb.ColdDay) *sitemodels.ColdDay { return sitemodels.ColdDayFromDB(row, false) })
	totals, err := q.ColdDayTotals(r.Context())
	if herr == nil {
		herr = apiutils.ErrorFor(err, "cold days")
	}
	if herr != nil {
		writeError(w, r, siteContext.AcceptType, herr)
		return
	}
	page := &sitemodels.ColdDays{Days: days, Total: totals.Total, TotalBytes: totals.Bytes}
	WriteResponse(w, r, siteContext, http.StatusOK, page, site_settings_templ.ColdDays(page, skip, limit), true)
}

// ColdDayHandler is one cold day, by its kind and date, with each of its values
func ColdDayHandler(w http.ResponseWriter, r *http.Request) {
	siteContext := coldDaysAdmin(w, r)
	if siteContext == nil {
		return
	}
	row, err := siteContext.SiteDB.Queries().GetColdDay(r.Context(), sitedb.GetColdDayParams{Kind: chi.URLParam(r, "kind"), Date: chi.URLParam(r, "date")})
	if herr := apiutils.ErrorFor(err, "cold day"); herr != nil {
		writeError(w, r, siteContext.AcceptType, herr)
		return
	}
	day := sitemodels.ColdDayFromDB(row, true)
	WriteResponse(w, r, siteContext, http.StatusOK, day, site_settings_templ.ColdDay(day), true)
}
