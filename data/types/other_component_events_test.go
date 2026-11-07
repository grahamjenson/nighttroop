package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/types"
)

// A component event's other attributes are NULL with none set, and otherwise a JSON object of only those set, which
// reads back
func TestComponentEventOther(t *testing.T) {
	code, serial, pumpType := "1234", "P123", "Omnipod"
	for _, test := range []struct {
		name  string
		other types.ComponentEventOther
		value any
	}{
		{"none set is NULL", types.ComponentEventOther{}, nil},
		{"only those set", types.ComponentEventOther{SensorCode: &code}, `{"sensor_code":"1234"}`},
		{"a pump's", types.ComponentEventOther{PumpSerial: &serial, PumpType: &pumpType}, `{"pump_serial":"P123","pump_type":"Omnipod"}`},
	} {
		t.Run(test.name, func(t *testing.T) {
			value, err := test.other.Value()
			require.NoError(t, err)
			require.Equal(t, test.value, value)

			var read types.ComponentEventOther
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
