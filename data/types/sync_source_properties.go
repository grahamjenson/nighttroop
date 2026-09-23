package types

import (
	"encoding/json"
	"fmt"

	"nighttroop.t1d.nz/utils/consts"
)

type SyncSourceProperties map[string]any

func (s *SyncSourceProperties) Scan(value any) error {
	if value == nil {
		*s = nil
		return nil
	}

	b, ok := value.([]byte)
	if !ok {
		str, ok := value.(string)
		if !ok {
			return fmt.Errorf("SyncSourceProperties: expected string or []byte, got %T", value)
		}
		b = []byte(str)
	}

	return json.Unmarshal(b, s)
}

func (s *SyncSourceProperties) DexcomRegion() consts.DexcomRegion {
	if s == nil {
		return consts.DexcomRegionOther
	}
	regionRaw := (*s)["region"]
	switch regionRaw := regionRaw.(type) {
	case bool:
		if regionRaw {
			return consts.DexcomRegionUS
		}
		return consts.DexcomRegionOther
	case string:
		return consts.NewDexcomRegionFromString(regionRaw)
	default:
		return consts.DexcomRegionOther
	}
}

func (s *SyncSourceProperties) IsDexcomRegionUSA() bool {
	if s == nil {
		return false
	}
	regionRaw := (*s)["region"]
	switch regionRaw := regionRaw.(type) {
	case bool:
		return regionRaw
	case string:
		return regionRaw == "US"
	default:
		return false
	}

}
