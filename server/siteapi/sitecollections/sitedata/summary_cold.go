package sitedata

import (
	"context"
	"slices"
	"time"

	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/retention"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/models/sitemodels"
	"nighttroop.t1d.nz/utils/httputils"
)

// coldBefore is the time, in ms, that a day must end by for its summary to be built from its cold days: the start of
// the UTC day before now's, as a UTC day is packaged in the run after it ends. A later day is built from its rows.
func coldBefore(now time.Time) int64 {
	return now.UTC().Truncate(24 * time.Hour).Add(-24 * time.Hour).UnixMilli()
}

// packagedMargin is how long before a site was last backed up, which packages it first, its data may have been
// written and not be packaged: longer than a backup of it takes
const packagedMargin = time.Hour

// builtSummaryDays are the site's days from the date first to last, as 2026-01-31, up to today, at most a year and a fortnight of
// them, oldest first, from midnight in its timezone, each built as it is read and not kept. A day that ended before
// coldBefore is built from its cold days, which are there when its rows are deleted. A later day, and today so far,
// is built from its rows up to now, as is a day with data written since the site was last packaged while its rows
// are all there, so a day is never behind its data. A day with no data is left out, but, with gaps, one after
// a day with data, so the days read on with none missing.
func builtSummaryDays(ctx context.Context, q *sitedb.Queries, siteContext *httputils.SiteContext, first, last string, now time.Time, gaps bool) ([]*sitedb.SummaryPeriod, error) {
	location := siteContext.GetSiteTimezone()
	from, err := time.ParseInLocation(time.DateOnly, first, location)
	if err != nil {
		return nil, err
	}
	// midnights are the starts of the days, and of the day after the last
	midnights, cold := []time.Time{from}, 0
	for day := from; len(midnights) <= summaryDailyMaxDays+14 && day.Format(time.DateOnly) <= last && !day.After(now); {
		day = time.Date(day.Year(), day.Month(), day.Day()+1, 0, 0, 0, 0, location)
		midnights = append(midnights, day)
		if day.UnixMilli() <= coldBefore(now) {
			cold = len(midnights) - 1
		}
	}
	start, end := from.UnixMilli(), midnights[cold].UnixMilli()
	// The series of all the cold days, read once
	var series sitemodels.DaySeries
	for of, into := range map[retention.Series]*[]float64{retention.Glucose: &series.Glucose, retention.IOB: &series.IOB, retention.IOBBasal: &series.IOBBasal,
		retention.COB: &series.COB, retention.BasalRate: &series.BasalRate, retention.InsulinBolus: &series.Bolus,
		retention.InsulinBasal: &series.Injected, retention.Carbs: &series.Carbs} {
		if *into, err = db.ColdSeries(ctx, q, of, start, end); err != nil {
			return nil, err
		}
	}
	// The meter's readings are counted from their rows, which are never deleted
	meter, err := q.SearchGlucoseReadings(ctx, sitedb.SearchGlucoseReadingsParams{DateGTE: &start, DateLT: &end, TypeEq: new("mbg"), Limit: maxRows})
	if err != nil {
		return nil, err
	}
	// The days with data that may not be packaged yet, which are built from rows while the rows are there: all of them
	// where the server deletes none, and else those of the hot days
	since, rowsFrom := int64(0), int64(0)
	if siteContext.SiteModel != nil && siteContext.SiteModel.BackedUpAt != nil {
		since = *siteContext.SiteModel.BackedUpAt - packagedMargin.Milliseconds()
	}
	if siteContext.ControlDB != nil && siteContext.ControlDB.ServerConfig.RetentionDeleteRows {
		rowsFrom = retention.HotSince(now, siteContext.ControlDB.ServerConfig.RetentionHotDays).UnixMilli()
	}
	written, err := q.SummaryDataWrittenSince(ctx, sitedb.SummaryDataWrittenSinceParams{Since: since, Before: end})
	if err != nil {
		return nil, err
	}
	low, high := siteContext.GetSiteRangeMgdl()
	var days []*sitedb.SummaryPeriod
	for i, midnight := range midnights[:len(midnights)-1] {
		dayStart, dayEnd := midnight.UnixMilli(), midnights[i+1].UnixMilli()
		var day *sitedb.SummaryPeriod
		notPackaged := dayStart >= rowsFrom && slices.ContainsFunc(written, func(at int64) bool { return at >= dayStart && at < dayEnd })
		if i >= cold || notPackaged {
			if day, err = buildSummaryDaily(ctx, q, siteContext, midnight.Format(time.DateOnly), now.UnixMilli()); err != nil {
				return nil, err
			}
		} else {
			a, b := (dayStart-start)/sitemodels.SummaryDailySlotMs, (dayEnd-start)/sitemodels.SummaryDailySlotMs
			readings := 0
			for _, reading := range meter {
				if reading.OccurredAt >= dayStart && reading.OccurredAt < dayEnd {
					readings++
				}
			}
			day = sitemodels.NewSummaryDailyFromSeries(midnight.Format(time.DateOnly), location.String(), dayStart, dayEnd, float64(low), float64(high), sitemodels.DaySeries{
				Glucose: series.Glucose[a:b], IOB: series.IOB[a:b], IOBBasal: series.IOBBasal[a:b], COB: series.COB[a:b],
				BasalRate: series.BasalRate[a:b], Bolus: series.Bolus[a:b], Injected: series.Injected[a:b], Carbs: series.Carbs[a:b]}, int64(readings))
		}
		if day.Days > 0 || gaps && len(days) > 0 || i == len(midnights)-2 && midnights[i+1].After(now) {
			days = append(days, day)
		}
	}
	return days, nil
}

