
package site


// GetID returns the ID value.
func (m *Activity) GetID() string {
	return m.ID
}

// GetName returns the Name value.
func (m *Activity) GetName() *string {
	return m.Name
}

// GetNotes returns the Notes value.
func (m *Activity) GetNotes() *string {
	return m.Notes
}

// GetActivityType returns the ActivityType value.
func (m *Activity) GetActivityType() *string {
	return m.ActivityType
}

// GetDurationMinutes returns the DurationMinutes value.
func (m *Activity) GetDurationMinutes() *int64 {
	return m.DurationMinutes
}

// GetCalories returns the Calories value.
func (m *Activity) GetCalories() *int64 {
	return m.Calories
}

// GetSteps returns the Steps value.
func (m *Activity) GetSteps() *int64 {
	return m.Steps
}

// GetHeartRate returns the HeartRate value.
func (m *Activity) GetHeartRate() *int64 {
	return m.HeartRate
}

// GetRawJsonData returns the RawJsonData value.
func (m *Activity) GetRawJsonData() *string {
	return m.RawJsonData
}

// GetDeletedAt returns the DeletedAt value.
func (m *Activity) GetDeletedAt() *int64 {
	return m.DeletedAt
}

// GetUtcOffsetMinutes returns the UtcOffsetMinutes value.
func (m *Activity) GetUtcOffsetMinutes() *int64 {
	return m.UtcOffsetMinutes
}

// GetOccurredAt returns the OccurredAt value.
func (m *Activity) GetOccurredAt() int64 {
	return m.OccurredAt
}

// GetCreatedAt returns the CreatedAt value.
func (m *Activity) GetCreatedAt() int64 {
	return m.CreatedAt
}

// GetUpdatedAt returns the UpdatedAt value.
func (m *Activity) GetUpdatedAt() int64 {
	return m.UpdatedAt
}

// GetCreatedByAppName returns the CreatedByAppName value.
func (m *Activity) GetCreatedByAppName() string {
	return m.CreatedByAppName
}

// GetCreatedByDeviceName returns the CreatedByDeviceName value.
func (m *Activity) GetCreatedByDeviceName() string {
	return m.CreatedByDeviceName
}

// GetUpdatedByIdentityName returns the UpdatedByIdentityName value.
func (m *Activity) GetUpdatedByIdentityName() *string {
	return m.UpdatedByIdentityName
}

// GetCreatedByIdentityName returns the CreatedByIdentityName value.
func (m *Activity) GetCreatedByIdentityName() string {
	return m.CreatedByIdentityName
}

// GetID returns the ID value.
func (m *CgmState) GetID() string {
	return m.ID
}

// GetCgmState returns the CgmState value.
func (m *CgmState) GetCgmState() *string {
	return m.CgmState
}

// GetCgmTimestamp returns the CgmTimestamp value.
func (m *CgmState) GetCgmTimestamp() *int64 {
	return m.CgmTimestamp
}

// GetCgmInsertionTime returns the CgmInsertionTime value.
func (m *CgmState) GetCgmInsertionTime() *int64 {
	return m.CgmInsertionTime
}

// GetRawJsonData returns the RawJsonData value.
func (m *CgmState) GetRawJsonData() *string {
	return m.RawJsonData
}

// GetDeletedAt returns the DeletedAt value.
func (m *CgmState) GetDeletedAt() *int64 {
	return m.DeletedAt
}

// GetUtcOffsetMinutes returns the UtcOffsetMinutes value.
func (m *CgmState) GetUtcOffsetMinutes() *int64 {
	return m.UtcOffsetMinutes
}

// GetOccurredAt returns the OccurredAt value.
func (m *CgmState) GetOccurredAt() int64 {
	return m.OccurredAt
}

// GetCreatedAt returns the CreatedAt value.
func (m *CgmState) GetCreatedAt() int64 {
	return m.CreatedAt
}

// GetUpdatedAt returns the UpdatedAt value.
func (m *CgmState) GetUpdatedAt() int64 {
	return m.UpdatedAt
}

// GetUpdatedByIdentityName returns the UpdatedByIdentityName value.
func (m *CgmState) GetUpdatedByIdentityName() *string {
	return m.UpdatedByIdentityName
}

// GetCreatedByAppName returns the CreatedByAppName value.
func (m *CgmState) GetCreatedByAppName() string {
	return m.CreatedByAppName
}

// GetCreatedByDeviceName returns the CreatedByDeviceName value.
func (m *CgmState) GetCreatedByDeviceName() string {
	return m.CreatedByDeviceName
}

// GetCreatedByIdentityName returns the CreatedByIdentityName value.
func (m *CgmState) GetCreatedByIdentityName() string {
	return m.CreatedByIdentityName
}

// GetID returns the ID value.
func (m *DeviceState) GetID() string {
	return m.ID
}

