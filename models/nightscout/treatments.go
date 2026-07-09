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
	Identifier *IdentifierString `json:"identifier,omitempty"`

	EventType *TreatmentEventType `json:"eventType,omitempty"`

	Glucose     *FlexibleFloat64 `json:"glucose"`
	GlucoseType *GlucoseType     `json:"glucoseType"` // "Sensor" or "Finger"

	Carbs    *FlexibleFloat64 `json:"carbs"`
	Insulin  *FlexibleFloat64 `json:"insulin"`
	Protein  *FlexibleFloat64 `json:"protein"`
	Fat      *FlexibleFloat64 `json:"fat"`
	PreBolus *FlexibleFloat64 `json:"preBolus"`

	SplitNow *FlexibleFloat64 `json:"splitNow"`
	SplitExt *FlexibleFloat64 `json:"splitExt"`

	Duration *FlexibleInt64 `json:"duration"` // minutes

	Notes     *SanitizedString `json:"notes"`
	EnteredBy *ShortString     `json:"enteredBy"`

	// Treatment-specific fields
	Profile      *ShortString     `json:"profile"`
	Percentage   *FlexibleFloat64 `json:"percentage"`
	Rate         *FlexibleFloat64 `json:"rate"`
	Absolute     *FlexibleFloat64 `json:"absolute"`
	Reason       *ShortString     `json:"reason"`
	TargetTop    *FlexibleFloat64 `json:"targetTop"`
	TargetBottom *FlexibleFloat64 `json:"targetBottom"`
	Units        *Units           `json:"units,omitempty"`

	// Default String Fields
	Device  *ShortString `json:"device,omitempty"`  // Device used for the activity
	App     *ShortString `json:"app,omitempty"`     // App used for the activity
	Subject *ShortString `json:"subject,omitempty"` // Subject of the activity

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
		ID IdentifierString `json:"_id"`
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
	Announcement      TreatmentEventType = "Announcement"
	BGCheck           TreatmentEventType = "BG Check"
	BolusWizard       TreatmentEventType = "Bolus Wizard"
	CarbCorrection    TreatmentEventType = "Carb Correction"
	ComboBolus        TreatmentEventType = "Combo Bolus"
	CorrectionBolus   TreatmentEventType = "Correction Bolus"
	DADAlert          TreatmentEventType = "D.A.D. Alert"
	Exercise          TreatmentEventType = "Exercise"
	InsulinChange     TreatmentEventType = "Insulin Change"
	MealBolus         TreatmentEventType = "Meal Bolus"
	Note              TreatmentEventType = "Note"
	OpenAPSOffline    TreatmentEventType = "OpenAPS Offline"
	ProfileSwitch     TreatmentEventType = "Profile Switch"
	PumpBatteryChange TreatmentEventType = "Pump Battery Change"
	Question          TreatmentEventType = "Question"
	SensorChange      TreatmentEventType = "Sensor Change"
	SensorStart       TreatmentEventType = "Sensor Start"
	SiteChange        TreatmentEventType = "Site Change"
	SnackBolus        TreatmentEventType = "Snack Bolus"
	TempBasal         TreatmentEventType = "Temp Basal"
	TemporaryTarget   TreatmentEventType = "Temporary Target"
	//
	Unknown TreatmentEventType = "Unknown"
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
	case DADAlert:
		return "dad_alert"
	case OpenAPSOffline:
		return "openaps_offline"
	case BolusWizard:
		return "bolus_wizard"
	case Announcement:
		return "announcement"
	case BGCheck:
		return "bg_check"
	case SnackBolus:
		return "snack_bolus"
	case Exercise:
		return "exercise"
	case SiteChange:
		return "site_change"
	case SensorStart:
		return "sensor_start"
	case SensorChange:
		return "sensor_change"
	case InsulinChange:
		return "insulin_change"
	case PumpBatteryChange:
		return "pump_battery_change"
	case Question:
		return "question"
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
