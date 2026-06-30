package nsutils

import (
	"fmt"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// We are only going to support a limited set of query params

func Test_Query_DateGTE(t *testing.T) {
	date := "2023-01-01T00:10:00Z"
	tm, _ := time.Parse(time.RFC3339, date) // Ensure the date is valid
	dateMills := fmt.Sprintf("%d", tm.UnixMilli())
	dateSeconds := fmt.Sprintf("%d", tm.Unix())
	dateDifferentFormat := tm.Format(time.DateTime)
	queries := []url.Values{
		{
			"find[date][$gte]": []string{date},
		},
		{
			"date$gte": []string{date},
		},
		{
			"find[date][$gte]": []string{dateMills},
		},
		{
			"find[date][$gte]": []string{dateSeconds},
		},
		{
			"find[date][$gte]": []string{dateDifferentFormat},
		},
	}

	for _, q := range queries {
		query, err := parseQuery(q)
		if !assert.NoError(t, err) {
			return
		}
		if !assert.NotNil(t, query.DateGTE) {
			return
		}
		assert.Equal(t, date, query.DateGTE.TimeString())
	}
}

func Test_Fields(t *testing.T) {
	queries := []url.Values{
		{
			"fields": []string{"date", "value", "type"},
		},
		{
			"fields": []string{"date,value,type"},
		},
		{
			"fields": []string{"date,value", "type"},
		},
		{
			"fields": []string{"date|value|type"},
		},
	}
	for _, q := range queries {
		query, err := parseQuery(q)
		if !assert.NoError(t, err) {
			return
		}
		if !assert.NotNil(t, query.Fields) {
			return
		}
		assert.Equal(t, []string{"date", "value", "type"}, query.Fields.ToArray())
	}
}
