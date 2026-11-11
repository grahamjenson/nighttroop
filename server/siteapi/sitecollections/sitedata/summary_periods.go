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
	now := time.Now()
	period, at, err := periodID(id, siteContext.GetSiteTimezone(), now)
	if err != nil {
		return nil, err
	}
	q, err := collectionTx.Site()
	if err != nil {
		return nil, err
	}
	return SummaryPeriodAt(ctx, q, siteContext, period, at, now)
}

// periodID is the period and date an id names, as week-2026-01-31, or not found for any other id or a date not from
// today back a year
func periodID(id string, location *time.Location, now time.Time) (string, time.Time, error) {
	period, date, _ := strings.Cut(id, "-")
	at, err := time.ParseInLocation(time.DateOnly, date, location)
	if err != nil || !slices.Contains(append([]string{"day"}, SummaryPeriods...), period) {
		return "", at, notAPeriod("the id is day, week, fortnight or month and a date, like week-2026-01-31")
	}
	if date > LocalDate(now, location, 0) || date < LocalDate(now, location, -summaryDailyMaxDays) {
		return "", at, notAPeriod("the date is from today back a year")
	}
	return period, at, nil
}

// SummariesCheckedOverlap is how far before a site's summaries_checked_at its data is looked through again, as a
// write's updated_at is taken before it commits
const SummariesCheckedOverlap = time.Minute

// SummariesOutOfDate is whether a report's summaries, the period's and its days', before today, are to be built again:
// marked stale, or with data in them written since the site's summaries were last checked
func SummariesOutOfDate(ctx context.Context, q *sitedb.Queries, siteContext *httputils.SiteContext, rows []*sitedb.SummaryPeriod, now time.Time) (bool, error) {
	today := PeriodStart("day", now, siteContext.GetSiteTimezone(), siteContext.GetSiteWeekStartDay()).UnixMilli()
	start, end := today, int64(0)
	for _, row := range rows {
		if row.StaleAt != nil {
			return true, nil
		}
		start, end = min(start, row.StartAt), max(end, row.EndAt)
	}
	checked := siteContext.SiteModel.SummariesCheckedAt
	if checked == nil || start >= today {
		return false, nil
	}
	// Without the rebuild's overlap, as what was written just before it was checked has been built
	written, err := q.SummaryDataWrittenSince(ctx, sitedb.SummaryDataWrittenSinceParams{Since: *checked, Before: min(end, today)})
	return slices.ContainsFunc(written, func(at int64) bool { return at >= start }), err
}

// SummaryReport is the report the id names and its days, oldest first: a day, as GetRow reads it, and itself; a month,
// as GetRow reads it, and its days kept and today so far while it is in it; and a week or fortnight, the ReportDays to
// the id's date, built from its days kept and today so far when it is the last
func SummaryReport(ctx context.Context, q *sitedb.Queries, siteContext *httputils.SiteContext, id string, now time.Time) (*sitedb.SummaryPeriod, []*sitedb.SummaryPeriod, error) {
	period, at, err := periodID(id, siteContext.GetSiteTimezone(), now)
	if err != nil {
		return nil, nil, err
	}
	if n, ok := sitemodels.ReportDays[period]; ok {
		first, last := at.AddDate(0, 0, 1-n), at.Format(time.DateOnly)
		days, err := SummaryDaysSoFar(ctx, q, siteContext, first.Format(time.DateOnly), last, now)
		row := sitemodels.NewSummaryPeriod(period, first.Format(time.DateOnly), at.Location().String(), first.UnixMilli(), at.AddDate(0, 0, 1).UnixMilli(), now.UnixMilli(), days)
		return row, days, err
	}
	row, err := SummaryPeriodAt(ctx, q, siteContext, period, at, now)
	if err != nil || period == "day" {
		return row, []*sitedb.SummaryPeriod{row}, err
	}
	location := siteContext.GetSiteTimezone()
	first, end := time.UnixMilli(row.StartAt).In(location), time.UnixMilli(row.EndAt).In(location)
	days, err := SummaryDays(ctx, q, first.Format(time.DateOnly), end.AddDate(0, 0, -1).Format(time.DateOnly))
	if err != nil || !now.Before(end) {
		return row, days, err
	}
	today, err := buildSummaryDaily(ctx, q, siteContext, LocalDate(now, location, 0), now.UnixMilli())
	return row, append(days, today), err
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
		row, _, err := summaryPeriodFrom(ctx, q, period, start, now, today...)
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
	day.AgpMgdlJson = sitemodels.NewSummaryPeriod("week", weekStart, day.Timezone, 0, 0, now.UnixMilli(), days).AgpMgdlJson
	return day, nil
}

// A period is only read as JSON, as its report pages, as /reports/agp, have its days too, so pages get nothing
func (s *SummaryPeriodsCollection) GetHTMXResponse(*sitemodels.SiteSummaryPeriod) templ.Component {
	return templ.NopComponent
}
