package site

// goverter:converter
// goverter:output:format function
// goverter:output:file ./z_converter.go
type ToCreateParams interface {
	ActivityToCreate(source Activity) CreateActivityParams

	BolusToCreate(source Bolus) CreateBolusParams

	CarbsToCreate(source Carb) CreateCarbParams

	DeviceStatusToCreate(source DeviceStatus) CreateDeviceStatusParams

	FoodToCreate(source Food) CreateFoodParams

	NoteToCreate(source Note) CreateNoteParams

	ProfileToCreate(source Profile) CreateProfileParams

	TempBasalToCreate(source TempBasal) CreateTempBasalParams

	TempTargetToCreate(source TempTarget) CreateTempTargetParams

	//////
	// GLUCOSE READINGS
	//////

	GlucoseReadingToCreate(source GlucoseReading) CreateGlucoseReadingParams
	GlucoseReadingToUpdate(source GlucoseReading) UpdateGlucoseReadingParams
}
