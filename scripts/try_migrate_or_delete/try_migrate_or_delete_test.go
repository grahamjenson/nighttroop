package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// newDB is a DB of a kind at path in the data folder's DB folder, opened as the server opens it
func newDB(t *testing.T, dataDir, path, name string) {
	t.Helper()
	require.NoError(t, openers[name](context.Background(), filepath.Join(dataDir, "db", path)))
}

func TestTryMigrateOrDelete(t *testing.T) {
	ctx := context.Background()
	dataDir := t.TempDir()
	require.NoError(t, tryMigrateOrDelete(ctx, dataDir), "a server with no DBs yet is left as it is")

	// The DBs a server makes, and its other files, are kept
	for _, name := range []string{"controldb", "jobsdb", "logsdb"} {
		newDB(t, dataDir, name+".sqlite", name)
	}
	newDB(t, dataDir, "site/a/abcdefgh.sqlite", "sitedb")
	newDB(t, dataDir, "site/c/cdefghij.sqlite", "sitedb")
	backup := filepath.Join(dataDir, "backup", "controldb.sqlite")
	require.NoError(t, os.MkdirAll(filepath.Dir(backup), 0o700))
	require.NoError(t, os.WriteFile(backup, []byte("a backup"), 0o600))
	clear(preparedSiteSchemas)
	require.NoError(t, tryMigrateOrDelete(ctx, dataDir))
	require.FileExists(t, filepath.Join(dataDir, "db", "controldb.sqlite"))
	require.FileExists(t, filepath.Join(dataDir, "db", "site/a/abcdefgh.sqlite"))
	require.FileExists(t, backup)
	require.Len(t, preparedSiteSchemas, 1, "site DBs with the same schema have their queries prepared once")

	// A site DB whose tables its queries do not fit, as one made before a migration was edited, has another schema, so
	// its queries are prepared, and deletes every DB and backup, not only it
	newDB(t, dataDir, "site/b/bcdefghi.sqlite", "jobsdb")
	require.NoError(t, tryMigrateOrDelete(ctx, dataDir))
	require.NoDirExists(t, filepath.Join(dataDir, "db"))
	require.NoDirExists(t, filepath.Join(dataDir, "backup"))
}
