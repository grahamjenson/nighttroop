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

	NoteToCreate(source Note) CreateNoteParams
	NoteToUpdate(source Note) UpdateNoteParams

	ProfileSwitchToCreate(source ProfileSwitch) CreateProfileSwitchParams
	ProfileSwitchToUpdate(source ProfileSwitch) UpdateProfileSwitchParams

	TempBasalToCreate(source TempBasal) CreateTempBasalParams
	TempBasalToUpdate(source TempBasal) UpdateTempBasalParams

	TempTargetToCreate(source TempTarget) CreateTempTargetParams
	TempTargetToUpdate(source TempTarget) UpdateTempTargetParams

	CGMStateToCreate(source CgmState) CreateCGMStateParams
	CGMStateToUpdate(source CgmState) UpdateCGMStateParams

	DeviceStateToCreate(source DeviceState) CreateDeviceStateParams
	DeviceStateToUpdate(source DeviceState) UpdateDeviceStateParams

	LoopStateToCreate(source LoopState) CreateLoopStateParams
	LoopStateToUpdate(source LoopState) UpdateLoopStateParams

	PumpStateToCreate(source PumpState) CreatePumpStateParams
	PumpStateToUpdate(source PumpState) UpdatePumpStateParams

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
	//goverter:ignore GlucoseReading InsulinDose FoodIntake TempBasal TempTarget ProfileSwitch DeviceState LoopState PumpState CgmState Activity Note
	EventGroupToEventGroupsBundle(source EventGroup) EventGroupsBundle
}