// GetDeviceState returns the DeviceState value.
func (m *DeviceState) GetDeviceState() *string {
	return m.DeviceState
}

// GetDeviceTimestamp returns the DeviceTimestamp value.
func (m *DeviceState) GetDeviceTimestamp() *int64 {
	return m.DeviceTimestamp
}

// GetDeviceBatteryPercent returns the DeviceBatteryPercent value.
func (m *DeviceState) GetDeviceBatteryPercent() *float64 {
	return m.DeviceBatteryPercent
}

// GetDeviceChargingState returns the DeviceChargingState value.
func (m *DeviceState) GetDeviceChargingState() *string {
	return m.DeviceChargingState
}

// GetRawJsonData returns the RawJsonData value.
func (m *DeviceState) GetRawJsonData() *string {
	return m.RawJsonData
}

// GetDeletedAt returns the DeletedAt value.
func (m *DeviceState) GetDeletedAt() *int64 {
	return m.DeletedAt
}

// GetUtcOffsetMinutes returns the UtcOffsetMinutes value.
func (m *DeviceState) GetUtcOffsetMinutes() *int64 {
	return m.UtcOffsetMinutes
}

// GetOccurredAt returns the OccurredAt value.
func (m *DeviceState) GetOccurredAt() int64 {
	return m.OccurredAt
}

// GetCreatedAt returns the CreatedAt value.
func (m *DeviceState) GetCreatedAt() int64 {
	return m.CreatedAt
}

// GetUpdatedAt returns the UpdatedAt value.
func (m *DeviceState) GetUpdatedAt() int64 {
	return m.UpdatedAt
}

// GetUpdatedByIdentityName returns the UpdatedByIdentityName value.
func (m *DeviceState) GetUpdatedByIdentityName() *string {
	return m.UpdatedByIdentityName
}

// GetCreatedByAppName returns the CreatedByAppName value.
func (m *DeviceState) GetCreatedByAppName() string {
	return m.CreatedByAppName
}

// GetCreatedByDeviceName returns the CreatedByDeviceName value.
func (m *DeviceState) GetCreatedByDeviceName() string {
	return m.CreatedByDeviceName
}

// GetCreatedByIdentityName returns the CreatedByIdentityName value.
func (m *DeviceState) GetCreatedByIdentityName() string {
	return m.CreatedByIdentityName
}

// GetID returns the ID value.
func (m *EventGroup) GetID() string {
	return m.ID
}

// GetGroupType returns the GroupType value.
func (m *EventGroup) GetGroupType() string {
	return m.GroupType
}

// GetGlucoseReadingID returns the GlucoseReadingID value.
func (m *EventGroup) GetGlucoseReadingID() *string {
	return m.GlucoseReadingID
}

// GetInsulinDoseID returns the InsulinDoseID value.
func (m *EventGroup) GetInsulinDoseID() *string {
	return m.InsulinDoseID
}

// GetFoodIntakeID returns the FoodIntakeID value.
func (m *EventGroup) GetFoodIntakeID() *string {
	return m.FoodIntakeID
}

// GetTempBasalID returns the TempBasalID value.
func (m *EventGroup) GetTempBasalID() *string {
	return m.TempBasalID
}

// GetTempTargetID returns the TempTargetID value.
func (m *EventGroup) GetTempTargetID() *string {
	return m.TempTargetID
}

// GetProfileSwitchID returns the ProfileSwitchID value.
func (m *EventGroup) GetProfileSwitchID() *string {
	return m.ProfileSwitchID
}

// GetDeviceStateID returns the DeviceStateID value.
func (m *EventGroup) GetDeviceStateID() *string {
	return m.DeviceStateID
}

// GetLoopStateID returns the LoopStateID value.
func (m *EventGroup) GetLoopStateID() *string {
	return m.LoopStateID
}

// GetPumpStateID returns the PumpStateID value.
func (m *EventGroup) GetPumpStateID() *string {
	return m.PumpStateID
}

// GetCgmStateID returns the CgmStateID value.
func (m *EventGroup) GetCgmStateID() *string {
	return m.CgmStateID
}

// GetActivityID returns the ActivityID value.
func (m *EventGroup) GetActivityID() *string {
	return m.ActivityID
}

// GetNoteID returns the NoteID value.
func (m *EventGroup) GetNoteID() *string {
	return m.NoteID
}

// GetRawJsonData returns the RawJsonData value.
func (m *EventGroup) GetRawJsonData() *string {
	return m.RawJsonData
}

// GetDeletedAt returns the DeletedAt value.
func (m *EventGroup) GetDeletedAt() *int64 {
	return m.DeletedAt
}

// GetUtcOffsetMinutes returns the UtcOffsetMinutes value.
func (m *EventGroup) GetUtcOffsetMinutes() *int64 {
	return m.UtcOffsetMinutes
}

