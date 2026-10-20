package apiutils

import (
	"net/url"

	"nighttroop.t1d.nz/utils/apiutils/apiquery"
)

type Query = apiquery.Query

func QueryFromParams(values url.Values) (*Query, *HTTPError) {
	queryValues, err := apiquery.ParseQueryValues(values)
	if err != nil {
		return nil, ErrBadRequest
	}
	query := apiquery.QueryValuesToQuery(queryValues)
	return &query, nil
}
