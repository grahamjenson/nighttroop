package migrateup

import (
	"bytes"
	"database/sql"
	"log/slog"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/stretchr/testify/require"

	_ "modernc.org/sqlite"
)

const (
	createA = "CREATE TABLE a (id INTEGER PRIMARY KEY);"
	createB = "CREATE TABLE b (id INTEGER PRIMARY KEY);"
	// failing fails after creating a table, which its transaction rolls back
	failing = "CREATE TABLE c (id INTEGER PRIMARY KEY); INSERT INTO missing VALUES (1);"
)

// migrations are a migrations folder with each up migration's SQL, numbered from 1
func migrations(ups ...string) fstest.MapFS {
	fsys := fstest.MapFS{}
	for i, up := range ups {
		name := "migrations/00000" + string(rune('1'+i)) + "_m"
		fsys[name+".up.sql"] = &fstest.MapFile{Data: []byte(up)}
		fsys[name+".down.sql"] = &fstest.MapFile{Data: []byte("")}
	}
	return fsys
}

// version is the version db is at and whether it is dirty
func version(t *testing.T, db *sql.DB, fsys fstest.MapFS) (uint, bool) {
	src, err := iofs.New(fsys, "migrations")
	require.NoError(t, err)
	driver, err := sqlite.WithInstance(db, &sqlite.Config{})
	require.NoError(t, err)
	m, err := migrate.NewWithInstance("iofs", src, "sqlite", driver)
	require.NoError(t, err)
	v, dirty, err := m.Version()
	require.NoError(t, err)
	return v, dirty
}

func openDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite", "file:"+filepath.Join(t.TempDir(), "test.sqlite"))
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return db
}

func TestUpRecoversADirtyDB(t *testing.T) {
	tests := []struct {
		name          string
		broken, fixed fstest.MapFS
		failed        uint
		forcedTo      string
	}{
		{"a later migration", migrations(createA, failing), migrations(createA, createB), 2, "forced_to_version=1"},
		{"the first migration", migrations(failing), migrations(createA), 1, "forced_to_version=-1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := openDB(t)

			err := Up("testdb", db, tt.broken)
			require.Error(t, err)
			require.Regexp(t, `^testdb migration \d failed: .*no such table: missing`, err.Error())
			v, dirty := version(t, db, tt.broken)
			require.Equal(t, tt.failed, v)
			require.True(t, dirty)

			// The migration keeps failing at each open, with the same plain error
			err = Up("testdb", db, tt.broken)
			require.Error(t, err)
			require.Regexp(t, `^testdb migration \d failed: .*no such table: missing`, err.Error())

			// Once fixed, the next open forces it back, says so and migrates it
			var logs bytes.Buffer
			before := slog.Default()
			slog.SetDefault(slog.New(slog.NewTextHandler(&logs, nil)))
			t.Cleanup(func() { slog.SetDefault(before) })

			require.NoError(t, Up("testdb", db, tt.fixed))
			v, dirty = version(t, db, tt.fixed)
			require.Equal(t, tt.failed, v)
			require.False(t, dirty)
			require.Contains(t, logs.String(), "db=testdb")
			require.Contains(t, logs.String(), tt.forcedTo)

			// It opens again without being forced back
			logs.Reset()
			require.NoError(t, Up("testdb", db, tt.fixed))
			require.Empty(t, logs.String())
		})
	}
}

func TestUpRefusesADBDirtyAtAMigrationItDoesNotHave(t *testing.T) {
	db := openDB(t)
	require.Error(t, Up("testdb", db, migrations(createA, createB, failing)))

	err := Up("testdb", db, migrations(createA, createB))
	require.EqualError(t, err, "testdb is dirty at migration 3, which is not one of its migrations: "+
		"prev for version 3 migrations: file does not exist")
}
