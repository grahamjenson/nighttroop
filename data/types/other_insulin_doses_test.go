package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/types"
)

// A dose's other attributes are NULL with none set, and otherwise a JSON object of only those set, which reads back,
// a loop's automatic kept when false
func TestInsulinDoseOther(t *testing.T) {
	now, ext, automatic := 0.7, 0.3, false
	for _, test := range []struct {
		name  string
		other types.InsulinDoseOther
		value any
	}{
		{"none set is NULL", types.InsulinDoseOther{}, nil},
		{"only those set", types.InsulinDoseOther{SplitNow: &now, SplitExt: &ext}, `{"split_now":0.7,"split_ext":0.3}`},
		{"automatic false is kept", types.InsulinDoseOther{Automatic: &automatic}, `{"automatic":false}`},
	} {
		t.Run(test.name, func(t *testing.T) {
			value, err := test.other.Value()
			require.NoError(t, err)
			require.Equal(t, test.value, value)

			var read types.InsulinDoseOther
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
