package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"slices"
	"time"

	"nighttroop.t1d.nz/data/retention"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/utils/glucose"
)

// dayMs is a UTC day, which a cold day is
const dayMs = int64(24 * time.Hour / time.Millisecond)

// sealSite seals the site's days that had a row written since since, in ms, 0 for all of them, into cold_days, each
// kind of data as a series. A site never sealed has all its days sealed.
//
// Glucose: each day that ended, the newest CGM reading of the best source in each 5 minutes. When the server deletes
// rows (RETENTION_DELETE_ROWS), the CGM readings older than its hot days are then deleted, but only those their cold
// day has.
//
// Events (what a person did: a meal, a dose, a meter's reading, a note, a target, a switch): each day that ended, as
// one list, see sealEventsDay.
//
// Treatments (the basal rate, insulin and carbs every 5 minutes): each day that ended, see sealTreatmentDays.
//
// What a loop's phone reports (its IOB and COB, the pump's reservoir, the batteries): each day up to now, today too,
// as device reports are thinned from an hour old (truncateSite), so it runs before they are, and a day keeps the
// values it has.
//
// See docs/retention.md for the other tables, the summaries and the truncation this is to take the place of.
func (db *ControlDB) sealSite(ctx context.Context, site *SiteDB, since int64, now time.Time) error {
	q, today := site.Queries(), now.UTC().Truncate(24*time.Hour)
	hot := retention.HotSince(now, db.ServerConfig.RetentionHotDays)
	if sealed, err := q.HasColdDays(ctx); err != nil {
		return fmt.Errorf("cannot seal site %s: %w", site.Subdomain, err)
	} else if !sealed {
		since = 0
	}
	sealed := func(day int64, err error) error {
		if err == nil {
			return nil
		}
		return fmt.Errorf("cannot seal %s of site %s: %w", time.UnixMilli(day*dayMs).UTC().Format(time.DateOnly), site.Subdomain, err)
	}
	days, err := q.CGMDaysWrittenSince(ctx, sitedb.CGMDaysWrittenSinceParams{Since: since, Before: today.UnixMilli()})
	if err != nil {
		return fmt.Errorf("cannot seal site %s: %w", site.Subdomain, err)
	}
	for _, day := range days {
		// A day whose rows may be deleted keeps what it has, and takes what was written since
		if err := sealed(day, sealGlucoseDay(ctx, q, day*dayMs, db.ServerConfig.RetentionDeleteRows && day*dayMs < hot.UnixMilli())); err != nil {
			return err
		}
	}
	if days, err = q.DeviceReportDaysWrittenSince(ctx, sitedb.DeviceReportDaysWrittenSinceParams{Since: since, Before: now.UnixMilli() + 1}); err != nil {
		return fmt.Errorf("cannot seal site %s: %w", site.Subdomain, err)
	}
	for _, day := range days {
		if err := sealed(day, sealDeviceDay(ctx, q, day*dayMs)); err != nil {
			return err
		}
	}
	if days, err = q.EventDaysWrittenSince(ctx, sitedb.EventDaysWrittenSinceParams{Since: since, Before: today.UnixMilli()}); err != nil {
		return fmt.Errorf("cannot seal site %s: %w", site.Subdomain, err)
	}
	// The days that just ended too, as a day with no event of its own still has what was in force at its start
	for _, day := range lastDays(days, today) {
		if err := sealed(day, sealEventsDay(ctx, q, day*dayMs)); err != nil {
			return err
		}
	}
	if err := db.sealTreatmentDays(ctx, site, since, today); err != nil {
		return fmt.Errorf("cannot seal the treatments of site %s: %w", site.Subdomain, err)
	}
	if db.ServerConfig.RetentionDeleteRows {
		if _, err := q.DeleteSealedCGMReadings(ctx, sitedb.DeleteSealedCGMReadingsParams{Before: hot.UnixMilli()}); err != nil {
			return fmt.Errorf("cannot delete the sealed readings of site %s: %w", site.Subdomain, err)
		}
	}
	return nil
}

// DaySeries is a day of one series, a value a step, NaN where there is none
type DaySeries struct {
	Series retention.Series
	Values []float64
}

