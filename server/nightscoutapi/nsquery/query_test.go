package nsquery

import (
	"fmt"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Kept first so it is the first parse in this package, like the first requests after a restart
func Test_ConcurrentFirstParse(t *testing.T) {
	start := make(chan struct{})
	var wg sync.WaitGroup
	for range 100 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			_, err := ParseQueryValues(url.Values{"count": []string{"10"}})
			assert.NoError(t, err)
		}()
	}
	close(start)
	wg.Wait()
}

// We are only going to support a limited set of query params, so each query is checked for all of them
func TestParseQueryValues(t *testing.T) {
	date := "2023-01-01T00:10:00Z"
	tm, err := time.Parse(time.RFC3339, date)
	require.NoError(t, err)
	fields := []string{"date", "value", "type"}
	asc := SortOccurredAtAsc.String()
	tests := []struct {
		name    string
		query   url.Values
		dateGTE string   // "" when not set
		fields  []string // nil when not set
		sort    string   // "" when there is no sort order
		filters bool
	}{
		{"date gte", url.Values{"find[date][$gte]": {date}}, date, nil, "", true},
		{"date gte short", url.Values{"date$gte": {date}}, date, nil, "", true},
		{"date gte in milliseconds", url.Values{"find[date][$gte]": {fmt.Sprintf("%d", tm.UnixMilli())}}, date, nil, "", true},
		{"date gte in seconds", url.Values{"find[date][$gte]": {fmt.Sprintf("%d", tm.Unix())}}, date, nil, "", true},
		{"date gte in another format", url.Values{"find[date][$gte]": {tm.Format(time.DateTime)}}, date, nil, "", true},
		{"date lte", url.Values{"find[date][$lte]": {date}}, "", nil, "", true},
		{"date lte short", url.Values{"date$lte": {date}}, "", nil, "", true},
		{"fields listed", url.Values{"fields": {"date", "value", "type"}}, "", fields, "", false},
		{"fields with commas", url.Values{"fields": {"date,value,type"}}, "", fields, "", false},
		{"fields listed and with commas", url.Values{"fields": {"date,value", "type"}}, "", fields, "", false},
		{"fields with pipes", url.Values{"fields": {"date|value|type"}}, "", fields, "", false},
		{"sort by date", url.Values{"sort": {"date"}}, "", nil, asc, false},
		{"sort by mills", url.Values{"sort": {"mills"}}, "", nil, asc, false},
		{"sort by dateString", url.Values{"sort": {"dateString"}}, "", nil, asc, false},
		{"sort by an invalid field", url.Values{"sort": {"invalid_field"}}, "", nil, "", false},
		{"limit", url.Values{"limit": {"1"}}, "", nil, "", false},
		{"skip", url.Values{"skip": {"1"}}, "", nil, "", false},
		{"nothing", url.Values{}, "", nil, "", false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			query, err := ParseQueryValues(test.query)
			require.NoError(t, err)

			dateGTE := ""
			if query.DateGTE != nil {
				dateGTE = query.DateGTE.TimeString()
			}
			require.Equal(t, test.dateGTE, dateGTE)

			var fields []string
			if query.Fields != nil {
				fields = query.Fields.ToArray()
			}
			require.Equal(t, test.fields, fields)

			sort := ""
			if order := query.SortOrder(); order != nil {
				sort = order.String()
			}
			require.Equal(t, test.sort, sort)

			require.Equal(t, test.filters, query.HasAnyFilters())
		})
	}
}

// Trio and iAPS filter treatments by who entered them, each $ne alone or in a $and adding to a $nin, and on whether
// they have carbs or a duration, a number at least 0
func TestParseQueryValuesTrioFilters(t *testing.T) {
	query, err := ParseQueryValues(url.Values{
		"find[carbs][$exists]": {"true"}, "find[duration][$exists]": {"true"},
		"find[$and][0][enteredBy][$ne]": {"Trio"}, "find[$and][1][enteredBy][$ne]": {"iAPS"}, "find[enteredBy][$ne]": {"loop://iPhone"},
	})
	require.NoError(t, err)
	require.ElementsMatch(t, []string{"Trio", "iAPS", "loop://iPhone"}, query.EnteredByNin.ToArray())
	require.Equal(t, int64(0), int64(*query.CarbsGTE))
	require.Equal(t, int64(0), int64(*query.DurationGTE))
	require.Empty(t, query.Other)
	require.True(t, query.HasAnyFilters())

	query, err = ParseQueryValues(url.Values{"find[enteredBy]": {"remote"}, "find[carbs][$exists]": {"true"}, "find[carbs][$gte]": {"10"}})
	require.NoError(t, err)
	require.Equal(t, "remote", query.EnteredByEq.String())
	require.Equal(t, int64(10), int64(*query.CarbsGTE), "a $gte sent is kept over the $exists")

	// Only $exists=true is a filter it knows
	query, err = ParseQueryValues(url.Values{"find[carbs][$exists]": {"false"}})
	require.NoError(t, err)
	require.Nil(t, query.CarbsGTE)
	require.Contains(t, query.Other, "find[carbs][$exists]")
}
