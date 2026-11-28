package db

import (
	"context"
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
	if from >= to {
		return values, nil
	}
	date := func(at int64) string { return time.UnixMilli(at).UTC().Format(time.DateOnly) }
	days, err := q.ColdDaysNewestFirst(ctx, sitedb.ColdDaysNewestFirstParams{Kind: series.Name, FromDate: date(from), ToDate: date(to - 1), Limit: (to-from)/dayMs + 2})
	if err != nil {
		return nil, err
	}
	for _, day := range days {
		at, err := time.Parse(time.DateOnly, day.Date)
		of, unpackErr := retention.UnpackGrid(day.Body)
		if err = errors.Join(err, unpackErr); err != nil {
			return nil, fmt.Errorf("cannot read the cold day %s of %s: %w", day.Date, series.Name, err)
		}
		start := at.UnixMilli()
		for i, value := range of {
			if at := start + int64(i)*step; at >= from && at < to {
				values[(at-from)/step] = value
			}
		}
	}
	return values, nil
}
