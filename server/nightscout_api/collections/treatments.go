package collections

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"nighttroop.t1d.nz/data/sql/user"
	"nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/nsutils"
	"nighttroop.t1d.nz/utils"
	"nighttroop.t1d.nz/utils/httputils"
)

type TreatmentCollection struct{}

func (c TreatmentCollection) Create(ctx context.Context, data map[string]any) (*nsutils.Record, bool, error) {
	userContext := httputils.GetUserContext(ctx)

	var treatmentModel Treatment
	raw_json, err := MapToStruct(data, &treatmentModel)
	if err != nil {
		return nil, false, fmt.Errorf("failed to map data to treatment model: %w", err)
	}
	externalID := utils.NextEID("ns")
	created := false
	switch treatmentModel.EventType {
	case "Carb Correction":
		err = userContext.UserDB.Queries.CreateCarb(ctx, user.CreateCarbParams{
			Eid:                   externalID,
			Amount:                treatmentModel.Carbs.Float64(),
			RawJsonData:           &raw_json,
			CreatedByIdentityName: userContext.IdentityModel.Name,
			CreatedByIdentityID:   userContext.IdentityModel.ID,
		})
		if err != nil {
			return nil, false, fmt.Errorf("failed to create carb in database: %w", err)
		}
		created = true
	}
	if !created {
		return nil, false, fmt.Errorf("unsupported treatment event type: %s", treatmentModel.EventType)
	}
	return &nsutils.Record{
		Collection:  "treatments",
		ID:          externalID,
		Data:        data,
		SrvModified: time.Now().UnixMilli(),
	}, true, nil
}

func (c TreatmentCollection) Get(ctx context.Context, identifier string) (*nsutils.Record, error) {
	return nil, nil
}

func (c TreatmentCollection) Search(ctx context.Context, query nsutils.Query) ([]nsutils.Record, error) {
	userContext := httputils.GetUserContext(ctx)
	records := []nsutils.Record{}
	carbs, err := userContext.UserDB.Queries.AllCarbs(ctx)
	if err != nil {
		return nil, err
	}
	records = append(records, CreateRecordSlice(carbs, CarbsToRecord)...)

	return records, nil
}

func (c TreatmentCollection) Update(ctx context.Context, identifier string, patch map[string]any) (*nsutils.Record, error) {
	return nil, nil
}

func (c TreatmentCollection) Delete(ctx context.Context, identifier string, permanent bool) error {
	return nil
}

func (c TreatmentCollection) DeleteMatching(ctx context.Context, query nsutils.Query, permanent bool) (int, error) {
	return 0, nil
}

////
// Models
////

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

func (t *TreatmentEventType) UnmarshalText(text []byte) error {
	switch strings.ToLower(strings.TrimSpace(string(text))) {
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
		*t = Unknown
	}

	return nil
}

func (t *TreatmentEventType) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	return t.UnmarshalText([]byte(s))
}

type Treatment struct {
	Identifier string             `json:"identifier,omitempty"`
	EventType  TreatmentEventType `json:"eventType,omitempty"`
	Type       string             `json:"type,omitempty"`
	CreatedAt  string             `json:"created_at,omitempty"`
	Date       int64              `json:"date,omitempty"`
	Insulin    FlexibleFloat64    `json:"insulin,omitempty"`
	Carbs      FlexibleFloat64    `json:"carbs,omitempty"`
	Notes      string             `json:"notes,omitempty"`
	EnteredBy  string             `json:"enteredBy,omitempty"`
	Payload    map[string]any     `json:"payload,omitempty"`
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

func CarbsToRecord(carbs *user.Carb) nsutils.Record {
	if carbs == nil {
		return nsutils.Record{}
	}

	treatmentModel := Treatment{
		Identifier: fmt.Sprintf("%d", carbs.ID),
		EventType:  "Carb Correction",
		Carbs:      FlexibleFloat64(carbs.Amount),
		Date:       carbs.CreatedAt,
		CreatedAt:  dateString(carbs.CreatedAt),
		EnteredBy:  fmt.Sprintf("%d", carbs.CreatedByIdentityID),
	}

	return nsutils.Record{
		Collection:  "treatments",
		ID:          fmt.Sprintf("%d", carbs.ID),
		Data:        StructToMap(treatmentModel),
		SrvCreated:  carbs.CreatedAt,
		SrvModified: carbs.CreatedAt,
		Subject:     "",
		IsValid:     true,
		DeletedAt:   nil,
	}
}
