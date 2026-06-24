package nightscout

import (
	"fmt"

	"nighttroop.t1d.nz/data/sql/user"
)

//go:generate goverter gen .

// goverter:converter
// goverter:output:format function
// goverter:output:file ./z_converters.go
type MapConverter interface {

	// goverter:useZeroValueOnPointerInconsistency
	// goverter:ignore CreatedByIdentityID CreatedByIdentityName
	// goverter:map ActivityType Type
	// goverter:map Identifier ID
	// goverter:map Date OccurredAt
	// goverter:map Duration DurationMinutes
	ActivityToCreate(source Activity) user.CreateActivityParams

	// goverter:useZeroValueOnPointerInconsistency
	// goverter:ignore CreatedAt
	// goverter:map ID Identifier
	// goverter:map Type ActivityType
	// goverter:map OccurredAt Date
	// goverter:map DurationMinutes Duration
	// goverter:map CreatedAt SrvCreated
	// goverter:map UpdatedAt SrvModified
	DBModelToActivity(source user.Activity) Activity
}

// ///
// Custom Converters
// ///

func (e *Entry) ToDBModel() *user.GlucoseReading {
	return &user.GlucoseReading{
		ID:                    *e.Identifier,
		Value:                 float64(*e.SGV),
		Type:                  string(SGVType),
		RawJsonData:           &e.RawJSONData,
		UtcOffsetMinutes:      e.UTCOffset.Int64(),
		OccurredAt:            e.Date.Time().UnixMilli(),
		CreatedByIdentityID:   0,  // This will need to be set by the caller based on the user context
		CreatedByIdentityName: "", // This will need to be set by the caller based on the user context
	}
}

func (e *Entry) FromDBModel(sgv *user.GlucoseReading) error {
	e.Identifier = ptr(sgv.ID)
	entryType := EntryType(sgv.Type)
	e.Type = &entryType
	e.SGV = ptr(FlexibleFloat64(sgv.Value))
	dateInt := DateInt(sgv.OccurredAt)
	e.Date = &dateInt
	dateString := DateStringFromInt64(sgv.OccurredAt)
	e.DateString = &dateString
	e.UTCOffset = ptr(FlexibleInt64(sgv.UtcOffsetMinutes))
	device := fmt.Sprintf("%d", sgv.CreatedByIdentityID)
	e.Device = &device

	return nil
}

func (e *Treatment) ToCarbDBModel() (*user.Carb, error) {
	if e.Identifier == nil || e.Carbs == nil || e.Date == nil {
		return nil, fmt.Errorf("missing required fields for DB model")
	}

	return &user.Carb{
		ID:               *e.Identifier,
		Amount:           float64(*e.Carbs),
		Type:             string(CarbCorrection),
		RawJsonData:      &e.RawJSONData,
		UtcOffsetMinutes: e.UTCOffset.Int64(),
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
		RawJsonData:      &e.RawJSONData,
		UtcOffsetMinutes: e.UTCOffset.Int64(),
		OccurredAt:       e.Date.Time().UnixMilli(),
	}, nil
}

func (e *Treatment) ToTempTargetDBModel() (*user.TempTarget, error) {
	if e.Identifier == nil || e.Insulin == nil || e.Date == nil {
		return nil, fmt.Errorf("missing required fields for DB model")
	}

	return &user.TempTarget{
		ID:               *e.Identifier,
		Hightarget:       e.TargetTop.Float64(),
		Lowtarget:        e.TargetBottom.Float64(),
		Start:            e.CreatedAt.Time().UnixMilli(),
		Duration:         e.Duration.Int64(),
		RawJsonData:      &e.RawJSONData,
		UtcOffsetMinutes: e.UTCOffset.Int64(),
	}, nil
}
func (e *Treatment) FromTempTargetDBModel(sgv *user.TempTarget) error {
	e.Identifier = ptr(sgv.ID)
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
	e.Identifier = ptr(sgv.ID)
	eventType := CarbCorrection
	e.EventType = &eventType
	e.Carbs = ptr(FlexibleFloat64(sgv.Amount))
	dateInt := DateStringFromInt64(sgv.OccurredAt)
	e.CreatedAt = &dateInt

	e.UTCOffset = ptr(FlexibleInt64(sgv.UtcOffsetMinutes))

	return nil
}

func (e *Treatment) FromBolusDBModel(sgv *user.Bolus) error {
	e.Identifier = ptr(sgv.ID)
	eventType := CorrectionBolus
	e.EventType = &eventType
	e.Insulin = ptr(FlexibleFloat64(sgv.Value))
	dateInt := DateInt(sgv.OccurredAt)
	e.Date = &dateInt

	e.UTCOffset = ptr(FlexibleInt64(sgv.UtcOffsetMinutes))

	return nil
}
