package sitedb

import (
	"context"
	"reflect"
)

// MarkStale marks the summaries the record row happened in to be built again, as it was deleted from them or moved
// away, when it is one with a time and not nil
func (q *Queries) MarkStale(ctx context.Context, row any) error {
	timed, ok := row.(interface{ GetOccurredAt() int64 })
	if value := reflect.ValueOf(row); !ok || value.Kind() == reflect.Pointer && value.IsNil() {
		return nil
	}
	return q.MarkSummariesStale(ctx, MarkSummariesStaleParams{At: timed.GetOccurredAt()})
}