// TreatmentDaySeries builds the treatment series (the basal rate, insulin and carbs) of the site's UTC day that starts
// at start, in ms, from its rows. It is set by the package that reads profiles and temp basals, sitedata, which this
// package cannot import, and nil in a program without it, which then seals no treatment series.
var TreatmentDaySeries func(ctx context.Context, q *sitedb.Queries, site *controldb.Site, start int64) ([]DaySeries, error)

// daysBack is how many of the days that just ended are sealed though nothing was written in them, as a day has a
// basal rate from its profile alone, and what was in force when it started
const daysBack = 2

// lastDays are days, as days since 1970, with the last daysBack that ended before today
func lastDays(days []int64, today time.Time) []int64 {
	for back := int64(1); back <= daysBack; back++ {
		if day := today.UnixMilli()/dayMs - back; !slices.Contains(days, day) {
			days = append(days, day)
		}
	}
	return days
}

// sealTreatmentDays seals the treatment series of the days that ended and had a treatment written since since, and of
// the last days that ended, from their rows alone, as none of their rows is deleted yet. A series with no value in a
// day is written only to empty a day that had some.
func (db *ControlDB) sealTreatmentDays(ctx context.Context, site *SiteDB, since int64, today time.Time) error {
	if TreatmentDaySeries == nil {
		return nil
	}
	q := site.Queries()
	days, err := q.TreatmentDaysWrittenSince(ctx, sitedb.TreatmentDaysWrittenSinceParams{Since: since, Before: today.UnixMilli()})
	if err != nil {
		return err
	}
	days = lastDays(days, today)
	model, err := db.Queries.GetSiteInAnyState(ctx, controldb.GetSiteInAnyStateParams{ID: site.Subdomain})
	if err != nil {
		return err
	}
	for _, day := range days {
		built, err := TreatmentDaySeries(ctx, q, model, day*dayMs)
		if err != nil {
			return err
		}
		date := time.UnixMilli(day * dayMs).UTC().Format(time.DateOnly)
		for _, series := range built {
			if !slices.ContainsFunc(series.Values, func(value float64) bool { return !math.IsNaN(value) }) {
				if _, err := q.GetColdDay(ctx, sitedb.GetColdDayParams{Kind: series.Series.Name, Date: date}); err != nil {
					continue
				}
			}
			if err := writeSeriesDay(ctx, q, series.Series, date, series.Values); err != nil {
				return err
			}
		}
	}
	return nil
}

// point is a value of a series at a time, in ms
type point struct {
	at    int64
	value float64
}

// sealSeriesDay writes one series of the UTC day that starts at start, in ms: each point, the oldest first, takes the
// step it is in, so a step keeps its newest. With merge, the day's rows may be gone or thinned, so it starts from the
// day as it was sealed before, with the steps at cleared emptied, and takes the points. It reads the day back before
// it is kept, so a day is never sealed as what cannot be read, or with other values.
func sealSeriesDay(ctx context.Context, q *sitedb.Queries, series retention.Series, start int64, merge bool, cleared []int64, points []point) error {
	date, values := time.UnixMilli(start).UTC().Format(time.DateOnly), series.Empty()
	if merge {
		switch sealed, err := q.GetColdDay(ctx, sitedb.GetColdDayParams{Kind: series.Name, Date: date}); {
		case err == nil:
			if values, err = retention.UnpackGrid(sealed.Body); err != nil {
				return err
			}
		case !errors.Is(err, sql.ErrNoRows):
			return err
		}
	}
	step := func(at int64) int64 { return min(max((at-start)/series.Step.Milliseconds(), 0), int64(len(values))-1) }
	for _, at := range cleared {
		values[step(at)] = math.NaN()
	}
	for _, p := range points {
		values[step(p.at)] = math.Round(p.value*series.Scale) / series.Scale
	}
	return writeSeriesDay(ctx, q, series, date, values)
}

