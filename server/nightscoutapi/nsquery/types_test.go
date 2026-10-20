package nsquery_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/server/nightscoutapi/nsquery"
)

// The values of $in and $nin go to SQLite's json_each, so they must be valid JSON whatever they contain
func TestQueryStringArrayStringpIsJSON(t *testing.T) {
	tests := [][]string{
		{"abc"},
		{"a", "b"},
		{`a"b`},
		{`back\slash`},
		{"new\nline"},
	}
	for _, values := range tests {
		t.Run(values[0], func(t *testing.T) {
			qsa := nsquery.QueryStringArray(values)
			var decoded []string
			require.NoError(t, json.Unmarshal([]byte(*qsa.Stringp()), &decoded))
			require.Equal(t, values, decoded)
		})
	}
}