// BuiltSummaries are what the Summaries page of the month starting first shows, each built as it is read from the
// site's days and newest first: by "day", the month's days with data that are over; by "week" and "fortnight", those
// that are over with a day in the month and a day with data; and by "month", the months of its year that are over
// and have a day with data.
func BuiltSummaries(ctx context.Context, q *sitedb.Queries, siteContext *httputils.SiteContext, first, now time.Time) (map[string][]*sitedb.SummaryPeriod, error) {
	location, weekStart := siteContext.GetSiteTimezone(), siteContext.GetSiteWeekStartDay()
	year, monthEnd := time.Date(first.Year(), 1, 1, 0, 0, 0, 0, location), first.AddDate(0, 1, 0)
	// All the days the page's periods are of, read once: from the fortnight the year starts in to the year's end
	from := PeriodStart("fortnight", year, location, weekStart)
	days, err := builtSummaryDays(ctx, q, siteContext, from.Format(time.DateOnly), year.AddDate(1, 0, -1).Format(time.DateOnly), now, false)
	if err != nil {
		return nil, err
	}
	days = slices.DeleteFunc(days, func(day *sitedb.SummaryPeriod) bool { return day.EndAt > now.UnixMilli() })
	within := func(start, end time.Time) []*sitedb.SummaryPeriod {
		var of []*sitedb.SummaryPeriod
		for _, day := range days {
			if day.StartAt >= start.UnixMilli() && day.StartAt < end.UnixMilli() {
				of = append(of, day)
			}
		}
		return of
	}
	built := map[string][]*sitedb.SummaryPeriod{"day": within(first, monthEnd)}
	slices.Reverse(built["day"])
	// add is the period from start, when it is over and has a day with data
	add := func(period string, start time.Time) {
		end := PeriodEnd(period, start)
		if of := within(start, end); len(of) > 0 && !end.After(now) {
			row := sitemodels.NewSummaryPeriod(period, start.Format(time.DateOnly), location.String(), start.UnixMilli(), end.UnixMilli(), now.UnixMilli(), of)
			built[period] = append([]*sitedb.SummaryPeriod{row}, built[period]...)
		}
	}
	for _, period := range []string{"week", "fortnight"} {
		for start := PeriodStart(period, first, location, weekStart); start.Before(monthEnd); start = PeriodEnd(period, start) {
			add(period, start)
		}
	}
	for month := year; month.Year() == year.Year(); month = month.AddDate(0, 1, 0) {
		add("month", month)
	}
	return built, nil
}
