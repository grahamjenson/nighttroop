package nightscout

import (
	"encoding/json"
	"fmt"
	"strings"
)

////
// Each model needs to be able to
// 1. Unmarshal from JSON
// 2. Unmarshal from a Map
// 3. From DB Model
// 4. To Map
// 5. To DB Model
////
// Models
////

type Entry struct {
	Identifier *string    `json:"identifier,omitempty"`
	Type       *EntryType `json:"type,omitempty"`

	SGV       *FlexibleFloat64 `json:"sgv,omitempty"`
	MBG       *FlexibleFloat64 `json:"mbg,omitempty"`
	Direction *string          `json:"direction,omitempty"`
	Trend     *FlexibleFloat64 `json:"trend,omitempty"`
	Units     *Units           `json:"units,omitempty"`

	Device *string `json:"device,omitempty"`

	Date       *DateInt       `json:"date,omitempty"`
	DateString *DateString    `json:"dateString,omitempty"`
	Mills      *DateInt       `json:"mills,omitempty"`
	UTCOffset  *FlexibleInt64 `json:"utcOffset,omitempty"`

	Filtered   *FlexibleFloat64 `json:"filtered,omitempty"`
	Unfiltered *FlexibleFloat64 `json:"unfiltered,omitempty"`
	RSSI       *FlexibleFloat64 `json:"rssi,omitempty"`
	Noise      *FlexibleFloat64 `json:"noise,omitempty"`
	SysTime    *DateString      `json:"sysTime,omitempty"`

	SrvCreated  *DateInt `json:"srvCreated,omitempty"`
	SrvModified *DateInt `json:"srvModified,omitempty"`

	RawJSONData string `json:"-"`
}

func (t *Entry) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, t); err != nil {
		return err
	}
	t.RawJSONData = string(data)
	return nil
}

//////
// Subtypes and Constants
//////

type EntryType string

const (
	SGVType EntryType = "sgv"
	MBGType EntryType = "mbg"
)

func (t *EntryType) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	switch strings.ToLower(strings.TrimSpace(s)) {
	case "sgv":
		*t = SGVType
	case "mbg":
		*t = MBGType
	default:
		return fmt.Errorf("unknown entry type: %s", s)
	}

	return nil
}
