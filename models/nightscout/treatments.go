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

	Notes     *SanitizedString `json:"notes"`
	EnteredBy *SanitizedString `json:"enteredBy"`

	// Treatment-specific fields
	Profile      *SanitizedString `json:"profile"`
	Percentage   *FlexibleFloat64 `json:"percentage"`
	Rate         *FlexibleFloat64 `json:"rate"`
	Absolute     *FlexibleFloat64 `json:"absolute"`
	Reason       *SanitizedString `json:"reason"`
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

	// Set default CreatedAt (leave UTCOffset nil to be later replaces by user timezone)
	if t.CreatedAt == nil {
		// This is UTC not User timezone, will have to set UTC offset minutes at API level
		ds, now := DateStringNow()
		t.CreatedAt = ds
		t.Date = ptr(DateIntFromTime(now))
	} else {
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

func (t TreatmentEventType) ToLocalTypeString() string {
	switch t {
	case CorrectionBolus:
		return "correction_bolus"
	case ComboBolus:
		return "combo_bolus"
	case TemporaryTarget:
		return "temporary_target"
	case MealBolus:
		return "meal_bolus"
	case TempBasal:
		return "temp_basal"
	case Note:
		return "note"
	case ProfileSwitch:
		return "profile_switch"
	case CarbCorrection:
		return "carb_correction"
	default:
		return "unknown"
	}
}

func FromLocalTypeString(s string) *TreatmentEventType {
	cb := CorrectionBolus
	cob := ComboBolus
	tt := TemporaryTarget
	mb := MealBolus
	tb := TempBasal
	n := Note
	ps := ProfileSwitch
	cc := CarbCorrection

	switch strings.ToLower(strings.TrimSpace(s)) {
	case "correction_bolus":
		return &cb
	case "combo_bolus":
		return &cob
	case "temporary_target":
		return &tt
	case "meal_bolus":
		return &mb
	case "temp_basal":
		return &tb
	case "note":
		return &n
	case "profile_switch":
		return &ps
	case "carb_correction":
		return &cc
	}

	return nil
}

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
