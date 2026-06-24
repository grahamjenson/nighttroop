package nightscout

import "encoding/json"

type Activity struct {
	Identifier *string `json:"identifier"`

	Name *string `json:"name,omitempty"` // Name of the activity

	ActivityType *string        `json:"activityType"` // Type of activity
	Duration     *FlexibleInt64 `json:"duration"`     // Duration in minutes
	Steps        *FlexibleInt64 `json:"steps"`        // Step count
	Calories     *FlexibleInt64 `json:"calories"`     // Calories burned
	HeartRate    *FlexibleInt64 `json:"heartRate"`    // Heart rate
	Notes        *string        `json:"notes"`        // Notes

	// Date fields
	CreatedAt   *DateString `json:"created_at"` // ISO 8601 string
	SrvCreated  *DateInt    `json:"srvCreated,omitempty"`
	SrvModified *DateInt    `json:"srvModified,omitempty"`

	// Non-JSON values to make mapping easier
	Date             *DateInt       `json:"-"` // Unix timestamp in milliseconds
	UtcOffsetMinutes *FlexibleInt64 `json:"-"`
	RawJsonData      string         `json:"-"`
}

func (t *Activity) UnmarshalJSON(data []byte) error {
	type Alias Activity
	aux := struct {
		ID string `json:"_id"`
		*Alias
	}{
		Alias: (*Alias)(t),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Raw Value
	t.RawJsonData = string(data)

	// Identitfiers
	if t.Identifier == nil {
		t.Identifier = &aux.ID
	}

	// Cleaning Updates
	if t.CreatedAt != nil {
		t.Date, t.UtcOffsetMinutes = t.CreatedAt.MillisUTCOffset()
	}

	return nil
}

func (t *Activity) MarshalJSON() ([]byte, error) {
	type Alias Activity
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(t),
	})
}
