package nightscoutmodels

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// ///
// Notes
// ///

type JSONString string

func (s *JSONString) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return fmt.Errorf("expected JSON string, got %s", data)
	}

	if len(str) > 2048 {
		str = str[:2048]
	}

	*s = JSONString(str)
	return nil
}

func (s *JSONString) String() string {
	return string(*s)
}

type LongJSONString string

func (s *LongJSONString) String() string {
	return string(*s)
}

func (s *LongJSONString) Stringp() *string {
	if s == nil {
		return nil
	}
	str := string(*s)
	return &str
}

func (n *LongJSONString) UnmarshalJSON(data []byte) error {
	// We are going to sanitize notes here, no html allowed, trim white space, convert to string
	var v string
	if err := json.Unmarshal(data, &v); err != nil {
		return fmt.Errorf("expected notes string, got %s", data)
	}

	// Test that this is valid json string
	var testJSON any
	err := json.Unmarshal([]byte(v), &testJSON)
	if err != nil {
		return fmt.Errorf("invalid JSON string, got %s", data)
	}

	// Concat to 1024 characters
	if len(v) > 4024 {
		v = v[:4024]
	}

	*n = LongJSONString(v)
	return nil
}

type IdentifierString string

func (i *IdentifierString) String() string {
	if i == nil {
		return ""
	}
	return string(*i)
}

func (i *IdentifierString) Stringp() *string {
	if i == nil {
		return nil
	}
	str := string(*i)
	return &str
}

func (i *IdentifierString) UnmarshalJSON(data []byte) error {
	// limit in possible characters and length, convert everything to a string
	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		return fmt.Errorf("expected identifier string, got %s", data)
	}

	str := ""

	switch v := v.(type) {
	case string:
		str = v
	case float64:
		str = strconv.FormatFloat(v, 'f', -1, 64)
	case json.Number:
		str = v.String()
	default:
		str = fmt.Sprintf("%v", v)
	}

	if len(str) > 32 {
		str = str[:32]
	}

	if len(str) == 0 {
		return fmt.Errorf("identifier string cannot be empty")
	}

	*i = IdentifierString(str)
	return nil
}

////
// Utils
////

func ptr[T any](v T) *T {
	return &v
}
