package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/types"
)

// A row's other attributes are NULL with none set, and otherwise a JSON object of only those set, which reads back
func TestGlucoseReadingOther(t *testing.T) {
	noise, rssi := 1.0, -62.0
	for _, test := range []struct {
		name  string
		other types.GlucoseReadingOther
		value any
	}{
		{"none set is NULL", types.GlucoseReadingOther{}, nil},
		{"only those set", types.GlucoseReadingOther{Noise: &noise, Rssi: &rssi}, `{"noise":1,"rssi":-62}`},
	} {
		t.Run(test.name, func(t *testing.T) {
			value, err := test.other.Value()
			require.NoError(t, err)
			require.Equal(t, test.value, value)

			var read types.GlucoseReadingOther
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
