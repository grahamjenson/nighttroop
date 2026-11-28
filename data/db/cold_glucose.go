package db

import (
	"context"
	"fmt"
	"math"
	"time"

	"nighttroop.t1d.nz/data/retention"
	"nighttroop.t1d.nz/data/sql/sitedb"
)

// coldIDSuffix ends the ID of a reading made from a cold day. With the 8 hex digits of its time before it, the ID is
// shaped as a MongoDB ObjectId, as each reading's is, and is the same each time the reading is made.
const coldIDSuffix = "c01dc01dc01dc01d"

// ColdReadingID is the ID of the reading made from a cold day's step at at, in ms
func ColdReadingID(at int64) string {
	return fmt.Sprintf("%08x%s", at/1000, coldIDSuffix)
}

// ColdGlucoseReadings are the CGM readings of the site's sealed days from from up to, not including, to, in ms, the
// newest first, at most limit of them: one for each step of a day's glucose series that has a value and that keep
// takes, at the start of its step. They are made as rows are, so what reads rows reads them, but they are not in the
// DB: a day's rows are deleted once they are old (see sealSite), and its cold day is what is left of them.
func ColdGlucoseReadings(ctx context.Context, q *sitedb.Queries, from, to int64, limit int, keep func(at int64, mgdl float64) bool) ([]*sitedb.GlucoseReading, error) {
	date := func(at int64) string { return time.UnixMilli(at).UTC().Format(time.DateOnly) }
	var readings []*sitedb.GlucoseReading
	for last := to - 1; from < to && len(readings) < limit; {
		// A day has at most 288 readings every 5 minutes, so a few more days than the readings wanted need
		days, err := q.ColdDaysNewestFirst(ctx, sitedb.ColdDaysNewestFirstParams{Kind: retention.Glucose.Name, FromDate: date(from), ToDate: date(last), Limit: int64((limit-len(readings))/200 + 2)})
		if err != nil || len(days) == 0 {
			return readings, err
		}
		for _, day := range days {
			start, err := time.Parse(time.DateOnly, day.Date)
			if err != nil {
				return nil, err
			}
			values, err := retention.UnpackGrid(day.Body)
			if err != nil || len(values) == 0 {
				return nil, fmt.Errorf("cannot read the cold day %s: %w", day.Date, err)
			}
			step := dayMs / int64(len(values))
			for i := len(values) - 1; i >= 0 && len(readings) < limit; i-- {
				at := start.UnixMilli() + int64(i)*step
				if math.IsNaN(values[i]) || at < from || at >= to || !keep(at, values[i]) {
					continue
				}
				readings = append(readings, &sitedb.GlucoseReading{ID: ColdReadingID(at), Mgdl: values[i], ReadingType: "sgv", OccurredAt: at,
					CreatedAt: at, UpdatedAt: at, UtcOffsetMinutes: new(int64(0)), CreatedByIdentityName: "nighttroop", CreatedByDeviceName: new("nighttroop")})
			}
			last = start.UnixMilli() - 1
		}
	}
	return readings, nil
}
