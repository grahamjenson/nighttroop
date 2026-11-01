package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// newDB is a DB of a kind at path in dir, made by its migrations as the server makes it
func newDB(t *testing.T, dir, path, name string) {
	t.Helper()
	k := kinds[name]
	db, err := k.open(filepath.Join(dir, path), 0)
	require.NoError(t, err)
	require.NoError(t, k.migrate(db))
	require.NoError(t, db.Close())
}

func TestCheck(t *testing.T) {
	ctx := context.Background()
	dir := filepath.Join(t.TempDir(), "db")
	require.NoError(t, check(ctx, dir), "a server with no DBs yet is valid")

	// The DBs a server makes, and files that are not its DBs, are valid
	for _, name := range []string{"controldb", "jobsdb", "logsdb"} {
		newDB(t, dir, name+".sqlite", name)
	}
	newDB(t, dir, "site/a/abcdefgh.sqlite", "sitedb")
	newDB(t, dir, "site/b/bcdefghi.sqlite", "sitedb")
	require.NoError(t, os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("not a DB"), 0o600))
	require.NoError(t, check(ctx, dir))

	// A site DB made by other migrations, as one made before a migration was edited is, is not
	newDB(t, dir, "site/c/cdefghij.sqlite", "jobsdb")
	require.ErrorContains(t, check(ctx, dir), "site/c/cdefghij.sqlite does not fit its migrations")
	require.NoError(t, os.Remove(filepath.Join(dir, "site/c/cdefghij.sqlite")))

	// Nor is a file that is not a database at all
	require.NoError(t, os.WriteFile(filepath.Join(dir, "controldb.sqlite"), []byte("not a DB"), 0o600))
	require.ErrorContains(t, check(ctx, dir), "controldb.sqlite cannot be migrated")
}
