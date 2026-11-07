package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/types"
)

// A food intake's other attributes are NULL with none set, and otherwise a JSON object of only those set, which reads
// back
func TestFoodIntakeOther(t *testing.T) {
	category, fat, preBolus := "Breakfast", 12.5, -15.0
	for _, test := range []struct {
		name  string
		other types.FoodIntakeOther
		value any
	}{
		{"none set is NULL", types.FoodIntakeOther{}, nil},
		{"only those set", types.FoodIntakeOther{Category: &category, Fat: &fat, PreBolus: &preBolus},
			`{"category":"Breakfast","fat":12.5,"pre_bolus":-15}`},
	} {
		t.Run(test.name, func(t *testing.T) {
			value, err := test.other.Value()
			require.NoError(t, err)
			require.Equal(t, test.value, value)

			var read types.FoodIntakeOther
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
