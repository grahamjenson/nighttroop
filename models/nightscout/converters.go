package nightscout

import (
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
	// goverter:map . DateString | ModelDateStringFromDateIntAndUTCOffset
	DBModelToEntry(source user.GlucoseReading) Entry

	///////
	// TREATMENTS
	///////

	///////
	// Bolus
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
	TreatmentToCreateBolus(source Treatment, idPrefix string, createdByIdentityID int64, createdByIdentityName string) user.CreateBolusParams

	// goverter:useZeroValueOnPointerInconsistency
	// goverter:map  Identifier ID
	// goverter:map  Date OccurredAt
	// goverter:map  Insulin Value
	TreatmentToUpdateBolus(source Treatment) user.UpdateBolusParams

	// goverter:useZeroValueOnPointerInconsistency
	// goverter:ignore Glucose GlucoseType Carbs Duration Notes Profile Percentage Rate Absolute Reason TargetTop TargetBottom Units
	// goverter:map ID Identifier
	// goverter:map Value Insulin
	// goverter:map CreatedByIdentityName EnteredBy
	// goverter:map . EventType  	    | DBTypeToTreatmentType
	// goverter:map . CreatedAt         | ModelDateStringFromDateIntAndUTCOffset
	// goverter:map OccurredAt Date
	// goverter:map CreatedAt SrvCreated
	// goverter:map UpdatedAt SrvModified
	BolusDBModelToTreatment(source user.Bolus) Treatment

	///////
	// Carbs
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
	// goverter:map  Carbs Amount
	TreatmentToCreateCarb(source Treatment, idPrefix string, createdByIdentityID int64, createdByIdentityName string) user.CreateCarbParams

	// goverter:useZeroValueOnPointerInconsistency
	// goverter:map  Identifier ID
	// goverter:map  Date OccurredAt
	// goverter:map  Carbs Amount
	TreatmentToUpdateCarbs(source Treatment) user.UpdateCarbParams

	// goverter:useZeroValueOnPointerInconsistency
	// goverter:ignore Glucose GlucoseType Insulin Duration Notes Profile Percentage Rate Absolute Reason TargetTop TargetBottom Units
	// goverter:map ID Identifier
	// goverter:map Amount Carbs
	// goverter:map CreatedByIdentityName EnteredBy
	// goverter:map . EventType  	    | DBTypeToTreatmentType
	// goverter:map . CreatedAt         | ModelDateStringFromDateIntAndUTCOffset
	// goverter:map OccurredAt Date
	// goverter:map CreatedAt SrvCreated
	// goverter:map UpdatedAt SrvModified
	CarbsDBModelToTreatment(source user.Carb) Treatment

	///////
	// Temp Basal
	///////

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
	TreatmentToCreateTempBasal(source Treatment, idPrefix string, createdByIdentityID int64, createdByIdentityName string) user.CreateTempBasalParams

	// goverter:useZeroValueOnPointerInconsistency
	// goverter:map  Identifier ID
	// goverter:map  Date Start
	// goverter:map  . Value | TempBasalValuePtr
	// goverter:map Duration DurationMinutes
	TreatmentToUpdateTempBasal(source Treatment) user.UpdateTempBasalParams

	// goverter:useZeroValueOnPointerInconsistency
	// goverter:ignore Glucose GlucoseType Insulin Carbs Duration Notes Profile Percentage Rate Absolute Reason TargetTop TargetBottom Units
	// goverter:map ID Identifier
	// goverter:map Value Rate
	// goverter:map Value Absolute
	// goverter:map CreatedByIdentityName EnteredBy
	// goverter:map . EventType  	    | DBTypeToTreatmentType
	// goverter:map . CreatedAt         | ModelDateStringFromDateIntAndUTCOffset
	// goverter:map Start Date
	// goverter:map CreatedAt SrvCreated
	// goverter:map UpdatedAt SrvModified
	TempBasalDBModelToTreatment(source user.TempBasal) Treatment

	///////
	// Temp Targets
	///////
	// goverter:useZeroValueOnPointerInconsistency
	// goverter:context idPrefix
	// goverter:context createdByIdentityID
	// goverter:context createdByIdentityName
	// goverter:map  Date ID                | AssignID
	// goverter:map  Date Start
	// goverter:map  CreatedByIdentityID    | AssignCreatedByID
	// goverter:map  CreatedByIdentityName  | AssignCreatedByName
	// goverter:map Duration DurationMinutes
	// goverter:map TargetTop HighTarget
	// goverter:map TargetBottom LowTarget
	TreatmentToCreateTempTarget(source Treatment, idPrefix string, createdByIdentityID int64, createdByIdentityName string) user.CreateTempTargetParams

	// goverter:useZeroValueOnPointerInconsistency
	// goverter:map  Identifier ID
	// goverter:map  Date Start
	// goverter:map Duration DurationMinutes
	// goverter:map TargetTop HighTarget
	// goverter:map TargetBottom LowTarget
	TreatmentToUpdateTempTarget(source Treatment) user.UpdateTempTargetParams

	///////
	// Notes
	///////
	// goverter:useZeroValueOnPointerInconsistency
	// goverter:context idPrefix
	// goverter:context createdByIdentityID
	// goverter:context createdByIdentityName
	// goverter:map  Date ID                | AssignID
	// goverter:map  Date OccurredAt
	// goverter:map  CreatedByIdentityID    | AssignCreatedByID
	// goverter:map  CreatedByIdentityName  | AssignCreatedByName
	// goverter:map  Notes Content
	TreatmentToCreateNote(source Treatment, idPrefix string, createdByIdentityID int64, createdByIdentityName string) user.CreateNoteParams

	// goverter:useZeroValueOnPointerInconsistency
	// goverter:map  Identifier ID
	// goverter:map  Date OccurredAt
	// goverter:map  Notes Content
	TreatmentToUpdateNote(source Treatment) user.UpdateNoteParams
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

func ToDateInt(createdAt int64) *DateInt {
	x := DateInt(createdAt)
	return &x
}

// ////
// GlucoseReading
// ////
func GlucoseReadingTypeToValue(gr user.GlucoseReading) *FlexibleFloat64 {
	switch gr.Type {
	case "sgv":
		return ptr(FlexibleFloat64(gr.Value))
	case "mbg":
		return ptr(FlexibleFloat64(gr.Value))
	}
	return nil
}

func ModelDateStringFromDateIntAndUTCOffset(m interface{}) *DateString {
	if m == nil {
		return nil
	}

	switch v := m.(type) {
	case user.GlucoseReading:
		ds := DateStringFromDateIntAndUTCOffset(DateInt(v.OccurredAt), FlexibleInt64(v.UtcOffsetMinutes))
		return &ds
	case *user.GlucoseReading:
		ds := DateStringFromDateIntAndUTCOffset(DateInt(v.OccurredAt), FlexibleInt64(v.UtcOffsetMinutes))
		return &ds
	case user.Bolus:
		ds := DateStringFromDateIntAndUTCOffset(DateInt(v.OccurredAt), FlexibleInt64(v.UtcOffsetMinutes))
		return &ds
	case *user.Bolus:
		ds := DateStringFromDateIntAndUTCOffset(DateInt(v.OccurredAt), FlexibleInt64(v.UtcOffsetMinutes))
		return &ds
	case user.Carb:
		ds := DateStringFromDateIntAndUTCOffset(DateInt(v.OccurredAt), FlexibleInt64(v.UtcOffsetMinutes))
		return &ds
	case *user.Carb:
		ds := DateStringFromDateIntAndUTCOffset(DateInt(v.OccurredAt), FlexibleInt64(v.UtcOffsetMinutes))
		return &ds
	case user.TempBasal:
		ds := DateStringFromDateIntAndUTCOffset(DateInt(v.Start), FlexibleInt64(v.UtcOffsetMinutes))
		return &ds
	case *user.TempBasal:
		ds := DateStringFromDateIntAndUTCOffset(DateInt(v.Start), FlexibleInt64(v.UtcOffsetMinutes))
		return &ds
	case user.TempTarget:
		ds := DateStringFromDateIntAndUTCOffset(DateInt(v.Start), FlexibleInt64(v.UtcOffsetMinutes))
		return &ds
	case *user.TempTarget:
		ds := DateStringFromDateIntAndUTCOffset(DateInt(v.Start), FlexibleInt64(v.UtcOffsetMinutes))
		return &ds
	case user.Note:
		ds := DateStringFromDateIntAndUTCOffset(DateInt(v.OccurredAt), FlexibleInt64(v.UtcOffsetMinutes))
		return &ds
	case *user.Note:
		ds := DateStringFromDateIntAndUTCOffset(DateInt(v.OccurredAt), FlexibleInt64(v.UtcOffsetMinutes))
		return &ds
	default:
		return nil
	}

}

// Treatments

func TempBasalValue(t Treatment) float64 {
	v := TempBasalValuePtr(t)
	if v == nil {
		return 0
	}
	return *v
}

func TempBasalValuePtr(t Treatment) *float64 {
	if t.Rate != nil {
		x := float64(*t.Rate)
		return &x
	}
	if t.Absolute != nil {
		x := float64(*t.Absolute)
		return &x
	}
	// TODO handle t.Percent
	return nil
}

func TreatmentTypeToDBType(t *TreatmentEventType) string {
	if t == nil {
		return ""
	}

	return t.ToLocalTypeString()
}

func DBTypeToTreatmentType(m interface{}) *TreatmentEventType {
	if m == nil {
		return nil
	}

	cb := CorrectionBolus
	tt := TemporaryTarget
	tb := TempBasal
	n := Note
	cc := CarbCorrection

	switch v := m.(type) {
	case user.Bolus:
		x := FromLocalTypeString(v.Type)
		if x == nil {
			return &cb
		}
		return x
	case *user.Bolus:
		x := FromLocalTypeString(v.Type)
		if x == nil {
			return &cb
		}
		return x
	case user.Carb:
		x := FromLocalTypeString(v.Type)
		if x == nil {
			return &cc
		}
		return x
	case *user.Carb:
		x := FromLocalTypeString(v.Type)
		if x == nil {
			return &cc
		}
		return x
	case user.TempBasal:
		return &tb
	case *user.TempBasal:
		return &tb
	case user.TempTarget:
		return &tt
	case *user.TempTarget:
		return &tt
	case user.Note:
		return &n
	case *user.Note:
		return &n
	default:
		return nil
	}

}
