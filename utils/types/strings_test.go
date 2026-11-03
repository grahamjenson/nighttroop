package types_test

import (
	"encoding/json"
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