// GetOccurredAt returns the OccurredAt value.
func (m *EventGroup) GetOccurredAt() int64 {
	return m.OccurredAt
}

// GetCreatedAt returns the CreatedAt value.
func (m *EventGroup) GetCreatedAt() int64 {
	return m.CreatedAt
}

// GetUpdatedAt returns the UpdatedAt value.
func (m *EventGroup) GetUpdatedAt() int64 {
	return m.UpdatedAt
}

// GetCreatedByAppName returns the CreatedByAppName value.
func (m *EventGroup) GetCreatedByAppName() string {
	return m.CreatedByAppName
}

// GetCreatedByDeviceName returns the CreatedByDeviceName value.
func (m *EventGroup) GetCreatedByDeviceName() string {
	return m.CreatedByDeviceName
}

// GetUpdatedByIdentityName returns the UpdatedByIdentityName value.
func (m *EventGroup) GetUpdatedByIdentityName() *string {
	return m.UpdatedByIdentityName
}

// GetCreatedByIdentityName returns the CreatedByIdentityName value.
func (m *EventGroup) GetCreatedByIdentityName() string {
	return m.CreatedByIdentityName
}

// GetID returns the ID value.
func (m *FoodIntake) GetID() string {
	return m.ID
}

// GetName returns the Name value.
func (m *FoodIntake) GetName() *string {
	return m.Name
}

// GetFoodType returns the FoodType value.
func (m *FoodIntake) GetFoodType() *string {
	return m.FoodType
}

// GetCarbs returns the Carbs value.
func (m *FoodIntake) GetCarbs() *float64 {
	return m.Carbs
}

// GetFat returns the Fat value.
func (m *FoodIntake) GetFat() *float64 {
	return m.Fat
}

// GetProtein returns the Protein value.
func (m *FoodIntake) GetProtein() *float64 {
	return m.Protein
}

// GetFibre returns the Fibre value.
func (m *FoodIntake) GetFibre() *float64 {
	return m.Fibre
}

// GetCalories returns the Calories value.
func (m *FoodIntake) GetCalories() *float64 {
	return m.Calories
}

// GetRawJsonData returns the RawJsonData value.
func (m *FoodIntake) GetRawJsonData() *string {
	return m.RawJsonData
}

// GetDeletedAt returns the DeletedAt value.
func (m *FoodIntake) GetDeletedAt() *int64 {
	return m.DeletedAt
}

// GetUtcOffsetMinutes returns the UtcOffsetMinutes value.
func (m *FoodIntake) GetUtcOffsetMinutes() *int64 {
	return m.UtcOffsetMinutes
}

// GetOccurredAt returns the OccurredAt value.
func (m *FoodIntake) GetOccurredAt() int64 {
	return m.OccurredAt
}

// GetCreatedAt returns the CreatedAt value.
func (m *FoodIntake) GetCreatedAt() int64 {
	return m.CreatedAt
}

// GetUpdatedAt returns the UpdatedAt value.
func (m *FoodIntake) GetUpdatedAt() int64 {
	return m.UpdatedAt
}

// GetCreatedByAppName returns the CreatedByAppName value.
func (m *FoodIntake) GetCreatedByAppName() string {
	return m.CreatedByAppName
}

// GetCreatedByDeviceName returns the CreatedByDeviceName value.
func (m *FoodIntake) GetCreatedByDeviceName() string {
	return m.CreatedByDeviceName
}

// GetUpdatedByIdentityName returns the UpdatedByIdentityName value.
func (m *FoodIntake) GetUpdatedByIdentityName() *string {
	return m.UpdatedByIdentityName
}

// GetCreatedByIdentityName returns the CreatedByIdentityName value.
func (m *FoodIntake) GetCreatedByIdentityName() string {
	return m.CreatedByIdentityName
}

// GetID returns the ID value.
func (m *FoodPreset) GetID() string {
	return m.ID
}

// GetName returns the Name value.
func (m *FoodPreset) GetName() string {
	return m.Name
}

// GetFoodType returns the FoodType value.
func (m *FoodPreset) GetFoodType() string {
	return m.FoodType
}

// GetCarbs returns the Carbs value.
func (m *FoodPreset) GetCarbs() *float64 {
	return m.Carbs
}

// GetFat returns the Fat value.
func (m *FoodPreset) GetFat() *float64 {
	return m.Fat
}

// GetProtein returns the Protein value.
func (m *FoodPreset) GetProtein() *float64 {
	return m.Protein
}

// GetFibre returns the Fibre value.
func (m *FoodPreset) GetFibre() *float64 {
	return m.Fibre
}

// GetCalories returns the Calories value.
func (m *FoodPreset) GetCalories() *float64 {
	return m.Calories
}

