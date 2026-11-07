package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/types"
)

// A treatment's other attributes are NULL with none set, and otherwise a JSON object of only those set, which reads
// back, AAPS's autoForced false and empty reasons kept
func TestEventGroupOther(t *testing.T) {
	uuid, mode, duration := "abc-123", "DISCONNECTED_PUMP", int64(1800000)
	autoForced, reasons := false, ""
	for _, test := range []struct {
		name  string
		other types.EventGroupOther
		value any
	}{
		{"none set is NULL", types.EventGroupOther{}, nil},
		{"only those set", types.EventGroupOther{Uuid: &uuid}, `{"uuid":"abc-123"}`},
		{"a running mode", types.EventGroupOther{Mode: &mode, AutoForced: &autoForced, Reasons: &reasons, OriginalDurationMs: &duration},
			`{"mode":"DISCONNECTED_PUMP","auto_forced":false,"reasons":"","original_duration_ms":1800000}`},
	} {
		t.Run(test.name, func(t *testing.T) {
			value, err := test.other.Value()
			require.NoError(t, err)
			require.Equal(t, test.value, value)

			var read types.EventGroupOther
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
