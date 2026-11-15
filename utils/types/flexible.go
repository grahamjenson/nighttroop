package types

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strconv"
	"strings"
)

// Flexible types accept JSON numbers, numeric strings and bool strings, as Nightscout clients send all of them

// typeError is a value sent that is not of the type kept, as the JSON decoder's own, so it says the field's path
func typeError(data []byte, kept reflect.Type) error {
	return &json.UnmarshalTypeError{Value: string(data[:min(len(data), 64)]), Type: kept}
}

// //////
// FlexibleFloat64
// ///////

type FlexibleFloat64 float64

func (f *FlexibleFloat64) UnmarshalJSON(data []byte) error {
	var v any
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	if err := dec.Decode(&v); err != nil {
		return typeError(data, reflect.TypeFor[float64]())
	}

	switch val := v.(type) {
	case float64:
		*f = FlexibleFloat64(val)
	case json.Number:
		if asFloat, err := val.Float64(); err == nil {
			*f = FlexibleFloat64(asFloat)
			return nil
		}
		return typeError(data, reflect.TypeFor[float64]())
	case string:
		var err error
		newF, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return typeError(data, reflect.TypeFor[float64]())
		}
		*f = FlexibleFloat64(newF)
	default:
		return typeError(data, reflect.TypeFor[float64]())
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
		return typeError(data, reflect.TypeFor[int64]())
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
			return typeError(data, reflect.TypeFor[int64]())
		}
		*i = FlexibleInt64(int64(asFloat))
	case string:
		var err error
		newF, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return typeError(data, reflect.TypeFor[int64]())
		}
		*i = FlexibleInt64(int64(newF))
	default:
		return typeError(data, reflect.TypeFor[int64]())
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
		return typeError(data, reflect.TypeFor[bool]())
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
			return typeError(data, reflect.TypeFor[bool]())
		}
	default:
		return typeError(data, reflect.TypeFor[bool]())
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