// GetRawJsonData returns the RawJsonData value.
func (m *FoodPreset) GetRawJsonData() *string {
	return m.RawJsonData
}

// GetCreatedAt returns the CreatedAt value.
func (m *FoodPreset) GetCreatedAt() int64 {
	return m.CreatedAt
}

// GetUpdatedAt returns the UpdatedAt value.
func (m *FoodPreset) GetUpdatedAt() int64 {
	return m.UpdatedAt
}

// GetCreatedByAppName returns the CreatedByAppName value.
func (m *FoodPreset) GetCreatedByAppName() string {
	return m.CreatedByAppName
}

// GetCreatedByDeviceName returns the CreatedByDeviceName value.
func (m *FoodPreset) GetCreatedByDeviceName() string {
	return m.CreatedByDeviceName
}

// GetUpdatedByIdentityName returns the UpdatedByIdentityName value.
func (m *FoodPreset) GetUpdatedByIdentityName() *string {
	return m.UpdatedByIdentityName
}

// GetCreatedByIdentityName returns the CreatedByIdentityName value.
func (m *FoodPreset) GetCreatedByIdentityName() string {
	return m.CreatedByIdentityName
}

// GetID returns the ID value.
func (m *GlucoseReading) GetID() string {
	return m.ID
}

// GetReading returns the Reading value.
func (m *GlucoseReading) GetReading() float64 {
	return m.Reading
}

// GetReadingType returns the ReadingType value.
func (m *GlucoseReading) GetReadingType() string {
	return m.ReadingType
}

// GetRawJsonData returns the RawJsonData value.
func (m *GlucoseReading) GetRawJsonData() *string {
	return m.RawJsonData
}

// GetDeletedAt returns the DeletedAt value.
func (m *GlucoseReading) GetDeletedAt() *int64 {
	return m.DeletedAt
}

// GetUtcOffsetMinutes returns the UtcOffsetMinutes value.
func (m *GlucoseReading) GetUtcOffsetMinutes() *int64 {
	return m.UtcOffsetMinutes
}

// GetOccurredAt returns the OccurredAt value.
func (m *GlucoseReading) GetOccurredAt() int64 {
	return m.OccurredAt
}

// GetCreatedAt returns the CreatedAt value.
func (m *GlucoseReading) GetCreatedAt() int64 {
	return m.CreatedAt
}

// GetUpdatedAt returns the UpdatedAt value.
func (m *GlucoseReading) GetUpdatedAt() int64 {
	return m.UpdatedAt
}

// GetCreatedByAppName returns the CreatedByAppName value.
func (m *GlucoseReading) GetCreatedByAppName() string {
	return m.CreatedByAppName
}

// GetCreatedByDeviceName returns the CreatedByDeviceName value.
func (m *GlucoseReading) GetCreatedByDeviceName() string {
	return m.CreatedByDeviceName
}

// GetUpdatedByIdentityName returns the UpdatedByIdentityName value.
func (m *GlucoseReading) GetUpdatedByIdentityName() *string {
	return m.UpdatedByIdentityName
}

// GetCreatedByIdentityName returns the CreatedByIdentityName value.
func (m *GlucoseReading) GetCreatedByIdentityName() string {
	return m.CreatedByIdentityName
}

// GetID returns the ID value.
func (m *InsulinDose) GetID() string {
	return m.ID
}

// GetDose returns the Dose value.
func (m *InsulinDose) GetDose() float64 {
	return m.Dose
}

// GetDoseType returns the DoseType value.
func (m *InsulinDose) GetDoseType() *string {
	return m.DoseType
}

// GetRawJsonData returns the RawJsonData value.
func (m *InsulinDose) GetRawJsonData() *string {
	return m.RawJsonData
}

// GetDeletedAt returns the DeletedAt value.
func (m *InsulinDose) GetDeletedAt() *int64 {
	return m.DeletedAt
}

// GetUtcOffsetMinutes returns the UtcOffsetMinutes value.
func (m *InsulinDose) GetUtcOffsetMinutes() *int64 {
	return m.UtcOffsetMinutes
}

// GetOccurredAt returns the OccurredAt value.
func (m *InsulinDose) GetOccurredAt() int64 {
	return m.OccurredAt
}

// GetCreatedAt returns the CreatedAt value.
func (m *InsulinDose) GetCreatedAt() int64 {
	return m.CreatedAt
}

// GetUpdatedAt returns the UpdatedAt value.
func (m *InsulinDose) GetUpdatedAt() int64 {
	return m.UpdatedAt
}

// GetCreatedByAppName returns the CreatedByAppName value.
func (m *InsulinDose) GetCreatedByAppName() string {
	return m.CreatedByAppName
}

