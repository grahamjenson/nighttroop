package nightscout

import "encoding/json"

type Setting struct {
	Identifier *IdentifierString `json:"identifier,omitempty"`
	Type       *SanitizedString  `json:"type,omitempty"`

	Value *SanitizedString `json:"value,omitempty"`

	// Date Fields
	SrvCreated  *DateInt `json:"srvCreated,omitempty"`
	SrvModified *DateInt `json:"srvModified,omitempty"`
}

func (t *Setting) UnmarshalJSON(data []byte) error {
	type Alias Setting
	aux := struct {
		ID IdentifierString `json:"_id"`
		*Alias
	}{
		Alias: (*Alias)(t),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Identitfiers
	if t.Identifier == nil {
		t.Identifier = &aux.ID
	}

	return nil
}
