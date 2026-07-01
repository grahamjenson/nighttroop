package nightscout

import (
	"nighttroop.t1d.nz/data/sql/user"
	"nighttroop.t1d.nz/utils"
)

//go:generate goverter gen .

type CreateContext struct {
	IDPrefix              string
	CreatedByIdentityID   int64
	CreatedByIdentityName string
}

type UpdateContext struct {
	UpdatedByIdentityID   int64
	UpdatedByIdentityName string
}

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
	// goverter:map  App CreatedByAppName
	// goverter:map  Device CreatedByDeviceName
	// goverter:context ctx
	// goverter:map  Date ID                | AssignID
	// goverter:map  CreatedByIdentityID    | AssignCreatedByID
	// goverter:map  CreatedByIdentityName  | AssignCreatedByName
	ActivityToCreate(source Activity, ctx CreateContext) user.CreateActivityParams

	// goverter:useZeroValueOnPointerInconsistency
	// goverter:map . CreatedAt | ModelDateStringFromDateIntAndUTCOffset
	// goverter:map CreatedByIdentityName Subject
	// goverter:map CreatedByAppName App
	// goverter:map CreatedByDeviceName Device
	// goverter:map ID Identifier
	// goverter:map Type ActivityType
	// goverter:map OccurredAt Date
	// goverter:map DurationMinutes Duration
	// goverter:map CreatedAt SrvCreated
	// goverter:map UpdatedAt SrvModified
	DBModelToActivity(source *user.Activity) Activity

	///////
	// DEVICE STATUS TODO
	///////

	// goverter:useZeroValueOnPointerInconsistency
	// goverter:ignore InsertionTime Type Status ChargingState Timestamp Iob Cob Basaliob Treatmentiob PumpReservoir
	// goverter:map Identifier ID
	// goverter:map Date OccurredAt
	// goverter:map UploaderBattery BatteryPercentage
	// goverter:map  App CreatedByAppName
	// goverter:map  Device CreatedByDeviceName
	// goverter:context ctx
	// goverter:map  Date ID                | AssignID
	// goverter:map  CreatedByIdentityID    | AssignCreatedByID
	// goverter:map  CreatedByIdentityName  | AssignCreatedByName
	DeviceStatusToCreate(source DeviceStatus, ctx CreateContext) user.CreateDeviceStatusParams

	// goverter:useZeroValueOnPointerInconsistency
	// goverter:ignore Uploader Loop Pump OpenAPS
	// goverter:map ID Identifier
	// goverter:map . CreatedAt | ModelDateStringFromDateIntAndUTCOffset
	// goverter:map CreatedByIdentityName Subject
	// goverter:map CreatedByAppName App
	// goverter:map CreatedByDeviceName Device
	// goverter:map OccurredAt Date
	// goverter:map BatteryPercentage UploaderBattery
	// goverter:map CreatedByIdentityName Device
	// goverter:map CreatedAt SrvCreated
	// goverter:map UpdatedAt SrvModified
	DBModelToDeviceStatus(source *user.DeviceStatus) DeviceStatus

	///////
	// ENTRIES
	///////

	// goverter:useZeroValueOnPointerInconsistency
	// goverter:context ctx
	// goverter:map  Date ID                | AssignID
	// goverter:map  Date OccurredAt
	// goverter:map  App CreatedByAppName
	// goverter:map  Device CreatedByDeviceName
	// goverter:map  CreatedByIdentityID    | AssignCreatedByID
	// goverter:map  CreatedByIdentityName  | AssignCreatedByName
	EntryToCreate(source Entry, ctx CreateContext) user.CreateGlucoseReadingParams

	// goverter:useZeroValueOnPointerInconsistency
	// goverter:map Identifier ID
	// goverter:map Date OccurredAt
	// goverter:map UpdatedByIdentityName | AssignUpdatedByName
	// goverter:map UpdatedByIdentityID   | AssignUpdatedByID
	// goverter:context ctx
	EntryToUpdate(source Entry, ctx UpdateContext) user.UpdateGlucoseReadingParams

	// goverter:useZeroValueOnPointerInconsistency
	// goverter:ignore Direction Trend Units Filtered Unfiltered RSSI Noise SysTime
	// goverter:map CreatedByIdentityName Subject
	// goverter:map CreatedByAppName App
	// goverter:map CreatedByDeviceName Device
	// goverter:map ID Identifier
	// goverter:map . SGV | SGVGlucoseReadingTypeToValue
	// goverter:map . MBG | MBGGlucoseReadingTypeToValue
	// goverter:map CreatedByIdentityName Device
	// goverter:map OccurredAt Date
	// goverter:map OccurredAt Mills
	// goverter:map CreatedAt SrvCreated
	// goverter:map UpdatedAt SrvModified
	// goverter:map . DateString | ModelDateStringFromDateIntAndUTCOffset
	DBModelToEntry(source *user.GlucoseReading) Entry

	///////
	// FOOD TODO
	///////

	// goverter:useZeroValueOnPointerInconsistency
	// goverter:ignore Calories
	// goverter:context ctx
	// goverter:map  App CreatedByAppName
	// goverter:map  Device CreatedByDeviceName
	// goverter:map  Date ID                | AssignID
	// goverter:map  CreatedByIdentityID    | AssignCreatedByID
	// goverter:map  CreatedByIdentityName  | AssignCreatedByName
	FoodToCreate(source Food, ctx CreateContext) user.CreateFoodParams

	// goverter:useZeroValueOnPointerInconsistency
	// goverter:ignore Device App Subject Food Portions HideAfterUse Hidden Position
	// goverter:ignore Category Subcategory Portion Unit Fat Protein Energy GI Date
	// goverter:map ID Identifier
	// goverter:map CreatedAt SrvCreated
	// goverter:map UpdatedAt SrvModified
	DBModelToFood(source *user.Food) Food

	///////
	// Profile TODO
	///////

	// goverter:useZeroValueOnPointerInconsistency
	// goverter:context ctx
	// goverter:ignore  Start DurationMinutes Source Percent Dia Units Timezone CarbRatiosJson InsulinSensitivitiesJson BasalRatesJson TargetsJson
	// goverter:map  App CreatedByAppName
	// goverter:map  Device CreatedByDeviceName
	// goverter:map  Date ID                | AssignID
	// goverter:map  DefaultProfile Name
	// goverter:map  CreatedByIdentityID    | AssignCreatedByID
	// goverter:map  CreatedByIdentityName  | AssignCreatedByName
	ProfileToCreate(source Profile, ctx CreateContext) user.CreateProfileParams

	// goverter:useZeroValueOnPointerInconsistency
	// goverter:ignore Device App Subject
	// goverter:ignore DefaultProfile Store
	// goverter:map ID Identifier
	// goverter:map . StartDate | ModelDateStringFromDateIntAndUTCOffset
	// goverter:map OccurredAt Mills
	// goverter:map OccurredAt Date
	// goverter:map CreatedAt SrvCreated
	// goverter:map UpdatedAt SrvModified
	DBModelToProfile(source *user.Profile) Profile

	///////
	// TREATMENTS
	///////

	///////
	// Bolus
	///////

	// goverter:useZeroValueOnPointerInconsistency
	// goverter:context ctx
	// goverter:map  App CreatedByAppName
	// goverter:map  Device CreatedByDeviceName
	// goverter:map  Date ID                | AssignID
	// goverter:map  Date OccurredAt
	// goverter:map  CreatedByIdentityID    | AssignCreatedByID
	// goverter:map  CreatedByIdentityName  | AssignCreatedByName
	// goverter:map  EventType Type 	    | TreatmentTypeToDBType
	// goverter:map  Insulin Value
	TreatmentToCreateBolus(source Treatment, ctx CreateContext) user.CreateBolusParams

	// goverter:useZeroValueOnPointerInconsistency
	// goverter:map  Identifier ID
	// goverter:map  Date OccurredAt
	// goverter:map  Insulin Value
	// goverter:map UpdatedByIdentityName | AssignUpdatedByName
	// goverter:map UpdatedByIdentityID   | AssignUpdatedByID
	// goverter:context ctx
	TreatmentToUpdateBolus(source Treatment, ctx UpdateContext) user.UpdateBolusParams

	// goverter:useZeroValueOnPointerInconsistency
	// goverter:ignore Glucose GlucoseType Carbs Duration Notes Profile Percentage Rate Absolute Reason TargetTop TargetBottom Units Protein Fat PreBolus SplitNow SplitExt Device App Subject
	// goverter:map ID Identifier
	// goverter:map Value Insulin
	// goverter:map CreatedByIdentityName EnteredBy
	// goverter:map . EventType  	    | DBTypeToTreatmentType
	// goverter:map . CreatedAt         | ModelDateStringFromDateIntAndUTCOffset
	// goverter:map OccurredAt Date
	// goverter:map CreatedAt SrvCreated
	// goverter:map UpdatedAt SrvModified
	BolusDBModelToTreatment(source *user.Bolus) Treatment

	///////
	// Carbs
	///////

	// goverter:useZeroValueOnPointerInconsistency
	// goverter:context ctx
	// goverter:map  App CreatedByAppName
	// goverter:map  Device CreatedByDeviceName
	// goverter:map  Date ID                | AssignID
	// goverter:map  Date OccurredAt
	// goverter:map  CreatedByIdentityID    | AssignCreatedByID
	// goverter:map  CreatedByIdentityName  | AssignCreatedByName
	// goverter:map  EventType Type 	    | TreatmentTypeToDBType
	// goverter:map  Carbs Amount
	TreatmentToCreateCarb(source Treatment, ctx CreateContext) user.CreateCarbParams

	// goverter:useZeroValueOnPointerInconsistency
	// goverter:map  Identifier ID
	// goverter:map  Date OccurredAt
	// goverter:map  Carbs Amount
	// goverter:map UpdatedByIdentityName | AssignUpdatedByName
	// goverter:map UpdatedByIdentityID   | AssignUpdatedByID
	// goverter:context ctx
	TreatmentToUpdateCarbs(source Treatment, ctx UpdateContext) user.UpdateCarbParams

	// goverter:useZeroValueOnPointerInconsistency
	// goverter:ignore Glucose GlucoseType Insulin Duration Notes Profile Percentage Rate Absolute Reason TargetTop TargetBottom Units Protein Fat PreBolus SplitNow SplitExt Device App Subject
	// goverter:map ID Identifier
	// goverter:map Amount Carbs
	// goverter:map CreatedByIdentityName EnteredBy
	// goverter:map . EventType  	    | DBTypeToTreatmentType
	// goverter:map . CreatedAt         | ModelDateStringFromDateIntAndUTCOffset
	// goverter:map OccurredAt Date
	// goverter:map CreatedAt SrvCreated
	// goverter:map UpdatedAt SrvModified
	CarbsDBModelToTreatment(source *user.Carb) Treatment

	///////
	// Temp Basal
	///////

	// goverter:useZeroValueOnPointerInconsistency
	// goverter:context ctx
	// goverter:map  App CreatedByAppName
	// goverter:map  Device CreatedByDeviceName
	// goverter:map  Date ID                | AssignID
	// goverter:map  Date Start
	// goverter:map  CreatedByIdentityID    | AssignCreatedByID
	// goverter:map  CreatedByIdentityName  | AssignCreatedByName
	// goverter:map  . Value | TempBasalValue
	// goverter:map Duration DurationMinutes
	TreatmentToCreateTempBasal(source Treatment, ctx CreateContext) user.CreateTempBasalParams

	// goverter:useZeroValueOnPointerInconsistency
	// goverter:map  Identifier ID
	// goverter:map  Date Start
	// goverter:map  . Value | TempBasalValuePtr
	// goverter:map Duration DurationMinutes
	// goverter:map UpdatedByIdentityName | AssignUpdatedByName
	// goverter:map UpdatedByIdentityID   | AssignUpdatedByID
	// goverter:context ctx
	TreatmentToUpdateTempBasal(source Treatment, ctx UpdateContext) user.UpdateTempBasalParams

	// goverter:useZeroValueOnPointerInconsistency
	// goverter:ignore Glucose GlucoseType Insulin Carbs Duration Notes Profile Percentage Rate Absolute Reason TargetTop TargetBottom Units Protein Fat PreBolus SplitNow SplitExt Device App Subject
	// goverter:map ID Identifier
	// goverter:map Value Rate
	// goverter:map Value Absolute
	// goverter:map CreatedByIdentityName EnteredBy
	// goverter:map . EventType  	    | DBTypeToTreatmentType
	// goverter:map . CreatedAt         | ModelDateStringFromDateIntAndUTCOffset
	// goverter:map Start Date
	// goverter:map CreatedAt SrvCreated
	// goverter:map UpdatedAt SrvModified
	TempBasalDBModelToTreatment(source *user.TempBasal) Treatment

	///////
	// Temp Targets
	///////
	// goverter:useZeroValueOnPointerInconsistency
	// goverter:context ctx
	// goverter:map  App CreatedByAppName
	// goverter:map  Device CreatedByDeviceName
	// goverter:map  Date ID                | AssignID
	// goverter:map  Date Start
	// goverter:map  CreatedByIdentityID    | AssignCreatedByID
	// goverter:map  CreatedByIdentityName  | AssignCreatedByName
	// goverter:map Duration DurationMinutes
	// goverter:map TargetTop HighTarget
	// goverter:map TargetBottom LowTarget
	TreatmentToCreateTempTarget(source Treatment, ctx CreateContext) user.CreateTempTargetParams

	// goverter:useZeroValueOnPointerInconsistency
	// goverter:map  Identifier ID
	// goverter:map  Date Start
	// goverter:map Duration DurationMinutes
	// goverter:map TargetTop HighTarget
	// goverter:map TargetBottom LowTarget
	// goverter:map UpdatedByIdentityName | AssignUpdatedByName
	// goverter:map UpdatedByIdentityID   | AssignUpdatedByID
	// goverter:context ctx
	TreatmentToUpdateTempTarget(source Treatment, ctx UpdateContext) user.UpdateTempTargetParams

	///////
	// Notes
	///////
	// goverter:useZeroValueOnPointerInconsistency
	// goverter:context ctx
	// goverter:map  App CreatedByAppName
	// goverter:map  Device CreatedByDeviceName
	// goverter:map  Date ID                | AssignID
	// goverter:map  Date OccurredAt
	// goverter:map  CreatedByIdentityID    | AssignCreatedByID
	// goverter:map  CreatedByIdentityName  | AssignCreatedByName
	// goverter:map  Notes Content
	TreatmentToCreateNote(source Treatment, ctx CreateContext) user.CreateNoteParams

	// goverter:useZeroValueOnPointerInconsistency
	// goverter:map  Identifier ID
	// goverter:map  Date OccurredAt
	// goverter:map  Notes Content
	// goverter:map UpdatedByIdentityName | AssignUpdatedByName
	// goverter:map UpdatedByIdentityID   | AssignUpdatedByID
	// goverter:context ctx
	TreatmentToUpdateNote(source Treatment, ctx UpdateContext) user.UpdateNoteParams

	////////
	// SETTINGS
	////////

	// goverter:useZeroValueOnPointerInconsistency
	// goverter:ignore Type
	// goverter:map ID Identifier
	// goverter:map CreatedAt SrvCreated
	// goverter:map UpdatedAt SrvModified
	DBModelToSetting(source *user.Setting) Setting
}

