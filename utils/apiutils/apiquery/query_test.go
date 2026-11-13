package apiquery_test

import (
	"net/url"
	"strings"
	"testing"
	"time"

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
		{query: "", limit: apiquery.DefaultQueryLimit, skip: 0},
		{query: "limit=20&skip=40", limit: 20, skip: 40},
		{query: "limit=1000", limit: 1000, skip: 0},
		{query: "limit=1001", limit: apiquery.MaxQueryLimit, skip: 0},
		{query: "limit=0", limit: apiquery.DefaultQueryLimit, skip: 0},
		{query: "limit=-5&skip=-5", limit: apiquery.DefaultQueryLimit, skip: 0},
		{query: "skip=32000", limit: apiquery.DefaultQueryLimit, skip: 32000},
		{query: "skip=32001", limit: apiquery.DefaultQueryLimit, skip: apiquery.MaxQuerySkip},
		{query: "skip=1000000000", limit: apiquery.DefaultQueryLimit, skip: apiquery.MaxQuerySkip},
		{query: "limit=20.0&skip=10", limit: 20, skip: 10},
		{query: "other=1", limit: apiquery.DefaultQueryLimit, skip: 0},
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

// A day is the rows on or before it, in the site's timezone: those before the next day starts there, from any cursor
func TestQueryBeforeDay(t *testing.T) {
	auckland, err := time.LoadLocation("Pacific/Auckland")
	require.NoError(t, err)
	tests := []struct {
		day      string
		location *time.Location
		before   *int64
	}{
		{"2026-11-01", time.UTC, ptr(time.Date(2026, 11, 2, 0, 0, 0, 0, time.UTC).UnixMilli())},
		{"2026-11-01", auckland, ptr(time.Date(2026, 11, 2, 0, 0, 0, 0, auckland).UnixMilli())},
		{"1 November", time.UTC, ptr(int64(5))},
		{"", time.UTC, ptr(int64(5))},
	}
	for _, test := range tests {
		t.Run(test.day+" "+test.location.String(), func(t *testing.T) {
			id := "row"
			query := apiquery.Query{Before: ptr(int64(5)), BeforeID: &id, BeforeDate: &test.day}
			query.BeforeDay(test.location)
			require.Equal(t, test.before, query.Before)
			if *test.before != 5 {
				require.Nil(t, query.BeforeID)
			}
		})
	}
}

func ptr[T any](value T) *T { return &value }