// writeSeriesDay keeps a day of a series, each value to the series' scale. It reads the day back before it is kept,
// so a day is never sealed as what cannot be read, or with other values.
func writeSeriesDay(ctx context.Context, q *sitedb.Queries, series retention.Series, date string, values []float64) error {
	for i, value := range values {
		values[i] = math.Round(value*series.Scale) / series.Scale
	}
	packed := retention.PackGrid(series, values)
	read, err := retention.UnpackGrid(packed)
	if err != nil || !slices.EqualFunc(values, read, func(a, b float64) bool { return a == b || (math.IsNaN(a) && math.IsNaN(b)) }) {
		return fmt.Errorf("the sealed day of %s does not read back as it was written: %v", series.Name, err)
	}
	return q.UpsertColdDay(ctx, sitedb.UpsertColdDayParams{Kind: series.Name, Date: date, Body: packed})
}

// sealGlucoseDay writes the glucose series of the UTC day that starts at start, in ms, from its CGM readings: the
// newest reading of the best source in each step. With merge, a reading written since takes its step, and a deleted
// one clears its step.
func sealGlucoseDay(ctx context.Context, q *sitedb.Queries, start int64, merge bool) error {
	rows, err := q.CGMReadingsOfDay(ctx, sitedb.CGMReadingsOfDayParams{From: start, To: start + dayMs})
	if err != nil {
		return err
	}
	var cleared []int64
	for _, row := range rows {
		if row.DeletedAt != nil {
			cleared = append(cleared, row.OccurredAt)
		}
	}
	// Newest first, so they are turned to the oldest first
	live := slices.DeleteFunc(slices.Clone(rows), func(row *sitedb.GlucoseReading) bool { return row.DeletedAt != nil })
	var points []point
	for _, reading := range slices.Backward(glucose.CleanReadings(live)) {
		points = append(points, point{reading.Row.OccurredAt, reading.Row.Mgdl})
	}
	return sealSeriesDay(ctx, q, retention.Glucose, start, merge, cleared, points)
}

// deviceSeries are the series a device report's numbers are sealed into, each with the number of a report it takes.
// The phone's battery is only the loop's phone's, as a carer's phone reports its own too.
var deviceSeries = []struct {
	series retention.Series
	value  func(*sitedb.DeviceReportNumbersOfDayRow) *float64
}{
	{retention.IOB, func(row *sitedb.DeviceReportNumbersOfDayRow) *float64 { return row.LoopIob }},
	{retention.IOBBasal, func(row *sitedb.DeviceReportNumbersOfDayRow) *float64 { return row.LoopIobBasal }},
	{retention.COB, func(row *sitedb.DeviceReportNumbersOfDayRow) *float64 { return row.LoopCob }},
	{retention.PumpReservoir, func(row *sitedb.DeviceReportNumbersOfDayRow) *float64 { return row.PumpReservoir }},
	{retention.PumpBattery, func(row *sitedb.DeviceReportNumbersOfDayRow) *float64 { return row.PumpBatteryPercent }},
	{retention.PhoneBattery, func(row *sitedb.DeviceReportNumbersOfDayRow) *float64 {
		if row.LoopName == nil {
			return nil
		}
		return row.DeviceBatteryPercent
	}},
}

// sealDeviceDay writes the device series of the UTC day that starts at start, in ms, from its device reports: each
// series a step's newest report with that number. A day keeps the values it has, as its reports are thinned as they
// age and deleted after a month, and a series no report has a number for is not written.
func sealDeviceDay(ctx context.Context, q *sitedb.Queries, start int64) error {
	rows, err := q.DeviceReportNumbersOfDay(ctx, sitedb.DeviceReportNumbersOfDayParams{From: start, To: start + dayMs})
	if err != nil {
		return err
	}
	for _, device := range deviceSeries {
		var cleared []int64
		var points []point
		for _, row := range rows {
			switch value := device.value(row); {
			case value == nil:
			case row.DeletedAt != nil:
				cleared = append(cleared, row.OccurredAt)
			default:
				points = append(points, point{row.OccurredAt, *value})
			}
		}
		if len(points) == 0 && len(cleared) == 0 {
			continue
		}
		if err := sealSeriesDay(ctx, q, device.series, start, true, cleared, points); err != nil {
			return err
		}
	}
	return nil
}
