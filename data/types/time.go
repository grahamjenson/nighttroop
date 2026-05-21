package types

import (
	"database/sql/driver"
	"fmt"
	"time"
)

type UnixMillisTime struct {
	time.Time
}

func (t *UnixMillisTime) Scan(value any) error {
	if value == nil {
		t.Time = time.Time{}
		return nil
	}

	switch v := value.(type) {
	case int64:
		t.Time = time.UnixMilli(v).UTC()
		return nil

	case int:
		t.Time = time.UnixMilli(int64(v)).UTC()
		return nil

	case []byte:
		var ms int64
		_, err := fmt.Sscan(string(v), &ms)
		if err != nil {
			return err
		}
		t.Time = time.UnixMilli(ms).UTC()
		return nil

	default:
		return fmt.Errorf("unsupported type for UnixMillisTime: %T", value)
	}
}

func (t UnixMillisTime) Value() (driver.Value, error) {
	if t.Time.IsZero() {
		return nil, nil
	}

	return t.Time.UnixMilli(), nil
}
