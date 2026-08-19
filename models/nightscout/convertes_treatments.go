package nightscout

import (
	"encoding/json"

	"nighttroop.t1d.nz/data/sql/site"
)

// goverter:converter
// goverter:output:format function
// goverter:output:file ./z_treatments_converters.go
// goverter:useZeroValueOnPointerInconsistency
type TreatmentsEvenetConverter interface {

	///////
	// CREATE
	///////

	// goverter:context ctx
	// goverter:map CreatedByIdentityName  | AssignCreatedByName
	// goverter:map ID                     | AssignID
	EventGroupCreate(source site.EventGroup, ctx ConverterContext) site.CreateEventGroupParams

	///////
	// Food Intake
	///////

	// goverter:context ctx
	// goverter:map CreatedByIdentityName  | AssignCreatedByName
	// goverter:map ID                     | AssignID
	FoodIntakeCreate(source site.FoodIntake, ctx ConverterContext) site.CreateFoodIntakeParams

	///////
	// Glucose Reading
	///////

	// goverter:context ctx
	// goverter:map CreatedByIdentityName  | AssignCreatedByName
	// goverter:map ID                     | AssignID
	GlucoseReadingCreate(source site.GlucoseReading, ctx ConverterContext) site.CreateGlucoseReadingParams

	///////
	// Insulin Dose
	///////

	// goverter:context ctx
	// goverter:map CreatedByIdentityName  | AssignCreatedByName
	// goverter:map ID                     | AssignID
	InsulinDoseCreate(source site.InsulinDose, ctx ConverterContext) site.CreateInsulinDoseParams

	///////
	// Notes
	///////

	// goverter:context ctx
	// goverter:map CreatedByIdentityName  | AssignCreatedByName
	// goverter:map ID                     | AssignID
	NotificationEventCreate(source site.NotificationEvent, ctx ConverterContext) site.CreateNotificationEventParams

	///////
	// Profile Switch
	///////

	// goverter:context ctx
	// goverter:map CreatedByIdentityName  | AssignCreatedByName
	// goverter:map ID                     | AssignID
	ProfileSwitchCreate(source site.ProfileSwitch, ctx ConverterContext) site.CreateProfileSwitchParams

	///////
	// Temp Basal
	///////

	// goverter:context ctx
	// goverter:map CreatedByIdentityName  | AssignCreatedByName
	// goverter:map ID                     | AssignID
	TempBasalCreate(source site.TempBasal, ctx ConverterContext) site.CreateTempBasalParams

	///////
	// Temp Targets
	///////

	// goverter:context ctx
	// goverter:map CreatedByIdentityName  | AssignCreatedByName
	// goverter:map ID                     | AssignID
	TempTargetCreate(source site.TempTarget, ctx ConverterContext) site.CreateTempTargetParams

	// goverter:context ctx
	// goverter:map CreatedByIdentityName  | AssignCreatedByName
	// goverter:map ID                     | AssignID
	ComponentEventCreate(source site.ComponentEvent, ctx ConverterContext) site.CreateComponentEventParams

	// goverter:context ctx
	// goverter:map CreatedByIdentityName  | AssignCreatedByName
	// goverter:map ID                     | AssignID
	ActivityCreate(source site.Activity, ctx ConverterContext) site.CreateActivityParams

	///////
	// Update
	///////

	// goverter:context ctx
	// goverter:map UpdatedByIdentityName | AssignUpdatedByName
	EventGroupUpdate(source site.EventGroup, ctx ConverterContext) site.UpdateEventGroupParams

	///////
	// Food Intake
	///////

	// goverter:context ctx
	// goverter:map UpdatedByIdentityName | AssignUpdatedByName
	FoodIntakeUpdate(source site.FoodIntake, ctx ConverterContext) site.UpdateFoodIntakeParams

	///////
	// Glucose Reading
	///////

	// goverter:context ctx
	// goverter:map UpdatedByIdentityName | AssignUpdatedByName
	GlucoseReadingUpdate(source site.GlucoseReading, ctx ConverterContext) site.UpdateGlucoseReadingParams

	///////
	// Insulin Dose
	///////

	// goverter:context ctx
	// goverter:map UpdatedByIdentityName | AssignUpdatedByName
	InsulinDoseUpdate(source site.InsulinDose, ctx ConverterContext) site.UpdateInsulinDoseParams

	///////
	// Notes
	///////

	// goverter:context ctx
	// goverter:map UpdatedByIdentityName | AssignUpdatedByName
	NotificationEventUpdate(source site.NotificationEvent, ctx ConverterContext) site.UpdateNotificationEventParams

	///////
	// Profile
	///////

	// goverter:context ctx
	// goverter:map UpdatedByIdentityName | AssignUpdatedByName
	ProfileUpdate(source site.ProfilePreset, ctx ConverterContext) site.UpdateProfilePresetParams

	///////
	// Profile Switch
	///////

	// goverter:context ctx
	// goverter:map UpdatedByIdentityName | AssignUpdatedByName
	ProfileSwitchUpdate(source site.ProfileSwitch, ctx ConverterContext) site.UpdateProfileSwitchParams

	///////
	// Temp Basal
	///////

	// goverter:context ctx
	// goverter:map UpdatedByIdentityName | AssignUpdatedByName
	TempBasalUpdate(source site.TempBasal, ctx ConverterContext) site.UpdateTempBasalParams

	///////
	// Temp Targets
	///////

	// goverter:context ctx
	// goverter:map UpdatedByIdentityName | AssignUpdatedByName
	TempTargetUpdate(source site.TempTarget, ctx ConverterContext) site.UpdateTempTargetParams

	// goverter:context ctx
	// goverter:map UpdatedByIdentityName | AssignUpdatedByName
	ComponentEventUpdate(source site.ComponentEvent, ctx ConverterContext) site.UpdateComponentEventParams

	// goverter:context ctx
	// goverter:map UpdatedByIdentityName | AssignUpdatedByName
	ActivityUpdate(source site.Activity, ctx ConverterContext) site.UpdateActivityParams

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
	// goverter:map EventGroup.Notes Notes
	// goverter:map EventGroup.Reason Reason
	// goverter:map EventGroup.GroupType EventType
	//
	// goverter:ignore PreBolus SplitNow SplitExt Timeshift OriginalProfileName
	//
	//
	// goverter:map GlucoseReading.Reading Glucose
	// goverter:map GlucoseReading.ReadingType GlucoseType
	//
	// goverter:map InsulinDose.Dose Insulin
	// goverter:map InsulinDose.DoseType Type
	//
	// goverter:map FoodIntake.Category Category
	// goverter:map FoodIntake.Subcategory Subcategory
	// goverter:map FoodIntake.Portions Portions
	// goverter:map FoodIntake.Carbs Carbs
	// goverter:map FoodIntake.Protein Protein
	// goverter:map FoodIntake.Fat Fat
	// goverter:map FoodIntake.Fibre Fibre
	// goverter:map FoodIntake.Energy Energy
	//
	// goverter:map TempBasal.Absolute Absolute
	// goverter:map TempBasal.Absolute Rate
	//
	// goverter:map TempTarget.HighTarget TargetTop
	// goverter:map TempTarget.LowTarget TargetBottom
	//
	// goverter:map ProfileSwitch.Name Name
	// goverter:map ProfileSwitch.Name Profile
	// goverter:map ProfileSwitch.Timeshift Timeshift
	// goverter:map ProfileSwitch.Units Units
	// goverter:map . ProfileJSON | ToProfileJSON
	//
	// goverter:map Activity.ActivityType ActivityType
	// goverter:map Activity.Steps Steps
	// goverter:map Activity.Calories Calories
	// goverter:map Activity.HeartRate HeartRate
	//
	// goverter:map ComponentEvent.PumpID PumpID
	// goverter:map ComponentEvent.PumpType PumpType
	// goverter:map ComponentEvent.PumpSerial PumpSerial
	//
	// SHARED FIELDS
	// goverter:map . DurationInMilliseconds     | MinutesToMilliseconds
	// goverter:map . Duration                   | DurationSelect
	// goverter:map . Percentage                 | PercentageSelect
	DBModelToTreatment(source *site.EventGroupsBundle) Treatment

	//////
	// Update Override
	//////

	// goverter:update target
	// goverter:ignore ID
	UpdateToEventGroup(source site.EventGroup, target *site.EventGroup)

	// goverter:update target
	// goverter:ignore ID
	UpdateToFoodIntake(source *site.FoodIntake, target *site.FoodIntake)

	// goverter:update target
	// goverter:ignore ID
	UpdateToGlucoseReading(source *site.GlucoseReading, target *site.GlucoseReading)

	// goverter:update target
	// goverter:ignore ID
	UpdateToInsulinDose(source *site.InsulinDose, target *site.InsulinDose)

	// goverter:update target
	// goverter:ignore ID
	UpdateToNotificationEvent(source *site.NotificationEvent, target *site.NotificationEvent)

	// goverter:update target
	// goverter:ignore ID
	UpdateToProfileSwitch(source *site.ProfileSwitch, target *site.ProfileSwitch)

	// goverter:update target
	// goverter:ignore ID
	UpdateToTempBasal(source *site.TempBasal, target *site.TempBasal)

	// goverter:update target
	// goverter:ignore ID
	UpdateToTempTarget(source *site.TempTarget, target *site.TempTarget)

	// goverter:update target
	// goverter:ignore ID
	UpdateToComponentEvent(source *site.ComponentEvent, target *site.ComponentEvent)

	// goverter:update target
	// goverter:ignore ID
	UpdateToActivity(source *site.Activity, target *site.Activity)
}