// //
// Custom Converter Methods
// //

// goverter:context ctx
func AssignID(date *DateInt, ctx CreateContext) string {
	return utils.NextDateID(ctx.IDPrefix, date.Int64())
}

// goverter:context ctx
func AssignCreatedByName(ctx CreateContext) string {
	return ctx.CreatedByIdentityName
}

// goverter:context ctx
func AssignCreatedByID(ctx CreateContext) int64 {
	return ctx.CreatedByIdentityID
}

// goverter:context ctx
func AssignUpdatedByName(ctx UpdateContext) *string {
	return ptr(ctx.UpdatedByIdentityName)
}

// goverter:context ctx
func AssignUpdatedByID(ctx UpdateContext) *int64 {
	return ptr(ctx.UpdatedByIdentityID)
}

func ToDateInt(createdAt int64) *DateInt {
	x := DateInt(createdAt)
	return &x
}

// ////
// GlucoseReading
// ////
func SGVGlucoseReadingTypeToValue(gr user.GlucoseReading) *FlexibleFloat64 {
	if gr.Type == "sgv" {
		return ptr(FlexibleFloat64(gr.Value))
	}
	return nil

}

func MBGGlucoseReadingTypeToValue(gr user.GlucoseReading) *FlexibleFloat64 {
	if gr.Type == "mbg" {
		return ptr(FlexibleFloat64(gr.Value))
	}
	return nil

}

func ModelDateStringFromDateIntAndUTCOffset(m interface{}) *DateString {
	if m == nil {
		return nil
	}

	// TODO this should be mapping to interfaces now we have getters
	switch v := m.(type) {
	case interface {
		GetOccurredAt() int64
		GetUtcOffsetMinutes() *int64
	}:
		ds := DateStringFromInt64AndUTCOffset(v.GetOccurredAt(), v.GetUtcOffsetMinutes())
		return &ds
	case interface {
		GetStart() int64
		GetUtcOffsetMinutes() *int64
	}:
		ds := DateStringFromInt64AndUTCOffset(v.GetStart(), v.GetUtcOffsetMinutes())
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
