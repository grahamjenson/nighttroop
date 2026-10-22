package siteapiutils

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/utils/apiutils"
)

func TestWriteError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want *apiutils.HTTPError
	}{
		{"no error", nil, nil},
		{"the store's own answer", apiutils.ErrMissingRequiredFieldWithHint("carbs"), apiutils.ErrMissingRequiredFieldWithHint("carbs")},
		{"no row", sql.ErrNoRows, apiutils.ErrNotFound},
		{"no row, wrapped", errors.Join(errors.New("updating"), sql.ErrNoRows), apiutils.ErrNotFound},
		{"anything else", errors.New("disk I/O error"), apiutils.ErrInternalServerError},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.want, writeError(test.err, "test rows"))
		})
	}
}
