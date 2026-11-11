package apiutils

import (
	"database/sql"
	"log/slog"
)

// Found is the model for the row a Get query returned, or nil when there is no such row
func Found[R, M any](row *R, err error, what string, toModel func(*R) *M) (*M, *HTTPError) {
	if err == sql.ErrNoRows || (err == nil && row == nil) {
		return nil, nil
	}
	if err != nil {
		slog.Error("Error fetching "+what, "error", err)
		return nil, ErrInternalServerError
	}
	return toModel(row), nil
}

// Searched is the models for the rows a search query returned
func Searched[R, M any](rows []*R, err error, what string, toModel func(*R) *M) ([]*M, *HTTPError) {
	if err != nil {
		slog.Error("Error searching "+what, "error", err)
		return nil, ErrInternalServerError
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
