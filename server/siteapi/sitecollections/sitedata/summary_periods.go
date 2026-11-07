package sitedata

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/a-h/templ"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/models/sitemodels"
	"nighttroop.t1d.nz/server/siteapi/siteapiutils"
	"nighttroop.t1d.nz/utils/apiutils"
	"nighttroop.t1d.nz/utils/httputils"
)

// SummaryPeriodsCollection is the site's days, weeks, fortnights and months, read by an id of the period and a date in
// it, as week-2026-10-28 for the week with 28 October 2026 in it
type SummaryPeriodsCollection struct{}

func (s *SummaryPeriodsCollection) GetName() string {
	return "summary_periods"
}

// GetRoute is where the collection is in its URL
func (s *SummaryPeriodsCollection) GetRoute() string {
	return "summary-periods"
}

func (s *SummaryPeriodsCollection) ToModel(row *sitedb.SummaryPeriod, siteContext *httputils.SiteContext) *sitemodels.SiteSummaryPeriod {
	if row == nil {
		return nil
	}
	return sitemodels.SummaryPeriodFromDB(row, siteContext)
}

// notAPeriod is the error for an id that names no period that can be read, saying why
func notAPeriod(hint string) *apiutils.HTTPError {
	return &apiutils.HTTPError{Status: http.StatusNotFound, Message: "Not found: " + hint}
}

// GetRow is the period named by the id with the date in it, from today back a year, in the site's timezone and from
// its week start day: the kept one, or else one built, and not kept, from its kept days so far
func (s *SummaryPeriodsCollection) GetRow(ctx context.Context, collectionTx *siteapiutils.CollectionTx, id string, siteContext *httputils.SiteContext) (*sitedb.SummaryPeriod, error) {
	period, date, _ := strings.Cut(id, "-")
	location, now := siteContext.GetSiteTimezone(), time.Now()
	at, err := time.ParseInLocation(time.DateOnly, date, location)
	if err != nil || !slices.Contains(append([]string{"day"}, SummaryPeriods...), period) {
		return nil, notAPeriod("the id is day, week, fortnight or month and a date, like week-2026-01-31")
	}
	if date > LocalDate(now, location, 0) || date < LocalDate(now, location, -summaryDailyMaxDays) {
		return nil, notAPeriod("the date is from today back a year")
	}
	q, err := collectionTx.Site()
	if err != nil {
		return nil, err
	}
	return SummaryPeriodAt(ctx, q, siteContext, period, at, now)
}

// SummaryPeriodAt is the site's day, week, fortnight or month with at in it. A day not kept, as today is not, is built
// from its data up to now, and its AGP is of the 7 days to it. A period not over has today so far in it too.
func SummaryPeriodAt(ctx context.Context, q *sitedb.Queries, siteContext *httputils.SiteContext, period string, at time.Time, now time.Time) (*sitedb.SummaryPeriod, error) {
	location := siteContext.GetSiteTimezone()
	start := PeriodStart(period, at, location, siteContext.GetSiteWeekStartDay())
	if period != "day" {
		var today []*sitedb.SummaryPeriod
		if now.Before(PeriodEnd(period, start)) {
			day, err := buildSummaryDaily(ctx, q, siteContext, LocalDate(now, location, 0), now.UnixMilli())
			if err != nil {
				return nil, err
			}
			today = append(today, day)
		}
		row, _, err := summaryPeriodFrom(ctx, q, period, start, today...)
		return row, err
	}
	date := start.Format(time.DateOnly)
	day, err := q.GetSummaryPeriod(ctx, sitedb.GetSummaryPeriodParams{Period: "day", Date: date})
	if errors.Is(err, sql.ErrNoRows) {
		day, err = buildSummaryDaily(ctx, q, siteContext, date, now.UnixMilli())
	}
	if err != nil {
		return nil, err
	}
	weekStart := start.AddDate(0, 0, -6).Format(time.DateOnly)
	days, err := q.SearchSummaryPeriods(ctx, sitedb.SearchSummaryPeriodsParams{Period: "day", DateGTE: &weekStart, DateLTE: &date, Limit: 7})
	if err != nil {
		return nil, err
	}
	days = append(slices.DeleteFunc(days, func(d *sitedb.SummaryPeriod) bool { return d.Date == date }), day)
	day.AgpMgdlJson = sitemodels.NewSummaryPeriod("week", weekStart, day.Timezone, 0, 0, days).AgpMgdlJson
	return day, nil
}

// A period is only read as JSON for now, so pages get nothing
func (s *SummaryPeriodsCollection) GetHTMXResponse(*sitemodels.SiteSummaryPeriod) templ.Component {
	return templ.NopComponent
}
