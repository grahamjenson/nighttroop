package types_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/utils/types"
)

// A client's id is given back as sent, as a string: a string as it is, a whole number with all its figures, past
// those a float keeps, and anything else as a ShortString reads it, whole, as the column limits decide its length
func TestClientID(t *testing.T) {
	tests := []struct{ sent, want string }{
		{`"4103"`, "4103"},
		{`4103`, "4103"},
		{`-17`, "-17"},
		{`1234567890123456789`, "1234567890123456789"},
		{`"5E3C1D2A-0000-4000-8000-000000000001"`, "5E3C1D2A-0000-4000-8000-000000000001"},
		{`4103.5`, "4103.5"},
		{`4.103e3`, "4103"},
		{`true`, "true"},
		{`"` + strings.Repeat("a", 70) + `"`, strings.Repeat("a", 70)},
		{strings.Repeat("9", 70), strings.Repeat("9", 70)},
	}
	for _, test := range tests {
		t.Run(test.sent, func(t *testing.T) {
			var id types.ClientID
			require.NoError(t, json.Unmarshal([]byte(test.sent), &id))
			require.Equal(t, test.want, string(id))
			written, err := json.Marshal(id)
			require.NoError(t, err)
			require.Equal(t, `"`+test.want+`"`, string(written))
		})
	}
}
