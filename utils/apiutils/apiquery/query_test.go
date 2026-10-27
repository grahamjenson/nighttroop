package apiquery_test

import (
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/utils/apiutils/apiquery"
)

func TestQueryLimitAndSkip(t *testing.T) {
	tests := []struct {
		query string
		limit int64
		skip  int64
		err   bool
	}{
		{query: "", limit: apiquery.MaxQueryLimit, skip: 0},
		{query: "limit=20&skip=40", limit: 20, skip: 40},
		{query: "limit=1000", limit: 1000, skip: 0},
		{query: "limit=1001", limit: apiquery.MaxQueryLimit, skip: 0},
		{query: "limit=0", limit: apiquery.MaxQueryLimit, skip: 0},
		{query: "limit=-5&skip=-5", limit: apiquery.MaxQueryLimit, skip: 0},
		{query: "limit=20.0&skip=10", limit: 20, skip: 10},
		{query: "other=1", limit: apiquery.MaxQueryLimit, skip: 0},
		{query: "limit=lots", err: true},
		{query: "skip=lots", err: true},
	}
	for _, test := range tests {
		t.Run(test.query, func(t *testing.T) {
			values, err := url.ParseQuery(test.query)
			require.NoError(t, err)
			queryValues, err := apiquery.ParseQueryValues(values)
			if test.err {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			query := apiquery.QueryValuesToQuery(queryValues)
			require.Equal(t, test.limit, query.Limit)
			require.Equal(t, test.skip, query.Skip)
		})
	}
}

func TestQueryHours(t *testing.T) {
	week := int64(apiquery.MaxQueryHours)
	tests := []struct {
		query string
		hours *int64
	}{
		{query: "", hours: nil},
		{query: "hours=6", hours: func() *int64 { h := int64(6); return &h }()},
		{query: "hours=1000", hours: &week},
		{query: "hours=0", hours: nil},
		{query: "hours=-1", hours: nil},
	}
	for _, test := range tests {
		t.Run(test.query, func(t *testing.T) {
			values, err := url.ParseQuery(test.query)
			require.NoError(t, err)
			queryValues, err := apiquery.ParseQueryValues(values)
			require.NoError(t, err)
			require.Equal(t, test.hours, apiquery.QueryValuesToQuery(queryValues).Hours)
		})
	}
}

func TestGetSearch(t *testing.T) {
	str := func(s string) *string { return &s }
	long := strings.Repeat("a", apiquery.MaxSearchLength+10)
	tests := []struct {
		name string
		in   *string
		want *string
	}{
		{"unset", nil, nil},
		{"blank is every row", str("   "), nil},
		{"trimmed", str("  synced "), str("synced")},
		{"cut to the longest", &long, str(long[:apiquery.MaxSearchLength])},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, apiquery.GetSearch(apiquery.QueryValues{Search: tt.in}))
		})
	}
}