// GetCreatedByDeviceName returns the CreatedByDeviceName value.
func (m *InsulinDose) GetCreatedByDeviceName() string {
	return m.CreatedByDeviceName
}

// GetUpdatedByIdentityName returns the UpdatedByIdentityName value.
func (m *InsulinDose) GetUpdatedByIdentityName() *string {
	return m.UpdatedByIdentityName
}

// GetCreatedByIdentityName returns the CreatedByIdentityName value.
func (m *InsulinDose) GetCreatedByIdentityName() string {
	return m.CreatedByIdentityName
}

// GetID returns the ID value.
func (m *LoopState) GetID() string {
	return m.ID
}

// GetLoopState returns the LoopState value.
func (m *LoopState) GetLoopState() *string {
	return m.LoopState
}

// GetLoopTimestamp returns the LoopTimestamp value.
func (m *LoopState) GetLoopTimestamp() *int64 {
	return m.LoopTimestamp
}

// GetIob returns the Iob value.
func (m *LoopState) GetIob() *float64 {
	return m.Iob
}

// GetIobBasal returns the IobBasal value.
func (m *LoopState) GetIobBasal() *float64 {
	return m.IobBasal
}

// GetIobTreatment returns the IobTreatment value.
func (m *LoopState) GetIobTreatment() *float64 {
	return m.IobTreatment
}

// GetCob returns the Cob value.
func (m *LoopState) GetCob() *float64 {
	return m.Cob
}

// GetRawJsonData returns the RawJsonData value.
func (m *LoopState) GetRawJsonData() *string {
	return m.RawJsonData
}

// GetDeletedAt returns the DeletedAt value.
func (m *LoopState) GetDeletedAt() *int64 {
	return m.DeletedAt
}

// GetUtcOffsetMinutes returns the UtcOffsetMinutes value.
func (m *LoopState) GetUtcOffsetMinutes() *int64 {
	return m.UtcOffsetMinutes
}

// GetOccurredAt returns the OccurredAt value.
func (m *LoopState) GetOccurredAt() int64 {
	return m.OccurredAt
}

// GetCreatedAt returns the CreatedAt value.
func (m *LoopState) GetCreatedAt() int64 {
	return m.CreatedAt
}

// GetUpdatedAt returns the UpdatedAt value.
func (m *LoopState) GetUpdatedAt() int64 {
	return m.UpdatedAt
}

// GetUpdatedByIdentityName returns the UpdatedByIdentityName value.
func (m *LoopState) GetUpdatedByIdentityName() *string {
	return m.UpdatedByIdentityName
}

// GetCreatedByAppName returns the CreatedByAppName value.
func (m *LoopState) GetCreatedByAppName() string {
	return m.CreatedByAppName
}

// GetCreatedByDeviceName returns the CreatedByDeviceName value.
func (m *LoopState) GetCreatedByDeviceName() string {
	return m.CreatedByDeviceName
}

// GetCreatedByIdentityName returns the CreatedByIdentityName value.
func (m *LoopState) GetCreatedByIdentityName() string {
	return m.CreatedByIdentityName
}

// GetID returns the ID value.
func (m *Note) GetID() string {
	return m.ID
}

// GetContent returns the Content value.
func (m *Note) GetContent() string {
	return m.Content
}

// GetReason returns the Reason value.
func (m *Note) GetReason() *string {
	return m.Reason
}

// GetRawJsonData returns the RawJsonData value.
func (m *Note) GetRawJsonData() *string {
	return m.RawJsonData
}

// GetDeletedAt returns the DeletedAt value.
func (m *Note) GetDeletedAt() *int64 {
	return m.DeletedAt
}

// GetUtcOffsetMinutes returns the UtcOffsetMinutes value.
func (m *Note) GetUtcOffsetMinutes() *int64 {
	return m.UtcOffsetMinutes
}

// GetOccurredAt returns the OccurredAt value.
func (m *Note) GetOccurredAt() int64 {
	return m.OccurredAt
}

// GetCreatedAt returns the CreatedAt value.
func (m *Note) GetCreatedAt() int64 {
	return m.CreatedAt
}

// GetUpdatedAt returns the UpdatedAt value.
func (m *Note) GetUpdatedAt() int64 {
	return m.UpdatedAt
}

// GetCreatedByAppName returns the CreatedByAppName value.
func (m *Note) GetCreatedByAppName() string {
	return m.CreatedByAppName
}

// GetCreatedByDeviceName returns the CreatedByDeviceName value.
func (m *Note) GetCreatedByDeviceName() string {
	return m.CreatedByDeviceName
}

// GetUpdatedByIdentityName returns the UpdatedByIdentityName value.
func (m *Note) GetUpdatedByIdentityName() *string {
	return m.UpdatedByIdentityName
}

// GetCreatedByIdentityName returns the CreatedByIdentityName value.
func (m *Note) GetCreatedByIdentityName() string {
	return m.CreatedByIdentityName
}

