package user

// goverter:converter
//
// goverter:output:file ./z_converter.go
type ToCreateParams interface {
	Activity(source Activity) CreateActivityParams

	Bolus(source Bolus) CreateBolusParams
	Carbs(source Carb) CreateCarbParams

	DeviceStatus(source DeviceStatus) CreateDeviceStatusParams
	Food(source Food) CreateFoodParams

	GlucoseReading(source GlucoseReading) CreateGlucoseReadingParams
	Note(source Note) CreateNoteParams

	Profile(source Profile) CreateProfileParams
	TempBasal(source TempBasal) CreateTempBasalParams
	TempTarget(source TempTarget) CreateTempTargetParams
}
