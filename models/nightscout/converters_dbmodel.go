package nightscout

import (
	"encoding/json"
	"fmt"

	"nighttroop.t1d.nz/data/sql/site"
)

// goverter:converter
// goverter:output:format function
// goverter:output:file ./z_dbmodel_converters.go
// goverter:useZeroValueOnPointerInconsistency
type DBModelConverter interface {

	///////
	// ACTIVITIES
	///////

	// goverter:map . CreatedAt | ModelDateStringFromDateIntAndUTCOffset
	// goverter:map CreatedByIdentityName Subject
	// goverter:map CreatedByAppName App
	// goverter:map CreatedByDeviceName Device
	// goverter:map ID Identifier
	// goverter:map OccurredAt Date
	//
	// goverter:map DurationMinutes Duration
	DBModelToActivity(source *site.Activity) Activity

	///////
	// ENTRIES TO GLUCOSE READINGS
	///////

	// goverter:ignore Direction Trend Units Filtered Unfiltered RSSI Noise SysTime
	// goverter:map CreatedByIdentityName Subject
	// goverter:map CreatedByAppName App
	// goverter:map CreatedByDeviceName Device
	// goverter:map ID Identifier
	// goverter:map . SGV | SGVGlucoseReadingTypeToValue
	// goverter:map . MBG | MBGGlucoseReadingTypeToValue
	// goverter:map OccurredAt Date
	// goverter:map OccurredAt Mills
	// goverter:map . DateString | ModelDateStringFromDateIntAndUTCOffset
	// goverter:map ReadingType Type
	DBModelToEntry(source *site.GlucoseReading) Entry

	///////
	// DEVICE REPORT TO DEVICE STATUS
	///////

	// goverter:ignore Uploader Loop Pump OpenAPS MergedLoopStatus
	// goverter:map CreatedByIdentityName Subject
	// goverter:map CreatedByAppName App
	// goverter:map CreatedByDeviceName Device
	// goverter:map ID Identifier
	// goverter:map OccurredAt Date
	// goverter:map DeviceBatteryPercent UploaderBattery
	// goverter:map . CreatedAt | ModelDateStringFromDateIntAndUTCOffset
	DBModelToDeviceStatus(source *site.DeviceReport) DeviceStatus

	///////
	// FOOD TO FOOD PRESET
	///////

	// goverter:map CreatedByIdentityName Subject
	// goverter:map CreatedByAppName App
	// goverter:map CreatedByDeviceName Device
	// goverter:map ID Identifier
	//
	// goverter:ignore Food Category Subcategory Portion Unit GI Portions HideAfterUse Hidden Position
	// goverter:map CreatedAt Date
	// goverter:map Calories Energy
	DBModelToFood(source *site.FoodPreset) Food

	///////
	// Profile To Profile Preset
	///////

	// goverter:map . StartDate | ModelDateStringFromDateIntAndUTCOffset
	// goverter:map CreatedByIdentityName Subject
	// goverter:map CreatedByAppName App
	// goverter:map CreatedByDeviceName Device
	// goverter:map ID Identifier
	// goverter:map OccurredAt Date
	//
	// goverter:ignore Store
	// goverter:map OccurredAt Mills
	// goverter:map Name DefaultProfile
	DBModelToProfileStore(source *site.ProfilePreset) ProfileStore

	// goverter:map CarbRatiosJson           CarbRatio  | JSONToScheduleValueSlice
	// goverter:map InsulinSensitivitiesJson Sens       | JSONToScheduleValueSlice
	// goverter:map BasalRatesJson           Basal      | JSONToScheduleValueSlice
	// goverter:map TargetsHighJson          TargetHigh | JSONToScheduleValueSlice
	// goverter:map TargetsLowJson           TargetLow  | JSONToScheduleValueSlice
	DBModelToProfile(source *site.ProfilePreset) Profile

	////////////////////////////////////////
	// TREATMENT TO CREATE/UPDATE
	////////////////////////////////////////

	///////
	// EventGroup
	///////

	// goverter:map . CreatedAt | ModelDateStringFromDateIntAndUTCOffset
	// goverter:map EventGroup.UtcOffsetMinutes UtcOffsetMinutes
	// goverter:map EventGroup.RawJsonData RawJsonData
	// goverter:map EventGroup.CreatedByIdentityName Subject
	// goverter:map EventGroup.CreatedByIdentityName EnteredBy
	// goverter:map EventGroup.CreatedByAppName App
	// goverter:map EventGroup.CreatedByDeviceName Device
	// goverter:map EventGroup.ID Identifier
	// goverter:map EventGroup.OccurredAt Date
	//
	// goverter:ignore PreBolus SplitNow SplitExt Profile Units PumpType ProfileJSON Timeshift
	// goverter:map EventGroup.GroupType EventType
	//
	// goverter:map GlucoseReading.Reading Glucose
	// goverter:map GlucoseReading.ReadingType GlucoseType
	//
	// goverter:map InsulinDose.Dose Insulin
	//
	// goverter:map FoodIntake.Carbs Carbs
	// goverter:map FoodIntake.Protein Protein
	// goverter:map FoodIntake.Fat Fat
	// goverter:map FoodIntake.Fibre Fibre
	// goverter:map FoodIntake.Calories Energy
	//
	// goverter:map TempBasal.DurationMinutes Duration
	// goverter:map TempBasal.Percent Rate
	// goverter:map TempBasal.Absolute Absolute
	//
	// goverter:map TempTarget.DurationMinutes Duration
	// goverter:map TempTarget.HighTarget TargetTop
	// goverter:map TempTarget.LowTarget TargetBottom
	//
	// goverter:map ProfileSwitch.Name Name
	// goverter:map ProfileSwitch.Name Profile
	// goverter:map ProfileSwitch.Name OriginalProfileName
	// goverter:map ProfileSwitch.Percent Percentage
	//
	// goverter:map Activity.ActivityType ActivityType
	// goverter:map Activity.Steps Steps
	// goverter:map Activity.Calories Calories
	// goverter:map Activity.DurationMinutes Duration
	// goverter:map Activity.HeartRate HeartRate
	//
	// goverter:map Note.Content Notes
	// goverter:map Note.Reason Reason
	DBModelToTreatment(source *site.EventGroupsBundle) Treatment
}

