package apiutils

import (
	"database/sql"
	"errors"
)

// Found is the model for the row a Get query returned, nil when there is no such row, or the error as ErrorFor says
func Found[R, M any](row *R, err error, what string, toModel func(*R) *M) (*M, *HTTPError) {
	if errors.Is(err, sql.ErrNoRows) || (err == nil && row == nil) {
		return nil, nil
	}
	if herr := ErrorFor(err, what); herr != nil {
		return nil, herr
	}
	return toModel(row), nil
}

// Searched is the models for the rows a search query returned, or the error as ErrorFor says
func Searched[R, M any](rows []*R, err error, what string, toModel func(*R) *M) ([]*M, *HTTPError) {
	if herr := ErrorFor(err, what); herr != nil {
		return nil, herr
	}
	models := make([]*M, 0, len(rows))
	for _, row := range rows {
		models = append(models, toModel(row))
	}
	return models, nil
}

// WithContext is a converter that also takes a context, e.g. the site's units, as one that does not
func WithContext[R, M, C any](toModel func(*R, C) *M, context C) func(*R) *M {
	return func(row *R) *M { return toModel(row, context) }
}
