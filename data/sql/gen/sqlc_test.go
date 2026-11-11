package gen

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// The generated queries join the tables, not the _vw views sqlc types them with, and nothing else is changed
func TestJoinTablesNotViews(t *testing.T) {
	tests := []struct {
		name  string
		query string
		want  string
	}{
		{"left join on a view", "LEFT JOIN glucose_readings_vw gr ON eg.glucose_reading_id = gr.id", "LEFT JOIN glucose_readings gr ON eg.glucose_reading_id = gr.id"},
		{"every view joined", "LEFT JOIN temp_basals_vw tb ON x\nLEFT JOIN activities_vw a ON y", "LEFT JOIN temp_basals tb ON x\nLEFT JOIN activities a ON y"},
		{"a table joined", "LEFT JOIN glucose_readings gr ON x", "LEFT JOIN glucose_readings gr ON x"},
		{"a view not joined", "SELECT * FROM glucose_readings_vw", "SELECT * FROM glucose_readings_vw"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			folder := t.TempDir()
			file := filepath.Join(folder, "event_group.sql.gen.go")
			require.NoError(t, os.WriteFile(file, []byte(test.query), 0o644))
			require.NoError(t, joinTablesNotViews(folder))
			got, err := os.ReadFile(file)
			require.NoError(t, err)
			require.Equal(t, test.want, string(got))
		})
	}
}

// A bound LIMIT or OFFSET must be wrapped in CAST so SQLite does not re-prepare the query on every run
func TestCheckLimitsCast(t *testing.T) {
	tests := []struct {
		name  string
		query string
		bare  bool
	}{
		{"bare sqlc.arg limit", "SELECT * FROM t LIMIT sqlc.arg('Limit');", true},
		{"bare sqlc.arg offset", "SELECT * FROM t LIMIT CAST(sqlc.arg('Limit') AS INTEGER) OFFSET sqlc.arg('Skip');", true},
		{"bare sqlc.narg limit", "SELECT * FROM t LIMIT sqlc.narg('Limit');", true},
		{"bare ? limit on its own line", "SELECT * FROM t\nLIMIT\n    ?\n", true},
		{"bare ?N limit", "SELECT * FROM t LIMIT ?1 OFFSET ?2;", true},
		{"bare @name limit", "SELECT * FROM t LIMIT @limit;", true},
		{"bare limit in parentheses", "SELECT * FROM t LIMIT (sqlc.arg('Limit'));", true},
		{"cast limit and offset", "SELECT * FROM t LIMIT CAST(sqlc.arg('Limit') AS INTEGER) OFFSET CAST(sqlc.arg('Skip') AS INTEGER);", false},
		{"constant limit and offset", "SELECT * FROM t LIMIT 1 OFFSET 9999;", false},
		{"limit in a comment", "-- Limit to 20 minutes search window\nSELECT 1;", false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			folder := t.TempDir()
			require.NoError(t, os.Mkdir(filepath.Join(folder, "queries"), 0o755))
			require.NoError(t, os.WriteFile(filepath.Join(folder, "queries", "q.sql"), []byte(test.query), 0o644))
			err := checkLimitsCast(folder)
			if test.bare {
				require.ErrorContains(t, err, "q.sql:")
			} else {
				require.NoError(t, err)
			}
		})
	}
}
