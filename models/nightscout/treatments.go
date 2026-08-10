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

	Name *ShortString `json:"name,omitempty"` // Name of the activity

	// GlucoseReading
	Glucose     *FlexibleFloat64 `json:"glucose,omitempty"`
	GlucoseType *GlucoseType     `json:"glucoseType,omitempty"` // "Sensor" or "Finger"

	// Food
	Carbs    *FlexibleFloat64 `json:"carbs,omitempty"`
	Insulin  *FlexibleFloat64 `json:"insulin,omitempty"`
	Protein  *FlexibleFloat64 `json:"protein,omitempty"`
	Fat      *FlexibleFloat64 `json:"fat,omitempty"`
	Fibre    *FlexibleFloat64 `json:"fibre,omitempty"`
	Energy   *FlexibleFloat64 `json:"energy,omitempty"`
	PreBolus *FlexibleFloat64 `json:"preBolus,omitempty"`

	// Exercise
	ActivityType *ShortString   `json:"activityType,omitempty"` // Type of activity
	Steps        *FlexibleInt64 `json:"steps,omitempty"`        // Step count
	Calories     *FlexibleInt64 `json:"calories,omitempty"`     // Calories burned
	HeartRate    *FlexibleInt64 `json:"heartRate,omitempty"`    // Heart rate

	SplitNow *FlexibleFloat64 `json:"splitNow,omitempty"`
	SplitExt *FlexibleFloat64 `json:"splitExt,omitempty"`

	Duration *FlexibleInt64 `json:"duration,omitempty"` // minutes

	Notes     *SanitizedString `json:"notes,omitempty"`
	EnteredBy *ShortString     `json:"enteredBy,omitempty"`

	// Treatment-specific fields
	Profile             *ShortString     `json:"profile,omitempty"`
	OriginalProfileName *ShortString     `json:"originalProfileName,omitempty"`
	ProfileJSON         *SanitizedString `json:"profileJson,omitempty"`
	Percentage          *FlexibleFloat64 `json:"percentage,omitempty"`
	Timeshift           *FlexibleInt64   `json:"timeshift,omitempty"`

	Rate         *FlexibleFloat64 `json:"rate,omitempty"`
	Absolute     *FlexibleFloat64 `json:"absolute,omitempty"`
	Reason       *ShortString     `json:"reason,omitempty"`
	TargetTop    *FlexibleFloat64 `json:"targetTop,omitempty"`
	TargetBottom *FlexibleFloat64 `json:"targetBottom,omitempty"`
	Units        *Units           `json:"units,omitempty"`
	PumpType     *ShortString     `json:"pumpType,omitempty"`

	// Default String Fields
	Device  *ShortString `json:"device,omitempty"`  // Device used for the activity
	App     *ShortString `json:"app,omitempty"`     // App used for the activity
	Subject *ShortString `json:"subject,omitempty"` // Subject of the activity

	// Date Fields
	CreatedAt        *DateString    `json:"created_at,omitempty"` // ISO8601 -> time.Time
	Date             *DateInt       `json:"-"`
	UtcOffsetMinutes *FlexibleInt64 `json:"-"`

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
	InsulinChange     TreatmentEventType = "Insulin Change" // device event
	MealBolus         TreatmentEventType = "Meal Bolus"
	Note              TreatmentEventType = "Note"
	OpenAPSOffline    TreatmentEventType = "OpenAPS Offline" // device event
	ProfileSwitch     TreatmentEventType = "Profile Switch"
	PumpBatteryChange TreatmentEventType = "Pump Battery Change" // device event
	Question          TreatmentEventType = "Question"
	SensorChange      TreatmentEventType = "Sensor Change" // device event
	SensorStart       TreatmentEventType = "Sensor Start"  // device event
	SiteChange        TreatmentEventType = "Site Change"   // device event
	SnackBolus        TreatmentEventType = "Snack Bolus"
	TempBasal         TreatmentEventType = "Temp Basal"
	TemporaryTarget   TreatmentEventType = "Temporary Target"
	//
	Unknown TreatmentEventType = "Unknown"
)

func (t TreatmentEventType) String() string {
	return string(t)
}

func ToSnakeCase(s string) string {
	return strings.ToLower(strings.Join(strings.Fields(s), "_"))
}

var TreatmentEventTypes = []TreatmentEventType{Announcement, BGCheck, BolusWizard, CarbCorrection, ComboBolus, CorrectionBolus, DADAlert, Exercise, InsulinChange, MealBolus, Note, OpenAPSOffline, ProfileSwitch, PumpBatteryChange, Question, SensorChange, SensorStart, SiteChange, SnackBolus, TempBasal, TemporaryTarget}

func (t *TreatmentEventType) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	switch ToSnakeCase(strings.TrimSpace(s)) {
	case "announcement":
		*t = Announcement
	case "bg_check":
		*t = BGCheck
	case "bolus_wizard":
		*t = BolusWizard
	case "carb_correction":
		*t = CarbCorrection
	case "combo_bolus":
		*t = ComboBolus
	case "correction_bolus":
		*t = CorrectionBolus
	case "d.a.d._alert":
		*t = DADAlert
	case "exercise":
		*t = Exercise
	case "insulin_change":
		*t = InsulinChange
	case "meal_bolus":
		*t = MealBolus
	case "note":
		*t = Note
	case "openaps_offline":
		*t = OpenAPSOffline
	case "profile_switch":
		*t = ProfileSwitch
	case "pump_battery_change":
		*t = PumpBatteryChange
	case "question":
		*t = Question
	case "sensor_change":
		*t = SensorChange
	case "sensor_start":
		*t = SensorStart
	case "site_change":
		*t = SiteChange
	case "snack_bolus":
		*t = SnackBolus
	case "temp_basal":
		*t = TempBasal
	case "temporary_target":
		*t = TemporaryTarget
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

func (t Treatment) GetDevice() *string {
	return t.Device.Stringp()
}

func (t Treatment) GetApp() *string {
	return t.App.Stringp()
}

func (t Treatment) GetSubject() *string {
	return t.Subject.Stringp()
}

func (t Treatment) GetDate() *int64 {
	return t.Date.Int64p()
}
