package nightscout

type Setting struct {
	Identifier *string `json:"identifier,omitempty"`
	Type       *string `json:"type,omitempty"`

	Value *any `json:"value,omitempty"`

	SrvCreated  *DateInt `json:"srvCreated,omitempty"`
	SrvModified *DateInt `json:"srvModified,omitempty"`

	RawJSON string `json:"-"`
}

func (e *Setting) FromMap(m map[string]any) error {
	return FromMap(m, e)
}

func (e *Setting) ToMap() (map[string]any, error) {
	return ToMap(e)
}
