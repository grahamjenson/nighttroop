package sitedata

import (
	"cmp"
	"context"
	"database/sql"
	"slices"
	"time"

	"github.com/a-h/templ"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/models/sitemodels"
	"nighttroop.t1d.nz/server/components/site_templ/site_data_templ"
	"nighttroop.t1d.nz/server/siteapi/siteapiutils"
	"nighttroop.t1d.nz/utils/apiutils"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/glucose"
	"nighttroop.t1d.nz/utils/httputils"
)

// summaryDailyMaxDays is how many days back a summary day can be built, so a request cannot build years of them
const summaryDailyMaxDays = 366

// SummaryDailyCollection is the site's days, each built from its data: searched newest first, built again for a date
// by a create, and deleted for good
type SummaryDailyCollection struct{}

func (s *SummaryDailyCollection) GetName() string {
	return "summary_daily"
}

// GetRoute is where the collection is in its URL
func (s *SummaryDailyCollection) GetRoute() string {
	return "summary-daily"
}

func (s *SummaryDailyCollection) ToModel(row *sitemodels.SummaryRow, siteContext *httputils.SiteContext) *sitemodels.SiteSummaryDaily {
	if row == nil {
		return nil
	}
	return sitemodels.SummaryDailyFromDB(row, siteContext)
}

// GetRow is the day named by the id, built as it is read, or not found for a day with no data
func (s *SummaryDailyCollection) GetRow(ctx context.Context, collectionTx *siteapiutils.CollectionTx, id string, siteContext *httputils.SiteContext) (*sitemodels.SummaryRow, error) {
	q, err := collectionTx.Site()
	if err != nil {
		return nil, err
	}
	if _, err := time.Parse(time.DateOnly, id); err != nil {
		return nil, sql.ErrNoRows
	}
	days, err := builtSummaryDays(ctx, q, siteContext, id, id, time.Now(), false)
	if err != nil || len(days) == 0 {
		return nil, cmp.Or(err, sql.ErrNoRows)
	}
	return days[0], nil
}

// SearchRows is the days, newest first, from today so far back to the first with data: a page is the limit of dates,
// back from the skip of them before today
func (s *SummaryDailyCollection) SearchRows(ctx context.Context, collectionTx *siteapiutils.CollectionTx, query *apiutils.Query, siteContext *httputils.SiteContext) ([]*sitemodels.SummaryRow, error) {
	q, err := collectionTx.Site()
	if err != nil || query.Skip >= summaryDailyMaxDays {
		return nil, err
	}
	now, location := time.Now(), siteContext.GetSiteTimezone()
	last := int(-query.Skip)
	days, err := builtSummaryDays(ctx, q, siteContext, LocalDate(now, location, last+1-int(min(query.Limit, summaryDailyMaxDays))), LocalDate(now, location, last), now, true)
	slices.Reverse(days)
	return days, err
}

// A day is only read as JSON for now, so pages get nothing
func (s *SummaryDailyCollection) GetHTMXResponse(*sitemodels.SiteSummaryDaily) templ.Component {
	return templ.NopComponent
}

// SearchHTMXResponse is the days' glucose as small graphs, which the glucose readings page shows
func (s *SummaryDailyCollection) SearchHTMXResponse(models []*sitemodels.SiteSummaryDaily, _ *apiutils.Query) templ.Component {
	return site_data_templ.GlucoseDays(models)
}

// SummaryDaysSoFar are the site's days with data from the date first to last, as 2026-01-31, oldest first, and today so
// far after them when last is today (see builtSummaryDays)
func SummaryDaysSoFar(ctx context.Context, q *sitedb.Queries, siteContext *httputils.SiteContext, first, last string, now time.Time) ([]*sitemodels.SummaryRow, error) {
	return builtSummaryDays(ctx, q, siteContext, first, last, now, false)
}

// LocalDate is the date days from now's, in location's time
func LocalDate(now time.Time, location *time.Location, days int) string {
	return now.In(location).AddDate(0, 0, days).Format(time.DateOnly)
}

// PeriodStart is the midnight, in location, that starts the day, week or fortnight, from weekStart (monday or sunday),
// or month that at is in. Fortnights are every other week from the first weekStart of 1970, so each week is always in
// the same one.
func PeriodStart(period string, at time.Time, location *time.Location, weekStart string) time.Time {
	local := at.In(location)
	switch period {
	case "day":
		return time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, location)
	case "month":
		return time.Date(local.Year(), local.Month(), 1, 0, 0, 0, 0, location)
	}
	sinceStart, firstStart := int(local.Weekday()), 3 // 4 Jan 1970, the first Sunday, is day 3
	if weekStart == "monday" {
		sinceStart, firstStart = (sinceStart+6)%7, 4
	}
	weekDay := time.Date(local.Year(), local.Month(), local.Day()-sinceStart, 0, 0, 0, 0, time.UTC).Unix() / 86400
	if period == "fortnight" {
		sinceStart += 7 * int((weekDay-int64(firstStart))/7%2)
	}
	return time.Date(local.Year(), local.Month(), local.Day()-sinceStart, 0, 0, 0, 0, location)
}

