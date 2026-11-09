package apiutils

import "context"

// Totals are how many rows a search matches over all its pages, and when the earliest and latest of them occurred
type Totals struct {
	Total      int64  `json:"total"`
	EarliestAt *int64 `json:"earliest_at"`
	LatestAt   *int64 `json:"latest_at"`
}

// SearchResult is one page of a search, newest first: its items, which page it is, and, for a collection that counts
// what it matches, its totals
type SearchResult[M any] struct {
	Items []*M  `json:"items"`
	Skip  int64 `json:"skip"`
	Limit int64 `json:"limit"`
	*Totals
}

type totalsKey struct{}

// WithTotals keeps a search's totals for the page showing it
func WithTotals(ctx context.Context, totals *Totals) context.Context {
	return context.WithValue(ctx, totalsKey{}, totals)
}

// TotalsFrom are the totals of the search a page shows, nil when its collection does not count them
func TotalsFrom(ctx context.Context) *Totals {
	totals, _ := ctx.Value(totalsKey{}).(*Totals)
	return totals
}
