package types

import (
	"bytes"
	"encoding/json"
	"regexp"
)

// wholeNumber is a JSON number with no fraction or exponent
var wholeNumber = regexp.MustCompile(`^-?[0-9]+$`)

// ClientID is an id a client gives its own record and finds it by, such as AAPS's pumpId or xDrip+'s uuid, so it is
// given back as sent, as a string: a string as it is, and a whole number with all its figures, past those a float
// keeps, as AAPS sends its pumpId as a Long and reads a quoted number back into it. Anything else is read as a
// ShortString is. It is not cut: one longer than its column is left out when the document is checked against the
// column limits.
type ClientID string

func (c *ClientID) UnmarshalJSON(data []byte) error {
	var short ShortString
	if err := json.Unmarshal(data, &short); err != nil {
		return err
	}
	if figures := bytes.TrimSpace(data); wholeNumber.Match(figures) {
		short = ShortString(figures)
	}
	*c = ClientID(short)
	return nil
}
