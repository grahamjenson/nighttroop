package nightscoutmodels

import (
	"encoding/json"
	"fmt"
	"nighttroop.t1d.nz/utils/types"
)

type Setting struct {
	Identifier *types.IdentifierString `json:"identifier,omitempty"`
	Type       *types.ShortString      `json:"type,omitempty"`
	Value      any                     `json:"value,omitempty"`
}

func (t *Setting) UnmarshalJSON(data []byte) error {
	type Alias Setting

	aux := struct {
		*Alias
		ID    *types.IdentifierString `json:"_id"`
		Value any                     `json:"value,omitempty"`
	}{
		Alias: (*Alias)(t),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if t.Identifier == nil {
		t.Identifier = aux.ID
	}

	// Just making sure it is one of the correct types
	value := aux.Value
	switch v := aux.Value.(type) {
	case string:
		if len(v) > 255 {
			return fmt.Errorf("string value exceeds 255 characters")
		}

	case int64, float64, bool:
	// A list, as of device names, which its setting checks; this only keeps it small
	case []any:
		if len(v) > 20 {
			return fmt.Errorf("list value exceeds 20 items")
		}
	default:
		return fmt.Errorf("unsupported value type")

	}

	t.Value = value
	return nil
}

func (t Setting) GetDevice() *string {
	return new("")
}

func (t Setting) GetApp() *string {
	return new("")
}

func (t Setting) GetSubject() *string {
	return new("")
}

func (t Setting) GetDate() *int64 {
	return nil
}

func (t Setting) GetUtcOffsetMinutes() *int64 {
	return new(int64(0))
}

func (t Setting) GetIdentifier() *string {
	if t.Identifier == nil {
		return new("")
	}
	return t.Identifier.Stringp()
}
