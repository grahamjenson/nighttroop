package nightscout

import (
	"encoding/json"
	"fmt"
	"strings"

	"nighttroop.t1d.nz/data/sql/user"
)

////
// Models
////

type Treatment struct {
	Identifier *string `json:"identifier,omitempty"`

	EventType *TreatmentEventType `json:"eventType,omitempty"`

	CreatedAt *DateString    `json:"created_at"`     // ISO8601 -> time.Time
	Date      *DateInt       `json:"date,omitempty"` // Unix timestamp in milliseconds, only used for some treatment types
	UTCOffset *FlexibleInt64 `json:"utcOffset,omitempty"`

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

	SrvCreated  *DateInt `json:"srvCreated"`
	SrvModified *DateInt `json:"srvModified"`

	RawJSON string `json:"-"`
}

func (e *Treatment) FromMap(m map[string]any) error {
	b, err := json.Marshal(m)
	if err != nil {
		return err
	}
	err = json.Unmarshal(b, e)
	e.RawJSON = string(b)
	return err
}

func (e *Treatment) ToMap() (map[string]any, error) {
	b, err := json.Marshal(e)
	if err != nil {
		return nil, err
	}
	var m map[string]any
	err = json.Unmarshal(b, &m)
	return m, err
}

func (e *Treatment) ToCarbDBModel() (*user.Carb, error) {
	if e.Identifier == nil || e.Carbs == nil || e.Date == nil {
		return nil, fmt.Errorf("missing required fields for DB model")
	}

	return &user.Carb{
		Eid:              *e.Identifier,
		Amount:           float64(*e.Carbs),
		Type:             string(CarbCorrection),
		RawJsonData:      &e.RawJSON,
		UtcOffsetMinutes: e.UTCOffset.Int64(),
		OccurredAt:       e.CreatedAt.Time().UnixMilli(),
	}, nil
}

func (e *Treatment) ToBolusDBModel() (*user.Bolus, error) {
	if e.Identifier == nil || e.Insulin == nil || e.Date == nil {
		return nil, fmt.Errorf("missing required fields for DB model")
	}

	return &user.Bolus{
		Eid:              *e.Identifier,
		Value:            e.Insulin.Float64(),
		Type:             string(CarbCorrection),
		RawJsonData:      &e.RawJSON,
		UtcOffsetMinutes: e.UTCOffset.Int64(),
		OccurredAt:       e.Date.Time().UnixMilli(),
	}, nil
}

func (e *Treatment) ToTempTargetDBModel() (*user.TempTarget, error) {
	if e.Identifier == nil || e.Insulin == nil || e.Date == nil {
		return nil, fmt.Errorf("missing required fields for DB model")
	}

	return &user.TempTarget{
		Eid:              *e.Identifier,
		Hightarget:       e.TargetTop.Float64(),
		Lowtarget:        e.TargetBottom.Float64(),
		Start:            e.CreatedAt.Time().UnixMilli(),
		Duration:         e.Duration.Int64(),
		RawJsonData:      &e.RawJSON,
		UtcOffsetMinutes: e.UTCOffset.Int64(),
	}, nil
}
func (e *Treatment) FromTempTargetDBModel(sgv *user.TempTarget) error {
	e.Identifier = ptr(sgv.Eid)
	eventType := TemporaryTarget
	e.EventType = &eventType
	e.TargetTop = ptr(FlexibleFloat64(sgv.Hightarget))
	e.TargetBottom = ptr(FlexibleFloat64(sgv.Lowtarget))
	dateInt := DateStringFromInt64(sgv.Start)
	e.CreatedAt = &dateInt

	e.UTCOffset = ptr(FlexibleInt64(sgv.UtcOffsetMinutes))

	return nil
}

func (e *Treatment) FromCarbDBModel(sgv *user.Carb) error {
	e.Identifier = ptr(sgv.Eid)
	eventType := CarbCorrection
	e.EventType = &eventType
	e.Carbs = ptr(FlexibleFloat64(sgv.Amount))
	dateInt := DateStringFromInt64(sgv.OccurredAt)
	e.CreatedAt = &dateInt

	e.UTCOffset = ptr(FlexibleInt64(sgv.UtcOffsetMinutes))

	return nil
}

func (e *Treatment) FromBolusDBModel(sgv *user.Bolus) error {
	e.Identifier = ptr(sgv.Eid)
	eventType := CorrectionBolus
	e.EventType = &eventType
	e.Insulin = ptr(FlexibleFloat64(sgv.Value))
	dateInt := DateInt(sgv.OccurredAt)
	e.Date = &dateInt

	e.UTCOffset = ptr(FlexibleInt64(sgv.UtcOffsetMinutes))

	return nil
}

////
// Mappers
////

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