// LastPeriodStart is the midnight, in location, that starts the last whole week, fortnight or month before now's
func LastPeriodStart(period string, now time.Time, location *time.Location, weekStart string) time.Time {
	return PeriodStart(period, PeriodStart(period, now, location, weekStart).AddDate(0, 0, -1), location, weekStart)
}

// PeriodEnd is the midnight that ends the day, week, fortnight or month from start
func PeriodEnd(period string, start time.Time) time.Time {
	if period == "month" {
		return start.AddDate(0, 1, 0)
	}
	return start.AddDate(0, 0, map[string]int{"day": 1, "week": 7, "fortnight": 14}[period])
}

// builtSummaryPeriod is the site's week, fortnight or month from start, built as it is read from its days so far
func builtSummaryPeriod(ctx context.Context, q *sitedb.Queries, siteContext *httputils.SiteContext, period string, start, now time.Time) (*sitemodels.SummaryRow, error) {
	date, end := start.Format(time.DateOnly), PeriodEnd(period, start)
	days, err := builtSummaryDays(ctx, q, siteContext, date, end.AddDate(0, 0, -1).Format(time.DateOnly), now, false)
	if err != nil {
		return nil, err
	}
	return sitemodels.NewSummaryPeriod(period, date, start.Location().String(), start.UnixMilli(), end.UnixMilli(), now.UnixMilli(), days), nil
}

// extendedLookback is the longest a dose's extended part runs, as its column takes
const extendedLookback = 7 * 24 * time.Hour

// buildSummaryDaily is the site's day named date, from midnight in its timezone, built from its clean readings,
// insulin, basal, food and loop reports up to until, so a day not yet over is its part so far. Its time in range is from the site's
// range low to its range high.
func buildSummaryDaily(ctx context.Context, q *sitedb.Queries, siteContext *httputils.SiteContext, date string, until int64) (*sitemodels.SummaryRow, error) {
	location := siteContext.GetSiteTimezone()
	midnight, err := time.ParseInLocation(time.DateOnly, date, location)
	if err != nil {
		return nil, err
	}
	start, end := midnight.UnixMilli(), midnight.AddDate(0, 0, 1).UnixMilli()
	upTo, excludeDeleted := min(end, until), false
	rows, err := q.SearchGlucoseReadings(ctx, sitedb.SearchGlucoseReadingsParams{DateGTE: new(start - trendLookback.Milliseconds()), DateLT: &upTo, Limit: maxRawReadings})
	if err != nil {
		return nil, err
	}
	doses, err := q.SearchInsulinDoses(ctx, sitedb.SearchInsulinDosesParams{DateGTE: &start, DateLT: &upTo, IncludeDeleted: &excludeDeleted, Limit: maxRows})
	if err != nil {
		return nil, err
	}
	// From a week before, the longest a dose's extended part runs, only those still running into the day, which gives
	// it their share, so a loop's week of SMBs is not read for them
	extended, err := q.InsulinDosesExtendedInto(ctx, sitedb.InsulinDosesExtendedIntoParams{Since: start - extendedLookback.Milliseconds(), At: start, Limit: maxRows})
	if err != nil {
		return nil, err
	}
	doses = append(doses, extended...)
	food, err := q.SearchFoodIntakes(ctx, sitedb.SearchFoodIntakesParams{DateGTE: &start, DateLT: &upTo, IncludeDeleted: &excludeDeleted, Limit: maxRows})
	if err != nil {
		return nil, err
	}
	reports, err := q.DeviceReportsLoopBetween(ctx, sitedb.DeviceReportsLoopBetweenParams{Start: start, End: upTo, Limit: maxRows})
	if err != nil {
		return nil, err
	}
	basalRows, err := q.SearchTempBasals(ctx, sitedb.SearchTempBasalsParams{DateGTE: &start, DateLT: &upTo, IncludeDeleted: &excludeDeleted, Limit: maxRows})
	basals, herr := apiutils.Searched(basalRows, err, "summary daily temp basals", sitemodels.TempBasalFromDB)
	if herr != nil {
		return nil, herr
	}
	minutes, herr := summaryMinutes(ctx, q, start, upTo, new(consts.UnitsMGDL), location, nil, nil, basals)
	if herr != nil {
		return nil, herr
	}
	low, high := siteContext.GetSiteRangeMgdl()
	return sitemodels.NewSummaryDaily(date, location.String(), start, end, until, float64(low), float64(high),
		glucose.CleanReadings(rows), doses, food, reports, start, minutes), nil
}