// //
// Custom Converter Methods
// //

func ModelDateStringFromDateIntAndUTCOffset(m interface{}) *DateString {
	fmt.Printf("ModelDateStringFromDateIntAndUTCOff`set called with: %v:%T\n", m, m)
	if m == nil {
		return nil
	}

	switch v := m.(type) {
	case interface {
		GetOccurredAt() int64
		GetUtcOffsetMinutes() *int64
	}:
		ds := DateStringFromInt64AndUTCOffset(v.GetOccurredAt(), v.GetUtcOffsetMinutes())
		fmt.Printf("ModelDateStringFromDateIntAndUTCOffset returning: %v\n", ds)
		return &ds
	case interface {
		GetStart() int64
		GetUtcOffsetMinutes() *int64
	}:
		ds := DateStringFromInt64AndUTCOffset(v.GetStart(), v.GetUtcOffsetMinutes())
		fmt.Printf("ModelDateStringFromDateIntAndUTCOffset returning: %v\n", ds)
		return &ds
	}

	fmt.Printf("ModelDateStringFromDateIntAndUTCOffset returning: nil\n")
	return nil
}

func SGVGlucoseReadingTypeToValue(gr site.GlucoseReading) *FlexibleFloat64 {
	if gr.ReadingType == "sgv" {
		return ptr(FlexibleFloat64(gr.Reading))
	}
	return nil

}

func MBGGlucoseReadingTypeToValue(gr site.GlucoseReading) *FlexibleFloat64 {
	if gr.ReadingType == "mbg" {
		return ptr(FlexibleFloat64(gr.Reading))
	}
	return nil

}

func JSONToScheduleValueSlice(jsonStr *string) []*ScheduleValue {
	if jsonStr == nil {
		return nil
	}
	var result []*ScheduleValue
	err := json.Unmarshal([]byte(*jsonStr), &result)
	if err != nil {
		return nil
	}
	for _, sv := range result {
		if sv == nil {
			continue
		}
		if sv.TimeAsSeconds == nil && sv.Time != nil {
			sv.TimeAsSeconds = sv.Time.ToSeconds()
		}
	}
	return result
}
