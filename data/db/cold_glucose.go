package db

import (
	"cmp"
	"context"
	"database/sql"
	"fmt"
	"math"
	"reflect"
	"slices"
	"strconv"
	"strings"
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

// eachColdDay reads the site's cold days of a kind from from up to, not including, to, in ms, the newest first, a
// few at a time, giving each to visit with the time its day starts, while more says more are wanted. wanted is how
// many records are still wanted, of which a day has about perDay, so no more days are read than are needed.
func eachColdDay(ctx context.Context, q *sitedb.Queries, kind string, from, to int64, perDay int, wanted func() int, visit func(start int64, body []byte) error) error {
	date := func(at int64) string { return time.UnixMilli(at).UTC().Format(time.DateOnly) }
	for last := to - 1; from < to && wanted() > 0; {
		days, err := q.ColdDaysNewestFirst(ctx, sitedb.ColdDaysNewestFirstParams{Kind: kind, FromDate: date(from), ToDate: date(last), Limit: int64(wanted()/perDay + 2)})
		if err != nil || len(days) == 0 {
			return err
		}
		for _, day := range days {
			start, err := time.Parse(time.DateOnly, day.Date)
			if err != nil {
				return err
			}
			if err := visit(start.UnixMilli(), day.Body); err != nil {
				return fmt.Errorf("cannot read the cold day %s of %s: %w", day.Date, kind, err)
			}
			last = start.UnixMilli() - 1
		}
	}
	return nil
}

// ColdGlucoseReadings are the CGM readings of the site's packaged days from from up to, not including, to, in ms, the
// newest first, at most limit of them: one for each step of a day's glucose series that has a value and that keep
// takes, at the start of its step. They are made as rows are, so what reads rows reads them, but they are not in the
// DB: a day's rows are deleted once they are old (see packageSite), and its cold day is what is left of them.
func ColdGlucoseReadings(ctx context.Context, q *sitedb.Queries, from, to int64, limit int, keep func(at int64, mgdl float64) bool) ([]*sitedb.GlucoseReading, error) {
	var readings []*sitedb.GlucoseReading
	// A day has at most 288 readings every 5 minutes, and most have fewer
	err := eachColdDay(ctx, q, retention.Glucose.Name, from, to, 200, func() int { return limit - len(readings) }, func(start int64, body []byte) error {
		values, err := retention.UnpackGrid(body)
		if err != nil || len(values) == 0 {
			return fmt.Errorf("not a series: %w", err)
		}
		step := dayMs / int64(len(values))
		for i := len(values) - 1; i >= 0 && len(readings) < limit; i-- {
			at := start + int64(i)*step
			if math.IsNaN(values[i]) || at < from || at >= to || !keep(at, values[i]) {
				continue
			}
			readings = append(readings, &sitedb.GlucoseReading{ID: ColdReadingID(at), Mgdl: values[i], ReadingType: "sgv", OccurredAt: at,
				CreatedAt: at, UpdatedAt: at, UtcOffsetMinutes: new(int64(0)), CreatedByIdentityName: "nighttroop", CreatedByDeviceName: new("nighttroop")})
		}
		return nil
	})
	return readings, err
}

// ColdEvents are the events of the site's packaged days from from up to, not including, to, in ms, the newest first, at
// most limit of them, that keep takes: each as it is kept, the columns of its row by their names, with "k", its table
// (see packageEventsDay). What was in force when a day started is not one of its events.
func ColdEvents(ctx context.Context, q *sitedb.Queries, from, to int64, limit int, keep func(event map[string]any) bool) ([]map[string]any, error) {
	var events []map[string]any
	err := eachColdDay(ctx, q, retention.EventsKind, from, to, 10, func() int { return limit - len(events) }, func(_ int64, body []byte) error {
		_, day, err := retention.ReadEvents(body)
		if err != nil {
			return err
		}
		for _, event := range slices.Backward(day.Events) {
			if at := EventAt(event); len(events) < limit && at >= from && at < to && keep(event) {
				events = append(events, event)
			}
		}
		return nil
	})
	return events, err
}

// EventAt is when a cold event occurred, in ms
func EventAt(event map[string]any) int64 {
	at, _ := event["occurred_at"].(float64)
	return int64(at)
}

// RowOf sets into, a row of table, from a cold event: each column the event kept, by its name, as eventOf read it. The
// times it was written are when it occurred, as a cold event does not keep them.
func RowOf(table string, event map[string]any, into any) {
	row := reflect.ValueOf(into).Elem()
	for i := range row.NumField() {
		field, name := row.Field(i), row.Type().Field(i).Name
		limit, ok := sitedb.ColumnLimitOf(table, name)
		kept, has := event[limit.Column]
		switch {
		case name == "CreatedAt" || name == "UpdatedAt":
			field.SetInt(EventAt(event))
		case !ok || !has || kept == nil:
		default:
			if field.Kind() == reflect.Pointer {
				field.Set(reflect.New(field.Type().Elem()))
				field = field.Elem()
			}
			// A number is kept as JSON's, so it is turned to the column's, and nothing is set from another kind
			if value := reflect.ValueOf(kept); value.Kind() == field.Kind() || (value.CanFloat() && (field.CanInt() || field.CanFloat())) {
				field.Set(value.Convert(field.Type()))
			}
		}
	}
}

// ColdReadingAt is the time, in ms, of the reading a cold day's ID names, and whether id is one (see ColdReadingID)
func ColdReadingAt(id string) (int64, bool) {
	seconds, err := strconv.ParseInt(strings.TrimSuffix(id, coldIDSuffix), 16, 64)
	return seconds * 1000, err == nil && len(id) == 8+len(coldIDSuffix) && strings.HasSuffix(id, coldIDSuffix)
}

// ColdGlucoseReading is the reading id names, made from its step of its cold day, or sql.ErrNoRows when id is not a
// cold day's or the step has no value
func ColdGlucoseReading(ctx context.Context, q *sitedb.Queries, id string) (*sitedb.GlucoseReading, error) {
	at, ok := ColdReadingAt(id)
	if !ok {
		return nil, sql.ErrNoRows
	}
	readings, err := ColdGlucoseReadings(ctx, q, at, at+1, 1, func(int64, float64) bool { return true })
	if err != nil || len(readings) == 0 {
		return nil, cmp.Or(err, sql.ErrNoRows)
	}
	return readings[0], nil
}

// ClearColdGlucoseReading deletes the reading id names from its cold day, by clearing its step, and is the reading it
// was, or sql.ErrNoRows when there was none. A day packaged again keeps the step clear, as it has no row.
func ClearColdGlucoseReading(ctx context.Context, q *sitedb.Queries, id string) (*sitedb.GlucoseReading, error) {
	reading, err := ColdGlucoseReading(ctx, q, id)
	if err != nil {
		return nil, err
	}
	start := reading.OccurredAt - reading.OccurredAt%dayMs
	values, err := ColdSeries(ctx, q, retention.Glucose, start, start+dayMs)
	if err != nil {
		return nil, err
	}
	values[(reading.OccurredAt-start)/retention.Glucose.Step.Milliseconds()] = math.NaN()
	return reading, writeSeriesDay(ctx, q, retention.Glucose, time.UnixMilli(start).UTC().Format(time.DateOnly), values)
}
