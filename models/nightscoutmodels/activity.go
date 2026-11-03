package nightscoutmodels

import (
	"encoding/json"
	"nighttroop.t1d.nz/utils/types"
)

type Activity struct {
	Identifier *types.IdentifierString `json:"identifier"`

	Name *types.ShortString `json:"name,omitempty"` // Name of the activity

	ActivityType    *types.ShortString     `json:"activityType,omitempty"` // Type of activity
	DurationMinutes *types.FlexibleInt64   `json:"duration,omitempty"`     // Duration in minutes
	Steps           *types.FlexibleInt64   `json:"steps,omitempty"`        // Step count
	Calories        *types.FlexibleInt64   `json:"calories,omitempty"`     // Calories burned
	HeartRate       *types.FlexibleInt64   `json:"heartRate,omitempty"`    // Heart rate
	Notes           *types.SanitizedString `json:"notes,omitempty"`        // Notes
	Reason          *types.SanitizedString `json:"reason,omitempty"`       // Reason

	// Date fields
	CreatedAt *types.DateString `json:"created_at,omitempty"` // ISO 8601 string

	// Default String Fields
	Device  *types.ShortString `json:"device,omitempty"`  // Device used for the activity
	App     *types.ShortString `json:"app,omitempty"`     // App used for the activity
	Subject *types.ShortString `json:"subject,omitempty"` // Subject of the activity

	// Non-JSON values to make mapping easier
	Date             *types.DateInt       `json:"date,omitempty"`      // Unix timestamp in milliseconds
	UtcOffsetMinutes *types.FlexibleInt64 `json:"utcOffset,omitempty"` // Minutes offset from UTC
	RawJsonData      string               `json:"-"`
}

func (t *Activity) UnmarshalJSON(data []byte) error {
	type Alias Activity
	aux := struct {
		ID *types.IdentifierString `json:"_id"`
		// xDrip+ sends its heart rate as {"type": "hr-bpm", "bpm"}, steps as steps-total and motion as
		// {"type": "motion-class", "class"}
		Type  *types.ShortString   `json:"type"`
		BPM   *types.FlexibleInt64 `json:"bpm"`
		Class *types.ShortString   `json:"class"`
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
	if t.ActivityType == nil {
		t.ActivityType = aux.Type
	}
	if t.HeartRate == nil {
		t.HeartRate = aux.BPM
	}
	if t.Name == nil {
		t.Name = aux.Class
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
