
package user

import (
	"encoding/json"
	"nighttroop.t1d.nz/utils/consts"
)
// GetID returns the ID value.
func (m *Activity) GetID() string {
	return m.ID
}

// GetName returns the Name value.
func (m *Activity) GetName() *string {
	return m.Name
}

// GetType returns the Type value.
func (m *Activity) GetType() *string {
	return m.Type
}

// GetNotes returns the Notes value.
func (m *Activity) GetNotes() *string {
	return m.Notes
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

// GetCreatedByIdentityName returns the CreatedByIdentityName value.
func (m *Activity) GetCreatedByIdentityName() string {
	return m.CreatedByIdentityName
}

// GetCreatedByIdentityID returns the CreatedByIdentityID value.
func (m *Activity) GetCreatedByIdentityID() int64 {
	return m.CreatedByIdentityID
}

// GetID returns the ID value.
func (m *Bolus) GetID() string {
	return m.ID
}

// GetValue returns the Value value.
func (m *Bolus) GetValue() float64 {
	return m.Value
}

// GetType returns the Type value.
func (m *Bolus) GetType() string {
	return m.Type
}

// GetRawJsonData returns the RawJsonData value.
func (m *Bolus) GetRawJsonData() *string {
	return m.RawJsonData
}

// GetDeletedAt returns the DeletedAt value.
func (m *Bolus) GetDeletedAt() *int64 {
	return m.DeletedAt
}

// GetUtcOffsetMinutes returns the UtcOffsetMinutes value.
func (m *Bolus) GetUtcOffsetMinutes() *int64 {
	return m.UtcOffsetMinutes
}

// GetOccurredAt returns the OccurredAt value.
func (m *Bolus) GetOccurredAt() int64 {
	return m.OccurredAt
}

// GetCreatedAt returns the CreatedAt value.
func (m *Bolus) GetCreatedAt() int64 {
	return m.CreatedAt
}

// GetUpdatedAt returns the UpdatedAt value.
func (m *Bolus) GetUpdatedAt() int64 {
	return m.UpdatedAt
}

// GetCreatedByIdentityName returns the CreatedByIdentityName value.
func (m *Bolus) GetCreatedByIdentityName() string {
	return m.CreatedByIdentityName
}

// GetCreatedByIdentityID returns the CreatedByIdentityID value.
func (m *Bolus) GetCreatedByIdentityID() int64 {
	return m.CreatedByIdentityID
}

// GetID returns the ID value.
func (m *Carb) GetID() string {
	return m.ID
}

// GetAmount returns the Amount value.
func (m *Carb) GetAmount() float64 {
	return m.Amount
}

// GetType returns the Type value.
func (m *Carb) GetType() string {
	return m.Type
}

// GetRawJsonData returns the RawJsonData value.
func (m *Carb) GetRawJsonData() *string {
	return m.RawJsonData
}

// GetDeletedAt returns the DeletedAt value.
func (m *Carb) GetDeletedAt() *int64 {
	return m.DeletedAt
}

// GetUtcOffsetMinutes returns the UtcOffsetMinutes value.
func (m *Carb) GetUtcOffsetMinutes() *int64 {
	return m.UtcOffsetMinutes
}

// GetOccurredAt returns the OccurredAt value.
func (m *Carb) GetOccurredAt() int64 {
	return m.OccurredAt
}

// GetCreatedAt returns the CreatedAt value.
func (m *Carb) GetCreatedAt() int64 {
	return m.CreatedAt
}

// GetUpdatedAt returns the UpdatedAt value.
func (m *Carb) GetUpdatedAt() int64 {
	return m.UpdatedAt
}

// GetCreatedByIdentityName returns the CreatedByIdentityName value.
func (m *Carb) GetCreatedByIdentityName() string {
	return m.CreatedByIdentityName
}

// GetCreatedByIdentityID returns the CreatedByIdentityID value.
func (m *Carb) GetCreatedByIdentityID() int64 {
	return m.CreatedByIdentityID
}

// GetID returns the ID value.
func (m *DeviceStatus) GetID() string {
	return m.ID
}

// GetIdentityID returns the IdentityID value.
func (m *DeviceStatus) GetIdentityID() int64 {
	return m.IdentityID
}

// GetBatteryLevel returns the BatteryLevel value.
func (m *DeviceStatus) GetBatteryLevel() *int64 {
	return m.BatteryLevel
}

// GetChargingState returns the ChargingState value.
func (m *DeviceStatus) GetChargingState() *string {
	return m.ChargingState
}

// GetPumpBattery returns the PumpBattery value.
func (m *DeviceStatus) GetPumpBattery() *int64 {
	return m.PumpBattery
}

// GetPumpStatus returns the PumpStatus value.
func (m *DeviceStatus) GetPumpStatus() *string {
	return m.PumpStatus
}

// GetPumpReservoir returns the PumpReservoir value.
func (m *DeviceStatus) GetPumpReservoir() *float64 {
	return m.PumpReservoir
}

// GetPumpInsertionTime returns the PumpInsertionTime value.
func (m *DeviceStatus) GetPumpInsertionTime() *int64 {
	return m.PumpInsertionTime
}

// GetLoopStatus returns the LoopStatus value.
func (m *DeviceStatus) GetLoopStatus() *string {
	return m.LoopStatus
}

// GetLoopTimestamp returns the LoopTimestamp value.
func (m *DeviceStatus) GetLoopTimestamp() *int64 {
	return m.LoopTimestamp
}

// GetCgmStatus returns the CgmStatus value.
func (m *DeviceStatus) GetCgmStatus() *string {
	return m.CgmStatus
}

// GetCgmInsertionTime returns the CgmInsertionTime value.
func (m *DeviceStatus) GetCgmInsertionTime() *int64 {
	return m.CgmInsertionTime
}

// GetRawJsonData returns the RawJsonData value.
func (m *DeviceStatus) GetRawJsonData() *string {
	return m.RawJsonData
}

// GetDeletedAt returns the DeletedAt value.
func (m *DeviceStatus) GetDeletedAt() *int64 {
	return m.DeletedAt
}

// GetUtcOffsetMinutes returns the UtcOffsetMinutes value.
func (m *DeviceStatus) GetUtcOffsetMinutes() *int64 {
	return m.UtcOffsetMinutes
}

// GetOccurredAt returns the OccurredAt value.
func (m *DeviceStatus) GetOccurredAt() int64 {
	return m.OccurredAt
}

// GetCreatedAt returns the CreatedAt value.
func (m *DeviceStatus) GetCreatedAt() int64 {
	return m.CreatedAt
}

// GetUpdatedAt returns the UpdatedAt value.
func (m *DeviceStatus) GetUpdatedAt() int64 {
	return m.UpdatedAt
}

// GetCreatedByIdentityName returns the CreatedByIdentityName value.
func (m *DeviceStatus) GetCreatedByIdentityName() string {
	return m.CreatedByIdentityName
}

// GetCreatedByIdentityID returns the CreatedByIdentityID value.
func (m *DeviceStatus) GetCreatedByIdentityID() int64 {
	return m.CreatedByIdentityID
}

// GetID returns the ID value.
func (m *Event) GetID() int64 {
	return m.ID
}

// GetType returns the Type value.
func (m *Event) GetType() *string {
	return m.Type
}

// GetDetails returns the Details value.
func (m *Event) GetDetails() *string {
	return m.Details
}

// GetStartAt returns the StartAt value.
func (m *Event) GetStartAt() int64 {
	return m.StartAt
}

// GetEndAt returns the EndAt value.
func (m *Event) GetEndAt() *int64 {
	return m.EndAt
}

// GetID returns the ID value.
func (m *Food) GetID() string {
	return m.ID
}

// GetName returns the Name value.
func (m *Food) GetName() *string {
	return m.Name
}

// GetCarbs returns the Carbs value.
func (m *Food) GetCarbs() *float64 {
	return m.Carbs
}

// GetCalories returns the Calories value.
func (m *Food) GetCalories() *float64 {
	return m.Calories
}

// GetRawJsonData returns the RawJsonData value.
func (m *Food) GetRawJsonData() *string {
	return m.RawJsonData
}

// GetDeletedAt returns the DeletedAt value.
func (m *Food) GetDeletedAt() *int64 {
	return m.DeletedAt
}

// GetCreatedAt returns the CreatedAt value.
func (m *Food) GetCreatedAt() int64 {
	return m.CreatedAt
}

// GetUpdatedAt returns the UpdatedAt value.
func (m *Food) GetUpdatedAt() int64 {
	return m.UpdatedAt
}

// GetCreatedByIdentityName returns the CreatedByIdentityName value.
func (m *Food) GetCreatedByIdentityName() string {
	return m.CreatedByIdentityName
}

// GetCreatedByIdentityID returns the CreatedByIdentityID value.
func (m *Food) GetCreatedByIdentityID() int64 {
	return m.CreatedByIdentityID
}

// GetID returns the ID value.
func (m *GlucoseReading) GetID() string {
	return m.ID
}

// GetValue returns the Value value.
func (m *GlucoseReading) GetValue() float64 {
	return m.Value
}

// GetType returns the Type value.
func (m *GlucoseReading) GetType() string {
	return m.Type
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

// GetCreatedByIdentityName returns the CreatedByIdentityName value.
func (m *GlucoseReading) GetCreatedByIdentityName() string {
	return m.CreatedByIdentityName
}

// GetCreatedByIdentityID returns the CreatedByIdentityID value.
func (m *GlucoseReading) GetCreatedByIdentityID() int64 {
	return m.CreatedByIdentityID
}

// GetID returns the ID value.
func (m *Identity) GetID() int64 {
	return m.ID
}

// GetName returns the Name value.
func (m *Identity) GetName() string {
	return m.Name
}

// GetPassword returns the Password value.
func (m *Identity) GetPassword() string {
	return m.Password
}

// GetTotpSecret returns the TotpSecret value.
func (m *Identity) GetTotpSecret() string {
	return m.TotpSecret
}

// GetRole returns the Role value.
func (m *Identity) GetRole() consts.IdentityRole {
	return m.Role
}

// GetCreatedAt returns the CreatedAt value.
func (m *Identity) GetCreatedAt() int64 {
	return m.CreatedAt
}

// GetLastSeenAt returns the LastSeenAt value.
func (m *Identity) GetLastSeenAt() *int64 {
	return m.LastSeenAt
}

// GetLastWriteAt returns the LastWriteAt value.
func (m *Identity) GetLastWriteAt() *int64 {
	return m.LastWriteAt
}

// GetDeletedAt returns the DeletedAt value.
func (m *Identity) GetDeletedAt() *int64 {
	return m.DeletedAt
}

// GetEnabled returns the Enabled value.
func (m *Identity) GetEnabled() bool {
	return m.Enabled
}

// GetID returns the ID value.
func (m *Note) GetID() string {
	return m.ID
}

// GetContent returns the Content value.
func (m *Note) GetContent() string {
	return m.Content
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

// GetCreatedByIdentityName returns the CreatedByIdentityName value.
func (m *Note) GetCreatedByIdentityName() string {
	return m.CreatedByIdentityName
}

// GetCreatedByIdentityID returns the CreatedByIdentityID value.
func (m *Note) GetCreatedByIdentityID() int64 {
	return m.CreatedByIdentityID
}

// GetID returns the ID value.
func (m *Profile) GetID() string {
	return m.ID
}

// GetName returns the Name value.
func (m *Profile) GetName() *string {
	return m.Name
}

// GetStart returns the Start value.
func (m *Profile) GetStart() *int64 {
	return m.Start
}

// GetDurationMinutes returns the DurationMinutes value.
func (m *Profile) GetDurationMinutes() *int64 {
	return m.DurationMinutes
}

// GetSource returns the Source value.
func (m *Profile) GetSource() *string {
	return m.Source
}

// GetPercent returns the Percent value.
func (m *Profile) GetPercent() *int64 {
	return m.Percent
}

// GetDia returns the Dia value.
func (m *Profile) GetDia() *int64 {
	return m.Dia
}

// GetUnits returns the Units value.
func (m *Profile) GetUnits() *string {
	return m.Units
}

// GetTimezone returns the Timezone value.
func (m *Profile) GetTimezone() *string {
	return m.Timezone
}

// GetCarbRatiosJson returns the CarbRatiosJson value.
func (m *Profile) GetCarbRatiosJson() *string {
	return m.CarbRatiosJson
}

// GetInsulinSensitivitiesJson returns the InsulinSensitivitiesJson value.
func (m *Profile) GetInsulinSensitivitiesJson() *string {
	return m.InsulinSensitivitiesJson
}

// GetBasalRatesJson returns the BasalRatesJson value.
func (m *Profile) GetBasalRatesJson() *string {
	return m.BasalRatesJson
}

// GetTargetsJson returns the TargetsJson value.
func (m *Profile) GetTargetsJson() *string {
	return m.TargetsJson
}

// GetRawJsonData returns the RawJsonData value.
func (m *Profile) GetRawJsonData() *string {
	return m.RawJsonData
}

// GetDeletedAt returns the DeletedAt value.
func (m *Profile) GetDeletedAt() *int64 {
	return m.DeletedAt
}

// GetUtcOffsetMinutes returns the UtcOffsetMinutes value.
func (m *Profile) GetUtcOffsetMinutes() *int64 {
	return m.UtcOffsetMinutes
}

// GetOccurredAt returns the OccurredAt value.
func (m *Profile) GetOccurredAt() int64 {
	return m.OccurredAt
}

// GetCreatedAt returns the CreatedAt value.
func (m *Profile) GetCreatedAt() int64 {
	return m.CreatedAt
}

// GetUpdatedAt returns the UpdatedAt value.
func (m *Profile) GetUpdatedAt() int64 {
	return m.UpdatedAt
}

// GetCreatedByIdentityName returns the CreatedByIdentityName value.
func (m *Profile) GetCreatedByIdentityName() string {
	return m.CreatedByIdentityName
}

// GetCreatedByIdentityID returns the CreatedByIdentityID value.
func (m *Profile) GetCreatedByIdentityID() int64 {
	return m.CreatedByIdentityID
}

// GetID returns the ID value.
func (m *Setting) GetID() string {
	return m.ID
}

// GetValue returns the Value value.
func (m *Setting) GetValue() json.RawMessage {
	return m.Value
}

// GetDeletedAt returns the DeletedAt value.
func (m *Setting) GetDeletedAt() *int64 {
	return m.DeletedAt
}

// GetCreatedAt returns the CreatedAt value.
func (m *Setting) GetCreatedAt() int64 {
	return m.CreatedAt
}

// GetUpdatedAt returns the UpdatedAt value.
func (m *Setting) GetUpdatedAt() int64 {
	return m.UpdatedAt
}

// GetName returns the Name value.
func (m *SqliteSequence) GetName() interface{} {
	return m.Name
}

// GetSeq returns the Seq value.
func (m *SqliteSequence) GetSeq() interface{} {
	return m.Seq
}

// GetID returns the ID value.
func (m *TempBasal) GetID() string {
	return m.ID
}

// GetStart returns the Start value.
func (m *TempBasal) GetStart() int64 {
	return m.Start
}

// GetDurationMinutes returns the DurationMinutes value.
func (m *TempBasal) GetDurationMinutes() int64 {
	return m.DurationMinutes
}

// GetValue returns the Value value.
func (m *TempBasal) GetValue() float64 {
	return m.Value
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

// GetCreatedAt returns the CreatedAt value.
func (m *TempBasal) GetCreatedAt() int64 {
	return m.CreatedAt
}

// GetUpdatedAt returns the UpdatedAt value.
func (m *TempBasal) GetUpdatedAt() int64 {
	return m.UpdatedAt
}

// GetCreatedByIdentityName returns the CreatedByIdentityName value.
func (m *TempBasal) GetCreatedByIdentityName() string {
	return m.CreatedByIdentityName
}

// GetCreatedByIdentityID returns the CreatedByIdentityID value.
func (m *TempBasal) GetCreatedByIdentityID() int64 {
	return m.CreatedByIdentityID
}

// GetID returns the ID value.
func (m *TempTarget) GetID() string {
	return m.ID
}

// GetStart returns the Start value.
func (m *TempTarget) GetStart() int64 {
	return m.Start
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

// GetCreatedAt returns the CreatedAt value.
func (m *TempTarget) GetCreatedAt() int64 {
	return m.CreatedAt
}

// GetUpdatedAt returns the UpdatedAt value.
func (m *TempTarget) GetUpdatedAt() int64 {
	return m.UpdatedAt
}

// GetCreatedByIdentityName returns the CreatedByIdentityName value.
func (m *TempTarget) GetCreatedByIdentityName() string {
	return m.CreatedByIdentityName
}

// GetCreatedByIdentityID returns the CreatedByIdentityID value.
func (m *TempTarget) GetCreatedByIdentityID() int64 {
	return m.CreatedByIdentityID
}

