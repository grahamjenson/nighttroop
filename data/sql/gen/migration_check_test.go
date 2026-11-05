package gen

import (
	"database/sql"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

// openTestDB is a DB on disk with foreign keys on, as the server opens one
func openTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite", "file:"+filepath.Join(t.TempDir(), "test.sqlite")+"?_foreign_keys=1")
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return db
}

// rebuild makes table t again with name's CHECK at most n long, keeping its rows, as docs/migrations.md says to
func rebuild(n string) string {
	return `PRAGMA defer_foreign_keys = ON;
CREATE TEMP TABLE t_copy AS SELECT * FROM t;
DROP TABLE t;
CREATE TABLE t (id INTEGER PRIMARY KEY, name TEXT NOT NULL CHECK(length(name) <= ` + n + `)) STRICT;
INSERT INTO t SELECT * FROM t_copy;
DROP TABLE t_copy;`
}

func migrations(files map[string]string) fstest.MapFS {
	fs := fstest.MapFS{}
	for name, sql := range files {
		fs["migrations/"+name] = &fstest.MapFile{Data: []byte(sql)}
	}
	return fs
}

func TestCheckMigrationsWithData(t *testing.T) {
	first := map[string]string{
		"000001_t.up.sql":   `CREATE TABLE t (id INTEGER PRIMARY KEY, name TEXT NOT NULL CHECK(length(name) <= 10)) STRICT;`,
		"000001_t.down.sql": `DROP TABLE t;`,
	}
	with := func(files map[string]string) map[string]string {
		all := map[string]string{}
		for name, sql := range first {
			all[name] = sql
		}
		for name, sql := range files {
			all[name] = sql
		}
		return all
	}
	tests := []struct {
		name  string
		files map[string]string
		err   string
	}{
		{"one migration", first, ""},
		{"a migration that widens a CHECK, but narrows it again going down", with(map[string]string{"000002_wider.up.sql": rebuild("20"), "000002_wider.down.sql": rebuild("10")}),
			"migrating down from version 2"},
		{"a migration that narrows a CHECK", with(map[string]string{"000002_narrower.up.sql": rebuild("5"), "000002_narrower.down.sql": rebuild("10")}),
			"migrating up to version 2"},
		{"a migration that adds a table", with(map[string]string{
			"000002_u.up.sql":   `CREATE TABLE u (id TEXT PRIMARY KEY CHECK(length(id) <= 8), t_id INTEGER REFERENCES t(id)) STRICT;`,
			"000002_u.down.sql": `DROP TABLE u;`,
		}), ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckMigrationsWithData(migrations(tt.files), openTestDB(t))
			if tt.err == "" {
				require.NoError(t, err)
				return
			}
			require.ErrorContains(t, err, tt.err)
			require.ErrorContains(t, err, "CHECK constraint failed")
		})
	}
}

// Each column is at the largest and the smallest its CHECKs allow, and a foreign key at its parent's row at the same end
func TestSeedLimits(t *testing.T) {
	db := openTestDB(t)
	_, err := db.Exec(`
CREATE TABLE parent (id TEXT PRIMARY KEY CHECK(length(id) <= 8)) STRICT;
CREATE TABLE child (
	id INTEGER PRIMARY KEY,
	parent_id TEXT NOT NULL REFERENCES parent(id) ON DELETE RESTRICT,
	name TEXT NOT NULL CHECK(length(name) <= 12),
	token TEXT UNIQUE CHECK(length(token) = 4),
	minutes INTEGER CHECK(minutes BETWEEN -5 AND 90),
	rate REAL CHECK(rate BETWEEN 0 AND 2.5),
	state TEXT NOT NULL CHECK(state IN ('live', 'archived')),
	settings_json TEXT CHECK(length(settings_json) <= 10 AND json_valid(settings_json)),
	notes TEXT,
	doubled INTEGER GENERATED ALWAYS AS (minutes * 2) VIRTUAL
) STRICT;
CREATE TABLE settings (id INTEGER PRIMARY KEY CHECK(id = 1), on_off INTEGER NOT NULL DEFAULT 0 CHECK(on_off IN (0, 1))) STRICT;
INSERT INTO settings (id) VALUES (1);`)
	require.NoError(t, err)
	require.NoError(t, SeedLimits(db))

	type row struct {
		parent, name, token, state, settings string
		minutes                              int64
		rate                                 float64
	}
	got := []row{}
	rows, err := db.Query(`SELECT parent_id, name, token, state, settings_json, minutes, rate FROM child ORDER BY id`)
	require.NoError(t, err)
	for rows.Next() {
		var r row
		require.NoError(t, rows.Scan(&r.parent, &r.name, &r.token, &r.state, &r.settings, &r.minutes, &r.rate))
		got = append(got, r)
	}
	require.NoError(t, rows.Close())
	require.Equal(t, []row{
		{strings.Repeat("z", 8), strings.Repeat("z", 12), "zzzz", "archived", `"zzzzzzzz"`, 90, 2.5},
		{"a", "a", "aaaa", "live", `"a"`, -5, 0},
	}, got)

	// A table of one row keeps it
	var settings int
	require.NoError(t, db.QueryRow(`SELECT COUNT(*) FROM settings`).Scan(&settings))
	require.Equal(t, 1, settings)
}
