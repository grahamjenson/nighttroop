package types

import "database/sql/driver"

// ProfileSwitchOther are a profile switch's other attributes: the AAPS switch it was made from, which AAPS's followers
// read an effective profile switch (a Note) by, its times in milliseconds
type ProfileSwitchOther struct {
	OriginalProfileName    *string `json:"original_profile_name,omitempty"`
	OriginalCustomizedName *string `json:"original_customized_name,omitempty"`
	OriginalTimeshiftMs    *int64  `json:"original_timeshift_ms,omitempty"`
	OriginalPercentage     *int64  `json:"original_percentage,omitempty"`
	OriginalDurationMs     *int64  `json:"original_duration_ms,omitempty"`
	OriginalEnd            *int64  `json:"original_end,omitempty"`
}

func (o ProfileSwitchOther) Value() (driver.Value, error) { return otherValue(o) }

func (o *ProfileSwitchOther) Scan(src any) error {
	*o = ProfileSwitchOther{}
	return otherScan(src, o)
}
