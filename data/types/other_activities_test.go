package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/types"
)

// An activity's other attributes are NULL with none set, and otherwise a JSON object of only those set, which reads back
func TestActivityOther(t *testing.T) {
	steps, heartRate := int64(1000), int64(120)
	for _, test := range []struct {
		name  string
		other types.ActivityOther
		value any
	}{
		{"none set is NULL", types.ActivityOther{}, nil},
		{"only those set", types.ActivityOther{Steps: &steps, HeartRate: &heartRate}, `{"steps":1000,"heart_rate":120}`},
	} {
		t.Run(test.name, func(t *testing.T) {
			value, err := test.other.Value()
			require.NoError(t, err)
			require.Equal(t, test.value, value)

			var read types.ActivityOther
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
