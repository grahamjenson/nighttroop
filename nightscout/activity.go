package nightscout

type Activity struct {
	Identifier *string `json:"identifier"`

	CreatedAt    *DateString      `json:"created_at"`   // ISO 8601 string
	ActivityType *string          `json:"activityType"` // Type of activity
	Duration     *FlexibleFloat64 `json:"duration"`     // Duration in minutes
	Steps        *FlexibleFloat64 `json:"steps"`        // Step count
	HeartRate    *FlexibleFloat64 `json:"heartRate"`    // Heart rate
	Notes        *string          `json:"notes"`        // Notes

	// API v3 additions

	SrvCreated  *DateInt `json:"srvCreated,omitempty"`
	SrvModified *DateInt `json:"srvModified,omitempty"`

	RawJSON string `json:"-"`
}

func (e *Activity) FromMap(m map[string]any) error {
	return FromMap(m, e)
}

func (e *Activity) ToMap() (map[string]any, error) {
	return ToMap(e)
}
