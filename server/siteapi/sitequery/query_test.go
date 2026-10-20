package sitequery_test

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/server/siteapi/sitequery"
)

func TestQueryLimitAndSkip(t *testing.T) {
	tests := []struct {
		query string
		limit int64
		skip  int64
		err   bool
	}{
		{query: "", limit: sitequery.MaxQueryLimit, skip: 0},
		{query: "limit=20&skip=40", limit: 20, skip: 40},
		{query: "limit=1000", limit: 1000, skip: 0},
		{query: "limit=1001", limit: sitequery.MaxQueryLimit, skip: 0},
		{query: "limit=0", limit: sitequery.MaxQueryLimit, skip: 0},
		{query: "limit=-5&skip=-5", limit: sitequery.MaxQueryLimit, skip: 0},
		{query: "limit=20.0&skip=10", limit: 20, skip: 10},
		{query: "other=1", limit: sitequery.MaxQueryLimit, skip: 0},
		{query: "limit=lots", err: true},
		{query: "skip=lots", err: true},
	}
	for _, test := range tests {
		t.Run(test.query, func(t *testing.T) {
			values, err := url.ParseQuery(test.query)
			require.NoError(t, err)
			queryValues, err := sitequery.ParseQueryValues(values)
			if test.err {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			query := sitequery.QueryValuesToQuery(queryValues)
			require.Equal(t, test.limit, query.Limit)
			require.Equal(t, test.skip, query.Skip)
		})
	}
}
