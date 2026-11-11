package dig_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"nighttroop.t1d.nz/utils/dig"
)

// value is what a getter found, or nil, so a failure shows the value and not its pointer
func value[T any](found *T) any {
	if found == nil {
		return nil
	}
	return *found
}

// TestDigTypes is each kind of value a decoded document may hold, and what each getter reads it as, nil for not at all
func TestDigTypes(t *testing.T) {
	tests := []struct {
		value   any
		boolean any
		int64   any
		float64 any
		string  any
	}{
		{true, true, int64(1), 1.0, "true"},
		{false, false, int64(0), 0.0, "false"},
		{"true", true, nil, nil, "true"},
		{"false", false, nil, nil, "false"},
		{"hello", nil, nil, nil, "hello"},
		{1, true, int64(1), 1.0, "1"},
		{0, false, int64(0), 0.0, "0"},
		{42, nil, int64(42), 42.0, "42"},
		{int64(42), nil, int64(42), 42.0, "42"},
		{float64(42), nil, int64(42), 42.0, "42"},
		{float64(42.5), nil, int64(42), 42.5, "42.5"},
		{"42", nil, int64(42), 42.0, "42"},
		{"42.5", nil, int64(42), 42.5, "42.5"},
		{json.Number("1"), true, int64(1), 1.0, "1"},
		{json.Number("0"), false, int64(0), 0.0, "0"},
		{json.Number("42"), nil, int64(42), 42.0, "42"},
		{json.Number("42.1"), nil, int64(42), 42.1, "42.1"},
		{json.Number("42.5"), nil, int64(42), 42.5, "42.5"},
		{[]int{1, 2, 3}, nil, nil, nil, nil},
		{nil, nil, nil, nil, nil},
	}
	for _, test := range tests {
		t.Run(fmt.Sprintf("%T %v", test.value, test.value), func(t *testing.T) {
			m := map[string]any{"a": test.value}
			assert.Equal(t, test.boolean, value(dig.Bool(m, "a")), "Bool")
			assert.Equal(t, test.int64, value(dig.Int64(m, "a")), "Int64")
			assert.Equal(t, test.float64, value(dig.Float64(m, "a")), "Float64")
			assert.Equal(t, test.string, value(dig.String(m, "a")), "String")
		})
	}
}

func TestDigValue(t *testing.T) {
	m := map[string]any{"level1": map[string]any{"level2": map[string]any{"value": 42}}}
	tests := []struct {
		keys  []string
		value any
	}{
		{[]string{"level1", "level2", "value"}, 42},
		{[]string{"level1", "nonexistent", "value"}, nil},
		{[]string{"level1", "level2", "value", "deeper"}, nil},
		{[]string{"missing"}, nil},
	}
	for _, test := range tests {
		assert.Equal(t, test.value, dig.Value(m, test.keys...), test.keys)
	}
}