////////////////////////////////////////
// TREATMENT TO CREATE/UPDATE
////////////////////////////////////////

func TreatmentToEventGroup(source Treatment) site.EventGroup {
	target := site.EventGroup{}
	SharedTreatmentConvertion(source, &target)

	target.GroupType = source.EventType.String()

	return target
}

///////
// Food Intake
///////

func TreatmentToFoodIntake(source Treatment) *site.FoodIntake {
	assigned := false
	target := site.FoodIntake{}
	SharedTreatmentConvertion(source, &target)

	if source.Name != nil {
		target.Name = source.Name.Stringp()
	}
	if source.Category != nil {
		target.Category = source.Category.Stringp()
	}
	if source.Subcategory != nil {
		target.Subcategory = source.Subcategory.Stringp()
	}
	if source.Carbs != nil {
		assigned = true
		target.Carbs = source.Carbs.Float64p()
	}
	if source.Fat != nil {
		assigned = true
		target.Fat = source.Fat.Float64p()
	}
	if source.Protein != nil {
		assigned = true
		target.Protein = source.Protein.Float64p()
	}
	if source.Fibre != nil {
		assigned = true
		target.Fibre = source.Fibre.Float64p()
	}
	if source.Energy != nil {
		assigned = true
		target.Energy = source.Energy.Float64p()
	}
	if source.Portions != nil {
		assigned = true
		target.Portions = source.Portions.Float64p()
	}

	if assigned {
		return &target
	}
	return nil
}

