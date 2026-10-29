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
