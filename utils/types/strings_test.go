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

// A short string is kept whole, as the column or JSON it is kept in limits it, and a number or a bool is kept as text
func TestShortString(t *testing.T) {
	tests := []struct {
		name, data, want string
	}{
		{"Trio's 160 character APNs token", `"` + strings.Repeat("ab", 80) + `"`, strings.Repeat("ab", 80)},
		{"longer than 64", `"` + strings.Repeat("a", 300) + `"`, strings.Repeat("a", 300)},
		{"a number", "12.5", "12.5"},
		{"a bool", "true", "true"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got types.ShortString
			assert.NoError(t, json.Unmarshal([]byte(tt.data), &got))
			assert.Equal(t, tt.want, string(got))
		})
	}
}

// A note is kept as sent, HTML, entities and the white space around it too, as xDrip+ grows its own note with any that
// comes back changed; one only of white space is none, and one too long is cut at a character, not inside one
func TestSanitizedString(t *testing.T) {
	long := strings.Repeat("é", 1030)
	tests := []struct {
		name, sent, want string
	}{
		{"plain", "Pasta", "Pasta"},
		{"xDrip+'s note", `xDrip note <b>bold</b> & "quoted" 100%% <script>x</script>`, `xDrip note <b>bold</b> & "quoted" 100%% <script>x</script>`},
		{"entities as typed", "a &amp; b &lt;c&gt;", "a &amp; b &lt;c&gt;"},
		{"white space around", "  Pasta\n", "  Pasta\n"},
		{"only white space", " \n\t ", ""},
		{"too long", long, strings.Repeat("é", 1024)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.sent)
			assert.NoError(t, err)
			var got types.SanitizedString
			assert.NoError(t, json.Unmarshal(data, &got))
			assert.Equal(t, tt.want, string(got))
		})
	}
}
