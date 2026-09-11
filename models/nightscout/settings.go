package nightscout

import (
	"encoding/json"
	"fmt"

	"nighttroop.t1d.nz/data/sql/control"
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
		Value json.RawMessage   `json:"value,omitempty"`
	}{
		Alias: (*Alias)(t),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if t.Identifier == nil {
		t.Identifier = aux.ID
	}

	// We are going to look up the type of the setting, and then process the Value into the correct safe/converting type.
	var value any
	typeOfSetting, ok := control.ALL_SETTINGS_TYPES[t.Identifier.String()]
	if !ok {
		value = new(ShortString)
		if err := json.Unmarshal(aux.Value, value); err != nil {
			if t.Identifier == nil {
				return fmt.Errorf("invalid value for setting: %w", err)
			}
			return fmt.Errorf("invalid value for setting %q: %w", *t.Identifier, err)
		}
	} else {
		var err error
		value, err = typeOfSetting.FromJSONRaw(aux.Value)
		if err != nil {
			if t.Identifier == nil {
				return fmt.Errorf("invalid value for setting: %w", err)
			}
			return fmt.Errorf("invalid value for setting %q: %w", *t.Identifier, err)
		}
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