///////
// Glucose Reading
///////

func TreatmentToGlucoseReading(source Treatment) *site.GlucoseReading {
	assigned := false
	target := site.GlucoseReading{ReadingType: "sgv"}
	SharedTreatmentConvertion(source, &target)

	if source.Glucose != nil {
		assigned = true
		target.Reading = source.Glucose.Float64()
	}

	if source.GlucoseType != nil {
		switch *source.GlucoseType {
		case Sensor:
			target.ReadingType = "sgv"
		case Finger:
			target.ReadingType = "mbg"
		case Meter:
			target.ReadingType = "mbg"
		}
	}
	if assigned {
		return &target
	}
	return nil
}

///////
// Insulin Dose
///////

func TreatmentToInsulinDose(source Treatment) *site.InsulinDose {
	assigned := false
	target := site.InsulinDose{}
	SharedTreatmentConvertion(source, &target)

	if source.Insulin != nil {
		assigned = true
		target.Dose = source.Insulin.Float64()
	}

	target.DoseType = source.Type.Stringp()

	if assigned {
		return &target
	}
	return nil
}

///////
// NotificationEvents
///////

func TreatmentToNotificationEvent(source Treatment) *site.NotificationEvent {
	assigned := false
	target := site.NotificationEvent{
		Severity:         0,
		NotificationType: "announcement",
		Title:            "Treatment Event",
	}
	SharedTreatmentConvertion(source, &target)

	// Always is assigned
	assigned = true

	if source.Name != nil {
		target.Title = source.Name.String()
	}
	if source.Notes != nil {
		target.Content = source.Notes.Stringp()
	}

	if assigned {
		return &target
	}
	return nil
}

///////
// Profile Switch
///////

