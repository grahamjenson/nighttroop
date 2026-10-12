package nightscoutmodels

import (
	"encoding/json"
	"fmt"
)

type Setting struct {
	Identifier *IdentifierString `json:"identifier,omitempty"`
	Type       *ShortString      `json:"type,omitempty"`
	Value      any               `json:"value,omitempty"`
}

func (t *Setting) UnmarshalJSON(data []byte) error {
	type Alias Setting

	aux := struct {
		*Alias
		ID    *IdentifierString `json:"_id"`
		Value any               `json:"value,omitempty"`
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
	default:
		return fmt.Errorf("unsupported value type")

	}

	t.Value = value
	return nil
}

func (t Setting) GetDevice() *string {
	return ptr("")
}

func (t Setting) GetApp() *string {
	return ptr("")
}

func (t Setting) GetSubject() *string {
	return ptr("")
}

func (t Setting) GetDate() *int64 {
	return nil
}

func (t Setting) GetUtcOffsetMinutes() *int64 {
	return ptr(int64(0))
}

func (t Setting) GetIdentifier() *string {
	if t.Identifier == nil {
		return ptr("")
	}
	return t.Identifier.Stringp()
}
