package nightscout

import (
	"encoding/json"
)

type Activity struct {
	Identifier *IdentifierString `json:"identifier"`

	Name *ShortString `json:"name,omitempty"` // Name of the activity

	ActivityType *ShortString     `json:"activityType,omitempty"` // Type of activity
	Duration     *FlexibleInt64   `json:"duration,omitempty"`     // Duration in minutes
	Steps        *FlexibleInt64   `json:"steps,omitempty"`        // Step count
	Calories     *FlexibleInt64   `json:"calories,omitempty"`     // Calories burned
	HeartRate    *FlexibleInt64   `json:"heartRate,omitempty"`    // Heart rate
	Notes        *SanitizedString `json:"notes,omitempty"`        // Notes

	// Date fields
	CreatedAt *DateString `json:"created_at,omitempty"` // ISO 8601 string

	// Default String Fields
	Device  *ShortString `json:"device,omitempty"`  // Device used for the activity
	App     *ShortString `json:"app,omitempty"`     // App used for the activity
	Subject *ShortString `json:"subject,omitempty"` // Subject of the activity

	// Non-JSON values to make mapping easier
	Date             *DateInt       `json:"-"` // Unix timestamp in milliseconds
	UtcOffsetMinutes *FlexibleInt64 `json:"-"`
	RawJsonData      string         `json:"-"`
}

func (t *Activity) UnmarshalJSON(data []byte) error {
	type Alias Activity
	aux := struct {
		ID *IdentifierString `json:"_id"`
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
		t.Identifier = aux.ID
	}

	// Set default CreatedAt (leave UTCOffset nil to be later replaces by user timezone)
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

func (t Activity) GetDevice() *string {
	return t.Device.Stringp()
}

func (t Activity) GetApp() *string {
	return t.App.Stringp()
}

func (t Activity) GetSubject() *string {
	return t.Subject.Stringp()
}

func (t Activity) GetDate() *int64 {
	return t.Date.Int64p()
}

func (t Activity) GetUtcOffsetMinutes() *int64 {
	return t.UtcOffsetMinutes.Int64p()
}

func (t Activity) GetIdentifier() *string {
	return t.Identifier.Stringp()
}
