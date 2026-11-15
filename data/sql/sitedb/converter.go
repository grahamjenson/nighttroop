package sitedb

// goverter:converter
// goverter:output:format function
// goverter:enum:unknown @ignore
// goverter:output:file ./z_converter.go
type ToCreateParams interface {

	// EVENTS
	ActivityToCreate(source Activity) CreateActivityParams
	ActivityToUpdate(source Activity) UpdateActivityParams

	FoodIntakeToCreate(source FoodIntake) CreateFoodIntakeParams
	FoodIntakeToUpdate(source FoodIntake) UpdateFoodIntakeParams

	GlucoseReadingToCreate(source GlucoseReading) CreateGlucoseReadingParams
	GlucoseReadingToUpdate(source GlucoseReading) UpdateGlucoseReadingParams

	InsulinDoseToCreate(source InsulinDose) CreateInsulinDoseParams
	InsulinDoseToUpdate(source InsulinDose) UpdateInsulinDoseParams

	NotificationEventToCreate(source NotificationEvent) CreateNotificationEventParams
	NotificationEventToUpdate(source NotificationEvent) UpdateNotificationEventParams

	ProfileSwitchToCreate(source ProfileSwitch) CreateProfileSwitchParams
	ProfileSwitchToUpdate(source ProfileSwitch) UpdateProfileSwitchParams

	TempBasalToCreate(source TempBasal) CreateTempBasalParams
	TempBasalToUpdate(source TempBasal) UpdateTempBasalParams

	TempTargetToCreate(source TempTarget) CreateTempTargetParams
	TempTargetToUpdate(source TempTarget) UpdateTempTargetParams

	DeviceReportToCreate(source DeviceReport) CreateDeviceReportParams

	// PRESETS
	FoodPresetToCreate(source FoodPreset) CreateFoodPresetParams
	FoodPresetToUpdate(source FoodPreset) UpdateFoodPresetParams

	ProfilePresetToCreate(source ProfilePreset) CreateProfilePresetParams
	ProfilePresetToUpdate(source ProfilePreset) UpdateProfilePresetParams

	// EVENT GROUP
	EventGroupToCreate(source EventGroup) CreateEventGroupParams

	SearchEventGroupToGetEventGroup(source SearchEventGroupsRow) GetEventGroupRow
	AllEventGroupToGetEventGroup(source AllEventGroupsRow) GetEventGroupRow

	//goverter:map . EventGroup
	//goverter:ignore GlucoseReading InsulinDose FoodIntake TempBasal TempTarget ProfileSwitch Activity NotificationEvent
	EventGroupToEventGroupsBundle(source EventGroup) EventGroupsBundle

	// EVENT GROUP EVENTS: each event its row joins, through the event's view, whose every column can be NULL, as the
	// event's own row. Copied field by field, rather than through JSON, as every treatment read makes them.
	// goverter:useZeroValueOnPointerInconsistency
	GlucoseReadingFromView(source GlucoseReadingsVw) GlucoseReading
	// goverter:useZeroValueOnPointerInconsistency
	InsulinDoseFromView(source InsulinDosesVw) InsulinDose
	// goverter:useZeroValueOnPointerInconsistency
	FoodIntakeFromView(source FoodIntakesVw) FoodIntake
	// goverter:useZeroValueOnPointerInconsistency
	TempBasalFromView(source TempBasalsVw) TempBasal
	// goverter:useZeroValueOnPointerInconsistency
	TempTargetFromView(source TempTargetsVw) TempTarget
	// goverter:useZeroValueOnPointerInconsistency
	ProfileSwitchFromView(source ProfileSwitchesVw) ProfileSwitch
	// goverter:useZeroValueOnPointerInconsistency
	ActivityFromView(source ActivitiesVw) Activity
	// goverter:useZeroValueOnPointerInconsistency
	NotificationEventFromView(source NotificationEventsVw) NotificationEvent
}
