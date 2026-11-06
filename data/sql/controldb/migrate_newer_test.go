package controldb_test

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/controldb"
)

// buildMigrations are the control DB's migrations and then those of a later build
func buildMigrations(t *testing.T, later ...string) fs.FS {
	migrations := fstest.MapFS{}
	files, err := fs.Glob(controldb.Migrations, "migrations/*.sql")
	require.NoError(t, err)
	for _, file := range files {
		data, err := fs.ReadFile(controldb.Migrations, file)
		require.NoError(t, err)
		migrations[file] = &fstest.MapFile{Data: data}
	}
	for i, sql := range later {
		migrations[fmt.Sprintf("migrations/%06d_later.up.sql", i+2)] = &fstest.MapFile{Data: []byte(sql)}
	}
	return migrations
}

// A build opens a DB a build one additive migration newer migrated, without migrating it, and refuses one that is not
// additive or is two migrations newer, as it cannot know it can use them
func TestMigrateUpOpensDBOneAdditiveVersionNewer(t *testing.T) {
	additive := "-- additive\nALTER TABLE sites ADD COLUMN note TEXT;"
	tests := []struct {
		name         string
		newer, older fs.FS
		err          string
	}{
		{"one additive migration newer", buildMigrations(t, additive), buildMigrations(t), ""},
		{"one newer that is not additive", buildMigrations(t, "ALTER TABLE sites ADD COLUMN note TEXT;"), buildMigrations(t),
			"the DB is at version 2, which this build, at 1, cannot use"},
		{"two additive migrations newer", buildMigrations(t, additive, "-- additive\nCREATE TABLE notes (id INTEGER PRIMARY KEY) STRICT;"),
			buildMigrations(t), "the DB is at version 3, which this build, at 1, cannot use"},
		{"one additive migration newer of two", buildMigrations(t, additive, "-- additive\nCREATE TABLE notes (id INTEGER PRIMARY KEY) STRICT;"),
			buildMigrations(t, additive), ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "control.sqlite")
			db, err := controldb.OpenSqliteDB(path, 0)
			require.NoError(t, err)
			require.NoError(t, controldb.MigrateUpWith(db, tt.newer))
			require.NoError(t, db.Close())

			db, err = controldb.OpenSqliteDB(path, 0)
			require.NoError(t, err)
			t.Cleanup(func() { db.Close() })
			err = controldb.MigrateUpWith(db, tt.older)
			if tt.err != "" {
				require.ErrorContains(t, err, tt.err)
				return
			}
			require.NoError(t, err)
			// Its queries still work on the newer schema
			queries, err := controldb.Prepare(t.Context(), db)
			require.NoError(t, err)
			_, err = queries.GetServerSettings(t.Context())
			require.NoError(t, err)

			// The newer build opens it again
			require.NoError(t, controldb.MigrateUpWith(db, tt.newer))
		})
	}
}
