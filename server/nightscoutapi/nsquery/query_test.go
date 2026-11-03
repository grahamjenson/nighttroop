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

// Each filter a client sends that Nighttroop does not know gives nothing; any other parameter is not a filter
func TestHasUnknownFilter(t *testing.T) {
	tests := []struct {
		key     string
		unknown bool
	}{
		{"find[date][$gte]", false},
		{"date$gte", false},
		{"sort$desc", false},
		{"find[uuid]", false},
		{"rr", false},
		{"secret", false},
		{"now", false},
		{"find[duration][$exists]", false},
		{"find[enteredBy]", false},
		{"find[enteredBy][$ne]", false},
		{"find[$and][0][enteredBy][$ne]", false},
		{"find[$and][12][enteredBy][$ne]", false},
		{"find[uuid_v2]", true},
		{"find[notes]", true},
		{"find[carbs][$exists]", false},
		{"find[sgv][$exists]", true},
		{"find[$and][0][notes][$ne]", true},
		{"find[$or][0][id][$eq]", true},
		{"find[enteredBy][$eq]", false},
		{"notes$eq", true},
		{"enteredBy$ne", true},
	}
	for _, test := range tests {
		t.Run(test.key, func(t *testing.T) {
			query, err := ParseQueryValues(url.Values{test.key: {"1"}})
			require.NoError(t, err)
			require.Equal(t, test.unknown, query.HasUnknownFilter())
		})
	}
}

// Trio sends each enteredBy it leaves out as find[$and][n][enteredBy][$ne] and iAPS repeats find[enteredBy][$ne],
// encoding them twice; each is one of the $nin
func TestParseEnteredByNotIn(t *testing.T) {
	tests := []struct {
		name  string
		query url.Values
		want  []string
	}{
		{"trio $and", url.Values{"find[$and][0][enteredBy][$ne]": {"Trio"}, "find[$and][1][enteredBy][$ne]": {"iAPS"}}, []string{"Trio", "iAPS"}},
		{"iaps repeated", url.Values{"find[enteredBy][$ne]": {"iAPS", "iAPS%20Watch"}}, []string{"iAPS", "iAPS Watch"}},
		{"with a $nin", url.Values{"find[enteredBy][$ne]": {"Trio"}, "find[enteredBy][$nin]": {"Loop"}}, []string{"Trio", "Loop"}},
		{"none", url.Values{"find[enteredBy]": {"remote"}}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qv, err := ParseQueryValues(tt.query)
			require.NoError(t, err)
			if tt.want == nil {
				assert.Nil(t, qv.EnteredByNin)
				assert.Equal(t, "remote", qv.EnteredByEq.String())
				return
			}
			require.NotNil(t, qv.EnteredByNin)
			assert.ElementsMatch(t, tt.want, qv.EnteredByNin.ToArray())
			assert.True(t, qv.HasAnyFilters())
		})
	}
}

// Each find[$or] alternative is a query with the others' shared values, and one without is itself
func TestOrBranches(t *testing.T) {
	values := url.Values{
		"find[$or][0][id][$eq]":         {"a"},
		"find[$or][1][dateString][$eq]": {"2026-11-02T12:00:00.000Z"},
		"count":                         {"1"},
	}
	branches := OrBranches(values)
	assert.ElementsMatch(t, []url.Values{
		{"find[id][$eq]": {"a"}, "count": {"1"}},
		{"find[dateString][$eq]": {"2026-11-02T12:00:00.000Z"}, "count": {"1"}},
	}, branches)

	plain := url.Values{"find[id][$eq]": {"a"}}
	assert.Equal(t, []url.Values{plain}, OrBranches(plain))
}
