package nightscout

import (
	"encoding/json"
	"fmt"
	"strings"
)

////
// Models
////

type Treatment struct {
	Identifier *string `json:"identifier,omitempty"`

	EventType *TreatmentEventType `json:"eventType,omitempty"`

	Glucose     *FlexibleFloat64 `json:"glucose"`
	GlucoseType *GlucoseType     `json:"glucoseType"` // "Sensor" or "Finger"

	Carbs   *FlexibleFloat64 `json:"carbs"`
	Insulin *FlexibleFloat64 `json:"insulin"`

	Duration *FlexibleInt64 `json:"duration"` // minutes

	Notes     *string `json:"notes"`
	EnteredBy *string `json:"enteredBy"`

	// Treatment-specific fields
	Profile      *string          `json:"profile"`
	Percentage   *FlexibleFloat64 `json:"percentage"`
	Absolute     *FlexibleFloat64 `json:"absolute"`
	Reason       *string          `json:"reason"`
	TargetTop    *FlexibleFloat64 `json:"targetTop"`
	TargetBottom *FlexibleFloat64 `json:"targetBottom"`
	Units        *Units           `json:"units,omitempty"`

	// Date Fields
	CreatedAt        *DateString    `json:"created_at"`     // ISO8601 -> time.Time
	Date             *DateInt       `json:"date,omitempty"` // Unix timestamp in milliseconds, only used for some treatment types
	UtcOffsetMinutes *FlexibleInt64 `json:"utcOffset,omitempty"`
	SrvCreated       *DateInt       `json:"srvCreated"`
	SrvModified      *DateInt       `json:"srvModified"`

	// Non-JSON values to make mapping easier
	RawJsonData string `json:"-"`
}

func (t *Treatment) UnmarshalJSON(data []byte) error {
	type Alias Treatment
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

//"eventType": "Correction Bolus",
// "eventType": "Temporary Target",
// "eventType": "Meal Bolus",
// "eventType": "Temp Basal",
// "eventType": "Note",
// "eventType": "Profile Switch",
// "eventType": "Carb Correction",

//////
// Subtypes and Constants
//////

type TreatmentEventType string

const (
	CorrectionBolus TreatmentEventType = "Correction Bolus"
	ComboBolus      TreatmentEventType = "Combo Bolus"
	TemporaryTarget TreatmentEventType = "Temporary Target"
	MealBolus       TreatmentEventType = "Meal Bolus"
	TempBasal       TreatmentEventType = "Temp Basal"
	Note            TreatmentEventType = "Note"
	ProfileSwitch   TreatmentEventType = "Profile Switch"
	CarbCorrection  TreatmentEventType = "Carb Correction"
	Unknown         TreatmentEventType = "Unknown"
)

var TreatmentEventTypes = []TreatmentEventType{CorrectionBolus, ComboBolus, TemporaryTarget, MealBolus, TempBasal, Note, ProfileSwitch, CarbCorrection}

func (t *TreatmentEventType) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	switch strings.ToLower(strings.TrimSpace(s)) {
	case "correction bolus":
		*t = CorrectionBolus
	case "temporary target":
		*t = TemporaryTarget
	case "meal bolus":
		*t = MealBolus
	case "temp basal":
		*t = TempBasal
	case "note":
		*t = Note
	case "profile switch":
		*t = ProfileSwitch
	case "carb correction":
		*t = CarbCorrection
	default:
		return fmt.Errorf("unknown treatment event type: %s", s)
	}

	return nil
}

type GlucoseType string

const (
	Sensor GlucoseType = "Sensor"
	Finger GlucoseType = "Finger"
)

func (t *GlucoseType) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	switch strings.ToLower(strings.TrimSpace(s)) {
	case "sensor":
		*t = Sensor
	case "finger":
		*t = Finger
	default:
		return fmt.Errorf("unknown glucose type: %s", s)
	}

	return nil
}
