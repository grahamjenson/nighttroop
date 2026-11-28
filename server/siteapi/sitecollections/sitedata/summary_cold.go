package sitedata

import (
	"context"
	"time"

	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/retention"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/models/sitemodels"
	"nighttroop.t1d.nz/utils/httputils"
)

// coldSummaryDaily is the site's day named date, from midnight in its timezone, built from its cold days and not from
// its rows, so a day whose rows are deleted has its summary: what buildSummaryDaily builds (see
// sitemodels.NewSummaryDailyFromSeries for what differs). The meter's readings are counted from their rows, which are
// never deleted. A day that is not packaged, as today, has no values.
func coldSummaryDaily(ctx context.Context, q *sitedb.Queries, siteContext *httputils.SiteContext, date string) (*sitedb.SummaryPeriod, error) {
	location := siteContext.GetSiteTimezone()
	midnight, err := time.ParseInLocation(time.DateOnly, date, location)
	if err != nil {
		return nil, err
	}
	start, end := midnight.UnixMilli(), midnight.AddDate(0, 0, 1).UnixMilli()
	var series sitemodels.DaySeries
	for of, into := range map[retention.Series]*[]float64{retention.Glucose: &series.Glucose, retention.IOB: &series.IOB, retention.IOBBasal: &series.IOBBasal,
		retention.COB: &series.COB, retention.BasalRate: &series.BasalRate, retention.InsulinBolus: &series.Bolus,
		retention.InsulinBasal: &series.Injected, retention.Carbs: &series.Carbs} {
		if *into, err = db.ColdSeries(ctx, q, of, start, end); err != nil {
			return nil, err
		}
	}
	meter, err := q.SearchGlucoseReadings(ctx, sitedb.SearchGlucoseReadingsParams{DateGTE: &start, DateLT: &end, TypeEq: new("mbg"), Limit: maxRows})
	if err != nil {
		return nil, err
	}
	low, high := siteContext.GetSiteRangeMgdl()
	return sitemodels.NewSummaryDailyFromSeries(date, location.String(), start, end, float64(low), float64(high), series, int64(len(meter))), nil
}
