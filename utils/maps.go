package utils

import (
	"encoding/json"
	"log/slog"
)

/////
// MAPs
/////

func ConvertViaJSON[F any, T any](src F, dst *T) error {
	b, err := json.Marshal(src)
	if err != nil {
		return err
	}

	return json.Unmarshal(b, dst)
}

func FromMap[T any](m map[string]any, dst *T) error {
	b, err := json.Marshal(m)
	if err != nil {
		return err
	}

	return json.Unmarshal(b, dst)
}

func ToMap[T any](src *T) (map[string]any, error) {
	b, err := json.Marshal(src)
	if err != nil {
		return nil, err
	}
	var m map[string]any
	err = json.Unmarshal(b, &m)
	return m, err
}

func DeepCloneMap(m map[string]any) map[string]any {
	if m == nil {
		return map[string]any{}
	}
	b, err := json.Marshal(m)
	if err != nil {
		slog.Warn("failed to marshal map for deep clone", "error", err)
		return map[string]any{}
	}
	var clone map[string]any
	err = json.Unmarshal(b, &clone)
	if err != nil {
		slog.Warn("failed to unmarshal map for deep clone", "error", err)
		return map[string]any{}
	}
	return clone
}