// GetID returns the ID value.
func (m *ProfilePreset) GetID() string {
	return m.ID
}

// GetName returns the Name value.
func (m *ProfilePreset) GetName() string {
	return m.Name
}

// GetCarbRatiosJson returns the CarbRatiosJson value.
func (m *ProfilePreset) GetCarbRatiosJson() *string {
	return m.CarbRatiosJson
}

// GetInsulinSensitivitiesJson returns the InsulinSensitivitiesJson value.
func (m *ProfilePreset) GetInsulinSensitivitiesJson() *string {
	return m.InsulinSensitivitiesJson
}

// GetBasalRatesJson returns the BasalRatesJson value.
func (m *ProfilePreset) GetBasalRatesJson() *string {
	return m.BasalRatesJson
}

// GetTargetsJson returns the TargetsJson value.
func (m *ProfilePreset) GetTargetsJson() *string {
	return m.TargetsJson
}

// GetRawJsonData returns the RawJsonData value.
func (m *ProfilePreset) GetRawJsonData() *string {
	return m.RawJsonData
}

// GetCreatedAt returns the CreatedAt value.
func (m *ProfilePreset) GetCreatedAt() int64 {
	return m.CreatedAt
}

// GetUpdatedAt returns the UpdatedAt value.
func (m *ProfilePreset) GetUpdatedAt() int64 {
	return m.UpdatedAt
}

// GetCreatedByAppName returns the CreatedByAppName value.
func (m *ProfilePreset) GetCreatedByAppName() string {
	return m.CreatedByAppName
}

// GetCreatedByDeviceName returns the CreatedByDeviceName value.
func (m *ProfilePreset) GetCreatedByDeviceName() string {
	return m.CreatedByDeviceName
}

// GetUpdatedByIdentityName returns the UpdatedByIdentityName value.
func (m *ProfilePreset) GetUpdatedByIdentityName() *string {
	return m.UpdatedByIdentityName
}

// GetCreatedByIdentityName returns the CreatedByIdentityName value.
func (m *ProfilePreset) GetCreatedByIdentityName() string {
	return m.CreatedByIdentityName
}

// GetID returns the ID value.
func (m *ProfileSwitch) GetID() string {
	return m.ID
}

// GetName returns the Name value.
func (m *ProfileSwitch) GetName() string {
	return m.Name
}

// GetDurationMinutes returns the DurationMinutes value.
func (m *ProfileSwitch) GetDurationMinutes() int64 {
	return m.DurationMinutes
}

// GetPercent returns the Percent value.
func (m *ProfileSwitch) GetPercent() float64 {
	return m.Percent
}

// GetCarbRatiosJson returns the CarbRatiosJson value.
func (m *ProfileSwitch) GetCarbRatiosJson() *string {
	return m.CarbRatiosJson
}

// GetInsulinSensitivitiesJson returns the InsulinSensitivitiesJson value.
func (m *ProfileSwitch) GetInsulinSensitivitiesJson() *string {
	return m.InsulinSensitivitiesJson
}

// GetBasalRatesJson returns the BasalRatesJson value.
func (m *ProfileSwitch) GetBasalRatesJson() *string {
	return m.BasalRatesJson
}

// GetTargetsJson returns the TargetsJson value.
func (m *ProfileSwitch) GetTargetsJson() *string {
	return m.TargetsJson
}

// GetRawJsonData returns the RawJsonData value.
func (m *ProfileSwitch) GetRawJsonData() *string {
	return m.RawJsonData
}

// GetDeletedAt returns the DeletedAt value.
func (m *ProfileSwitch) GetDeletedAt() *int64 {
	return m.DeletedAt
}

// GetUtcOffsetMinutes returns the UtcOffsetMinutes value.
func (m *ProfileSwitch) GetUtcOffsetMinutes() *int64 {
	return m.UtcOffsetMinutes
}

// GetOccurredAt returns the OccurredAt value.
func (m *ProfileSwitch) GetOccurredAt() int64 {
	return m.OccurredAt
}

// GetCreatedAt returns the CreatedAt value.
func (m *ProfileSwitch) GetCreatedAt() int64 {
	return m.CreatedAt
}

// GetUpdatedAt returns the UpdatedAt value.
func (m *ProfileSwitch) GetUpdatedAt() int64 {
	return m.UpdatedAt
}

// GetCreatedByAppName returns the CreatedByAppName value.
func (m *ProfileSwitch) GetCreatedByAppName() string {
	return m.CreatedByAppName
}

// GetCreatedByDeviceName returns the CreatedByDeviceName value.
func (m *ProfileSwitch) GetCreatedByDeviceName() string {
	return m.CreatedByDeviceName
}

