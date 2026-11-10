package apiutils_test

import (
	"context"
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
		// A request its client stopped waiting for is not the server's error, but one that ran out of time is
		{"cancelled", context.Canceled, apiutils.ErrClientClosedRequest},
		{"cancelled wrapped", fmt.Errorf("searching: %w", context.Canceled), apiutils.ErrClientClosedRequest},
		{"timed out", context.DeadlineExceeded, apiutils.ErrInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if httpErr := apiutils.ErrorFor(tt.err, "numbers"); httpErr != tt.httpErr {
				t.Errorf("ErrorFor(%v) = %v, want %v", tt.err, httpErr, tt.httpErr)
			}
		})
	}
}
