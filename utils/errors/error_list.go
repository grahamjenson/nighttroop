package errors

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type ErrorList []string

// Stored errors are shown to people, not parsed, so they are bounded to keep every
// errors column under its CHECK whatever an upstream service put in the message.
// A list of MaxErrors messages of MaxErrorLength runes fits 51200 characters even
// when JSON escapes every rune.
const (
	MaxErrorLength = 512
	MaxErrors      = 16
)

// Truncate cuts an error message to MaxErrorLength runes, marking the cut with an ellipsis
func Truncate(msg string) string {
	r := []rune(msg)
	if len(r) <= MaxErrorLength {
		return msg
	}
	return string(r[:MaxErrorLength]) + "…"
}

func (e ErrorList) AddAll(errs ErrorList) ErrorList {
	for _, err := range errs {
		e = e.Add(err)
	}

	return e
}

// Add appends a distinct, truncated error; once the list is full later errors are dropped
func (e ErrorList) Add(err string) ErrorList {
	err = Truncate(err)
	for _, existing := range e {
		if existing == err {
			return e
		}
	}
	if len(e) >= MaxErrors {
		return e
	}

	e = append(e, err)
	return e
}

func (e ErrorList) Value() (driver.Value, error) {
	if len(e) == 0 {
		return "[]", nil
	}
	byteArray, err := json.Marshal(e)
	return string(byteArray), err
}

func (e *ErrorList) Scan(src interface{}) error {
	if src == nil {
		*e = ErrorList{}
		return nil
	}
	switch src := src.(type) {
	case string:
		return json.Unmarshal([]byte(src), e)
	case []byte:
		return json.Unmarshal(src, e)
	default:
		return fmt.Errorf("unsupported type: %T", src)
	}
}
