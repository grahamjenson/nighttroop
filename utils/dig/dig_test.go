package dig_test

import (
	"encoding/json"
	"testing"

	"nighttroop.t1d.nz/utils/dig"
)

func TestDigBool(t *testing.T) {
	assertTrue := []any{"true", json.Number("1"), 1, true}
	for _, key := range assertTrue {
		got := dig.Bool(map[string]any{"a": key}, "a")
		if got == nil || *got != true {
			t.Errorf("Bool() = %v, want true for key %v", got, key)
		}
	}

	assertFalse := []any{"false", json.Number("0"), 0, false}
	for _, key := range assertFalse {
		got := dig.Bool(map[string]any{"a": key}, "a")
		if got == nil || *got != false {
			t.Errorf("Bool() = %v, want false for key %v", got, key)
		}
	}

	got := dig.Bool(map[string]any{"a": "notabool"}, "a")
	if got != nil {
		t.Errorf("Bool() = %v, want nil for invalid bool", got)
	}
}

func TestDigInt64(t *testing.T) {
	assertInt64 := []any{int64(42), float64(42), "42", json.Number("42"), json.Number("42.1")}
	for _, key := range assertInt64 {
		got := dig.Int64(map[string]any{"a": key}, "a")
		if got == nil || *got != 42 {
			t.Errorf("Int64() = %v, want 42 for key %v", got, key)
		}
	}

	got := dig.Int64(map[string]any{"a": "notanint"}, "a")
	if got != nil {
		t.Errorf("Int64() = %v, want nil for invalid int", got)
	}
}

func TestDigFloat64(t *testing.T) {
	assertFloat64 := []any{float64(42.5), "42.5", json.Number("42.5")}
	for _, key := range assertFloat64 {
		got := dig.Float64(map[string]any{"a": key}, "a")
		if got == nil || *got != 42.5 {
			t.Errorf("Float64() = %v, want 42.5 for key %v", got, key)
		}
	}

	got := dig.Float64(map[string]any{"a": "notafloat"}, "a")
	if got != nil {
		t.Errorf("Float64() = %v, want nil for invalid float", got)
	}
}

func TestDigString(t *testing.T) {
	assertString := []any{"hello", json.Number("42"), 42, true}

	for _, key := range assertString {
		got := dig.String(map[string]any{"a": key}, "a")
		if got == nil {
			t.Errorf("String() = nil, want non-nil for key %v", key)
		}
	}

	got := dig.String(map[string]any{"a": []int{1, 2, 3}}, "a")
	if got != nil {
		t.Errorf("String() = %v, want nil for invalid string", got)
	}
}

func TestDigValue(t *testing.T) {
	m := map[string]any{
		"level1": map[string]any{
			"level2": map[string]any{
				"value": 42,
			},
		},
	}

	got := dig.Value(m, "level1", "level2", "value")
	if got == nil || got.(int) != 42 {
		t.Errorf("Value() = %v, want 42", got)
	}

	got = dig.Value(m, "level1", "nonexistent", "value")
	if got != nil {
		t.Errorf("Value() = %v, want nil for nonexistent key", got)
	}
}
