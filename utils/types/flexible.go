package types

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// Flexible types accept JSON numbers, numeric strings and bool strings, as Nightscout clients send all of them

// //////
// FlexibleFloat64
// ///////

type FlexibleFloat64 float64

func (f *FlexibleFloat64) UnmarshalJSON(data []byte) error {
	var v any
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	if err := dec.Decode(&v); err != nil {
		return fmt.Errorf("expected float value, got %s", data)
	}

	switch val := v.(type) {
	case float64:
		*f = FlexibleFloat64(val)
	case json.Number:
		if asFloat, err := val.Float64(); err == nil {
			*f = FlexibleFloat64(asFloat)
			return nil
		}
		return fmt.Errorf("invalid float value %q", val)
	case string:
		var err error
		newF, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return fmt.Errorf("invalid float value %q", val)
		}
		*f = FlexibleFloat64(newF)
	default:
		return fmt.Errorf("expected number or string for float, got %v:%T", v, v)
	}
	return nil
}

func (f FlexibleFloat64) Float64() float64 {
	return float64(f)
}

func (f *FlexibleFloat64) Float64p() *float64 {
	if f == nil {
		return nil
	}
	v := float64(*f)
	return &v
}

// //////
// FlexibleInt64
// ///////

type FlexibleInt64 int64

func (i *FlexibleInt64) UnmarshalJSON(data []byte) error {
	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		return fmt.Errorf("expected int value, got %s", data)
	}

	switch val := v.(type) {
	case float64:
		*i = FlexibleInt64(int64(val))
	case json.Number:
		if asInt, err := val.Int64(); err == nil {
			*i = FlexibleInt64(asInt)
			return nil
		}
		asFloat, err := val.Float64()
		if err != nil {
			return fmt.Errorf("invalid int value %q: %w", val, err)
		}
		*i = FlexibleInt64(int64(asFloat))
	case string:
		var err error
		newF, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return fmt.Errorf("invalid int value %q: %w", val, err)
		}
		*i = FlexibleInt64(int64(newF))
	default:
		return fmt.Errorf("expected number or string for int, got %T", v)
	}

	return nil
}

func (i *FlexibleInt64) Int64p() *int64 {
	if i == nil {
		return nil
	}
	v := int64(*i)
	return &v
}

func (i *FlexibleInt64) Int64() int64 {
	if i == nil {
		return 0
	}
	return int64(*i)
}

func (i *FlexibleInt64) Int() int {
	if i == nil {
		return 0
	}
	return int(*i)
}

// //////
// FlexibleBool
// ///////

type FlexibleBool bool

func (b *FlexibleBool) UnmarshalJSON(data []byte) error {
	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		return fmt.Errorf("expected bool value, got %s", data)
	}

	switch val := v.(type) {
	case bool:
		*b = FlexibleBool(val)
	case string:
		lower := strings.ToLower(strings.TrimSpace(val))
		if lower == "true" || lower == "1" {
			*b = FlexibleBool(true)
		} else if lower == "false" || lower == "0" {
			*b = FlexibleBool(false)
		} else {
			return fmt.Errorf("invalid bool value %q", val)
		}
	default:
		return fmt.Errorf("expected bool or string for bool, got %T", v)
	}
	return nil
}

func (b *FlexibleBool) Boolp() *bool {
	if b == nil {
		return nil
	}
	v := bool(*b)
	return &v
}

func (b *FlexibleBool) Bool() bool {
	if b == nil {
		return false
	}
	return bool(*b)
}
