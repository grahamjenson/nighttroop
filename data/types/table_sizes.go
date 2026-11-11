package types

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// TableSize is the bytes a table of a DB takes, its indexes counted with it, and its records
type TableSize struct {
	Bytes   int64 `json:"bytes"`
	Records int64 `json:"records"`
}

// TableSizes are a DB's tables by name, kept as a JSON object; nil is none measured yet
type TableSizes map[string]TableSize

func (t TableSizes) Value() (driver.Value, error) {
	if t == nil {
		return nil, nil
	}
	b, err := json.Marshal(map[string]TableSize(t))
	return string(b), err
}

func (t *TableSizes) Scan(src any) error {
	switch src := src.(type) {
	case nil:
		*t = nil
		return nil
	case string:
		return json.Unmarshal([]byte(src), t)
	case []byte:
		return json.Unmarshal(src, t)
	default:
		return fmt.Errorf("unsupported type for TableSizes: %T", src)
	}
}
