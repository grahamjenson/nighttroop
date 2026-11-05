package apiutils_test

import (
	"database/sql"
	"errors"
	"fmt"
	"testing"

	"nighttroop.t1d.nz/utils/apiutils"
)

func TestErrorFor(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		httpErr *apiutils.HTTPError
	}{
		{"no error", nil, nil},
		{"its own", apiutils.ErrForbidden, apiutils.ErrForbidden},
		{"its own wrapped", fmt.Errorf("checking: %w", apiutils.ErrForbidden), apiutils.ErrForbidden},
		{"no rows", sql.ErrNoRows, apiutils.ErrNotFound},
		{"no rows wrapped", fmt.Errorf("getting: %w", sql.ErrNoRows), apiutils.ErrNotFound},
		{"anything else", errors.New("disk"), apiutils.ErrInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if httpErr := apiutils.ErrorFor(tt.err, "numbers"); httpErr != tt.httpErr {
				t.Errorf("ErrorFor(%v) = %v, want %v", tt.err, httpErr, tt.httpErr)
			}
		})
	}
}
