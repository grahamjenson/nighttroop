package nightscout

import "encoding/json"

type Setting struct {
	Identifier *string `json:"identifier,omitempty"`
	Type       *string `json:"type,omitempty"`

	Value *any `json:"value,omitempty"`

	SrvCreated  *DateInt `json:"srvCreated,omitempty"`
	SrvModified *DateInt `json:"srvModified,omitempty"`

	RawJSONData string `json:"-"`
}

func (t *Setting) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, t); err != nil {
		return err
	}
	t.RawJSONData = string(data)
	return nil
}