func TreatmentToProfileSwitch(source Treatment) *site.ProfileSwitch {
	assigned := false
	target := site.ProfileSwitch{}
	SharedTreatmentConvertion(source, &target)

	// TODO: This is the bit that will need more work, with the json of profile
	if source.Name != nil {
		assigned = true
		target.Name = source.Name.String()
	} else if source.Profile != nil {
		assigned = true
		target.Name = source.Profile.String()
	}

	if source.Percentage != nil {
		assigned = true
		target.Percent = source.Percentage.Float64p()
	}
	if source.Duration != nil {
		assigned = true
		target.DurationMinutes = source.Duration.Int64()
	}
	if source.Timeshift != nil {
		assigned = true
		target.Timeshift = source.Timeshift.Int64p()
	}

	if source.ProfileJSON != nil {
		var profileJSON Profile
		err := json.Unmarshal([]byte(*source.ProfileJSON), &profileJSON)
		if err == nil {
			assigned = true
			target.Dia = profileJSON.Dia.Float64p()
			target.Timezone = profileJSON.Timezone.Stringp()
			target.Units = profileJSON.Units.Stringp()

			basalBytes, _ := json.Marshal(profileJSON.Basal)
			carbRatioBytes, _ := json.Marshal(profileJSON.CarbRatio)
			sensBytes, _ := json.Marshal(profileJSON.Sens)
			targetLowBytes, _ := json.Marshal(profileJSON.TargetLow)
			targetHighBytes, _ := json.Marshal(profileJSON.TargetHigh)

			if basalBytes != nil {
				target.BasalRatesJson = ptr(string(basalBytes))
			} else {
				target.BasalRatesJson = ptr("[]")
			}

			if carbRatioBytes != nil {
				target.CarbRatiosJson = ptr(string(carbRatioBytes))
			} else {
				target.CarbRatiosJson = ptr("[]")
			}

			if sensBytes != nil {
				target.InsulinSensitivitiesJson = ptr(string(sensBytes))
			} else {
				target.InsulinSensitivitiesJson = ptr("[]")
			}

			if targetLowBytes != nil {
				target.TargetsLowJson = ptr(string(targetLowBytes))
			} else {
				target.TargetsLowJson = ptr("[]")
			}

			if targetHighBytes != nil {
				target.TargetsHighJson = ptr(string(targetHighBytes))
			} else {
				target.TargetsHighJson = ptr("[]")
			}
		}

	}
	if assigned {
		return &target
	}
	return nil
}

///////
// Temp Basal
///////

func TreatmentToTempBasal(source Treatment) *site.TempBasal {
	assigned := false
	target := site.TempBasal{}
	SharedTreatmentConvertion(source, &target)

	if source.Absolute != nil {
		assigned = true
		target.Absolute = source.Absolute.Float64()
	} else if source.Rate != nil {
		assigned = true
		target.Absolute = source.Rate.Float64()
	}

	if source.Percentage != nil {
		target.Percent = source.Percentage.Float64p()
	}

	target.DurationMinutes = source.Duration.Int64()

	if assigned {
		return &target
	}
	return nil
}

///////
// Temp Targets
///////

func TreatmentToTempTarget(source Treatment) *site.TempTarget {
	assigned := false
	target := site.TempTarget{}
	SharedTreatmentConvertion(source, &target)

	if source.TargetTop != nil {
		assigned = true
		target.HighTarget = source.TargetTop.Float64()
	} else if source.TargetBottom != nil {
		// If Target Top is missing then we assume both targets are the same
		assigned = true
		target.HighTarget = source.TargetBottom.Float64()
	}

	if source.TargetBottom != nil {
		assigned = true
		target.LowTarget = source.TargetBottom.Float64()
	} else if source.TargetTop != nil {
		// If Target Bottom is missing then we assume both targets are the same
		assigned = true
		target.LowTarget = source.TargetTop.Float64()
	}

	target.DurationMinutes = source.Duration.Int64()

	if assigned {
		return &target
	}
	return nil
}

func TreatmentToComponentEvent(source Treatment) *site.ComponentEvent {
	assigned := false
	target := site.ComponentEvent{EventType: source.EventType.String()}
	SharedTreatmentConvertion(source, &target)

	if source.PumpID != nil {
		assigned = true
		target.PumpID = source.PumpID.Stringp()
	}
	if source.PumpSerial != nil {
		assigned = true
		target.PumpSerial = source.PumpSerial.Stringp()
	}
	if source.PumpType != nil {
		assigned = true
		target.PumpType = source.PumpType.Stringp()
	}

	if assigned {
		return &target
	}
	return nil
}

func TreatmentToActivity(source Treatment) *site.Activity {
	assigned := false
	target := site.Activity{}
	SharedTreatmentConvertion(source, &target)

	if source.Name != nil {
		assigned = true
		target.Name = source.Name.Stringp()
	}

	if source.ActivityType != nil {
		assigned = true
		target.ActivityType = source.ActivityType.String()
	}

	if source.Calories != nil {
		assigned = true
		target.Calories = source.Calories.Int64p()
	}

	if source.Steps != nil {
		assigned = true
		target.Steps = source.Steps.Int64p()
	}

	if source.HeartRate != nil {
		assigned = true
		target.HeartRate = source.HeartRate.Int64p()
	}

	if assigned {
		return &target
	}
	return nil
}

