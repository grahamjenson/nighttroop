package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// A DB's tables are kept as their JSON, and decoded only by Tables: none measured is NULL in the DB, as an empty
// string is not JSON, and reads as no tables
func TestTableSizesAreDecodedWhereRead(t *testing.T) {
	var none TableSizes
	require.NoError(t, none.Scan(nil))
	value, err := none.Value()
	require.NoError(t, err)
	require.Nil(t, value)
	tables, err := none.Tables()
	require.NoError(t, err)
	require.Empty(t, tables)

	kept, err := NewTableSizes(map[string]TableSize{"glucose_readings": {Bytes: 4096, Records: 3}})
	require.NoError(t, err)
	value, err = kept.Value()
	require.NoError(t, err)
	require.JSONEq(t, `{"glucose_readings":{"bytes":4096,"records":3}}`, value.(string))

	for _, stored := range []any{value, []byte(value.(string))} {
		var read TableSizes
		require.NoError(t, read.Scan(stored))
		tables, err = read.Tables()
		require.NoError(t, err)
		require.Equal(t, map[string]TableSize{"glucose_readings": {Bytes: 4096, Records: 3}}, tables)
	}
	require.Error(t, new(TableSizes).Scan(1))
}
