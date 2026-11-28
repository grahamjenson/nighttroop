package sitedb

import (
	"context"
	"reflect"
	"time"
)

// MarkStale marks the cold days of the day the record row happened in to be packaged again, as it was deleted from the
// day for good or moved away, which leaves no row that says so, when it is one with a time and not nil
func (q *Queries) MarkStale(ctx context.Context, row any) error {
	timed, ok := row.(interface{ GetOccurredAt() int64 })
	if value := reflect.ValueOf(row); !ok || value.Kind() == reflect.Pointer && value.IsNil() {
		return nil
	}
	return q.MarkColdDaysStale(ctx, MarkColdDaysStaleParams{Date: time.UnixMilli(timed.GetOccurredAt()).UTC().Format(time.DateOnly)})
}
