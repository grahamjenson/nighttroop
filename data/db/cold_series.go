package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"time"

	"nighttroop.t1d.nz/data/retention"
	"nighttroop.t1d.nz/data/sql/sitedb"
)

// ColdSeries are the values of a series of the site's packaged days from from up to, not including, to, in ms: one for
// each of the series' steps, the first at from, NaN where a day is not packaged or has no value. A local day is read
// this way from the UTC days it is in. from is on a step.
func ColdSeries(ctx context.Context, q *sitedb.Queries, series retention.Series, from, to int64) ([]float64, error) {
	step := series.Step.Milliseconds()
	values := make([]float64, max(0, (to-from+step-1)/step))
	for i := range values {
		values[i] = math.NaN()
	}
	for start := from - from%dayMs; start < to; start += dayMs {
		day, err := q.GetColdDay(ctx, sitedb.GetColdDayParams{Kind: series.Name, Date: time.UnixMilli(start).UTC().Format(time.DateOnly)})
		if errors.Is(err, sql.ErrNoRows) {
			continue
		}
		if err != nil {
			return nil, err
		}
		of, err := retention.UnpackGrid(day.Body)
		if err != nil {
			return nil, fmt.Errorf("cannot read the cold day %s of %s: %w", day.Date, series.Name, err)
		}
		for i, value := range of {
			if at := start + int64(i)*step; at >= from && at < to {
				values[(at-from)/step] = value
			}
		}
	}
	return values, nil
}
