package db

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// A DB that is not backed up and does not open is removed, with its write-ahead log, and opened again. One that opens
// is left as it is, and one that still does not open is an error.
func TestOpenDisposable(t *testing.T) {
	tests := []struct {
		name            string
		failures, opens int
		removed, ok     bool
	}{
		{"opens", 0, 1, false, true},
		{"does not open once", 1, 2, true, true},
		{"never opens", 2, 2, true, false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "jobsdb.sqlite")
			for _, file := range dbFiles(path) {
				require.NoError(t, os.WriteFile(file, []byte("not a DB"), 0o600))
			}
			opens := 0
			opened, err := openDisposable(t.Context(), path, func() (*string, error) {
				opens++
				if opens <= test.failures {
					return nil, errors.New("file is not a database")
				}
				return &path, nil
			})
			require.Equal(t, test.opens, opens)
			require.Equal(t, test.ok, err == nil)
			require.Equal(t, test.ok, opened != nil)
			for _, file := range dbFiles(path) {
				_, statErr := os.Stat(file)
				require.Equal(t, test.removed, os.IsNotExist(statErr), file)
			}
		})
	}
}
