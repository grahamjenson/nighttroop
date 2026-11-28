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

// TableSizes are a DB's tables by name, as the JSON object they are kept as; empty is none measured yet. Every request
// of a site reads its site, and only the storage pages show its tables, so they are decoded where they are shown (Tables).
type TableSizes string

// NewTableSizes is the tables as they are kept
func NewTableSizes(tables map[string]TableSize) (TableSizes, error) {
	b, err := json.Marshal(tables)
	return TableSizes(b), err
}

// Tables are the tables by name, none when none is measured yet
func (t TableSizes) Tables() (map[string]TableSize, error) {
	tables := map[string]TableSize{}
	if t == "" {
		return tables, nil
	}
	return tables, json.Unmarshal([]byte(t), &tables)
}

func (t TableSizes) Value() (driver.Value, error) {
	if t == "" {
		return nil, nil
	}
	return string(t), nil
}

func (t *TableSizes) Scan(src any) error {
	switch src := src.(type) {
	case nil:
		*t = ""
	case string:
		*t = TableSizes(src)
	case []byte:
		*t = TableSizes(src)
	default:
		return fmt.Errorf("unsupported type for TableSizes: %T", src)
	}
	return nil
}
