package site

// goverter:converter
// goverter:output:format function
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

	ComponentEventToCreate(source ComponentEvent) CreateComponentEventParams
	ComponentEventToUpdate(source ComponentEvent) UpdateComponentEventParams

	DeviceReportToCreate(source DeviceReport) CreateDeviceReportParams
	DeviceReportToUpdate(source DeviceReport) UpdateDeviceReportParams

	// PRESETS
	FoodPresetToCreate(source FoodPreset) CreateFoodPresetParams
	FoodPresetToUpdate(source FoodPreset) UpdateFoodPresetParams

	ProfilePresetToCreate(source ProfilePreset) CreateProfilePresetParams
	ProfilePresetToUpdate(source ProfilePreset) UpdateProfilePresetParams

	// SITE SETTINGS
	SiteSettingsToUpdate(source SiteSetting) UpdateSiteSettingsParams

	// EVENT GROUP
	EventGroupToCreate(source EventGroup) CreateEventGroupParams
	EventGroupToUpdate(source EventGroup) UpdateEventGroupParams

	SearchEventGroupToGetEventGroup(source SearchEventGroupsRow) GetEventGroupRow

	//goverter:map . EventGroup
	//goverter:ignore GlucoseReading InsulinDose FoodIntake TempBasal TempTarget ProfileSwitch ComponentEvent Activity NotificationEvent
	EventGroupToEventGroupsBundle(source EventGroup) EventGroupsBundle
}
