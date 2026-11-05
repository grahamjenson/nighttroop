package gen

import (
	"database/sql"
	"maps"
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
	with := func(more ...map[string]string) map[string]string {
		all := maps.Clone(first)
		for _, files := range more {
			maps.Copy(all, files)
		}
		return all
	}
	// t made again with its columns in another order, and its rows copied with SELECT *, which swaps name and nick
	swapped := `PRAGMA defer_foreign_keys = ON;
CREATE TEMP TABLE t_copy AS SELECT * FROM t;
DROP TABLE t;
CREATE TABLE t (id INTEGER PRIMARY KEY, nick TEXT NOT NULL DEFAULT 'nick', name TEXT NOT NULL CHECK(length(name) <= 10)) STRICT;
INSERT INTO t SELECT * FROM t_copy;
DROP TABLE t_copy;`
	withNick := map[string]string{
		"000002_nick.up.sql":   `ALTER TABLE t ADD COLUMN nick TEXT NOT NULL DEFAULT 'nick';`,
		"000002_nick.down.sql": `ALTER TABLE t DROP COLUMN nick;`,
	}
	tests := []struct {
		name  string
		files map[string]string
		errs  []string
	}{
		{"one migration", first, nil},
		{"a migration that widens a CHECK, but narrows it again going down", with(map[string]string{"000002_wider.up.sql": rebuild("20"), "000002_wider.down.sql": rebuild("10")}),
			[]string{"migrating down from version 2", "CHECK constraint failed"}},
		{"a migration that narrows a CHECK", with(map[string]string{"000002_narrower.up.sql": rebuild("5"), "000002_narrower.down.sql": rebuild("10")}),
			[]string{"migrating up to version 2", "CHECK constraint failed"}},
		{"a migration that adds a table", with(map[string]string{
			"000002_u.up.sql":   `CREATE TABLE u (id TEXT PRIMARY KEY CHECK(length(id) <= 8), t_id INTEGER REFERENCES t(id)) STRICT;`,
			"000002_u.down.sql": `DROP TABLE u;`,
		}), nil},
		{"a migration that adds a column", with(withNick), nil},
		{"a migration that makes a table again keeping its rows", with(map[string]string{"000002_same.up.sql": rebuild("10"), "000002_same.down.sql": rebuild("10")}), nil},
		{"a migration that swaps columns", with(withNick, map[string]string{"000003_swap.up.sql": swapped, "000003_swap.down.sql": ""}),
			[]string{"migrating up to version 3", "a row of t was lost or changed"}},
		{"a migration that loses rows", with(map[string]string{"000002_lose.up.sql": `DELETE FROM t WHERE length(name) > 5;`, "000002_lose.down.sql": ""}),
			[]string{"migrating up to version 2", "a row of t was lost or changed"}},
		{"a migration that says it changes rows", with(map[string]string{"000002_lose.up.sql": "-- changes rows of t\nDELETE FROM t WHERE length(name) > 5;", "000002_lose.down.sql": ""}), nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckMigrationsWithData(migrations(tt.files), openTestDB(t))
			if tt.errs == nil {
				require.NoError(t, err)
				return
			}
			for _, want := range tt.errs {
				require.ErrorContains(t, err, want)
			}
		})
	}
}

// Each column is at the largest and the smallest its CHECKs allow, and a foreign key at its parent's row at the same end.
// A text with no exact value is made of a letter for its column, so columns swapped hold different rows.
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
		{strings.Repeat("z", 8), strings.Repeat("x", 12), "wwww", "archived", `"ssssssss"`, 90, 2.5},
		{"a", "c", "dddd", "live", `"h"`, -5, 0},
	}, got)

	// A table of one row keeps it
	var settings int
	require.NoError(t, db.QueryRow(`SELECT COUNT(*) FROM settings`).Scan(&settings))
	require.Equal(t, 1, settings)
}

// A table of only a row ID gets its rows too
func TestSeedLimitsTableOfOnlyARowID(t *testing.T) {
	db := openTestDB(t)
	_, err := db.Exec(`CREATE TABLE ids (id INTEGER PRIMARY KEY) STRICT;`)
	require.NoError(t, err)
	require.NoError(t, SeedLimits(db))

	var count int
	require.NoError(t, db.QueryRow(`SELECT COUNT(*) FROM ids`).Scan(&count))
	require.Equal(t, 2, count)
}
