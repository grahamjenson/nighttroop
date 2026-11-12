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

// A device token is kept whole, as Apple's are of no fixed length, and one too long to be a token is empty rather than
// cut or refuse the upload
func TestDeviceTokenString(t *testing.T) {
	tests := []struct {
		name  string
		token string
		want  string
	}{
		{"a 64 character token", strings.Repeat("ab", 32), strings.Repeat("ab", 32)},
		{"Trio's 160 character token", strings.Repeat("ab", 80), strings.Repeat("ab", 80)},
		{"the longest token", strings.Repeat("a", types.MaxDeviceTokenLength), strings.Repeat("a", types.MaxDeviceTokenLength)},
		{"a token too long", strings.Repeat("a", types.MaxDeviceTokenLength+1), ""},
		{"a token that is not hex", "asd", "asd"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.token)
			assert.NoError(t, err)
			var got types.DeviceTokenString
			assert.NoError(t, json.Unmarshal(data, &got))
			assert.Equal(t, tt.want, string(got))
		})
	}
}