// GetUpdatedByIdentityName returns the UpdatedByIdentityName value.
func (m *ProfileSwitch) GetUpdatedByIdentityName() *string {
	return m.UpdatedByIdentityName
}

// GetCreatedByIdentityName returns the CreatedByIdentityName value.
func (m *ProfileSwitch) GetCreatedByIdentityName() string {
	return m.CreatedByIdentityName
}

// GetID returns the ID value.
func (m *PumpState) GetID() string {
	return m.ID
}

// GetPumpState returns the PumpState value.
func (m *PumpState) GetPumpState() *string {
	return m.PumpState
}

// GetPumpTimestamp returns the PumpTimestamp value.
func (m *PumpState) GetPumpTimestamp() *int64 {
	return m.PumpTimestamp
}

// GetPumpInsertionTime returns the PumpInsertionTime value.
func (m *PumpState) GetPumpInsertionTime() *int64 {
	return m.PumpInsertionTime
}

// GetPumpReservoir returns the PumpReservoir value.
func (m *PumpState) GetPumpReservoir() *float64 {
	return m.PumpReservoir
}

// GetPumpBatteryPercent returns the PumpBatteryPercent value.
func (m *PumpState) GetPumpBatteryPercent() *float64 {
	return m.PumpBatteryPercent
}

// GetPumpChargingState returns the PumpChargingState value.
func (m *PumpState) GetPumpChargingState() *string {
	return m.PumpChargingState
}

// GetRawJsonData returns the RawJsonData value.
func (m *PumpState) GetRawJsonData() *string {
	return m.RawJsonData
}

// GetDeletedAt returns the DeletedAt value.
func (m *PumpState) GetDeletedAt() *int64 {
	return m.DeletedAt
}

// GetUtcOffsetMinutes returns the UtcOffsetMinutes value.
func (m *PumpState) GetUtcOffsetMinutes() *int64 {
	return m.UtcOffsetMinutes
}

// GetOccurredAt returns the OccurredAt value.
func (m *PumpState) GetOccurredAt() int64 {
	return m.OccurredAt
}

// GetCreatedAt returns the CreatedAt value.
func (m *PumpState) GetCreatedAt() int64 {
	return m.CreatedAt
}

// GetUpdatedAt returns the UpdatedAt value.
func (m *PumpState) GetUpdatedAt() int64 {
	return m.UpdatedAt
}

// GetUpdatedByIdentityName returns the UpdatedByIdentityName value.
func (m *PumpState) GetUpdatedByIdentityName() *string {
	return m.UpdatedByIdentityName
}

// GetCreatedByAppName returns the CreatedByAppName value.
func (m *PumpState) GetCreatedByAppName() string {
	return m.CreatedByAppName
}

// GetCreatedByDeviceName returns the CreatedByDeviceName value.
func (m *PumpState) GetCreatedByDeviceName() string {
	return m.CreatedByDeviceName
}

// GetCreatedByIdentityName returns the CreatedByIdentityName value.
func (m *PumpState) GetCreatedByIdentityName() string {
	return m.CreatedByIdentityName
}

// GetID returns the ID value.
func (m *SiteSetting) GetID() int64 {
	return m.ID
}

// GetUnits returns the Units value.
func (m *SiteSetting) GetUnits() string {
	return m.Units
}

// GetTimezone returns the Timezone value.
func (m *SiteSetting) GetTimezone() string {
	return m.Timezone
}

// GetBolusDiaHours returns the BolusDiaHours value.
func (m *SiteSetting) GetBolusDiaHours() float64 {
	return m.BolusDiaHours
}

// GetBolusOnsetMinutes returns the BolusOnsetMinutes value.
func (m *SiteSetting) GetBolusOnsetMinutes() float64 {
	return m.BolusOnsetMinutes
}

// GetBolusPeakMinutes returns the BolusPeakMinutes value.
func (m *SiteSetting) GetBolusPeakMinutes() float64 {
	return m.BolusPeakMinutes
}

// GetBasalDiaHours returns the BasalDiaHours value.
func (m *SiteSetting) GetBasalDiaHours() float64 {
	return m.BasalDiaHours
}

// GetBasalOnsetMinutes returns the BasalOnsetMinutes value.
func (m *SiteSetting) GetBasalOnsetMinutes() float64 {
	return m.BasalOnsetMinutes
}

// GetBasalPeakMinutes returns the BasalPeakMinutes value.
func (m *SiteSetting) GetBasalPeakMinutes() float64 {
	return m.BasalPeakMinutes
}

// GetCarbMaxAbsorptionHours returns the CarbMaxAbsorptionHours value.
func (m *SiteSetting) GetCarbMaxAbsorptionHours() float64 {
	return m.CarbMaxAbsorptionHours
}

