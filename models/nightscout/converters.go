package nightscout

import (
	"fmt"

	"nighttroop.t1d.nz/data/sql/user"
	"nighttroop.t1d.nz/utils"
)

//go:generate goverter gen .

// goverter:converter
// goverter:output:format function
// goverter:output:file ./z_converters.go
type MapConverter interface {

	///////
	// ACTIVITIES
	///////

	// goverter:useZeroValueOnPointerInconsistency
	// goverter:ignore CreatedByIdentityID CreatedByIdentityName
	// goverter:map ActivityType Type
	// goverter:map Identifier ID
	// goverter:map Date OccurredAt
	// goverter:map Duration DurationMinutes
	// goverter:context idPrefix
	// goverter:context createdByIdentityID
	// goverter:context createdByIdentityName
	// goverter:map  Date ID                | AssignID
	// goverter:map  CreatedByIdentityID    | AssignCreatedByID
	// goverter:map  CreatedByIdentityName  | AssignCreatedByName
	ActivityToCreate(source Activity, idPrefix string, createdByIdentityID int64, createdByIdentityName string) user.CreateActivityParams

	// goverter:useZeroValueOnPointerInconsistency
	// goverter:ignore CreatedAt
	// goverter:map ID Identifier
	// goverter:map Type ActivityType
	// goverter:map OccurredAt Date
	// goverter:map DurationMinutes Duration
	// goverter:map CreatedAt SrvCreated
	// goverter:map UpdatedAt SrvModified
	DBModelToActivity(source user.Activity) Activity

	///////
	// ENTRIES
	///////

	// goverter:useZeroValueOnPointerInconsistency
	// goverter:context idPrefix
	// goverter:context createdByIdentityID
	// goverter:context createdByIdentityName
	// goverter:map  Date ID                | AssignID
	// goverter:map  Date OccurredAt
	// goverter:map  CreatedByIdentityID    | AssignCreatedByID
	// goverter:map  CreatedByIdentityName  | AssignCreatedByName
	EntryToCreate(source Entry, idPrefix string, createdByIdentityID int64, createdByIdentityName string) user.CreateGlucoseReadingParams

	// goverter:useZeroValueOnPointerInconsistency
	// goverter:map Identifier ID
	// goverter:map Date OccurredAt
	EntryToUpdate(source Entry) user.UpdateGlucoseReadingParams

	// goverter:useZeroValueOnPointerInconsistency
	// goverter:ignore Direction Trend Units Filtered Unfiltered RSSI Noise SysTime
	// goverter:map ID Identifier
	// goverter:map . SGV | GlucoseReadingTypeToValue
	// goverter:map . MBG | GlucoseReadingTypeToValue
	// goverter:map CreatedByIdentityName Device
	// goverter:map OccurredAt Date
	// goverter:map OccurredAt Mills
	// goverter:map CreatedAt SrvCreated
	// goverter:map UpdatedAt SrvModified
	// goverter:map . DateString | GlucoseReadingDateStringFromDateIntAndUTCOffset
	DBModelToEntry(source user.GlucoseReading) Entry

	///////
	// TREATMENTS
	///////

	// goverter:useZeroValueOnPointerInconsistency
	// goverter:context idPrefix
	// goverter:context createdByIdentityID
	// goverter:context createdByIdentityName
	// goverter:map  Date ID                | AssignID
	// goverter:map  Date OccurredAt
	// goverter:map  CreatedByIdentityID    | AssignCreatedByID
	// goverter:map  CreatedByIdentityName  | AssignCreatedByName
	// goverter:map  EventType Type 	    | TreatmentTypeToDBType
	// goverter:map  Insulin Value
	TreatmentToCreateBolus(source Treatment, idPrefix string, createdByIdentityID int64, createdByIdentityName string) *user.CreateBolusParams

	// goverter:useZeroValueOnPointerInconsistency
	// goverter:context idPrefix
	// goverter:context createdByIdentityID
	// goverter:context createdByIdentityName
	// goverter:map  Date ID                | AssignID
	// goverter:map  Date OccurredAt
	// goverter:map  CreatedByIdentityID    | AssignCreatedByID
	// goverter:map  CreatedByIdentityName  | AssignCreatedByName
	// goverter:map  EventType Type 	    | TreatmentTypeToDBType
	// goverter:map  Carbs Amount
	TreatmentToCreateCarb(source Treatment, idPrefix string, createdByIdentityID int64, createdByIdentityName string) *user.CreateCarbParams

	// goverter:useZeroValueOnPointerInconsistency
	// goverter:context idPrefix
	// goverter:context createdByIdentityID
	// goverter:context createdByIdentityName
	// goverter:map  Date ID                | AssignID
	// goverter:map  Date Start
	// goverter:map  CreatedByIdentityID    | AssignCreatedByID
	// goverter:map  CreatedByIdentityName  | AssignCreatedByName
	// goverter:map  . Value | TempBasalValue
	// goverter:map Duration DurationMinutes
	TreatmentToCreateTempBasal(source Treatment, idPrefix string, createdByIdentityID int64, createdByIdentityName string) *user.CreateTempBasalParams

	// TreatmentToCreateTempTarget(source Treatment, idPrefix string, createdByIdentityID int64, createdByIdentityName string) *user.CreateTempTargetParams

	// goverter:useZeroValueOnPointerInconsistency
	// goverter:context idPrefix
	// goverter:context createdByIdentityID
	// goverter:context createdByIdentityName
	// goverter:map  Date ID                | AssignID
	// goverter:map  Date OccurredAt
	// goverter:map  CreatedByIdentityID    | AssignCreatedByID
	// goverter:map  CreatedByIdentityName  | AssignCreatedByName
	// goverter:map  Notes Content
	TreatmentToCreateNote(source Treatment, idPrefix string, createdByIdentityID int64, createdByIdentityName string) *user.CreateNoteParams
}

