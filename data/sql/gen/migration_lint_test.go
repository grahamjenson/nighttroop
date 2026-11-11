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
