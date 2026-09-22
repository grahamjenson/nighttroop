package errors

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type ErrorList []string

func (e ErrorList) AddAll(errs ErrorList) ErrorList {
	seen := make(map[string]struct{}, len(e)+len(errs))

	for _, err := range e {
		seen[err] = struct{}{}
	}

	for _, err := range errs {
		if _, exists := seen[err]; exists {
			continue
		}

		seen[err] = struct{}{}
		e = append(e, err)
	}

	return e
}

func (e ErrorList) Add(err string) ErrorList {
	for _, existing := range e {
		if existing == err {
			return e
		}
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
	switch src := src.(type) {
	case string:
		return json.Unmarshal([]byte(src), e)
	case []byte:
		return json.Unmarshal(src, e)
	default:
		return fmt.Errorf("unsupported type: %T", src)
	}
}
