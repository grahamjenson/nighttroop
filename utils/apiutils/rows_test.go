package apiutils_test

import (
	"database/sql"
	"errors"
	"strconv"
	"testing"

	"nighttroop.t1d.nz/utils/apiutils"
)

func TestFound(t *testing.T) {
	toModel := func(row *int) *string { s := strconv.Itoa(*row); return &s }
	one := 1
	tests := []struct {
		name    string
		row     *int
		err     error
		model   *string
		httpErr *apiutils.HTTPError
	}{
		{"found", &one, nil, toModel(&one), nil},
		{"no rows", nil, sql.ErrNoRows, nil, nil},
		{"nil row", nil, nil, nil, nil},
		{"error", nil, errors.New("disk"), nil, apiutils.ErrInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model, httpErr := apiutils.Found(tt.row, tt.err, "number", toModel)
			if httpErr != tt.httpErr || (model == nil) != (tt.model == nil) || (model != nil && *model != *tt.model) {
				t.Fatalf("got %v %v, want %v %v", model, httpErr, tt.model, tt.httpErr)
			}
		})
	}
}

func TestSearched(t *testing.T) {
	toModel := apiutils.WithContext(func(row *int, add int) *int { n := *row + add; return &n }, 10)
	one, two := 1, 2
	models, httpErr := apiutils.Searched([]*int{&one, &two}, nil, "numbers", toModel)
	if httpErr != nil || len(models) != 2 || *models[0] != 11 || *models[1] != 12 {
		t.Fatalf("unexpected models %v %v", models, httpErr)
	}
	// No rows is an empty list, not null, in the JSON API
	if models, _ := apiutils.Searched([]*int(nil), nil, "numbers", toModel); models == nil {
		t.Fatal("no rows should be an empty list")
	}
	if _, httpErr := apiutils.Searched([]*int(nil), errors.New("disk"), "numbers", toModel); httpErr != apiutils.ErrInternalServerError {
		t.Fatalf("unexpected error %v", httpErr)
	}
}
