package types_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"nighttroop.t1d.nz/utils/types"
)

// An id that is not an identifier, such as null or an empty string, is none rather than a bad upload
func TestIdentifierFromJSON(t *testing.T) {
	tests := []struct {
		data string
		want *string
	}{
		{"", nil},
		{"null", nil},
		{`""`, nil},
		{`"6F1C3B0E-7A8A-4C55-9D42-1B5E5E0B9A11"`, new("6F1C3B0E-7A8A-4C55-9D42-1B5E5E0B9A11")},
		{"123", new("123")},
	}
	for _, tt := range tests {
		t.Run(tt.data, func(t *testing.T) {
			got := types.IdentifierFromJSON(json.RawMessage(tt.data))
			if tt.want == nil {
				assert.Nil(t, got)
				return
			}
			assert.Equal(t, *tt.want, got.String())
		})
	}
}

// A pumpId is given back with the JSON type it was sent with, cut to 64 characters; one neither a number nor a string
// is empty and not stored
func TestShortNumberOrString(t *testing.T) {
	long := `"` + strings.Repeat("a", 70) + `"`
	tests := []struct {
		data, want string
	}{
		{"1794456578717", "1794456578717"},
		{`"1794456578717"`, `"1794456578717"`},
		{`"DANA_I"`, `"DANA_I"`},
		{"4102.5", "4102.5"},
		{long, `"` + strings.Repeat("a", 64) + `"`},
		{"true", `""`},
		{`{"id": 1}`, `""`},
	}
	for _, tt := range tests {
		t.Run(tt.data, func(t *testing.T) {
			var got types.ShortNumberOrString
			assert.NoError(t, json.Unmarshal([]byte(tt.data), &got))
			out, err := json.Marshal(got)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, string(out))
		})
	}
}
