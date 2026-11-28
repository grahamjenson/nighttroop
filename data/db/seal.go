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
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/utils/glucose"
)

// dayMs is a UTC day, which a cold day is
const dayMs = int64(24 * time.Hour / time.Millisecond)

// sealSite seals the site's days that ended and had a CGM reading written since since, in ms, 0 for all of them:
// each day's glucose is written to cold_days as a series, the newest reading of the best source in each 5 minutes.
// A site never sealed has all its days sealed. Then, when the server deletes rows (RETENTION_DELETE_ROWS), the CGM
// readings older than its hot days are deleted, but only those their cold day has. Only glucose is sealed yet: see
// docs/retention.md for the other tables, the summaries and the truncation this is to take the place of.
func (db *ControlDB) sealSite(ctx context.Context, site *SiteDB, since int64, now time.Time) error {
	q, today := site.Queries(), now.UTC().Truncate(24*time.Hour)
	hot := today.AddDate(0, 0, -db.ServerConfig.RetentionHotDays)
	if sealed, err := q.HasColdDays(ctx); err != nil {
		return fmt.Errorf("cannot seal site %s: %w", site.Subdomain, err)
	} else if !sealed {
		since = 0
	}
	days, err := q.CGMDaysWrittenSince(ctx, sitedb.CGMDaysWrittenSinceParams{Since: since, Before: today.UnixMilli()})
	if err != nil {
		return fmt.Errorf("cannot seal site %s: %w", site.Subdomain, err)
	}
	for _, day := range days {
		// A day whose rows may be deleted keeps what it has, and takes what was written since
		if err := sealGlucoseDay(ctx, q, day*dayMs, db.ServerConfig.RetentionDeleteRows && day*dayMs < hot.UnixMilli()); err != nil {
			return fmt.Errorf("cannot seal %s of site %s: %w", time.UnixMilli(day*dayMs).UTC().Format(time.DateOnly), site.Subdomain, err)
		}
	}
	if db.ServerConfig.RetentionDeleteRows {
		if _, err := q.DeleteSealedCGMReadings(ctx, sitedb.DeleteSealedCGMReadingsParams{Before: hot.UnixMilli()}); err != nil {
			return fmt.Errorf("cannot delete the sealed readings of site %s: %w", site.Subdomain, err)
		}
	}
	return nil
}

// sealGlucoseDay writes the glucose series of the UTC day that starts at start, in ms, from its CGM readings. With
// merge, the day's rows may be gone, so it starts from the day as it was sealed before: a reading written since takes
// its step, and a deleted one clears its step. It reads the day back before it is kept, so a day is never sealed as
// what cannot be read, or with a different number of values.
func sealGlucoseDay(ctx context.Context, q *sitedb.Queries, start int64, merge bool) error {
	rows, err := q.CGMReadingsOfDay(ctx, sitedb.CGMReadingsOfDayParams{From: start, To: start + dayMs})
	if err != nil {
		return err
	}
	series, date := retention.Glucose, time.UnixMilli(start).UTC().Format(time.DateOnly)
	values, step := series.Empty(), func(row *sitedb.GlucoseReading) int64 { return (row.OccurredAt - start) / series.Step.Milliseconds() }
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
	live := slices.DeleteFunc(slices.Clone(rows), func(row *sitedb.GlucoseReading) bool { return row.DeletedAt != nil })
	for _, row := range rows {
		if row.DeletedAt != nil {
			values[step(row)] = math.NaN()
		}
	}
	// Newest first, so the oldest is written first and the newest reading of a step is the one kept
	cleaned := glucose.CleanReadings(live)
	for _, reading := range slices.Backward(cleaned) {
		values[step(reading.Row)] = math.Round(reading.Row.Mgdl)
	}
	packed := retention.PackGrid(series, values)
	read, err := retention.UnpackGrid(packed)
	if err != nil || !slices.EqualFunc(values, read, func(a, b float64) bool { return a == b || (math.IsNaN(a) && math.IsNaN(b)) }) {
		return fmt.Errorf("the sealed day does not read back as it was written: %v", err)
	}
	return q.UpsertColdDay(ctx, sitedb.UpsertColdDayParams{Kind: series.Name, Date: date, Body: packed})
}