// //
// Custom Converter Methods
// //

// goverter:context idPrefix
func AssignID(date *DateInt, idPrefix string) string {
	return utils.NextDateID(idPrefix, date.Int64())
}

// goverter:context createdByIdentityName
func AssignCreatedByName(createdByIdentityName string) string {
	return createdByIdentityName
}

// goverter:context createdByIdentityID
func AssignCreatedByID(createdByIdentityID int64) int64 {
	return createdByIdentityID
}

// GlucoseReading
func GlucoseReadingTypeToValue(gr user.GlucoseReading) *FlexibleFloat64 {
	switch gr.Type {
	case "sgv":
		return ptr(FlexibleFloat64(gr.Value))
	case "mbg":
		return ptr(FlexibleFloat64(gr.Value))
	}
	return nil
}

func GlucoseReadingDateStringFromDateIntAndUTCOffset(gr user.GlucoseReading) *DateString {
	ds := DateStringFromDateIntAndUTCOffset(DateInt(gr.OccurredAt), FlexibleInt64(gr.UtcOffsetMinutes))
	return &ds
}

// Treatments

func TempBasalValue(t Treatment) float64 {
	if t.Rate != nil {
		return float64(*t.Rate)
	}
	if t.Absolute != nil {
		return float64(*t.Absolute)
	}
	// TODO handle t.Percent
	return 0
}

// /
func TreatmentTypeToDBType(t *TreatmentEventType) string {
	if t == nil {
		return ""
	}

	switch *t {
	case CorrectionBolus:
		return "bolus"
	case ComboBolus:
		return "combo_bolus"
	case TemporaryTarget:
		return "temp_target"
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

// ///
// Custom Converters
// ///

func (e *Treatment) ToCarbDBModel() (*user.Carb, error) {
	if e.Identifier == nil || e.Carbs == nil || e.Date == nil {
		return nil, fmt.Errorf("missing required fields for DB model")
	}

	return &user.Carb{
		ID:               *e.Identifier,
		Amount:           float64(*e.Carbs),
		Type:             string(CarbCorrection),
		RawJsonData:      &e.RawJsonData,
		UtcOffsetMinutes: e.UtcOffsetMinutes.Int64(),
		OccurredAt:       e.CreatedAt.Time().UnixMilli(),
	}, nil
}

func (e *Treatment) ToBolusDBModel() (*user.Bolus, error) {
	if e.Identifier == nil || e.Insulin == nil || e.Date == nil {
		return nil, fmt.Errorf("missing required fields for DB model")
	}

	return &user.Bolus{
		ID:               *e.Identifier,
		Value:            e.Insulin.Float64(),
		Type:             string(CorrectionBolus),
		RawJsonData:      &e.RawJsonData,
		UtcOffsetMinutes: e.UtcOffsetMinutes.Int64(),
		OccurredAt:       e.Date.Time().UnixMilli(),
	}, nil
}

func (e *Treatment) ToTempTargetDBModel() (*user.TempTarget, error) {
	if e.Identifier == nil || e.Insulin == nil || e.Date == nil {
		return nil, fmt.Errorf("missing required fields for DB model")
	}

	return &user.TempTarget{
		ID:               *e.Identifier,
		HighTarget:       e.TargetTop.Float64(),
		LowTarget:        e.TargetBottom.Float64(),
		Start:            e.CreatedAt.Time().UnixMilli(),
		DurationMinutes:  e.Duration.Int64(),
		RawJsonData:      &e.RawJsonData,
		UtcOffsetMinutes: e.UtcOffsetMinutes.Int64(),
	}, nil
}
func (e *Treatment) FromTempTargetDBModel(sgv *user.TempTarget) error {
	e.Identifier = ptr(sgv.ID)
	eventType := TemporaryTarget
	e.EventType = &eventType
	e.TargetTop = ptr(FlexibleFloat64(sgv.HighTarget))
	e.TargetBottom = ptr(FlexibleFloat64(sgv.LowTarget))

	e.UtcOffsetMinutes = ptr(FlexibleInt64(sgv.UtcOffsetMinutes))

	return nil
}

func (e *Treatment) FromCarbDBModel(sgv *user.Carb) error {
	e.Identifier = ptr(sgv.ID)
	eventType := CarbCorrection
	e.EventType = &eventType
	e.Carbs = ptr(FlexibleFloat64(sgv.Amount))

	e.UtcOffsetMinutes = ptr(FlexibleInt64(sgv.UtcOffsetMinutes))

	return nil
}

func (e *Treatment) FromBolusDBModel(sgv *user.Bolus) error {
	e.Identifier = ptr(sgv.ID)
	eventType := CorrectionBolus
	e.EventType = &eventType
	e.Insulin = ptr(FlexibleFloat64(sgv.Value))
	dateInt := DateInt(sgv.OccurredAt)
	e.Date = &dateInt

	e.UtcOffsetMinutes = ptr(FlexibleInt64(sgv.UtcOffsetMinutes))

	return nil
}
