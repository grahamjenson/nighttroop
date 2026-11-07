package main

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/require"

	"nighttroop.t1d.nz/data/sql/migrateup"
	"nighttroop.t1d.nz/data/sql/sitedb"
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
	sitePrepared = false
	require.NoError(t, tryMigrateOrDelete(ctx, dataDir))
	require.FileExists(t, filepath.Join(dataDir, "db", "controldb.sqlite"))
	require.FileExists(t, filepath.Join(dataDir, "db", "site/a/abcdefgh.sqlite"))
	require.FileExists(t, backup)
	require.True(t, sitePrepared, "site DBs have their queries prepared once")

	// A site DB whose tables its queries do not fit, as one made before a migration was edited, deletes every DB and
	// backup, not only it
	newDB(t, dataDir, "site/b/bcdefghi.sqlite", "jobsdb")
	require.NoError(t, tryMigrateOrDelete(ctx, dataDir))
	require.NoDirExists(t, filepath.Join(dataDir, "db"))
	require.NoDirExists(t, filepath.Join(dataDir, "backup"))
}

// editedMigrations are the site DB's migrations with old replaced by new in each, as before a migration was edited
func editedMigrations(t *testing.T, old, new string) fstest.MapFS {
	t.Helper()
	fsys := fstest.MapFS{}
	edited := false
	require.NoError(t, fs.WalkDir(sitedb.Migrations, "migrations", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		data, err := fs.ReadFile(sitedb.Migrations, path)
		edited = edited || strings.Contains(string(data), old)
		fsys[path] = &fstest.MapFile{Data: []byte(strings.ReplaceAll(string(data), old, new))}
		return err
	}))
	require.True(t, edited, "a migration has %q", old)
	return fsys
}

func TestTryMigrateOrDeleteChecksTheSchema(t *testing.T) {
	ctx := context.Background()
	dataDir := t.TempDir()
	newDB(t, dataDir, "controldb.sqlite", "controldb")
	newDB(t, dataDir, "site/a/abcdefgh.sqlite", "sitedb")
	require.NoError(t, tryMigrateOrDelete(ctx, dataDir), "DBs with the schema their migrations make are kept")
	require.FileExists(t, filepath.Join(dataDir, "db", "site/a/abcdefgh.sqlite"))

	// A site DB made before a CHECK was changed is at the same version, and every query still works on it, but it keeps
	// the old rule, so it deletes every DB
	db, err := sitedb.OpenSqliteDB(filepath.Join(dataDir, "db", "site/b/bcdefghi.sqlite"), 0)
	require.NoError(t, err)
	require.NoError(t, migrateup.Up("sitedb", db, editedMigrations(t, "mgdl <= 1000", "mgdl <= 2000")))
	queries, err := sitedb.Prepare(ctx, db)
	require.NoError(t, err, "its queries still work")
	require.NoError(t, queries.Close())
	require.NoError(t, db.Close())
	require.ErrorContains(t, openers["sitedb"](ctx, filepath.Join(dataDir, "db", "site/b/bcdefghi.sqlite")), "mgdl <= 2000")
	require.NoError(t, tryMigrateOrDelete(ctx, dataDir))
	require.NoDirExists(t, filepath.Join(dataDir, "db"))
}
