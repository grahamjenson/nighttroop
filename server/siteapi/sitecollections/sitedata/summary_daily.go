package sitedata

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/a-h/templ"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/models/sitemodels"
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

func (s *SummaryDailyCollection) ToModel(row *sitedb.SummaryDaily, siteContext *httputils.SiteContext) *sitemodels.SiteSummaryDaily {
	if row == nil {
		return nil
	}
	return sitemodels.SummaryDailyFromDB(row, siteContext)
}

func (s *SummaryDailyCollection) GetRow(ctx context.Context, collectionTx *siteapiutils.CollectionTx, id string, siteContext *httputils.SiteContext) (*sitedb.SummaryDaily, error) {
	q, err := collectionTx.Site()
	if err != nil {
		return nil, err
	}
	return q.GetSummaryDaily(ctx, sitedb.GetSummaryDailyParams{Date: id})
}

func (s *SummaryDailyCollection) SearchRows(ctx context.Context, collectionTx *siteapiutils.CollectionTx, query *apiutils.Query, siteContext *httputils.SiteContext) ([]*sitedb.SummaryDaily, error) {
	q, err := collectionTx.Site()
	if err != nil {
		return nil, err
	}
	return q.SearchSummaryDaily(ctx, sitedb.SearchSummaryDailyParams{Limit: query.Limit, Skip: query.Skip})
}

// CreateRow builds the model's date again, a day before today in the site's timezone and no more than a year ago
func (s *SummaryDailyCollection) CreateRow(ctx context.Context, collectionTx *siteapiutils.CollectionTx, id *string, siteContext *httputils.SiteContext, model *sitemodels.SiteSummaryDaily) (*sitedb.SummaryDaily, error) {
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
	q, err := collectionTx.Site()
	if err != nil {
		return err
	}
	_, err = q.HardDeleteSummaryDaily(ctx, sitedb.HardDeleteSummaryDailyParams{Date: id})
	return err
}

// The days are only read as JSON for now, so pages get nothing
func (s *SummaryDailyCollection) GetHTMXResponse(*sitemodels.SiteSummaryDaily) templ.Component {
	return templ.NopComponent
}

func (s *SummaryDailyCollection) SearchHTMXResponse([]*sitemodels.SiteSummaryDaily, *apiutils.Query) templ.Component {
	return templ.NopComponent
}

func (s *SummaryDailyCollection) CreateHTMXResponse(*sitemodels.SiteSummaryDaily) templ.Component {
	return templ.NopComponent
}

func (s *SummaryDailyCollection) DeleteHTMXResponse(*sitemodels.SiteSummaryDaily) templ.Component {
	return templ.NopComponent
}

// LocalDate is the date days from now's, in location's time
func LocalDate(now time.Time, location *time.Location, days int) string {
	return now.In(location).AddDate(0, 0, days).Format(time.DateOnly)
}

// RebuildSummaryDaily builds the site's day named date again and keeps it
func RebuildSummaryDaily(ctx context.Context, q *sitedb.Queries, siteContext *httputils.SiteContext, date string, now time.Time) (*sitedb.SummaryDaily, error) {
	day, err := buildSummaryDaily(ctx, q, siteContext, date, now.UnixMilli())
	if err != nil {
		return nil, err
	}
	return q.UpsertSummaryDaily(ctx, sitedb.UpsertSummaryDailyParams{Date: day.Date, Timezone: day.Timezone, StartAt: day.StartAt,
		EndAt: day.EndAt, TimeInRange: day.TimeInRange, TotalDailyDose: day.TotalDailyDose, TotalCarbs: day.TotalCarbs,
		GlucoseMgdlJson: day.GlucoseMgdlJson})
}

// RebuildSummaryDailyWeek deletes all the site's days and periods, as from a timezone they no longer follow, and
// builds the RebuildDays up to yesterday, and the last week and month of what days there are, again
func RebuildSummaryDailyWeek(ctx context.Context, q *sitedb.Queries, siteContext *httputils.SiteContext, now time.Time) error {
	if _, err := q.DeleteAllSummaryDaily(ctx); err != nil {
		return err
	}
	if _, err := q.DeleteAllSummaryPeriods(ctx); err != nil {
		return err
	}
	for days := -RebuildDays; days < 0; days++ {
		if _, err := RebuildSummaryDaily(ctx, q, siteContext, LocalDate(now, siteContext.GetSiteTimezone(), days), now); err != nil {
			return err
		}
	}
	return BuildLastSummaryPeriods(ctx, q, siteContext, now)
}

// SummaryPeriods are what the summary days are put together into: weeks from Monday, and months
var SummaryPeriods = []string{"week", "month"}

// LastPeriodStart is the midnight, in location, that starts the last whole week or month before now's
func LastPeriodStart(period string, now time.Time, location *time.Location) time.Time {
	local := now.In(location)
	if period == "week" {
		return time.Date(local.Year(), local.Month(), local.Day()-(int(local.Weekday())+6)%7-7, 0, 0, 0, 0, location)
	}
	return time.Date(local.Year(), local.Month()-1, 1, 0, 0, 0, 0, location)
}

// LastSummaryPeriod is the site's last whole week or month, and whether it was built, built from its days if not
func LastSummaryPeriod(ctx context.Context, q *sitedb.Queries, siteContext *httputils.SiteContext, period string, now time.Time) (*sitedb.SummaryPeriod, bool, error) {
	start := LastPeriodStart(period, now, siteContext.GetSiteTimezone())
	date := start.Format(time.DateOnly)
	row, err := q.GetSummaryPeriod(ctx, sitedb.GetSummaryPeriodParams{Period: period, Date: date})
	if !errors.Is(err, sql.ErrNoRows) {
		return row, err == nil, err
	}
	end := start.AddDate(0, 1, 0)
	if period == "week" {
		end = start.AddDate(0, 0, 7)
	}
	last := end.AddDate(0, 0, -1).Format(time.DateOnly)
	days, err := q.SearchSummaryDaily(ctx, sitedb.SearchSummaryDailyParams{DateGTE: &date, DateLTE: &last, Limit: 31})
	if err != nil {
		return nil, false, err
	}
	return sitemodels.NewSummaryPeriod(period, date, start.Location().String(), start.UnixMilli(), end.UnixMilli(), days), false, nil
}

// BuildLastSummaryPeriods builds and keeps the site's last whole week and month, when they are not yet built
func BuildLastSummaryPeriods(ctx context.Context, q *sitedb.Queries, siteContext *httputils.SiteContext, now time.Time) error {
	for _, period := range SummaryPeriods {
		row, built, err := LastSummaryPeriod(ctx, q, siteContext, period, now)
		if err != nil {
			return err
		}
		if built {
			continue
		}
		if _, err := q.UpsertSummaryPeriod(ctx, sitedb.UpsertSummaryPeriodParams{Period: row.Period, Date: row.Date, Timezone: row.Timezone,
			StartAt: row.StartAt, EndAt: row.EndAt, Days: row.Days, TimeInRange: row.TimeInRange, MeanDailyDose: row.MeanDailyDose,
			MeanDailyCarbs: row.MeanDailyCarbs}); err != nil {
			return err
		}
	}
	return nil
}

// buildSummaryDaily is the site's day named date, from midnight in its timezone, built from its clean readings,
// insulin, basal and food up to until, so a day not yet over is its part so far. Its time in range is from the site's
// range low to its range high.
func buildSummaryDaily(ctx context.Context, q *sitedb.Queries, siteContext *httputils.SiteContext, date string, until int64) (*sitedb.SummaryDaily, error) {
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
		glucose.CleanReadings(rows), doses, food, start, minutes), nil
}
