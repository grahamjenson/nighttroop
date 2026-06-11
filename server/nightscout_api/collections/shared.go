package collections

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/nsutils"
)

func CreateRecordSlice[T any](in []T, fn func(T) nsutils.Record) []nsutils.Record {
	out := make([]nsutils.Record, len(in))
	for i, v := range in {
		out[i] = fn(v)
	}
	return out
}

func StructToMap(in any) map[string]any {
	b, err := json.Marshal(in)
	if err != nil {
		slog.Error("Failed to marshal ns model to JSON", "error", err)
		return map[string]any{}
	}

	var m map[string]any
	err = json.Unmarshal(b, &m)
	if err != nil {
		slog.Error("Failed to unmarshal ns model from JSON", "error", err)
		return map[string]any{}
	}

	return m
}

func MapToStruct(m map[string]any, out any) (string, error) {
	b, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	return string(b), json.Unmarshal(b, out)
}

func dateString(t int64) string {
	return time.UnixMilli(t).UTC().Format(time.RFC3339Nano)
}

type FlexibleFloat64 float64

func (f *FlexibleFloat64) UnmarshalJSON(data []byte) error {
	// number
	var num float64
	if err := json.Unmarshal(data, &num); err == nil {
		*f = FlexibleFloat64(num)
		return nil
	}

	// string
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		v, err := strconv.ParseFloat(str, 64)
		if err != nil {
			return fmt.Errorf("invalid float value %q: %w", str, err)
		}
		*f = FlexibleFloat64(v)
		return nil
	}

	return fmt.Errorf("expected number or string, got %s", data)
}

func (f FlexibleFloat64) Float64() float64 {
	return float64(f)
}

func ptr[T any](v T) *T {
	return &v
}
