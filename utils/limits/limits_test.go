package limits

import (
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/sitedb"
)

// TestOverridesNarrow checks each override is of a column there is, and only narrows its limit, so the site never takes
// what its column does not
func TestOverridesNarrow(t *testing.T) {
	for column, override := range Overrides {
		limit, ok := Column(column)
		require.True(t, ok, "%s is not a column", column)
		require.True(t, narrows(limit, override), "%s's override %+v takes more than its column %+v", column, override, limit)
	}
}

// TestNarrows checks an override that takes what its column does not is caught
func TestNarrows(t *testing.T) {
	dose, _ := Column("insulin_doses.Dose")
	reading, _ := Column("glucose_readings.Mgdl")
	keyID, _ := Column("push_credentials.KeyID")
	notes, _ := Column("insulin_doses.Notes")
	tests := []struct {
		name     string
		column   sitedb.ColumnLimit
		override sitedb.ColumnLimit
		want     bool
	}{
		{"inside its column", dose, sitedb.ColumnLimit{Min: new(-100.0), Max: new(100.0)}, true},
		{"a min below its column's", dose, sitedb.ColumnLimit{Min: new(-600.0)}, false},
		{"a max above its column's", dose, sitedb.ColumnLimit{Max: new(600.0)}, false},
		{"a min at an excluded bound", reading, sitedb.ColumnLimit{Min: new(0.0)}, false},
		{"a shorter text", keyID, sitedb.ColumnLimit{MaxLength: 10}, true},
		{"a longer text", keyID, sitedb.ColumnLimit{MaxLength: 33}, false},
		{"a length of a text of any length", notes, sitedb.ColumnLimit{MaxLength: 2000}, false},
		{"a required column", dose, sitedb.ColumnLimit{Required: true}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, narrows(tt.column, tt.override))
		})
	}
}

// narrows is whether an override takes only what its column does: its bounds are ones the column takes, its length
// no more than the column's, and it says nothing else
func narrows(column, override sitedb.ColumnLimit) bool {
	bounds := (override.Min == nil || column.Takes(*override.Min)) && (override.Max == nil || column.Takes(*override.Max))
	length := override.MaxLength == 0 || (column.MaxLength > 0 && override.MaxLength <= column.MaxLength)
	rest := override == sitedb.ColumnLimit{Min: override.Min, Max: override.Max, MaxLength: override.MaxLength}
	return bounds && length && rest
}
