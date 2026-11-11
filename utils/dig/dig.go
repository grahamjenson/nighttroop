package dig

import (
	"encoding/json"
	"fmt"
	"strconv"
)

func Bool(m map[string]any, keys ...string) *bool {
	v := Value(m, keys...)
	if v == nil {
		return nil
	}

	switch v := v.(type) {
	case bool:
		return &v
	}

	value, err := strconv.ParseBool(fmt.Sprint(v))
	if err != nil {
		return nil
	}
	return &value

}

func String(m map[string]any, keys ...string) *string {
	v := Value(m, keys...)
	if v == nil {
		return nil
	}

	switch v := v.(type) {
	case string:
		return &v
	case json.Number:
		value := string(v)
		return &value
	case bool:
		value := strconv.FormatBool(v)
		return &value
	case float64, float32, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		value := fmt.Sprint(v)
		return &value
	}
	return nil
}

func Float64(m map[string]any, keys ...string) *float64 {
	v := Value(m, keys...)
	if v == nil {
		return nil
	}

	switch v := v.(type) {
	case float64:
		return &v

	case float32:
		value := float64(v)
		return &value

	case string:
		value, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return nil
		}
		return &value

	case bool:
		if v {
			value := float64(1)
			return &value
		}
		value := float64(0)
		return &value

	case int:
		value := float64(v)
		return &value

	case int8:
		value := float64(v)
		return &value

	case int16:
		value := float64(v)
		return &value

	case int32:
		value := float64(v)
		return &value

	case int64:
		value := float64(v)
		return &value

	case uint:
		value := float64(v)
		return &value

	case uint8:
		value := float64(v)
		return &value

	case uint16:
		value := float64(v)
		return &value

	case uint32:
		value := float64(v)
		return &value

	case uint64:
		value := float64(v)
		return &value

	case json.Number:
		value, err := v.Float64()
		if err != nil {
			return nil
		}
		return &value
	}

	return nil
}

func Int64(m map[string]any, keys ...string) *int64 {
	v := Value(m, keys...)
	if v == nil {
		return nil
	}

	switch v := v.(type) {
	case int:
		value := int64(v)
		return &value

	case int8:
		value := int64(v)
		return &value

	case int16:
		value := int64(v)
		return &value

	case int32:
		value := int64(v)
		return &value

	case int64:
		return &v

	case uint:
		value := int64(v)
		return &value

	case uint8:
		value := int64(v)
		return &value

	case uint16:
		value := int64(v)
		return &value

	case uint32:
		value := int64(v)
		return &value

	case uint64:
		value := int64(v)
		return &value

	case float32:
		value := int64(v)
		return &value

	case float64:
		value := int64(v)
		return &value

	case string:
		value, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			// Also allow "123.5"
			floatValue, floatErr := strconv.ParseFloat(v, 64)
			if floatErr != nil {
				return nil
			}
			value = int64(floatValue)
		}
		return &value

	case bool:
		if v {
			value := int64(1)
			return &value
		}
		value := int64(0)
		return &value

	case json.Number:
		value, err := v.Int64()
		if err == nil {
			return &value
		}

		floatValue, err := v.Float64()
		if err != nil {
			return nil
		}

		value = int64(floatValue)
		return &value
	}

	return nil
}

func Value(m map[string]any, keys ...string) any {
	var current any = m

	for _, key := range keys {
		obj, ok := current.(map[string]any)
		if !ok {
			return nil
		}

		current, ok = obj[key]
		if !ok || current == nil {
			return nil
		}
	}

	return current
}
