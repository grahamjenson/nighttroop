package types

import (
	"encoding/json"
	"fmt"
)

type SyncSourceProperties map[string]any

func (j *SyncSourceProperties) Scan(value any) error {
	if value == nil {
		*j = nil
		return nil
	}

	b, ok := value.([]byte)
	if !ok {
		s, ok := value.(string)
		if !ok {
			return fmt.Errorf("SyncSourceProperties: expected string or []byte, got %T", value)
		}
		b = []byte(s)
	}

	return json.Unmarshal(b, j)
}
