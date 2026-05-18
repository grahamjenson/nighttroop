package errors

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type ErrorList []string

func (e ErrorList) Value() (driver.Value, error) {
	if len(e) == 0 {
		return "[]", nil
	}
	return json.Marshal(e)
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
