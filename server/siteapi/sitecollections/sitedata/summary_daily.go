package sitedata

import (
	"context"
	"database/sql"
	"errors"
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

// RebuildDays is how many days, up to yesterday, are built again when the site's timezone changes
const RebuildDays = 7

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

func (s *SummaryDailyCollection) ToModel(row *sitedb.SummaryPeriod, siteContext *httputils.SiteContext) *sitemodels.SiteSummaryDaily {
	if row == nil {
		return nil
	}
	return sitemodels.SummaryDailyFromDB(row, siteContext)
}

func (s *SummaryDailyCollection) GetRow(ctx context.Context, collectionTx *siteapiutils.CollectionTx, id string, siteContext *httputils.SiteContext) (*sitedb.SummaryPeriod, error) {
	return siteapiutils.SiteQuery(ctx, collectionTx, (*sitedb.Queries).GetSummaryPeriod, sitedb.GetSummaryPeriodParams{Period: "day", Date: id})
}

// SearchRows is the kept days, newest first, after today so far on the first page, built from its clean readings and
// not kept, as it is not over
func (s *SummaryDailyCollection) SearchRows(ctx context.Context, collectionTx *siteapiutils.CollectionTx, query *apiutils.Query, siteContext *httputils.SiteContext) ([]*sitedb.SummaryPeriod, error) {
	q, err := collectionTx.Site()
	if err != nil {
		return nil, err
	}
	if query.Skip > 0 {
		return q.SearchSummaryPeriods(ctx, sitedb.SearchSummaryPeriodsParams{Period: "day", Limit: query.Limit, Skip: query.Skip - 1})
	}
	now := time.Now()
	today, err := buildSummaryDaily(ctx, q, siteContext, LocalDate(now, siteContext.GetSiteTimezone(), 0), now.UnixMilli())
	if err != nil {
		return nil, err
	}
	days, err := q.SearchSummaryPeriods(ctx, sitedb.SearchSummaryPeriodsParams{Period: "day", Limit: query.Limit - 1})
	return append([]*sitedb.SummaryPeriod{today}, days...), err
}

// CreateRow builds the model's date again, a day before today in the site's timezone and no more than a year ago
func (s *SummaryDailyCollection) CreateRow(ctx context.Context, collectionTx *siteapiutils.CollectionTx, id *string, siteContext *httputils.SiteContext, model *sitemodels.SiteSummaryDaily) (*sitedb.SummaryPeriod, error) {
	q, err := collectionTx.Site()
	if err != nil {
		return nil, err
	}
	location, now := siteContext.GetSiteTimezone(), time.Now()
	if _, err := time.Parse(time.DateOnly, model.Date); err != nil {
		return nil, apiutils.ErrBadBodyWithHint("date must be a day, like 2026-01-31")
	}
	if model.Date >= LocalDate(now, location, 0) || model.Date < LocalDate(now, location, -summaryDailyMaxDays) {
		return nil, apiutils.ErrBadBodyWithHint("date must be before today and within a year")
	}
	return RebuildSummaryDaily(ctx, q, siteContext, model.Date, now)
}

func (s *SummaryDailyCollection) DeleteRow(ctx context.Context, collectionTx *siteapiutils.CollectionTx, id string, siteContext *httputils.SiteContext) error {
	_, err := siteapiutils.SiteQuery(ctx, collectionTx, (*sitedb.Queries).HardDeleteSummaryPeriod, sitedb.HardDeleteSummaryPeriodParams{Period: "day", Date: id})
	return err
}

// A day is only read as JSON for now, so pages get nothing
func (s *SummaryDailyCollection) GetHTMXResponse(*sitemodels.SiteSummaryDaily) templ.Component {
	return templ.NopComponent
}

// SearchHTMXResponse is the days' glucose as small graphs, which the glucose readings page shows
func (s *SummaryDailyCollection) SearchHTMXResponse(models []*sitemodels.SiteSummaryDaily, _ *apiutils.Query) templ.Component {
	return site_data_templ.GlucoseDays(models)
}

func (s *SummaryDailyCollection) CreateHTMXResponse(*sitemodels.SiteSummaryDaily) templ.Component {
	return templ.NopComponent
}

func (s *SummaryDailyCollection) DeleteHTMXResponse(*sitemodels.SiteSummaryDaily) templ.Component {
	return templ.NopComponent
}

// SummaryDays are the site's kept days from the date first to last, as 2026-01-31, at most a year of them, oldest first
func SummaryDays(ctx context.Context, q *sitedb.Queries, first, last string) ([]*sitedb.SummaryPeriod, error) {
	days, err := q.SearchSummaryPeriods(ctx, sitedb.SearchSummaryPeriodsParams{Period: "day", DateGTE: &first, DateLTE: &last, Limit: summaryDailyMaxDays})
	slices.Reverse(days)
	return days, err
}

// LocalDate is the date days from now's, in location's time
func LocalDate(now time.Time, location *time.Location, days int) string {
	return now.In(location).AddDate(0, 0, days).Format(time.DateOnly)
}

// RebuildSummaryDaily builds the site's day named date again and keeps it
func RebuildSummaryDaily(ctx context.Context, q *sitedb.Queries, siteContext *httputils.SiteContext, date string, now time.Time) (*sitedb.SummaryPeriod, error) {
	day, err := buildSummaryDaily(ctx, q, siteContext, date, now.UnixMilli())
	if err != nil {
		return nil, err
	}
	return upsertSummaryPeriod(ctx, q, day)
}

func upsertSummaryPeriod(ctx context.Context, q *sitedb.Queries, row *sitedb.SummaryPeriod) (*sitedb.SummaryPeriod, error) {
	return q.UpsertSummaryPeriod(ctx, sitedb.UpsertSummaryPeriodParams{Period: row.Period, Date: row.Date, Timezone: row.Timezone,
		StartAt: row.StartAt, EndAt: row.EndAt, Days: row.Days, TimeInRange: row.TimeInRange, TotalDailyDose: row.TotalDailyDose,
		TotalDailyCarbs: row.TotalDailyCarbs, GlucoseMgdlJson: row.GlucoseMgdlJson, AgpMgdlJson: row.AgpMgdlJson, LoopIobJson: row.LoopIobJson,
		LoopIobBasalJson: row.LoopIobBasalJson, LoopIobTreatmentJson: row.LoopIobTreatmentJson, LoopCobJson: row.LoopCobJson})
}

// RebuildSummaryDailyWeek deletes all the site's days and periods, as from a timezone they no longer follow, and
// builds the RebuildDays up to yesterday, and the last week and month of what days there are, again
func RebuildSummaryDailyWeek(ctx context.Context, q *sitedb.Queries, siteContext *httputils.SiteContext, now time.Time) error {
	if _, err := q.DeleteSummaryPeriods(ctx, sitedb.DeleteSummaryPeriodsParams{}); err != nil {
		return err
	}
	for days := -RebuildDays; days < 0; days++ {
		if _, err := RebuildSummaryDaily(ctx, q, siteContext, LocalDate(now, siteContext.GetSiteTimezone(), days), now); err != nil {
			return err
		}
	}
	return BuildLastSummaryPeriods(ctx, q, siteContext, now)
}

// RebuildSummaryYear deletes all the site's days and periods, and builds again each day of the last year with data,
// and yesterday, and each whole week, fortnight and month of them with data, as when its admin asks for them to be
func RebuildSummaryYear(ctx context.Context, q *sitedb.Queries, siteContext *httputils.SiteContext, now time.Time) error {
	if _, err := q.DeleteSummaryPeriods(ctx, sitedb.DeleteSummaryPeriodsParams{}); err != nil {
		return err
	}
	location, weekStart := siteContext.GetSiteTimezone(), siteContext.GetSiteWeekStartDay()
	for days := -summaryDailyMaxDays; days < 0; days++ {
		day, err := buildSummaryDaily(ctx, q, siteContext, LocalDate(now, location, days), now.UnixMilli())
		if err == nil && (day.Days > 0 || days == -1) {
			_, err = upsertSummaryPeriod(ctx, q, day)
		}
		if err != nil {
			return err
		}
	}
	today := PeriodStart("day", now, location, weekStart)
	for _, period := range SummaryPeriods {
		for start := PeriodStart(period, now.AddDate(0, 0, -summaryDailyMaxDays), location, weekStart); !PeriodEnd(period, start).After(today); start = PeriodEnd(period, start) {
			row, _, err := summaryPeriodFrom(ctx, q, period, start)
			if err == nil && row.Days > 0 {
				_, err = upsertSummaryPeriod(ctx, q, row)
			}
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// RebuildSummaryWeeks deletes all the site's weeks and fortnights, as from a week start day they no longer follow,
// and builds the last ones again
func RebuildSummaryWeeks(ctx context.Context, q *sitedb.Queries, siteContext *httputils.SiteContext, now time.Time) error {
	for _, period := range []string{"week", "fortnight"} {
		if _, err := q.DeleteSummaryPeriods(ctx, sitedb.DeleteSummaryPeriodsParams{Period: &period}); err != nil {
			return err
		}
	}
	return BuildLastSummaryPeriods(ctx, q, siteContext, now)
}

// SummaryPeriods are what the summary days are put together into: weeks and fortnights from the site's week start
// day, and months
var SummaryPeriods = []string{"week", "fortnight", "month"}

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

// LastSummaryPeriod is the site's last whole week, fortnight or month, and whether it was built, built from its days
// if not
func LastSummaryPeriod(ctx context.Context, q *sitedb.Queries, siteContext *httputils.SiteContext, period string, now time.Time) (*sitedb.SummaryPeriod, bool, error) {
	return summaryPeriodFrom(ctx, q, period, LastPeriodStart(period, now, siteContext.GetSiteTimezone(), siteContext.GetSiteWeekStartDay()))
}

// summaryPeriodFrom is the site's week, fortnight or month from start, and whether it was built, built from the days
// kept in it, and any more days given, if not
func summaryPeriodFrom(ctx context.Context, q *sitedb.Queries, period string, start time.Time, more ...*sitedb.SummaryPeriod) (*sitedb.SummaryPeriod, bool, error) {
	date := start.Format(time.DateOnly)
	row, err := q.GetSummaryPeriod(ctx, sitedb.GetSummaryPeriodParams{Period: period, Date: date})
	if !errors.Is(err, sql.ErrNoRows) {
		return row, err == nil, err
	}
	end := PeriodEnd(period, start)
	last := end.AddDate(0, 0, -1).Format(time.DateOnly)
	days, err := q.SearchSummaryPeriods(ctx, sitedb.SearchSummaryPeriodsParams{Period: "day", DateGTE: &date, DateLTE: &last, Limit: 31})
	if err != nil {
		return nil, false, err
	}
	return sitemodels.NewSummaryPeriod(period, date, start.Location().String(), start.UnixMilli(), end.UnixMilli(), append(days, more...)), false, nil
}

// BuildLastSummaryPeriods builds and keeps the site's last whole week, fortnight and month, when they are not yet built
func BuildLastSummaryPeriods(ctx context.Context, q *sitedb.Queries, siteContext *httputils.SiteContext, now time.Time) error {
	for _, period := range SummaryPeriods {
		row, built, err := LastSummaryPeriod(ctx, q, siteContext, period, now)
		if err != nil {
			return err
		}
		if built {
			continue
		}
		if _, err := upsertSummaryPeriod(ctx, q, row); err != nil {
			return err
		}
	}
	return nil
}

// buildSummaryDaily is the site's day named date, from midnight in its timezone, built from its clean readings,
// insulin, basal, food and loop reports up to until, so a day not yet over is its part so far. Its time in range is from the site's
// range low to its range high.
func buildSummaryDaily(ctx context.Context, q *sitedb.Queries, siteContext *httputils.SiteContext, date string, until int64) (*sitedb.SummaryPeriod, error) {
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
	return sitemodels.NewSummaryDaily(date, location.String(), start, end, float64(low), float64(high),
		glucose.CleanReadings(rows), doses, food, reports, start, minutes), nil
}
