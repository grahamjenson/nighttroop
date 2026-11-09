package types_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/types"
)

// A device status's other attributes are NULL with none set, and otherwise a JSON object of only those set, its JSON
// ones nested as JSON rather than as text, which reads back
func TestDeviceReportOther(t *testing.T) {
	clock, model, loop := int64(1704067200000), "Omnipod", "loop"
	for _, test := range []struct {
		name  string
		other types.DeviceReportOther
		value any
	}{
		{"none set is NULL", types.DeviceReportOther{}, nil},
		{"only those set", types.DeviceReportOther{PumpClock: &clock, PumpModel: &model},
			`{"pump_clock":1704067200000,"pump_model":"Omnipod"}`},
		{"JSON nested as JSON", types.DeviceReportOther{LoopName: &loop, Override: json.RawMessage(`{"active":false}`)},
			`{"override":{"active":false},"loop_name":"loop"}`},
	} {
		t.Run(test.name, func(t *testing.T) {
			value, err := test.other.Value()
			require.NoError(t, err)
			require.Equal(t, test.value, value)

			var read types.DeviceReportOther
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
