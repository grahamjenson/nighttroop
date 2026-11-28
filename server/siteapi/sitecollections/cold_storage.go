package sitecollections

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/models/sitemodels"
	"nighttroop.t1d.nz/server/components/site_templ/site_settings_templ"
	"nighttroop.t1d.nz/server/siteapi/sitecollections/sitesettings"
	"nighttroop.t1d.nz/utils/apiutils"
	"nighttroop.t1d.nz/utils/apiutils/apiquery"
	"nighttroop.t1d.nz/utils/can"
	"nighttroop.t1d.nz/utils/httputils"
	"nighttroop.t1d.nz/utils/i18n"
)

// coldDaysShown is how many cold days a page lists unless limit says
const coldDaysShown = 100

// coldStorageAdmin is the context of an admin's request to look at the site's cold storage, or nil after refusing it.
// It is only looked at: a cold day is written by packaging its day's rows (see docs/retention.md), so there is no
// create, change or delete.
func coldStorageAdmin(w http.ResponseWriter, r *http.Request) *httputils.SiteContext {
	return SiteContextFor(w, r, (&sitesettings.SettingsCollection{Site: true}).GetName(), can.ActionGet)
}

// coldStorageFilter is the filter a request asks for: kind, date_gt and date_lt, each a UTC day as 2006-01-02, and
// order, oldest for the oldest first. A date that is not a day is refused.
func coldStorageFilter(r *http.Request) (*sitemodels.ColdStorageFilter, *apiutils.HTTPError) {
	query := r.URL.Query()
	filter := &sitemodels.ColdStorageFilter{Oldest: query.Get("order") == "oldest"}
	for name, value := range map[string]**string{"kind": &filter.Kind, "date_gt": &filter.DateGT, "date_lt": &filter.DateLT} {
		given := query.Get(name)
		if given == "" {
			continue
		}
		if name != "kind" {
			if _, err := time.Parse(time.DateOnly, given); err != nil {
				return nil, apiutils.ErrConstraintCheckFailedWithHint(i18n.NewError("errors.site.cold_storage_date"))
			}
		}
		*value = &given
	}
	return filter, nil
}

// ColdStorageHandler is the site's cold days of a kind, after and before a day, each when asked for, the newest first
// or the oldest, limit of them after skip, with how many there are and the kinds there are to filter by
func ColdStorageHandler(w http.ResponseWriter, r *http.Request) {
	siteContext := coldStorageAdmin(w, r)
	if siteContext == nil {
		return
	}
	filter, herr := coldStorageFilter(r)
	if herr != nil {
		writeError(w, r, siteContext.AcceptType, herr)
		return
	}
	number := func(name string, otherwise, most int64) int64 {
		if n, err := strconv.ParseInt(r.URL.Query().Get(name), 10, 64); err == nil && n >= 0 {
			return min(n, most)
		}
		return otherwise
	}
	filter.Limit, filter.Skip = max(number("limit", coldDaysShown, apiquery.MaxQueryLimit), 1), number("skip", 0, apiquery.MaxQuerySkip)
	q := siteContext.SiteDB.Queries()
	params := sitedb.ColdDaysPageParams{Kind: filter.Kind, DateGT: filter.DateGT, DateLT: filter.DateLT, Limit: filter.Limit, Skip: filter.Skip}
	var rows []*sitedb.ColdDay
	var err error
	if filter.Oldest {
		rows, err = q.ColdDaysPageOldest(r.Context(), sitedb.ColdDaysPageOldestParams(params))
	} else {
		rows, err = q.ColdDaysPage(r.Context(), params)
	}
	days, herr := apiutils.Searched(rows, err, "cold days", func(row *sitedb.ColdDay) *sitemodels.ColdDay { return sitemodels.ColdDayFromDB(row, false) })
	totals, err := q.ColdDayTotals(r.Context(), sitedb.ColdDayTotalsParams{Kind: filter.Kind, DateGT: filter.DateGT, DateLT: filter.DateLT})
	if herr == nil {
		herr = apiutils.ErrorFor(err, "cold days")
	}
	kinds, err := q.ColdDayKinds(r.Context())
	if herr == nil {
		herr = apiutils.ErrorFor(err, "cold days")
	}
	if herr != nil {
		writeError(w, r, siteContext.AcceptType, herr)
		return
	}
	page := &sitemodels.ColdDays{Days: days, Total: totals.Total, TotalBytes: totals.Bytes, Kinds: append([]string{}, kinds...)}
	WriteResponse(w, r, siteContext, http.StatusOK, page, site_settings_templ.ColdStorage(page, filter), true)
}

// ColdStorageDayHandler is one cold day, by its kind and date, with each of its values
func ColdStorageDayHandler(w http.ResponseWriter, r *http.Request) {
	siteContext := coldStorageAdmin(w, r)
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