// GetCarbMin5mImpact returns the CarbMin5mImpact value.
func (m *SiteSetting) GetCarbMin5mImpact() float64 {
	return m.CarbMin5mImpact
}

// GetUpdatedAt returns the UpdatedAt value.
func (m *SiteSetting) GetUpdatedAt() int64 {
	return m.UpdatedAt
}

// GetUpdatedByIdentityName returns the UpdatedByIdentityName value.
func (m *SiteSetting) GetUpdatedByIdentityName() *string {
	return m.UpdatedByIdentityName
}

// GetID returns the ID value.
func (m *TempBasal) GetID() string {
	return m.ID
}

// GetDurationMinutes returns the DurationMinutes value.
func (m *TempBasal) GetDurationMinutes() int64 {
	return m.DurationMinutes
}

// GetAbsolute returns the Absolute value.
func (m *TempBasal) GetAbsolute() float64 {
	return m.Absolute
}

// GetPercent returns the Percent value.
func (m *TempBasal) GetPercent() *float64 {
	return m.Percent
}

// GetRawJsonData returns the RawJsonData value.
func (m *TempBasal) GetRawJsonData() *string {
	return m.RawJsonData
}

// GetDeletedAt returns the DeletedAt value.
func (m *TempBasal) GetDeletedAt() *int64 {
	return m.DeletedAt
}

// GetUtcOffsetMinutes returns the UtcOffsetMinutes value.
func (m *TempBasal) GetUtcOffsetMinutes() *int64 {
	return m.UtcOffsetMinutes
}

// GetOccurredAt returns the OccurredAt value.
func (m *TempBasal) GetOccurredAt() int64 {
	return m.OccurredAt
}

// GetCreatedAt returns the CreatedAt value.
func (m *TempBasal) GetCreatedAt() int64 {
	return m.CreatedAt
}

// GetUpdatedAt returns the UpdatedAt value.
func (m *TempBasal) GetUpdatedAt() int64 {
	return m.UpdatedAt
}

// GetCreatedByAppName returns the CreatedByAppName value.
func (m *TempBasal) GetCreatedByAppName() string {
	return m.CreatedByAppName
}

// GetCreatedByDeviceName returns the CreatedByDeviceName value.
func (m *TempBasal) GetCreatedByDeviceName() string {
	return m.CreatedByDeviceName
}

// GetUpdatedByIdentityName returns the UpdatedByIdentityName value.
func (m *TempBasal) GetUpdatedByIdentityName() *string {
	return m.UpdatedByIdentityName
}

// GetCreatedByIdentityName returns the CreatedByIdentityName value.
func (m *TempBasal) GetCreatedByIdentityName() string {
	return m.CreatedByIdentityName
}

// GetID returns the ID value.
func (m *TempTarget) GetID() string {
	return m.ID
}

// GetDurationMinutes returns the DurationMinutes value.
func (m *TempTarget) GetDurationMinutes() int64 {
	return m.DurationMinutes
}

// GetHighTarget returns the HighTarget value.
func (m *TempTarget) GetHighTarget() float64 {
	return m.HighTarget
}

// GetLowTarget returns the LowTarget value.
func (m *TempTarget) GetLowTarget() float64 {
	return m.LowTarget
}

// GetRawJsonData returns the RawJsonData value.
func (m *TempTarget) GetRawJsonData() *string {
	return m.RawJsonData
}

// GetDeletedAt returns the DeletedAt value.
func (m *TempTarget) GetDeletedAt() *int64 {
	return m.DeletedAt
}

// GetUtcOffsetMinutes returns the UtcOffsetMinutes value.
func (m *TempTarget) GetUtcOffsetMinutes() *int64 {
	return m.UtcOffsetMinutes
}

// GetOccurredAt returns the OccurredAt value.
func (m *TempTarget) GetOccurredAt() int64 {
	return m.OccurredAt
}

// GetCreatedAt returns the CreatedAt value.
func (m *TempTarget) GetCreatedAt() int64 {
	return m.CreatedAt
}

// GetUpdatedAt returns the UpdatedAt value.
func (m *TempTarget) GetUpdatedAt() int64 {
	return m.UpdatedAt
}

// GetCreatedByAppName returns the CreatedByAppName value.
func (m *TempTarget) GetCreatedByAppName() string {
	return m.CreatedByAppName
}

// GetCreatedByDeviceName returns the CreatedByDeviceName value.
func (m *TempTarget) GetCreatedByDeviceName() string {
	return m.CreatedByDeviceName
}

// GetUpdatedByIdentityName returns the UpdatedByIdentityName value.
func (m *TempTarget) GetUpdatedByIdentityName() *string {
	return m.UpdatedByIdentityName
}

// GetCreatedByIdentityName returns the CreatedByIdentityName value.
func (m *TempTarget) GetCreatedByIdentityName() string {
	return m.CreatedByIdentityName
}

