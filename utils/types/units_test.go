package types_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/types"
)

// Units are stored, and written as JSON, as they were sent, and read back with what they mean
func TestUnitsKeepWhatWasSent(t *testing.T) {
	tests := []struct {
		sent string
		want consts.Units
	}{
		{"mg/dl", consts.UnitsMGDL},
		{"mg/dL", consts.UnitsMGDL},
		{"mgdl", consts.UnitsMGDL},
		{"mmol", consts.UnitsMMOLL},
		{"mmol/L", consts.UnitsMMOLL},
	}
	for _, tt := range tests {
		t.Run(tt.sent, func(t *testing.T) {
			var scanned types.Units
			require.NoError(t, scanned.Scan(tt.sent))
			assert.Equal(t, types.Units{Units: tt.want, Sent: tt.sent}, scanned)

			stored, err := scanned.Value()
			require.NoError(t, err)
			assert.Equal(t, tt.sent, stored)

			data, err := json.Marshal(scanned)
			require.NoError(t, err)
			var read types.Units
			require.NoError(t, json.Unmarshal(data, &read))
			assert.Equal(t, scanned, read)
		})
	}

	// Units from what they mean are written as Nighttroop writes them
	stored, err := types.UnitsFromEnum(consts.UnitsMMOLL).Value()
	require.NoError(t, err)
	assert.Equal(t, "mmol/L", stored)

	var bad types.Units
	assert.Error(t, bad.Scan("furlongs"))
	assert.Error(t, bad.Scan(42))
}
