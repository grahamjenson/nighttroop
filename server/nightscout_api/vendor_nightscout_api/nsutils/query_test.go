package nsutils

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

// We are only going to support a limited set of query params

func Test_Query_DateGTE(t *testing.T) {
	query, err := parseQuery(url.Values{
		"find[date][$gte]": []string{"2023-01-01T00:00:00Z"},
	})

	assert.NoError(t, err)
	if !assert.NotNil(t, query.DateGTE) {
		return
	}
	assert.Equal(t, "2023-01-01T00:00:00Z", query.DateGTE.TimeString())

	query, err = parseQuery(url.Values{
		"date$gte": []string{"2023-01-01T00:00:00Z"},
	})
	assert.NoError(t, err)
	if !assert.NotNil(t, query.DateGTE) {
		return
	}
	assert.Equal(t, "2023-01-01T00:00:00Z", query.DateGTE.TimeString())
	assert.Empty(t, query.Other)
}
