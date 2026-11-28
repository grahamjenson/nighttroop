package db

import (
	"bytes"
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

// packageSite packages the site's days that had a row written since since, in ms, 0 for all of them, into cold_days, each
// kind of data as a series. A site never packaged has all its days packaged.
//
// Glucose: each day that ended, the newest CGM reading of the best source in each 5 minutes. When the server deletes
// rows (RETENTION_DELETE_ROWS), the CGM readings older than its hot days are then deleted, but only those their cold
// day has.
//
// Events (what a person did: a meal, a dose, a meter's reading, a note, a target, a switch): each day that ended, as
// one list, see packageEventsDay.
//
// Treatments (the basal rate, insulin and carbs every 5 minutes): each day that ended, see packageTreatmentDays.
//
// What a loop's phone reports (its IOB and COB, the pump's reservoir, the batteries): each day up to now, today too,
// as device reports are thinned from an hour old (truncateSite), so it runs before they are, and a day keeps the
// values it has.
//
// See docs/retention.md for the other tables, the summaries and the truncation this is to take the place of.
func (db *ControlDB) packageSite(ctx context.Context, site *SiteDB, since int64, now time.Time) error {
	q, today := site.Queries(), now.UTC().Truncate(24*time.Hour)
	hot := retention.HotSince(now, db.ServerConfig.RetentionHotDays)
	if packaged, err := q.HasColdDays(ctx); err != nil {
		return fmt.Errorf("cannot package site %s: %w", site.Subdomain, err)
	} else if !packaged {
		since = 0
	}
	packaged := func(day int64, err error) error {
		if err == nil {
			return nil
		}
		return fmt.Errorf("cannot package %s of site %s: %w", time.UnixMilli(day*dayMs).UTC().Format(time.DateOnly), site.Subdomain, err)
	}
	// The days marked to be packaged again, as a record was deleted from them for good or moved away: those whose rows
	// are all there, so the hot days where the server deletes rows
	staleFrom := ""
	if db.ServerConfig.RetentionDeleteRows {
		staleFrom = hot.Format(time.DateOnly)
	}
	staleDates, err := q.StaleColdDays(ctx, sitedb.StaleColdDaysParams{FromDate: staleFrom})
	if err != nil {
		return fmt.Errorf("cannot package site %s: %w", site.Subdomain, err)
	}
	var stale []int64
	for _, date := range staleDates {
		if at, err := time.Parse(time.DateOnly, date); err == nil && at.Before(today) {
			stale = append(stale, at.UnixMilli()/dayMs)
		}
	}
	days, err := q.CGMDaysWrittenSince(ctx, sitedb.CGMDaysWrittenSinceParams{Since: since, Before: today.UnixMilli()})
	if err != nil {
		return fmt.Errorf("cannot package site %s: %w", site.Subdomain, err)
	}
	for _, day := range slices.Compact(slices.Sorted(slices.Values(append(days, stale...)))) {
		// A day whose rows may be deleted keeps what it has, and takes what was written since
		if err := packaged(day, packageGlucoseDay(ctx, q, day*dayMs, db.ServerConfig.RetentionDeleteRows && day*dayMs < hot.UnixMilli())); err != nil {
			return err
		}
	}
	if days, err = q.DeviceReportDaysWrittenSince(ctx, sitedb.DeviceReportDaysWrittenSinceParams{Since: since, Before: now.UnixMilli() + 1}); err != nil {
		return fmt.Errorf("cannot package site %s: %w", site.Subdomain, err)
	}
	for _, day := range days {
		if err := packaged(day, packageDeviceDay(ctx, q, day*dayMs)); err != nil {
			return err
		}
	}
	if err := db.packageTreatmentDays(ctx, site, since, now, stale); err != nil {
		return fmt.Errorf("cannot package the treatments of site %s: %w", site.Subdomain, err)
	}
	if err := q.ClearStaleColdDays(ctx, sitedb.ClearStaleColdDaysParams{FromDate: staleFrom}); err != nil {
		return fmt.Errorf("cannot package site %s: %w", site.Subdomain, err)
	}
	if db.ServerConfig.RetentionDeleteRows {
		if _, err := q.DeletePackagedCGMReadings(ctx, sitedb.DeletePackagedCGMReadingsParams{Before: hot.UnixMilli()}); err != nil {
			return fmt.Errorf("cannot delete the packaged readings of site %s: %w", site.Subdomain, err)
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
// package cannot import, and nil in a program without it, which then packages no treatment series.
var TreatmentDaySeries func(ctx context.Context, q *sitedb.Queries, site *controldb.Site, start int64) ([]DaySeries, error)

// daysBack is how many of the days that just ended are packaged though nothing was written in them, as a day has a
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

// PackagedSettle is how long a day older than the hot days must go without being packaged again before its treatment
// rows are deleted, and from when the day is no longer packaged again: a backfill writes a day's rows over some runs,
// and a day is packaged again while they come. It is a variable for a test to set.
var PackagedSettle = 24 * time.Hour

// packageTreatmentDays packages the treatment series and then the events of the days that ended and had a treatment or
// an event written since since, of the stale days, and of the last days that ended, from their rows alone. The events are written last,
// as a day's events say the day was packaged: when the server deletes rows, the rows of a day older than the hot days
// are deleted once its events were packaged after them and have then settled (see deletePackagedTreatments), and such
// a day is never packaged again, as its rows are not all there. A row written for it later stays a row.
//
// A series with no value in a day is written only to empty a day that had some.
func (db *ControlDB) packageTreatmentDays(ctx context.Context, site *SiteDB, since int64, now time.Time, stale []int64) error {
	q, today := site.Queries(), now.UTC().Truncate(24*time.Hour)
	hot, settled := retention.HotSince(now, db.ServerConfig.RetentionHotDays).UnixMilli(), now.Add(-PackagedSettle).UnixMilli()
	days, err := q.EventDaysWrittenSince(ctx, sitedb.EventDaysWrittenSinceParams{Since: since, Before: today.UnixMilli()})
	if err != nil {
		return err
	}
	treatmentDays, err := q.TreatmentDaysWrittenSince(ctx, sitedb.TreatmentDaysWrittenSinceParams{Since: since, Before: today.UnixMilli()})
	if err != nil {
		return err
	}
	for _, day := range append(treatmentDays, stale...) {
		if !slices.Contains(days, day) {
			days = append(days, day)
		}
	}
	var model *controldb.Site
	if TreatmentDaySeries != nil {
		if model, err = db.Queries.GetSiteInAnyState(ctx, controldb.GetSiteInAnyStateParams{ID: site.Subdomain}); err != nil {
			return err
		}
	}
	for _, day := range lastDays(days, today) {
		date := time.UnixMilli(day * dayMs).UTC().Format(time.DateOnly)
		if db.ServerConfig.RetentionDeleteRows && day*dayMs < hot {
			if events, err := q.GetColdDay(ctx, sitedb.GetColdDayParams{Kind: retention.EventsKind, Date: date}); err == nil && events.UpdatedAt <= settled {
				continue
			}
		}
		treated := false
		if TreatmentDaySeries != nil {
			built, err := TreatmentDaySeries(ctx, q, model, day*dayMs)
			if err != nil {
				return fmt.Errorf("%s: %w", date, err)
			}
			for _, series := range built {
				if !slices.ContainsFunc(series.Values, func(value float64) bool { return !math.IsNaN(value) }) {
					if _, err := q.GetColdDay(ctx, sitedb.GetColdDayParams{Kind: series.Series.Name, Date: date}); err != nil {
						continue
					}
				}
				treated = true
				if err := writeSeriesDay(ctx, q, series.Series, date, series.Values); err != nil {
					return fmt.Errorf("%s: %w", date, err)
				}
			}
		}
		// A day with a treatment series has its events written though it has none, to say it was packaged
		if err := packageEventsDay(ctx, q, day*dayMs, treated, db.ServerConfig.RetentionDeleteRows && day*dayMs < hot); err != nil {
			return fmt.Errorf("%s: %w", date, err)
		}
	}
	if !db.ServerConfig.RetentionDeleteRows || TreatmentDaySeries == nil {
		return nil
	}
	return deletePackagedTreatments(ctx, q, hot, settled)
}

// deletePackagedTreatments deletes the treatments and events from before before, in ms, whose day's events were
// packaged after they were last written, and not since settled, so none is deleted that its cold days do not have:
// the treatments first, then the events no treatment still refers to. A profile switch that may still be in force is
// kept, as the basal of the days that are rows is read from it. A meter's reading is kept (see
// DeletePackagedCGMReadings).
func deletePackagedTreatments(ctx context.Context, q *sitedb.Queries, before, settled int64) error {
	_, err := q.DeletePackagedEventGroups(ctx, sitedb.DeletePackagedEventGroupsParams{Before: before, SettledBefore: settled})
	deletes := []func() (int64, error){
		func() (int64, error) {
			return q.DeletePackagedInsulinDoses(ctx, sitedb.DeletePackagedInsulinDosesParams{Before: before, SettledBefore: settled})
		},
		func() (int64, error) {
			return q.DeletePackagedFoodIntakes(ctx, sitedb.DeletePackagedFoodIntakesParams{Before: before, SettledBefore: settled})
		},
		func() (int64, error) {
			return q.DeletePackagedTempBasals(ctx, sitedb.DeletePackagedTempBasalsParams{Before: before, SettledBefore: settled})
		},
		func() (int64, error) {
			return q.DeletePackagedTempTargets(ctx, sitedb.DeletePackagedTempTargetsParams{Before: before})
		},
		func() (int64, error) {
			return q.DeletePackagedActivities(ctx, sitedb.DeletePackagedActivitiesParams{Before: before})
		},
		func() (int64, error) {
			return q.DeletePackagedNotificationEvents(ctx, sitedb.DeletePackagedNotificationEventsParams{Before: before})
		},
		func() (int64, error) {
			return q.DeletePackagedProfileSwitches(ctx, sitedb.DeletePackagedProfileSwitchesParams{Before: before, SettledBefore: settled})
		},
	}
	for _, deleteRows := range deletes {
		if err == nil {
			_, err = deleteRows()
		}
	}
	return err
}

// point is a value of a series at a time, in ms
type point struct {
	at    int64
	value float64
}

// packageSeriesDay writes one series of the UTC day that starts at start, in ms: each point, the oldest first, takes the
// step it is in, so a step keeps its newest. With merge, the day's rows may be gone or thinned, so it starts from the
// day as it was packaged before, with the steps at cleared emptied, and takes the points. It reads the day back before
// it is kept, so a day is never packaged as what cannot be read, or with other values.
func packageSeriesDay(ctx context.Context, q *sitedb.Queries, series retention.Series, start int64, merge bool, cleared []int64, points []point) error {
	date, values := time.UnixMilli(start).UTC().Format(time.DateOnly), series.Empty()
	if merge {
		switch packaged, err := q.GetColdDay(ctx, sitedb.GetColdDayParams{Kind: series.Name, Date: date}); {
		case err == nil:
			if values, err = retention.UnpackGrid(packaged.Body); err != nil {
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
// so a day is never packaged as what cannot be read, or with other values.
func writeSeriesDay(ctx context.Context, q *sitedb.Queries, series retention.Series, date string, values []float64) error {
	for i, value := range values {
		values[i] = math.Round(value*series.Scale) / series.Scale
	}
	packed := retention.PackGrid(series, values)
	read, err := retention.UnpackGrid(packed)
	if err != nil || !slices.EqualFunc(values, read, func(a, b float64) bool { return a == b || (math.IsNaN(a) && math.IsNaN(b)) }) {
		return fmt.Errorf("the packaged day of %s does not read back as it was written: %v", series.Name, err)
	}
	return upsertColdDay(ctx, q, series.Name, date, packed)
}

// upsertColdDay keeps the cold day of kind for the UTC date. One that is the same, packaged after every row of its day
// was last written and not marked to be packaged again, is not written again, so it keeps when it was packaged and the
// wait before its rows are deleted (PackagedSettle) goes on. Any other is written, and says it was packaged now, after
// each of its rows, as a row is deleted only when its cold day was packaged after it was last written.
func upsertColdDay(ctx context.Context, q *sitedb.Queries, kind, date string, body []byte) error {
	day, err := time.Parse(time.DateOnly, date)
	if err != nil {
		return err
	}
	switch packaged, err := q.GetColdDay(ctx, sitedb.GetColdDayParams{Kind: kind, Date: date}); {
	case errors.Is(err, sql.ErrNoRows):
	case err != nil:
		return err
	case packaged.UpdatedAt > 0 && bytes.Equal(packaged.Body, body):
		written, err := q.LastWrittenOfDay(ctx, sitedb.LastWrittenOfDayParams{From: day.UnixMilli(), To: day.UnixMilli() + dayMs})
		if err != nil || packaged.UpdatedAt >= written {
			return err
		}
	}
	return q.UpsertColdDay(ctx, sitedb.UpsertColdDayParams{Kind: kind, Date: date, Body: body})
}

// packageGlucoseDay writes the glucose series of the UTC day that starts at start, in ms, from its CGM readings: the
// newest reading of the best source in each step. With merge, a reading written since takes its step, and a deleted
// one clears its step.
func packageGlucoseDay(ctx context.Context, q *sitedb.Queries, start int64, merge bool) error {
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
	return packageSeriesDay(ctx, q, retention.Glucose, start, merge, cleared, points)
}

// deviceSeries are the series a device report's numbers are packaged into, each with the number of a report it takes.
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

// packageDeviceDay writes the device series of the UTC day that starts at start, in ms, from its device reports: each
// series a step's newest report with that number. A day keeps the values it has, as its reports are thinned as they
// age and deleted after a month, and a series no report has a number for is not written.
func packageDeviceDay(ctx context.Context, q *sitedb.Queries, start int64) error {
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
		if err := packageSeriesDay(ctx, q, device.series, start, true, cleared, points); err != nil {
			return err
		}
	}
	return nil
}
