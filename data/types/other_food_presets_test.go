package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/types"
)

// A food preset's other attributes are NULL with none set, and otherwise a JSON object of only those set, which reads
// back
func TestFoodPresetOther(t *testing.T) {
	subcategory, protein, portions := "Bread", 4.5, 2.0
	for _, test := range []struct {
		name  string
		other types.FoodPresetOther
		value any
	}{
		{"none set is NULL", types.FoodPresetOther{}, nil},
		{"only those set", types.FoodPresetOther{Subcategory: &subcategory, Protein: &protein, Portions: &portions},
			`{"subcategory":"Bread","protein":4.5,"portions":2}`},
	} {
		t.Run(test.name, func(t *testing.T) {
			value, err := test.other.Value()
			require.NoError(t, err)
			require.Equal(t, test.value, value)

			var read types.FoodPresetOther
			require.NoError(t, read.Scan(value))
			require.Equal(t, test.other, read)

			// An update sends those set as JSON, and nothing for none
			if json := types.OtherJSON(test.other); test.value == nil {
				require.Nil(t, json)
			} else {
				require.Equal(t, test.value, *json)
			}
		})
	}
}
