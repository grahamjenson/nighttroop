package types_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/utils/types"
)

func TestDateStringUnmarshalNumbers(t *testing.T) {
	tests := []struct {
		json string
		want string
	}{
		{json: `1719300000000`, want: "2024-06-25T07:20:00Z"},
		{json: `1719300000000.5`, want: "2024-06-25T07:20:00Z"},
		{json: `1.7193e12`, want: "2024-06-25T07:20:00Z"},
		{json: `"2024-06-25T07:20:00Z"`, want: "2024-06-25T07:20:00Z"},
	}
	for _, test := range tests {
		t.Run(test.json, func(t *testing.T) {
			var date types.DateString
			require.NoError(t, json.Unmarshal([]byte(test.json), &date))
			require.Equal(t, test.want, date.String())
		})
	}
}
