package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/types"
)

// A device status's other attributes, those without a column, of which there are none yet, are NULL
func TestDeviceReportOther(t *testing.T) {
	value, err := types.DeviceReportOther{}.Value()
	require.NoError(t, err)
	require.Nil(t, value)
	var read types.DeviceReportOther
	require.NoError(t, read.Scan(value))
	require.NoError(t, read.Scan(`{}`))
	require.Nil(t, types.OtherJSON(read))
}
