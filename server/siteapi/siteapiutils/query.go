package siteapiutils

import (
	"net/url"

	"nighttroop.t1d.nz/server/siteapi/sitequery"
)

type Query = sitequery.Query

func QueryFromParams(values url.Values) (*Query, *HTTPError) {
	queryValues, err := sitequery.ParseQueryValues(values)
	if err != nil {
		return nil, ErrBadRequest
	}
	query := sitequery.QueryValuesToQuery(queryValues)
	return &query, nil
}
