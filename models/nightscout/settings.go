package nightscout

import "encoding/json"

type Setting struct {
	Identifier *string `json:"identifier,omitempty"`
	Type       *string `json:"type,omitempty"`

	Value *any `json:"value,omitempty"`

	// Date Fields
	SrvCreated  *DateInt `json:"srvCreated,omitempty"`
	SrvModified *DateInt `json:"srvModified,omitempty"`

	// Non-JSON values to make mapping easier
	RawJsonData string `json:"-"`
}

func (t *Setting) UnmarshalJSON(data []byte) error {
	type Alias Setting
	aux := struct {
		ID string `json:"_id"`
		*Alias
	}{
		Alias: (*Alias)(t),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Raw Value
	t.RawJsonData = string(data)

	// Identitfiers
	if t.Identifier == nil {
		t.Identifier = &aux.ID
	}

	return nil
}
