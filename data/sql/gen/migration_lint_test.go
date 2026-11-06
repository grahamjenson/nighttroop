package gen

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLintMigrations(t *testing.T) {
	tests := []struct {
		name, sql, err string
	}{
		{"no PRAGMA", `CREATE TABLE t (id INTEGER PRIMARY KEY) STRICT;`, ""},
		{"defer_foreign_keys", "PRAGMA defer_foreign_keys = ON;\nDROP TABLE t;", ""},
		{"defer_foreign_keys of a schema, in capitals", `pragma main.DEFER_FOREIGN_KEYS = ON;`, ""},
		{"a PRAGMA in a comment", "-- not PRAGMA foreign_keys = OFF, which does nothing here\n/* nor PRAGMA journal_mode */\nDROP TABLE t;", ""},
		{"a table-valued pragma function", `INSERT INTO u SELECT name FROM pragma_table_info('t');`, ""},
		{"foreign_keys, which does nothing in a transaction", `PRAGMA foreign_keys = OFF;`, "000002_x.up.sql runs PRAGMA foreign_keys"},
		{"ignore_check_constraints, which lasts past it", `CREATE TABLE u (id INTEGER PRIMARY KEY) STRICT; PRAGMA ignore_check_constraints = ON;`, "000002_x.up.sql runs PRAGMA ignore_check_constraints"},
		{"a PRAGMA of a schema", `PRAGMA main.journal_mode = DELETE;`, "000002_x.up.sql runs PRAGMA journal_mode"},
		{"a PRAGMA reading", `PRAGMA foreign_key_check;`, "000002_x.up.sql runs PRAGMA foreign_key_check"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := LintMigrations(migrations(map[string]string{
				"000001_t.up.sql":   `CREATE TABLE t (id INTEGER PRIMARY KEY) STRICT;`,
				"000001_t.down.sql": `DROP TABLE t;`,
				"000002_x.up.sql":   tt.sql,
				"000002_x.down.sql": ``,
			}))
			if tt.err == "" {
				require.NoError(t, err)
				return
			}
			require.ErrorContains(t, err, tt.err)
			require.ErrorContains(t, err, "only PRAGMA defer_foreign_keys is allowed")
		})
	}
}

// A migration marked additive leaves the DB one a build without it can go on using
func TestLintAdditiveMigrations(t *testing.T) {
	tests := []struct {
		name, sql, err string
	}{
		{"a new table, index, view and column", "-- additive\nCREATE TABLE u (id INTEGER PRIMARY KEY, t_id INTEGER REFERENCES t(id) ON DELETE CASCADE) STRICT;\n" +
			"CREATE INDEX u_t ON u(t_id);\nCREATE VIEW u_vw AS SELECT * FROM u;\nALTER TABLE t ADD COLUMN note TEXT CHECK(length(note) <= 32);", ""},
		{"a column referencing a table ON DELETE SET NULL", "-- additive\nALTER TABLE t ADD u_id INTEGER REFERENCES u(id) ON DELETE SET NULL;", ""},
		{"not marked, so it may change the schema", "DROP TABLE t;", ""},
		{"additive only in a comment of a line", "-- not -- additive\nDROP TABLE t;", ""},
		{"a dropped table", "-- additive\nDROP TABLE t;", `000002_x.up.sql runs "DROP TABLE t"`},
		{"a dropped column", "-- additive\nALTER TABLE t DROP COLUMN note;", `runs "ALTER TABLE t DROP COLUMN note"`},
		{"a renamed table", "-- additive\nALTER TABLE t RENAME TO u;", `runs "ALTER TABLE t RENAME TO u"`},
		{"a renamed column", "-- additive\nALTER TABLE t RENAME COLUMN id TO t_id;", `runs "ALTER TABLE t RENAME COLUMN id TO t_id"`},
		{"a table made again, as a changed CHECK needs", "-- additive\nCREATE TEMP TABLE t_copy AS SELECT * FROM t;", `runs "CREATE TEMP TABLE`},
		{"changed rows", "-- additive\nUPDATE t SET id = id + 1;", `runs "UPDATE t SET id = id + 1"`},
		{"a unique index, which could refuse its rows", "-- additive\nCREATE UNIQUE INDEX t_id ON t(id);", `runs "CREATE UNIQUE INDEX`},
		{"a trigger", "-- additive\nCREATE TRIGGER t_ins AFTER INSERT ON t BEGIN SELECT 1; END;", `runs "CREATE TRIGGER`},
		{"a reference that stops a delete", "-- additive\nCREATE TABLE u (t_id INTEGER REFERENCES t(id)) STRICT;", "000002_x.up.sql references a table without ON DELETE CASCADE"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := LintMigrations(migrations(map[string]string{
				"000001_t.up.sql":   `CREATE TABLE t (id INTEGER PRIMARY KEY) STRICT;`,
				"000001_t.down.sql": `DROP TABLE t;`,
				"000002_x.up.sql":   tt.sql,
				"000002_x.down.sql": "-- additive\nDROP TABLE u;",
			}))
			if tt.err == "" {
				require.NoError(t, err)
				return
			}
			require.ErrorContains(t, err, tt.err)
			require.ErrorContains(t, err, "a build without an additive migration goes on using the DB after it")
		})
	}
}