func SharedTreatmentConvertion(source Treatment, target interface {
	SetOccurredAt(int64)
	SetCreatedByAppName(*string)
	SetCreatedByDeviceName(*string)
	SetNotes(*string)
	SetReason(*string)
	SetRawJsonData(*string)
	SetUtcOffsetMinutes(*int64)
}) {
	target.SetOccurredAt(OrNow(source.Date))
	target.SetCreatedByAppName(source.App.Stringp())
	target.SetCreatedByDeviceName(source.Device.Stringp())
	target.SetNotes(source.Notes.Stringp())
	target.SetReason(source.Reason.Stringp())
	target.SetRawJsonData(&source.RawJsonData)
	target.SetUtcOffsetMinutes(source.UtcOffsetMinutes.Int64p())

	if value, ok := target.(interface {
		SetDurationMinutes(*int64)
	}); ok {
		value.SetDurationMinutes(source.Duration.Int64p())
	}
}

func MinutesToMilliseconds(source *site.EventGroupsBundle) *FlexibleInt64 {

	if source.TempBasal != nil {
		value := FlexibleInt64(source.TempBasal.DurationMinutes * 60 * 1000)
		return &value
	}
	if source.TempTarget != nil {
		value := FlexibleInt64(source.TempTarget.DurationMinutes * 60 * 1000)
		return &value

	}

	if source.ProfileSwitch != nil {
		value := FlexibleInt64(source.ProfileSwitch.DurationMinutes * 60 * 1000)
		return &value
	}

	if source.Activity != nil && source.Activity.DurationMinutes != nil {
		value := FlexibleInt64(*source.Activity.DurationMinutes * 60 * 1000)
		return &value
	}

	return nil
}

func DurationSelect(source *site.EventGroupsBundle) *FlexibleInt64 {
	if source.TempBasal != nil {
		value := FlexibleInt64(source.TempBasal.DurationMinutes)
		return &value
	}

	if source.TempTarget != nil {
		value := FlexibleInt64(source.TempTarget.DurationMinutes)
		return &value

	}

	if source.ProfileSwitch != nil {
		value := FlexibleInt64(source.ProfileSwitch.DurationMinutes)
		return &value
	}

	if source.Activity != nil && source.Activity.DurationMinutes != nil {
		value := FlexibleInt64(*source.Activity.DurationMinutes)
		return &value
	}

	return nil
}

func PercentageSelect(source *site.EventGroupsBundle) *FlexibleFloat64 {
	if source.TempBasal != nil {
		if source.TempBasal.Percent != nil {
			value := FlexibleFloat64(*source.TempBasal.Percent)
			return &value
		}
	}

	if source.ProfileSwitch != nil {
		if source.ProfileSwitch.Percent != nil {
			value := FlexibleFloat64(*source.ProfileSwitch.Percent)
			return &value
		}
	}

	return nil
}

func ToProfileJSON(source *site.EventGroupsBundle) *LongJSONString {
	if source == nil || source.ProfileSwitch == nil {
		return nil
	}
	profile := Profile{}
	profileSwitch := *source.ProfileSwitch
	if profileSwitch.Dia != nil {
		profile.Dia = ptr(FlexibleFloat64(*profileSwitch.Dia))
	}
	if profileSwitch.Timezone != nil {
		profile.Timezone = ptr(Timezone(*profileSwitch.Timezone))
	}
	if profileSwitch.Units != nil {
		profile.Units = ptr(Units(*profileSwitch.Units))
	}
	if profileSwitch.CarbRatiosJson != nil {
		json.Unmarshal([]byte(*profileSwitch.CarbRatiosJson), &profile.CarbRatio)
	}
	if profileSwitch.InsulinSensitivitiesJson != nil {
		json.Unmarshal([]byte(*profileSwitch.InsulinSensitivitiesJson), &profile.Sens)
	}
	if profileSwitch.BasalRatesJson != nil {
		json.Unmarshal([]byte(*profileSwitch.BasalRatesJson), &profile.Basal)
	}
	if profileSwitch.TargetsLowJson != nil {
		json.Unmarshal([]byte(*profileSwitch.TargetsLowJson), &profile.TargetLow)
	}
	if profileSwitch.TargetsHighJson != nil {
		json.Unmarshal([]byte(*profileSwitch.TargetsHighJson), &profile.TargetHigh)
	}

	profileBytes, err := json.Marshal(profile)
	if err != nil {
		return nil
	}
	profileJSON := LongJSONString(string(profileBytes))
	return &profileJSON
}
