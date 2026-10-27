package utils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDeleteFiles(t *testing.T) {
	dir := t.TempDir()
	present := filepath.Join(dir, "present.db")
	require.NoError(t, os.WriteFile(present, []byte("x"), 0600))
	require.True(t, FileExists(present))

	// A file that is not there is already deleted
	require.NoError(t, DeleteFiles([]string{present, filepath.Join(dir, "missing.db")}))
	require.False(t, FileExists(present))

	// A directory with a file in it cannot be removed, and says which path failed once
	full := filepath.Join(dir, "full")
	require.NoError(t, MkDirs(full))
	require.NoError(t, os.WriteFile(filepath.Join(full, "a"), nil, 0600))
	err := DeleteFile(full)
	require.ErrorContains(t, err, "cannot delete file "+full+":")
}
