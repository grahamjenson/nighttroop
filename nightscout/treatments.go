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
	Identifier *string             `json:"identifier,omitempty"`
	EventType  *TreatmentEventType `json:"eventType,omitempty"`
	Type       *string             `json:"type,omitempty"`
	CreatedAt  *DateString         `json:"created_at,omitempty"`
	Date       *DateInt            `json:"date,omitempty"`
	UTCOffset  *FlexibleInt64      `json:"utcOffset,omitempty"`
	Insulin    *FlexibleFloat64    `json:"insulin,omitempty"`
	Carbs      *FlexibleFloat64    `json:"carbs,omitempty"`
	Notes      *string             `json:"notes,omitempty"`
	EnteredBy  *string             `json:"enteredBy,omitempty"`
	RawJSON    string              `json:"-"`
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
		Eid:                   *e.Identifier,
		Amount:                float64(*e.Carbs),
		Type:                  string(CarbCorrection),
		RawJsonData:           &e.RawJSON,
		UtcOffsetMinutes:      e.UTCOffset.Int64(),
		OccurredAt:            e.Date.Time().UnixMilli(),
		CreatedByIdentityID:   0,  // This will need to be set by the caller based on the user context
		CreatedByIdentityName: "", // This will need to be set by the caller based on the user context
	}, nil
}

func (e *Treatment) FromCarbDBModel(sgv *user.Carb) error {
	e.Identifier = ptr(sgv.Eid)
	eventType := TreatmentEventType(sgv.Type)
	e.EventType = &eventType
	e.Carbs = ptr(FlexibleFloat64(sgv.Amount))
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
