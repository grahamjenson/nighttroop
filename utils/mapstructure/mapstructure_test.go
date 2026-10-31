package mapstructure

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/utils/types"
)

type query struct {
	Limit  *types.FlexibleInt64 `mapstructure:"limit,omitempty"`
	Device *string              `mapstructure:"device,omitempty"`
	Code   *types.ShortString   `mapstructure:"code"`
	Other  map[string]any       `mapstructure:",remain"`
}

func TestDecode(t *testing.T) {
	ptr := func(s string) *string { return &s }
	tests := []struct {
		name   string
		values map[string]string
		match  func(key, name string) bool
		want   query
		err    string
	}{
		{name: "no values sets nothing", values: map[string]string{}, want: query{}},
		{name: "a string field is the value", values: map[string]string{"device": "xDrip"}, want: query{Device: ptr("xDrip")}},
		{name: "a type with UnmarshalJSON is given the value as a JSON string", values: map[string]string{"limit": "10", "code": "abc"},
			want: query{Limit: new(types.FlexibleInt64(10)), Code: new(types.ShortString("abc"))}},
		{name: "keys match tag names case-insensitively", values: map[string]string{"LIMIT": "5"}, want: query{Limit: new(types.FlexibleInt64(5))}},
		{name: "unmatched values go to the remain field", values: map[string]string{"device": "x", "count": "3"},
			want: query{Device: ptr("x"), Other: map[string]any{"count": "3"}}},
		{name: "a custom match maps other keys to a field",
			values: map[string]string{"count": "7"}, match: func(key, name string) bool { return strings.EqualFold(map[string]string{"count": "limit"}[key], name) },
			want: query{Limit: new(types.FlexibleInt64(7))}},
		{name: "an exact key is used before a custom match", values: map[string]string{"limit": "1", "count": "2"},
			match: func(key, name string) bool { return key == "count" && name == "limit" }, want: query{Limit: new(types.FlexibleInt64(1)), Other: map[string]any{"count": "2"}}},
		{name: "a value its type cannot unmarshal is an error naming the field", values: map[string]string{"limit": "ten"}, err: "'limit'"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got query
			err := Decode(tt.values, &got, tt.match)
			if tt.err != "" {
				require.ErrorContains(t, err, tt.err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

// A field must be a pointer to a string or to a type with its own UnmarshalJSON
func TestDecodeRefusesOtherFields(t *testing.T) {
	var notPointer struct {
		Count int64 `mapstructure:"count"`
	}
	require.ErrorContains(t, Decode(map[string]string{"count": "1"}, &notPointer, nil), "must be a pointer")
	var noUnmarshal struct {
		Count *int64 `mapstructure:"count"`
	}
	require.ErrorContains(t, Decode(map[string]string{"count": "1"}, &noUnmarshal, nil), "cannot be decoded")
}
