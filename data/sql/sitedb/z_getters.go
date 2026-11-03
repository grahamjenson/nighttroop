
package sitedb

import (
	"nighttroop.t1d.nz/utils/consts"
)


// GetID returns the ID value.
func (m ActivitiesVw) GetID() *string {
	return m.ID
}

// SetID sets the ID value and returns the updated ActivitiesVw.
func (m *ActivitiesVw) SetID(v *string) {
	m.ID = v
}

// GetName returns the Name value.
func (m ActivitiesVw) GetName() *string {
	return m.Name
}

// SetName sets the Name value and returns the updated ActivitiesVw.
func (m *ActivitiesVw) SetName(v *string) {
	m.Name = v
}

// GetActivityType returns the ActivityType value.
func (m ActivitiesVw) GetActivityType() *string {
	return m.ActivityType
}

// SetActivityType sets the ActivityType value and returns the updated ActivitiesVw.
func (m *ActivitiesVw) SetActivityType(v *string) {
	m.ActivityType = v
}

// GetDurationMinutes returns the DurationMinutes value.
func (m ActivitiesVw) GetDurationMinutes() *int64 {
	return m.DurationMinutes
}

// SetDurationMinutes sets the DurationMinutes value and returns the updated ActivitiesVw.
func (m *ActivitiesVw) SetDurationMinutes(v *int64) {
	m.DurationMinutes = v
}

// GetCalories returns the Calories value.
func (m ActivitiesVw) GetCalories() *int64 {
	return m.Calories
}

// SetCalories sets the Calories value and returns the updated ActivitiesVw.
func (m *ActivitiesVw) SetCalories(v *int64) {
	m.Calories = v
}

// GetSteps returns the Steps value.
func (m ActivitiesVw) GetSteps() *int64 {
	return m.Steps
}

// SetSteps sets the Steps value and returns the updated ActivitiesVw.
func (m *ActivitiesVw) SetSteps(v *int64) {
	m.Steps = v
}

// GetHeartRate returns the HeartRate value.
func (m ActivitiesVw) GetHeartRate() *int64 {
	return m.HeartRate
}

// SetHeartRate sets the HeartRate value and returns the updated ActivitiesVw.
func (m *ActivitiesVw) SetHeartRate(v *int64) {
	m.HeartRate = v
}

// GetNotes returns the Notes value.
func (m ActivitiesVw) GetNotes() *string {
	return m.Notes
}

// SetNotes sets the Notes value and returns the updated ActivitiesVw.
func (m *ActivitiesVw) SetNotes(v *string) {
	m.Notes = v
}

// GetReason returns the Reason value.
func (m ActivitiesVw) GetReason() *string {
	return m.Reason
}

// SetReason sets the Reason value and returns the updated ActivitiesVw.
func (m *ActivitiesVw) SetReason(v *string) {
	m.Reason = v
}

// GetRawJsonData returns the RawJsonData value.
func (m ActivitiesVw) GetRawJsonData() *string {
	return m.RawJsonData
}

// SetRawJsonData sets the RawJsonData value and returns the updated ActivitiesVw.
func (m *ActivitiesVw) SetRawJsonData(v *string) {
	m.RawJsonData = v
}

// GetDeletedAt returns the DeletedAt value.
func (m ActivitiesVw) GetDeletedAt() *int64 {
	return m.DeletedAt
}

// SetDeletedAt sets the DeletedAt value and returns the updated ActivitiesVw.
func (m *ActivitiesVw) SetDeletedAt(v *int64) {
	m.DeletedAt = v
}

// GetUtcOffsetMinutes returns the UtcOffsetMinutes value.
func (m ActivitiesVw) GetUtcOffsetMinutes() *int64 {
	return m.UtcOffsetMinutes
}

// SetUtcOffsetMinutes sets the UtcOffsetMinutes value and returns the updated ActivitiesVw.
func (m *ActivitiesVw) SetUtcOffsetMinutes(v *int64) {
	m.UtcOffsetMinutes = v
}

// GetOccurredAt returns the OccurredAt value.
func (m ActivitiesVw) GetOccurredAt() *int64 {
	return m.OccurredAt
}

// SetOccurredAt sets the OccurredAt value and returns the updated ActivitiesVw.
func (m *ActivitiesVw) SetOccurredAt(v *int64) {
	m.OccurredAt = v
}

// GetCreatedAt returns the CreatedAt value.
func (m ActivitiesVw) GetCreatedAt() *int64 {
	return m.CreatedAt
}

// SetCreatedAt sets the CreatedAt value and returns the updated ActivitiesVw.
func (m *ActivitiesVw) SetCreatedAt(v *int64) {
	m.CreatedAt = v
}

// GetUpdatedAt returns the UpdatedAt value.
func (m ActivitiesVw) GetUpdatedAt() *int64 {
	return m.UpdatedAt
}

// SetUpdatedAt sets the UpdatedAt value and returns the updated ActivitiesVw.
func (m *ActivitiesVw) SetUpdatedAt(v *int64) {
	m.UpdatedAt = v
}

// GetCreatedByAppName returns the CreatedByAppName value.
func (m ActivitiesVw) GetCreatedByAppName() *string {
	return m.CreatedByAppName
}

// SetCreatedByAppName sets the CreatedByAppName value and returns the updated ActivitiesVw.
func (m *ActivitiesVw) SetCreatedByAppName(v *string) {
	m.CreatedByAppName = v
}

// GetCreatedByDeviceName returns the CreatedByDeviceName value.
func (m ActivitiesVw) GetCreatedByDeviceName() *string {
	return m.CreatedByDeviceName
}

// SetCreatedByDeviceName sets the CreatedByDeviceName value and returns the updated ActivitiesVw.
func (m *ActivitiesVw) SetCreatedByDeviceName(v *string) {
	m.CreatedByDeviceName = v
}

// GetUpdatedByIdentityID returns the UpdatedByIdentityID value.
func (m ActivitiesVw) GetUpdatedByIdentityID() *int64 {
	return m.UpdatedByIdentityID
}

// SetUpdatedByIdentityID sets the UpdatedByIdentityID value and returns the updated ActivitiesVw.
func (m *ActivitiesVw) SetUpdatedByIdentityID(v *int64) {
	m.UpdatedByIdentityID = v
}

// GetUpdatedByIdentityName returns the UpdatedByIdentityName value.
func (m ActivitiesVw) GetUpdatedByIdentityName() *string {
	return m.UpdatedByIdentityName
}

// SetUpdatedByIdentityName sets the UpdatedByIdentityName value and returns the updated ActivitiesVw.
func (m *ActivitiesVw) SetUpdatedByIdentityName(v *string) {
	m.UpdatedByIdentityName = v
}

// GetCreatedByIdentityID returns the CreatedByIdentityID value.
func (m ActivitiesVw) GetCreatedByIdentityID() *int64 {
	return m.CreatedByIdentityID
}

// SetCreatedByIdentityID sets the CreatedByIdentityID value and returns the updated ActivitiesVw.
func (m *ActivitiesVw) SetCreatedByIdentityID(v *int64) {
	m.CreatedByIdentityID = v
}

// GetCreatedByIdentityName returns the CreatedByIdentityName value.
func (m ActivitiesVw) GetCreatedByIdentityName() *string {
	return m.CreatedByIdentityName
}

// SetCreatedByIdentityName sets the CreatedByIdentityName value and returns the updated ActivitiesVw.
func (m *ActivitiesVw) SetCreatedByIdentityName(v *string) {
	m.CreatedByIdentityName = v
}

// GetID returns the ID value.
func (m Activity) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated Activity.
func (m *Activity) SetID(v string) {
	m.ID = v
}

// GetName returns the Name value.
func (m Activity) GetName() *string {
	return m.Name
}

// SetName sets the Name value and returns the updated Activity.
func (m *Activity) SetName(v *string) {
	m.Name = v
}

// GetActivityType returns the ActivityType value.
func (m Activity) GetActivityType() string {
	return m.ActivityType
}

// SetActivityType sets the ActivityType value and returns the updated Activity.
func (m *Activity) SetActivityType(v string) {
	m.ActivityType = v
}

// GetDurationMinutes returns the DurationMinutes value.
func (m Activity) GetDurationMinutes() *int64 {
	return m.DurationMinutes
}

// SetDurationMinutes sets the DurationMinutes value and returns the updated Activity.
func (m *Activity) SetDurationMinutes(v *int64) {
	m.DurationMinutes = v
}

// GetCalories returns the Calories value.
func (m Activity) GetCalories() *int64 {
	return m.Calories
}

// SetCalories sets the Calories value and returns the updated Activity.
func (m *Activity) SetCalories(v *int64) {
	m.Calories = v
}

// GetSteps returns the Steps value.
func (m Activity) GetSteps() *int64 {
	return m.Steps
}

// SetSteps sets the Steps value and returns the updated Activity.
func (m *Activity) SetSteps(v *int64) {
	m.Steps = v
}

// GetHeartRate returns the HeartRate value.
func (m Activity) GetHeartRate() *int64 {
	return m.HeartRate
}

// SetHeartRate sets the HeartRate value and returns the updated Activity.
func (m *Activity) SetHeartRate(v *int64) {
	m.HeartRate = v
}

// GetNotes returns the Notes value.
func (m Activity) GetNotes() *string {
	return m.Notes
}

// SetNotes sets the Notes value and returns the updated Activity.
func (m *Activity) SetNotes(v *string) {
	m.Notes = v
}

// GetReason returns the Reason value.
func (m Activity) GetReason() *string {
	return m.Reason
}

// SetReason sets the Reason value and returns the updated Activity.
func (m *Activity) SetReason(v *string) {
	m.Reason = v
}

// GetRawJsonData returns the RawJsonData value.
func (m Activity) GetRawJsonData() *string {
	return m.RawJsonData
}

// SetRawJsonData sets the RawJsonData value and returns the updated Activity.
func (m *Activity) SetRawJsonData(v *string) {
	m.RawJsonData = v
}

// GetDeletedAt returns the DeletedAt value.
func (m Activity) GetDeletedAt() *int64 {
	return m.DeletedAt
}

// SetDeletedAt sets the DeletedAt value and returns the updated Activity.
func (m *Activity) SetDeletedAt(v *int64) {
	m.DeletedAt = v
}

// GetUtcOffsetMinutes returns the UtcOffsetMinutes value.
func (m Activity) GetUtcOffsetMinutes() *int64 {
	return m.UtcOffsetMinutes
}

// SetUtcOffsetMinutes sets the UtcOffsetMinutes value and returns the updated Activity.
func (m *Activity) SetUtcOffsetMinutes(v *int64) {
	m.UtcOffsetMinutes = v
}

// GetOccurredAt returns the OccurredAt value.
func (m Activity) GetOccurredAt() int64 {
	return m.OccurredAt
}

// SetOccurredAt sets the OccurredAt value and returns the updated Activity.
func (m *Activity) SetOccurredAt(v int64) {
	m.OccurredAt = v
}

// GetCreatedAt returns the CreatedAt value.
func (m Activity) GetCreatedAt() int64 {
	return m.CreatedAt
}

// SetCreatedAt sets the CreatedAt value and returns the updated Activity.
func (m *Activity) SetCreatedAt(v int64) {
	m.CreatedAt = v
}

// GetUpdatedAt returns the UpdatedAt value.
func (m Activity) GetUpdatedAt() int64 {
	return m.UpdatedAt
}

// SetUpdatedAt sets the UpdatedAt value and returns the updated Activity.
func (m *Activity) SetUpdatedAt(v int64) {
	m.UpdatedAt = v
}

// GetCreatedByAppName returns the CreatedByAppName value.
func (m Activity) GetCreatedByAppName() *string {
	return m.CreatedByAppName
}

// SetCreatedByAppName sets the CreatedByAppName value and returns the updated Activity.
func (m *Activity) SetCreatedByAppName(v *string) {
	m.CreatedByAppName = v
}

// GetCreatedByDeviceName returns the CreatedByDeviceName value.
func (m Activity) GetCreatedByDeviceName() *string {
	return m.CreatedByDeviceName
}

// SetCreatedByDeviceName sets the CreatedByDeviceName value and returns the updated Activity.
func (m *Activity) SetCreatedByDeviceName(v *string) {
	m.CreatedByDeviceName = v
}

// GetUpdatedByIdentityID returns the UpdatedByIdentityID value.
func (m Activity) GetUpdatedByIdentityID() *int64 {
	return m.UpdatedByIdentityID
}

// SetUpdatedByIdentityID sets the UpdatedByIdentityID value and returns the updated Activity.
func (m *Activity) SetUpdatedByIdentityID(v *int64) {
	m.UpdatedByIdentityID = v
}

// GetUpdatedByIdentityName returns the UpdatedByIdentityName value.
func (m Activity) GetUpdatedByIdentityName() *string {
	return m.UpdatedByIdentityName
}

// SetUpdatedByIdentityName sets the UpdatedByIdentityName value and returns the updated Activity.
func (m *Activity) SetUpdatedByIdentityName(v *string) {
	m.UpdatedByIdentityName = v
}

// GetCreatedByIdentityID returns the CreatedByIdentityID value.
func (m Activity) GetCreatedByIdentityID() *int64 {
	return m.CreatedByIdentityID
}

// SetCreatedByIdentityID sets the CreatedByIdentityID value and returns the updated Activity.
func (m *Activity) SetCreatedByIdentityID(v *int64) {
	m.CreatedByIdentityID = v
}

// GetCreatedByIdentityName returns the CreatedByIdentityName value.
func (m Activity) GetCreatedByIdentityName() string {
	return m.CreatedByIdentityName
}

// SetCreatedByIdentityName sets the CreatedByIdentityName value and returns the updated Activity.
func (m *Activity) SetCreatedByIdentityName(v string) {
	m.CreatedByIdentityName = v
}

// GetID returns the ID value.
func (m ComponentEvent) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated ComponentEvent.
func (m *ComponentEvent) SetID(v string) {
	m.ID = v
}

// GetEventType returns the EventType value.
func (m ComponentEvent) GetEventType() string {
	return m.EventType
}

// SetEventType sets the EventType value and returns the updated ComponentEvent.
func (m *ComponentEvent) SetEventType(v string) {
	m.EventType = v
}

// GetSensorCode returns the SensorCode value.
func (m ComponentEvent) GetSensorCode() *string {
	return m.SensorCode
}

// SetSensorCode sets the SensorCode value and returns the updated ComponentEvent.
func (m *ComponentEvent) SetSensorCode(v *string) {
	m.SensorCode = v
}

// GetTransmitterID returns the TransmitterID value.
func (m ComponentEvent) GetTransmitterID() *string {
	return m.TransmitterID
}

// SetTransmitterID sets the TransmitterID value and returns the updated ComponentEvent.
func (m *ComponentEvent) SetTransmitterID(v *string) {
	m.TransmitterID = v
}

// GetPumpID returns the PumpID value.
func (m ComponentEvent) GetPumpID() *string {
	return m.PumpID
}

// SetPumpID sets the PumpID value and returns the updated ComponentEvent.
func (m *ComponentEvent) SetPumpID(v *string) {
	m.PumpID = v
}

// GetPumpSerial returns the PumpSerial value.
func (m ComponentEvent) GetPumpSerial() *string {
	return m.PumpSerial
}

// SetPumpSerial sets the PumpSerial value and returns the updated ComponentEvent.
func (m *ComponentEvent) SetPumpSerial(v *string) {
	m.PumpSerial = v
}

// GetPumpType returns the PumpType value.
func (m ComponentEvent) GetPumpType() *string {
	return m.PumpType
}

// SetPumpType sets the PumpType value and returns the updated ComponentEvent.
func (m *ComponentEvent) SetPumpType(v *string) {
	m.PumpType = v
}

// GetNotes returns the Notes value.
func (m ComponentEvent) GetNotes() *string {
	return m.Notes
}

// SetNotes sets the Notes value and returns the updated ComponentEvent.
func (m *ComponentEvent) SetNotes(v *string) {
	m.Notes = v
}

// GetReason returns the Reason value.
func (m ComponentEvent) GetReason() *string {
	return m.Reason
}

// SetReason sets the Reason value and returns the updated ComponentEvent.
func (m *ComponentEvent) SetReason(v *string) {
	m.Reason = v
}

// GetRawJsonData returns the RawJsonData value.
func (m ComponentEvent) GetRawJsonData() *string {
	return m.RawJsonData
}

// SetRawJsonData sets the RawJsonData value and returns the updated ComponentEvent.
func (m *ComponentEvent) SetRawJsonData(v *string) {
	m.RawJsonData = v
}

// GetDeletedAt returns the DeletedAt value.
func (m ComponentEvent) GetDeletedAt() *int64 {
	return m.DeletedAt
}

// SetDeletedAt sets the DeletedAt value and returns the updated ComponentEvent.
func (m *ComponentEvent) SetDeletedAt(v *int64) {
	m.DeletedAt = v
}

// GetUtcOffsetMinutes returns the UtcOffsetMinutes value.
func (m ComponentEvent) GetUtcOffsetMinutes() *int64 {
	return m.UtcOffsetMinutes
}

// SetUtcOffsetMinutes sets the UtcOffsetMinutes value and returns the updated ComponentEvent.
func (m *ComponentEvent) SetUtcOffsetMinutes(v *int64) {
	m.UtcOffsetMinutes = v
}

// GetOccurredAt returns the OccurredAt value.
func (m ComponentEvent) GetOccurredAt() int64 {
	return m.OccurredAt
}

// SetOccurredAt sets the OccurredAt value and returns the updated ComponentEvent.
func (m *ComponentEvent) SetOccurredAt(v int64) {
	m.OccurredAt = v
}

// GetCreatedAt returns the CreatedAt value.
func (m ComponentEvent) GetCreatedAt() int64 {
	return m.CreatedAt
}

// SetCreatedAt sets the CreatedAt value and returns the updated ComponentEvent.
func (m *ComponentEvent) SetCreatedAt(v int64) {
	m.CreatedAt = v
}

// GetUpdatedAt returns the UpdatedAt value.
func (m ComponentEvent) GetUpdatedAt() int64 {
	return m.UpdatedAt
}

// SetUpdatedAt sets the UpdatedAt value and returns the updated ComponentEvent.
func (m *ComponentEvent) SetUpdatedAt(v int64) {
	m.UpdatedAt = v
}

// GetUpdatedByIdentityID returns the UpdatedByIdentityID value.
func (m ComponentEvent) GetUpdatedByIdentityID() *int64 {
	return m.UpdatedByIdentityID
}

// SetUpdatedByIdentityID sets the UpdatedByIdentityID value and returns the updated ComponentEvent.
func (m *ComponentEvent) SetUpdatedByIdentityID(v *int64) {
	m.UpdatedByIdentityID = v
}

// GetUpdatedByIdentityName returns the UpdatedByIdentityName value.
func (m ComponentEvent) GetUpdatedByIdentityName() *string {
	return m.UpdatedByIdentityName
}

// SetUpdatedByIdentityName sets the UpdatedByIdentityName value and returns the updated ComponentEvent.
func (m *ComponentEvent) SetUpdatedByIdentityName(v *string) {
	m.UpdatedByIdentityName = v
}

// GetCreatedByAppName returns the CreatedByAppName value.
func (m ComponentEvent) GetCreatedByAppName() *string {
	return m.CreatedByAppName
}

// SetCreatedByAppName sets the CreatedByAppName value and returns the updated ComponentEvent.
func (m *ComponentEvent) SetCreatedByAppName(v *string) {
	m.CreatedByAppName = v
}

// GetCreatedByDeviceName returns the CreatedByDeviceName value.
func (m ComponentEvent) GetCreatedByDeviceName() *string {
	return m.CreatedByDeviceName
}

// SetCreatedByDeviceName sets the CreatedByDeviceName value and returns the updated ComponentEvent.
func (m *ComponentEvent) SetCreatedByDeviceName(v *string) {
	m.CreatedByDeviceName = v
}

// GetCreatedByIdentityID returns the CreatedByIdentityID value.
func (m ComponentEvent) GetCreatedByIdentityID() *int64 {
	return m.CreatedByIdentityID
}

// SetCreatedByIdentityID sets the CreatedByIdentityID value and returns the updated ComponentEvent.
func (m *ComponentEvent) SetCreatedByIdentityID(v *int64) {
	m.CreatedByIdentityID = v
}

// GetCreatedByIdentityName returns the CreatedByIdentityName value.
func (m ComponentEvent) GetCreatedByIdentityName() string {
	return m.CreatedByIdentityName
}

// SetCreatedByIdentityName sets the CreatedByIdentityName value and returns the updated ComponentEvent.
func (m *ComponentEvent) SetCreatedByIdentityName(v string) {
	m.CreatedByIdentityName = v
}

// GetID returns the ID value.
func (m ComponentEventsVw) GetID() *string {
	return m.ID
}

// SetID sets the ID value and returns the updated ComponentEventsVw.
func (m *ComponentEventsVw) SetID(v *string) {
	m.ID = v
}

// GetEventType returns the EventType value.
func (m ComponentEventsVw) GetEventType() *string {
	return m.EventType
}

// SetEventType sets the EventType value and returns the updated ComponentEventsVw.
func (m *ComponentEventsVw) SetEventType(v *string) {
	m.EventType = v
}

// GetSensorCode returns the SensorCode value.
func (m ComponentEventsVw) GetSensorCode() *string {
	return m.SensorCode
}

// SetSensorCode sets the SensorCode value and returns the updated ComponentEventsVw.
func (m *ComponentEventsVw) SetSensorCode(v *string) {
	m.SensorCode = v
}

// GetTransmitterID returns the TransmitterID value.
func (m ComponentEventsVw) GetTransmitterID() *string {
	return m.TransmitterID
}

// SetTransmitterID sets the TransmitterID value and returns the updated ComponentEventsVw.
func (m *ComponentEventsVw) SetTransmitterID(v *string) {
	m.TransmitterID = v
}

// GetPumpID returns the PumpID value.
func (m ComponentEventsVw) GetPumpID() *string {
	return m.PumpID
}

// SetPumpID sets the PumpID value and returns the updated ComponentEventsVw.
func (m *ComponentEventsVw) SetPumpID(v *string) {
	m.PumpID = v
}

// GetPumpSerial returns the PumpSerial value.
func (m ComponentEventsVw) GetPumpSerial() *string {
	return m.PumpSerial
}

// SetPumpSerial sets the PumpSerial value and returns the updated ComponentEventsVw.
func (m *ComponentEventsVw) SetPumpSerial(v *string) {
	m.PumpSerial = v
}

// GetPumpType returns the PumpType value.
func (m ComponentEventsVw) GetPumpType() *string {
	return m.PumpType
}

// SetPumpType sets the PumpType value and returns the updated ComponentEventsVw.
func (m *ComponentEventsVw) SetPumpType(v *string) {
	m.PumpType = v
}

// GetNotes returns the Notes value.
func (m ComponentEventsVw) GetNotes() *string {
	return m.Notes
}

// SetNotes sets the Notes value and returns the updated ComponentEventsVw.
func (m *ComponentEventsVw) SetNotes(v *string) {
	m.Notes = v
}

// GetReason returns the Reason value.
func (m ComponentEventsVw) GetReason() *string {
	return m.Reason
}

// SetReason sets the Reason value and returns the updated ComponentEventsVw.
func (m *ComponentEventsVw) SetReason(v *string) {
	m.Reason = v
}

// GetRawJsonData returns the RawJsonData value.
func (m ComponentEventsVw) GetRawJsonData() *string {
	return m.RawJsonData
}

// SetRawJsonData sets the RawJsonData value and returns the updated ComponentEventsVw.
func (m *ComponentEventsVw) SetRawJsonData(v *string) {
	m.RawJsonData = v
}

// GetDeletedAt returns the DeletedAt value.
func (m ComponentEventsVw) GetDeletedAt() *int64 {
	return m.DeletedAt
}

// SetDeletedAt sets the DeletedAt value and returns the updated ComponentEventsVw.
func (m *ComponentEventsVw) SetDeletedAt(v *int64) {
	m.DeletedAt = v
}

// GetUtcOffsetMinutes returns the UtcOffsetMinutes value.
func (m ComponentEventsVw) GetUtcOffsetMinutes() *int64 {
	return m.UtcOffsetMinutes
}

// SetUtcOffsetMinutes sets the UtcOffsetMinutes value and returns the updated ComponentEventsVw.
func (m *ComponentEventsVw) SetUtcOffsetMinutes(v *int64) {
	m.UtcOffsetMinutes = v
}

// GetOccurredAt returns the OccurredAt value.
func (m ComponentEventsVw) GetOccurredAt() *int64 {
	return m.OccurredAt
}

// SetOccurredAt sets the OccurredAt value and returns the updated ComponentEventsVw.
func (m *ComponentEventsVw) SetOccurredAt(v *int64) {
	m.OccurredAt = v
}

// GetCreatedAt returns the CreatedAt value.
func (m ComponentEventsVw) GetCreatedAt() *int64 {
	return m.CreatedAt
}

// SetCreatedAt sets the CreatedAt value and returns the updated ComponentEventsVw.
func (m *ComponentEventsVw) SetCreatedAt(v *int64) {
	m.CreatedAt = v
}

// GetUpdatedAt returns the UpdatedAt value.
func (m ComponentEventsVw) GetUpdatedAt() *int64 {
	return m.UpdatedAt
}

// SetUpdatedAt sets the UpdatedAt value and returns the updated ComponentEventsVw.
func (m *ComponentEventsVw) SetUpdatedAt(v *int64) {
	m.UpdatedAt = v
}

// GetUpdatedByIdentityID returns the UpdatedByIdentityID value.
func (m ComponentEventsVw) GetUpdatedByIdentityID() *int64 {
	return m.UpdatedByIdentityID
}

// SetUpdatedByIdentityID sets the UpdatedByIdentityID value and returns the updated ComponentEventsVw.
func (m *ComponentEventsVw) SetUpdatedByIdentityID(v *int64) {
	m.UpdatedByIdentityID = v
}

// GetUpdatedByIdentityName returns the UpdatedByIdentityName value.
func (m ComponentEventsVw) GetUpdatedByIdentityName() *string {
	return m.UpdatedByIdentityName
}

// SetUpdatedByIdentityName sets the UpdatedByIdentityName value and returns the updated ComponentEventsVw.
func (m *ComponentEventsVw) SetUpdatedByIdentityName(v *string) {
	m.UpdatedByIdentityName = v
}

// GetCreatedByAppName returns the CreatedByAppName value.
func (m ComponentEventsVw) GetCreatedByAppName() *string {
	return m.CreatedByAppName
}

// SetCreatedByAppName sets the CreatedByAppName value and returns the updated ComponentEventsVw.
func (m *ComponentEventsVw) SetCreatedByAppName(v *string) {
	m.CreatedByAppName = v
}

// GetCreatedByDeviceName returns the CreatedByDeviceName value.
func (m ComponentEventsVw) GetCreatedByDeviceName() *string {
	return m.CreatedByDeviceName
}

// SetCreatedByDeviceName sets the CreatedByDeviceName value and returns the updated ComponentEventsVw.
func (m *ComponentEventsVw) SetCreatedByDeviceName(v *string) {
	m.CreatedByDeviceName = v
}

// GetCreatedByIdentityID returns the CreatedByIdentityID value.
func (m ComponentEventsVw) GetCreatedByIdentityID() *int64 {
	return m.CreatedByIdentityID
}

// SetCreatedByIdentityID sets the CreatedByIdentityID value and returns the updated ComponentEventsVw.
func (m *ComponentEventsVw) SetCreatedByIdentityID(v *int64) {
	m.CreatedByIdentityID = v
}

// GetCreatedByIdentityName returns the CreatedByIdentityName value.
func (m ComponentEventsVw) GetCreatedByIdentityName() *string {
	return m.CreatedByIdentityName
}

// SetCreatedByIdentityName sets the CreatedByIdentityName value and returns the updated ComponentEventsVw.
func (m *ComponentEventsVw) SetCreatedByIdentityName(v *string) {
	m.CreatedByIdentityName = v
}

// GetID returns the ID value.
func (m DeviceReport) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated DeviceReport.
func (m *DeviceReport) SetID(v string) {
	m.ID = v
}

// GetDeviceBatteryPercent returns the DeviceBatteryPercent value.
func (m DeviceReport) GetDeviceBatteryPercent() *float64 {
	return m.DeviceBatteryPercent
}

// SetDeviceBatteryPercent sets the DeviceBatteryPercent value and returns the updated DeviceReport.
func (m *DeviceReport) SetDeviceBatteryPercent(v *float64) {
	m.DeviceBatteryPercent = v
}

// GetIsDeviceCharging returns the IsDeviceCharging value.
func (m DeviceReport) GetIsDeviceCharging() *bool {
	return m.IsDeviceCharging
}

// SetIsDeviceCharging sets the IsDeviceCharging value and returns the updated DeviceReport.
func (m *DeviceReport) SetIsDeviceCharging(v *bool) {
	m.IsDeviceCharging = v
}

// GetPumpState returns the PumpState value.
func (m DeviceReport) GetPumpState() *string {
	return m.PumpState
}

// SetPumpState sets the PumpState value and returns the updated DeviceReport.
func (m *DeviceReport) SetPumpState(v *string) {
	m.PumpState = v
}

// GetPumpReservoir returns the PumpReservoir value.
func (m DeviceReport) GetPumpReservoir() *float64 {
	return m.PumpReservoir
}

// SetPumpReservoir sets the PumpReservoir value and returns the updated DeviceReport.
func (m *DeviceReport) SetPumpReservoir(v *float64) {
	m.PumpReservoir = v
}

// GetPumpBatteryPercent returns the PumpBatteryPercent value.
func (m DeviceReport) GetPumpBatteryPercent() *float64 {
	return m.PumpBatteryPercent
}

// SetPumpBatteryPercent sets the PumpBatteryPercent value and returns the updated DeviceReport.
func (m *DeviceReport) SetPumpBatteryPercent(v *float64) {
	m.PumpBatteryPercent = v
}

// GetPumpChargingState returns the PumpChargingState value.
func (m DeviceReport) GetPumpChargingState() *string {
	return m.PumpChargingState
}

// SetPumpChargingState sets the PumpChargingState value and returns the updated DeviceReport.
func (m *DeviceReport) SetPumpChargingState(v *string) {
	m.PumpChargingState = v
}

// GetPumpClock returns the PumpClock value.
func (m DeviceReport) GetPumpClock() *int64 {
	return m.PumpClock
}

// SetPumpClock sets the PumpClock value and returns the updated DeviceReport.
func (m *DeviceReport) SetPumpClock(v *int64) {
	m.PumpClock = v
}

// GetPumpManufacturer returns the PumpManufacturer value.
func (m DeviceReport) GetPumpManufacturer() *string {
	return m.PumpManufacturer
}

// SetPumpManufacturer sets the PumpManufacturer value and returns the updated DeviceReport.
func (m *DeviceReport) SetPumpManufacturer(v *string) {
	m.PumpManufacturer = v
}

// GetPumpModel returns the PumpModel value.
func (m DeviceReport) GetPumpModel() *string {
	return m.PumpModel
}

// SetPumpModel sets the PumpModel value and returns the updated DeviceReport.
func (m *DeviceReport) SetPumpModel(v *string) {
	m.PumpModel = v
}

// GetOverrideJson returns the OverrideJson value.
func (m DeviceReport) GetOverrideJson() *string {
	return m.OverrideJson
}

// SetOverrideJson sets the OverrideJson value and returns the updated DeviceReport.
func (m *DeviceReport) SetOverrideJson(v *string) {
	m.OverrideJson = v
}

// GetLoopName returns the LoopName value.
func (m DeviceReport) GetLoopName() *string {
	return m.LoopName
}

// SetLoopName sets the LoopName value and returns the updated DeviceReport.
func (m *DeviceReport) SetLoopName(v *string) {
	m.LoopName = v
}

// GetConfigurationSettingsJson returns the ConfigurationSettingsJson value.
func (m DeviceReport) GetConfigurationSettingsJson() *string {
	return m.ConfigurationSettingsJson
}

// SetConfigurationSettingsJson sets the ConfigurationSettingsJson value and returns the updated DeviceReport.
func (m *DeviceReport) SetConfigurationSettingsJson(v *string) {
	m.ConfigurationSettingsJson = v
}

// GetLoopSettingsJson returns the LoopSettingsJson value.
func (m DeviceReport) GetLoopSettingsJson() *string {
	return m.LoopSettingsJson
}

// SetLoopSettingsJson sets the LoopSettingsJson value and returns the updated DeviceReport.
func (m *DeviceReport) SetLoopSettingsJson(v *string) {
	m.LoopSettingsJson = v
}

// GetLoopIob returns the LoopIob value.
func (m DeviceReport) GetLoopIob() *float64 {
	return m.LoopIob
}

// SetLoopIob sets the LoopIob value and returns the updated DeviceReport.
func (m *DeviceReport) SetLoopIob(v *float64) {
	m.LoopIob = v
}

// GetLoopIobBasal returns the LoopIobBasal value.
func (m DeviceReport) GetLoopIobBasal() *float64 {
	return m.LoopIobBasal
}

// SetLoopIobBasal sets the LoopIobBasal value and returns the updated DeviceReport.
func (m *DeviceReport) SetLoopIobBasal(v *float64) {
	m.LoopIobBasal = v
}

// GetLoopIobTreatment returns the LoopIobTreatment value.
func (m DeviceReport) GetLoopIobTreatment() *float64 {
	return m.LoopIobTreatment
}

// SetLoopIobTreatment sets the LoopIobTreatment value and returns the updated DeviceReport.
func (m *DeviceReport) SetLoopIobTreatment(v *float64) {
	m.LoopIobTreatment = v
}

// GetLoopCob returns the LoopCob value.
func (m DeviceReport) GetLoopCob() *float64 {
	return m.LoopCob
}

// SetLoopCob sets the LoopCob value and returns the updated DeviceReport.
func (m *DeviceReport) SetLoopCob(v *float64) {
	m.LoopCob = v
}

// GetLoopPredictionsStartTime returns the LoopPredictionsStartTime value.
func (m DeviceReport) GetLoopPredictionsStartTime() *int64 {
	return m.LoopPredictionsStartTime
}

// SetLoopPredictionsStartTime sets the LoopPredictionsStartTime value and returns the updated DeviceReport.
func (m *DeviceReport) SetLoopPredictionsStartTime(v *int64) {
	m.LoopPredictionsStartTime = v
}

// GetLoopPredictionsDefault returns the LoopPredictionsDefault value.
func (m DeviceReport) GetLoopPredictionsDefault() *string {
	return m.LoopPredictionsDefault
}

// SetLoopPredictionsDefault sets the LoopPredictionsDefault value and returns the updated DeviceReport.
func (m *DeviceReport) SetLoopPredictionsDefault(v *string) {
	m.LoopPredictionsDefault = v
}

// GetLoopPredictionsZt returns the LoopPredictionsZt value.
func (m DeviceReport) GetLoopPredictionsZt() *string {
	return m.LoopPredictionsZt
}

// SetLoopPredictionsZt sets the LoopPredictionsZt value and returns the updated DeviceReport.
func (m *DeviceReport) SetLoopPredictionsZt(v *string) {
	m.LoopPredictionsZt = v
}

// GetRawJsonData returns the RawJsonData value.
func (m DeviceReport) GetRawJsonData() *string {
	return m.RawJsonData
}

// SetRawJsonData sets the RawJsonData value and returns the updated DeviceReport.
func (m *DeviceReport) SetRawJsonData(v *string) {
	m.RawJsonData = v
}

// GetDeletedAt returns the DeletedAt value.
func (m DeviceReport) GetDeletedAt() *int64 {
	return m.DeletedAt
}

// SetDeletedAt sets the DeletedAt value and returns the updated DeviceReport.
func (m *DeviceReport) SetDeletedAt(v *int64) {
	m.DeletedAt = v
}

// GetUtcOffsetMinutes returns the UtcOffsetMinutes value.
func (m DeviceReport) GetUtcOffsetMinutes() *int64 {
	return m.UtcOffsetMinutes
}

// SetUtcOffsetMinutes sets the UtcOffsetMinutes value and returns the updated DeviceReport.
func (m *DeviceReport) SetUtcOffsetMinutes(v *int64) {
	m.UtcOffsetMinutes = v
}

// GetOccurredAt returns the OccurredAt value.
func (m DeviceReport) GetOccurredAt() int64 {
	return m.OccurredAt
}

// SetOccurredAt sets the OccurredAt value and returns the updated DeviceReport.
func (m *DeviceReport) SetOccurredAt(v int64) {
	m.OccurredAt = v
}

// GetCreatedAt returns the CreatedAt value.
func (m DeviceReport) GetCreatedAt() int64 {
	return m.CreatedAt
}

// SetCreatedAt sets the CreatedAt value and returns the updated DeviceReport.
func (m *DeviceReport) SetCreatedAt(v int64) {
	m.CreatedAt = v
}

// GetUpdatedAt returns the UpdatedAt value.
func (m DeviceReport) GetUpdatedAt() int64 {
	return m.UpdatedAt
}

// SetUpdatedAt sets the UpdatedAt value and returns the updated DeviceReport.
func (m *DeviceReport) SetUpdatedAt(v int64) {
	m.UpdatedAt = v
}

// GetUpdatedByIdentityID returns the UpdatedByIdentityID value.
func (m DeviceReport) GetUpdatedByIdentityID() *int64 {
	return m.UpdatedByIdentityID
}

// SetUpdatedByIdentityID sets the UpdatedByIdentityID value and returns the updated DeviceReport.
func (m *DeviceReport) SetUpdatedByIdentityID(v *int64) {
	m.UpdatedByIdentityID = v
}

// GetUpdatedByIdentityName returns the UpdatedByIdentityName value.
func (m DeviceReport) GetUpdatedByIdentityName() *string {
	return m.UpdatedByIdentityName
}

// SetUpdatedByIdentityName sets the UpdatedByIdentityName value and returns the updated DeviceReport.
func (m *DeviceReport) SetUpdatedByIdentityName(v *string) {
	m.UpdatedByIdentityName = v
}

// GetCreatedByAppName returns the CreatedByAppName value.
func (m DeviceReport) GetCreatedByAppName() *string {
	return m.CreatedByAppName
}

// SetCreatedByAppName sets the CreatedByAppName value and returns the updated DeviceReport.
func (m *DeviceReport) SetCreatedByAppName(v *string) {
	m.CreatedByAppName = v
}

// GetCreatedByDeviceName returns the CreatedByDeviceName value.
func (m DeviceReport) GetCreatedByDeviceName() *string {
	return m.CreatedByDeviceName
}

// SetCreatedByDeviceName sets the CreatedByDeviceName value and returns the updated DeviceReport.
func (m *DeviceReport) SetCreatedByDeviceName(v *string) {
	m.CreatedByDeviceName = v
}

// GetCreatedByIdentityID returns the CreatedByIdentityID value.
func (m DeviceReport) GetCreatedByIdentityID() *int64 {
	return m.CreatedByIdentityID
}

// SetCreatedByIdentityID sets the CreatedByIdentityID value and returns the updated DeviceReport.
func (m *DeviceReport) SetCreatedByIdentityID(v *int64) {
	m.CreatedByIdentityID = v
}

// GetCreatedByIdentityName returns the CreatedByIdentityName value.
func (m DeviceReport) GetCreatedByIdentityName() string {
	return m.CreatedByIdentityName
}

// SetCreatedByIdentityName sets the CreatedByIdentityName value and returns the updated DeviceReport.
func (m *DeviceReport) SetCreatedByIdentityName(v string) {
	m.CreatedByIdentityName = v
}

// GetID returns the ID value.
func (m EventGroup) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated EventGroup.
func (m *EventGroup) SetID(v string) {
	m.ID = v
}

// GetGroupType returns the GroupType value.
func (m EventGroup) GetGroupType() string {
	return m.GroupType
}

// SetGroupType sets the GroupType value and returns the updated EventGroup.
func (m *EventGroup) SetGroupType(v string) {
	m.GroupType = v
}

// GetGlucoseReadingID returns the GlucoseReadingID value.
func (m EventGroup) GetGlucoseReadingID() *string {
	return m.GlucoseReadingID
}

// SetGlucoseReadingID sets the GlucoseReadingID value and returns the updated EventGroup.
func (m *EventGroup) SetGlucoseReadingID(v *string) {
	m.GlucoseReadingID = v
}

// GetInsulinDoseID returns the InsulinDoseID value.
func (m EventGroup) GetInsulinDoseID() *string {
	return m.InsulinDoseID
}

// SetInsulinDoseID sets the InsulinDoseID value and returns the updated EventGroup.
func (m *EventGroup) SetInsulinDoseID(v *string) {
	m.InsulinDoseID = v
}

// GetFoodIntakeID returns the FoodIntakeID value.
func (m EventGroup) GetFoodIntakeID() *string {
	return m.FoodIntakeID
}

// SetFoodIntakeID sets the FoodIntakeID value and returns the updated EventGroup.
func (m *EventGroup) SetFoodIntakeID(v *string) {
	m.FoodIntakeID = v
}

// GetTempBasalID returns the TempBasalID value.
func (m EventGroup) GetTempBasalID() *string {
	return m.TempBasalID
}

// SetTempBasalID sets the TempBasalID value and returns the updated EventGroup.
func (m *EventGroup) SetTempBasalID(v *string) {
	m.TempBasalID = v
}

// GetTempTargetID returns the TempTargetID value.
func (m EventGroup) GetTempTargetID() *string {
	return m.TempTargetID
}

// SetTempTargetID sets the TempTargetID value and returns the updated EventGroup.
func (m *EventGroup) SetTempTargetID(v *string) {
	m.TempTargetID = v
}

// GetProfileSwitchID returns the ProfileSwitchID value.
func (m EventGroup) GetProfileSwitchID() *string {
	return m.ProfileSwitchID
}

// SetProfileSwitchID sets the ProfileSwitchID value and returns the updated EventGroup.
func (m *EventGroup) SetProfileSwitchID(v *string) {
	m.ProfileSwitchID = v
}

// GetComponentEventID returns the ComponentEventID value.
func (m EventGroup) GetComponentEventID() *string {
	return m.ComponentEventID
}

// SetComponentEventID sets the ComponentEventID value and returns the updated EventGroup.
func (m *EventGroup) SetComponentEventID(v *string) {
	m.ComponentEventID = v
}

// GetActivityID returns the ActivityID value.
func (m EventGroup) GetActivityID() *string {
	return m.ActivityID
}

// SetActivityID sets the ActivityID value and returns the updated EventGroup.
func (m *EventGroup) SetActivityID(v *string) {
	m.ActivityID = v
}

// GetNotificationEventID returns the NotificationEventID value.
func (m EventGroup) GetNotificationEventID() *string {
	return m.NotificationEventID
}

// SetNotificationEventID sets the NotificationEventID value and returns the updated EventGroup.
func (m *EventGroup) SetNotificationEventID(v *string) {
	m.NotificationEventID = v
}

// GetEnteredBy returns the EnteredBy value.
func (m EventGroup) GetEnteredBy() *string {
	return m.EnteredBy
}

// SetEnteredBy sets the EnteredBy value and returns the updated EventGroup.
func (m *EventGroup) SetEnteredBy(v *string) {
	m.EnteredBy = v
}

// GetUuid returns the Uuid value.
func (m EventGroup) GetUuid() *string {
	return m.Uuid
}

// SetUuid sets the Uuid value and returns the updated EventGroup.
func (m *EventGroup) SetUuid(v *string) {
	m.Uuid = v
}

// GetNotes returns the Notes value.
func (m EventGroup) GetNotes() *string {
	return m.Notes
}

// SetNotes sets the Notes value and returns the updated EventGroup.
func (m *EventGroup) SetNotes(v *string) {
	m.Notes = v
}

// GetReason returns the Reason value.
func (m EventGroup) GetReason() *string {
	return m.Reason
}

// SetReason sets the Reason value and returns the updated EventGroup.
func (m *EventGroup) SetReason(v *string) {
	m.Reason = v
}

// GetDurationMinutes returns the DurationMinutes value.
func (m EventGroup) GetDurationMinutes() *int64 {
	return m.DurationMinutes
}

// SetDurationMinutes sets the DurationMinutes value and returns the updated EventGroup.
func (m *EventGroup) SetDurationMinutes(v *int64) {
	m.DurationMinutes = v
}

// GetRawJsonData returns the RawJsonData value.
func (m EventGroup) GetRawJsonData() *string {
	return m.RawJsonData
}

// SetRawJsonData sets the RawJsonData value and returns the updated EventGroup.
func (m *EventGroup) SetRawJsonData(v *string) {
	m.RawJsonData = v
}

// GetDeletedAt returns the DeletedAt value.
func (m EventGroup) GetDeletedAt() *int64 {
	return m.DeletedAt
}

// SetDeletedAt sets the DeletedAt value and returns the updated EventGroup.
func (m *EventGroup) SetDeletedAt(v *int64) {
	m.DeletedAt = v
}

// GetUtcOffsetMinutes returns the UtcOffsetMinutes value.
func (m EventGroup) GetUtcOffsetMinutes() *int64 {
	return m.UtcOffsetMinutes
}

// SetUtcOffsetMinutes sets the UtcOffsetMinutes value and returns the updated EventGroup.
func (m *EventGroup) SetUtcOffsetMinutes(v *int64) {
	m.UtcOffsetMinutes = v
}

// GetOccurredAt returns the OccurredAt value.
func (m EventGroup) GetOccurredAt() int64 {
	return m.OccurredAt
}

// SetOccurredAt sets the OccurredAt value and returns the updated EventGroup.
func (m *EventGroup) SetOccurredAt(v int64) {
	m.OccurredAt = v
}

// GetCreatedAt returns the CreatedAt value.
func (m EventGroup) GetCreatedAt() int64 {
	return m.CreatedAt
}

// SetCreatedAt sets the CreatedAt value and returns the updated EventGroup.
func (m *EventGroup) SetCreatedAt(v int64) {
	m.CreatedAt = v
}

// GetUpdatedAt returns the UpdatedAt value.
func (m EventGroup) GetUpdatedAt() int64 {
	return m.UpdatedAt
}

// SetUpdatedAt sets the UpdatedAt value and returns the updated EventGroup.
func (m *EventGroup) SetUpdatedAt(v int64) {
	m.UpdatedAt = v
}

// GetCreatedByAppName returns the CreatedByAppName value.
func (m EventGroup) GetCreatedByAppName() *string {
	return m.CreatedByAppName
}

// SetCreatedByAppName sets the CreatedByAppName value and returns the updated EventGroup.
func (m *EventGroup) SetCreatedByAppName(v *string) {
	m.CreatedByAppName = v
}

// GetCreatedByDeviceName returns the CreatedByDeviceName value.
func (m EventGroup) GetCreatedByDeviceName() *string {
	return m.CreatedByDeviceName
}

// SetCreatedByDeviceName sets the CreatedByDeviceName value and returns the updated EventGroup.
func (m *EventGroup) SetCreatedByDeviceName(v *string) {
	m.CreatedByDeviceName = v
}

// GetUpdatedByIdentityID returns the UpdatedByIdentityID value.
func (m EventGroup) GetUpdatedByIdentityID() *int64 {
	return m.UpdatedByIdentityID
}

// SetUpdatedByIdentityID sets the UpdatedByIdentityID value and returns the updated EventGroup.
func (m *EventGroup) SetUpdatedByIdentityID(v *int64) {
	m.UpdatedByIdentityID = v
}

// GetUpdatedByIdentityName returns the UpdatedByIdentityName value.
func (m EventGroup) GetUpdatedByIdentityName() *string {
	return m.UpdatedByIdentityName
}

// SetUpdatedByIdentityName sets the UpdatedByIdentityName value and returns the updated EventGroup.
func (m *EventGroup) SetUpdatedByIdentityName(v *string) {
	m.UpdatedByIdentityName = v
}

// GetCreatedByIdentityID returns the CreatedByIdentityID value.
func (m EventGroup) GetCreatedByIdentityID() *int64 {
	return m.CreatedByIdentityID
}

// SetCreatedByIdentityID sets the CreatedByIdentityID value and returns the updated EventGroup.
func (m *EventGroup) SetCreatedByIdentityID(v *int64) {
	m.CreatedByIdentityID = v
}

// GetCreatedByIdentityName returns the CreatedByIdentityName value.
func (m EventGroup) GetCreatedByIdentityName() string {
	return m.CreatedByIdentityName
}

// SetCreatedByIdentityName sets the CreatedByIdentityName value and returns the updated EventGroup.
func (m *EventGroup) SetCreatedByIdentityName(v string) {
	m.CreatedByIdentityName = v
}

// GetID returns the ID value.
func (m FoodIntake) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated FoodIntake.
func (m *FoodIntake) SetID(v string) {
	m.ID = v
}

// GetName returns the Name value.
func (m FoodIntake) GetName() *string {
	return m.Name
}

// SetName sets the Name value and returns the updated FoodIntake.
func (m *FoodIntake) SetName(v *string) {
	m.Name = v
}

// GetFoodType returns the FoodType value.
func (m FoodIntake) GetFoodType() *string {
	return m.FoodType
}

// SetFoodType sets the FoodType value and returns the updated FoodIntake.
func (m *FoodIntake) SetFoodType(v *string) {
	m.FoodType = v
}

// GetCategory returns the Category value.
func (m FoodIntake) GetCategory() *string {
	return m.Category
}

// SetCategory sets the Category value and returns the updated FoodIntake.
func (m *FoodIntake) SetCategory(v *string) {
	m.Category = v
}

// GetSubcategory returns the Subcategory value.
func (m FoodIntake) GetSubcategory() *string {
	return m.Subcategory
}

// SetSubcategory sets the Subcategory value and returns the updated FoodIntake.
func (m *FoodIntake) SetSubcategory(v *string) {
	m.Subcategory = v
}

// GetAbsorptionTimeMinutes returns the AbsorptionTimeMinutes value.
func (m FoodIntake) GetAbsorptionTimeMinutes() *int64 {
	return m.AbsorptionTimeMinutes
}

// SetAbsorptionTimeMinutes sets the AbsorptionTimeMinutes value and returns the updated FoodIntake.
func (m *FoodIntake) SetAbsorptionTimeMinutes(v *int64) {
	m.AbsorptionTimeMinutes = v
}

// GetPreBolus returns the PreBolus value.
func (m FoodIntake) GetPreBolus() *float64 {
	return m.PreBolus
}

// SetPreBolus sets the PreBolus value and returns the updated FoodIntake.
func (m *FoodIntake) SetPreBolus(v *float64) {
	m.PreBolus = v
}

// GetCarbs returns the Carbs value.
func (m FoodIntake) GetCarbs() *float64 {
	return m.Carbs
}

// SetCarbs sets the Carbs value and returns the updated FoodIntake.
func (m *FoodIntake) SetCarbs(v *float64) {
	m.Carbs = v
}

// GetFat returns the Fat value.
func (m FoodIntake) GetFat() *float64 {
	return m.Fat
}

// SetFat sets the Fat value and returns the updated FoodIntake.
func (m *FoodIntake) SetFat(v *float64) {
	m.Fat = v
}

// GetProtein returns the Protein value.
func (m FoodIntake) GetProtein() *float64 {
	return m.Protein
}

// SetProtein sets the Protein value and returns the updated FoodIntake.
func (m *FoodIntake) SetProtein(v *float64) {
	m.Protein = v
}

// GetFibre returns the Fibre value.
func (m FoodIntake) GetFibre() *float64 {
	return m.Fibre
}

// SetFibre sets the Fibre value and returns the updated FoodIntake.
func (m *FoodIntake) SetFibre(v *float64) {
	m.Fibre = v
}

// GetEnergy returns the Energy value.
func (m FoodIntake) GetEnergy() *float64 {
	return m.Energy
}

// SetEnergy sets the Energy value and returns the updated FoodIntake.
func (m *FoodIntake) SetEnergy(v *float64) {
	m.Energy = v
}

// GetPortions returns the Portions value.
func (m FoodIntake) GetPortions() *float64 {
	return m.Portions
}

// SetPortions sets the Portions value and returns the updated FoodIntake.
func (m *FoodIntake) SetPortions(v *float64) {
	m.Portions = v
}

// GetNotes returns the Notes value.
func (m FoodIntake) GetNotes() *string {
	return m.Notes
}

// SetNotes sets the Notes value and returns the updated FoodIntake.
func (m *FoodIntake) SetNotes(v *string) {
	m.Notes = v
}

// GetReason returns the Reason value.
func (m FoodIntake) GetReason() *string {
	return m.Reason
}

// SetReason sets the Reason value and returns the updated FoodIntake.
func (m *FoodIntake) SetReason(v *string) {
	m.Reason = v
}

// GetRawJsonData returns the RawJsonData value.
func (m FoodIntake) GetRawJsonData() *string {
	return m.RawJsonData
}

// SetRawJsonData sets the RawJsonData value and returns the updated FoodIntake.
func (m *FoodIntake) SetRawJsonData(v *string) {
	m.RawJsonData = v
}

// GetDeletedAt returns the DeletedAt value.
func (m FoodIntake) GetDeletedAt() *int64 {
	return m.DeletedAt
}

// SetDeletedAt sets the DeletedAt value and returns the updated FoodIntake.
func (m *FoodIntake) SetDeletedAt(v *int64) {
	m.DeletedAt = v
}

// GetUtcOffsetMinutes returns the UtcOffsetMinutes value.
func (m FoodIntake) GetUtcOffsetMinutes() *int64 {
	return m.UtcOffsetMinutes
}

// SetUtcOffsetMinutes sets the UtcOffsetMinutes value and returns the updated FoodIntake.
func (m *FoodIntake) SetUtcOffsetMinutes(v *int64) {
	m.UtcOffsetMinutes = v
}

// GetOccurredAt returns the OccurredAt value.
func (m FoodIntake) GetOccurredAt() int64 {
	return m.OccurredAt
}

// SetOccurredAt sets the OccurredAt value and returns the updated FoodIntake.
func (m *FoodIntake) SetOccurredAt(v int64) {
	m.OccurredAt = v
}

// GetCreatedAt returns the CreatedAt value.
func (m FoodIntake) GetCreatedAt() int64 {
	return m.CreatedAt
}

// SetCreatedAt sets the CreatedAt value and returns the updated FoodIntake.
func (m *FoodIntake) SetCreatedAt(v int64) {
	m.CreatedAt = v
}

// GetUpdatedAt returns the UpdatedAt value.
func (m FoodIntake) GetUpdatedAt() int64 {
	return m.UpdatedAt
}

// SetUpdatedAt sets the UpdatedAt value and returns the updated FoodIntake.
func (m *FoodIntake) SetUpdatedAt(v int64) {
	m.UpdatedAt = v
}

// GetCreatedByAppName returns the CreatedByAppName value.
func (m FoodIntake) GetCreatedByAppName() *string {
	return m.CreatedByAppName
}

// SetCreatedByAppName sets the CreatedByAppName value and returns the updated FoodIntake.
func (m *FoodIntake) SetCreatedByAppName(v *string) {
	m.CreatedByAppName = v
}

// GetCreatedByDeviceName returns the CreatedByDeviceName value.
func (m FoodIntake) GetCreatedByDeviceName() *string {
	return m.CreatedByDeviceName
}

// SetCreatedByDeviceName sets the CreatedByDeviceName value and returns the updated FoodIntake.
func (m *FoodIntake) SetCreatedByDeviceName(v *string) {
	m.CreatedByDeviceName = v
}

// GetUpdatedByIdentityID returns the UpdatedByIdentityID value.
func (m FoodIntake) GetUpdatedByIdentityID() *int64 {
	return m.UpdatedByIdentityID
}

// SetUpdatedByIdentityID sets the UpdatedByIdentityID value and returns the updated FoodIntake.
func (m *FoodIntake) SetUpdatedByIdentityID(v *int64) {
	m.UpdatedByIdentityID = v
}

// GetUpdatedByIdentityName returns the UpdatedByIdentityName value.
func (m FoodIntake) GetUpdatedByIdentityName() *string {
	return m.UpdatedByIdentityName
}

// SetUpdatedByIdentityName sets the UpdatedByIdentityName value and returns the updated FoodIntake.
func (m *FoodIntake) SetUpdatedByIdentityName(v *string) {
	m.UpdatedByIdentityName = v
}

// GetCreatedByIdentityID returns the CreatedByIdentityID value.
func (m FoodIntake) GetCreatedByIdentityID() *int64 {
	return m.CreatedByIdentityID
}

// SetCreatedByIdentityID sets the CreatedByIdentityID value and returns the updated FoodIntake.
func (m *FoodIntake) SetCreatedByIdentityID(v *int64) {
	m.CreatedByIdentityID = v
}

// GetCreatedByIdentityName returns the CreatedByIdentityName value.
func (m FoodIntake) GetCreatedByIdentityName() string {
	return m.CreatedByIdentityName
}

// SetCreatedByIdentityName sets the CreatedByIdentityName value and returns the updated FoodIntake.
func (m *FoodIntake) SetCreatedByIdentityName(v string) {
	m.CreatedByIdentityName = v
}

// GetID returns the ID value.
func (m FoodIntakesVw) GetID() *string {
	return m.ID
}

// SetID sets the ID value and returns the updated FoodIntakesVw.
func (m *FoodIntakesVw) SetID(v *string) {
	m.ID = v
}

// GetName returns the Name value.
func (m FoodIntakesVw) GetName() *string {
	return m.Name
}

// SetName sets the Name value and returns the updated FoodIntakesVw.
func (m *FoodIntakesVw) SetName(v *string) {
	m.Name = v
}

// GetFoodType returns the FoodType value.
func (m FoodIntakesVw) GetFoodType() *string {
	return m.FoodType
}

// SetFoodType sets the FoodType value and returns the updated FoodIntakesVw.
func (m *FoodIntakesVw) SetFoodType(v *string) {
	m.FoodType = v
}

// GetCategory returns the Category value.
func (m FoodIntakesVw) GetCategory() *string {
	return m.Category
}

// SetCategory sets the Category value and returns the updated FoodIntakesVw.
func (m *FoodIntakesVw) SetCategory(v *string) {
	m.Category = v
}

// GetSubcategory returns the Subcategory value.
func (m FoodIntakesVw) GetSubcategory() *string {
	return m.Subcategory
}

// SetSubcategory sets the Subcategory value and returns the updated FoodIntakesVw.
func (m *FoodIntakesVw) SetSubcategory(v *string) {
	m.Subcategory = v
}

// GetAbsorptionTimeMinutes returns the AbsorptionTimeMinutes value.
func (m FoodIntakesVw) GetAbsorptionTimeMinutes() *int64 {
	return m.AbsorptionTimeMinutes
}

// SetAbsorptionTimeMinutes sets the AbsorptionTimeMinutes value and returns the updated FoodIntakesVw.
func (m *FoodIntakesVw) SetAbsorptionTimeMinutes(v *int64) {
	m.AbsorptionTimeMinutes = v
}

// GetPreBolus returns the PreBolus value.
func (m FoodIntakesVw) GetPreBolus() *float64 {
	return m.PreBolus
}

// SetPreBolus sets the PreBolus value and returns the updated FoodIntakesVw.
func (m *FoodIntakesVw) SetPreBolus(v *float64) {
	m.PreBolus = v
}

// GetCarbs returns the Carbs value.
func (m FoodIntakesVw) GetCarbs() *float64 {
	return m.Carbs
}

// SetCarbs sets the Carbs value and returns the updated FoodIntakesVw.
func (m *FoodIntakesVw) SetCarbs(v *float64) {
	m.Carbs = v
}

// GetFat returns the Fat value.
func (m FoodIntakesVw) GetFat() *float64 {
	return m.Fat
}

// SetFat sets the Fat value and returns the updated FoodIntakesVw.
func (m *FoodIntakesVw) SetFat(v *float64) {
	m.Fat = v
}

// GetProtein returns the Protein value.
func (m FoodIntakesVw) GetProtein() *float64 {
	return m.Protein
}

// SetProtein sets the Protein value and returns the updated FoodIntakesVw.
func (m *FoodIntakesVw) SetProtein(v *float64) {
	m.Protein = v
}

// GetFibre returns the Fibre value.
func (m FoodIntakesVw) GetFibre() *float64 {
	return m.Fibre
}

// SetFibre sets the Fibre value and returns the updated FoodIntakesVw.
func (m *FoodIntakesVw) SetFibre(v *float64) {
	m.Fibre = v
}

// GetEnergy returns the Energy value.
func (m FoodIntakesVw) GetEnergy() *float64 {
	return m.Energy
}

// SetEnergy sets the Energy value and returns the updated FoodIntakesVw.
func (m *FoodIntakesVw) SetEnergy(v *float64) {
	m.Energy = v
}

// GetPortions returns the Portions value.
func (m FoodIntakesVw) GetPortions() *float64 {
	return m.Portions
}

// SetPortions sets the Portions value and returns the updated FoodIntakesVw.
func (m *FoodIntakesVw) SetPortions(v *float64) {
	m.Portions = v
}

// GetNotes returns the Notes value.
func (m FoodIntakesVw) GetNotes() *string {
	return m.Notes
}

// SetNotes sets the Notes value and returns the updated FoodIntakesVw.
func (m *FoodIntakesVw) SetNotes(v *string) {
	m.Notes = v
}

// GetReason returns the Reason value.
func (m FoodIntakesVw) GetReason() *string {
	return m.Reason
}

// SetReason sets the Reason value and returns the updated FoodIntakesVw.
func (m *FoodIntakesVw) SetReason(v *string) {
	m.Reason = v
}

// GetRawJsonData returns the RawJsonData value.
func (m FoodIntakesVw) GetRawJsonData() *string {
	return m.RawJsonData
}

// SetRawJsonData sets the RawJsonData value and returns the updated FoodIntakesVw.
func (m *FoodIntakesVw) SetRawJsonData(v *string) {
	m.RawJsonData = v
}

// GetDeletedAt returns the DeletedAt value.
func (m FoodIntakesVw) GetDeletedAt() *int64 {
	return m.DeletedAt
}

// SetDeletedAt sets the DeletedAt value and returns the updated FoodIntakesVw.
func (m *FoodIntakesVw) SetDeletedAt(v *int64) {
	m.DeletedAt = v
}

// GetUtcOffsetMinutes returns the UtcOffsetMinutes value.
func (m FoodIntakesVw) GetUtcOffsetMinutes() *int64 {
	return m.UtcOffsetMinutes
}

// SetUtcOffsetMinutes sets the UtcOffsetMinutes value and returns the updated FoodIntakesVw.
func (m *FoodIntakesVw) SetUtcOffsetMinutes(v *int64) {
	m.UtcOffsetMinutes = v
}

// GetOccurredAt returns the OccurredAt value.
func (m FoodIntakesVw) GetOccurredAt() *int64 {
	return m.OccurredAt
}

// SetOccurredAt sets the OccurredAt value and returns the updated FoodIntakesVw.
func (m *FoodIntakesVw) SetOccurredAt(v *int64) {
	m.OccurredAt = v
}

// GetCreatedAt returns the CreatedAt value.
func (m FoodIntakesVw) GetCreatedAt() *int64 {
	return m.CreatedAt
}

// SetCreatedAt sets the CreatedAt value and returns the updated FoodIntakesVw.
func (m *FoodIntakesVw) SetCreatedAt(v *int64) {
	m.CreatedAt = v
}

// GetUpdatedAt returns the UpdatedAt value.
func (m FoodIntakesVw) GetUpdatedAt() *int64 {
	return m.UpdatedAt
}

// SetUpdatedAt sets the UpdatedAt value and returns the updated FoodIntakesVw.
func (m *FoodIntakesVw) SetUpdatedAt(v *int64) {
	m.UpdatedAt = v
}

// GetCreatedByAppName returns the CreatedByAppName value.
func (m FoodIntakesVw) GetCreatedByAppName() *string {
	return m.CreatedByAppName
}

// SetCreatedByAppName sets the CreatedByAppName value and returns the updated FoodIntakesVw.
func (m *FoodIntakesVw) SetCreatedByAppName(v *string) {
	m.CreatedByAppName = v
}

// GetCreatedByDeviceName returns the CreatedByDeviceName value.
func (m FoodIntakesVw) GetCreatedByDeviceName() *string {
	return m.CreatedByDeviceName
}

// SetCreatedByDeviceName sets the CreatedByDeviceName value and returns the updated FoodIntakesVw.
func (m *FoodIntakesVw) SetCreatedByDeviceName(v *string) {
	m.CreatedByDeviceName = v
}

// GetUpdatedByIdentityID returns the UpdatedByIdentityID value.
func (m FoodIntakesVw) GetUpdatedByIdentityID() *int64 {
	return m.UpdatedByIdentityID
}

// SetUpdatedByIdentityID sets the UpdatedByIdentityID value and returns the updated FoodIntakesVw.
func (m *FoodIntakesVw) SetUpdatedByIdentityID(v *int64) {
	m.UpdatedByIdentityID = v
}

// GetUpdatedByIdentityName returns the UpdatedByIdentityName value.
func (m FoodIntakesVw) GetUpdatedByIdentityName() *string {
	return m.UpdatedByIdentityName
}

// SetUpdatedByIdentityName sets the UpdatedByIdentityName value and returns the updated FoodIntakesVw.
func (m *FoodIntakesVw) SetUpdatedByIdentityName(v *string) {
	m.UpdatedByIdentityName = v
}

// GetCreatedByIdentityID returns the CreatedByIdentityID value.
func (m FoodIntakesVw) GetCreatedByIdentityID() *int64 {
	return m.CreatedByIdentityID
}

// SetCreatedByIdentityID sets the CreatedByIdentityID value and returns the updated FoodIntakesVw.
func (m *FoodIntakesVw) SetCreatedByIdentityID(v *int64) {
	m.CreatedByIdentityID = v
}

// GetCreatedByIdentityName returns the CreatedByIdentityName value.
func (m FoodIntakesVw) GetCreatedByIdentityName() *string {
	return m.CreatedByIdentityName
}

// SetCreatedByIdentityName sets the CreatedByIdentityName value and returns the updated FoodIntakesVw.
func (m *FoodIntakesVw) SetCreatedByIdentityName(v *string) {
	m.CreatedByIdentityName = v
}

// GetID returns the ID value.
func (m FoodPreset) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated FoodPreset.
func (m *FoodPreset) SetID(v string) {
	m.ID = v
}

// GetName returns the Name value.
func (m FoodPreset) GetName() string {
	return m.Name
}

// SetName sets the Name value and returns the updated FoodPreset.
func (m *FoodPreset) SetName(v string) {
	m.Name = v
}

// GetFoodType returns the FoodType value.
func (m FoodPreset) GetFoodType() *string {
	return m.FoodType
}

// SetFoodType sets the FoodType value and returns the updated FoodPreset.
func (m *FoodPreset) SetFoodType(v *string) {
	m.FoodType = v
}

// GetCategory returns the Category value.
func (m FoodPreset) GetCategory() *string {
	return m.Category
}

// SetCategory sets the Category value and returns the updated FoodPreset.
func (m *FoodPreset) SetCategory(v *string) {
	m.Category = v
}

// GetSubcategory returns the Subcategory value.
func (m FoodPreset) GetSubcategory() *string {
	return m.Subcategory
}

// SetSubcategory sets the Subcategory value and returns the updated FoodPreset.
func (m *FoodPreset) SetSubcategory(v *string) {
	m.Subcategory = v
}

// GetAbsorptionTimeMinutes returns the AbsorptionTimeMinutes value.
func (m FoodPreset) GetAbsorptionTimeMinutes() *int64 {
	return m.AbsorptionTimeMinutes
}

// SetAbsorptionTimeMinutes sets the AbsorptionTimeMinutes value and returns the updated FoodPreset.
func (m *FoodPreset) SetAbsorptionTimeMinutes(v *int64) {
	m.AbsorptionTimeMinutes = v
}

// GetPreBolus returns the PreBolus value.
func (m FoodPreset) GetPreBolus() *float64 {
	return m.PreBolus
}

// SetPreBolus sets the PreBolus value and returns the updated FoodPreset.
func (m *FoodPreset) SetPreBolus(v *float64) {
	m.PreBolus = v
}

// GetCarbs returns the Carbs value.
func (m FoodPreset) GetCarbs() *float64 {
	return m.Carbs
}

// SetCarbs sets the Carbs value and returns the updated FoodPreset.
func (m *FoodPreset) SetCarbs(v *float64) {
	m.Carbs = v
}

// GetFat returns the Fat value.
func (m FoodPreset) GetFat() *float64 {
	return m.Fat
}

// SetFat sets the Fat value and returns the updated FoodPreset.
func (m *FoodPreset) SetFat(v *float64) {
	m.Fat = v
}

// GetProtein returns the Protein value.
func (m FoodPreset) GetProtein() *float64 {
	return m.Protein
}

// SetProtein sets the Protein value and returns the updated FoodPreset.
func (m *FoodPreset) SetProtein(v *float64) {
	m.Protein = v
}

// GetFibre returns the Fibre value.
func (m FoodPreset) GetFibre() *float64 {
	return m.Fibre
}

// SetFibre sets the Fibre value and returns the updated FoodPreset.
func (m *FoodPreset) SetFibre(v *float64) {
	m.Fibre = v
}

// GetEnergy returns the Energy value.
func (m FoodPreset) GetEnergy() *float64 {
	return m.Energy
}

// SetEnergy sets the Energy value and returns the updated FoodPreset.
func (m *FoodPreset) SetEnergy(v *float64) {
	m.Energy = v
}

// GetPortions returns the Portions value.
func (m FoodPreset) GetPortions() *float64 {
	return m.Portions
}

// SetPortions sets the Portions value and returns the updated FoodPreset.
func (m *FoodPreset) SetPortions(v *float64) {
	m.Portions = v
}

// GetRawJsonData returns the RawJsonData value.
func (m FoodPreset) GetRawJsonData() *string {
	return m.RawJsonData
}

// SetRawJsonData sets the RawJsonData value and returns the updated FoodPreset.
func (m *FoodPreset) SetRawJsonData(v *string) {
	m.RawJsonData = v
}

// GetCreatedAt returns the CreatedAt value.
func (m FoodPreset) GetCreatedAt() int64 {
	return m.CreatedAt
}

// SetCreatedAt sets the CreatedAt value and returns the updated FoodPreset.
func (m *FoodPreset) SetCreatedAt(v int64) {
	m.CreatedAt = v
}

// GetUpdatedAt returns the UpdatedAt value.
func (m FoodPreset) GetUpdatedAt() int64 {
	return m.UpdatedAt
}

// SetUpdatedAt sets the UpdatedAt value and returns the updated FoodPreset.
func (m *FoodPreset) SetUpdatedAt(v int64) {
	m.UpdatedAt = v
}

// GetCreatedByAppName returns the CreatedByAppName value.
func (m FoodPreset) GetCreatedByAppName() *string {
	return m.CreatedByAppName
}

// SetCreatedByAppName sets the CreatedByAppName value and returns the updated FoodPreset.
func (m *FoodPreset) SetCreatedByAppName(v *string) {
	m.CreatedByAppName = v
}

// GetCreatedByDeviceName returns the CreatedByDeviceName value.
func (m FoodPreset) GetCreatedByDeviceName() *string {
	return m.CreatedByDeviceName
}

// SetCreatedByDeviceName sets the CreatedByDeviceName value and returns the updated FoodPreset.
func (m *FoodPreset) SetCreatedByDeviceName(v *string) {
	m.CreatedByDeviceName = v
}

// GetUpdatedByIdentityID returns the UpdatedByIdentityID value.
func (m FoodPreset) GetUpdatedByIdentityID() *int64 {
	return m.UpdatedByIdentityID
}

// SetUpdatedByIdentityID sets the UpdatedByIdentityID value and returns the updated FoodPreset.
func (m *FoodPreset) SetUpdatedByIdentityID(v *int64) {
	m.UpdatedByIdentityID = v
}

// GetUpdatedByIdentityName returns the UpdatedByIdentityName value.
func (m FoodPreset) GetUpdatedByIdentityName() *string {
	return m.UpdatedByIdentityName
}

// SetUpdatedByIdentityName sets the UpdatedByIdentityName value and returns the updated FoodPreset.
func (m *FoodPreset) SetUpdatedByIdentityName(v *string) {
	m.UpdatedByIdentityName = v
}

// GetCreatedByIdentityID returns the CreatedByIdentityID value.
func (m FoodPreset) GetCreatedByIdentityID() *int64 {
	return m.CreatedByIdentityID
}

// SetCreatedByIdentityID sets the CreatedByIdentityID value and returns the updated FoodPreset.
func (m *FoodPreset) SetCreatedByIdentityID(v *int64) {
	m.CreatedByIdentityID = v
}

// GetCreatedByIdentityName returns the CreatedByIdentityName value.
func (m FoodPreset) GetCreatedByIdentityName() string {
	return m.CreatedByIdentityName
}

// SetCreatedByIdentityName sets the CreatedByIdentityName value and returns the updated FoodPreset.
func (m *FoodPreset) SetCreatedByIdentityName(v string) {
	m.CreatedByIdentityName = v
}

// GetID returns the ID value.
func (m GlucoseReading) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated GlucoseReading.
func (m *GlucoseReading) SetID(v string) {
	m.ID = v
}

// GetMgdl returns the Mgdl value.
func (m GlucoseReading) GetMgdl() float64 {
	return m.Mgdl
}

// SetMgdl sets the Mgdl value and returns the updated GlucoseReading.
func (m *GlucoseReading) SetMgdl(v float64) {
	m.Mgdl = v
}

// GetReadingType returns the ReadingType value.
func (m GlucoseReading) GetReadingType() string {
	return m.ReadingType
}

// SetReadingType sets the ReadingType value and returns the updated GlucoseReading.
func (m *GlucoseReading) SetReadingType(v string) {
	m.ReadingType = v
}

// GetUnits returns the Units value.
func (m GlucoseReading) GetUnits() *consts.Units {
	return m.Units
}

// SetUnits sets the Units value and returns the updated GlucoseReading.
func (m *GlucoseReading) SetUnits(v *consts.Units) {
	m.Units = v
}

// GetMeasurementMethod returns the MeasurementMethod value.
func (m GlucoseReading) GetMeasurementMethod() *string {
	return m.MeasurementMethod
}

// SetMeasurementMethod sets the MeasurementMethod value and returns the updated GlucoseReading.
func (m *GlucoseReading) SetMeasurementMethod(v *string) {
	m.MeasurementMethod = v
}

// GetDelta returns the Delta value.
func (m GlucoseReading) GetDelta() *float64 {
	return m.Delta
}

// SetDelta sets the Delta value and returns the updated GlucoseReading.
func (m *GlucoseReading) SetDelta(v *float64) {
	m.Delta = v
}

// GetDirection returns the Direction value.
func (m GlucoseReading) GetDirection() *string {
	return m.Direction
}

// SetDirection sets the Direction value and returns the updated GlucoseReading.
func (m *GlucoseReading) SetDirection(v *string) {
	m.Direction = v
}

// GetNoise returns the Noise value.
func (m GlucoseReading) GetNoise() *float64 {
	return m.Noise
}

// SetNoise sets the Noise value and returns the updated GlucoseReading.
func (m *GlucoseReading) SetNoise(v *float64) {
	m.Noise = v
}

// GetFiltered returns the Filtered value.
func (m GlucoseReading) GetFiltered() *float64 {
	return m.Filtered
}

// SetFiltered sets the Filtered value and returns the updated GlucoseReading.
func (m *GlucoseReading) SetFiltered(v *float64) {
	m.Filtered = v
}

// GetUnfiltered returns the Unfiltered value.
func (m GlucoseReading) GetUnfiltered() *float64 {
	return m.Unfiltered
}

// SetUnfiltered sets the Unfiltered value and returns the updated GlucoseReading.
func (m *GlucoseReading) SetUnfiltered(v *float64) {
	m.Unfiltered = v
}

// GetRssi returns the Rssi value.
func (m GlucoseReading) GetRssi() *float64 {
	return m.Rssi
}

// SetRssi sets the Rssi value and returns the updated GlucoseReading.
func (m *GlucoseReading) SetRssi(v *float64) {
	m.Rssi = v
}

// GetNotes returns the Notes value.
func (m GlucoseReading) GetNotes() *string {
	return m.Notes
}

// SetNotes sets the Notes value and returns the updated GlucoseReading.
func (m *GlucoseReading) SetNotes(v *string) {
	m.Notes = v
}

// GetReason returns the Reason value.
func (m GlucoseReading) GetReason() *string {
	return m.Reason
}

// SetReason sets the Reason value and returns the updated GlucoseReading.
func (m *GlucoseReading) SetReason(v *string) {
	m.Reason = v
}

// GetRawJsonData returns the RawJsonData value.
func (m GlucoseReading) GetRawJsonData() *string {
	return m.RawJsonData
}

// SetRawJsonData sets the RawJsonData value and returns the updated GlucoseReading.
func (m *GlucoseReading) SetRawJsonData(v *string) {
	m.RawJsonData = v
}

// GetDeletedAt returns the DeletedAt value.
func (m GlucoseReading) GetDeletedAt() *int64 {
	return m.DeletedAt
}

// SetDeletedAt sets the DeletedAt value and returns the updated GlucoseReading.
func (m *GlucoseReading) SetDeletedAt(v *int64) {
	m.DeletedAt = v
}

// GetUtcOffsetMinutes returns the UtcOffsetMinutes value.
func (m GlucoseReading) GetUtcOffsetMinutes() *int64 {
	return m.UtcOffsetMinutes
}

// SetUtcOffsetMinutes sets the UtcOffsetMinutes value and returns the updated GlucoseReading.
func (m *GlucoseReading) SetUtcOffsetMinutes(v *int64) {
	m.UtcOffsetMinutes = v
}

// GetOccurredAt returns the OccurredAt value.
func (m GlucoseReading) GetOccurredAt() int64 {
	return m.OccurredAt
}

// SetOccurredAt sets the OccurredAt value and returns the updated GlucoseReading.
func (m *GlucoseReading) SetOccurredAt(v int64) {
	m.OccurredAt = v
}

// GetCreatedAt returns the CreatedAt value.
func (m GlucoseReading) GetCreatedAt() int64 {
	return m.CreatedAt
}

// SetCreatedAt sets the CreatedAt value and returns the updated GlucoseReading.
func (m *GlucoseReading) SetCreatedAt(v int64) {
	m.CreatedAt = v
}

// GetUpdatedAt returns the UpdatedAt value.
func (m GlucoseReading) GetUpdatedAt() int64 {
	return m.UpdatedAt
}

// SetUpdatedAt sets the UpdatedAt value and returns the updated GlucoseReading.
func (m *GlucoseReading) SetUpdatedAt(v int64) {
	m.UpdatedAt = v
}

// GetCreatedByAppName returns the CreatedByAppName value.
func (m GlucoseReading) GetCreatedByAppName() *string {
	return m.CreatedByAppName
}

// SetCreatedByAppName sets the CreatedByAppName value and returns the updated GlucoseReading.
func (m *GlucoseReading) SetCreatedByAppName(v *string) {
	m.CreatedByAppName = v
}

// GetCreatedByDeviceName returns the CreatedByDeviceName value.
func (m GlucoseReading) GetCreatedByDeviceName() *string {
	return m.CreatedByDeviceName
}

// SetCreatedByDeviceName sets the CreatedByDeviceName value and returns the updated GlucoseReading.
func (m *GlucoseReading) SetCreatedByDeviceName(v *string) {
	m.CreatedByDeviceName = v
}

// GetUpdatedByIdentityID returns the UpdatedByIdentityID value.
func (m GlucoseReading) GetUpdatedByIdentityID() *int64 {
	return m.UpdatedByIdentityID
}

// SetUpdatedByIdentityID sets the UpdatedByIdentityID value and returns the updated GlucoseReading.
func (m *GlucoseReading) SetUpdatedByIdentityID(v *int64) {
	m.UpdatedByIdentityID = v
}

// GetUpdatedByIdentityName returns the UpdatedByIdentityName value.
func (m GlucoseReading) GetUpdatedByIdentityName() *string {
	return m.UpdatedByIdentityName
}

// SetUpdatedByIdentityName sets the UpdatedByIdentityName value and returns the updated GlucoseReading.
func (m *GlucoseReading) SetUpdatedByIdentityName(v *string) {
	m.UpdatedByIdentityName = v
}

// GetCreatedByIdentityID returns the CreatedByIdentityID value.
func (m GlucoseReading) GetCreatedByIdentityID() *int64 {
	return m.CreatedByIdentityID
}

// SetCreatedByIdentityID sets the CreatedByIdentityID value and returns the updated GlucoseReading.
func (m *GlucoseReading) SetCreatedByIdentityID(v *int64) {
	m.CreatedByIdentityID = v
}

// GetCreatedByIdentityName returns the CreatedByIdentityName value.
func (m GlucoseReading) GetCreatedByIdentityName() string {
	return m.CreatedByIdentityName
}

// SetCreatedByIdentityName sets the CreatedByIdentityName value and returns the updated GlucoseReading.
func (m *GlucoseReading) SetCreatedByIdentityName(v string) {
	m.CreatedByIdentityName = v
}

// GetID returns the ID value.
func (m GlucoseReadingsVw) GetID() *string {
	return m.ID
}

// SetID sets the ID value and returns the updated GlucoseReadingsVw.
func (m *GlucoseReadingsVw) SetID(v *string) {
	m.ID = v
}

// GetMgdl returns the Mgdl value.
func (m GlucoseReadingsVw) GetMgdl() *float64 {
	return m.Mgdl
}

// SetMgdl sets the Mgdl value and returns the updated GlucoseReadingsVw.
func (m *GlucoseReadingsVw) SetMgdl(v *float64) {
	m.Mgdl = v
}

// GetReadingType returns the ReadingType value.
func (m GlucoseReadingsVw) GetReadingType() *string {
	return m.ReadingType
}

// SetReadingType sets the ReadingType value and returns the updated GlucoseReadingsVw.
func (m *GlucoseReadingsVw) SetReadingType(v *string) {
	m.ReadingType = v
}

// GetUnits returns the Units value.
func (m GlucoseReadingsVw) GetUnits() *consts.Units {
	return m.Units
}

// SetUnits sets the Units value and returns the updated GlucoseReadingsVw.
func (m *GlucoseReadingsVw) SetUnits(v *consts.Units) {
	m.Units = v
}

// GetMeasurementMethod returns the MeasurementMethod value.
func (m GlucoseReadingsVw) GetMeasurementMethod() *string {
	return m.MeasurementMethod
}

// SetMeasurementMethod sets the MeasurementMethod value and returns the updated GlucoseReadingsVw.
func (m *GlucoseReadingsVw) SetMeasurementMethod(v *string) {
	m.MeasurementMethod = v
}

// GetDelta returns the Delta value.
func (m GlucoseReadingsVw) GetDelta() *float64 {
	return m.Delta
}

// SetDelta sets the Delta value and returns the updated GlucoseReadingsVw.
func (m *GlucoseReadingsVw) SetDelta(v *float64) {
	m.Delta = v
}

// GetDirection returns the Direction value.
func (m GlucoseReadingsVw) GetDirection() *string {
	return m.Direction
}

// SetDirection sets the Direction value and returns the updated GlucoseReadingsVw.
func (m *GlucoseReadingsVw) SetDirection(v *string) {
	m.Direction = v
}

// GetNoise returns the Noise value.
func (m GlucoseReadingsVw) GetNoise() *float64 {
	return m.Noise
}

// SetNoise sets the Noise value and returns the updated GlucoseReadingsVw.
func (m *GlucoseReadingsVw) SetNoise(v *float64) {
	m.Noise = v
}

// GetFiltered returns the Filtered value.
func (m GlucoseReadingsVw) GetFiltered() *float64 {
	return m.Filtered
}

// SetFiltered sets the Filtered value and returns the updated GlucoseReadingsVw.
func (m *GlucoseReadingsVw) SetFiltered(v *float64) {
	m.Filtered = v
}

// GetUnfiltered returns the Unfiltered value.
func (m GlucoseReadingsVw) GetUnfiltered() *float64 {
	return m.Unfiltered
}

// SetUnfiltered sets the Unfiltered value and returns the updated GlucoseReadingsVw.
func (m *GlucoseReadingsVw) SetUnfiltered(v *float64) {
	m.Unfiltered = v
}

// GetRssi returns the Rssi value.
func (m GlucoseReadingsVw) GetRssi() *float64 {
	return m.Rssi
}

// SetRssi sets the Rssi value and returns the updated GlucoseReadingsVw.
func (m *GlucoseReadingsVw) SetRssi(v *float64) {
	m.Rssi = v
}

// GetNotes returns the Notes value.
func (m GlucoseReadingsVw) GetNotes() *string {
	return m.Notes
}

// SetNotes sets the Notes value and returns the updated GlucoseReadingsVw.
func (m *GlucoseReadingsVw) SetNotes(v *string) {
	m.Notes = v
}

// GetReason returns the Reason value.
func (m GlucoseReadingsVw) GetReason() *string {
	return m.Reason
}

// SetReason sets the Reason value and returns the updated GlucoseReadingsVw.
func (m *GlucoseReadingsVw) SetReason(v *string) {
	m.Reason = v
}

// GetRawJsonData returns the RawJsonData value.
func (m GlucoseReadingsVw) GetRawJsonData() *string {
	return m.RawJsonData
}

// SetRawJsonData sets the RawJsonData value and returns the updated GlucoseReadingsVw.
func (m *GlucoseReadingsVw) SetRawJsonData(v *string) {
	m.RawJsonData = v
}

// GetDeletedAt returns the DeletedAt value.
func (m GlucoseReadingsVw) GetDeletedAt() *int64 {
	return m.DeletedAt
}

// SetDeletedAt sets the DeletedAt value and returns the updated GlucoseReadingsVw.
func (m *GlucoseReadingsVw) SetDeletedAt(v *int64) {
	m.DeletedAt = v
}

// GetUtcOffsetMinutes returns the UtcOffsetMinutes value.
func (m GlucoseReadingsVw) GetUtcOffsetMinutes() *int64 {
	return m.UtcOffsetMinutes
}

// SetUtcOffsetMinutes sets the UtcOffsetMinutes value and returns the updated GlucoseReadingsVw.
func (m *GlucoseReadingsVw) SetUtcOffsetMinutes(v *int64) {
	m.UtcOffsetMinutes = v
}

// GetOccurredAt returns the OccurredAt value.
func (m GlucoseReadingsVw) GetOccurredAt() *int64 {
	return m.OccurredAt
}

// SetOccurredAt sets the OccurredAt value and returns the updated GlucoseReadingsVw.
func (m *GlucoseReadingsVw) SetOccurredAt(v *int64) {
	m.OccurredAt = v
}

// GetCreatedAt returns the CreatedAt value.
func (m GlucoseReadingsVw) GetCreatedAt() *int64 {
	return m.CreatedAt
}

// SetCreatedAt sets the CreatedAt value and returns the updated GlucoseReadingsVw.
func (m *GlucoseReadingsVw) SetCreatedAt(v *int64) {
	m.CreatedAt = v
}

// GetUpdatedAt returns the UpdatedAt value.
func (m GlucoseReadingsVw) GetUpdatedAt() *int64 {
	return m.UpdatedAt
}

// SetUpdatedAt sets the UpdatedAt value and returns the updated GlucoseReadingsVw.
func (m *GlucoseReadingsVw) SetUpdatedAt(v *int64) {
	m.UpdatedAt = v
}

// GetCreatedByAppName returns the CreatedByAppName value.
func (m GlucoseReadingsVw) GetCreatedByAppName() *string {
	return m.CreatedByAppName
}

// SetCreatedByAppName sets the CreatedByAppName value and returns the updated GlucoseReadingsVw.
func (m *GlucoseReadingsVw) SetCreatedByAppName(v *string) {
	m.CreatedByAppName = v
}

// GetCreatedByDeviceName returns the CreatedByDeviceName value.
func (m GlucoseReadingsVw) GetCreatedByDeviceName() *string {
	return m.CreatedByDeviceName
}

// SetCreatedByDeviceName sets the CreatedByDeviceName value and returns the updated GlucoseReadingsVw.
func (m *GlucoseReadingsVw) SetCreatedByDeviceName(v *string) {
	m.CreatedByDeviceName = v
}

// GetUpdatedByIdentityID returns the UpdatedByIdentityID value.
func (m GlucoseReadingsVw) GetUpdatedByIdentityID() *int64 {
	return m.UpdatedByIdentityID
}

// SetUpdatedByIdentityID sets the UpdatedByIdentityID value and returns the updated GlucoseReadingsVw.
func (m *GlucoseReadingsVw) SetUpdatedByIdentityID(v *int64) {
	m.UpdatedByIdentityID = v
}

// GetUpdatedByIdentityName returns the UpdatedByIdentityName value.
func (m GlucoseReadingsVw) GetUpdatedByIdentityName() *string {
	return m.UpdatedByIdentityName
}

// SetUpdatedByIdentityName sets the UpdatedByIdentityName value and returns the updated GlucoseReadingsVw.
func (m *GlucoseReadingsVw) SetUpdatedByIdentityName(v *string) {
	m.UpdatedByIdentityName = v
}

// GetCreatedByIdentityID returns the CreatedByIdentityID value.
func (m GlucoseReadingsVw) GetCreatedByIdentityID() *int64 {
	return m.CreatedByIdentityID
}

// SetCreatedByIdentityID sets the CreatedByIdentityID value and returns the updated GlucoseReadingsVw.
func (m *GlucoseReadingsVw) SetCreatedByIdentityID(v *int64) {
	m.CreatedByIdentityID = v
}

// GetCreatedByIdentityName returns the CreatedByIdentityName value.
func (m GlucoseReadingsVw) GetCreatedByIdentityName() *string {
	return m.CreatedByIdentityName
}

// SetCreatedByIdentityName sets the CreatedByIdentityName value and returns the updated GlucoseReadingsVw.
func (m *GlucoseReadingsVw) SetCreatedByIdentityName(v *string) {
	m.CreatedByIdentityName = v
}

// GetID returns the ID value.
func (m InsulinDose) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated InsulinDose.
func (m *InsulinDose) SetID(v string) {
	m.ID = v
}

// GetDose returns the Dose value.
func (m InsulinDose) GetDose() float64 {
	return m.Dose
}

// SetDose sets the Dose value and returns the updated InsulinDose.
func (m *InsulinDose) SetDose(v float64) {
	m.Dose = v
}

// GetDoseType returns the DoseType value.
func (m InsulinDose) GetDoseType() *string {
	return m.DoseType
}

// SetDoseType sets the DoseType value and returns the updated InsulinDose.
func (m *InsulinDose) SetDoseType(v *string) {
	m.DoseType = v
}

// GetSplitNow returns the SplitNow value.
func (m InsulinDose) GetSplitNow() *float64 {
	return m.SplitNow
}

// SetSplitNow sets the SplitNow value and returns the updated InsulinDose.
func (m *InsulinDose) SetSplitNow(v *float64) {
	m.SplitNow = v
}

// GetSplitExt returns the SplitExt value.
func (m InsulinDose) GetSplitExt() *float64 {
	return m.SplitExt
}

// SetSplitExt sets the SplitExt value and returns the updated InsulinDose.
func (m *InsulinDose) SetSplitExt(v *float64) {
	m.SplitExt = v
}

// GetNotes returns the Notes value.
func (m InsulinDose) GetNotes() *string {
	return m.Notes
}

// SetNotes sets the Notes value and returns the updated InsulinDose.
func (m *InsulinDose) SetNotes(v *string) {
	m.Notes = v
}

// GetReason returns the Reason value.
func (m InsulinDose) GetReason() *string {
	return m.Reason
}

// SetReason sets the Reason value and returns the updated InsulinDose.
func (m *InsulinDose) SetReason(v *string) {
	m.Reason = v
}

// GetRawJsonData returns the RawJsonData value.
func (m InsulinDose) GetRawJsonData() *string {
	return m.RawJsonData
}

// SetRawJsonData sets the RawJsonData value and returns the updated InsulinDose.
func (m *InsulinDose) SetRawJsonData(v *string) {
	m.RawJsonData = v
}

// GetDeletedAt returns the DeletedAt value.
func (m InsulinDose) GetDeletedAt() *int64 {
	return m.DeletedAt
}

// SetDeletedAt sets the DeletedAt value and returns the updated InsulinDose.
func (m *InsulinDose) SetDeletedAt(v *int64) {
	m.DeletedAt = v
}

// GetUtcOffsetMinutes returns the UtcOffsetMinutes value.
func (m InsulinDose) GetUtcOffsetMinutes() *int64 {
	return m.UtcOffsetMinutes
}

// SetUtcOffsetMinutes sets the UtcOffsetMinutes value and returns the updated InsulinDose.
func (m *InsulinDose) SetUtcOffsetMinutes(v *int64) {
	m.UtcOffsetMinutes = v
}

// GetOccurredAt returns the OccurredAt value.
func (m InsulinDose) GetOccurredAt() int64 {
	return m.OccurredAt
}

// SetOccurredAt sets the OccurredAt value and returns the updated InsulinDose.
func (m *InsulinDose) SetOccurredAt(v int64) {
	m.OccurredAt = v
}

// GetCreatedAt returns the CreatedAt value.
func (m InsulinDose) GetCreatedAt() int64 {
	return m.CreatedAt
}

// SetCreatedAt sets the CreatedAt value and returns the updated InsulinDose.
func (m *InsulinDose) SetCreatedAt(v int64) {
	m.CreatedAt = v
}

// GetUpdatedAt returns the UpdatedAt value.
func (m InsulinDose) GetUpdatedAt() int64 {
	return m.UpdatedAt
}

// SetUpdatedAt sets the UpdatedAt value and returns the updated InsulinDose.
func (m *InsulinDose) SetUpdatedAt(v int64) {
	m.UpdatedAt = v
}

// GetCreatedByAppName returns the CreatedByAppName value.
func (m InsulinDose) GetCreatedByAppName() *string {
	return m.CreatedByAppName
}

// SetCreatedByAppName sets the CreatedByAppName value and returns the updated InsulinDose.
func (m *InsulinDose) SetCreatedByAppName(v *string) {
	m.CreatedByAppName = v
}

// GetCreatedByDeviceName returns the CreatedByDeviceName value.
func (m InsulinDose) GetCreatedByDeviceName() *string {
	return m.CreatedByDeviceName
}

// SetCreatedByDeviceName sets the CreatedByDeviceName value and returns the updated InsulinDose.
func (m *InsulinDose) SetCreatedByDeviceName(v *string) {
	m.CreatedByDeviceName = v
}

// GetUpdatedByIdentityID returns the UpdatedByIdentityID value.
func (m InsulinDose) GetUpdatedByIdentityID() *int64 {
	return m.UpdatedByIdentityID
}

// SetUpdatedByIdentityID sets the UpdatedByIdentityID value and returns the updated InsulinDose.
func (m *InsulinDose) SetUpdatedByIdentityID(v *int64) {
	m.UpdatedByIdentityID = v
}

// GetUpdatedByIdentityName returns the UpdatedByIdentityName value.
func (m InsulinDose) GetUpdatedByIdentityName() *string {
	return m.UpdatedByIdentityName
}

// SetUpdatedByIdentityName sets the UpdatedByIdentityName value and returns the updated InsulinDose.
func (m *InsulinDose) SetUpdatedByIdentityName(v *string) {
	m.UpdatedByIdentityName = v
}

// GetCreatedByIdentityID returns the CreatedByIdentityID value.
func (m InsulinDose) GetCreatedByIdentityID() *int64 {
	return m.CreatedByIdentityID
}

// SetCreatedByIdentityID sets the CreatedByIdentityID value and returns the updated InsulinDose.
func (m *InsulinDose) SetCreatedByIdentityID(v *int64) {
	m.CreatedByIdentityID = v
}

// GetCreatedByIdentityName returns the CreatedByIdentityName value.
func (m InsulinDose) GetCreatedByIdentityName() string {
	return m.CreatedByIdentityName
}

// SetCreatedByIdentityName sets the CreatedByIdentityName value and returns the updated InsulinDose.
func (m *InsulinDose) SetCreatedByIdentityName(v string) {
	m.CreatedByIdentityName = v
}

// GetID returns the ID value.
func (m InsulinDosesVw) GetID() *string {
	return m.ID
}

// SetID sets the ID value and returns the updated InsulinDosesVw.
func (m *InsulinDosesVw) SetID(v *string) {
	m.ID = v
}

// GetDose returns the Dose value.
func (m InsulinDosesVw) GetDose() *float64 {
	return m.Dose
}

// SetDose sets the Dose value and returns the updated InsulinDosesVw.
func (m *InsulinDosesVw) SetDose(v *float64) {
	m.Dose = v
}

// GetDoseType returns the DoseType value.
func (m InsulinDosesVw) GetDoseType() *string {
	return m.DoseType
}

// SetDoseType sets the DoseType value and returns the updated InsulinDosesVw.
func (m *InsulinDosesVw) SetDoseType(v *string) {
	m.DoseType = v
}

// GetSplitNow returns the SplitNow value.
func (m InsulinDosesVw) GetSplitNow() *float64 {
	return m.SplitNow
}

// SetSplitNow sets the SplitNow value and returns the updated InsulinDosesVw.
func (m *InsulinDosesVw) SetSplitNow(v *float64) {
	m.SplitNow = v
}

// GetSplitExt returns the SplitExt value.
func (m InsulinDosesVw) GetSplitExt() *float64 {
	return m.SplitExt
}

// SetSplitExt sets the SplitExt value and returns the updated InsulinDosesVw.
func (m *InsulinDosesVw) SetSplitExt(v *float64) {
	m.SplitExt = v
}

// GetNotes returns the Notes value.
func (m InsulinDosesVw) GetNotes() *string {
	return m.Notes
}

// SetNotes sets the Notes value and returns the updated InsulinDosesVw.
func (m *InsulinDosesVw) SetNotes(v *string) {
	m.Notes = v
}

// GetReason returns the Reason value.
func (m InsulinDosesVw) GetReason() *string {
	return m.Reason
}

// SetReason sets the Reason value and returns the updated InsulinDosesVw.
func (m *InsulinDosesVw) SetReason(v *string) {
	m.Reason = v
}

// GetRawJsonData returns the RawJsonData value.
func (m InsulinDosesVw) GetRawJsonData() *string {
	return m.RawJsonData
}

// SetRawJsonData sets the RawJsonData value and returns the updated InsulinDosesVw.
func (m *InsulinDosesVw) SetRawJsonData(v *string) {
	m.RawJsonData = v
}

// GetDeletedAt returns the DeletedAt value.
func (m InsulinDosesVw) GetDeletedAt() *int64 {
	return m.DeletedAt
}

// SetDeletedAt sets the DeletedAt value and returns the updated InsulinDosesVw.
func (m *InsulinDosesVw) SetDeletedAt(v *int64) {
	m.DeletedAt = v
}

// GetUtcOffsetMinutes returns the UtcOffsetMinutes value.
func (m InsulinDosesVw) GetUtcOffsetMinutes() *int64 {
	return m.UtcOffsetMinutes
}

// SetUtcOffsetMinutes sets the UtcOffsetMinutes value and returns the updated InsulinDosesVw.
func (m *InsulinDosesVw) SetUtcOffsetMinutes(v *int64) {
	m.UtcOffsetMinutes = v
}

// GetOccurredAt returns the OccurredAt value.
func (m InsulinDosesVw) GetOccurredAt() *int64 {
	return m.OccurredAt
}

// SetOccurredAt sets the OccurredAt value and returns the updated InsulinDosesVw.
func (m *InsulinDosesVw) SetOccurredAt(v *int64) {
	m.OccurredAt = v
}

// GetCreatedAt returns the CreatedAt value.
func (m InsulinDosesVw) GetCreatedAt() *int64 {
	return m.CreatedAt
}

// SetCreatedAt sets the CreatedAt value and returns the updated InsulinDosesVw.
func (m *InsulinDosesVw) SetCreatedAt(v *int64) {
	m.CreatedAt = v
}

// GetUpdatedAt returns the UpdatedAt value.
func (m InsulinDosesVw) GetUpdatedAt() *int64 {
	return m.UpdatedAt
}

// SetUpdatedAt sets the UpdatedAt value and returns the updated InsulinDosesVw.
func (m *InsulinDosesVw) SetUpdatedAt(v *int64) {
	m.UpdatedAt = v
}

// GetCreatedByAppName returns the CreatedByAppName value.
func (m InsulinDosesVw) GetCreatedByAppName() *string {
	return m.CreatedByAppName
}

// SetCreatedByAppName sets the CreatedByAppName value and returns the updated InsulinDosesVw.
func (m *InsulinDosesVw) SetCreatedByAppName(v *string) {
	m.CreatedByAppName = v
}

// GetCreatedByDeviceName returns the CreatedByDeviceName value.
func (m InsulinDosesVw) GetCreatedByDeviceName() *string {
	return m.CreatedByDeviceName
}

// SetCreatedByDeviceName sets the CreatedByDeviceName value and returns the updated InsulinDosesVw.
func (m *InsulinDosesVw) SetCreatedByDeviceName(v *string) {
	m.CreatedByDeviceName = v
}

// GetUpdatedByIdentityID returns the UpdatedByIdentityID value.
func (m InsulinDosesVw) GetUpdatedByIdentityID() *int64 {
	return m.UpdatedByIdentityID
}

// SetUpdatedByIdentityID sets the UpdatedByIdentityID value and returns the updated InsulinDosesVw.
func (m *InsulinDosesVw) SetUpdatedByIdentityID(v *int64) {
	m.UpdatedByIdentityID = v
}

// GetUpdatedByIdentityName returns the UpdatedByIdentityName value.
func (m InsulinDosesVw) GetUpdatedByIdentityName() *string {
	return m.UpdatedByIdentityName
}

// SetUpdatedByIdentityName sets the UpdatedByIdentityName value and returns the updated InsulinDosesVw.
func (m *InsulinDosesVw) SetUpdatedByIdentityName(v *string) {
	m.UpdatedByIdentityName = v
}

// GetCreatedByIdentityID returns the CreatedByIdentityID value.
func (m InsulinDosesVw) GetCreatedByIdentityID() *int64 {
	return m.CreatedByIdentityID
}

// SetCreatedByIdentityID sets the CreatedByIdentityID value and returns the updated InsulinDosesVw.
func (m *InsulinDosesVw) SetCreatedByIdentityID(v *int64) {
	m.CreatedByIdentityID = v
}

// GetCreatedByIdentityName returns the CreatedByIdentityName value.
func (m InsulinDosesVw) GetCreatedByIdentityName() *string {
	return m.CreatedByIdentityName
}

// SetCreatedByIdentityName sets the CreatedByIdentityName value and returns the updated InsulinDosesVw.
func (m *InsulinDosesVw) SetCreatedByIdentityName(v *string) {
	m.CreatedByIdentityName = v
}

// GetID returns the ID value.
func (m NotificationEvent) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated NotificationEvent.
func (m *NotificationEvent) SetID(v string) {
	m.ID = v
}

// GetNotificationType returns the NotificationType value.
func (m NotificationEvent) GetNotificationType() string {
	return m.NotificationType
}

// SetNotificationType sets the NotificationType value and returns the updated NotificationEvent.
func (m *NotificationEvent) SetNotificationType(v string) {
	m.NotificationType = v
}

// GetTitle returns the Title value.
func (m NotificationEvent) GetTitle() string {
	return m.Title
}

// SetTitle sets the Title value and returns the updated NotificationEvent.
func (m *NotificationEvent) SetTitle(v string) {
	m.Title = v
}

// GetSeverity returns the Severity value.
func (m NotificationEvent) GetSeverity() int64 {
	return m.Severity
}

// SetSeverity sets the Severity value and returns the updated NotificationEvent.
func (m *NotificationEvent) SetSeverity(v int64) {
	m.Severity = v
}

// GetContent returns the Content value.
func (m NotificationEvent) GetContent() *string {
	return m.Content
}

// SetContent sets the Content value and returns the updated NotificationEvent.
func (m *NotificationEvent) SetContent(v *string) {
	m.Content = v
}

// GetNotes returns the Notes value.
func (m NotificationEvent) GetNotes() *string {
	return m.Notes
}

// SetNotes sets the Notes value and returns the updated NotificationEvent.
func (m *NotificationEvent) SetNotes(v *string) {
	m.Notes = v
}

// GetReason returns the Reason value.
func (m NotificationEvent) GetReason() *string {
	return m.Reason
}

// SetReason sets the Reason value and returns the updated NotificationEvent.
func (m *NotificationEvent) SetReason(v *string) {
	m.Reason = v
}

// GetRawJsonData returns the RawJsonData value.
func (m NotificationEvent) GetRawJsonData() *string {
	return m.RawJsonData
}

// SetRawJsonData sets the RawJsonData value and returns the updated NotificationEvent.
func (m *NotificationEvent) SetRawJsonData(v *string) {
	m.RawJsonData = v
}

// GetDeletedAt returns the DeletedAt value.
func (m NotificationEvent) GetDeletedAt() *int64 {
	return m.DeletedAt
}

// SetDeletedAt sets the DeletedAt value and returns the updated NotificationEvent.
func (m *NotificationEvent) SetDeletedAt(v *int64) {
	m.DeletedAt = v
}

// GetUtcOffsetMinutes returns the UtcOffsetMinutes value.
func (m NotificationEvent) GetUtcOffsetMinutes() *int64 {
	return m.UtcOffsetMinutes
}

// SetUtcOffsetMinutes sets the UtcOffsetMinutes value and returns the updated NotificationEvent.
func (m *NotificationEvent) SetUtcOffsetMinutes(v *int64) {
	m.UtcOffsetMinutes = v
}

// GetOccurredAt returns the OccurredAt value.
func (m NotificationEvent) GetOccurredAt() int64 {
	return m.OccurredAt
}

// SetOccurredAt sets the OccurredAt value and returns the updated NotificationEvent.
func (m *NotificationEvent) SetOccurredAt(v int64) {
	m.OccurredAt = v
}

// GetCreatedAt returns the CreatedAt value.
func (m NotificationEvent) GetCreatedAt() int64 {
	return m.CreatedAt
}

// SetCreatedAt sets the CreatedAt value and returns the updated NotificationEvent.
func (m *NotificationEvent) SetCreatedAt(v int64) {
	m.CreatedAt = v
}

// GetUpdatedAt returns the UpdatedAt value.
func (m NotificationEvent) GetUpdatedAt() int64 {
	return m.UpdatedAt
}

// SetUpdatedAt sets the UpdatedAt value and returns the updated NotificationEvent.
func (m *NotificationEvent) SetUpdatedAt(v int64) {
	m.UpdatedAt = v
}

// GetCreatedByAppName returns the CreatedByAppName value.
func (m NotificationEvent) GetCreatedByAppName() *string {
	return m.CreatedByAppName
}

// SetCreatedByAppName sets the CreatedByAppName value and returns the updated NotificationEvent.
func (m *NotificationEvent) SetCreatedByAppName(v *string) {
	m.CreatedByAppName = v
}

// GetCreatedByDeviceName returns the CreatedByDeviceName value.
func (m NotificationEvent) GetCreatedByDeviceName() *string {
	return m.CreatedByDeviceName
}

// SetCreatedByDeviceName sets the CreatedByDeviceName value and returns the updated NotificationEvent.
func (m *NotificationEvent) SetCreatedByDeviceName(v *string) {
	m.CreatedByDeviceName = v
}

// GetUpdatedByIdentityID returns the UpdatedByIdentityID value.
func (m NotificationEvent) GetUpdatedByIdentityID() *int64 {
	return m.UpdatedByIdentityID
}

// SetUpdatedByIdentityID sets the UpdatedByIdentityID value and returns the updated NotificationEvent.
func (m *NotificationEvent) SetUpdatedByIdentityID(v *int64) {
	m.UpdatedByIdentityID = v
}

// GetUpdatedByIdentityName returns the UpdatedByIdentityName value.
func (m NotificationEvent) GetUpdatedByIdentityName() *string {
	return m.UpdatedByIdentityName
}

// SetUpdatedByIdentityName sets the UpdatedByIdentityName value and returns the updated NotificationEvent.
func (m *NotificationEvent) SetUpdatedByIdentityName(v *string) {
	m.UpdatedByIdentityName = v
}

// GetCreatedByIdentityID returns the CreatedByIdentityID value.
func (m NotificationEvent) GetCreatedByIdentityID() *int64 {
	return m.CreatedByIdentityID
}

// SetCreatedByIdentityID sets the CreatedByIdentityID value and returns the updated NotificationEvent.
func (m *NotificationEvent) SetCreatedByIdentityID(v *int64) {
	m.CreatedByIdentityID = v
}

// GetCreatedByIdentityName returns the CreatedByIdentityName value.
func (m NotificationEvent) GetCreatedByIdentityName() string {
	return m.CreatedByIdentityName
}

// SetCreatedByIdentityName sets the CreatedByIdentityName value and returns the updated NotificationEvent.
func (m *NotificationEvent) SetCreatedByIdentityName(v string) {
	m.CreatedByIdentityName = v
}

// GetID returns the ID value.
func (m NotificationEventsVw) GetID() *string {
	return m.ID
}

// SetID sets the ID value and returns the updated NotificationEventsVw.
func (m *NotificationEventsVw) SetID(v *string) {
	m.ID = v
}

// GetNotificationType returns the NotificationType value.
func (m NotificationEventsVw) GetNotificationType() *string {
	return m.NotificationType
}

// SetNotificationType sets the NotificationType value and returns the updated NotificationEventsVw.
func (m *NotificationEventsVw) SetNotificationType(v *string) {
	m.NotificationType = v
}

// GetTitle returns the Title value.
func (m NotificationEventsVw) GetTitle() *string {
	return m.Title
}

// SetTitle sets the Title value and returns the updated NotificationEventsVw.
func (m *NotificationEventsVw) SetTitle(v *string) {
	m.Title = v
}

// GetSeverity returns the Severity value.
func (m NotificationEventsVw) GetSeverity() *int64 {
	return m.Severity
}

// SetSeverity sets the Severity value and returns the updated NotificationEventsVw.
func (m *NotificationEventsVw) SetSeverity(v *int64) {
	m.Severity = v
}

// GetContent returns the Content value.
func (m NotificationEventsVw) GetContent() *string {
	return m.Content
}

// SetContent sets the Content value and returns the updated NotificationEventsVw.
func (m *NotificationEventsVw) SetContent(v *string) {
	m.Content = v
}

// GetNotes returns the Notes value.
func (m NotificationEventsVw) GetNotes() *string {
	return m.Notes
}

// SetNotes sets the Notes value and returns the updated NotificationEventsVw.
func (m *NotificationEventsVw) SetNotes(v *string) {
	m.Notes = v
}

// GetReason returns the Reason value.
func (m NotificationEventsVw) GetReason() *string {
	return m.Reason
}

// SetReason sets the Reason value and returns the updated NotificationEventsVw.
func (m *NotificationEventsVw) SetReason(v *string) {
	m.Reason = v
}

// GetRawJsonData returns the RawJsonData value.
func (m NotificationEventsVw) GetRawJsonData() *string {
	return m.RawJsonData
}

// SetRawJsonData sets the RawJsonData value and returns the updated NotificationEventsVw.
func (m *NotificationEventsVw) SetRawJsonData(v *string) {
	m.RawJsonData = v
}

// GetDeletedAt returns the DeletedAt value.
func (m NotificationEventsVw) GetDeletedAt() *int64 {
	return m.DeletedAt
}

// SetDeletedAt sets the DeletedAt value and returns the updated NotificationEventsVw.
func (m *NotificationEventsVw) SetDeletedAt(v *int64) {
	m.DeletedAt = v
}

// GetUtcOffsetMinutes returns the UtcOffsetMinutes value.
func (m NotificationEventsVw) GetUtcOffsetMinutes() *int64 {
	return m.UtcOffsetMinutes
}

// SetUtcOffsetMinutes sets the UtcOffsetMinutes value and returns the updated NotificationEventsVw.
func (m *NotificationEventsVw) SetUtcOffsetMinutes(v *int64) {
	m.UtcOffsetMinutes = v
}

// GetOccurredAt returns the OccurredAt value.
func (m NotificationEventsVw) GetOccurredAt() *int64 {
	return m.OccurredAt
}

// SetOccurredAt sets the OccurredAt value and returns the updated NotificationEventsVw.
func (m *NotificationEventsVw) SetOccurredAt(v *int64) {
	m.OccurredAt = v
}

// GetCreatedAt returns the CreatedAt value.
func (m NotificationEventsVw) GetCreatedAt() *int64 {
	return m.CreatedAt
}

// SetCreatedAt sets the CreatedAt value and returns the updated NotificationEventsVw.
func (m *NotificationEventsVw) SetCreatedAt(v *int64) {
	m.CreatedAt = v
}

// GetUpdatedAt returns the UpdatedAt value.
func (m NotificationEventsVw) GetUpdatedAt() *int64 {
	return m.UpdatedAt
}

// SetUpdatedAt sets the UpdatedAt value and returns the updated NotificationEventsVw.
func (m *NotificationEventsVw) SetUpdatedAt(v *int64) {
	m.UpdatedAt = v
}

// GetCreatedByAppName returns the CreatedByAppName value.
func (m NotificationEventsVw) GetCreatedByAppName() *string {
	return m.CreatedByAppName
}

// SetCreatedByAppName sets the CreatedByAppName value and returns the updated NotificationEventsVw.
func (m *NotificationEventsVw) SetCreatedByAppName(v *string) {
	m.CreatedByAppName = v
}

// GetCreatedByDeviceName returns the CreatedByDeviceName value.
func (m NotificationEventsVw) GetCreatedByDeviceName() *string {
	return m.CreatedByDeviceName
}

// SetCreatedByDeviceName sets the CreatedByDeviceName value and returns the updated NotificationEventsVw.
func (m *NotificationEventsVw) SetCreatedByDeviceName(v *string) {
	m.CreatedByDeviceName = v
}

// GetUpdatedByIdentityID returns the UpdatedByIdentityID value.
func (m NotificationEventsVw) GetUpdatedByIdentityID() *int64 {
	return m.UpdatedByIdentityID
}

// SetUpdatedByIdentityID sets the UpdatedByIdentityID value and returns the updated NotificationEventsVw.
func (m *NotificationEventsVw) SetUpdatedByIdentityID(v *int64) {
	m.UpdatedByIdentityID = v
}

// GetUpdatedByIdentityName returns the UpdatedByIdentityName value.
func (m NotificationEventsVw) GetUpdatedByIdentityName() *string {
	return m.UpdatedByIdentityName
}

// SetUpdatedByIdentityName sets the UpdatedByIdentityName value and returns the updated NotificationEventsVw.
func (m *NotificationEventsVw) SetUpdatedByIdentityName(v *string) {
	m.UpdatedByIdentityName = v
}

// GetCreatedByIdentityID returns the CreatedByIdentityID value.
func (m NotificationEventsVw) GetCreatedByIdentityID() *int64 {
	return m.CreatedByIdentityID
}

// SetCreatedByIdentityID sets the CreatedByIdentityID value and returns the updated NotificationEventsVw.
func (m *NotificationEventsVw) SetCreatedByIdentityID(v *int64) {
	m.CreatedByIdentityID = v
}

// GetCreatedByIdentityName returns the CreatedByIdentityName value.
func (m NotificationEventsVw) GetCreatedByIdentityName() *string {
	return m.CreatedByIdentityName
}

// SetCreatedByIdentityName sets the CreatedByIdentityName value and returns the updated NotificationEventsVw.
func (m *NotificationEventsVw) SetCreatedByIdentityName(v *string) {
	m.CreatedByIdentityName = v
}

// GetID returns the ID value.
func (m ProfilePreset) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated ProfilePreset.
func (m *ProfilePreset) SetID(v string) {
	m.ID = v
}

// GetName returns the Name value.
func (m ProfilePreset) GetName() string {
	return m.Name
}

// SetName sets the Name value and returns the updated ProfilePreset.
func (m *ProfilePreset) SetName(v string) {
	m.Name = v
}

// GetProfileStoreID returns the ProfileStoreID value.
func (m ProfilePreset) GetProfileStoreID() string {
	return m.ProfileStoreID
}

// SetProfileStoreID sets the ProfileStoreID value and returns the updated ProfilePreset.
func (m *ProfilePreset) SetProfileStoreID(v string) {
	m.ProfileStoreID = v
}

// GetIsDefault returns the IsDefault value.
func (m ProfilePreset) GetIsDefault() bool {
	return m.IsDefault
}

// SetIsDefault sets the IsDefault value and returns the updated ProfilePreset.
func (m *ProfilePreset) SetIsDefault(v bool) {
	m.IsDefault = v
}

// GetDia returns the Dia value.
func (m ProfilePreset) GetDia() *float64 {
	return m.Dia
}

// SetDia sets the Dia value and returns the updated ProfilePreset.
func (m *ProfilePreset) SetDia(v *float64) {
	m.Dia = v
}

// GetUnits returns the Units value.
func (m ProfilePreset) GetUnits() *consts.Units {
	return m.Units
}

// SetUnits sets the Units value and returns the updated ProfilePreset.
func (m *ProfilePreset) SetUnits(v *consts.Units) {
	m.Units = v
}

// GetTimezone returns the Timezone value.
func (m ProfilePreset) GetTimezone() *string {
	return m.Timezone
}

// SetTimezone sets the Timezone value and returns the updated ProfilePreset.
func (m *ProfilePreset) SetTimezone(v *string) {
	m.Timezone = v
}

// GetCarbsHr returns the CarbsHr value.
func (m ProfilePreset) GetCarbsHr() *float64 {
	return m.CarbsHr
}

// SetCarbsHr sets the CarbsHr value and returns the updated ProfilePreset.
func (m *ProfilePreset) SetCarbsHr(v *float64) {
	m.CarbsHr = v
}

// GetDelay returns the Delay value.
func (m ProfilePreset) GetDelay() *float64 {
	return m.Delay
}

// SetDelay sets the Delay value and returns the updated ProfilePreset.
func (m *ProfilePreset) SetDelay(v *float64) {
	m.Delay = v
}

// GetCarbRatiosJson returns the CarbRatiosJson value.
func (m ProfilePreset) GetCarbRatiosJson() *string {
	return m.CarbRatiosJson
}

// SetCarbRatiosJson sets the CarbRatiosJson value and returns the updated ProfilePreset.
func (m *ProfilePreset) SetCarbRatiosJson(v *string) {
	m.CarbRatiosJson = v
}

// GetInsulinSensitivitiesJson returns the InsulinSensitivitiesJson value.
func (m ProfilePreset) GetInsulinSensitivitiesJson() *string {
	return m.InsulinSensitivitiesJson
}

// SetInsulinSensitivitiesJson sets the InsulinSensitivitiesJson value and returns the updated ProfilePreset.
func (m *ProfilePreset) SetInsulinSensitivitiesJson(v *string) {
	m.InsulinSensitivitiesJson = v
}

// GetBasalRatesJson returns the BasalRatesJson value.
func (m ProfilePreset) GetBasalRatesJson() *string {
	return m.BasalRatesJson
}

// SetBasalRatesJson sets the BasalRatesJson value and returns the updated ProfilePreset.
func (m *ProfilePreset) SetBasalRatesJson(v *string) {
	m.BasalRatesJson = v
}

// GetTargetsHighJson returns the TargetsHighJson value.
func (m ProfilePreset) GetTargetsHighJson() *string {
	return m.TargetsHighJson
}

// SetTargetsHighJson sets the TargetsHighJson value and returns the updated ProfilePreset.
func (m *ProfilePreset) SetTargetsHighJson(v *string) {
	m.TargetsHighJson = v
}

// GetTargetsLowJson returns the TargetsLowJson value.
func (m ProfilePreset) GetTargetsLowJson() *string {
	return m.TargetsLowJson
}

// SetTargetsLowJson sets the TargetsLowJson value and returns the updated ProfilePreset.
func (m *ProfilePreset) SetTargetsLowJson(v *string) {
	m.TargetsLowJson = v
}

// GetLoopSettingsJson returns the LoopSettingsJson value.
func (m ProfilePreset) GetLoopSettingsJson() *string {
	return m.LoopSettingsJson
}

// SetLoopSettingsJson sets the LoopSettingsJson value and returns the updated ProfilePreset.
func (m *ProfilePreset) SetLoopSettingsJson(v *string) {
	m.LoopSettingsJson = v
}

// GetRemoteSettingsJson returns the RemoteSettingsJson value.
func (m ProfilePreset) GetRemoteSettingsJson() *string {
	return m.RemoteSettingsJson
}

// SetRemoteSettingsJson sets the RemoteSettingsJson value and returns the updated ProfilePreset.
func (m *ProfilePreset) SetRemoteSettingsJson(v *string) {
	m.RemoteSettingsJson = v
}

// GetRawJsonData returns the RawJsonData value.
func (m ProfilePreset) GetRawJsonData() *string {
	return m.RawJsonData
}

// SetRawJsonData sets the RawJsonData value and returns the updated ProfilePreset.
func (m *ProfilePreset) SetRawJsonData(v *string) {
	m.RawJsonData = v
}

// GetUtcOffsetMinutes returns the UtcOffsetMinutes value.
func (m ProfilePreset) GetUtcOffsetMinutes() *int64 {
	return m.UtcOffsetMinutes
}

// SetUtcOffsetMinutes sets the UtcOffsetMinutes value and returns the updated ProfilePreset.
func (m *ProfilePreset) SetUtcOffsetMinutes(v *int64) {
	m.UtcOffsetMinutes = v
}

// GetOccurredAt returns the OccurredAt value.
func (m ProfilePreset) GetOccurredAt() int64 {
	return m.OccurredAt
}

// SetOccurredAt sets the OccurredAt value and returns the updated ProfilePreset.
func (m *ProfilePreset) SetOccurredAt(v int64) {
	m.OccurredAt = v
}

// GetCreatedAt returns the CreatedAt value.
func (m ProfilePreset) GetCreatedAt() int64 {
	return m.CreatedAt
}

// SetCreatedAt sets the CreatedAt value and returns the updated ProfilePreset.
func (m *ProfilePreset) SetCreatedAt(v int64) {
	m.CreatedAt = v
}

// GetUpdatedAt returns the UpdatedAt value.
func (m ProfilePreset) GetUpdatedAt() int64 {
	return m.UpdatedAt
}

// SetUpdatedAt sets the UpdatedAt value and returns the updated ProfilePreset.
func (m *ProfilePreset) SetUpdatedAt(v int64) {
	m.UpdatedAt = v
}

// GetCreatedByAppName returns the CreatedByAppName value.
func (m ProfilePreset) GetCreatedByAppName() *string {
	return m.CreatedByAppName
}

// SetCreatedByAppName sets the CreatedByAppName value and returns the updated ProfilePreset.
func (m *ProfilePreset) SetCreatedByAppName(v *string) {
	m.CreatedByAppName = v
}

// GetCreatedByDeviceName returns the CreatedByDeviceName value.
func (m ProfilePreset) GetCreatedByDeviceName() *string {
	return m.CreatedByDeviceName
}

// SetCreatedByDeviceName sets the CreatedByDeviceName value and returns the updated ProfilePreset.
func (m *ProfilePreset) SetCreatedByDeviceName(v *string) {
	m.CreatedByDeviceName = v
}

// GetUpdatedByIdentityID returns the UpdatedByIdentityID value.
func (m ProfilePreset) GetUpdatedByIdentityID() *int64 {
	return m.UpdatedByIdentityID
}

// SetUpdatedByIdentityID sets the UpdatedByIdentityID value and returns the updated ProfilePreset.
func (m *ProfilePreset) SetUpdatedByIdentityID(v *int64) {
	m.UpdatedByIdentityID = v
}

// GetUpdatedByIdentityName returns the UpdatedByIdentityName value.
func (m ProfilePreset) GetUpdatedByIdentityName() *string {
	return m.UpdatedByIdentityName
}

// SetUpdatedByIdentityName sets the UpdatedByIdentityName value and returns the updated ProfilePreset.
func (m *ProfilePreset) SetUpdatedByIdentityName(v *string) {
	m.UpdatedByIdentityName = v
}

// GetCreatedByIdentityID returns the CreatedByIdentityID value.
func (m ProfilePreset) GetCreatedByIdentityID() *int64 {
	return m.CreatedByIdentityID
}

// SetCreatedByIdentityID sets the CreatedByIdentityID value and returns the updated ProfilePreset.
func (m *ProfilePreset) SetCreatedByIdentityID(v *int64) {
	m.CreatedByIdentityID = v
}

// GetCreatedByIdentityName returns the CreatedByIdentityName value.
func (m ProfilePreset) GetCreatedByIdentityName() string {
	return m.CreatedByIdentityName
}

// SetCreatedByIdentityName sets the CreatedByIdentityName value and returns the updated ProfilePreset.
func (m *ProfilePreset) SetCreatedByIdentityName(v string) {
	m.CreatedByIdentityName = v
}

// GetID returns the ID value.
func (m ProfileSwitch) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated ProfileSwitch.
func (m *ProfileSwitch) SetID(v string) {
	m.ID = v
}

// GetName returns the Name value.
func (m ProfileSwitch) GetName() string {
	return m.Name
}

// SetName sets the Name value and returns the updated ProfileSwitch.
func (m *ProfileSwitch) SetName(v string) {
	m.Name = v
}

// GetDurationMinutes returns the DurationMinutes value.
func (m ProfileSwitch) GetDurationMinutes() int64 {
	return m.DurationMinutes
}

// SetDurationMinutes sets the DurationMinutes value and returns the updated ProfileSwitch.
func (m *ProfileSwitch) SetDurationMinutes(v int64) {
	m.DurationMinutes = v
}

// GetPercent returns the Percent value.
func (m ProfileSwitch) GetPercent() *float64 {
	return m.Percent
}

// SetPercent sets the Percent value and returns the updated ProfileSwitch.
func (m *ProfileSwitch) SetPercent(v *float64) {
	m.Percent = v
}

// GetTimeshift returns the Timeshift value.
func (m ProfileSwitch) GetTimeshift() *int64 {
	return m.Timeshift
}

// SetTimeshift sets the Timeshift value and returns the updated ProfileSwitch.
func (m *ProfileSwitch) SetTimeshift(v *int64) {
	m.Timeshift = v
}

// GetDia returns the Dia value.
func (m ProfileSwitch) GetDia() *float64 {
	return m.Dia
}

// SetDia sets the Dia value and returns the updated ProfileSwitch.
func (m *ProfileSwitch) SetDia(v *float64) {
	m.Dia = v
}

// GetUnits returns the Units value.
func (m ProfileSwitch) GetUnits() *consts.Units {
	return m.Units
}

// SetUnits sets the Units value and returns the updated ProfileSwitch.
func (m *ProfileSwitch) SetUnits(v *consts.Units) {
	m.Units = v
}

// GetTimezone returns the Timezone value.
func (m ProfileSwitch) GetTimezone() *string {
	return m.Timezone
}

// SetTimezone sets the Timezone value and returns the updated ProfileSwitch.
func (m *ProfileSwitch) SetTimezone(v *string) {
	m.Timezone = v
}

// GetCarbsHr returns the CarbsHr value.
func (m ProfileSwitch) GetCarbsHr() *float64 {
	return m.CarbsHr
}

// SetCarbsHr sets the CarbsHr value and returns the updated ProfileSwitch.
func (m *ProfileSwitch) SetCarbsHr(v *float64) {
	m.CarbsHr = v
}

// GetDelay returns the Delay value.
func (m ProfileSwitch) GetDelay() *float64 {
	return m.Delay
}

// SetDelay sets the Delay value and returns the updated ProfileSwitch.
func (m *ProfileSwitch) SetDelay(v *float64) {
	m.Delay = v
}

// GetCarbRatiosJson returns the CarbRatiosJson value.
func (m ProfileSwitch) GetCarbRatiosJson() *string {
	return m.CarbRatiosJson
}

// SetCarbRatiosJson sets the CarbRatiosJson value and returns the updated ProfileSwitch.
func (m *ProfileSwitch) SetCarbRatiosJson(v *string) {
	m.CarbRatiosJson = v
}

// GetInsulinSensitivitiesJson returns the InsulinSensitivitiesJson value.
func (m ProfileSwitch) GetInsulinSensitivitiesJson() *string {
	return m.InsulinSensitivitiesJson
}

// SetInsulinSensitivitiesJson sets the InsulinSensitivitiesJson value and returns the updated ProfileSwitch.
func (m *ProfileSwitch) SetInsulinSensitivitiesJson(v *string) {
	m.InsulinSensitivitiesJson = v
}

// GetBasalRatesJson returns the BasalRatesJson value.
func (m ProfileSwitch) GetBasalRatesJson() *string {
	return m.BasalRatesJson
}

// SetBasalRatesJson sets the BasalRatesJson value and returns the updated ProfileSwitch.
func (m *ProfileSwitch) SetBasalRatesJson(v *string) {
	m.BasalRatesJson = v
}

// GetTargetsHighJson returns the TargetsHighJson value.
func (m ProfileSwitch) GetTargetsHighJson() *string {
	return m.TargetsHighJson
}

// SetTargetsHighJson sets the TargetsHighJson value and returns the updated ProfileSwitch.
func (m *ProfileSwitch) SetTargetsHighJson(v *string) {
	m.TargetsHighJson = v
}

// GetTargetsLowJson returns the TargetsLowJson value.
func (m ProfileSwitch) GetTargetsLowJson() *string {
	return m.TargetsLowJson
}

// SetTargetsLowJson sets the TargetsLowJson value and returns the updated ProfileSwitch.
func (m *ProfileSwitch) SetTargetsLowJson(v *string) {
	m.TargetsLowJson = v
}

// GetNotes returns the Notes value.
func (m ProfileSwitch) GetNotes() *string {
	return m.Notes
}

// SetNotes sets the Notes value and returns the updated ProfileSwitch.
func (m *ProfileSwitch) SetNotes(v *string) {
	m.Notes = v
}

// GetReason returns the Reason value.
func (m ProfileSwitch) GetReason() *string {
	return m.Reason
}

// SetReason sets the Reason value and returns the updated ProfileSwitch.
func (m *ProfileSwitch) SetReason(v *string) {
	m.Reason = v
}

// GetRawJsonData returns the RawJsonData value.
func (m ProfileSwitch) GetRawJsonData() *string {
	return m.RawJsonData
}

// SetRawJsonData sets the RawJsonData value and returns the updated ProfileSwitch.
func (m *ProfileSwitch) SetRawJsonData(v *string) {
	m.RawJsonData = v
}

// GetDeletedAt returns the DeletedAt value.
func (m ProfileSwitch) GetDeletedAt() *int64 {
	return m.DeletedAt
}

// SetDeletedAt sets the DeletedAt value and returns the updated ProfileSwitch.
func (m *ProfileSwitch) SetDeletedAt(v *int64) {
	m.DeletedAt = v
}

// GetUtcOffsetMinutes returns the UtcOffsetMinutes value.
func (m ProfileSwitch) GetUtcOffsetMinutes() *int64 {
	return m.UtcOffsetMinutes
}

// SetUtcOffsetMinutes sets the UtcOffsetMinutes value and returns the updated ProfileSwitch.
func (m *ProfileSwitch) SetUtcOffsetMinutes(v *int64) {
	m.UtcOffsetMinutes = v
}

// GetOccurredAt returns the OccurredAt value.
func (m ProfileSwitch) GetOccurredAt() int64 {
	return m.OccurredAt
}

// SetOccurredAt sets the OccurredAt value and returns the updated ProfileSwitch.
func (m *ProfileSwitch) SetOccurredAt(v int64) {
	m.OccurredAt = v
}

// GetCreatedAt returns the CreatedAt value.
func (m ProfileSwitch) GetCreatedAt() int64 {
	return m.CreatedAt
}

// SetCreatedAt sets the CreatedAt value and returns the updated ProfileSwitch.
func (m *ProfileSwitch) SetCreatedAt(v int64) {
	m.CreatedAt = v
}

// GetUpdatedAt returns the UpdatedAt value.
func (m ProfileSwitch) GetUpdatedAt() int64 {
	return m.UpdatedAt
}

// SetUpdatedAt sets the UpdatedAt value and returns the updated ProfileSwitch.
func (m *ProfileSwitch) SetUpdatedAt(v int64) {
	m.UpdatedAt = v
}

// GetCreatedByAppName returns the CreatedByAppName value.
func (m ProfileSwitch) GetCreatedByAppName() *string {
	return m.CreatedByAppName
}

// SetCreatedByAppName sets the CreatedByAppName value and returns the updated ProfileSwitch.
func (m *ProfileSwitch) SetCreatedByAppName(v *string) {
	m.CreatedByAppName = v
}

// GetCreatedByDeviceName returns the CreatedByDeviceName value.
func (m ProfileSwitch) GetCreatedByDeviceName() *string {
	return m.CreatedByDeviceName
}

// SetCreatedByDeviceName sets the CreatedByDeviceName value and returns the updated ProfileSwitch.
func (m *ProfileSwitch) SetCreatedByDeviceName(v *string) {
	m.CreatedByDeviceName = v
}

// GetUpdatedByIdentityID returns the UpdatedByIdentityID value.
func (m ProfileSwitch) GetUpdatedByIdentityID() *int64 {
	return m.UpdatedByIdentityID
}

// SetUpdatedByIdentityID sets the UpdatedByIdentityID value and returns the updated ProfileSwitch.
func (m *ProfileSwitch) SetUpdatedByIdentityID(v *int64) {
	m.UpdatedByIdentityID = v
}

// GetUpdatedByIdentityName returns the UpdatedByIdentityName value.
func (m ProfileSwitch) GetUpdatedByIdentityName() *string {
	return m.UpdatedByIdentityName
}

// SetUpdatedByIdentityName sets the UpdatedByIdentityName value and returns the updated ProfileSwitch.
func (m *ProfileSwitch) SetUpdatedByIdentityName(v *string) {
	m.UpdatedByIdentityName = v
}

// GetCreatedByIdentityID returns the CreatedByIdentityID value.
func (m ProfileSwitch) GetCreatedByIdentityID() *int64 {
	return m.CreatedByIdentityID
}

// SetCreatedByIdentityID sets the CreatedByIdentityID value and returns the updated ProfileSwitch.
func (m *ProfileSwitch) SetCreatedByIdentityID(v *int64) {
	m.CreatedByIdentityID = v
}

// GetCreatedByIdentityName returns the CreatedByIdentityName value.
func (m ProfileSwitch) GetCreatedByIdentityName() string {
	return m.CreatedByIdentityName
}

// SetCreatedByIdentityName sets the CreatedByIdentityName value and returns the updated ProfileSwitch.
func (m *ProfileSwitch) SetCreatedByIdentityName(v string) {
	m.CreatedByIdentityName = v
}

// GetID returns the ID value.
func (m ProfileSwitchesVw) GetID() *string {
	return m.ID
}

// SetID sets the ID value and returns the updated ProfileSwitchesVw.
func (m *ProfileSwitchesVw) SetID(v *string) {
	m.ID = v
}

// GetName returns the Name value.
func (m ProfileSwitchesVw) GetName() *string {
	return m.Name
}

// SetName sets the Name value and returns the updated ProfileSwitchesVw.
func (m *ProfileSwitchesVw) SetName(v *string) {
	m.Name = v
}

// GetDurationMinutes returns the DurationMinutes value.
func (m ProfileSwitchesVw) GetDurationMinutes() *int64 {
	return m.DurationMinutes
}

// SetDurationMinutes sets the DurationMinutes value and returns the updated ProfileSwitchesVw.
func (m *ProfileSwitchesVw) SetDurationMinutes(v *int64) {
	m.DurationMinutes = v
}

// GetPercent returns the Percent value.
func (m ProfileSwitchesVw) GetPercent() *float64 {
	return m.Percent
}

// SetPercent sets the Percent value and returns the updated ProfileSwitchesVw.
func (m *ProfileSwitchesVw) SetPercent(v *float64) {
	m.Percent = v
}

// GetTimeshift returns the Timeshift value.
func (m ProfileSwitchesVw) GetTimeshift() *int64 {
	return m.Timeshift
}

// SetTimeshift sets the Timeshift value and returns the updated ProfileSwitchesVw.
func (m *ProfileSwitchesVw) SetTimeshift(v *int64) {
	m.Timeshift = v
}

// GetDia returns the Dia value.
func (m ProfileSwitchesVw) GetDia() *float64 {
	return m.Dia
}

// SetDia sets the Dia value and returns the updated ProfileSwitchesVw.
func (m *ProfileSwitchesVw) SetDia(v *float64) {
	m.Dia = v
}

// GetUnits returns the Units value.
func (m ProfileSwitchesVw) GetUnits() *consts.Units {
	return m.Units
}

// SetUnits sets the Units value and returns the updated ProfileSwitchesVw.
func (m *ProfileSwitchesVw) SetUnits(v *consts.Units) {
	m.Units = v
}

// GetTimezone returns the Timezone value.
func (m ProfileSwitchesVw) GetTimezone() *string {
	return m.Timezone
}

// SetTimezone sets the Timezone value and returns the updated ProfileSwitchesVw.
func (m *ProfileSwitchesVw) SetTimezone(v *string) {
	m.Timezone = v
}

// GetCarbsHr returns the CarbsHr value.
func (m ProfileSwitchesVw) GetCarbsHr() *float64 {
	return m.CarbsHr
}

// SetCarbsHr sets the CarbsHr value and returns the updated ProfileSwitchesVw.
func (m *ProfileSwitchesVw) SetCarbsHr(v *float64) {
	m.CarbsHr = v
}

// GetDelay returns the Delay value.
func (m ProfileSwitchesVw) GetDelay() *float64 {
	return m.Delay
}

// SetDelay sets the Delay value and returns the updated ProfileSwitchesVw.
func (m *ProfileSwitchesVw) SetDelay(v *float64) {
	m.Delay = v
}

// GetCarbRatiosJson returns the CarbRatiosJson value.
func (m ProfileSwitchesVw) GetCarbRatiosJson() *string {
	return m.CarbRatiosJson
}

// SetCarbRatiosJson sets the CarbRatiosJson value and returns the updated ProfileSwitchesVw.
func (m *ProfileSwitchesVw) SetCarbRatiosJson(v *string) {
	m.CarbRatiosJson = v
}

// GetInsulinSensitivitiesJson returns the InsulinSensitivitiesJson value.
func (m ProfileSwitchesVw) GetInsulinSensitivitiesJson() *string {
	return m.InsulinSensitivitiesJson
}

// SetInsulinSensitivitiesJson sets the InsulinSensitivitiesJson value and returns the updated ProfileSwitchesVw.
func (m *ProfileSwitchesVw) SetInsulinSensitivitiesJson(v *string) {
	m.InsulinSensitivitiesJson = v
}

// GetBasalRatesJson returns the BasalRatesJson value.
func (m ProfileSwitchesVw) GetBasalRatesJson() *string {
	return m.BasalRatesJson
}

// SetBasalRatesJson sets the BasalRatesJson value and returns the updated ProfileSwitchesVw.
func (m *ProfileSwitchesVw) SetBasalRatesJson(v *string) {
	m.BasalRatesJson = v
}

// GetTargetsHighJson returns the TargetsHighJson value.
func (m ProfileSwitchesVw) GetTargetsHighJson() *string {
	return m.TargetsHighJson
}

// SetTargetsHighJson sets the TargetsHighJson value and returns the updated ProfileSwitchesVw.
func (m *ProfileSwitchesVw) SetTargetsHighJson(v *string) {
	m.TargetsHighJson = v
}

// GetTargetsLowJson returns the TargetsLowJson value.
func (m ProfileSwitchesVw) GetTargetsLowJson() *string {
	return m.TargetsLowJson
}

// SetTargetsLowJson sets the TargetsLowJson value and returns the updated ProfileSwitchesVw.
func (m *ProfileSwitchesVw) SetTargetsLowJson(v *string) {
	m.TargetsLowJson = v
}

// GetNotes returns the Notes value.
func (m ProfileSwitchesVw) GetNotes() *string {
	return m.Notes
}

// SetNotes sets the Notes value and returns the updated ProfileSwitchesVw.
func (m *ProfileSwitchesVw) SetNotes(v *string) {
	m.Notes = v
}

// GetReason returns the Reason value.
func (m ProfileSwitchesVw) GetReason() *string {
	return m.Reason
}

// SetReason sets the Reason value and returns the updated ProfileSwitchesVw.
func (m *ProfileSwitchesVw) SetReason(v *string) {
	m.Reason = v
}

// GetRawJsonData returns the RawJsonData value.
func (m ProfileSwitchesVw) GetRawJsonData() *string {
	return m.RawJsonData
}

// SetRawJsonData sets the RawJsonData value and returns the updated ProfileSwitchesVw.
func (m *ProfileSwitchesVw) SetRawJsonData(v *string) {
	m.RawJsonData = v
}

// GetDeletedAt returns the DeletedAt value.
func (m ProfileSwitchesVw) GetDeletedAt() *int64 {
	return m.DeletedAt
}

// SetDeletedAt sets the DeletedAt value and returns the updated ProfileSwitchesVw.
func (m *ProfileSwitchesVw) SetDeletedAt(v *int64) {
	m.DeletedAt = v
}

// GetUtcOffsetMinutes returns the UtcOffsetMinutes value.
func (m ProfileSwitchesVw) GetUtcOffsetMinutes() *int64 {
	return m.UtcOffsetMinutes
}

// SetUtcOffsetMinutes sets the UtcOffsetMinutes value and returns the updated ProfileSwitchesVw.
func (m *ProfileSwitchesVw) SetUtcOffsetMinutes(v *int64) {
	m.UtcOffsetMinutes = v
}

// GetOccurredAt returns the OccurredAt value.
func (m ProfileSwitchesVw) GetOccurredAt() *int64 {
	return m.OccurredAt
}

// SetOccurredAt sets the OccurredAt value and returns the updated ProfileSwitchesVw.
func (m *ProfileSwitchesVw) SetOccurredAt(v *int64) {
	m.OccurredAt = v
}

// GetCreatedAt returns the CreatedAt value.
func (m ProfileSwitchesVw) GetCreatedAt() *int64 {
	return m.CreatedAt
}

// SetCreatedAt sets the CreatedAt value and returns the updated ProfileSwitchesVw.
func (m *ProfileSwitchesVw) SetCreatedAt(v *int64) {
	m.CreatedAt = v
}

// GetUpdatedAt returns the UpdatedAt value.
func (m ProfileSwitchesVw) GetUpdatedAt() *int64 {
	return m.UpdatedAt
}

// SetUpdatedAt sets the UpdatedAt value and returns the updated ProfileSwitchesVw.
func (m *ProfileSwitchesVw) SetUpdatedAt(v *int64) {
	m.UpdatedAt = v
}

// GetCreatedByAppName returns the CreatedByAppName value.
func (m ProfileSwitchesVw) GetCreatedByAppName() *string {
	return m.CreatedByAppName
}

// SetCreatedByAppName sets the CreatedByAppName value and returns the updated ProfileSwitchesVw.
func (m *ProfileSwitchesVw) SetCreatedByAppName(v *string) {
	m.CreatedByAppName = v
}

// GetCreatedByDeviceName returns the CreatedByDeviceName value.
func (m ProfileSwitchesVw) GetCreatedByDeviceName() *string {
	return m.CreatedByDeviceName
}

// SetCreatedByDeviceName sets the CreatedByDeviceName value and returns the updated ProfileSwitchesVw.
func (m *ProfileSwitchesVw) SetCreatedByDeviceName(v *string) {
	m.CreatedByDeviceName = v
}

// GetUpdatedByIdentityID returns the UpdatedByIdentityID value.
func (m ProfileSwitchesVw) GetUpdatedByIdentityID() *int64 {
	return m.UpdatedByIdentityID
}

// SetUpdatedByIdentityID sets the UpdatedByIdentityID value and returns the updated ProfileSwitchesVw.
func (m *ProfileSwitchesVw) SetUpdatedByIdentityID(v *int64) {
	m.UpdatedByIdentityID = v
}

// GetUpdatedByIdentityName returns the UpdatedByIdentityName value.
func (m ProfileSwitchesVw) GetUpdatedByIdentityName() *string {
	return m.UpdatedByIdentityName
}

// SetUpdatedByIdentityName sets the UpdatedByIdentityName value and returns the updated ProfileSwitchesVw.
func (m *ProfileSwitchesVw) SetUpdatedByIdentityName(v *string) {
	m.UpdatedByIdentityName = v
}

// GetCreatedByIdentityID returns the CreatedByIdentityID value.
func (m ProfileSwitchesVw) GetCreatedByIdentityID() *int64 {
	return m.CreatedByIdentityID
}

// SetCreatedByIdentityID sets the CreatedByIdentityID value and returns the updated ProfileSwitchesVw.
func (m *ProfileSwitchesVw) SetCreatedByIdentityID(v *int64) {
	m.CreatedByIdentityID = v
}

// GetCreatedByIdentityName returns the CreatedByIdentityName value.
func (m ProfileSwitchesVw) GetCreatedByIdentityName() *string {
	return m.CreatedByIdentityName
}

// SetCreatedByIdentityName sets the CreatedByIdentityName value and returns the updated ProfileSwitchesVw.
func (m *ProfileSwitchesVw) SetCreatedByIdentityName(v *string) {
	m.CreatedByIdentityName = v
}

// GetPeriod returns the Period value.
func (m SummaryPeriod) GetPeriod() string {
	return m.Period
}

// SetPeriod sets the Period value and returns the updated SummaryPeriod.
func (m *SummaryPeriod) SetPeriod(v string) {
	m.Period = v
}

// GetDate returns the Date value.
func (m SummaryPeriod) GetDate() string {
	return m.Date
}

// SetDate sets the Date value and returns the updated SummaryPeriod.
func (m *SummaryPeriod) SetDate(v string) {
	m.Date = v
}

// GetTimezone returns the Timezone value.
func (m SummaryPeriod) GetTimezone() string {
	return m.Timezone
}

// SetTimezone sets the Timezone value and returns the updated SummaryPeriod.
func (m *SummaryPeriod) SetTimezone(v string) {
	m.Timezone = v
}

// GetStartAt returns the StartAt value.
func (m SummaryPeriod) GetStartAt() int64 {
	return m.StartAt
}

// SetStartAt sets the StartAt value and returns the updated SummaryPeriod.
func (m *SummaryPeriod) SetStartAt(v int64) {
	m.StartAt = v
}

// GetEndAt returns the EndAt value.
func (m SummaryPeriod) GetEndAt() int64 {
	return m.EndAt
}

// SetEndAt sets the EndAt value and returns the updated SummaryPeriod.
func (m *SummaryPeriod) SetEndAt(v int64) {
	m.EndAt = v
}

// GetDays returns the Days value.
func (m SummaryPeriod) GetDays() int64 {
	return m.Days
}

// SetDays sets the Days value and returns the updated SummaryPeriod.
func (m *SummaryPeriod) SetDays(v int64) {
	m.Days = v
}

// GetTimeInRange returns the TimeInRange value.
func (m SummaryPeriod) GetTimeInRange() *float64 {
	return m.TimeInRange
}

// SetTimeInRange sets the TimeInRange value and returns the updated SummaryPeriod.
func (m *SummaryPeriod) SetTimeInRange(v *float64) {
	m.TimeInRange = v
}

// GetTotalDailyDose returns the TotalDailyDose value.
func (m SummaryPeriod) GetTotalDailyDose() float64 {
	return m.TotalDailyDose
}

// SetTotalDailyDose sets the TotalDailyDose value and returns the updated SummaryPeriod.
func (m *SummaryPeriod) SetTotalDailyDose(v float64) {
	m.TotalDailyDose = v
}

// GetTotalDailyCarbs returns the TotalDailyCarbs value.
func (m SummaryPeriod) GetTotalDailyCarbs() float64 {
	return m.TotalDailyCarbs
}

// SetTotalDailyCarbs sets the TotalDailyCarbs value and returns the updated SummaryPeriod.
func (m *SummaryPeriod) SetTotalDailyCarbs(v float64) {
	m.TotalDailyCarbs = v
}

// GetGlucoseMgdlJson returns the GlucoseMgdlJson value.
func (m SummaryPeriod) GetGlucoseMgdlJson() string {
	return m.GlucoseMgdlJson
}

// SetGlucoseMgdlJson sets the GlucoseMgdlJson value and returns the updated SummaryPeriod.
func (m *SummaryPeriod) SetGlucoseMgdlJson(v string) {
	m.GlucoseMgdlJson = v
}

// GetAgpMgdlJson returns the AgpMgdlJson value.
func (m SummaryPeriod) GetAgpMgdlJson() *string {
	return m.AgpMgdlJson
}

// SetAgpMgdlJson sets the AgpMgdlJson value and returns the updated SummaryPeriod.
func (m *SummaryPeriod) SetAgpMgdlJson(v *string) {
	m.AgpMgdlJson = v
}

// GetCreatedAt returns the CreatedAt value.
func (m SummaryPeriod) GetCreatedAt() int64 {
	return m.CreatedAt
}

// SetCreatedAt sets the CreatedAt value and returns the updated SummaryPeriod.
func (m *SummaryPeriod) SetCreatedAt(v int64) {
	m.CreatedAt = v
}

// GetUpdatedAt returns the UpdatedAt value.
func (m SummaryPeriod) GetUpdatedAt() int64 {
	return m.UpdatedAt
}

// SetUpdatedAt sets the UpdatedAt value and returns the updated SummaryPeriod.
func (m *SummaryPeriod) SetUpdatedAt(v int64) {
	m.UpdatedAt = v
}

// GetID returns the ID value.
func (m TempBasal) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated TempBasal.
func (m *TempBasal) SetID(v string) {
	m.ID = v
}

// GetDurationMinutes returns the DurationMinutes value.
func (m TempBasal) GetDurationMinutes() int64 {
	return m.DurationMinutes
}

// SetDurationMinutes sets the DurationMinutes value and returns the updated TempBasal.
func (m *TempBasal) SetDurationMinutes(v int64) {
	m.DurationMinutes = v
}

// GetAbsolute returns the Absolute value.
func (m TempBasal) GetAbsolute() float64 {
	return m.Absolute
}

// SetAbsolute sets the Absolute value and returns the updated TempBasal.
func (m *TempBasal) SetAbsolute(v float64) {
	m.Absolute = v
}

// GetPercent returns the Percent value.
func (m TempBasal) GetPercent() *float64 {
	return m.Percent
}

// SetPercent sets the Percent value and returns the updated TempBasal.
func (m *TempBasal) SetPercent(v *float64) {
	m.Percent = v
}

// GetNotes returns the Notes value.
func (m TempBasal) GetNotes() *string {
	return m.Notes
}

// SetNotes sets the Notes value and returns the updated TempBasal.
func (m *TempBasal) SetNotes(v *string) {
	m.Notes = v
}

// GetReason returns the Reason value.
func (m TempBasal) GetReason() *string {
	return m.Reason
}

// SetReason sets the Reason value and returns the updated TempBasal.
func (m *TempBasal) SetReason(v *string) {
	m.Reason = v
}

// GetRawJsonData returns the RawJsonData value.
func (m TempBasal) GetRawJsonData() *string {
	return m.RawJsonData
}

// SetRawJsonData sets the RawJsonData value and returns the updated TempBasal.
func (m *TempBasal) SetRawJsonData(v *string) {
	m.RawJsonData = v
}

// GetDeletedAt returns the DeletedAt value.
func (m TempBasal) GetDeletedAt() *int64 {
	return m.DeletedAt
}

// SetDeletedAt sets the DeletedAt value and returns the updated TempBasal.
func (m *TempBasal) SetDeletedAt(v *int64) {
	m.DeletedAt = v
}

// GetUtcOffsetMinutes returns the UtcOffsetMinutes value.
func (m TempBasal) GetUtcOffsetMinutes() *int64 {
	return m.UtcOffsetMinutes
}

// SetUtcOffsetMinutes sets the UtcOffsetMinutes value and returns the updated TempBasal.
func (m *TempBasal) SetUtcOffsetMinutes(v *int64) {
	m.UtcOffsetMinutes = v
}

// GetOccurredAt returns the OccurredAt value.
func (m TempBasal) GetOccurredAt() int64 {
	return m.OccurredAt
}

// SetOccurredAt sets the OccurredAt value and returns the updated TempBasal.
func (m *TempBasal) SetOccurredAt(v int64) {
	m.OccurredAt = v
}

// GetCreatedAt returns the CreatedAt value.
func (m TempBasal) GetCreatedAt() int64 {
	return m.CreatedAt
}

// SetCreatedAt sets the CreatedAt value and returns the updated TempBasal.
func (m *TempBasal) SetCreatedAt(v int64) {
	m.CreatedAt = v
}

// GetUpdatedAt returns the UpdatedAt value.
func (m TempBasal) GetUpdatedAt() int64 {
	return m.UpdatedAt
}

// SetUpdatedAt sets the UpdatedAt value and returns the updated TempBasal.
func (m *TempBasal) SetUpdatedAt(v int64) {
	m.UpdatedAt = v
}

// GetCreatedByAppName returns the CreatedByAppName value.
func (m TempBasal) GetCreatedByAppName() *string {
	return m.CreatedByAppName
}

// SetCreatedByAppName sets the CreatedByAppName value and returns the updated TempBasal.
func (m *TempBasal) SetCreatedByAppName(v *string) {
	m.CreatedByAppName = v
}

// GetCreatedByDeviceName returns the CreatedByDeviceName value.
func (m TempBasal) GetCreatedByDeviceName() *string {
	return m.CreatedByDeviceName
}

// SetCreatedByDeviceName sets the CreatedByDeviceName value and returns the updated TempBasal.
func (m *TempBasal) SetCreatedByDeviceName(v *string) {
	m.CreatedByDeviceName = v
}

// GetUpdatedByIdentityID returns the UpdatedByIdentityID value.
func (m TempBasal) GetUpdatedByIdentityID() *int64 {
	return m.UpdatedByIdentityID
}

// SetUpdatedByIdentityID sets the UpdatedByIdentityID value and returns the updated TempBasal.
func (m *TempBasal) SetUpdatedByIdentityID(v *int64) {
	m.UpdatedByIdentityID = v
}

// GetUpdatedByIdentityName returns the UpdatedByIdentityName value.
func (m TempBasal) GetUpdatedByIdentityName() *string {
	return m.UpdatedByIdentityName
}

// SetUpdatedByIdentityName sets the UpdatedByIdentityName value and returns the updated TempBasal.
func (m *TempBasal) SetUpdatedByIdentityName(v *string) {
	m.UpdatedByIdentityName = v
}

// GetCreatedByIdentityID returns the CreatedByIdentityID value.
func (m TempBasal) GetCreatedByIdentityID() *int64 {
	return m.CreatedByIdentityID
}

// SetCreatedByIdentityID sets the CreatedByIdentityID value and returns the updated TempBasal.
func (m *TempBasal) SetCreatedByIdentityID(v *int64) {
	m.CreatedByIdentityID = v
}

// GetCreatedByIdentityName returns the CreatedByIdentityName value.
func (m TempBasal) GetCreatedByIdentityName() string {
	return m.CreatedByIdentityName
}

// SetCreatedByIdentityName sets the CreatedByIdentityName value and returns the updated TempBasal.
func (m *TempBasal) SetCreatedByIdentityName(v string) {
	m.CreatedByIdentityName = v
}

// GetID returns the ID value.
func (m TempBasalsVw) GetID() *string {
	return m.ID
}

// SetID sets the ID value and returns the updated TempBasalsVw.
func (m *TempBasalsVw) SetID(v *string) {
	m.ID = v
}

// GetDurationMinutes returns the DurationMinutes value.
func (m TempBasalsVw) GetDurationMinutes() *int64 {
	return m.DurationMinutes
}

// SetDurationMinutes sets the DurationMinutes value and returns the updated TempBasalsVw.
func (m *TempBasalsVw) SetDurationMinutes(v *int64) {
	m.DurationMinutes = v
}

// GetAbsolute returns the Absolute value.
func (m TempBasalsVw) GetAbsolute() *float64 {
	return m.Absolute
}

// SetAbsolute sets the Absolute value and returns the updated TempBasalsVw.
func (m *TempBasalsVw) SetAbsolute(v *float64) {
	m.Absolute = v
}

// GetPercent returns the Percent value.
func (m TempBasalsVw) GetPercent() *float64 {
	return m.Percent
}

// SetPercent sets the Percent value and returns the updated TempBasalsVw.
func (m *TempBasalsVw) SetPercent(v *float64) {
	m.Percent = v
}

// GetNotes returns the Notes value.
func (m TempBasalsVw) GetNotes() *string {
	return m.Notes
}

// SetNotes sets the Notes value and returns the updated TempBasalsVw.
func (m *TempBasalsVw) SetNotes(v *string) {
	m.Notes = v
}

// GetReason returns the Reason value.
func (m TempBasalsVw) GetReason() *string {
	return m.Reason
}

// SetReason sets the Reason value and returns the updated TempBasalsVw.
func (m *TempBasalsVw) SetReason(v *string) {
	m.Reason = v
}

// GetRawJsonData returns the RawJsonData value.
func (m TempBasalsVw) GetRawJsonData() *string {
	return m.RawJsonData
}

// SetRawJsonData sets the RawJsonData value and returns the updated TempBasalsVw.
func (m *TempBasalsVw) SetRawJsonData(v *string) {
	m.RawJsonData = v
}

// GetDeletedAt returns the DeletedAt value.
func (m TempBasalsVw) GetDeletedAt() *int64 {
	return m.DeletedAt
}

// SetDeletedAt sets the DeletedAt value and returns the updated TempBasalsVw.
func (m *TempBasalsVw) SetDeletedAt(v *int64) {
	m.DeletedAt = v
}

// GetUtcOffsetMinutes returns the UtcOffsetMinutes value.
func (m TempBasalsVw) GetUtcOffsetMinutes() *int64 {
	return m.UtcOffsetMinutes
}

// SetUtcOffsetMinutes sets the UtcOffsetMinutes value and returns the updated TempBasalsVw.
func (m *TempBasalsVw) SetUtcOffsetMinutes(v *int64) {
	m.UtcOffsetMinutes = v
}

// GetOccurredAt returns the OccurredAt value.
func (m TempBasalsVw) GetOccurredAt() *int64 {
	return m.OccurredAt
}

// SetOccurredAt sets the OccurredAt value and returns the updated TempBasalsVw.
func (m *TempBasalsVw) SetOccurredAt(v *int64) {
	m.OccurredAt = v
}

// GetCreatedAt returns the CreatedAt value.
func (m TempBasalsVw) GetCreatedAt() *int64 {
	return m.CreatedAt
}

// SetCreatedAt sets the CreatedAt value and returns the updated TempBasalsVw.
func (m *TempBasalsVw) SetCreatedAt(v *int64) {
	m.CreatedAt = v
}

// GetUpdatedAt returns the UpdatedAt value.
func (m TempBasalsVw) GetUpdatedAt() *int64 {
	return m.UpdatedAt
}

// SetUpdatedAt sets the UpdatedAt value and returns the updated TempBasalsVw.
func (m *TempBasalsVw) SetUpdatedAt(v *int64) {
	m.UpdatedAt = v
}

// GetCreatedByAppName returns the CreatedByAppName value.
func (m TempBasalsVw) GetCreatedByAppName() *string {
	return m.CreatedByAppName
}

// SetCreatedByAppName sets the CreatedByAppName value and returns the updated TempBasalsVw.
func (m *TempBasalsVw) SetCreatedByAppName(v *string) {
	m.CreatedByAppName = v
}

// GetCreatedByDeviceName returns the CreatedByDeviceName value.
func (m TempBasalsVw) GetCreatedByDeviceName() *string {
	return m.CreatedByDeviceName
}

// SetCreatedByDeviceName sets the CreatedByDeviceName value and returns the updated TempBasalsVw.
func (m *TempBasalsVw) SetCreatedByDeviceName(v *string) {
	m.CreatedByDeviceName = v
}

// GetUpdatedByIdentityID returns the UpdatedByIdentityID value.
func (m TempBasalsVw) GetUpdatedByIdentityID() *int64 {
	return m.UpdatedByIdentityID
}

// SetUpdatedByIdentityID sets the UpdatedByIdentityID value and returns the updated TempBasalsVw.
func (m *TempBasalsVw) SetUpdatedByIdentityID(v *int64) {
	m.UpdatedByIdentityID = v
}

// GetUpdatedByIdentityName returns the UpdatedByIdentityName value.
func (m TempBasalsVw) GetUpdatedByIdentityName() *string {
	return m.UpdatedByIdentityName
}

// SetUpdatedByIdentityName sets the UpdatedByIdentityName value and returns the updated TempBasalsVw.
func (m *TempBasalsVw) SetUpdatedByIdentityName(v *string) {
	m.UpdatedByIdentityName = v
}

// GetCreatedByIdentityID returns the CreatedByIdentityID value.
func (m TempBasalsVw) GetCreatedByIdentityID() *int64 {
	return m.CreatedByIdentityID
}

// SetCreatedByIdentityID sets the CreatedByIdentityID value and returns the updated TempBasalsVw.
func (m *TempBasalsVw) SetCreatedByIdentityID(v *int64) {
	m.CreatedByIdentityID = v
}

// GetCreatedByIdentityName returns the CreatedByIdentityName value.
func (m TempBasalsVw) GetCreatedByIdentityName() *string {
	return m.CreatedByIdentityName
}

// SetCreatedByIdentityName sets the CreatedByIdentityName value and returns the updated TempBasalsVw.
func (m *TempBasalsVw) SetCreatedByIdentityName(v *string) {
	m.CreatedByIdentityName = v
}

// GetID returns the ID value.
func (m TempTarget) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated TempTarget.
func (m *TempTarget) SetID(v string) {
	m.ID = v
}

// GetDurationMinutes returns the DurationMinutes value.
func (m TempTarget) GetDurationMinutes() int64 {
	return m.DurationMinutes
}

// SetDurationMinutes sets the DurationMinutes value and returns the updated TempTarget.
func (m *TempTarget) SetDurationMinutes(v int64) {
	m.DurationMinutes = v
}

// GetHighTargetMgdl returns the HighTargetMgdl value.
func (m TempTarget) GetHighTargetMgdl() float64 {
	return m.HighTargetMgdl
}

// SetHighTargetMgdl sets the HighTargetMgdl value and returns the updated TempTarget.
func (m *TempTarget) SetHighTargetMgdl(v float64) {
	m.HighTargetMgdl = v
}

// GetLowTargetMgdl returns the LowTargetMgdl value.
func (m TempTarget) GetLowTargetMgdl() float64 {
	return m.LowTargetMgdl
}

// SetLowTargetMgdl sets the LowTargetMgdl value and returns the updated TempTarget.
func (m *TempTarget) SetLowTargetMgdl(v float64) {
	m.LowTargetMgdl = v
}

// GetInsulinNeedsScaleFactor returns the InsulinNeedsScaleFactor value.
func (m TempTarget) GetInsulinNeedsScaleFactor() *float64 {
	return m.InsulinNeedsScaleFactor
}

// SetInsulinNeedsScaleFactor sets the InsulinNeedsScaleFactor value and returns the updated TempTarget.
func (m *TempTarget) SetInsulinNeedsScaleFactor(v *float64) {
	m.InsulinNeedsScaleFactor = v
}

// GetNotes returns the Notes value.
func (m TempTarget) GetNotes() *string {
	return m.Notes
}

// SetNotes sets the Notes value and returns the updated TempTarget.
func (m *TempTarget) SetNotes(v *string) {
	m.Notes = v
}

// GetReason returns the Reason value.
func (m TempTarget) GetReason() *string {
	return m.Reason
}

// SetReason sets the Reason value and returns the updated TempTarget.
func (m *TempTarget) SetReason(v *string) {
	m.Reason = v
}

// GetRawJsonData returns the RawJsonData value.
func (m TempTarget) GetRawJsonData() *string {
	return m.RawJsonData
}

// SetRawJsonData sets the RawJsonData value and returns the updated TempTarget.
func (m *TempTarget) SetRawJsonData(v *string) {
	m.RawJsonData = v
}

// GetDeletedAt returns the DeletedAt value.
func (m TempTarget) GetDeletedAt() *int64 {
	return m.DeletedAt
}

// SetDeletedAt sets the DeletedAt value and returns the updated TempTarget.
func (m *TempTarget) SetDeletedAt(v *int64) {
	m.DeletedAt = v
}

// GetUtcOffsetMinutes returns the UtcOffsetMinutes value.
func (m TempTarget) GetUtcOffsetMinutes() *int64 {
	return m.UtcOffsetMinutes
}

// SetUtcOffsetMinutes sets the UtcOffsetMinutes value and returns the updated TempTarget.
func (m *TempTarget) SetUtcOffsetMinutes(v *int64) {
	m.UtcOffsetMinutes = v
}

// GetOccurredAt returns the OccurredAt value.
func (m TempTarget) GetOccurredAt() int64 {
	return m.OccurredAt
}

// SetOccurredAt sets the OccurredAt value and returns the updated TempTarget.
func (m *TempTarget) SetOccurredAt(v int64) {
	m.OccurredAt = v
}

// GetCreatedAt returns the CreatedAt value.
func (m TempTarget) GetCreatedAt() int64 {
	return m.CreatedAt
}

// SetCreatedAt sets the CreatedAt value and returns the updated TempTarget.
func (m *TempTarget) SetCreatedAt(v int64) {
	m.CreatedAt = v
}

// GetUpdatedAt returns the UpdatedAt value.
func (m TempTarget) GetUpdatedAt() int64 {
	return m.UpdatedAt
}

// SetUpdatedAt sets the UpdatedAt value and returns the updated TempTarget.
func (m *TempTarget) SetUpdatedAt(v int64) {
	m.UpdatedAt = v
}

// GetCreatedByAppName returns the CreatedByAppName value.
func (m TempTarget) GetCreatedByAppName() *string {
	return m.CreatedByAppName
}

// SetCreatedByAppName sets the CreatedByAppName value and returns the updated TempTarget.
func (m *TempTarget) SetCreatedByAppName(v *string) {
	m.CreatedByAppName = v
}

// GetCreatedByDeviceName returns the CreatedByDeviceName value.
func (m TempTarget) GetCreatedByDeviceName() *string {
	return m.CreatedByDeviceName
}

// SetCreatedByDeviceName sets the CreatedByDeviceName value and returns the updated TempTarget.
func (m *TempTarget) SetCreatedByDeviceName(v *string) {
	m.CreatedByDeviceName = v
}

// GetUpdatedByIdentityID returns the UpdatedByIdentityID value.
func (m TempTarget) GetUpdatedByIdentityID() *int64 {
	return m.UpdatedByIdentityID
}

// SetUpdatedByIdentityID sets the UpdatedByIdentityID value and returns the updated TempTarget.
func (m *TempTarget) SetUpdatedByIdentityID(v *int64) {
	m.UpdatedByIdentityID = v
}

// GetUpdatedByIdentityName returns the UpdatedByIdentityName value.
func (m TempTarget) GetUpdatedByIdentityName() *string {
	return m.UpdatedByIdentityName
}

// SetUpdatedByIdentityName sets the UpdatedByIdentityName value and returns the updated TempTarget.
func (m *TempTarget) SetUpdatedByIdentityName(v *string) {
	m.UpdatedByIdentityName = v
}

// GetCreatedByIdentityID returns the CreatedByIdentityID value.
func (m TempTarget) GetCreatedByIdentityID() *int64 {
	return m.CreatedByIdentityID
}

// SetCreatedByIdentityID sets the CreatedByIdentityID value and returns the updated TempTarget.
func (m *TempTarget) SetCreatedByIdentityID(v *int64) {
	m.CreatedByIdentityID = v
}

// GetCreatedByIdentityName returns the CreatedByIdentityName value.
func (m TempTarget) GetCreatedByIdentityName() string {
	return m.CreatedByIdentityName
}

// SetCreatedByIdentityName sets the CreatedByIdentityName value and returns the updated TempTarget.
func (m *TempTarget) SetCreatedByIdentityName(v string) {
	m.CreatedByIdentityName = v
}

// GetID returns the ID value.
func (m TempTargetsVw) GetID() *string {
	return m.ID
}

// SetID sets the ID value and returns the updated TempTargetsVw.
func (m *TempTargetsVw) SetID(v *string) {
	m.ID = v
}

// GetDurationMinutes returns the DurationMinutes value.
func (m TempTargetsVw) GetDurationMinutes() *int64 {
	return m.DurationMinutes
}

// SetDurationMinutes sets the DurationMinutes value and returns the updated TempTargetsVw.
func (m *TempTargetsVw) SetDurationMinutes(v *int64) {
	m.DurationMinutes = v
}

// GetHighTargetMgdl returns the HighTargetMgdl value.
func (m TempTargetsVw) GetHighTargetMgdl() *float64 {
	return m.HighTargetMgdl
}

// SetHighTargetMgdl sets the HighTargetMgdl value and returns the updated TempTargetsVw.
func (m *TempTargetsVw) SetHighTargetMgdl(v *float64) {
	m.HighTargetMgdl = v
}

// GetLowTargetMgdl returns the LowTargetMgdl value.
func (m TempTargetsVw) GetLowTargetMgdl() *float64 {
	return m.LowTargetMgdl
}

// SetLowTargetMgdl sets the LowTargetMgdl value and returns the updated TempTargetsVw.
func (m *TempTargetsVw) SetLowTargetMgdl(v *float64) {
	m.LowTargetMgdl = v
}

// GetInsulinNeedsScaleFactor returns the InsulinNeedsScaleFactor value.
func (m TempTargetsVw) GetInsulinNeedsScaleFactor() *float64 {
	return m.InsulinNeedsScaleFactor
}

// SetInsulinNeedsScaleFactor sets the InsulinNeedsScaleFactor value and returns the updated TempTargetsVw.
func (m *TempTargetsVw) SetInsulinNeedsScaleFactor(v *float64) {
	m.InsulinNeedsScaleFactor = v
}

// GetNotes returns the Notes value.
func (m TempTargetsVw) GetNotes() *string {
	return m.Notes
}

// SetNotes sets the Notes value and returns the updated TempTargetsVw.
func (m *TempTargetsVw) SetNotes(v *string) {
	m.Notes = v
}

// GetReason returns the Reason value.
func (m TempTargetsVw) GetReason() *string {
	return m.Reason
}

// SetReason sets the Reason value and returns the updated TempTargetsVw.
func (m *TempTargetsVw) SetReason(v *string) {
	m.Reason = v
}

// GetRawJsonData returns the RawJsonData value.
func (m TempTargetsVw) GetRawJsonData() *string {
	return m.RawJsonData
}

// SetRawJsonData sets the RawJsonData value and returns the updated TempTargetsVw.
func (m *TempTargetsVw) SetRawJsonData(v *string) {
	m.RawJsonData = v
}

// GetDeletedAt returns the DeletedAt value.
func (m TempTargetsVw) GetDeletedAt() *int64 {
	return m.DeletedAt
}

// SetDeletedAt sets the DeletedAt value and returns the updated TempTargetsVw.
func (m *TempTargetsVw) SetDeletedAt(v *int64) {
	m.DeletedAt = v
}

// GetUtcOffsetMinutes returns the UtcOffsetMinutes value.
func (m TempTargetsVw) GetUtcOffsetMinutes() *int64 {
	return m.UtcOffsetMinutes
}

// SetUtcOffsetMinutes sets the UtcOffsetMinutes value and returns the updated TempTargetsVw.
func (m *TempTargetsVw) SetUtcOffsetMinutes(v *int64) {
	m.UtcOffsetMinutes = v
}

// GetOccurredAt returns the OccurredAt value.
func (m TempTargetsVw) GetOccurredAt() *int64 {
	return m.OccurredAt
}

// SetOccurredAt sets the OccurredAt value and returns the updated TempTargetsVw.
func (m *TempTargetsVw) SetOccurredAt(v *int64) {
	m.OccurredAt = v
}

// GetCreatedAt returns the CreatedAt value.
func (m TempTargetsVw) GetCreatedAt() *int64 {
	return m.CreatedAt
}

// SetCreatedAt sets the CreatedAt value and returns the updated TempTargetsVw.
func (m *TempTargetsVw) SetCreatedAt(v *int64) {
	m.CreatedAt = v
}

// GetUpdatedAt returns the UpdatedAt value.
func (m TempTargetsVw) GetUpdatedAt() *int64 {
	return m.UpdatedAt
}

// SetUpdatedAt sets the UpdatedAt value and returns the updated TempTargetsVw.
func (m *TempTargetsVw) SetUpdatedAt(v *int64) {
	m.UpdatedAt = v
}

// GetCreatedByAppName returns the CreatedByAppName value.
func (m TempTargetsVw) GetCreatedByAppName() *string {
	return m.CreatedByAppName
}

// SetCreatedByAppName sets the CreatedByAppName value and returns the updated TempTargetsVw.
func (m *TempTargetsVw) SetCreatedByAppName(v *string) {
	m.CreatedByAppName = v
}

// GetCreatedByDeviceName returns the CreatedByDeviceName value.
func (m TempTargetsVw) GetCreatedByDeviceName() *string {
	return m.CreatedByDeviceName
}

// SetCreatedByDeviceName sets the CreatedByDeviceName value and returns the updated TempTargetsVw.
func (m *TempTargetsVw) SetCreatedByDeviceName(v *string) {
	m.CreatedByDeviceName = v
}

// GetUpdatedByIdentityID returns the UpdatedByIdentityID value.
func (m TempTargetsVw) GetUpdatedByIdentityID() *int64 {
	return m.UpdatedByIdentityID
}

// SetUpdatedByIdentityID sets the UpdatedByIdentityID value and returns the updated TempTargetsVw.
func (m *TempTargetsVw) SetUpdatedByIdentityID(v *int64) {
	m.UpdatedByIdentityID = v
}

// GetUpdatedByIdentityName returns the UpdatedByIdentityName value.
func (m TempTargetsVw) GetUpdatedByIdentityName() *string {
	return m.UpdatedByIdentityName
}

// SetUpdatedByIdentityName sets the UpdatedByIdentityName value and returns the updated TempTargetsVw.
func (m *TempTargetsVw) SetUpdatedByIdentityName(v *string) {
	m.UpdatedByIdentityName = v
}

// GetCreatedByIdentityID returns the CreatedByIdentityID value.
func (m TempTargetsVw) GetCreatedByIdentityID() *int64 {
	return m.CreatedByIdentityID
}

// SetCreatedByIdentityID sets the CreatedByIdentityID value and returns the updated TempTargetsVw.
func (m *TempTargetsVw) SetCreatedByIdentityID(v *int64) {
	m.CreatedByIdentityID = v
}

// GetCreatedByIdentityName returns the CreatedByIdentityName value.
func (m TempTargetsVw) GetCreatedByIdentityName() *string {
	return m.CreatedByIdentityName
}

// SetCreatedByIdentityName sets the CreatedByIdentityName value and returns the updated TempTargetsVw.
func (m *TempTargetsVw) SetCreatedByIdentityName(v *string) {
	m.CreatedByIdentityName = v
}

// GetIncludeDeleted returns the IncludeDeleted value.
func (m AllDeviceReportsParams) GetIncludeDeleted() bool {
	return m.IncludeDeleted
}

// SetIncludeDeleted sets the IncludeDeleted value and returns the updated AllDeviceReportsParams.
func (m *AllDeviceReportsParams) SetIncludeDeleted(v bool) {
	m.IncludeDeleted = v
}

// GetSkip returns the Skip value.
func (m AllDeviceReportsParams) GetSkip() int64 {
	return m.Skip
}

// SetSkip sets the Skip value and returns the updated AllDeviceReportsParams.
func (m *AllDeviceReportsParams) SetSkip(v int64) {
	m.Skip = v
}

// GetLimit returns the Limit value.
func (m AllDeviceReportsParams) GetLimit() int64 {
	return m.Limit
}

// SetLimit sets the Limit value and returns the updated AllDeviceReportsParams.
func (m *AllDeviceReportsParams) SetLimit(v int64) {
	m.Limit = v
}

// GetID returns the ID value.
func (m CreateDeviceReportParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated CreateDeviceReportParams.
func (m *CreateDeviceReportParams) SetID(v string) {
	m.ID = v
}

// GetDeviceBatteryPercent returns the DeviceBatteryPercent value.
func (m CreateDeviceReportParams) GetDeviceBatteryPercent() *float64 {
	return m.DeviceBatteryPercent
}

// SetDeviceBatteryPercent sets the DeviceBatteryPercent value and returns the updated CreateDeviceReportParams.
func (m *CreateDeviceReportParams) SetDeviceBatteryPercent(v *float64) {
	m.DeviceBatteryPercent = v
}

// GetIsDeviceCharging returns the IsDeviceCharging value.
func (m CreateDeviceReportParams) GetIsDeviceCharging() *bool {
	return m.IsDeviceCharging
}

// SetIsDeviceCharging sets the IsDeviceCharging value and returns the updated CreateDeviceReportParams.
func (m *CreateDeviceReportParams) SetIsDeviceCharging(v *bool) {
	m.IsDeviceCharging = v
}

// GetPumpState returns the PumpState value.
func (m CreateDeviceReportParams) GetPumpState() *string {
	return m.PumpState
}

// SetPumpState sets the PumpState value and returns the updated CreateDeviceReportParams.
func (m *CreateDeviceReportParams) SetPumpState(v *string) {
	m.PumpState = v
}

// GetPumpReservoir returns the PumpReservoir value.
func (m CreateDeviceReportParams) GetPumpReservoir() *float64 {
	return m.PumpReservoir
}

// SetPumpReservoir sets the PumpReservoir value and returns the updated CreateDeviceReportParams.
func (m *CreateDeviceReportParams) SetPumpReservoir(v *float64) {
	m.PumpReservoir = v
}

// GetPumpBatteryPercent returns the PumpBatteryPercent value.
func (m CreateDeviceReportParams) GetPumpBatteryPercent() *float64 {
	return m.PumpBatteryPercent
}

// SetPumpBatteryPercent sets the PumpBatteryPercent value and returns the updated CreateDeviceReportParams.
func (m *CreateDeviceReportParams) SetPumpBatteryPercent(v *float64) {
	m.PumpBatteryPercent = v
}

// GetPumpChargingState returns the PumpChargingState value.
func (m CreateDeviceReportParams) GetPumpChargingState() *string {
	return m.PumpChargingState
}

// SetPumpChargingState sets the PumpChargingState value and returns the updated CreateDeviceReportParams.
func (m *CreateDeviceReportParams) SetPumpChargingState(v *string) {
	m.PumpChargingState = v
}

// GetPumpClock returns the PumpClock value.
func (m CreateDeviceReportParams) GetPumpClock() *int64 {
	return m.PumpClock
}

// SetPumpClock sets the PumpClock value and returns the updated CreateDeviceReportParams.
func (m *CreateDeviceReportParams) SetPumpClock(v *int64) {
	m.PumpClock = v
}

// GetPumpManufacturer returns the PumpManufacturer value.
func (m CreateDeviceReportParams) GetPumpManufacturer() *string {
	return m.PumpManufacturer
}

// SetPumpManufacturer sets the PumpManufacturer value and returns the updated CreateDeviceReportParams.
func (m *CreateDeviceReportParams) SetPumpManufacturer(v *string) {
	m.PumpManufacturer = v
}

// GetPumpModel returns the PumpModel value.
func (m CreateDeviceReportParams) GetPumpModel() *string {
	return m.PumpModel
}

// SetPumpModel sets the PumpModel value and returns the updated CreateDeviceReportParams.
func (m *CreateDeviceReportParams) SetPumpModel(v *string) {
	m.PumpModel = v
}

// GetOverrideJson returns the OverrideJson value.
func (m CreateDeviceReportParams) GetOverrideJson() *string {
	return m.OverrideJson
}

// SetOverrideJson sets the OverrideJson value and returns the updated CreateDeviceReportParams.
func (m *CreateDeviceReportParams) SetOverrideJson(v *string) {
	m.OverrideJson = v
}

// GetLoopName returns the LoopName value.
func (m CreateDeviceReportParams) GetLoopName() *string {
	return m.LoopName
}

// SetLoopName sets the LoopName value and returns the updated CreateDeviceReportParams.
func (m *CreateDeviceReportParams) SetLoopName(v *string) {
	m.LoopName = v
}

// GetConfigurationSettingsJson returns the ConfigurationSettingsJson value.
func (m CreateDeviceReportParams) GetConfigurationSettingsJson() *string {
	return m.ConfigurationSettingsJson
}

// SetConfigurationSettingsJson sets the ConfigurationSettingsJson value and returns the updated CreateDeviceReportParams.
func (m *CreateDeviceReportParams) SetConfigurationSettingsJson(v *string) {
	m.ConfigurationSettingsJson = v
}

// GetLoopSettingsJson returns the LoopSettingsJson value.
func (m CreateDeviceReportParams) GetLoopSettingsJson() *string {
	return m.LoopSettingsJson
}

// SetLoopSettingsJson sets the LoopSettingsJson value and returns the updated CreateDeviceReportParams.
func (m *CreateDeviceReportParams) SetLoopSettingsJson(v *string) {
	m.LoopSettingsJson = v
}

// GetLoopIob returns the LoopIob value.
func (m CreateDeviceReportParams) GetLoopIob() *float64 {
	return m.LoopIob
}

// SetLoopIob sets the LoopIob value and returns the updated CreateDeviceReportParams.
func (m *CreateDeviceReportParams) SetLoopIob(v *float64) {
	m.LoopIob = v
}

// GetLoopIobBasal returns the LoopIobBasal value.
func (m CreateDeviceReportParams) GetLoopIobBasal() *float64 {
	return m.LoopIobBasal
}

// SetLoopIobBasal sets the LoopIobBasal value and returns the updated CreateDeviceReportParams.
func (m *CreateDeviceReportParams) SetLoopIobBasal(v *float64) {
	m.LoopIobBasal = v
}

// GetLoopIobTreatment returns the LoopIobTreatment value.
func (m CreateDeviceReportParams) GetLoopIobTreatment() *float64 {
	return m.LoopIobTreatment
}

// SetLoopIobTreatment sets the LoopIobTreatment value and returns the updated CreateDeviceReportParams.
func (m *CreateDeviceReportParams) SetLoopIobTreatment(v *float64) {
	m.LoopIobTreatment = v
}

// GetLoopCob returns the LoopCob value.
func (m CreateDeviceReportParams) GetLoopCob() *float64 {
	return m.LoopCob
}

// SetLoopCob sets the LoopCob value and returns the updated CreateDeviceReportParams.
func (m *CreateDeviceReportParams) SetLoopCob(v *float64) {
	m.LoopCob = v
}

// GetLoopPredictionsStartTime returns the LoopPredictionsStartTime value.
func (m CreateDeviceReportParams) GetLoopPredictionsStartTime() *int64 {
	return m.LoopPredictionsStartTime
}

// SetLoopPredictionsStartTime sets the LoopPredictionsStartTime value and returns the updated CreateDeviceReportParams.
func (m *CreateDeviceReportParams) SetLoopPredictionsStartTime(v *int64) {
	m.LoopPredictionsStartTime = v
}

// GetLoopPredictionsDefault returns the LoopPredictionsDefault value.
func (m CreateDeviceReportParams) GetLoopPredictionsDefault() *string {
	return m.LoopPredictionsDefault
}

// SetLoopPredictionsDefault sets the LoopPredictionsDefault value and returns the updated CreateDeviceReportParams.
func (m *CreateDeviceReportParams) SetLoopPredictionsDefault(v *string) {
	m.LoopPredictionsDefault = v
}

// GetLoopPredictionsZt returns the LoopPredictionsZt value.
func (m CreateDeviceReportParams) GetLoopPredictionsZt() *string {
	return m.LoopPredictionsZt
}

// SetLoopPredictionsZt sets the LoopPredictionsZt value and returns the updated CreateDeviceReportParams.
func (m *CreateDeviceReportParams) SetLoopPredictionsZt(v *string) {
	m.LoopPredictionsZt = v
}

// GetUtcOffsetMinutes returns the UtcOffsetMinutes value.
func (m CreateDeviceReportParams) GetUtcOffsetMinutes() *int64 {
	return m.UtcOffsetMinutes
}

// SetUtcOffsetMinutes sets the UtcOffsetMinutes value and returns the updated CreateDeviceReportParams.
func (m *CreateDeviceReportParams) SetUtcOffsetMinutes(v *int64) {
	m.UtcOffsetMinutes = v
}

// GetOccurredAt returns the OccurredAt value.
func (m CreateDeviceReportParams) GetOccurredAt() int64 {
	return m.OccurredAt
}

// SetOccurredAt sets the OccurredAt value and returns the updated CreateDeviceReportParams.
func (m *CreateDeviceReportParams) SetOccurredAt(v int64) {
	m.OccurredAt = v
}

// GetRawJsonData returns the RawJsonData value.
func (m CreateDeviceReportParams) GetRawJsonData() *string {
	return m.RawJsonData
}

// SetRawJsonData sets the RawJsonData value and returns the updated CreateDeviceReportParams.
func (m *CreateDeviceReportParams) SetRawJsonData(v *string) {
	m.RawJsonData = v
}

// GetCreatedByAppName returns the CreatedByAppName value.
func (m CreateDeviceReportParams) GetCreatedByAppName() *string {
	return m.CreatedByAppName
}

// SetCreatedByAppName sets the CreatedByAppName value and returns the updated CreateDeviceReportParams.
func (m *CreateDeviceReportParams) SetCreatedByAppName(v *string) {
	m.CreatedByAppName = v
}

// GetCreatedByDeviceName returns the CreatedByDeviceName value.
func (m CreateDeviceReportParams) GetCreatedByDeviceName() *string {
	return m.CreatedByDeviceName
}

// SetCreatedByDeviceName sets the CreatedByDeviceName value and returns the updated CreateDeviceReportParams.
func (m *CreateDeviceReportParams) SetCreatedByDeviceName(v *string) {
	m.CreatedByDeviceName = v
}

// GetCreatedByIdentityName returns the CreatedByIdentityName value.
func (m CreateDeviceReportParams) GetCreatedByIdentityName() string {
	return m.CreatedByIdentityName
}

// SetCreatedByIdentityName sets the CreatedByIdentityName value and returns the updated CreateDeviceReportParams.
func (m *CreateDeviceReportParams) SetCreatedByIdentityName(v string) {
	m.CreatedByIdentityName = v
}

// GetCreatedByIdentityID returns the CreatedByIdentityID value.
func (m CreateDeviceReportParams) GetCreatedByIdentityID() *int64 {
	return m.CreatedByIdentityID
}

// SetCreatedByIdentityID sets the CreatedByIdentityID value and returns the updated CreateDeviceReportParams.
func (m *CreateDeviceReportParams) SetCreatedByIdentityID(v *int64) {
	m.CreatedByIdentityID = v
}

// GetDevice returns the Device value.
func (m DeviceReportsForDeviceParams) GetDevice() *string {
	return m.Device
}

// SetDevice sets the Device value and returns the updated DeviceReportsForDeviceParams.
func (m *DeviceReportsForDeviceParams) SetDevice(v *string) {
	m.Device = v
}

// GetSkip returns the Skip value.
func (m DeviceReportsForDeviceParams) GetSkip() int64 {
	return m.Skip
}

// SetSkip sets the Skip value and returns the updated DeviceReportsForDeviceParams.
func (m *DeviceReportsForDeviceParams) SetSkip(v int64) {
	m.Skip = v
}

// GetLimit returns the Limit value.
func (m DeviceReportsForDeviceParams) GetLimit() int64 {
	return m.Limit
}

// SetLimit sets the Limit value and returns the updated DeviceReportsForDeviceParams.
func (m *DeviceReportsForDeviceParams) SetLimit(v int64) {
	m.Limit = v
}

// GetID returns the ID value.
func (m GetDeviceReportParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated GetDeviceReportParams.
func (m *GetDeviceReportParams) SetID(v string) {
	m.ID = v
}

// GetIncludeDeleted returns the IncludeDeleted value.
func (m GetDeviceReportParams) GetIncludeDeleted() bool {
	return m.IncludeDeleted
}

// SetIncludeDeleted sets the IncludeDeleted value and returns the updated GetDeviceReportParams.
func (m *GetDeviceReportParams) SetIncludeDeleted(v bool) {
	m.IncludeDeleted = v
}

// GetID returns the ID value.
func (m HardDeleteDeviceReportParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated HardDeleteDeviceReportParams.
func (m *HardDeleteDeviceReportParams) SetID(v string) {
	m.ID = v
}

// GetSince returns the Since value.
func (m RecentDeviceNamesParams) GetSince() int64 {
	return m.Since
}

// SetSince sets the Since value and returns the updated RecentDeviceNamesParams.
func (m *RecentDeviceNamesParams) SetSince(v int64) {
	m.Since = v
}

// GetLimit returns the Limit value.
func (m RecentDeviceNamesParams) GetLimit() int64 {
	return m.Limit
}

// SetLimit sets the Limit value and returns the updated RecentDeviceNamesParams.
func (m *RecentDeviceNamesParams) SetLimit(v int64) {
	m.Limit = v
}

// GetDateGTE returns the DateGTE value.
func (m SearchDeviceReportsParams) GetDateGTE() *int64 {
	return m.DateGTE
}

// SetDateGTE sets the DateGTE value and returns the updated SearchDeviceReportsParams.
func (m *SearchDeviceReportsParams) SetDateGTE(v *int64) {
	m.DateGTE = v
}

// GetDateGT returns the DateGT value.
func (m SearchDeviceReportsParams) GetDateGT() *int64 {
	return m.DateGT
}

// SetDateGT sets the DateGT value and returns the updated SearchDeviceReportsParams.
func (m *SearchDeviceReportsParams) SetDateGT(v *int64) {
	m.DateGT = v
}

// GetDateEQ returns the DateEQ value.
func (m SearchDeviceReportsParams) GetDateEQ() *int64 {
	return m.DateEQ
}

// SetDateEQ sets the DateEQ value and returns the updated SearchDeviceReportsParams.
func (m *SearchDeviceReportsParams) SetDateEQ(v *int64) {
	m.DateEQ = v
}

// GetDateLTE returns the DateLTE value.
func (m SearchDeviceReportsParams) GetDateLTE() *int64 {
	return m.DateLTE
}

// SetDateLTE sets the DateLTE value and returns the updated SearchDeviceReportsParams.
func (m *SearchDeviceReportsParams) SetDateLTE(v *int64) {
	m.DateLTE = v
}

// GetDateLT returns the DateLT value.
func (m SearchDeviceReportsParams) GetDateLT() *int64 {
	return m.DateLT
}

// SetDateLT sets the DateLT value and returns the updated SearchDeviceReportsParams.
func (m *SearchDeviceReportsParams) SetDateLT(v *int64) {
	m.DateLT = v
}

// GetCreatedAtGTE returns the CreatedAtGTE value.
func (m SearchDeviceReportsParams) GetCreatedAtGTE() *int64 {
	return m.CreatedAtGTE
}

// SetCreatedAtGTE sets the CreatedAtGTE value and returns the updated SearchDeviceReportsParams.
func (m *SearchDeviceReportsParams) SetCreatedAtGTE(v *int64) {
	m.CreatedAtGTE = v
}

// GetCreatedAtLTE returns the CreatedAtLTE value.
func (m SearchDeviceReportsParams) GetCreatedAtLTE() *int64 {
	return m.CreatedAtLTE
}

// SetCreatedAtLTE sets the CreatedAtLTE value and returns the updated SearchDeviceReportsParams.
func (m *SearchDeviceReportsParams) SetCreatedAtLTE(v *int64) {
	m.CreatedAtLTE = v
}

// GetCreatedAtGT returns the CreatedAtGT value.
func (m SearchDeviceReportsParams) GetCreatedAtGT() *int64 {
	return m.CreatedAtGT
}

// SetCreatedAtGT sets the CreatedAtGT value and returns the updated SearchDeviceReportsParams.
func (m *SearchDeviceReportsParams) SetCreatedAtGT(v *int64) {
	m.CreatedAtGT = v
}

// GetCreatedAtLT returns the CreatedAtLT value.
func (m SearchDeviceReportsParams) GetCreatedAtLT() *int64 {
	return m.CreatedAtLT
}

// SetCreatedAtLT sets the CreatedAtLT value and returns the updated SearchDeviceReportsParams.
func (m *SearchDeviceReportsParams) SetCreatedAtLT(v *int64) {
	m.CreatedAtLT = v
}

// GetUpdatedGTE returns the UpdatedGTE value.
func (m SearchDeviceReportsParams) GetUpdatedGTE() *int64 {
	return m.UpdatedGTE
}

// SetUpdatedGTE sets the UpdatedGTE value and returns the updated SearchDeviceReportsParams.
func (m *SearchDeviceReportsParams) SetUpdatedGTE(v *int64) {
	m.UpdatedGTE = v
}

// GetUpdatedLTE returns the UpdatedLTE value.
func (m SearchDeviceReportsParams) GetUpdatedLTE() *int64 {
	return m.UpdatedLTE
}

// SetUpdatedLTE sets the UpdatedLTE value and returns the updated SearchDeviceReportsParams.
func (m *SearchDeviceReportsParams) SetUpdatedLTE(v *int64) {
	m.UpdatedLTE = v
}

// GetUpdatedGT returns the UpdatedGT value.
func (m SearchDeviceReportsParams) GetUpdatedGT() *int64 {
	return m.UpdatedGT
}

// SetUpdatedGT sets the UpdatedGT value and returns the updated SearchDeviceReportsParams.
func (m *SearchDeviceReportsParams) SetUpdatedGT(v *int64) {
	m.UpdatedGT = v
}

// GetUpdatedLT returns the UpdatedLT value.
func (m SearchDeviceReportsParams) GetUpdatedLT() *int64 {
	return m.UpdatedLT
}

// SetUpdatedLT sets the UpdatedLT value and returns the updated SearchDeviceReportsParams.
func (m *SearchDeviceReportsParams) SetUpdatedLT(v *int64) {
	m.UpdatedLT = v
}

// GetIDEq returns the IDEq value.
func (m SearchDeviceReportsParams) GetIDEq() *string {
	return m.IDEq
}

// SetIDEq sets the IDEq value and returns the updated SearchDeviceReportsParams.
func (m *SearchDeviceReportsParams) SetIDEq(v *string) {
	m.IDEq = v
}

// GetIDIn returns the IDIn value.
func (m SearchDeviceReportsParams) GetIDIn() *string {
	return m.IDIn
}

// SetIDIn sets the IDIn value and returns the updated SearchDeviceReportsParams.
func (m *SearchDeviceReportsParams) SetIDIn(v *string) {
	m.IDIn = v
}

// GetDeviceEq returns the DeviceEq value.
func (m SearchDeviceReportsParams) GetDeviceEq() *string {
	return m.DeviceEq
}

// SetDeviceEq sets the DeviceEq value and returns the updated SearchDeviceReportsParams.
func (m *SearchDeviceReportsParams) SetDeviceEq(v *string) {
	m.DeviceEq = v
}

// GetAppEq returns the AppEq value.
func (m SearchDeviceReportsParams) GetAppEq() *string {
	return m.AppEq
}

// SetAppEq sets the AppEq value and returns the updated SearchDeviceReportsParams.
func (m *SearchDeviceReportsParams) SetAppEq(v *string) {
	m.AppEq = v
}

// GetIncludeDeleted returns the IncludeDeleted value.
func (m SearchDeviceReportsParams) GetIncludeDeleted() *bool {
	return m.IncludeDeleted
}

// SetIncludeDeleted sets the IncludeDeleted value and returns the updated SearchDeviceReportsParams.
func (m *SearchDeviceReportsParams) SetIncludeDeleted(v *bool) {
	m.IncludeDeleted = v
}

// GetSkip returns the Skip value.
func (m SearchDeviceReportsParams) GetSkip() int64 {
	return m.Skip
}

// SetSkip sets the Skip value and returns the updated SearchDeviceReportsParams.
func (m *SearchDeviceReportsParams) SetSkip(v int64) {
	m.Skip = v
}

// GetLimit returns the Limit value.
func (m SearchDeviceReportsParams) GetLimit() int64 {
	return m.Limit
}

// SetLimit sets the Limit value and returns the updated SearchDeviceReportsParams.
func (m *SearchDeviceReportsParams) SetLimit(v int64) {
	m.Limit = v
}

// GetUpdatedByIdentityName returns the UpdatedByIdentityName value.
func (m SoftDeleteDeviceReportParams) GetUpdatedByIdentityName() *string {
	return m.UpdatedByIdentityName
}

// SetUpdatedByIdentityName sets the UpdatedByIdentityName value and returns the updated SoftDeleteDeviceReportParams.
func (m *SoftDeleteDeviceReportParams) SetUpdatedByIdentityName(v *string) {
	m.UpdatedByIdentityName = v
}

// GetID returns the ID value.
func (m SoftDeleteDeviceReportParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated SoftDeleteDeviceReportParams.
func (m *SoftDeleteDeviceReportParams) SetID(v string) {
	m.ID = v
}

// GetRawJsonData returns the RawJsonData value.
func (m UpdateDeviceReportParams) GetRawJsonData() *string {
	return m.RawJsonData
}

// SetRawJsonData sets the RawJsonData value and returns the updated UpdateDeviceReportParams.
func (m *UpdateDeviceReportParams) SetRawJsonData(v *string) {
	m.RawJsonData = v
}

// GetUpdatedByIdentityID returns the UpdatedByIdentityID value.
func (m UpdateDeviceReportParams) GetUpdatedByIdentityID() *int64 {
	return m.UpdatedByIdentityID
}

// SetUpdatedByIdentityID sets the UpdatedByIdentityID value and returns the updated UpdateDeviceReportParams.
func (m *UpdateDeviceReportParams) SetUpdatedByIdentityID(v *int64) {
	m.UpdatedByIdentityID = v
}

// GetUpdatedByIdentityName returns the UpdatedByIdentityName value.
func (m UpdateDeviceReportParams) GetUpdatedByIdentityName() *string {
	return m.UpdatedByIdentityName
}

// SetUpdatedByIdentityName sets the UpdatedByIdentityName value and returns the updated UpdateDeviceReportParams.
func (m *UpdateDeviceReportParams) SetUpdatedByIdentityName(v *string) {
	m.UpdatedByIdentityName = v
}

// GetID returns the ID value.
func (m UpdateDeviceReportParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated UpdateDeviceReportParams.
func (m *UpdateDeviceReportParams) SetID(v string) {
	m.ID = v
}

// GetIncludeDeleted returns the IncludeDeleted value.
func (m AllActivitiesParams) GetIncludeDeleted() bool {
	return m.IncludeDeleted
}

// SetIncludeDeleted sets the IncludeDeleted value and returns the updated AllActivitiesParams.
func (m *AllActivitiesParams) SetIncludeDeleted(v bool) {
	m.IncludeDeleted = v
}

// GetSkip returns the Skip value.
func (m AllActivitiesParams) GetSkip() int64 {
	return m.Skip
}

// SetSkip sets the Skip value and returns the updated AllActivitiesParams.
func (m *AllActivitiesParams) SetSkip(v int64) {
	m.Skip = v
}

// GetLimit returns the Limit value.
func (m AllActivitiesParams) GetLimit() int64 {
	return m.Limit
}

// SetLimit sets the Limit value and returns the updated AllActivitiesParams.
func (m *AllActivitiesParams) SetLimit(v int64) {
	m.Limit = v
}

// GetID returns the ID value.
func (m CreateActivityParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated CreateActivityParams.
func (m *CreateActivityParams) SetID(v string) {
	m.ID = v
}

// GetName returns the Name value.
func (m CreateActivityParams) GetName() *string {
	return m.Name
}

// SetName sets the Name value and returns the updated CreateActivityParams.
func (m *CreateActivityParams) SetName(v *string) {
	m.Name = v
}

// GetActivityType returns the ActivityType value.
func (m CreateActivityParams) GetActivityType() string {
	return m.ActivityType
}

// SetActivityType sets the ActivityType value and returns the updated CreateActivityParams.
func (m *CreateActivityParams) SetActivityType(v string) {
	m.ActivityType = v
}

// GetDurationMinutes returns the DurationMinutes value.
func (m CreateActivityParams) GetDurationMinutes() *int64 {
	return m.DurationMinutes
}

// SetDurationMinutes sets the DurationMinutes value and returns the updated CreateActivityParams.
func (m *CreateActivityParams) SetDurationMinutes(v *int64) {
	m.DurationMinutes = v
}

// GetCalories returns the Calories value.
func (m CreateActivityParams) GetCalories() *int64 {
	return m.Calories
}

// SetCalories sets the Calories value and returns the updated CreateActivityParams.
func (m *CreateActivityParams) SetCalories(v *int64) {
	m.Calories = v
}

// GetSteps returns the Steps value.
func (m CreateActivityParams) GetSteps() *int64 {
	return m.Steps
}

// SetSteps sets the Steps value and returns the updated CreateActivityParams.
func (m *CreateActivityParams) SetSteps(v *int64) {
	m.Steps = v
}

// GetHeartRate returns the HeartRate value.
func (m CreateActivityParams) GetHeartRate() *int64 {
	return m.HeartRate
}

// SetHeartRate sets the HeartRate value and returns the updated CreateActivityParams.
func (m *CreateActivityParams) SetHeartRate(v *int64) {
	m.HeartRate = v
}

// GetNotes returns the Notes value.
func (m CreateActivityParams) GetNotes() *string {
	return m.Notes
}

// SetNotes sets the Notes value and returns the updated CreateActivityParams.
func (m *CreateActivityParams) SetNotes(v *string) {
	m.Notes = v
}

// GetReason returns the Reason value.
func (m CreateActivityParams) GetReason() *string {
	return m.Reason
}

// SetReason sets the Reason value and returns the updated CreateActivityParams.
func (m *CreateActivityParams) SetReason(v *string) {
	m.Reason = v
}

// GetOccurredAt returns the OccurredAt value.
func (m CreateActivityParams) GetOccurredAt() int64 {
	return m.OccurredAt
}

// SetOccurredAt sets the OccurredAt value and returns the updated CreateActivityParams.
func (m *CreateActivityParams) SetOccurredAt(v int64) {
	m.OccurredAt = v
}

// GetUtcOffsetMinutes returns the UtcOffsetMinutes value.
func (m CreateActivityParams) GetUtcOffsetMinutes() *int64 {
	return m.UtcOffsetMinutes
}

// SetUtcOffsetMinutes sets the UtcOffsetMinutes value and returns the updated CreateActivityParams.
func (m *CreateActivityParams) SetUtcOffsetMinutes(v *int64) {
	m.UtcOffsetMinutes = v
}

// GetRawJsonData returns the RawJsonData value.
func (m CreateActivityParams) GetRawJsonData() *string {
	return m.RawJsonData
}

// SetRawJsonData sets the RawJsonData value and returns the updated CreateActivityParams.
func (m *CreateActivityParams) SetRawJsonData(v *string) {
	m.RawJsonData = v
}

// GetCreatedByAppName returns the CreatedByAppName value.
func (m CreateActivityParams) GetCreatedByAppName() *string {
	return m.CreatedByAppName
}

// SetCreatedByAppName sets the CreatedByAppName value and returns the updated CreateActivityParams.
func (m *CreateActivityParams) SetCreatedByAppName(v *string) {
	m.CreatedByAppName = v
}

// GetCreatedByDeviceName returns the CreatedByDeviceName value.
func (m CreateActivityParams) GetCreatedByDeviceName() *string {
	return m.CreatedByDeviceName
}

// SetCreatedByDeviceName sets the CreatedByDeviceName value and returns the updated CreateActivityParams.
func (m *CreateActivityParams) SetCreatedByDeviceName(v *string) {
	m.CreatedByDeviceName = v
}

// GetCreatedByIdentityName returns the CreatedByIdentityName value.
func (m CreateActivityParams) GetCreatedByIdentityName() string {
	return m.CreatedByIdentityName
}

// SetCreatedByIdentityName sets the CreatedByIdentityName value and returns the updated CreateActivityParams.
func (m *CreateActivityParams) SetCreatedByIdentityName(v string) {
	m.CreatedByIdentityName = v
}

// GetCreatedByIdentityID returns the CreatedByIdentityID value.
func (m CreateActivityParams) GetCreatedByIdentityID() *int64 {
	return m.CreatedByIdentityID
}

// SetCreatedByIdentityID sets the CreatedByIdentityID value and returns the updated CreateActivityParams.
func (m *CreateActivityParams) SetCreatedByIdentityID(v *int64) {
	m.CreatedByIdentityID = v
}

// GetID returns the ID value.
func (m GetActivityParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated GetActivityParams.
func (m *GetActivityParams) SetID(v string) {
	m.ID = v
}

// GetIncludeDeleted returns the IncludeDeleted value.
func (m GetActivityParams) GetIncludeDeleted() bool {
	return m.IncludeDeleted
}

// SetIncludeDeleted sets the IncludeDeleted value and returns the updated GetActivityParams.
func (m *GetActivityParams) SetIncludeDeleted(v bool) {
	m.IncludeDeleted = v
}

// GetActivityType returns the ActivityType value.
func (m HardDeleteActivitiesByTypeParams) GetActivityType() string {
	return m.ActivityType
}

// SetActivityType sets the ActivityType value and returns the updated HardDeleteActivitiesByTypeParams.
func (m *HardDeleteActivitiesByTypeParams) SetActivityType(v string) {
	m.ActivityType = v
}

// GetID returns the ID value.
func (m HardDeleteActivityParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated HardDeleteActivityParams.
func (m *HardDeleteActivityParams) SetID(v string) {
	m.ID = v
}

// GetDateGTE returns the DateGTE value.
func (m SearchActivitiesParams) GetDateGTE() *int64 {
	return m.DateGTE
}

// SetDateGTE sets the DateGTE value and returns the updated SearchActivitiesParams.
func (m *SearchActivitiesParams) SetDateGTE(v *int64) {
	m.DateGTE = v
}

// GetDateGT returns the DateGT value.
func (m SearchActivitiesParams) GetDateGT() *int64 {
	return m.DateGT
}

// SetDateGT sets the DateGT value and returns the updated SearchActivitiesParams.
func (m *SearchActivitiesParams) SetDateGT(v *int64) {
	m.DateGT = v
}

// GetDateEQ returns the DateEQ value.
func (m SearchActivitiesParams) GetDateEQ() *int64 {
	return m.DateEQ
}

// SetDateEQ sets the DateEQ value and returns the updated SearchActivitiesParams.
func (m *SearchActivitiesParams) SetDateEQ(v *int64) {
	m.DateEQ = v
}

// GetDateLTE returns the DateLTE value.
func (m SearchActivitiesParams) GetDateLTE() *int64 {
	return m.DateLTE
}

// SetDateLTE sets the DateLTE value and returns the updated SearchActivitiesParams.
func (m *SearchActivitiesParams) SetDateLTE(v *int64) {
	m.DateLTE = v
}

// GetDateLT returns the DateLT value.
func (m SearchActivitiesParams) GetDateLT() *int64 {
	return m.DateLT
}

// SetDateLT sets the DateLT value and returns the updated SearchActivitiesParams.
func (m *SearchActivitiesParams) SetDateLT(v *int64) {
	m.DateLT = v
}

// GetCreatedAtGTE returns the CreatedAtGTE value.
func (m SearchActivitiesParams) GetCreatedAtGTE() *int64 {
	return m.CreatedAtGTE
}

// SetCreatedAtGTE sets the CreatedAtGTE value and returns the updated SearchActivitiesParams.
func (m *SearchActivitiesParams) SetCreatedAtGTE(v *int64) {
	m.CreatedAtGTE = v
}

// GetCreatedAtLTE returns the CreatedAtLTE value.
func (m SearchActivitiesParams) GetCreatedAtLTE() *int64 {
	return m.CreatedAtLTE
}

// SetCreatedAtLTE sets the CreatedAtLTE value and returns the updated SearchActivitiesParams.
func (m *SearchActivitiesParams) SetCreatedAtLTE(v *int64) {
	m.CreatedAtLTE = v
}

// GetCreatedAtGT returns the CreatedAtGT value.
func (m SearchActivitiesParams) GetCreatedAtGT() *int64 {
	return m.CreatedAtGT
}

// SetCreatedAtGT sets the CreatedAtGT value and returns the updated SearchActivitiesParams.
func (m *SearchActivitiesParams) SetCreatedAtGT(v *int64) {
	m.CreatedAtGT = v
}

// GetCreatedAtLT returns the CreatedAtLT value.
func (m SearchActivitiesParams) GetCreatedAtLT() *int64 {
	return m.CreatedAtLT
}

// SetCreatedAtLT sets the CreatedAtLT value and returns the updated SearchActivitiesParams.
func (m *SearchActivitiesParams) SetCreatedAtLT(v *int64) {
	m.CreatedAtLT = v
}

// GetUpdatedGTE returns the UpdatedGTE value.
func (m SearchActivitiesParams) GetUpdatedGTE() *int64 {
	return m.UpdatedGTE
}

// SetUpdatedGTE sets the UpdatedGTE value and returns the updated SearchActivitiesParams.
func (m *SearchActivitiesParams) SetUpdatedGTE(v *int64) {
	m.UpdatedGTE = v
}

// GetUpdatedLTE returns the UpdatedLTE value.
func (m SearchActivitiesParams) GetUpdatedLTE() *int64 {
	return m.UpdatedLTE
}

// SetUpdatedLTE sets the UpdatedLTE value and returns the updated SearchActivitiesParams.
func (m *SearchActivitiesParams) SetUpdatedLTE(v *int64) {
	m.UpdatedLTE = v
}

// GetUpdatedGT returns the UpdatedGT value.
func (m SearchActivitiesParams) GetUpdatedGT() *int64 {
	return m.UpdatedGT
}

// SetUpdatedGT sets the UpdatedGT value and returns the updated SearchActivitiesParams.
func (m *SearchActivitiesParams) SetUpdatedGT(v *int64) {
	m.UpdatedGT = v
}

// GetUpdatedLT returns the UpdatedLT value.
func (m SearchActivitiesParams) GetUpdatedLT() *int64 {
	return m.UpdatedLT
}

// SetUpdatedLT sets the UpdatedLT value and returns the updated SearchActivitiesParams.
func (m *SearchActivitiesParams) SetUpdatedLT(v *int64) {
	m.UpdatedLT = v
}

// GetIDEq returns the IDEq value.
func (m SearchActivitiesParams) GetIDEq() *string {
	return m.IDEq
}

// SetIDEq sets the IDEq value and returns the updated SearchActivitiesParams.
func (m *SearchActivitiesParams) SetIDEq(v *string) {
	m.IDEq = v
}

// GetIDIn returns the IDIn value.
func (m SearchActivitiesParams) GetIDIn() *string {
	return m.IDIn
}

// SetIDIn sets the IDIn value and returns the updated SearchActivitiesParams.
func (m *SearchActivitiesParams) SetIDIn(v *string) {
	m.IDIn = v
}

// GetNameEq returns the NameEq value.
func (m SearchActivitiesParams) GetNameEq() *string {
	return m.NameEq
}

// SetNameEq sets the NameEq value and returns the updated SearchActivitiesParams.
func (m *SearchActivitiesParams) SetNameEq(v *string) {
	m.NameEq = v
}

// GetDeviceEq returns the DeviceEq value.
func (m SearchActivitiesParams) GetDeviceEq() *string {
	return m.DeviceEq
}

// SetDeviceEq sets the DeviceEq value and returns the updated SearchActivitiesParams.
func (m *SearchActivitiesParams) SetDeviceEq(v *string) {
	m.DeviceEq = v
}

// GetAppEq returns the AppEq value.
func (m SearchActivitiesParams) GetAppEq() *string {
	return m.AppEq
}

// SetAppEq sets the AppEq value and returns the updated SearchActivitiesParams.
func (m *SearchActivitiesParams) SetAppEq(v *string) {
	m.AppEq = v
}

// GetTypeEq returns the TypeEq value.
func (m SearchActivitiesParams) GetTypeEq() *string {
	return m.TypeEq
}

// SetTypeEq sets the TypeEq value and returns the updated SearchActivitiesParams.
func (m *SearchActivitiesParams) SetTypeEq(v *string) {
	m.TypeEq = v
}

// GetTypeNeq returns the TypeNeq value.
func (m SearchActivitiesParams) GetTypeNeq() *string {
	return m.TypeNeq
}

// SetTypeNeq sets the TypeNeq value and returns the updated SearchActivitiesParams.
func (m *SearchActivitiesParams) SetTypeNeq(v *string) {
	m.TypeNeq = v
}

// GetTypeIn returns the TypeIn value.
func (m SearchActivitiesParams) GetTypeIn() *string {
	return m.TypeIn
}

// SetTypeIn sets the TypeIn value and returns the updated SearchActivitiesParams.
func (m *SearchActivitiesParams) SetTypeIn(v *string) {
	m.TypeIn = v
}

// GetTypeNin returns the TypeNin value.
func (m SearchActivitiesParams) GetTypeNin() *string {
	return m.TypeNin
}

// SetTypeNin sets the TypeNin value and returns the updated SearchActivitiesParams.
func (m *SearchActivitiesParams) SetTypeNin(v *string) {
	m.TypeNin = v
}

// GetDurationEQ returns the DurationEQ value.
func (m SearchActivitiesParams) GetDurationEQ() *int64 {
	return m.DurationEQ
}

// SetDurationEQ sets the DurationEQ value and returns the updated SearchActivitiesParams.
func (m *SearchActivitiesParams) SetDurationEQ(v *int64) {
	m.DurationEQ = v
}

// GetDurationGTE returns the DurationGTE value.
func (m SearchActivitiesParams) GetDurationGTE() *int64 {
	return m.DurationGTE
}

// SetDurationGTE sets the DurationGTE value and returns the updated SearchActivitiesParams.
func (m *SearchActivitiesParams) SetDurationGTE(v *int64) {
	m.DurationGTE = v
}

// GetDurationGT returns the DurationGT value.
func (m SearchActivitiesParams) GetDurationGT() *int64 {
	return m.DurationGT
}

// SetDurationGT sets the DurationGT value and returns the updated SearchActivitiesParams.
func (m *SearchActivitiesParams) SetDurationGT(v *int64) {
	m.DurationGT = v
}

// GetDurationLTE returns the DurationLTE value.
func (m SearchActivitiesParams) GetDurationLTE() *int64 {
	return m.DurationLTE
}

// SetDurationLTE sets the DurationLTE value and returns the updated SearchActivitiesParams.
func (m *SearchActivitiesParams) SetDurationLTE(v *int64) {
	m.DurationLTE = v
}

// GetDurationLT returns the DurationLT value.
func (m SearchActivitiesParams) GetDurationLT() *int64 {
	return m.DurationLT
}

// SetDurationLT sets the DurationLT value and returns the updated SearchActivitiesParams.
func (m *SearchActivitiesParams) SetDurationLT(v *int64) {
	m.DurationLT = v
}

// GetIncludeDeleted returns the IncludeDeleted value.
func (m SearchActivitiesParams) GetIncludeDeleted() *bool {
	return m.IncludeDeleted
}

// SetIncludeDeleted sets the IncludeDeleted value and returns the updated SearchActivitiesParams.
func (m *SearchActivitiesParams) SetIncludeDeleted(v *bool) {
	m.IncludeDeleted = v
}

// GetSkip returns the Skip value.
func (m SearchActivitiesParams) GetSkip() int64 {
	return m.Skip
}

// SetSkip sets the Skip value and returns the updated SearchActivitiesParams.
func (m *SearchActivitiesParams) SetSkip(v int64) {
	m.Skip = v
}

// GetLimit returns the Limit value.
func (m SearchActivitiesParams) GetLimit() int64 {
	return m.Limit
}

// SetLimit sets the Limit value and returns the updated SearchActivitiesParams.
func (m *SearchActivitiesParams) SetLimit(v int64) {
	m.Limit = v
}

// GetUpdatedByIdentityName returns the UpdatedByIdentityName value.
func (m SoftDeleteActivitiesByTypeParams) GetUpdatedByIdentityName() *string {
	return m.UpdatedByIdentityName
}

// SetUpdatedByIdentityName sets the UpdatedByIdentityName value and returns the updated SoftDeleteActivitiesByTypeParams.
func (m *SoftDeleteActivitiesByTypeParams) SetUpdatedByIdentityName(v *string) {
	m.UpdatedByIdentityName = v
}

// GetActivityType returns the ActivityType value.
func (m SoftDeleteActivitiesByTypeParams) GetActivityType() string {
	return m.ActivityType
}

// SetActivityType sets the ActivityType value and returns the updated SoftDeleteActivitiesByTypeParams.
func (m *SoftDeleteActivitiesByTypeParams) SetActivityType(v string) {
	m.ActivityType = v
}

// GetUpdatedByIdentityName returns the UpdatedByIdentityName value.
func (m SoftDeleteActivityParams) GetUpdatedByIdentityName() *string {
	return m.UpdatedByIdentityName
}

// SetUpdatedByIdentityName sets the UpdatedByIdentityName value and returns the updated SoftDeleteActivityParams.
func (m *SoftDeleteActivityParams) SetUpdatedByIdentityName(v *string) {
	m.UpdatedByIdentityName = v
}

// GetID returns the ID value.
func (m SoftDeleteActivityParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated SoftDeleteActivityParams.
func (m *SoftDeleteActivityParams) SetID(v string) {
	m.ID = v
}

// GetName returns the Name value.
func (m UpdateActivityParams) GetName() *string {
	return m.Name
}

// SetName sets the Name value and returns the updated UpdateActivityParams.
func (m *UpdateActivityParams) SetName(v *string) {
	m.Name = v
}

// GetActivityType returns the ActivityType value.
func (m UpdateActivityParams) GetActivityType() *string {
	return m.ActivityType
}

// SetActivityType sets the ActivityType value and returns the updated UpdateActivityParams.
func (m *UpdateActivityParams) SetActivityType(v *string) {
	m.ActivityType = v
}

// GetDurationMinutes returns the DurationMinutes value.
func (m UpdateActivityParams) GetDurationMinutes() *int64 {
	return m.DurationMinutes
}

// SetDurationMinutes sets the DurationMinutes value and returns the updated UpdateActivityParams.
func (m *UpdateActivityParams) SetDurationMinutes(v *int64) {
	m.DurationMinutes = v
}

// GetCalories returns the Calories value.
func (m UpdateActivityParams) GetCalories() *int64 {
	return m.Calories
}

// SetCalories sets the Calories value and returns the updated UpdateActivityParams.
func (m *UpdateActivityParams) SetCalories(v *int64) {
	m.Calories = v
}

// GetSteps returns the Steps value.
func (m UpdateActivityParams) GetSteps() *int64 {
	return m.Steps
}

// SetSteps sets the Steps value and returns the updated UpdateActivityParams.
func (m *UpdateActivityParams) SetSteps(v *int64) {
	m.Steps = v
}

// GetHeartRate returns the HeartRate value.
func (m UpdateActivityParams) GetHeartRate() *int64 {
	return m.HeartRate
}

// SetHeartRate sets the HeartRate value and returns the updated UpdateActivityParams.
func (m *UpdateActivityParams) SetHeartRate(v *int64) {
	m.HeartRate = v
}

// GetNotes returns the Notes value.
func (m UpdateActivityParams) GetNotes() *string {
	return m.Notes
}

// SetNotes sets the Notes value and returns the updated UpdateActivityParams.
func (m *UpdateActivityParams) SetNotes(v *string) {
	m.Notes = v
}

// GetReason returns the Reason value.
func (m UpdateActivityParams) GetReason() *string {
	return m.Reason
}

// SetReason sets the Reason value and returns the updated UpdateActivityParams.
func (m *UpdateActivityParams) SetReason(v *string) {
	m.Reason = v
}

// GetRawJsonData returns the RawJsonData value.
func (m UpdateActivityParams) GetRawJsonData() *string {
	return m.RawJsonData
}

// SetRawJsonData sets the RawJsonData value and returns the updated UpdateActivityParams.
func (m *UpdateActivityParams) SetRawJsonData(v *string) {
	m.RawJsonData = v
}

// GetUpdatedByIdentityID returns the UpdatedByIdentityID value.
func (m UpdateActivityParams) GetUpdatedByIdentityID() *int64 {
	return m.UpdatedByIdentityID
}

// SetUpdatedByIdentityID sets the UpdatedByIdentityID value and returns the updated UpdateActivityParams.
func (m *UpdateActivityParams) SetUpdatedByIdentityID(v *int64) {
	m.UpdatedByIdentityID = v
}

// GetUpdatedByIdentityName returns the UpdatedByIdentityName value.
func (m UpdateActivityParams) GetUpdatedByIdentityName() *string {
	return m.UpdatedByIdentityName
}

// SetUpdatedByIdentityName sets the UpdatedByIdentityName value and returns the updated UpdateActivityParams.
func (m *UpdateActivityParams) SetUpdatedByIdentityName(v *string) {
	m.UpdatedByIdentityName = v
}

// GetID returns the ID value.
func (m UpdateActivityParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated UpdateActivityParams.
func (m *UpdateActivityParams) SetID(v string) {
	m.ID = v
}

// GetIncludeDeleted returns the IncludeDeleted value.
func (m AllComponentEventsParams) GetIncludeDeleted() bool {
	return m.IncludeDeleted
}

// SetIncludeDeleted sets the IncludeDeleted value and returns the updated AllComponentEventsParams.
func (m *AllComponentEventsParams) SetIncludeDeleted(v bool) {
	m.IncludeDeleted = v
}

// GetSkip returns the Skip value.
func (m AllComponentEventsParams) GetSkip() int64 {
	return m.Skip
}

// SetSkip sets the Skip value and returns the updated AllComponentEventsParams.
func (m *AllComponentEventsParams) SetSkip(v int64) {
	m.Skip = v
}

// GetLimit returns the Limit value.
func (m AllComponentEventsParams) GetLimit() int64 {
	return m.Limit
}

// SetLimit sets the Limit value and returns the updated AllComponentEventsParams.
func (m *AllComponentEventsParams) SetLimit(v int64) {
	m.Limit = v
}

// GetID returns the ID value.
func (m CreateComponentEventParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated CreateComponentEventParams.
func (m *CreateComponentEventParams) SetID(v string) {
	m.ID = v
}

// GetEventType returns the EventType value.
func (m CreateComponentEventParams) GetEventType() string {
	return m.EventType
}

// SetEventType sets the EventType value and returns the updated CreateComponentEventParams.
func (m *CreateComponentEventParams) SetEventType(v string) {
	m.EventType = v
}

// GetSensorCode returns the SensorCode value.
func (m CreateComponentEventParams) GetSensorCode() *string {
	return m.SensorCode
}

// SetSensorCode sets the SensorCode value and returns the updated CreateComponentEventParams.
func (m *CreateComponentEventParams) SetSensorCode(v *string) {
	m.SensorCode = v
}

// GetTransmitterID returns the TransmitterID value.
func (m CreateComponentEventParams) GetTransmitterID() *string {
	return m.TransmitterID
}

// SetTransmitterID sets the TransmitterID value and returns the updated CreateComponentEventParams.
func (m *CreateComponentEventParams) SetTransmitterID(v *string) {
	m.TransmitterID = v
}

// GetPumpID returns the PumpID value.
func (m CreateComponentEventParams) GetPumpID() *string {
	return m.PumpID
}

// SetPumpID sets the PumpID value and returns the updated CreateComponentEventParams.
func (m *CreateComponentEventParams) SetPumpID(v *string) {
	m.PumpID = v
}

// GetPumpSerial returns the PumpSerial value.
func (m CreateComponentEventParams) GetPumpSerial() *string {
	return m.PumpSerial
}

// SetPumpSerial sets the PumpSerial value and returns the updated CreateComponentEventParams.
func (m *CreateComponentEventParams) SetPumpSerial(v *string) {
	m.PumpSerial = v
}

// GetPumpType returns the PumpType value.
func (m CreateComponentEventParams) GetPumpType() *string {
	return m.PumpType
}

// SetPumpType sets the PumpType value and returns the updated CreateComponentEventParams.
func (m *CreateComponentEventParams) SetPumpType(v *string) {
	m.PumpType = v
}

// GetNotes returns the Notes value.
func (m CreateComponentEventParams) GetNotes() *string {
	return m.Notes
}

// SetNotes sets the Notes value and returns the updated CreateComponentEventParams.
func (m *CreateComponentEventParams) SetNotes(v *string) {
	m.Notes = v
}

// GetReason returns the Reason value.
func (m CreateComponentEventParams) GetReason() *string {
	return m.Reason
}

// SetReason sets the Reason value and returns the updated CreateComponentEventParams.
func (m *CreateComponentEventParams) SetReason(v *string) {
	m.Reason = v
}

// GetUtcOffsetMinutes returns the UtcOffsetMinutes value.
func (m CreateComponentEventParams) GetUtcOffsetMinutes() *int64 {
	return m.UtcOffsetMinutes
}

// SetUtcOffsetMinutes sets the UtcOffsetMinutes value and returns the updated CreateComponentEventParams.
func (m *CreateComponentEventParams) SetUtcOffsetMinutes(v *int64) {
	m.UtcOffsetMinutes = v
}

// GetOccurredAt returns the OccurredAt value.
func (m CreateComponentEventParams) GetOccurredAt() int64 {
	return m.OccurredAt
}

// SetOccurredAt sets the OccurredAt value and returns the updated CreateComponentEventParams.
func (m *CreateComponentEventParams) SetOccurredAt(v int64) {
	m.OccurredAt = v
}

// GetRawJsonData returns the RawJsonData value.
func (m CreateComponentEventParams) GetRawJsonData() *string {
	return m.RawJsonData
}

// SetRawJsonData sets the RawJsonData value and returns the updated CreateComponentEventParams.
func (m *CreateComponentEventParams) SetRawJsonData(v *string) {
	m.RawJsonData = v
}

// GetCreatedByAppName returns the CreatedByAppName value.
func (m CreateComponentEventParams) GetCreatedByAppName() *string {
	return m.CreatedByAppName
}

// SetCreatedByAppName sets the CreatedByAppName value and returns the updated CreateComponentEventParams.
func (m *CreateComponentEventParams) SetCreatedByAppName(v *string) {
	m.CreatedByAppName = v
}

// GetCreatedByDeviceName returns the CreatedByDeviceName value.
func (m CreateComponentEventParams) GetCreatedByDeviceName() *string {
	return m.CreatedByDeviceName
}

// SetCreatedByDeviceName sets the CreatedByDeviceName value and returns the updated CreateComponentEventParams.
func (m *CreateComponentEventParams) SetCreatedByDeviceName(v *string) {
	m.CreatedByDeviceName = v
}

// GetCreatedByIdentityName returns the CreatedByIdentityName value.
func (m CreateComponentEventParams) GetCreatedByIdentityName() string {
	return m.CreatedByIdentityName
}

// SetCreatedByIdentityName sets the CreatedByIdentityName value and returns the updated CreateComponentEventParams.
func (m *CreateComponentEventParams) SetCreatedByIdentityName(v string) {
	m.CreatedByIdentityName = v
}

// GetCreatedByIdentityID returns the CreatedByIdentityID value.
func (m CreateComponentEventParams) GetCreatedByIdentityID() *int64 {
	return m.CreatedByIdentityID
}

// SetCreatedByIdentityID sets the CreatedByIdentityID value and returns the updated CreateComponentEventParams.
func (m *CreateComponentEventParams) SetCreatedByIdentityID(v *int64) {
	m.CreatedByIdentityID = v
}

// GetID returns the ID value.
func (m GetComponentEventParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated GetComponentEventParams.
func (m *GetComponentEventParams) SetID(v string) {
	m.ID = v
}

// GetIncludeDeleted returns the IncludeDeleted value.
func (m GetComponentEventParams) GetIncludeDeleted() bool {
	return m.IncludeDeleted
}

// SetIncludeDeleted sets the IncludeDeleted value and returns the updated GetComponentEventParams.
func (m *GetComponentEventParams) SetIncludeDeleted(v bool) {
	m.IncludeDeleted = v
}

// GetEventType returns the EventType value.
func (m GetComponentEventByTypeParams) GetEventType() string {
	return m.EventType
}

// SetEventType sets the EventType value and returns the updated GetComponentEventByTypeParams.
func (m *GetComponentEventByTypeParams) SetEventType(v string) {
	m.EventType = v
}

// GetID returns the ID value.
func (m HardDeleteComponentEventParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated HardDeleteComponentEventParams.
func (m *HardDeleteComponentEventParams) SetID(v string) {
	m.ID = v
}

// GetDateGTE returns the DateGTE value.
func (m SearchComponentEventsParams) GetDateGTE() *int64 {
	return m.DateGTE
}

// SetDateGTE sets the DateGTE value and returns the updated SearchComponentEventsParams.
func (m *SearchComponentEventsParams) SetDateGTE(v *int64) {
	m.DateGTE = v
}

// GetDateGT returns the DateGT value.
func (m SearchComponentEventsParams) GetDateGT() *int64 {
	return m.DateGT
}

// SetDateGT sets the DateGT value and returns the updated SearchComponentEventsParams.
func (m *SearchComponentEventsParams) SetDateGT(v *int64) {
	m.DateGT = v
}

// GetDateEQ returns the DateEQ value.
func (m SearchComponentEventsParams) GetDateEQ() *int64 {
	return m.DateEQ
}

// SetDateEQ sets the DateEQ value and returns the updated SearchComponentEventsParams.
func (m *SearchComponentEventsParams) SetDateEQ(v *int64) {
	m.DateEQ = v
}

// GetDateLTE returns the DateLTE value.
func (m SearchComponentEventsParams) GetDateLTE() *int64 {
	return m.DateLTE
}

// SetDateLTE sets the DateLTE value and returns the updated SearchComponentEventsParams.
func (m *SearchComponentEventsParams) SetDateLTE(v *int64) {
	m.DateLTE = v
}

// GetDateLT returns the DateLT value.
func (m SearchComponentEventsParams) GetDateLT() *int64 {
	return m.DateLT
}

// SetDateLT sets the DateLT value and returns the updated SearchComponentEventsParams.
func (m *SearchComponentEventsParams) SetDateLT(v *int64) {
	m.DateLT = v
}

// GetCreatedAtGTE returns the CreatedAtGTE value.
func (m SearchComponentEventsParams) GetCreatedAtGTE() *int64 {
	return m.CreatedAtGTE
}

// SetCreatedAtGTE sets the CreatedAtGTE value and returns the updated SearchComponentEventsParams.
func (m *SearchComponentEventsParams) SetCreatedAtGTE(v *int64) {
	m.CreatedAtGTE = v
}

// GetCreatedAtLTE returns the CreatedAtLTE value.
func (m SearchComponentEventsParams) GetCreatedAtLTE() *int64 {
	return m.CreatedAtLTE
}

// SetCreatedAtLTE sets the CreatedAtLTE value and returns the updated SearchComponentEventsParams.
func (m *SearchComponentEventsParams) SetCreatedAtLTE(v *int64) {
	m.CreatedAtLTE = v
}

// GetCreatedAtGT returns the CreatedAtGT value.
func (m SearchComponentEventsParams) GetCreatedAtGT() *int64 {
	return m.CreatedAtGT
}

// SetCreatedAtGT sets the CreatedAtGT value and returns the updated SearchComponentEventsParams.
func (m *SearchComponentEventsParams) SetCreatedAtGT(v *int64) {
	m.CreatedAtGT = v
}

// GetCreatedAtLT returns the CreatedAtLT value.
func (m SearchComponentEventsParams) GetCreatedAtLT() *int64 {
	return m.CreatedAtLT
}

// SetCreatedAtLT sets the CreatedAtLT value and returns the updated SearchComponentEventsParams.
func (m *SearchComponentEventsParams) SetCreatedAtLT(v *int64) {
	m.CreatedAtLT = v
}

// GetUpdatedGTE returns the UpdatedGTE value.
func (m SearchComponentEventsParams) GetUpdatedGTE() *int64 {
	return m.UpdatedGTE
}

// SetUpdatedGTE sets the UpdatedGTE value and returns the updated SearchComponentEventsParams.
func (m *SearchComponentEventsParams) SetUpdatedGTE(v *int64) {
	m.UpdatedGTE = v
}

// GetUpdatedLTE returns the UpdatedLTE value.
func (m SearchComponentEventsParams) GetUpdatedLTE() *int64 {
	return m.UpdatedLTE
}

// SetUpdatedLTE sets the UpdatedLTE value and returns the updated SearchComponentEventsParams.
func (m *SearchComponentEventsParams) SetUpdatedLTE(v *int64) {
	m.UpdatedLTE = v
}

// GetUpdatedGT returns the UpdatedGT value.
func (m SearchComponentEventsParams) GetUpdatedGT() *int64 {
	return m.UpdatedGT
}

// SetUpdatedGT sets the UpdatedGT value and returns the updated SearchComponentEventsParams.
func (m *SearchComponentEventsParams) SetUpdatedGT(v *int64) {
	m.UpdatedGT = v
}

// GetUpdatedLT returns the UpdatedLT value.
func (m SearchComponentEventsParams) GetUpdatedLT() *int64 {
	return m.UpdatedLT
}

// SetUpdatedLT sets the UpdatedLT value and returns the updated SearchComponentEventsParams.
func (m *SearchComponentEventsParams) SetUpdatedLT(v *int64) {
	m.UpdatedLT = v
}

// GetIDEq returns the IDEq value.
func (m SearchComponentEventsParams) GetIDEq() *string {
	return m.IDEq
}

// SetIDEq sets the IDEq value and returns the updated SearchComponentEventsParams.
func (m *SearchComponentEventsParams) SetIDEq(v *string) {
	m.IDEq = v
}

// GetIDIn returns the IDIn value.
func (m SearchComponentEventsParams) GetIDIn() *string {
	return m.IDIn
}

// SetIDIn sets the IDIn value and returns the updated SearchComponentEventsParams.
func (m *SearchComponentEventsParams) SetIDIn(v *string) {
	m.IDIn = v
}

// GetComponentEq returns the ComponentEq value.
func (m SearchComponentEventsParams) GetComponentEq() *string {
	return m.ComponentEq
}

// SetComponentEq sets the ComponentEq value and returns the updated SearchComponentEventsParams.
func (m *SearchComponentEventsParams) SetComponentEq(v *string) {
	m.ComponentEq = v
}

// GetAppEq returns the AppEq value.
func (m SearchComponentEventsParams) GetAppEq() *string {
	return m.AppEq
}

// SetAppEq sets the AppEq value and returns the updated SearchComponentEventsParams.
func (m *SearchComponentEventsParams) SetAppEq(v *string) {
	m.AppEq = v
}

// GetTypeEq returns the TypeEq value.
func (m SearchComponentEventsParams) GetTypeEq() *string {
	return m.TypeEq
}

// SetTypeEq sets the TypeEq value and returns the updated SearchComponentEventsParams.
func (m *SearchComponentEventsParams) SetTypeEq(v *string) {
	m.TypeEq = v
}

// GetTypeNeq returns the TypeNeq value.
func (m SearchComponentEventsParams) GetTypeNeq() *string {
	return m.TypeNeq
}

// SetTypeNeq sets the TypeNeq value and returns the updated SearchComponentEventsParams.
func (m *SearchComponentEventsParams) SetTypeNeq(v *string) {
	m.TypeNeq = v
}

// GetTypeIn returns the TypeIn value.
func (m SearchComponentEventsParams) GetTypeIn() *string {
	return m.TypeIn
}

// SetTypeIn sets the TypeIn value and returns the updated SearchComponentEventsParams.
func (m *SearchComponentEventsParams) SetTypeIn(v *string) {
	m.TypeIn = v
}

// GetTypeNin returns the TypeNin value.
func (m SearchComponentEventsParams) GetTypeNin() *string {
	return m.TypeNin
}

// SetTypeNin sets the TypeNin value and returns the updated SearchComponentEventsParams.
func (m *SearchComponentEventsParams) SetTypeNin(v *string) {
	m.TypeNin = v
}

// GetIncludeDeleted returns the IncludeDeleted value.
func (m SearchComponentEventsParams) GetIncludeDeleted() *bool {
	return m.IncludeDeleted
}

// SetIncludeDeleted sets the IncludeDeleted value and returns the updated SearchComponentEventsParams.
func (m *SearchComponentEventsParams) SetIncludeDeleted(v *bool) {
	m.IncludeDeleted = v
}

// GetSkip returns the Skip value.
func (m SearchComponentEventsParams) GetSkip() int64 {
	return m.Skip
}

// SetSkip sets the Skip value and returns the updated SearchComponentEventsParams.
func (m *SearchComponentEventsParams) SetSkip(v int64) {
	m.Skip = v
}

// GetLimit returns the Limit value.
func (m SearchComponentEventsParams) GetLimit() int64 {
	return m.Limit
}

// SetLimit sets the Limit value and returns the updated SearchComponentEventsParams.
func (m *SearchComponentEventsParams) SetLimit(v int64) {
	m.Limit = v
}

// GetUpdatedByIdentityName returns the UpdatedByIdentityName value.
func (m SoftDeleteComponentEventParams) GetUpdatedByIdentityName() *string {
	return m.UpdatedByIdentityName
}

// SetUpdatedByIdentityName sets the UpdatedByIdentityName value and returns the updated SoftDeleteComponentEventParams.
func (m *SoftDeleteComponentEventParams) SetUpdatedByIdentityName(v *string) {
	m.UpdatedByIdentityName = v
}

// GetID returns the ID value.
func (m SoftDeleteComponentEventParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated SoftDeleteComponentEventParams.
func (m *SoftDeleteComponentEventParams) SetID(v string) {
	m.ID = v
}

// GetEventType returns the EventType value.
func (m UpdateComponentEventParams) GetEventType() *string {
	return m.EventType
}

// SetEventType sets the EventType value and returns the updated UpdateComponentEventParams.
func (m *UpdateComponentEventParams) SetEventType(v *string) {
	m.EventType = v
}

// GetSensorCode returns the SensorCode value.
func (m UpdateComponentEventParams) GetSensorCode() *string {
	return m.SensorCode
}

// SetSensorCode sets the SensorCode value and returns the updated UpdateComponentEventParams.
func (m *UpdateComponentEventParams) SetSensorCode(v *string) {
	m.SensorCode = v
}

// GetTransmitterID returns the TransmitterID value.
func (m UpdateComponentEventParams) GetTransmitterID() *string {
	return m.TransmitterID
}

// SetTransmitterID sets the TransmitterID value and returns the updated UpdateComponentEventParams.
func (m *UpdateComponentEventParams) SetTransmitterID(v *string) {
	m.TransmitterID = v
}

// GetNotes returns the Notes value.
func (m UpdateComponentEventParams) GetNotes() *string {
	return m.Notes
}

// SetNotes sets the Notes value and returns the updated UpdateComponentEventParams.
func (m *UpdateComponentEventParams) SetNotes(v *string) {
	m.Notes = v
}

// GetReason returns the Reason value.
func (m UpdateComponentEventParams) GetReason() *string {
	return m.Reason
}

// SetReason sets the Reason value and returns the updated UpdateComponentEventParams.
func (m *UpdateComponentEventParams) SetReason(v *string) {
	m.Reason = v
}

// GetRawJsonData returns the RawJsonData value.
func (m UpdateComponentEventParams) GetRawJsonData() *string {
	return m.RawJsonData
}

// SetRawJsonData sets the RawJsonData value and returns the updated UpdateComponentEventParams.
func (m *UpdateComponentEventParams) SetRawJsonData(v *string) {
	m.RawJsonData = v
}

// GetUpdatedByIdentityID returns the UpdatedByIdentityID value.
func (m UpdateComponentEventParams) GetUpdatedByIdentityID() *int64 {
	return m.UpdatedByIdentityID
}

// SetUpdatedByIdentityID sets the UpdatedByIdentityID value and returns the updated UpdateComponentEventParams.
func (m *UpdateComponentEventParams) SetUpdatedByIdentityID(v *int64) {
	m.UpdatedByIdentityID = v
}

// GetUpdatedByIdentityName returns the UpdatedByIdentityName value.
func (m UpdateComponentEventParams) GetUpdatedByIdentityName() *string {
	return m.UpdatedByIdentityName
}

// SetUpdatedByIdentityName sets the UpdatedByIdentityName value and returns the updated UpdateComponentEventParams.
func (m *UpdateComponentEventParams) SetUpdatedByIdentityName(v *string) {
	m.UpdatedByIdentityName = v
}

// GetID returns the ID value.
func (m UpdateComponentEventParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated UpdateComponentEventParams.
func (m *UpdateComponentEventParams) SetID(v string) {
	m.ID = v
}

// GetIncludeDeleted returns the IncludeDeleted value.
func (m AllFoodIntakesParams) GetIncludeDeleted() bool {
	return m.IncludeDeleted
}

// SetIncludeDeleted sets the IncludeDeleted value and returns the updated AllFoodIntakesParams.
func (m *AllFoodIntakesParams) SetIncludeDeleted(v bool) {
	m.IncludeDeleted = v
}

// GetSkip returns the Skip value.
func (m AllFoodIntakesParams) GetSkip() int64 {
	return m.Skip
}

// SetSkip sets the Skip value and returns the updated AllFoodIntakesParams.
func (m *AllFoodIntakesParams) SetSkip(v int64) {
	m.Skip = v
}

// GetLimit returns the Limit value.
func (m AllFoodIntakesParams) GetLimit() int64 {
	return m.Limit
}

// SetLimit sets the Limit value and returns the updated AllFoodIntakesParams.
func (m *AllFoodIntakesParams) SetLimit(v int64) {
	m.Limit = v
}

// GetID returns the ID value.
func (m CreateFoodIntakeParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated CreateFoodIntakeParams.
func (m *CreateFoodIntakeParams) SetID(v string) {
	m.ID = v
}

// GetName returns the Name value.
func (m CreateFoodIntakeParams) GetName() *string {
	return m.Name
}

// SetName sets the Name value and returns the updated CreateFoodIntakeParams.
func (m *CreateFoodIntakeParams) SetName(v *string) {
	m.Name = v
}

// GetFoodType returns the FoodType value.
func (m CreateFoodIntakeParams) GetFoodType() *string {
	return m.FoodType
}

// SetFoodType sets the FoodType value and returns the updated CreateFoodIntakeParams.
func (m *CreateFoodIntakeParams) SetFoodType(v *string) {
	m.FoodType = v
}

// GetCategory returns the Category value.
func (m CreateFoodIntakeParams) GetCategory() *string {
	return m.Category
}

// SetCategory sets the Category value and returns the updated CreateFoodIntakeParams.
func (m *CreateFoodIntakeParams) SetCategory(v *string) {
	m.Category = v
}

// GetSubcategory returns the Subcategory value.
func (m CreateFoodIntakeParams) GetSubcategory() *string {
	return m.Subcategory
}

// SetSubcategory sets the Subcategory value and returns the updated CreateFoodIntakeParams.
func (m *CreateFoodIntakeParams) SetSubcategory(v *string) {
	m.Subcategory = v
}

// GetAbsorptionTimeMinutes returns the AbsorptionTimeMinutes value.
func (m CreateFoodIntakeParams) GetAbsorptionTimeMinutes() *int64 {
	return m.AbsorptionTimeMinutes
}

// SetAbsorptionTimeMinutes sets the AbsorptionTimeMinutes value and returns the updated CreateFoodIntakeParams.
func (m *CreateFoodIntakeParams) SetAbsorptionTimeMinutes(v *int64) {
	m.AbsorptionTimeMinutes = v
}

// GetPreBolus returns the PreBolus value.
func (m CreateFoodIntakeParams) GetPreBolus() *float64 {
	return m.PreBolus
}

// SetPreBolus sets the PreBolus value and returns the updated CreateFoodIntakeParams.
func (m *CreateFoodIntakeParams) SetPreBolus(v *float64) {
	m.PreBolus = v
}

// GetCarbs returns the Carbs value.
func (m CreateFoodIntakeParams) GetCarbs() *float64 {
	return m.Carbs
}

// SetCarbs sets the Carbs value and returns the updated CreateFoodIntakeParams.
func (m *CreateFoodIntakeParams) SetCarbs(v *float64) {
	m.Carbs = v
}

// GetFat returns the Fat value.
func (m CreateFoodIntakeParams) GetFat() *float64 {
	return m.Fat
}

// SetFat sets the Fat value and returns the updated CreateFoodIntakeParams.
func (m *CreateFoodIntakeParams) SetFat(v *float64) {
	m.Fat = v
}

// GetProtein returns the Protein value.
func (m CreateFoodIntakeParams) GetProtein() *float64 {
	return m.Protein
}

// SetProtein sets the Protein value and returns the updated CreateFoodIntakeParams.
func (m *CreateFoodIntakeParams) SetProtein(v *float64) {
	m.Protein = v
}

// GetFibre returns the Fibre value.
func (m CreateFoodIntakeParams) GetFibre() *float64 {
	return m.Fibre
}

// SetFibre sets the Fibre value and returns the updated CreateFoodIntakeParams.
func (m *CreateFoodIntakeParams) SetFibre(v *float64) {
	m.Fibre = v
}

// GetEnergy returns the Energy value.
func (m CreateFoodIntakeParams) GetEnergy() *float64 {
	return m.Energy
}

// SetEnergy sets the Energy value and returns the updated CreateFoodIntakeParams.
func (m *CreateFoodIntakeParams) SetEnergy(v *float64) {
	m.Energy = v
}

// GetPortions returns the Portions value.
func (m CreateFoodIntakeParams) GetPortions() *float64 {
	return m.Portions
}

// SetPortions sets the Portions value and returns the updated CreateFoodIntakeParams.
func (m *CreateFoodIntakeParams) SetPortions(v *float64) {
	m.Portions = v
}

// GetNotes returns the Notes value.
func (m CreateFoodIntakeParams) GetNotes() *string {
	return m.Notes
}

// SetNotes sets the Notes value and returns the updated CreateFoodIntakeParams.
func (m *CreateFoodIntakeParams) SetNotes(v *string) {
	m.Notes = v
}

// GetReason returns the Reason value.
func (m CreateFoodIntakeParams) GetReason() *string {
	return m.Reason
}

// SetReason sets the Reason value and returns the updated CreateFoodIntakeParams.
func (m *CreateFoodIntakeParams) SetReason(v *string) {
	m.Reason = v
}

// GetUtcOffsetMinutes returns the UtcOffsetMinutes value.
func (m CreateFoodIntakeParams) GetUtcOffsetMinutes() *int64 {
	return m.UtcOffsetMinutes
}

// SetUtcOffsetMinutes sets the UtcOffsetMinutes value and returns the updated CreateFoodIntakeParams.
func (m *CreateFoodIntakeParams) SetUtcOffsetMinutes(v *int64) {
	m.UtcOffsetMinutes = v
}

// GetOccurredAt returns the OccurredAt value.
func (m CreateFoodIntakeParams) GetOccurredAt() int64 {
	return m.OccurredAt
}

// SetOccurredAt sets the OccurredAt value and returns the updated CreateFoodIntakeParams.
func (m *CreateFoodIntakeParams) SetOccurredAt(v int64) {
	m.OccurredAt = v
}

// GetRawJsonData returns the RawJsonData value.
func (m CreateFoodIntakeParams) GetRawJsonData() *string {
	return m.RawJsonData
}

// SetRawJsonData sets the RawJsonData value and returns the updated CreateFoodIntakeParams.
func (m *CreateFoodIntakeParams) SetRawJsonData(v *string) {
	m.RawJsonData = v
}

// GetCreatedByAppName returns the CreatedByAppName value.
func (m CreateFoodIntakeParams) GetCreatedByAppName() *string {
	return m.CreatedByAppName
}

// SetCreatedByAppName sets the CreatedByAppName value and returns the updated CreateFoodIntakeParams.
func (m *CreateFoodIntakeParams) SetCreatedByAppName(v *string) {
	m.CreatedByAppName = v
}

// GetCreatedByDeviceName returns the CreatedByDeviceName value.
func (m CreateFoodIntakeParams) GetCreatedByDeviceName() *string {
	return m.CreatedByDeviceName
}

// SetCreatedByDeviceName sets the CreatedByDeviceName value and returns the updated CreateFoodIntakeParams.
func (m *CreateFoodIntakeParams) SetCreatedByDeviceName(v *string) {
	m.CreatedByDeviceName = v
}

// GetCreatedByIdentityName returns the CreatedByIdentityName value.
func (m CreateFoodIntakeParams) GetCreatedByIdentityName() string {
	return m.CreatedByIdentityName
}

// SetCreatedByIdentityName sets the CreatedByIdentityName value and returns the updated CreateFoodIntakeParams.
func (m *CreateFoodIntakeParams) SetCreatedByIdentityName(v string) {
	m.CreatedByIdentityName = v
}

// GetCreatedByIdentityID returns the CreatedByIdentityID value.
func (m CreateFoodIntakeParams) GetCreatedByIdentityID() *int64 {
	return m.CreatedByIdentityID
}

// SetCreatedByIdentityID sets the CreatedByIdentityID value and returns the updated CreateFoodIntakeParams.
func (m *CreateFoodIntakeParams) SetCreatedByIdentityID(v *int64) {
	m.CreatedByIdentityID = v
}

// GetID returns the ID value.
func (m GetFoodIntakeParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated GetFoodIntakeParams.
func (m *GetFoodIntakeParams) SetID(v string) {
	m.ID = v
}

// GetIncludeDeleted returns the IncludeDeleted value.
func (m GetFoodIntakeParams) GetIncludeDeleted() bool {
	return m.IncludeDeleted
}

// SetIncludeDeleted sets the IncludeDeleted value and returns the updated GetFoodIntakeParams.
func (m *GetFoodIntakeParams) SetIncludeDeleted(v bool) {
	m.IncludeDeleted = v
}

// GetID returns the ID value.
func (m HardDeleteFoodIntakeParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated HardDeleteFoodIntakeParams.
func (m *HardDeleteFoodIntakeParams) SetID(v string) {
	m.ID = v
}

// GetFoodType returns the FoodType value.
func (m HardDeleteFoodIntakesByTypeParams) GetFoodType() *string {
	return m.FoodType
}

// SetFoodType sets the FoodType value and returns the updated HardDeleteFoodIntakesByTypeParams.
func (m *HardDeleteFoodIntakesByTypeParams) SetFoodType(v *string) {
	m.FoodType = v
}

// GetDateGTE returns the DateGTE value.
func (m SearchFoodIntakesParams) GetDateGTE() *int64 {
	return m.DateGTE
}

// SetDateGTE sets the DateGTE value and returns the updated SearchFoodIntakesParams.
func (m *SearchFoodIntakesParams) SetDateGTE(v *int64) {
	m.DateGTE = v
}

// GetDateGT returns the DateGT value.
func (m SearchFoodIntakesParams) GetDateGT() *int64 {
	return m.DateGT
}

// SetDateGT sets the DateGT value and returns the updated SearchFoodIntakesParams.
func (m *SearchFoodIntakesParams) SetDateGT(v *int64) {
	m.DateGT = v
}

// GetDateEQ returns the DateEQ value.
func (m SearchFoodIntakesParams) GetDateEQ() *int64 {
	return m.DateEQ
}

// SetDateEQ sets the DateEQ value and returns the updated SearchFoodIntakesParams.
func (m *SearchFoodIntakesParams) SetDateEQ(v *int64) {
	m.DateEQ = v
}

// GetDateLTE returns the DateLTE value.
func (m SearchFoodIntakesParams) GetDateLTE() *int64 {
	return m.DateLTE
}

// SetDateLTE sets the DateLTE value and returns the updated SearchFoodIntakesParams.
func (m *SearchFoodIntakesParams) SetDateLTE(v *int64) {
	m.DateLTE = v
}

// GetDateLT returns the DateLT value.
func (m SearchFoodIntakesParams) GetDateLT() *int64 {
	return m.DateLT
}

// SetDateLT sets the DateLT value and returns the updated SearchFoodIntakesParams.
func (m *SearchFoodIntakesParams) SetDateLT(v *int64) {
	m.DateLT = v
}

// GetCreatedAtGTE returns the CreatedAtGTE value.
func (m SearchFoodIntakesParams) GetCreatedAtGTE() *int64 {
	return m.CreatedAtGTE
}

// SetCreatedAtGTE sets the CreatedAtGTE value and returns the updated SearchFoodIntakesParams.
func (m *SearchFoodIntakesParams) SetCreatedAtGTE(v *int64) {
	m.CreatedAtGTE = v
}

// GetCreatedAtLTE returns the CreatedAtLTE value.
func (m SearchFoodIntakesParams) GetCreatedAtLTE() *int64 {
	return m.CreatedAtLTE
}

// SetCreatedAtLTE sets the CreatedAtLTE value and returns the updated SearchFoodIntakesParams.
func (m *SearchFoodIntakesParams) SetCreatedAtLTE(v *int64) {
	m.CreatedAtLTE = v
}

// GetCreatedAtGT returns the CreatedAtGT value.
func (m SearchFoodIntakesParams) GetCreatedAtGT() *int64 {
	return m.CreatedAtGT
}

// SetCreatedAtGT sets the CreatedAtGT value and returns the updated SearchFoodIntakesParams.
func (m *SearchFoodIntakesParams) SetCreatedAtGT(v *int64) {
	m.CreatedAtGT = v
}

// GetCreatedAtLT returns the CreatedAtLT value.
func (m SearchFoodIntakesParams) GetCreatedAtLT() *int64 {
	return m.CreatedAtLT
}

// SetCreatedAtLT sets the CreatedAtLT value and returns the updated SearchFoodIntakesParams.
func (m *SearchFoodIntakesParams) SetCreatedAtLT(v *int64) {
	m.CreatedAtLT = v
}

// GetUpdatedGTE returns the UpdatedGTE value.
func (m SearchFoodIntakesParams) GetUpdatedGTE() *int64 {
	return m.UpdatedGTE
}

// SetUpdatedGTE sets the UpdatedGTE value and returns the updated SearchFoodIntakesParams.
func (m *SearchFoodIntakesParams) SetUpdatedGTE(v *int64) {
	m.UpdatedGTE = v
}

// GetUpdatedLTE returns the UpdatedLTE value.
func (m SearchFoodIntakesParams) GetUpdatedLTE() *int64 {
	return m.UpdatedLTE
}

// SetUpdatedLTE sets the UpdatedLTE value and returns the updated SearchFoodIntakesParams.
func (m *SearchFoodIntakesParams) SetUpdatedLTE(v *int64) {
	m.UpdatedLTE = v
}

// GetUpdatedGT returns the UpdatedGT value.
func (m SearchFoodIntakesParams) GetUpdatedGT() *int64 {
	return m.UpdatedGT
}

// SetUpdatedGT sets the UpdatedGT value and returns the updated SearchFoodIntakesParams.
func (m *SearchFoodIntakesParams) SetUpdatedGT(v *int64) {
	m.UpdatedGT = v
}

// GetUpdatedLT returns the UpdatedLT value.
func (m SearchFoodIntakesParams) GetUpdatedLT() *int64 {
	return m.UpdatedLT
}

// SetUpdatedLT sets the UpdatedLT value and returns the updated SearchFoodIntakesParams.
func (m *SearchFoodIntakesParams) SetUpdatedLT(v *int64) {
	m.UpdatedLT = v
}

// GetIDEq returns the IDEq value.
func (m SearchFoodIntakesParams) GetIDEq() *string {
	return m.IDEq
}

// SetIDEq sets the IDEq value and returns the updated SearchFoodIntakesParams.
func (m *SearchFoodIntakesParams) SetIDEq(v *string) {
	m.IDEq = v
}

// GetIDIn returns the IDIn value.
func (m SearchFoodIntakesParams) GetIDIn() *string {
	return m.IDIn
}

// SetIDIn sets the IDIn value and returns the updated SearchFoodIntakesParams.
func (m *SearchFoodIntakesParams) SetIDIn(v *string) {
	m.IDIn = v
}

// GetNameEq returns the NameEq value.
func (m SearchFoodIntakesParams) GetNameEq() *string {
	return m.NameEq
}

// SetNameEq sets the NameEq value and returns the updated SearchFoodIntakesParams.
func (m *SearchFoodIntakesParams) SetNameEq(v *string) {
	m.NameEq = v
}

// GetDeviceEq returns the DeviceEq value.
func (m SearchFoodIntakesParams) GetDeviceEq() *string {
	return m.DeviceEq
}

// SetDeviceEq sets the DeviceEq value and returns the updated SearchFoodIntakesParams.
func (m *SearchFoodIntakesParams) SetDeviceEq(v *string) {
	m.DeviceEq = v
}

// GetAppEq returns the AppEq value.
func (m SearchFoodIntakesParams) GetAppEq() *string {
	return m.AppEq
}

// SetAppEq sets the AppEq value and returns the updated SearchFoodIntakesParams.
func (m *SearchFoodIntakesParams) SetAppEq(v *string) {
	m.AppEq = v
}

// GetTypeEq returns the TypeEq value.
func (m SearchFoodIntakesParams) GetTypeEq() *string {
	return m.TypeEq
}

// SetTypeEq sets the TypeEq value and returns the updated SearchFoodIntakesParams.
func (m *SearchFoodIntakesParams) SetTypeEq(v *string) {
	m.TypeEq = v
}

// GetTypeNeq returns the TypeNeq value.
func (m SearchFoodIntakesParams) GetTypeNeq() *string {
	return m.TypeNeq
}

// SetTypeNeq sets the TypeNeq value and returns the updated SearchFoodIntakesParams.
func (m *SearchFoodIntakesParams) SetTypeNeq(v *string) {
	m.TypeNeq = v
}

// GetTypeIn returns the TypeIn value.
func (m SearchFoodIntakesParams) GetTypeIn() *string {
	return m.TypeIn
}

// SetTypeIn sets the TypeIn value and returns the updated SearchFoodIntakesParams.
func (m *SearchFoodIntakesParams) SetTypeIn(v *string) {
	m.TypeIn = v
}

// GetTypeNin returns the TypeNin value.
func (m SearchFoodIntakesParams) GetTypeNin() *string {
	return m.TypeNin
}

// SetTypeNin sets the TypeNin value and returns the updated SearchFoodIntakesParams.
func (m *SearchFoodIntakesParams) SetTypeNin(v *string) {
	m.TypeNin = v
}

// GetCarbsEQ returns the CarbsEQ value.
func (m SearchFoodIntakesParams) GetCarbsEQ() *int64 {
	return m.CarbsEQ
}

// SetCarbsEQ sets the CarbsEQ value and returns the updated SearchFoodIntakesParams.
func (m *SearchFoodIntakesParams) SetCarbsEQ(v *int64) {
	m.CarbsEQ = v
}

// GetCarbsGTE returns the CarbsGTE value.
func (m SearchFoodIntakesParams) GetCarbsGTE() *int64 {
	return m.CarbsGTE
}

// SetCarbsGTE sets the CarbsGTE value and returns the updated SearchFoodIntakesParams.
func (m *SearchFoodIntakesParams) SetCarbsGTE(v *int64) {
	m.CarbsGTE = v
}

// GetCarbsGT returns the CarbsGT value.
func (m SearchFoodIntakesParams) GetCarbsGT() *int64 {
	return m.CarbsGT
}

// SetCarbsGT sets the CarbsGT value and returns the updated SearchFoodIntakesParams.
func (m *SearchFoodIntakesParams) SetCarbsGT(v *int64) {
	m.CarbsGT = v
}

// GetCarbsLTE returns the CarbsLTE value.
func (m SearchFoodIntakesParams) GetCarbsLTE() *int64 {
	return m.CarbsLTE
}

// SetCarbsLTE sets the CarbsLTE value and returns the updated SearchFoodIntakesParams.
func (m *SearchFoodIntakesParams) SetCarbsLTE(v *int64) {
	m.CarbsLTE = v
}

// GetCarbsLT returns the CarbsLT value.
func (m SearchFoodIntakesParams) GetCarbsLT() *int64 {
	return m.CarbsLT
}

// SetCarbsLT sets the CarbsLT value and returns the updated SearchFoodIntakesParams.
func (m *SearchFoodIntakesParams) SetCarbsLT(v *int64) {
	m.CarbsLT = v
}

// GetIncludeDeleted returns the IncludeDeleted value.
func (m SearchFoodIntakesParams) GetIncludeDeleted() *bool {
	return m.IncludeDeleted
}

// SetIncludeDeleted sets the IncludeDeleted value and returns the updated SearchFoodIntakesParams.
func (m *SearchFoodIntakesParams) SetIncludeDeleted(v *bool) {
	m.IncludeDeleted = v
}

// GetSkip returns the Skip value.
func (m SearchFoodIntakesParams) GetSkip() int64 {
	return m.Skip
}

// SetSkip sets the Skip value and returns the updated SearchFoodIntakesParams.
func (m *SearchFoodIntakesParams) SetSkip(v int64) {
	m.Skip = v
}

// GetLimit returns the Limit value.
func (m SearchFoodIntakesParams) GetLimit() int64 {
	return m.Limit
}

// SetLimit sets the Limit value and returns the updated SearchFoodIntakesParams.
func (m *SearchFoodIntakesParams) SetLimit(v int64) {
	m.Limit = v
}

// GetUpdatedByIdentityName returns the UpdatedByIdentityName value.
func (m SoftDeleteFoodIntakeParams) GetUpdatedByIdentityName() *string {
	return m.UpdatedByIdentityName
}

// SetUpdatedByIdentityName sets the UpdatedByIdentityName value and returns the updated SoftDeleteFoodIntakeParams.
func (m *SoftDeleteFoodIntakeParams) SetUpdatedByIdentityName(v *string) {
	m.UpdatedByIdentityName = v
}

// GetID returns the ID value.
func (m SoftDeleteFoodIntakeParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated SoftDeleteFoodIntakeParams.
func (m *SoftDeleteFoodIntakeParams) SetID(v string) {
	m.ID = v
}

// GetUpdatedByIdentityName returns the UpdatedByIdentityName value.
func (m SoftDeleteFoodIntakesByTypeParams) GetUpdatedByIdentityName() *string {
	return m.UpdatedByIdentityName
}

// SetUpdatedByIdentityName sets the UpdatedByIdentityName value and returns the updated SoftDeleteFoodIntakesByTypeParams.
func (m *SoftDeleteFoodIntakesByTypeParams) SetUpdatedByIdentityName(v *string) {
	m.UpdatedByIdentityName = v
}

// GetFoodType returns the FoodType value.
func (m SoftDeleteFoodIntakesByTypeParams) GetFoodType() *string {
	return m.FoodType
}

// SetFoodType sets the FoodType value and returns the updated SoftDeleteFoodIntakesByTypeParams.
func (m *SoftDeleteFoodIntakesByTypeParams) SetFoodType(v *string) {
	m.FoodType = v
}

// GetName returns the Name value.
func (m UpdateFoodIntakeParams) GetName() *string {
	return m.Name
}

// SetName sets the Name value and returns the updated UpdateFoodIntakeParams.
func (m *UpdateFoodIntakeParams) SetName(v *string) {
	m.Name = v
}

// GetFoodType returns the FoodType value.
func (m UpdateFoodIntakeParams) GetFoodType() *string {
	return m.FoodType
}

// SetFoodType sets the FoodType value and returns the updated UpdateFoodIntakeParams.
func (m *UpdateFoodIntakeParams) SetFoodType(v *string) {
	m.FoodType = v
}

// GetCategory returns the Category value.
func (m UpdateFoodIntakeParams) GetCategory() *string {
	return m.Category
}

// SetCategory sets the Category value and returns the updated UpdateFoodIntakeParams.
func (m *UpdateFoodIntakeParams) SetCategory(v *string) {
	m.Category = v
}

// GetSubcategory returns the Subcategory value.
func (m UpdateFoodIntakeParams) GetSubcategory() *string {
	return m.Subcategory
}

// SetSubcategory sets the Subcategory value and returns the updated UpdateFoodIntakeParams.
func (m *UpdateFoodIntakeParams) SetSubcategory(v *string) {
	m.Subcategory = v
}

// GetCarbs returns the Carbs value.
func (m UpdateFoodIntakeParams) GetCarbs() *float64 {
	return m.Carbs
}

// SetCarbs sets the Carbs value and returns the updated UpdateFoodIntakeParams.
func (m *UpdateFoodIntakeParams) SetCarbs(v *float64) {
	m.Carbs = v
}

// GetFat returns the Fat value.
func (m UpdateFoodIntakeParams) GetFat() *float64 {
	return m.Fat
}

// SetFat sets the Fat value and returns the updated UpdateFoodIntakeParams.
func (m *UpdateFoodIntakeParams) SetFat(v *float64) {
	m.Fat = v
}

// GetProtein returns the Protein value.
func (m UpdateFoodIntakeParams) GetProtein() *float64 {
	return m.Protein
}

// SetProtein sets the Protein value and returns the updated UpdateFoodIntakeParams.
func (m *UpdateFoodIntakeParams) SetProtein(v *float64) {
	m.Protein = v
}

// GetFibre returns the Fibre value.
func (m UpdateFoodIntakeParams) GetFibre() *float64 {
	return m.Fibre
}

// SetFibre sets the Fibre value and returns the updated UpdateFoodIntakeParams.
func (m *UpdateFoodIntakeParams) SetFibre(v *float64) {
	m.Fibre = v
}

// GetEnergy returns the Energy value.
func (m UpdateFoodIntakeParams) GetEnergy() *float64 {
	return m.Energy
}

// SetEnergy sets the Energy value and returns the updated UpdateFoodIntakeParams.
func (m *UpdateFoodIntakeParams) SetEnergy(v *float64) {
	m.Energy = v
}

// GetAbsorptionTimeMinutes returns the AbsorptionTimeMinutes value.
func (m UpdateFoodIntakeParams) GetAbsorptionTimeMinutes() *int64 {
	return m.AbsorptionTimeMinutes
}

// SetAbsorptionTimeMinutes sets the AbsorptionTimeMinutes value and returns the updated UpdateFoodIntakeParams.
func (m *UpdateFoodIntakeParams) SetAbsorptionTimeMinutes(v *int64) {
	m.AbsorptionTimeMinutes = v
}

// GetPreBolus returns the PreBolus value.
func (m UpdateFoodIntakeParams) GetPreBolus() *float64 {
	return m.PreBolus
}

// SetPreBolus sets the PreBolus value and returns the updated UpdateFoodIntakeParams.
func (m *UpdateFoodIntakeParams) SetPreBolus(v *float64) {
	m.PreBolus = v
}

// GetPortions returns the Portions value.
func (m UpdateFoodIntakeParams) GetPortions() *float64 {
	return m.Portions
}

// SetPortions sets the Portions value and returns the updated UpdateFoodIntakeParams.
func (m *UpdateFoodIntakeParams) SetPortions(v *float64) {
	m.Portions = v
}

// GetNotes returns the Notes value.
func (m UpdateFoodIntakeParams) GetNotes() *string {
	return m.Notes
}

// SetNotes sets the Notes value and returns the updated UpdateFoodIntakeParams.
func (m *UpdateFoodIntakeParams) SetNotes(v *string) {
	m.Notes = v
}

// GetReason returns the Reason value.
func (m UpdateFoodIntakeParams) GetReason() *string {
	return m.Reason
}

// SetReason sets the Reason value and returns the updated UpdateFoodIntakeParams.
func (m *UpdateFoodIntakeParams) SetReason(v *string) {
	m.Reason = v
}

// GetRawJsonData returns the RawJsonData value.
func (m UpdateFoodIntakeParams) GetRawJsonData() *string {
	return m.RawJsonData
}

// SetRawJsonData sets the RawJsonData value and returns the updated UpdateFoodIntakeParams.
func (m *UpdateFoodIntakeParams) SetRawJsonData(v *string) {
	m.RawJsonData = v
}

// GetUpdatedByIdentityID returns the UpdatedByIdentityID value.
func (m UpdateFoodIntakeParams) GetUpdatedByIdentityID() *int64 {
	return m.UpdatedByIdentityID
}

// SetUpdatedByIdentityID sets the UpdatedByIdentityID value and returns the updated UpdateFoodIntakeParams.
func (m *UpdateFoodIntakeParams) SetUpdatedByIdentityID(v *int64) {
	m.UpdatedByIdentityID = v
}

// GetUpdatedByIdentityName returns the UpdatedByIdentityName value.
func (m UpdateFoodIntakeParams) GetUpdatedByIdentityName() *string {
	return m.UpdatedByIdentityName
}

// SetUpdatedByIdentityName sets the UpdatedByIdentityName value and returns the updated UpdateFoodIntakeParams.
func (m *UpdateFoodIntakeParams) SetUpdatedByIdentityName(v *string) {
	m.UpdatedByIdentityName = v
}

// GetID returns the ID value.
func (m UpdateFoodIntakeParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated UpdateFoodIntakeParams.
func (m *UpdateFoodIntakeParams) SetID(v string) {
	m.ID = v
}

// GetMinutes returns the Minutes value.
func (m AlarmGlucoseReadingsParams) GetMinutes() int64 {
	return m.Minutes
}

// SetMinutes sets the Minutes value and returns the updated AlarmGlucoseReadingsParams.
func (m *AlarmGlucoseReadingsParams) SetMinutes(v int64) {
	m.Minutes = v
}

// GetLimit returns the Limit value.
func (m AlarmGlucoseReadingsParams) GetLimit() int64 {
	return m.Limit
}

// SetLimit sets the Limit value and returns the updated AlarmGlucoseReadingsParams.
func (m *AlarmGlucoseReadingsParams) SetLimit(v int64) {
	m.Limit = v
}

// GetIncludeDeleted returns the IncludeDeleted value.
func (m AllGlucoseReadingsParams) GetIncludeDeleted() bool {
	return m.IncludeDeleted
}

// SetIncludeDeleted sets the IncludeDeleted value and returns the updated AllGlucoseReadingsParams.
func (m *AllGlucoseReadingsParams) SetIncludeDeleted(v bool) {
	m.IncludeDeleted = v
}

// GetSkip returns the Skip value.
func (m AllGlucoseReadingsParams) GetSkip() int64 {
	return m.Skip
}

// SetSkip sets the Skip value and returns the updated AllGlucoseReadingsParams.
func (m *AllGlucoseReadingsParams) SetSkip(v int64) {
	m.Skip = v
}

// GetLimit returns the Limit value.
func (m AllGlucoseReadingsParams) GetLimit() int64 {
	return m.Limit
}

// SetLimit sets the Limit value and returns the updated AllGlucoseReadingsParams.
func (m *AllGlucoseReadingsParams) SetLimit(v int64) {
	m.Limit = v
}

// GetID returns the ID value.
func (m CreateGlucoseReadingParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated CreateGlucoseReadingParams.
func (m *CreateGlucoseReadingParams) SetID(v string) {
	m.ID = v
}

// GetMgdl returns the Mgdl value.
func (m CreateGlucoseReadingParams) GetMgdl() float64 {
	return m.Mgdl
}

// SetMgdl sets the Mgdl value and returns the updated CreateGlucoseReadingParams.
func (m *CreateGlucoseReadingParams) SetMgdl(v float64) {
	m.Mgdl = v
}

// GetReadingType returns the ReadingType value.
func (m CreateGlucoseReadingParams) GetReadingType() string {
	return m.ReadingType
}

// SetReadingType sets the ReadingType value and returns the updated CreateGlucoseReadingParams.
func (m *CreateGlucoseReadingParams) SetReadingType(v string) {
	m.ReadingType = v
}

// GetUnits returns the Units value.
func (m CreateGlucoseReadingParams) GetUnits() *consts.Units {
	return m.Units
}

// SetUnits sets the Units value and returns the updated CreateGlucoseReadingParams.
func (m *CreateGlucoseReadingParams) SetUnits(v *consts.Units) {
	m.Units = v
}

// GetMeasurementMethod returns the MeasurementMethod value.
func (m CreateGlucoseReadingParams) GetMeasurementMethod() *string {
	return m.MeasurementMethod
}

// SetMeasurementMethod sets the MeasurementMethod value and returns the updated CreateGlucoseReadingParams.
func (m *CreateGlucoseReadingParams) SetMeasurementMethod(v *string) {
	m.MeasurementMethod = v
}

// GetDelta returns the Delta value.
func (m CreateGlucoseReadingParams) GetDelta() *float64 {
	return m.Delta
}

// SetDelta sets the Delta value and returns the updated CreateGlucoseReadingParams.
func (m *CreateGlucoseReadingParams) SetDelta(v *float64) {
	m.Delta = v
}

// GetDirection returns the Direction value.
func (m CreateGlucoseReadingParams) GetDirection() *string {
	return m.Direction
}

// SetDirection sets the Direction value and returns the updated CreateGlucoseReadingParams.
func (m *CreateGlucoseReadingParams) SetDirection(v *string) {
	m.Direction = v
}

// GetNoise returns the Noise value.
func (m CreateGlucoseReadingParams) GetNoise() *float64 {
	return m.Noise
}

// SetNoise sets the Noise value and returns the updated CreateGlucoseReadingParams.
func (m *CreateGlucoseReadingParams) SetNoise(v *float64) {
	m.Noise = v
}

// GetFiltered returns the Filtered value.
func (m CreateGlucoseReadingParams) GetFiltered() *float64 {
	return m.Filtered
}

// SetFiltered sets the Filtered value and returns the updated CreateGlucoseReadingParams.
func (m *CreateGlucoseReadingParams) SetFiltered(v *float64) {
	m.Filtered = v
}

// GetUnfiltered returns the Unfiltered value.
func (m CreateGlucoseReadingParams) GetUnfiltered() *float64 {
	return m.Unfiltered
}

// SetUnfiltered sets the Unfiltered value and returns the updated CreateGlucoseReadingParams.
func (m *CreateGlucoseReadingParams) SetUnfiltered(v *float64) {
	m.Unfiltered = v
}

// GetRssi returns the Rssi value.
func (m CreateGlucoseReadingParams) GetRssi() *float64 {
	return m.Rssi
}

// SetRssi sets the Rssi value and returns the updated CreateGlucoseReadingParams.
func (m *CreateGlucoseReadingParams) SetRssi(v *float64) {
	m.Rssi = v
}

// GetNotes returns the Notes value.
func (m CreateGlucoseReadingParams) GetNotes() *string {
	return m.Notes
}

// SetNotes sets the Notes value and returns the updated CreateGlucoseReadingParams.
func (m *CreateGlucoseReadingParams) SetNotes(v *string) {
	m.Notes = v
}

// GetReason returns the Reason value.
func (m CreateGlucoseReadingParams) GetReason() *string {
	return m.Reason
}

// SetReason sets the Reason value and returns the updated CreateGlucoseReadingParams.
func (m *CreateGlucoseReadingParams) SetReason(v *string) {
	m.Reason = v
}

// GetUtcOffsetMinutes returns the UtcOffsetMinutes value.
func (m CreateGlucoseReadingParams) GetUtcOffsetMinutes() *int64 {
	return m.UtcOffsetMinutes
}

// SetUtcOffsetMinutes sets the UtcOffsetMinutes value and returns the updated CreateGlucoseReadingParams.
func (m *CreateGlucoseReadingParams) SetUtcOffsetMinutes(v *int64) {
	m.UtcOffsetMinutes = v
}

// GetOccurredAt returns the OccurredAt value.
func (m CreateGlucoseReadingParams) GetOccurredAt() int64 {
	return m.OccurredAt
}

// SetOccurredAt sets the OccurredAt value and returns the updated CreateGlucoseReadingParams.
func (m *CreateGlucoseReadingParams) SetOccurredAt(v int64) {
	m.OccurredAt = v
}

// GetRawJsonData returns the RawJsonData value.
func (m CreateGlucoseReadingParams) GetRawJsonData() *string {
	return m.RawJsonData
}

// SetRawJsonData sets the RawJsonData value and returns the updated CreateGlucoseReadingParams.
func (m *CreateGlucoseReadingParams) SetRawJsonData(v *string) {
	m.RawJsonData = v
}

// GetCreatedByAppName returns the CreatedByAppName value.
func (m CreateGlucoseReadingParams) GetCreatedByAppName() *string {
	return m.CreatedByAppName
}

// SetCreatedByAppName sets the CreatedByAppName value and returns the updated CreateGlucoseReadingParams.
func (m *CreateGlucoseReadingParams) SetCreatedByAppName(v *string) {
	m.CreatedByAppName = v
}

// GetCreatedByDeviceName returns the CreatedByDeviceName value.
func (m CreateGlucoseReadingParams) GetCreatedByDeviceName() *string {
	return m.CreatedByDeviceName
}

// SetCreatedByDeviceName sets the CreatedByDeviceName value and returns the updated CreateGlucoseReadingParams.
func (m *CreateGlucoseReadingParams) SetCreatedByDeviceName(v *string) {
	m.CreatedByDeviceName = v
}

// GetCreatedByIdentityName returns the CreatedByIdentityName value.
func (m CreateGlucoseReadingParams) GetCreatedByIdentityName() string {
	return m.CreatedByIdentityName
}

// SetCreatedByIdentityName sets the CreatedByIdentityName value and returns the updated CreateGlucoseReadingParams.
func (m *CreateGlucoseReadingParams) SetCreatedByIdentityName(v string) {
	m.CreatedByIdentityName = v
}

// GetCreatedByIdentityID returns the CreatedByIdentityID value.
func (m CreateGlucoseReadingParams) GetCreatedByIdentityID() *int64 {
	return m.CreatedByIdentityID
}

// SetCreatedByIdentityID sets the CreatedByIdentityID value and returns the updated CreateGlucoseReadingParams.
func (m *CreateGlucoseReadingParams) SetCreatedByIdentityID(v *int64) {
	m.CreatedByIdentityID = v
}

// GetIncludeDeleted returns the IncludeDeleted value.
func (m GetGlucoseReadingParams) GetIncludeDeleted() bool {
	return m.IncludeDeleted
}

// SetIncludeDeleted sets the IncludeDeleted value and returns the updated GetGlucoseReadingParams.
func (m *GetGlucoseReadingParams) SetIncludeDeleted(v bool) {
	m.IncludeDeleted = v
}

// GetID returns the ID value.
func (m GetGlucoseReadingParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated GetGlucoseReadingParams.
func (m *GetGlucoseReadingParams) SetID(v string) {
	m.ID = v
}

// GetID returns the ID value.
func (m HardDeleteGlucoseReadingParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated HardDeleteGlucoseReadingParams.
func (m *HardDeleteGlucoseReadingParams) SetID(v string) {
	m.ID = v
}

// GetReadingType returns the ReadingType value.
func (m HardDeleteGlucoseReadingsByTypeParams) GetReadingType() string {
	return m.ReadingType
}

// SetReadingType sets the ReadingType value and returns the updated HardDeleteGlucoseReadingsByTypeParams.
func (m *HardDeleteGlucoseReadingsByTypeParams) SetReadingType(v string) {
	m.ReadingType = v
}

// GetDateGTE returns the DateGTE value.
func (m SearchGlucoseReadingsParams) GetDateGTE() *int64 {
	return m.DateGTE
}

// SetDateGTE sets the DateGTE value and returns the updated SearchGlucoseReadingsParams.
func (m *SearchGlucoseReadingsParams) SetDateGTE(v *int64) {
	m.DateGTE = v
}

// GetDateGT returns the DateGT value.
func (m SearchGlucoseReadingsParams) GetDateGT() *int64 {
	return m.DateGT
}

// SetDateGT sets the DateGT value and returns the updated SearchGlucoseReadingsParams.
func (m *SearchGlucoseReadingsParams) SetDateGT(v *int64) {
	m.DateGT = v
}

// GetDateEQ returns the DateEQ value.
func (m SearchGlucoseReadingsParams) GetDateEQ() *int64 {
	return m.DateEQ
}

// SetDateEQ sets the DateEQ value and returns the updated SearchGlucoseReadingsParams.
func (m *SearchGlucoseReadingsParams) SetDateEQ(v *int64) {
	m.DateEQ = v
}

// GetDateLTE returns the DateLTE value.
func (m SearchGlucoseReadingsParams) GetDateLTE() *int64 {
	return m.DateLTE
}

// SetDateLTE sets the DateLTE value and returns the updated SearchGlucoseReadingsParams.
func (m *SearchGlucoseReadingsParams) SetDateLTE(v *int64) {
	m.DateLTE = v
}

// GetDateLT returns the DateLT value.
func (m SearchGlucoseReadingsParams) GetDateLT() *int64 {
	return m.DateLT
}

// SetDateLT sets the DateLT value and returns the updated SearchGlucoseReadingsParams.
func (m *SearchGlucoseReadingsParams) SetDateLT(v *int64) {
	m.DateLT = v
}

// GetCreatedAtGTE returns the CreatedAtGTE value.
func (m SearchGlucoseReadingsParams) GetCreatedAtGTE() *int64 {
	return m.CreatedAtGTE
}

// SetCreatedAtGTE sets the CreatedAtGTE value and returns the updated SearchGlucoseReadingsParams.
func (m *SearchGlucoseReadingsParams) SetCreatedAtGTE(v *int64) {
	m.CreatedAtGTE = v
}

// GetCreatedAtLTE returns the CreatedAtLTE value.
func (m SearchGlucoseReadingsParams) GetCreatedAtLTE() *int64 {
	return m.CreatedAtLTE
}

// SetCreatedAtLTE sets the CreatedAtLTE value and returns the updated SearchGlucoseReadingsParams.
func (m *SearchGlucoseReadingsParams) SetCreatedAtLTE(v *int64) {
	m.CreatedAtLTE = v
}

// GetCreatedAtGT returns the CreatedAtGT value.
func (m SearchGlucoseReadingsParams) GetCreatedAtGT() *int64 {
	return m.CreatedAtGT
}

// SetCreatedAtGT sets the CreatedAtGT value and returns the updated SearchGlucoseReadingsParams.
func (m *SearchGlucoseReadingsParams) SetCreatedAtGT(v *int64) {
	m.CreatedAtGT = v
}

// GetCreatedAtLT returns the CreatedAtLT value.
func (m SearchGlucoseReadingsParams) GetCreatedAtLT() *int64 {
	return m.CreatedAtLT
}

// SetCreatedAtLT sets the CreatedAtLT value and returns the updated SearchGlucoseReadingsParams.
func (m *SearchGlucoseReadingsParams) SetCreatedAtLT(v *int64) {
	m.CreatedAtLT = v
}

// GetUpdatedGTE returns the UpdatedGTE value.
func (m SearchGlucoseReadingsParams) GetUpdatedGTE() *int64 {
	return m.UpdatedGTE
}

// SetUpdatedGTE sets the UpdatedGTE value and returns the updated SearchGlucoseReadingsParams.
func (m *SearchGlucoseReadingsParams) SetUpdatedGTE(v *int64) {
	m.UpdatedGTE = v
}

// GetUpdatedLTE returns the UpdatedLTE value.
func (m SearchGlucoseReadingsParams) GetUpdatedLTE() *int64 {
	return m.UpdatedLTE
}

// SetUpdatedLTE sets the UpdatedLTE value and returns the updated SearchGlucoseReadingsParams.
func (m *SearchGlucoseReadingsParams) SetUpdatedLTE(v *int64) {
	m.UpdatedLTE = v
}

// GetUpdatedGT returns the UpdatedGT value.
func (m SearchGlucoseReadingsParams) GetUpdatedGT() *int64 {
	return m.UpdatedGT
}

// SetUpdatedGT sets the UpdatedGT value and returns the updated SearchGlucoseReadingsParams.
func (m *SearchGlucoseReadingsParams) SetUpdatedGT(v *int64) {
	m.UpdatedGT = v
}

// GetUpdatedLT returns the UpdatedLT value.
func (m SearchGlucoseReadingsParams) GetUpdatedLT() *int64 {
	return m.UpdatedLT
}

// SetUpdatedLT sets the UpdatedLT value and returns the updated SearchGlucoseReadingsParams.
func (m *SearchGlucoseReadingsParams) SetUpdatedLT(v *int64) {
	m.UpdatedLT = v
}

// GetTypeEq returns the TypeEq value.
func (m SearchGlucoseReadingsParams) GetTypeEq() *string {
	return m.TypeEq
}

// SetTypeEq sets the TypeEq value and returns the updated SearchGlucoseReadingsParams.
func (m *SearchGlucoseReadingsParams) SetTypeEq(v *string) {
	m.TypeEq = v
}

// GetTypeNeq returns the TypeNeq value.
func (m SearchGlucoseReadingsParams) GetTypeNeq() *string {
	return m.TypeNeq
}

// SetTypeNeq sets the TypeNeq value and returns the updated SearchGlucoseReadingsParams.
func (m *SearchGlucoseReadingsParams) SetTypeNeq(v *string) {
	m.TypeNeq = v
}

// GetTypeIn returns the TypeIn value.
func (m SearchGlucoseReadingsParams) GetTypeIn() *string {
	return m.TypeIn
}

// SetTypeIn sets the TypeIn value and returns the updated SearchGlucoseReadingsParams.
func (m *SearchGlucoseReadingsParams) SetTypeIn(v *string) {
	m.TypeIn = v
}

// GetTypeNin returns the TypeNin value.
func (m SearchGlucoseReadingsParams) GetTypeNin() *string {
	return m.TypeNin
}

// SetTypeNin sets the TypeNin value and returns the updated SearchGlucoseReadingsParams.
func (m *SearchGlucoseReadingsParams) SetTypeNin(v *string) {
	m.TypeNin = v
}

// GetIDEq returns the IDEq value.
func (m SearchGlucoseReadingsParams) GetIDEq() *string {
	return m.IDEq
}

// SetIDEq sets the IDEq value and returns the updated SearchGlucoseReadingsParams.
func (m *SearchGlucoseReadingsParams) SetIDEq(v *string) {
	m.IDEq = v
}

// GetIDIn returns the IDIn value.
func (m SearchGlucoseReadingsParams) GetIDIn() *string {
	return m.IDIn
}

// SetIDIn sets the IDIn value and returns the updated SearchGlucoseReadingsParams.
func (m *SearchGlucoseReadingsParams) SetIDIn(v *string) {
	m.IDIn = v
}

// GetDeviceEq returns the DeviceEq value.
func (m SearchGlucoseReadingsParams) GetDeviceEq() *string {
	return m.DeviceEq
}

// SetDeviceEq sets the DeviceEq value and returns the updated SearchGlucoseReadingsParams.
func (m *SearchGlucoseReadingsParams) SetDeviceEq(v *string) {
	m.DeviceEq = v
}

// GetAppEq returns the AppEq value.
func (m SearchGlucoseReadingsParams) GetAppEq() *string {
	return m.AppEq
}

// SetAppEq sets the AppEq value and returns the updated SearchGlucoseReadingsParams.
func (m *SearchGlucoseReadingsParams) SetAppEq(v *string) {
	m.AppEq = v
}

// GetSgvEQ returns the SgvEQ value.
func (m SearchGlucoseReadingsParams) GetSgvEQ() *int64 {
	return m.SgvEQ
}

// SetSgvEQ sets the SgvEQ value and returns the updated SearchGlucoseReadingsParams.
func (m *SearchGlucoseReadingsParams) SetSgvEQ(v *int64) {
	m.SgvEQ = v
}

// GetSgvGTE returns the SgvGTE value.
func (m SearchGlucoseReadingsParams) GetSgvGTE() *int64 {
	return m.SgvGTE
}

// SetSgvGTE sets the SgvGTE value and returns the updated SearchGlucoseReadingsParams.
func (m *SearchGlucoseReadingsParams) SetSgvGTE(v *int64) {
	m.SgvGTE = v
}

// GetSgvGT returns the SgvGT value.
func (m SearchGlucoseReadingsParams) GetSgvGT() *int64 {
	return m.SgvGT
}

// SetSgvGT sets the SgvGT value and returns the updated SearchGlucoseReadingsParams.
func (m *SearchGlucoseReadingsParams) SetSgvGT(v *int64) {
	m.SgvGT = v
}

// GetSgvLTE returns the SgvLTE value.
func (m SearchGlucoseReadingsParams) GetSgvLTE() *int64 {
	return m.SgvLTE
}

// SetSgvLTE sets the SgvLTE value and returns the updated SearchGlucoseReadingsParams.
func (m *SearchGlucoseReadingsParams) SetSgvLTE(v *int64) {
	m.SgvLTE = v
}

// GetSgvLT returns the SgvLT value.
func (m SearchGlucoseReadingsParams) GetSgvLT() *int64 {
	return m.SgvLT
}

// SetSgvLT sets the SgvLT value and returns the updated SearchGlucoseReadingsParams.
func (m *SearchGlucoseReadingsParams) SetSgvLT(v *int64) {
	m.SgvLT = v
}

// GetIncludeDeleted returns the IncludeDeleted value.
func (m SearchGlucoseReadingsParams) GetIncludeDeleted() bool {
	return m.IncludeDeleted
}

// SetIncludeDeleted sets the IncludeDeleted value and returns the updated SearchGlucoseReadingsParams.
func (m *SearchGlucoseReadingsParams) SetIncludeDeleted(v bool) {
	m.IncludeDeleted = v
}

// GetSkip returns the Skip value.
func (m SearchGlucoseReadingsParams) GetSkip() int64 {
	return m.Skip
}

// SetSkip sets the Skip value and returns the updated SearchGlucoseReadingsParams.
func (m *SearchGlucoseReadingsParams) SetSkip(v int64) {
	m.Skip = v
}

// GetLimit returns the Limit value.
func (m SearchGlucoseReadingsParams) GetLimit() int64 {
	return m.Limit
}

// SetLimit sets the Limit value and returns the updated SearchGlucoseReadingsParams.
func (m *SearchGlucoseReadingsParams) SetLimit(v int64) {
	m.Limit = v
}

// GetUpdatedByIdentityName returns the UpdatedByIdentityName value.
func (m SoftDeleteGlucoseReadingParams) GetUpdatedByIdentityName() *string {
	return m.UpdatedByIdentityName
}

// SetUpdatedByIdentityName sets the UpdatedByIdentityName value and returns the updated SoftDeleteGlucoseReadingParams.
func (m *SoftDeleteGlucoseReadingParams) SetUpdatedByIdentityName(v *string) {
	m.UpdatedByIdentityName = v
}

// GetID returns the ID value.
func (m SoftDeleteGlucoseReadingParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated SoftDeleteGlucoseReadingParams.
func (m *SoftDeleteGlucoseReadingParams) SetID(v string) {
	m.ID = v
}

// GetUpdatedByIdentityName returns the UpdatedByIdentityName value.
func (m SoftDeleteGlucoseReadingsByTypeParams) GetUpdatedByIdentityName() *string {
	return m.UpdatedByIdentityName
}

// SetUpdatedByIdentityName sets the UpdatedByIdentityName value and returns the updated SoftDeleteGlucoseReadingsByTypeParams.
func (m *SoftDeleteGlucoseReadingsByTypeParams) SetUpdatedByIdentityName(v *string) {
	m.UpdatedByIdentityName = v
}

// GetReadingType returns the ReadingType value.
func (m SoftDeleteGlucoseReadingsByTypeParams) GetReadingType() string {
	return m.ReadingType
}

// SetReadingType sets the ReadingType value and returns the updated SoftDeleteGlucoseReadingsByTypeParams.
func (m *SoftDeleteGlucoseReadingsByTypeParams) SetReadingType(v string) {
	m.ReadingType = v
}

// GetMgdl returns the Mgdl value.
func (m UpdateGlucoseReadingParams) GetMgdl() *float64 {
	return m.Mgdl
}

// SetMgdl sets the Mgdl value and returns the updated UpdateGlucoseReadingParams.
func (m *UpdateGlucoseReadingParams) SetMgdl(v *float64) {
	m.Mgdl = v
}

// GetDelta returns the Delta value.
func (m UpdateGlucoseReadingParams) GetDelta() *float64 {
	return m.Delta
}

// SetDelta sets the Delta value and returns the updated UpdateGlucoseReadingParams.
func (m *UpdateGlucoseReadingParams) SetDelta(v *float64) {
	m.Delta = v
}

// GetDirection returns the Direction value.
func (m UpdateGlucoseReadingParams) GetDirection() *string {
	return m.Direction
}

// SetDirection sets the Direction value and returns the updated UpdateGlucoseReadingParams.
func (m *UpdateGlucoseReadingParams) SetDirection(v *string) {
	m.Direction = v
}

// GetNoise returns the Noise value.
func (m UpdateGlucoseReadingParams) GetNoise() *float64 {
	return m.Noise
}

// SetNoise sets the Noise value and returns the updated UpdateGlucoseReadingParams.
func (m *UpdateGlucoseReadingParams) SetNoise(v *float64) {
	m.Noise = v
}

// GetFiltered returns the Filtered value.
func (m UpdateGlucoseReadingParams) GetFiltered() *float64 {
	return m.Filtered
}

// SetFiltered sets the Filtered value and returns the updated UpdateGlucoseReadingParams.
func (m *UpdateGlucoseReadingParams) SetFiltered(v *float64) {
	m.Filtered = v
}

// GetUnfiltered returns the Unfiltered value.
func (m UpdateGlucoseReadingParams) GetUnfiltered() *float64 {
	return m.Unfiltered
}

// SetUnfiltered sets the Unfiltered value and returns the updated UpdateGlucoseReadingParams.
func (m *UpdateGlucoseReadingParams) SetUnfiltered(v *float64) {
	m.Unfiltered = v
}

// GetRssi returns the Rssi value.
func (m UpdateGlucoseReadingParams) GetRssi() *float64 {
	return m.Rssi
}

// SetRssi sets the Rssi value and returns the updated UpdateGlucoseReadingParams.
func (m *UpdateGlucoseReadingParams) SetRssi(v *float64) {
	m.Rssi = v
}

// GetNotes returns the Notes value.
func (m UpdateGlucoseReadingParams) GetNotes() *string {
	return m.Notes
}

// SetNotes sets the Notes value and returns the updated UpdateGlucoseReadingParams.
func (m *UpdateGlucoseReadingParams) SetNotes(v *string) {
	m.Notes = v
}

// GetReason returns the Reason value.
func (m UpdateGlucoseReadingParams) GetReason() *string {
	return m.Reason
}

// SetReason sets the Reason value and returns the updated UpdateGlucoseReadingParams.
func (m *UpdateGlucoseReadingParams) SetReason(v *string) {
	m.Reason = v
}

// GetRawJsonData returns the RawJsonData value.
func (m UpdateGlucoseReadingParams) GetRawJsonData() *string {
	return m.RawJsonData
}

// SetRawJsonData sets the RawJsonData value and returns the updated UpdateGlucoseReadingParams.
func (m *UpdateGlucoseReadingParams) SetRawJsonData(v *string) {
	m.RawJsonData = v
}

// GetUpdatedByIdentityID returns the UpdatedByIdentityID value.
func (m UpdateGlucoseReadingParams) GetUpdatedByIdentityID() *int64 {
	return m.UpdatedByIdentityID
}

// SetUpdatedByIdentityID sets the UpdatedByIdentityID value and returns the updated UpdateGlucoseReadingParams.
func (m *UpdateGlucoseReadingParams) SetUpdatedByIdentityID(v *int64) {
	m.UpdatedByIdentityID = v
}

// GetUpdatedByIdentityName returns the UpdatedByIdentityName value.
func (m UpdateGlucoseReadingParams) GetUpdatedByIdentityName() *string {
	return m.UpdatedByIdentityName
}

// SetUpdatedByIdentityName sets the UpdatedByIdentityName value and returns the updated UpdateGlucoseReadingParams.
func (m *UpdateGlucoseReadingParams) SetUpdatedByIdentityName(v *string) {
	m.UpdatedByIdentityName = v
}

// GetID returns the ID value.
func (m UpdateGlucoseReadingParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated UpdateGlucoseReadingParams.
func (m *UpdateGlucoseReadingParams) SetID(v string) {
	m.ID = v
}

// GetIncludeDeleted returns the IncludeDeleted value.
func (m AllEventGroupsParams) GetIncludeDeleted() bool {
	return m.IncludeDeleted
}

// SetIncludeDeleted sets the IncludeDeleted value and returns the updated AllEventGroupsParams.
func (m *AllEventGroupsParams) SetIncludeDeleted(v bool) {
	m.IncludeDeleted = v
}

// GetSkip returns the Skip value.
func (m AllEventGroupsParams) GetSkip() int64 {
	return m.Skip
}

// SetSkip sets the Skip value and returns the updated AllEventGroupsParams.
func (m *AllEventGroupsParams) SetSkip(v int64) {
	m.Skip = v
}

// GetLimit returns the Limit value.
func (m AllEventGroupsParams) GetLimit() int64 {
	return m.Limit
}

// SetLimit sets the Limit value and returns the updated AllEventGroupsParams.
func (m *AllEventGroupsParams) SetLimit(v int64) {
	m.Limit = v
}

// GetID returns the ID value.
func (m CreateEventGroupParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated CreateEventGroupParams.
func (m *CreateEventGroupParams) SetID(v string) {
	m.ID = v
}

// GetGroupType returns the GroupType value.
func (m CreateEventGroupParams) GetGroupType() string {
	return m.GroupType
}

// SetGroupType sets the GroupType value and returns the updated CreateEventGroupParams.
func (m *CreateEventGroupParams) SetGroupType(v string) {
	m.GroupType = v
}

// GetGlucoseReadingID returns the GlucoseReadingID value.
func (m CreateEventGroupParams) GetGlucoseReadingID() *string {
	return m.GlucoseReadingID
}

// SetGlucoseReadingID sets the GlucoseReadingID value and returns the updated CreateEventGroupParams.
func (m *CreateEventGroupParams) SetGlucoseReadingID(v *string) {
	m.GlucoseReadingID = v
}

// GetInsulinDoseID returns the InsulinDoseID value.
func (m CreateEventGroupParams) GetInsulinDoseID() *string {
	return m.InsulinDoseID
}

// SetInsulinDoseID sets the InsulinDoseID value and returns the updated CreateEventGroupParams.
func (m *CreateEventGroupParams) SetInsulinDoseID(v *string) {
	m.InsulinDoseID = v
}

// GetFoodIntakeID returns the FoodIntakeID value.
func (m CreateEventGroupParams) GetFoodIntakeID() *string {
	return m.FoodIntakeID
}

// SetFoodIntakeID sets the FoodIntakeID value and returns the updated CreateEventGroupParams.
func (m *CreateEventGroupParams) SetFoodIntakeID(v *string) {
	m.FoodIntakeID = v
}

// GetTempBasalID returns the TempBasalID value.
func (m CreateEventGroupParams) GetTempBasalID() *string {
	return m.TempBasalID
}

// SetTempBasalID sets the TempBasalID value and returns the updated CreateEventGroupParams.
func (m *CreateEventGroupParams) SetTempBasalID(v *string) {
	m.TempBasalID = v
}

// GetTempTargetID returns the TempTargetID value.
func (m CreateEventGroupParams) GetTempTargetID() *string {
	return m.TempTargetID
}

// SetTempTargetID sets the TempTargetID value and returns the updated CreateEventGroupParams.
func (m *CreateEventGroupParams) SetTempTargetID(v *string) {
	m.TempTargetID = v
}

// GetProfileSwitchID returns the ProfileSwitchID value.
func (m CreateEventGroupParams) GetProfileSwitchID() *string {
	return m.ProfileSwitchID
}

// SetProfileSwitchID sets the ProfileSwitchID value and returns the updated CreateEventGroupParams.
func (m *CreateEventGroupParams) SetProfileSwitchID(v *string) {
	m.ProfileSwitchID = v
}

// GetComponentEventID returns the ComponentEventID value.
func (m CreateEventGroupParams) GetComponentEventID() *string {
	return m.ComponentEventID
}

// SetComponentEventID sets the ComponentEventID value and returns the updated CreateEventGroupParams.
func (m *CreateEventGroupParams) SetComponentEventID(v *string) {
	m.ComponentEventID = v
}

// GetActivityID returns the ActivityID value.
func (m CreateEventGroupParams) GetActivityID() *string {
	return m.ActivityID
}

// SetActivityID sets the ActivityID value and returns the updated CreateEventGroupParams.
func (m *CreateEventGroupParams) SetActivityID(v *string) {
	m.ActivityID = v
}

// GetNotificationEventID returns the NotificationEventID value.
func (m CreateEventGroupParams) GetNotificationEventID() *string {
	return m.NotificationEventID
}

// SetNotificationEventID sets the NotificationEventID value and returns the updated CreateEventGroupParams.
func (m *CreateEventGroupParams) SetNotificationEventID(v *string) {
	m.NotificationEventID = v
}

// GetNotes returns the Notes value.
func (m CreateEventGroupParams) GetNotes() *string {
	return m.Notes
}

// SetNotes sets the Notes value and returns the updated CreateEventGroupParams.
func (m *CreateEventGroupParams) SetNotes(v *string) {
	m.Notes = v
}

// GetReason returns the Reason value.
func (m CreateEventGroupParams) GetReason() *string {
	return m.Reason
}

// SetReason sets the Reason value and returns the updated CreateEventGroupParams.
func (m *CreateEventGroupParams) SetReason(v *string) {
	m.Reason = v
}

// GetDurationMinutes returns the DurationMinutes value.
func (m CreateEventGroupParams) GetDurationMinutes() *int64 {
	return m.DurationMinutes
}

// SetDurationMinutes sets the DurationMinutes value and returns the updated CreateEventGroupParams.
func (m *CreateEventGroupParams) SetDurationMinutes(v *int64) {
	m.DurationMinutes = v
}

// GetEnteredBy returns the EnteredBy value.
func (m CreateEventGroupParams) GetEnteredBy() *string {
	return m.EnteredBy
}

// SetEnteredBy sets the EnteredBy value and returns the updated CreateEventGroupParams.
func (m *CreateEventGroupParams) SetEnteredBy(v *string) {
	m.EnteredBy = v
}

// GetUuid returns the Uuid value.
func (m CreateEventGroupParams) GetUuid() *string {
	return m.Uuid
}

// SetUuid sets the Uuid value and returns the updated CreateEventGroupParams.
func (m *CreateEventGroupParams) SetUuid(v *string) {
	m.Uuid = v
}

// GetOccurredAt returns the OccurredAt value.
func (m CreateEventGroupParams) GetOccurredAt() int64 {
	return m.OccurredAt
}

// SetOccurredAt sets the OccurredAt value and returns the updated CreateEventGroupParams.
func (m *CreateEventGroupParams) SetOccurredAt(v int64) {
	m.OccurredAt = v
}

// GetUtcOffsetMinutes returns the UtcOffsetMinutes value.
func (m CreateEventGroupParams) GetUtcOffsetMinutes() *int64 {
	return m.UtcOffsetMinutes
}

// SetUtcOffsetMinutes sets the UtcOffsetMinutes value and returns the updated CreateEventGroupParams.
func (m *CreateEventGroupParams) SetUtcOffsetMinutes(v *int64) {
	m.UtcOffsetMinutes = v
}

// GetRawJsonData returns the RawJsonData value.
func (m CreateEventGroupParams) GetRawJsonData() *string {
	return m.RawJsonData
}

// SetRawJsonData sets the RawJsonData value and returns the updated CreateEventGroupParams.
func (m *CreateEventGroupParams) SetRawJsonData(v *string) {
	m.RawJsonData = v
}

// GetCreatedByAppName returns the CreatedByAppName value.
func (m CreateEventGroupParams) GetCreatedByAppName() *string {
	return m.CreatedByAppName
}

// SetCreatedByAppName sets the CreatedByAppName value and returns the updated CreateEventGroupParams.
func (m *CreateEventGroupParams) SetCreatedByAppName(v *string) {
	m.CreatedByAppName = v
}

// GetCreatedByDeviceName returns the CreatedByDeviceName value.
func (m CreateEventGroupParams) GetCreatedByDeviceName() *string {
	return m.CreatedByDeviceName
}

// SetCreatedByDeviceName sets the CreatedByDeviceName value and returns the updated CreateEventGroupParams.
func (m *CreateEventGroupParams) SetCreatedByDeviceName(v *string) {
	m.CreatedByDeviceName = v
}

// GetCreatedByIdentityName returns the CreatedByIdentityName value.
func (m CreateEventGroupParams) GetCreatedByIdentityName() string {
	return m.CreatedByIdentityName
}

// SetCreatedByIdentityName sets the CreatedByIdentityName value and returns the updated CreateEventGroupParams.
func (m *CreateEventGroupParams) SetCreatedByIdentityName(v string) {
	m.CreatedByIdentityName = v
}

// GetCreatedByIdentityID returns the CreatedByIdentityID value.
func (m CreateEventGroupParams) GetCreatedByIdentityID() *int64 {
	return m.CreatedByIdentityID
}

// SetCreatedByIdentityID sets the CreatedByIdentityID value and returns the updated CreateEventGroupParams.
func (m *CreateEventGroupParams) SetCreatedByIdentityID(v *int64) {
	m.CreatedByIdentityID = v
}

// GetGroupID returns the GroupID value.
func (m EventsToGroupParams) GetGroupID() string {
	return m.GroupID
}

// SetGroupID sets the GroupID value and returns the updated EventsToGroupParams.
func (m *EventsToGroupParams) SetGroupID(v string) {
	m.GroupID = v
}

// GetFrom returns the From value.
func (m EventsToGroupParams) GetFrom() int64 {
	return m.From
}

// SetFrom sets the From value and returns the updated EventsToGroupParams.
func (m *EventsToGroupParams) SetFrom(v int64) {
	m.From = v
}

// GetTo returns the To value.
func (m EventsToGroupParams) GetTo() int64 {
	return m.To
}

// SetTo sets the To value and returns the updated EventsToGroupParams.
func (m *EventsToGroupParams) SetTo(v int64) {
	m.To = v
}

// GetID returns the ID value.
func (m GetEventGroupParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated GetEventGroupParams.
func (m *GetEventGroupParams) SetID(v string) {
	m.ID = v
}

// GetIncludeDeleted returns the IncludeDeleted value.
func (m GetEventGroupParams) GetIncludeDeleted() bool {
	return m.IncludeDeleted
}

// SetIncludeDeleted sets the IncludeDeleted value and returns the updated GetEventGroupParams.
func (m *GetEventGroupParams) SetIncludeDeleted(v bool) {
	m.IncludeDeleted = v
}

// GetGroupType returns the GroupType value.
func (m GetEventGroupOfTypeWithoutJoinsParams) GetGroupType() string {
	return m.GroupType
}

// SetGroupType sets the GroupType value and returns the updated GetEventGroupOfTypeWithoutJoinsParams.
func (m *GetEventGroupOfTypeWithoutJoinsParams) SetGroupType(v string) {
	m.GroupType = v
}

// GetIncludeDeleted returns the IncludeDeleted value.
func (m GetEventGroupOfTypeWithoutJoinsParams) GetIncludeDeleted() bool {
	return m.IncludeDeleted
}

// SetIncludeDeleted sets the IncludeDeleted value and returns the updated GetEventGroupOfTypeWithoutJoinsParams.
func (m *GetEventGroupOfTypeWithoutJoinsParams) SetIncludeDeleted(v bool) {
	m.IncludeDeleted = v
}

// GetLimit returns the Limit value.
func (m GetEventGroupOfTypeWithoutJoinsParams) GetLimit() int64 {
	return m.Limit
}

// SetLimit sets the Limit value and returns the updated GetEventGroupOfTypeWithoutJoinsParams.
func (m *GetEventGroupOfTypeWithoutJoinsParams) SetLimit(v int64) {
	m.Limit = v
}

// GetID returns the ID value.
func (m GetEventGroupWithoutJoinsParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated GetEventGroupWithoutJoinsParams.
func (m *GetEventGroupWithoutJoinsParams) SetID(v string) {
	m.ID = v
}

// GetIncludeDeleted returns the IncludeDeleted value.
func (m GetEventGroupWithoutJoinsParams) GetIncludeDeleted() bool {
	return m.IncludeDeleted
}

// SetIncludeDeleted sets the IncludeDeleted value and returns the updated GetEventGroupWithoutJoinsParams.
func (m *GetEventGroupWithoutJoinsParams) SetIncludeDeleted(v bool) {
	m.IncludeDeleted = v
}

// GetID returns the ID value.
func (m HardDeleteEventGroupParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated HardDeleteEventGroupParams.
func (m *HardDeleteEventGroupParams) SetID(v string) {
	m.ID = v
}

// GetGroupType returns the GroupType value.
func (m HardDeleteEventGroupsByTypeParams) GetGroupType() string {
	return m.GroupType
}

// SetGroupType sets the GroupType value and returns the updated HardDeleteEventGroupsByTypeParams.
func (m *HardDeleteEventGroupsByTypeParams) SetGroupType(v string) {
	m.GroupType = v
}

// GetGroupType returns the GroupType value.
func (m RegroupEventGroupParams) GetGroupType() *string {
	return m.GroupType
}

// SetGroupType sets the GroupType value and returns the updated RegroupEventGroupParams.
func (m *RegroupEventGroupParams) SetGroupType(v *string) {
	m.GroupType = v
}

// GetOccurredAt returns the OccurredAt value.
func (m RegroupEventGroupParams) GetOccurredAt() *int64 {
	return m.OccurredAt
}

// SetOccurredAt sets the OccurredAt value and returns the updated RegroupEventGroupParams.
func (m *RegroupEventGroupParams) SetOccurredAt(v *int64) {
	m.OccurredAt = v
}

// GetGlucoseReadingID returns the GlucoseReadingID value.
func (m RegroupEventGroupParams) GetGlucoseReadingID() *string {
	return m.GlucoseReadingID
}

// SetGlucoseReadingID sets the GlucoseReadingID value and returns the updated RegroupEventGroupParams.
func (m *RegroupEventGroupParams) SetGlucoseReadingID(v *string) {
	m.GlucoseReadingID = v
}

// GetInsulinDoseID returns the InsulinDoseID value.
func (m RegroupEventGroupParams) GetInsulinDoseID() *string {
	return m.InsulinDoseID
}

// SetInsulinDoseID sets the InsulinDoseID value and returns the updated RegroupEventGroupParams.
func (m *RegroupEventGroupParams) SetInsulinDoseID(v *string) {
	m.InsulinDoseID = v
}

// GetFoodIntakeID returns the FoodIntakeID value.
func (m RegroupEventGroupParams) GetFoodIntakeID() *string {
	return m.FoodIntakeID
}

// SetFoodIntakeID sets the FoodIntakeID value and returns the updated RegroupEventGroupParams.
func (m *RegroupEventGroupParams) SetFoodIntakeID(v *string) {
	m.FoodIntakeID = v
}

// GetTempBasalID returns the TempBasalID value.
func (m RegroupEventGroupParams) GetTempBasalID() *string {
	return m.TempBasalID
}

// SetTempBasalID sets the TempBasalID value and returns the updated RegroupEventGroupParams.
func (m *RegroupEventGroupParams) SetTempBasalID(v *string) {
	m.TempBasalID = v
}

// GetTempTargetID returns the TempTargetID value.
func (m RegroupEventGroupParams) GetTempTargetID() *string {
	return m.TempTargetID
}

// SetTempTargetID sets the TempTargetID value and returns the updated RegroupEventGroupParams.
func (m *RegroupEventGroupParams) SetTempTargetID(v *string) {
	m.TempTargetID = v
}

// GetProfileSwitchID returns the ProfileSwitchID value.
func (m RegroupEventGroupParams) GetProfileSwitchID() *string {
	return m.ProfileSwitchID
}

// SetProfileSwitchID sets the ProfileSwitchID value and returns the updated RegroupEventGroupParams.
func (m *RegroupEventGroupParams) SetProfileSwitchID(v *string) {
	m.ProfileSwitchID = v
}

// GetComponentEventID returns the ComponentEventID value.
func (m RegroupEventGroupParams) GetComponentEventID() *string {
	return m.ComponentEventID
}

// SetComponentEventID sets the ComponentEventID value and returns the updated RegroupEventGroupParams.
func (m *RegroupEventGroupParams) SetComponentEventID(v *string) {
	m.ComponentEventID = v
}

// GetActivityID returns the ActivityID value.
func (m RegroupEventGroupParams) GetActivityID() *string {
	return m.ActivityID
}

// SetActivityID sets the ActivityID value and returns the updated RegroupEventGroupParams.
func (m *RegroupEventGroupParams) SetActivityID(v *string) {
	m.ActivityID = v
}

// GetNotificationEventID returns the NotificationEventID value.
func (m RegroupEventGroupParams) GetNotificationEventID() *string {
	return m.NotificationEventID
}

// SetNotificationEventID sets the NotificationEventID value and returns the updated RegroupEventGroupParams.
func (m *RegroupEventGroupParams) SetNotificationEventID(v *string) {
	m.NotificationEventID = v
}

// GetNotes returns the Notes value.
func (m RegroupEventGroupParams) GetNotes() *string {
	return m.Notes
}

// SetNotes sets the Notes value and returns the updated RegroupEventGroupParams.
func (m *RegroupEventGroupParams) SetNotes(v *string) {
	m.Notes = v
}

// GetReason returns the Reason value.
func (m RegroupEventGroupParams) GetReason() *string {
	return m.Reason
}

// SetReason sets the Reason value and returns the updated RegroupEventGroupParams.
func (m *RegroupEventGroupParams) SetReason(v *string) {
	m.Reason = v
}

// GetUpdatedByIdentityID returns the UpdatedByIdentityID value.
func (m RegroupEventGroupParams) GetUpdatedByIdentityID() *int64 {
	return m.UpdatedByIdentityID
}

// SetUpdatedByIdentityID sets the UpdatedByIdentityID value and returns the updated RegroupEventGroupParams.
func (m *RegroupEventGroupParams) SetUpdatedByIdentityID(v *int64) {
	m.UpdatedByIdentityID = v
}

// GetUpdatedByIdentityName returns the UpdatedByIdentityName value.
func (m RegroupEventGroupParams) GetUpdatedByIdentityName() *string {
	return m.UpdatedByIdentityName
}

// SetUpdatedByIdentityName sets the UpdatedByIdentityName value and returns the updated RegroupEventGroupParams.
func (m *RegroupEventGroupParams) SetUpdatedByIdentityName(v *string) {
	m.UpdatedByIdentityName = v
}

// GetID returns the ID value.
func (m RegroupEventGroupParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated RegroupEventGroupParams.
func (m *RegroupEventGroupParams) SetID(v string) {
	m.ID = v
}

// GetDateGTE returns the DateGTE value.
func (m SearchEventGroupsParams) GetDateGTE() *int64 {
	return m.DateGTE
}

// SetDateGTE sets the DateGTE value and returns the updated SearchEventGroupsParams.
func (m *SearchEventGroupsParams) SetDateGTE(v *int64) {
	m.DateGTE = v
}

// GetDateGT returns the DateGT value.
func (m SearchEventGroupsParams) GetDateGT() *int64 {
	return m.DateGT
}

// SetDateGT sets the DateGT value and returns the updated SearchEventGroupsParams.
func (m *SearchEventGroupsParams) SetDateGT(v *int64) {
	m.DateGT = v
}

// GetDateEQ returns the DateEQ value.
func (m SearchEventGroupsParams) GetDateEQ() *int64 {
	return m.DateEQ
}

// SetDateEQ sets the DateEQ value and returns the updated SearchEventGroupsParams.
func (m *SearchEventGroupsParams) SetDateEQ(v *int64) {
	m.DateEQ = v
}

// GetDateLTE returns the DateLTE value.
func (m SearchEventGroupsParams) GetDateLTE() *int64 {
	return m.DateLTE
}

// SetDateLTE sets the DateLTE value and returns the updated SearchEventGroupsParams.
func (m *SearchEventGroupsParams) SetDateLTE(v *int64) {
	m.DateLTE = v
}

// GetDateLT returns the DateLT value.
func (m SearchEventGroupsParams) GetDateLT() *int64 {
	return m.DateLT
}

// SetDateLT sets the DateLT value and returns the updated SearchEventGroupsParams.
func (m *SearchEventGroupsParams) SetDateLT(v *int64) {
	m.DateLT = v
}

// GetCreatedAtGTE returns the CreatedAtGTE value.
func (m SearchEventGroupsParams) GetCreatedAtGTE() *int64 {
	return m.CreatedAtGTE
}

// SetCreatedAtGTE sets the CreatedAtGTE value and returns the updated SearchEventGroupsParams.
func (m *SearchEventGroupsParams) SetCreatedAtGTE(v *int64) {
	m.CreatedAtGTE = v
}

// GetCreatedAtLTE returns the CreatedAtLTE value.
func (m SearchEventGroupsParams) GetCreatedAtLTE() *int64 {
	return m.CreatedAtLTE
}

// SetCreatedAtLTE sets the CreatedAtLTE value and returns the updated SearchEventGroupsParams.
func (m *SearchEventGroupsParams) SetCreatedAtLTE(v *int64) {
	m.CreatedAtLTE = v
}

// GetCreatedAtGT returns the CreatedAtGT value.
func (m SearchEventGroupsParams) GetCreatedAtGT() *int64 {
	return m.CreatedAtGT
}

// SetCreatedAtGT sets the CreatedAtGT value and returns the updated SearchEventGroupsParams.
func (m *SearchEventGroupsParams) SetCreatedAtGT(v *int64) {
	m.CreatedAtGT = v
}

// GetCreatedAtLT returns the CreatedAtLT value.
func (m SearchEventGroupsParams) GetCreatedAtLT() *int64 {
	return m.CreatedAtLT
}

// SetCreatedAtLT sets the CreatedAtLT value and returns the updated SearchEventGroupsParams.
func (m *SearchEventGroupsParams) SetCreatedAtLT(v *int64) {
	m.CreatedAtLT = v
}

// GetUpdatedGTE returns the UpdatedGTE value.
func (m SearchEventGroupsParams) GetUpdatedGTE() *int64 {
	return m.UpdatedGTE
}

// SetUpdatedGTE sets the UpdatedGTE value and returns the updated SearchEventGroupsParams.
func (m *SearchEventGroupsParams) SetUpdatedGTE(v *int64) {
	m.UpdatedGTE = v
}

// GetUpdatedLTE returns the UpdatedLTE value.
func (m SearchEventGroupsParams) GetUpdatedLTE() *int64 {
	return m.UpdatedLTE
}

// SetUpdatedLTE sets the UpdatedLTE value and returns the updated SearchEventGroupsParams.
func (m *SearchEventGroupsParams) SetUpdatedLTE(v *int64) {
	m.UpdatedLTE = v
}

// GetUpdatedGT returns the UpdatedGT value.
func (m SearchEventGroupsParams) GetUpdatedGT() *int64 {
	return m.UpdatedGT
}

// SetUpdatedGT sets the UpdatedGT value and returns the updated SearchEventGroupsParams.
func (m *SearchEventGroupsParams) SetUpdatedGT(v *int64) {
	m.UpdatedGT = v
}

// GetUpdatedLT returns the UpdatedLT value.
func (m SearchEventGroupsParams) GetUpdatedLT() *int64 {
	return m.UpdatedLT
}

// SetUpdatedLT sets the UpdatedLT value and returns the updated SearchEventGroupsParams.
func (m *SearchEventGroupsParams) SetUpdatedLT(v *int64) {
	m.UpdatedLT = v
}

// GetTypeEq returns the TypeEq value.
func (m SearchEventGroupsParams) GetTypeEq() *string {
	return m.TypeEq
}

// SetTypeEq sets the TypeEq value and returns the updated SearchEventGroupsParams.
func (m *SearchEventGroupsParams) SetTypeEq(v *string) {
	m.TypeEq = v
}

// GetTypeNeq returns the TypeNeq value.
func (m SearchEventGroupsParams) GetTypeNeq() *string {
	return m.TypeNeq
}

// SetTypeNeq sets the TypeNeq value and returns the updated SearchEventGroupsParams.
func (m *SearchEventGroupsParams) SetTypeNeq(v *string) {
	m.TypeNeq = v
}

// GetTypeIn returns the TypeIn value.
func (m SearchEventGroupsParams) GetTypeIn() *string {
	return m.TypeIn
}

// SetTypeIn sets the TypeIn value and returns the updated SearchEventGroupsParams.
func (m *SearchEventGroupsParams) SetTypeIn(v *string) {
	m.TypeIn = v
}

// GetTypeNin returns the TypeNin value.
func (m SearchEventGroupsParams) GetTypeNin() *string {
	return m.TypeNin
}

// SetTypeNin sets the TypeNin value and returns the updated SearchEventGroupsParams.
func (m *SearchEventGroupsParams) SetTypeNin(v *string) {
	m.TypeNin = v
}

// GetIDEq returns the IDEq value.
func (m SearchEventGroupsParams) GetIDEq() *string {
	return m.IDEq
}

// SetIDEq sets the IDEq value and returns the updated SearchEventGroupsParams.
func (m *SearchEventGroupsParams) SetIDEq(v *string) {
	m.IDEq = v
}

// GetIDIn returns the IDIn value.
func (m SearchEventGroupsParams) GetIDIn() *string {
	return m.IDIn
}

// SetIDIn sets the IDIn value and returns the updated SearchEventGroupsParams.
func (m *SearchEventGroupsParams) SetIDIn(v *string) {
	m.IDIn = v
}

// GetDeviceEq returns the DeviceEq value.
func (m SearchEventGroupsParams) GetDeviceEq() *string {
	return m.DeviceEq
}

// SetDeviceEq sets the DeviceEq value and returns the updated SearchEventGroupsParams.
func (m *SearchEventGroupsParams) SetDeviceEq(v *string) {
	m.DeviceEq = v
}

// GetAppEq returns the AppEq value.
func (m SearchEventGroupsParams) GetAppEq() *string {
	return m.AppEq
}

// SetAppEq sets the AppEq value and returns the updated SearchEventGroupsParams.
func (m *SearchEventGroupsParams) SetAppEq(v *string) {
	m.AppEq = v
}

// GetNameEq returns the NameEq value.
func (m SearchEventGroupsParams) GetNameEq() *string {
	return m.NameEq
}

// SetNameEq sets the NameEq value and returns the updated SearchEventGroupsParams.
func (m *SearchEventGroupsParams) SetNameEq(v *string) {
	m.NameEq = v
}

// GetSgvEQ returns the SgvEQ value.
func (m SearchEventGroupsParams) GetSgvEQ() *int64 {
	return m.SgvEQ
}

// SetSgvEQ sets the SgvEQ value and returns the updated SearchEventGroupsParams.
func (m *SearchEventGroupsParams) SetSgvEQ(v *int64) {
	m.SgvEQ = v
}

// GetSgvGTE returns the SgvGTE value.
func (m SearchEventGroupsParams) GetSgvGTE() *int64 {
	return m.SgvGTE
}

// SetSgvGTE sets the SgvGTE value and returns the updated SearchEventGroupsParams.
func (m *SearchEventGroupsParams) SetSgvGTE(v *int64) {
	m.SgvGTE = v
}

// GetSgvGT returns the SgvGT value.
func (m SearchEventGroupsParams) GetSgvGT() *int64 {
	return m.SgvGT
}

// SetSgvGT sets the SgvGT value and returns the updated SearchEventGroupsParams.
func (m *SearchEventGroupsParams) SetSgvGT(v *int64) {
	m.SgvGT = v
}

// GetSgvLTE returns the SgvLTE value.
func (m SearchEventGroupsParams) GetSgvLTE() *int64 {
	return m.SgvLTE
}

// SetSgvLTE sets the SgvLTE value and returns the updated SearchEventGroupsParams.
func (m *SearchEventGroupsParams) SetSgvLTE(v *int64) {
	m.SgvLTE = v
}

// GetSgvLT returns the SgvLT value.
func (m SearchEventGroupsParams) GetSgvLT() *int64 {
	return m.SgvLT
}

// SetSgvLT sets the SgvLT value and returns the updated SearchEventGroupsParams.
func (m *SearchEventGroupsParams) SetSgvLT(v *int64) {
	m.SgvLT = v
}

// GetCarbsEQ returns the CarbsEQ value.
func (m SearchEventGroupsParams) GetCarbsEQ() *int64 {
	return m.CarbsEQ
}

// SetCarbsEQ sets the CarbsEQ value and returns the updated SearchEventGroupsParams.
func (m *SearchEventGroupsParams) SetCarbsEQ(v *int64) {
	m.CarbsEQ = v
}

// GetCarbsGTE returns the CarbsGTE value.
func (m SearchEventGroupsParams) GetCarbsGTE() *int64 {
	return m.CarbsGTE
}

// SetCarbsGTE sets the CarbsGTE value and returns the updated SearchEventGroupsParams.
func (m *SearchEventGroupsParams) SetCarbsGTE(v *int64) {
	m.CarbsGTE = v
}

// GetCarbsGT returns the CarbsGT value.
func (m SearchEventGroupsParams) GetCarbsGT() *int64 {
	return m.CarbsGT
}

// SetCarbsGT sets the CarbsGT value and returns the updated SearchEventGroupsParams.
func (m *SearchEventGroupsParams) SetCarbsGT(v *int64) {
	m.CarbsGT = v
}

// GetCarbsLTE returns the CarbsLTE value.
func (m SearchEventGroupsParams) GetCarbsLTE() *int64 {
	return m.CarbsLTE
}

// SetCarbsLTE sets the CarbsLTE value and returns the updated SearchEventGroupsParams.
func (m *SearchEventGroupsParams) SetCarbsLTE(v *int64) {
	m.CarbsLTE = v
}

// GetCarbsLT returns the CarbsLT value.
func (m SearchEventGroupsParams) GetCarbsLT() *int64 {
	return m.CarbsLT
}

// SetCarbsLT sets the CarbsLT value and returns the updated SearchEventGroupsParams.
func (m *SearchEventGroupsParams) SetCarbsLT(v *int64) {
	m.CarbsLT = v
}

// GetInsulinEQ returns the InsulinEQ value.
func (m SearchEventGroupsParams) GetInsulinEQ() *int64 {
	return m.InsulinEQ
}

// SetInsulinEQ sets the InsulinEQ value and returns the updated SearchEventGroupsParams.
func (m *SearchEventGroupsParams) SetInsulinEQ(v *int64) {
	m.InsulinEQ = v
}

// GetInsulinGTE returns the InsulinGTE value.
func (m SearchEventGroupsParams) GetInsulinGTE() *int64 {
	return m.InsulinGTE
}

// SetInsulinGTE sets the InsulinGTE value and returns the updated SearchEventGroupsParams.
func (m *SearchEventGroupsParams) SetInsulinGTE(v *int64) {
	m.InsulinGTE = v
}

// GetInsulinGT returns the InsulinGT value.
func (m SearchEventGroupsParams) GetInsulinGT() *int64 {
	return m.InsulinGT
}

// SetInsulinGT sets the InsulinGT value and returns the updated SearchEventGroupsParams.
func (m *SearchEventGroupsParams) SetInsulinGT(v *int64) {
	m.InsulinGT = v
}

// GetInsulinLTE returns the InsulinLTE value.
func (m SearchEventGroupsParams) GetInsulinLTE() *int64 {
	return m.InsulinLTE
}

// SetInsulinLTE sets the InsulinLTE value and returns the updated SearchEventGroupsParams.
func (m *SearchEventGroupsParams) SetInsulinLTE(v *int64) {
	m.InsulinLTE = v
}

// GetInsulinLT returns the InsulinLT value.
func (m SearchEventGroupsParams) GetInsulinLT() *int64 {
	return m.InsulinLT
}

// SetInsulinLT sets the InsulinLT value and returns the updated SearchEventGroupsParams.
func (m *SearchEventGroupsParams) SetInsulinLT(v *int64) {
	m.InsulinLT = v
}

// GetDurationEQ returns the DurationEQ value.
func (m SearchEventGroupsParams) GetDurationEQ() *int64 {
	return m.DurationEQ
}

// SetDurationEQ sets the DurationEQ value and returns the updated SearchEventGroupsParams.
func (m *SearchEventGroupsParams) SetDurationEQ(v *int64) {
	m.DurationEQ = v
}

// GetDurationGTE returns the DurationGTE value.
func (m SearchEventGroupsParams) GetDurationGTE() *int64 {
	return m.DurationGTE
}

// SetDurationGTE sets the DurationGTE value and returns the updated SearchEventGroupsParams.
func (m *SearchEventGroupsParams) SetDurationGTE(v *int64) {
	m.DurationGTE = v
}

// GetDurationGT returns the DurationGT value.
func (m SearchEventGroupsParams) GetDurationGT() *int64 {
	return m.DurationGT
}

// SetDurationGT sets the DurationGT value and returns the updated SearchEventGroupsParams.
func (m *SearchEventGroupsParams) SetDurationGT(v *int64) {
	m.DurationGT = v
}

// GetDurationLTE returns the DurationLTE value.
func (m SearchEventGroupsParams) GetDurationLTE() *int64 {
	return m.DurationLTE
}

// SetDurationLTE sets the DurationLTE value and returns the updated SearchEventGroupsParams.
func (m *SearchEventGroupsParams) SetDurationLTE(v *int64) {
	m.DurationLTE = v
}

// GetDurationLT returns the DurationLT value.
func (m SearchEventGroupsParams) GetDurationLT() *int64 {
	return m.DurationLT
}

// SetDurationLT sets the DurationLT value and returns the updated SearchEventGroupsParams.
func (m *SearchEventGroupsParams) SetDurationLT(v *int64) {
	m.DurationLT = v
}

// GetCarbsExists returns the CarbsExists value.
func (m SearchEventGroupsParams) GetCarbsExists() *bool {
	return m.CarbsExists
}

// SetCarbsExists sets the CarbsExists value and returns the updated SearchEventGroupsParams.
func (m *SearchEventGroupsParams) SetCarbsExists(v *bool) {
	m.CarbsExists = v
}

// GetDurationExists returns the DurationExists value.
func (m SearchEventGroupsParams) GetDurationExists() *bool {
	return m.DurationExists
}

// SetDurationExists sets the DurationExists value and returns the updated SearchEventGroupsParams.
func (m *SearchEventGroupsParams) SetDurationExists(v *bool) {
	m.DurationExists = v
}

// GetEnteredByEq returns the EnteredByEq value.
func (m SearchEventGroupsParams) GetEnteredByEq() *string {
	return m.EnteredByEq
}

// SetEnteredByEq sets the EnteredByEq value and returns the updated SearchEventGroupsParams.
func (m *SearchEventGroupsParams) SetEnteredByEq(v *string) {
	m.EnteredByEq = v
}

// GetEnteredByNin returns the EnteredByNin value.
func (m SearchEventGroupsParams) GetEnteredByNin() *string {
	return m.EnteredByNin
}

// SetEnteredByNin sets the EnteredByNin value and returns the updated SearchEventGroupsParams.
func (m *SearchEventGroupsParams) SetEnteredByNin(v *string) {
	m.EnteredByNin = v
}

// GetIncludeDeleted returns the IncludeDeleted value.
func (m SearchEventGroupsParams) GetIncludeDeleted() bool {
	return m.IncludeDeleted
}

// SetIncludeDeleted sets the IncludeDeleted value and returns the updated SearchEventGroupsParams.
func (m *SearchEventGroupsParams) SetIncludeDeleted(v bool) {
	m.IncludeDeleted = v
}

// GetSkip returns the Skip value.
func (m SearchEventGroupsParams) GetSkip() int64 {
	return m.Skip
}

// SetSkip sets the Skip value and returns the updated SearchEventGroupsParams.
func (m *SearchEventGroupsParams) SetSkip(v int64) {
	m.Skip = v
}

// GetLimit returns the Limit value.
func (m SearchEventGroupsParams) GetLimit() int64 {
	return m.Limit
}

// SetLimit sets the Limit value and returns the updated SearchEventGroupsParams.
func (m *SearchEventGroupsParams) SetLimit(v int64) {
	m.Limit = v
}

// GetUpdatedByIdentityName returns the UpdatedByIdentityName value.
func (m SoftDeleteEventGroupParams) GetUpdatedByIdentityName() *string {
	return m.UpdatedByIdentityName
}

// SetUpdatedByIdentityName sets the UpdatedByIdentityName value and returns the updated SoftDeleteEventGroupParams.
func (m *SoftDeleteEventGroupParams) SetUpdatedByIdentityName(v *string) {
	m.UpdatedByIdentityName = v
}

// GetID returns the ID value.
func (m SoftDeleteEventGroupParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated SoftDeleteEventGroupParams.
func (m *SoftDeleteEventGroupParams) SetID(v string) {
	m.ID = v
}

// GetUpdatedByIdentityName returns the UpdatedByIdentityName value.
func (m SoftDeleteEventGroupsByTypeParams) GetUpdatedByIdentityName() *string {
	return m.UpdatedByIdentityName
}

// SetUpdatedByIdentityName sets the UpdatedByIdentityName value and returns the updated SoftDeleteEventGroupsByTypeParams.
func (m *SoftDeleteEventGroupsByTypeParams) SetUpdatedByIdentityName(v *string) {
	m.UpdatedByIdentityName = v
}

// GetGroupType returns the GroupType value.
func (m SoftDeleteEventGroupsByTypeParams) GetGroupType() string {
	return m.GroupType
}

// SetGroupType sets the GroupType value and returns the updated SoftDeleteEventGroupsByTypeParams.
func (m *SoftDeleteEventGroupsByTypeParams) SetGroupType(v string) {
	m.GroupType = v
}

// GetGlucoseReadingID returns the GlucoseReadingID value.
func (m UpdateEventGroupParams) GetGlucoseReadingID() *string {
	return m.GlucoseReadingID
}

// SetGlucoseReadingID sets the GlucoseReadingID value and returns the updated UpdateEventGroupParams.
func (m *UpdateEventGroupParams) SetGlucoseReadingID(v *string) {
	m.GlucoseReadingID = v
}

// GetInsulinDoseID returns the InsulinDoseID value.
func (m UpdateEventGroupParams) GetInsulinDoseID() *string {
	return m.InsulinDoseID
}

// SetInsulinDoseID sets the InsulinDoseID value and returns the updated UpdateEventGroupParams.
func (m *UpdateEventGroupParams) SetInsulinDoseID(v *string) {
	m.InsulinDoseID = v
}

// GetFoodIntakeID returns the FoodIntakeID value.
func (m UpdateEventGroupParams) GetFoodIntakeID() *string {
	return m.FoodIntakeID
}

// SetFoodIntakeID sets the FoodIntakeID value and returns the updated UpdateEventGroupParams.
func (m *UpdateEventGroupParams) SetFoodIntakeID(v *string) {
	m.FoodIntakeID = v
}

// GetTempBasalID returns the TempBasalID value.
func (m UpdateEventGroupParams) GetTempBasalID() *string {
	return m.TempBasalID
}

// SetTempBasalID sets the TempBasalID value and returns the updated UpdateEventGroupParams.
func (m *UpdateEventGroupParams) SetTempBasalID(v *string) {
	m.TempBasalID = v
}

// GetTempTargetID returns the TempTargetID value.
func (m UpdateEventGroupParams) GetTempTargetID() *string {
	return m.TempTargetID
}

// SetTempTargetID sets the TempTargetID value and returns the updated UpdateEventGroupParams.
func (m *UpdateEventGroupParams) SetTempTargetID(v *string) {
	m.TempTargetID = v
}

// GetProfileSwitchID returns the ProfileSwitchID value.
func (m UpdateEventGroupParams) GetProfileSwitchID() *string {
	return m.ProfileSwitchID
}

// SetProfileSwitchID sets the ProfileSwitchID value and returns the updated UpdateEventGroupParams.
func (m *UpdateEventGroupParams) SetProfileSwitchID(v *string) {
	m.ProfileSwitchID = v
}

// GetComponentEventID returns the ComponentEventID value.
func (m UpdateEventGroupParams) GetComponentEventID() *string {
	return m.ComponentEventID
}

// SetComponentEventID sets the ComponentEventID value and returns the updated UpdateEventGroupParams.
func (m *UpdateEventGroupParams) SetComponentEventID(v *string) {
	m.ComponentEventID = v
}

// GetActivityID returns the ActivityID value.
func (m UpdateEventGroupParams) GetActivityID() *string {
	return m.ActivityID
}

// SetActivityID sets the ActivityID value and returns the updated UpdateEventGroupParams.
func (m *UpdateEventGroupParams) SetActivityID(v *string) {
	m.ActivityID = v
}

// GetNotificationEventID returns the NotificationEventID value.
func (m UpdateEventGroupParams) GetNotificationEventID() *string {
	return m.NotificationEventID
}

// SetNotificationEventID sets the NotificationEventID value and returns the updated UpdateEventGroupParams.
func (m *UpdateEventGroupParams) SetNotificationEventID(v *string) {
	m.NotificationEventID = v
}

// GetNotes returns the Notes value.
func (m UpdateEventGroupParams) GetNotes() *string {
	return m.Notes
}

// SetNotes sets the Notes value and returns the updated UpdateEventGroupParams.
func (m *UpdateEventGroupParams) SetNotes(v *string) {
	m.Notes = v
}

// GetReason returns the Reason value.
func (m UpdateEventGroupParams) GetReason() *string {
	return m.Reason
}

// SetReason sets the Reason value and returns the updated UpdateEventGroupParams.
func (m *UpdateEventGroupParams) SetReason(v *string) {
	m.Reason = v
}

// GetEnteredBy returns the EnteredBy value.
func (m UpdateEventGroupParams) GetEnteredBy() *string {
	return m.EnteredBy
}

// SetEnteredBy sets the EnteredBy value and returns the updated UpdateEventGroupParams.
func (m *UpdateEventGroupParams) SetEnteredBy(v *string) {
	m.EnteredBy = v
}

// GetUuid returns the Uuid value.
func (m UpdateEventGroupParams) GetUuid() *string {
	return m.Uuid
}

// SetUuid sets the Uuid value and returns the updated UpdateEventGroupParams.
func (m *UpdateEventGroupParams) SetUuid(v *string) {
	m.Uuid = v
}

// GetRawJsonData returns the RawJsonData value.
func (m UpdateEventGroupParams) GetRawJsonData() *string {
	return m.RawJsonData
}

// SetRawJsonData sets the RawJsonData value and returns the updated UpdateEventGroupParams.
func (m *UpdateEventGroupParams) SetRawJsonData(v *string) {
	m.RawJsonData = v
}

// GetUpdatedByIdentityID returns the UpdatedByIdentityID value.
func (m UpdateEventGroupParams) GetUpdatedByIdentityID() *int64 {
	return m.UpdatedByIdentityID
}

// SetUpdatedByIdentityID sets the UpdatedByIdentityID value and returns the updated UpdateEventGroupParams.
func (m *UpdateEventGroupParams) SetUpdatedByIdentityID(v *int64) {
	m.UpdatedByIdentityID = v
}

// GetUpdatedByIdentityName returns the UpdatedByIdentityName value.
func (m UpdateEventGroupParams) GetUpdatedByIdentityName() *string {
	return m.UpdatedByIdentityName
}

// SetUpdatedByIdentityName sets the UpdatedByIdentityName value and returns the updated UpdateEventGroupParams.
func (m *UpdateEventGroupParams) SetUpdatedByIdentityName(v *string) {
	m.UpdatedByIdentityName = v
}

// GetID returns the ID value.
func (m UpdateEventGroupParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated UpdateEventGroupParams.
func (m *UpdateEventGroupParams) SetID(v string) {
	m.ID = v
}

// GetIncludeDeleted returns the IncludeDeleted value.
func (m AllInsulinDosesParams) GetIncludeDeleted() bool {
	return m.IncludeDeleted
}

// SetIncludeDeleted sets the IncludeDeleted value and returns the updated AllInsulinDosesParams.
func (m *AllInsulinDosesParams) SetIncludeDeleted(v bool) {
	m.IncludeDeleted = v
}

// GetSkip returns the Skip value.
func (m AllInsulinDosesParams) GetSkip() int64 {
	return m.Skip
}

// SetSkip sets the Skip value and returns the updated AllInsulinDosesParams.
func (m *AllInsulinDosesParams) SetSkip(v int64) {
	m.Skip = v
}

// GetLimit returns the Limit value.
func (m AllInsulinDosesParams) GetLimit() int64 {
	return m.Limit
}

// SetLimit sets the Limit value and returns the updated AllInsulinDosesParams.
func (m *AllInsulinDosesParams) SetLimit(v int64) {
	m.Limit = v
}

// GetID returns the ID value.
func (m CreateInsulinDoseParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated CreateInsulinDoseParams.
func (m *CreateInsulinDoseParams) SetID(v string) {
	m.ID = v
}

// GetDose returns the Dose value.
func (m CreateInsulinDoseParams) GetDose() float64 {
	return m.Dose
}

// SetDose sets the Dose value and returns the updated CreateInsulinDoseParams.
func (m *CreateInsulinDoseParams) SetDose(v float64) {
	m.Dose = v
}

// GetDoseType returns the DoseType value.
func (m CreateInsulinDoseParams) GetDoseType() *string {
	return m.DoseType
}

// SetDoseType sets the DoseType value and returns the updated CreateInsulinDoseParams.
func (m *CreateInsulinDoseParams) SetDoseType(v *string) {
	m.DoseType = v
}

// GetSplitNow returns the SplitNow value.
func (m CreateInsulinDoseParams) GetSplitNow() *float64 {
	return m.SplitNow
}

// SetSplitNow sets the SplitNow value and returns the updated CreateInsulinDoseParams.
func (m *CreateInsulinDoseParams) SetSplitNow(v *float64) {
	m.SplitNow = v
}

// GetSplitExt returns the SplitExt value.
func (m CreateInsulinDoseParams) GetSplitExt() *float64 {
	return m.SplitExt
}

// SetSplitExt sets the SplitExt value and returns the updated CreateInsulinDoseParams.
func (m *CreateInsulinDoseParams) SetSplitExt(v *float64) {
	m.SplitExt = v
}

// GetNotes returns the Notes value.
func (m CreateInsulinDoseParams) GetNotes() *string {
	return m.Notes
}

// SetNotes sets the Notes value and returns the updated CreateInsulinDoseParams.
func (m *CreateInsulinDoseParams) SetNotes(v *string) {
	m.Notes = v
}

// GetReason returns the Reason value.
func (m CreateInsulinDoseParams) GetReason() *string {
	return m.Reason
}

// SetReason sets the Reason value and returns the updated CreateInsulinDoseParams.
func (m *CreateInsulinDoseParams) SetReason(v *string) {
	m.Reason = v
}

// GetUtcOffsetMinutes returns the UtcOffsetMinutes value.
func (m CreateInsulinDoseParams) GetUtcOffsetMinutes() *int64 {
	return m.UtcOffsetMinutes
}

// SetUtcOffsetMinutes sets the UtcOffsetMinutes value and returns the updated CreateInsulinDoseParams.
func (m *CreateInsulinDoseParams) SetUtcOffsetMinutes(v *int64) {
	m.UtcOffsetMinutes = v
}

// GetOccurredAt returns the OccurredAt value.
func (m CreateInsulinDoseParams) GetOccurredAt() int64 {
	return m.OccurredAt
}

// SetOccurredAt sets the OccurredAt value and returns the updated CreateInsulinDoseParams.
func (m *CreateInsulinDoseParams) SetOccurredAt(v int64) {
	m.OccurredAt = v
}

// GetRawJsonData returns the RawJsonData value.
func (m CreateInsulinDoseParams) GetRawJsonData() *string {
	return m.RawJsonData
}

// SetRawJsonData sets the RawJsonData value and returns the updated CreateInsulinDoseParams.
func (m *CreateInsulinDoseParams) SetRawJsonData(v *string) {
	m.RawJsonData = v
}

// GetCreatedByAppName returns the CreatedByAppName value.
func (m CreateInsulinDoseParams) GetCreatedByAppName() *string {
	return m.CreatedByAppName
}

// SetCreatedByAppName sets the CreatedByAppName value and returns the updated CreateInsulinDoseParams.
func (m *CreateInsulinDoseParams) SetCreatedByAppName(v *string) {
	m.CreatedByAppName = v
}

// GetCreatedByDeviceName returns the CreatedByDeviceName value.
func (m CreateInsulinDoseParams) GetCreatedByDeviceName() *string {
	return m.CreatedByDeviceName
}

// SetCreatedByDeviceName sets the CreatedByDeviceName value and returns the updated CreateInsulinDoseParams.
func (m *CreateInsulinDoseParams) SetCreatedByDeviceName(v *string) {
	m.CreatedByDeviceName = v
}

// GetCreatedByIdentityName returns the CreatedByIdentityName value.
func (m CreateInsulinDoseParams) GetCreatedByIdentityName() string {
	return m.CreatedByIdentityName
}

// SetCreatedByIdentityName sets the CreatedByIdentityName value and returns the updated CreateInsulinDoseParams.
func (m *CreateInsulinDoseParams) SetCreatedByIdentityName(v string) {
	m.CreatedByIdentityName = v
}

// GetCreatedByIdentityID returns the CreatedByIdentityID value.
func (m CreateInsulinDoseParams) GetCreatedByIdentityID() *int64 {
	return m.CreatedByIdentityID
}

// SetCreatedByIdentityID sets the CreatedByIdentityID value and returns the updated CreateInsulinDoseParams.
func (m *CreateInsulinDoseParams) SetCreatedByIdentityID(v *int64) {
	m.CreatedByIdentityID = v
}

// GetID returns the ID value.
func (m GetInsulinDoseParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated GetInsulinDoseParams.
func (m *GetInsulinDoseParams) SetID(v string) {
	m.ID = v
}

// GetIncludeDeleted returns the IncludeDeleted value.
func (m GetInsulinDoseParams) GetIncludeDeleted() bool {
	return m.IncludeDeleted
}

// SetIncludeDeleted sets the IncludeDeleted value and returns the updated GetInsulinDoseParams.
func (m *GetInsulinDoseParams) SetIncludeDeleted(v bool) {
	m.IncludeDeleted = v
}

// GetID returns the ID value.
func (m HardDeleteInsulinDoseParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated HardDeleteInsulinDoseParams.
func (m *HardDeleteInsulinDoseParams) SetID(v string) {
	m.ID = v
}

// GetDateGTE returns the DateGTE value.
func (m SearchInsulinDosesParams) GetDateGTE() *int64 {
	return m.DateGTE
}

// SetDateGTE sets the DateGTE value and returns the updated SearchInsulinDosesParams.
func (m *SearchInsulinDosesParams) SetDateGTE(v *int64) {
	m.DateGTE = v
}

// GetDateGT returns the DateGT value.
func (m SearchInsulinDosesParams) GetDateGT() *int64 {
	return m.DateGT
}

// SetDateGT sets the DateGT value and returns the updated SearchInsulinDosesParams.
func (m *SearchInsulinDosesParams) SetDateGT(v *int64) {
	m.DateGT = v
}

// GetDateEQ returns the DateEQ value.
func (m SearchInsulinDosesParams) GetDateEQ() *int64 {
	return m.DateEQ
}

// SetDateEQ sets the DateEQ value and returns the updated SearchInsulinDosesParams.
func (m *SearchInsulinDosesParams) SetDateEQ(v *int64) {
	m.DateEQ = v
}

// GetDateLTE returns the DateLTE value.
func (m SearchInsulinDosesParams) GetDateLTE() *int64 {
	return m.DateLTE
}

// SetDateLTE sets the DateLTE value and returns the updated SearchInsulinDosesParams.
func (m *SearchInsulinDosesParams) SetDateLTE(v *int64) {
	m.DateLTE = v
}

// GetDateLT returns the DateLT value.
func (m SearchInsulinDosesParams) GetDateLT() *int64 {
	return m.DateLT
}

// SetDateLT sets the DateLT value and returns the updated SearchInsulinDosesParams.
func (m *SearchInsulinDosesParams) SetDateLT(v *int64) {
	m.DateLT = v
}

// GetCreatedAtGTE returns the CreatedAtGTE value.
func (m SearchInsulinDosesParams) GetCreatedAtGTE() *int64 {
	return m.CreatedAtGTE
}

// SetCreatedAtGTE sets the CreatedAtGTE value and returns the updated SearchInsulinDosesParams.
func (m *SearchInsulinDosesParams) SetCreatedAtGTE(v *int64) {
	m.CreatedAtGTE = v
}

// GetCreatedAtLTE returns the CreatedAtLTE value.
func (m SearchInsulinDosesParams) GetCreatedAtLTE() *int64 {
	return m.CreatedAtLTE
}

// SetCreatedAtLTE sets the CreatedAtLTE value and returns the updated SearchInsulinDosesParams.
func (m *SearchInsulinDosesParams) SetCreatedAtLTE(v *int64) {
	m.CreatedAtLTE = v
}

// GetCreatedAtGT returns the CreatedAtGT value.
func (m SearchInsulinDosesParams) GetCreatedAtGT() *int64 {
	return m.CreatedAtGT
}

// SetCreatedAtGT sets the CreatedAtGT value and returns the updated SearchInsulinDosesParams.
func (m *SearchInsulinDosesParams) SetCreatedAtGT(v *int64) {
	m.CreatedAtGT = v
}

// GetCreatedAtLT returns the CreatedAtLT value.
func (m SearchInsulinDosesParams) GetCreatedAtLT() *int64 {
	return m.CreatedAtLT
}

// SetCreatedAtLT sets the CreatedAtLT value and returns the updated SearchInsulinDosesParams.
func (m *SearchInsulinDosesParams) SetCreatedAtLT(v *int64) {
	m.CreatedAtLT = v
}

// GetUpdatedGTE returns the UpdatedGTE value.
func (m SearchInsulinDosesParams) GetUpdatedGTE() *int64 {
	return m.UpdatedGTE
}

// SetUpdatedGTE sets the UpdatedGTE value and returns the updated SearchInsulinDosesParams.
func (m *SearchInsulinDosesParams) SetUpdatedGTE(v *int64) {
	m.UpdatedGTE = v
}

// GetUpdatedLTE returns the UpdatedLTE value.
func (m SearchInsulinDosesParams) GetUpdatedLTE() *int64 {
	return m.UpdatedLTE
}

// SetUpdatedLTE sets the UpdatedLTE value and returns the updated SearchInsulinDosesParams.
func (m *SearchInsulinDosesParams) SetUpdatedLTE(v *int64) {
	m.UpdatedLTE = v
}

// GetUpdatedGT returns the UpdatedGT value.
func (m SearchInsulinDosesParams) GetUpdatedGT() *int64 {
	return m.UpdatedGT
}

// SetUpdatedGT sets the UpdatedGT value and returns the updated SearchInsulinDosesParams.
func (m *SearchInsulinDosesParams) SetUpdatedGT(v *int64) {
	m.UpdatedGT = v
}

// GetUpdatedLT returns the UpdatedLT value.
func (m SearchInsulinDosesParams) GetUpdatedLT() *int64 {
	return m.UpdatedLT
}

// SetUpdatedLT sets the UpdatedLT value and returns the updated SearchInsulinDosesParams.
func (m *SearchInsulinDosesParams) SetUpdatedLT(v *int64) {
	m.UpdatedLT = v
}

// GetIDEq returns the IDEq value.
func (m SearchInsulinDosesParams) GetIDEq() *string {
	return m.IDEq
}

// SetIDEq sets the IDEq value and returns the updated SearchInsulinDosesParams.
func (m *SearchInsulinDosesParams) SetIDEq(v *string) {
	m.IDEq = v
}

// GetIDIn returns the IDIn value.
func (m SearchInsulinDosesParams) GetIDIn() *string {
	return m.IDIn
}

// SetIDIn sets the IDIn value and returns the updated SearchInsulinDosesParams.
func (m *SearchInsulinDosesParams) SetIDIn(v *string) {
	m.IDIn = v
}

// GetDeviceEq returns the DeviceEq value.
func (m SearchInsulinDosesParams) GetDeviceEq() *string {
	return m.DeviceEq
}

// SetDeviceEq sets the DeviceEq value and returns the updated SearchInsulinDosesParams.
func (m *SearchInsulinDosesParams) SetDeviceEq(v *string) {
	m.DeviceEq = v
}

// GetAppEq returns the AppEq value.
func (m SearchInsulinDosesParams) GetAppEq() *string {
	return m.AppEq
}

// SetAppEq sets the AppEq value and returns the updated SearchInsulinDosesParams.
func (m *SearchInsulinDosesParams) SetAppEq(v *string) {
	m.AppEq = v
}

// GetTypeEq returns the TypeEq value.
func (m SearchInsulinDosesParams) GetTypeEq() *string {
	return m.TypeEq
}

// SetTypeEq sets the TypeEq value and returns the updated SearchInsulinDosesParams.
func (m *SearchInsulinDosesParams) SetTypeEq(v *string) {
	m.TypeEq = v
}

// GetTypeNeq returns the TypeNeq value.
func (m SearchInsulinDosesParams) GetTypeNeq() *string {
	return m.TypeNeq
}

// SetTypeNeq sets the TypeNeq value and returns the updated SearchInsulinDosesParams.
func (m *SearchInsulinDosesParams) SetTypeNeq(v *string) {
	m.TypeNeq = v
}

// GetTypeIn returns the TypeIn value.
func (m SearchInsulinDosesParams) GetTypeIn() *string {
	return m.TypeIn
}

// SetTypeIn sets the TypeIn value and returns the updated SearchInsulinDosesParams.
func (m *SearchInsulinDosesParams) SetTypeIn(v *string) {
	m.TypeIn = v
}

// GetTypeNin returns the TypeNin value.
func (m SearchInsulinDosesParams) GetTypeNin() *string {
	return m.TypeNin
}

// SetTypeNin sets the TypeNin value and returns the updated SearchInsulinDosesParams.
func (m *SearchInsulinDosesParams) SetTypeNin(v *string) {
	m.TypeNin = v
}

// GetInsulinEQ returns the InsulinEQ value.
func (m SearchInsulinDosesParams) GetInsulinEQ() *int64 {
	return m.InsulinEQ
}

// SetInsulinEQ sets the InsulinEQ value and returns the updated SearchInsulinDosesParams.
func (m *SearchInsulinDosesParams) SetInsulinEQ(v *int64) {
	m.InsulinEQ = v
}

// GetInsulinGTE returns the InsulinGTE value.
func (m SearchInsulinDosesParams) GetInsulinGTE() *int64 {
	return m.InsulinGTE
}

// SetInsulinGTE sets the InsulinGTE value and returns the updated SearchInsulinDosesParams.
func (m *SearchInsulinDosesParams) SetInsulinGTE(v *int64) {
	m.InsulinGTE = v
}

// GetInsulinGT returns the InsulinGT value.
func (m SearchInsulinDosesParams) GetInsulinGT() *int64 {
	return m.InsulinGT
}

// SetInsulinGT sets the InsulinGT value and returns the updated SearchInsulinDosesParams.
func (m *SearchInsulinDosesParams) SetInsulinGT(v *int64) {
	m.InsulinGT = v
}

// GetInsulinLTE returns the InsulinLTE value.
func (m SearchInsulinDosesParams) GetInsulinLTE() *int64 {
	return m.InsulinLTE
}

// SetInsulinLTE sets the InsulinLTE value and returns the updated SearchInsulinDosesParams.
func (m *SearchInsulinDosesParams) SetInsulinLTE(v *int64) {
	m.InsulinLTE = v
}

// GetInsulinLT returns the InsulinLT value.
func (m SearchInsulinDosesParams) GetInsulinLT() *int64 {
	return m.InsulinLT
}

// SetInsulinLT sets the InsulinLT value and returns the updated SearchInsulinDosesParams.
func (m *SearchInsulinDosesParams) SetInsulinLT(v *int64) {
	m.InsulinLT = v
}

// GetIncludeDeleted returns the IncludeDeleted value.
func (m SearchInsulinDosesParams) GetIncludeDeleted() *bool {
	return m.IncludeDeleted
}

// SetIncludeDeleted sets the IncludeDeleted value and returns the updated SearchInsulinDosesParams.
func (m *SearchInsulinDosesParams) SetIncludeDeleted(v *bool) {
	m.IncludeDeleted = v
}

// GetSkip returns the Skip value.
func (m SearchInsulinDosesParams) GetSkip() int64 {
	return m.Skip
}

// SetSkip sets the Skip value and returns the updated SearchInsulinDosesParams.
func (m *SearchInsulinDosesParams) SetSkip(v int64) {
	m.Skip = v
}

// GetLimit returns the Limit value.
func (m SearchInsulinDosesParams) GetLimit() int64 {
	return m.Limit
}

// SetLimit sets the Limit value and returns the updated SearchInsulinDosesParams.
func (m *SearchInsulinDosesParams) SetLimit(v int64) {
	m.Limit = v
}

// GetUpdatedByIdentityName returns the UpdatedByIdentityName value.
func (m SoftDeleteInsulinDoseParams) GetUpdatedByIdentityName() *string {
	return m.UpdatedByIdentityName
}

// SetUpdatedByIdentityName sets the UpdatedByIdentityName value and returns the updated SoftDeleteInsulinDoseParams.
func (m *SoftDeleteInsulinDoseParams) SetUpdatedByIdentityName(v *string) {
	m.UpdatedByIdentityName = v
}

// GetID returns the ID value.
func (m SoftDeleteInsulinDoseParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated SoftDeleteInsulinDoseParams.
func (m *SoftDeleteInsulinDoseParams) SetID(v string) {
	m.ID = v
}

// GetDoseType returns the DoseType value.
func (m UpdateInsulinDoseParams) GetDoseType() *string {
	return m.DoseType
}

// SetDoseType sets the DoseType value and returns the updated UpdateInsulinDoseParams.
func (m *UpdateInsulinDoseParams) SetDoseType(v *string) {
	m.DoseType = v
}

// GetDose returns the Dose value.
func (m UpdateInsulinDoseParams) GetDose() *float64 {
	return m.Dose
}

// SetDose sets the Dose value and returns the updated UpdateInsulinDoseParams.
func (m *UpdateInsulinDoseParams) SetDose(v *float64) {
	m.Dose = v
}

// GetSplitNow returns the SplitNow value.
func (m UpdateInsulinDoseParams) GetSplitNow() *float64 {
	return m.SplitNow
}

// SetSplitNow sets the SplitNow value and returns the updated UpdateInsulinDoseParams.
func (m *UpdateInsulinDoseParams) SetSplitNow(v *float64) {
	m.SplitNow = v
}

// GetSplitExt returns the SplitExt value.
func (m UpdateInsulinDoseParams) GetSplitExt() *float64 {
	return m.SplitExt
}

// SetSplitExt sets the SplitExt value and returns the updated UpdateInsulinDoseParams.
func (m *UpdateInsulinDoseParams) SetSplitExt(v *float64) {
	m.SplitExt = v
}

// GetNotes returns the Notes value.
func (m UpdateInsulinDoseParams) GetNotes() *string {
	return m.Notes
}

// SetNotes sets the Notes value and returns the updated UpdateInsulinDoseParams.
func (m *UpdateInsulinDoseParams) SetNotes(v *string) {
	m.Notes = v
}

// GetReason returns the Reason value.
func (m UpdateInsulinDoseParams) GetReason() *string {
	return m.Reason
}

// SetReason sets the Reason value and returns the updated UpdateInsulinDoseParams.
func (m *UpdateInsulinDoseParams) SetReason(v *string) {
	m.Reason = v
}

// GetRawJsonData returns the RawJsonData value.
func (m UpdateInsulinDoseParams) GetRawJsonData() *string {
	return m.RawJsonData
}

// SetRawJsonData sets the RawJsonData value and returns the updated UpdateInsulinDoseParams.
func (m *UpdateInsulinDoseParams) SetRawJsonData(v *string) {
	m.RawJsonData = v
}

// GetUpdatedByIdentityID returns the UpdatedByIdentityID value.
func (m UpdateInsulinDoseParams) GetUpdatedByIdentityID() *int64 {
	return m.UpdatedByIdentityID
}

// SetUpdatedByIdentityID sets the UpdatedByIdentityID value and returns the updated UpdateInsulinDoseParams.
func (m *UpdateInsulinDoseParams) SetUpdatedByIdentityID(v *int64) {
	m.UpdatedByIdentityID = v
}

// GetUpdatedByIdentityName returns the UpdatedByIdentityName value.
func (m UpdateInsulinDoseParams) GetUpdatedByIdentityName() *string {
	return m.UpdatedByIdentityName
}

// SetUpdatedByIdentityName sets the UpdatedByIdentityName value and returns the updated UpdateInsulinDoseParams.
func (m *UpdateInsulinDoseParams) SetUpdatedByIdentityName(v *string) {
	m.UpdatedByIdentityName = v
}

// GetID returns the ID value.
func (m UpdateInsulinDoseParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated UpdateInsulinDoseParams.
func (m *UpdateInsulinDoseParams) SetID(v string) {
	m.ID = v
}

// GetIncludeDeleted returns the IncludeDeleted value.
func (m AllNotificationEventsParams) GetIncludeDeleted() bool {
	return m.IncludeDeleted
}

// SetIncludeDeleted sets the IncludeDeleted value and returns the updated AllNotificationEventsParams.
func (m *AllNotificationEventsParams) SetIncludeDeleted(v bool) {
	m.IncludeDeleted = v
}

// GetSkip returns the Skip value.
func (m AllNotificationEventsParams) GetSkip() int64 {
	return m.Skip
}

// SetSkip sets the Skip value and returns the updated AllNotificationEventsParams.
func (m *AllNotificationEventsParams) SetSkip(v int64) {
	m.Skip = v
}

// GetLimit returns the Limit value.
func (m AllNotificationEventsParams) GetLimit() int64 {
	return m.Limit
}

// SetLimit sets the Limit value and returns the updated AllNotificationEventsParams.
func (m *AllNotificationEventsParams) SetLimit(v int64) {
	m.Limit = v
}

// GetID returns the ID value.
func (m CreateNotificationEventParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated CreateNotificationEventParams.
func (m *CreateNotificationEventParams) SetID(v string) {
	m.ID = v
}

// GetTitle returns the Title value.
func (m CreateNotificationEventParams) GetTitle() string {
	return m.Title
}

// SetTitle sets the Title value and returns the updated CreateNotificationEventParams.
func (m *CreateNotificationEventParams) SetTitle(v string) {
	m.Title = v
}

// GetSeverity returns the Severity value.
func (m CreateNotificationEventParams) GetSeverity() int64 {
	return m.Severity
}

// SetSeverity sets the Severity value and returns the updated CreateNotificationEventParams.
func (m *CreateNotificationEventParams) SetSeverity(v int64) {
	m.Severity = v
}

// GetContent returns the Content value.
func (m CreateNotificationEventParams) GetContent() *string {
	return m.Content
}

// SetContent sets the Content value and returns the updated CreateNotificationEventParams.
func (m *CreateNotificationEventParams) SetContent(v *string) {
	m.Content = v
}

// GetNotificationType returns the NotificationType value.
func (m CreateNotificationEventParams) GetNotificationType() string {
	return m.NotificationType
}

// SetNotificationType sets the NotificationType value and returns the updated CreateNotificationEventParams.
func (m *CreateNotificationEventParams) SetNotificationType(v string) {
	m.NotificationType = v
}

// GetNotes returns the Notes value.
func (m CreateNotificationEventParams) GetNotes() *string {
	return m.Notes
}

// SetNotes sets the Notes value and returns the updated CreateNotificationEventParams.
func (m *CreateNotificationEventParams) SetNotes(v *string) {
	m.Notes = v
}

// GetReason returns the Reason value.
func (m CreateNotificationEventParams) GetReason() *string {
	return m.Reason
}

// SetReason sets the Reason value and returns the updated CreateNotificationEventParams.
func (m *CreateNotificationEventParams) SetReason(v *string) {
	m.Reason = v
}

// GetUtcOffsetMinutes returns the UtcOffsetMinutes value.
func (m CreateNotificationEventParams) GetUtcOffsetMinutes() *int64 {
	return m.UtcOffsetMinutes
}

// SetUtcOffsetMinutes sets the UtcOffsetMinutes value and returns the updated CreateNotificationEventParams.
func (m *CreateNotificationEventParams) SetUtcOffsetMinutes(v *int64) {
	m.UtcOffsetMinutes = v
}

// GetOccurredAt returns the OccurredAt value.
func (m CreateNotificationEventParams) GetOccurredAt() int64 {
	return m.OccurredAt
}

// SetOccurredAt sets the OccurredAt value and returns the updated CreateNotificationEventParams.
func (m *CreateNotificationEventParams) SetOccurredAt(v int64) {
	m.OccurredAt = v
}

// GetRawJsonData returns the RawJsonData value.
func (m CreateNotificationEventParams) GetRawJsonData() *string {
	return m.RawJsonData
}

// SetRawJsonData sets the RawJsonData value and returns the updated CreateNotificationEventParams.
func (m *CreateNotificationEventParams) SetRawJsonData(v *string) {
	m.RawJsonData = v
}

// GetCreatedByAppName returns the CreatedByAppName value.
func (m CreateNotificationEventParams) GetCreatedByAppName() *string {
	return m.CreatedByAppName
}

// SetCreatedByAppName sets the CreatedByAppName value and returns the updated CreateNotificationEventParams.
func (m *CreateNotificationEventParams) SetCreatedByAppName(v *string) {
	m.CreatedByAppName = v
}

// GetCreatedByDeviceName returns the CreatedByDeviceName value.
func (m CreateNotificationEventParams) GetCreatedByDeviceName() *string {
	return m.CreatedByDeviceName
}

// SetCreatedByDeviceName sets the CreatedByDeviceName value and returns the updated CreateNotificationEventParams.
func (m *CreateNotificationEventParams) SetCreatedByDeviceName(v *string) {
	m.CreatedByDeviceName = v
}

// GetCreatedByIdentityName returns the CreatedByIdentityName value.
func (m CreateNotificationEventParams) GetCreatedByIdentityName() string {
	return m.CreatedByIdentityName
}

// SetCreatedByIdentityName sets the CreatedByIdentityName value and returns the updated CreateNotificationEventParams.
func (m *CreateNotificationEventParams) SetCreatedByIdentityName(v string) {
	m.CreatedByIdentityName = v
}

// GetCreatedByIdentityID returns the CreatedByIdentityID value.
func (m CreateNotificationEventParams) GetCreatedByIdentityID() *int64 {
	return m.CreatedByIdentityID
}

// SetCreatedByIdentityID sets the CreatedByIdentityID value and returns the updated CreateNotificationEventParams.
func (m *CreateNotificationEventParams) SetCreatedByIdentityID(v *int64) {
	m.CreatedByIdentityID = v
}

// GetID returns the ID value.
func (m GetNotificationEventParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated GetNotificationEventParams.
func (m *GetNotificationEventParams) SetID(v string) {
	m.ID = v
}

// GetIncludeDeleted returns the IncludeDeleted value.
func (m GetNotificationEventParams) GetIncludeDeleted() bool {
	return m.IncludeDeleted
}

// SetIncludeDeleted sets the IncludeDeleted value and returns the updated GetNotificationEventParams.
func (m *GetNotificationEventParams) SetIncludeDeleted(v bool) {
	m.IncludeDeleted = v
}

// GetID returns the ID value.
func (m HardDeleteNotificationEventParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated HardDeleteNotificationEventParams.
func (m *HardDeleteNotificationEventParams) SetID(v string) {
	m.ID = v
}

// GetDateGTE returns the DateGTE value.
func (m SearchNotificationEventsParams) GetDateGTE() *int64 {
	return m.DateGTE
}

// SetDateGTE sets the DateGTE value and returns the updated SearchNotificationEventsParams.
func (m *SearchNotificationEventsParams) SetDateGTE(v *int64) {
	m.DateGTE = v
}

// GetDateGT returns the DateGT value.
func (m SearchNotificationEventsParams) GetDateGT() *int64 {
	return m.DateGT
}

// SetDateGT sets the DateGT value and returns the updated SearchNotificationEventsParams.
func (m *SearchNotificationEventsParams) SetDateGT(v *int64) {
	m.DateGT = v
}

// GetDateEQ returns the DateEQ value.
func (m SearchNotificationEventsParams) GetDateEQ() *int64 {
	return m.DateEQ
}

// SetDateEQ sets the DateEQ value and returns the updated SearchNotificationEventsParams.
func (m *SearchNotificationEventsParams) SetDateEQ(v *int64) {
	m.DateEQ = v
}

// GetDateLTE returns the DateLTE value.
func (m SearchNotificationEventsParams) GetDateLTE() *int64 {
	return m.DateLTE
}

// SetDateLTE sets the DateLTE value and returns the updated SearchNotificationEventsParams.
func (m *SearchNotificationEventsParams) SetDateLTE(v *int64) {
	m.DateLTE = v
}

// GetDateLT returns the DateLT value.
func (m SearchNotificationEventsParams) GetDateLT() *int64 {
	return m.DateLT
}

// SetDateLT sets the DateLT value and returns the updated SearchNotificationEventsParams.
func (m *SearchNotificationEventsParams) SetDateLT(v *int64) {
	m.DateLT = v
}

// GetCreatedAtGTE returns the CreatedAtGTE value.
func (m SearchNotificationEventsParams) GetCreatedAtGTE() *int64 {
	return m.CreatedAtGTE
}

// SetCreatedAtGTE sets the CreatedAtGTE value and returns the updated SearchNotificationEventsParams.
func (m *SearchNotificationEventsParams) SetCreatedAtGTE(v *int64) {
	m.CreatedAtGTE = v
}

// GetCreatedAtLTE returns the CreatedAtLTE value.
func (m SearchNotificationEventsParams) GetCreatedAtLTE() *int64 {
	return m.CreatedAtLTE
}

// SetCreatedAtLTE sets the CreatedAtLTE value and returns the updated SearchNotificationEventsParams.
func (m *SearchNotificationEventsParams) SetCreatedAtLTE(v *int64) {
	m.CreatedAtLTE = v
}

// GetCreatedAtGT returns the CreatedAtGT value.
func (m SearchNotificationEventsParams) GetCreatedAtGT() *int64 {
	return m.CreatedAtGT
}

// SetCreatedAtGT sets the CreatedAtGT value and returns the updated SearchNotificationEventsParams.
func (m *SearchNotificationEventsParams) SetCreatedAtGT(v *int64) {
	m.CreatedAtGT = v
}

// GetCreatedAtLT returns the CreatedAtLT value.
func (m SearchNotificationEventsParams) GetCreatedAtLT() *int64 {
	return m.CreatedAtLT
}

// SetCreatedAtLT sets the CreatedAtLT value and returns the updated SearchNotificationEventsParams.
func (m *SearchNotificationEventsParams) SetCreatedAtLT(v *int64) {
	m.CreatedAtLT = v
}

// GetUpdatedGTE returns the UpdatedGTE value.
func (m SearchNotificationEventsParams) GetUpdatedGTE() *int64 {
	return m.UpdatedGTE
}

// SetUpdatedGTE sets the UpdatedGTE value and returns the updated SearchNotificationEventsParams.
func (m *SearchNotificationEventsParams) SetUpdatedGTE(v *int64) {
	m.UpdatedGTE = v
}

// GetUpdatedLTE returns the UpdatedLTE value.
func (m SearchNotificationEventsParams) GetUpdatedLTE() *int64 {
	return m.UpdatedLTE
}

// SetUpdatedLTE sets the UpdatedLTE value and returns the updated SearchNotificationEventsParams.
func (m *SearchNotificationEventsParams) SetUpdatedLTE(v *int64) {
	m.UpdatedLTE = v
}

// GetUpdatedGT returns the UpdatedGT value.
func (m SearchNotificationEventsParams) GetUpdatedGT() *int64 {
	return m.UpdatedGT
}

// SetUpdatedGT sets the UpdatedGT value and returns the updated SearchNotificationEventsParams.
func (m *SearchNotificationEventsParams) SetUpdatedGT(v *int64) {
	m.UpdatedGT = v
}

// GetUpdatedLT returns the UpdatedLT value.
func (m SearchNotificationEventsParams) GetUpdatedLT() *int64 {
	return m.UpdatedLT
}

// SetUpdatedLT sets the UpdatedLT value and returns the updated SearchNotificationEventsParams.
func (m *SearchNotificationEventsParams) SetUpdatedLT(v *int64) {
	m.UpdatedLT = v
}

// GetIDEq returns the IDEq value.
func (m SearchNotificationEventsParams) GetIDEq() *string {
	return m.IDEq
}

// SetIDEq sets the IDEq value and returns the updated SearchNotificationEventsParams.
func (m *SearchNotificationEventsParams) SetIDEq(v *string) {
	m.IDEq = v
}

// GetIDIn returns the IDIn value.
func (m SearchNotificationEventsParams) GetIDIn() *string {
	return m.IDIn
}

// SetIDIn sets the IDIn value and returns the updated SearchNotificationEventsParams.
func (m *SearchNotificationEventsParams) SetIDIn(v *string) {
	m.IDIn = v
}

// GetTitleEq returns the TitleEq value.
func (m SearchNotificationEventsParams) GetTitleEq() *string {
	return m.TitleEq
}

// SetTitleEq sets the TitleEq value and returns the updated SearchNotificationEventsParams.
func (m *SearchNotificationEventsParams) SetTitleEq(v *string) {
	m.TitleEq = v
}

// GetDeviceEq returns the DeviceEq value.
func (m SearchNotificationEventsParams) GetDeviceEq() *string {
	return m.DeviceEq
}

// SetDeviceEq sets the DeviceEq value and returns the updated SearchNotificationEventsParams.
func (m *SearchNotificationEventsParams) SetDeviceEq(v *string) {
	m.DeviceEq = v
}

// GetAppEq returns the AppEq value.
func (m SearchNotificationEventsParams) GetAppEq() *string {
	return m.AppEq
}

// SetAppEq sets the AppEq value and returns the updated SearchNotificationEventsParams.
func (m *SearchNotificationEventsParams) SetAppEq(v *string) {
	m.AppEq = v
}

// GetIncludeDeleted returns the IncludeDeleted value.
func (m SearchNotificationEventsParams) GetIncludeDeleted() *bool {
	return m.IncludeDeleted
}

// SetIncludeDeleted sets the IncludeDeleted value and returns the updated SearchNotificationEventsParams.
func (m *SearchNotificationEventsParams) SetIncludeDeleted(v *bool) {
	m.IncludeDeleted = v
}

// GetSkip returns the Skip value.
func (m SearchNotificationEventsParams) GetSkip() int64 {
	return m.Skip
}

// SetSkip sets the Skip value and returns the updated SearchNotificationEventsParams.
func (m *SearchNotificationEventsParams) SetSkip(v int64) {
	m.Skip = v
}

// GetLimit returns the Limit value.
func (m SearchNotificationEventsParams) GetLimit() int64 {
	return m.Limit
}

// SetLimit sets the Limit value and returns the updated SearchNotificationEventsParams.
func (m *SearchNotificationEventsParams) SetLimit(v int64) {
	m.Limit = v
}

// GetUpdatedByIdentityName returns the UpdatedByIdentityName value.
func (m SoftDeleteNotificationEventParams) GetUpdatedByIdentityName() *string {
	return m.UpdatedByIdentityName
}

// SetUpdatedByIdentityName sets the UpdatedByIdentityName value and returns the updated SoftDeleteNotificationEventParams.
func (m *SoftDeleteNotificationEventParams) SetUpdatedByIdentityName(v *string) {
	m.UpdatedByIdentityName = v
}

// GetID returns the ID value.
func (m SoftDeleteNotificationEventParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated SoftDeleteNotificationEventParams.
func (m *SoftDeleteNotificationEventParams) SetID(v string) {
	m.ID = v
}

// GetTitle returns the Title value.
func (m UpdateNotificationEventParams) GetTitle() *string {
	return m.Title
}

// SetTitle sets the Title value and returns the updated UpdateNotificationEventParams.
func (m *UpdateNotificationEventParams) SetTitle(v *string) {
	m.Title = v
}

// GetSeverity returns the Severity value.
func (m UpdateNotificationEventParams) GetSeverity() *int64 {
	return m.Severity
}

// SetSeverity sets the Severity value and returns the updated UpdateNotificationEventParams.
func (m *UpdateNotificationEventParams) SetSeverity(v *int64) {
	m.Severity = v
}

// GetNotificationType returns the NotificationType value.
func (m UpdateNotificationEventParams) GetNotificationType() *string {
	return m.NotificationType
}

// SetNotificationType sets the NotificationType value and returns the updated UpdateNotificationEventParams.
func (m *UpdateNotificationEventParams) SetNotificationType(v *string) {
	m.NotificationType = v
}

// GetContent returns the Content value.
func (m UpdateNotificationEventParams) GetContent() *string {
	return m.Content
}

// SetContent sets the Content value and returns the updated UpdateNotificationEventParams.
func (m *UpdateNotificationEventParams) SetContent(v *string) {
	m.Content = v
}

// GetNotes returns the Notes value.
func (m UpdateNotificationEventParams) GetNotes() *string {
	return m.Notes
}

// SetNotes sets the Notes value and returns the updated UpdateNotificationEventParams.
func (m *UpdateNotificationEventParams) SetNotes(v *string) {
	m.Notes = v
}

// GetReason returns the Reason value.
func (m UpdateNotificationEventParams) GetReason() *string {
	return m.Reason
}

// SetReason sets the Reason value and returns the updated UpdateNotificationEventParams.
func (m *UpdateNotificationEventParams) SetReason(v *string) {
	m.Reason = v
}

// GetRawJsonData returns the RawJsonData value.
func (m UpdateNotificationEventParams) GetRawJsonData() *string {
	return m.RawJsonData
}

// SetRawJsonData sets the RawJsonData value and returns the updated UpdateNotificationEventParams.
func (m *UpdateNotificationEventParams) SetRawJsonData(v *string) {
	m.RawJsonData = v
}

// GetUpdatedByIdentityID returns the UpdatedByIdentityID value.
func (m UpdateNotificationEventParams) GetUpdatedByIdentityID() *int64 {
	return m.UpdatedByIdentityID
}

// SetUpdatedByIdentityID sets the UpdatedByIdentityID value and returns the updated UpdateNotificationEventParams.
func (m *UpdateNotificationEventParams) SetUpdatedByIdentityID(v *int64) {
	m.UpdatedByIdentityID = v
}

// GetUpdatedByIdentityName returns the UpdatedByIdentityName value.
func (m UpdateNotificationEventParams) GetUpdatedByIdentityName() *string {
	return m.UpdatedByIdentityName
}

// SetUpdatedByIdentityName sets the UpdatedByIdentityName value and returns the updated UpdateNotificationEventParams.
func (m *UpdateNotificationEventParams) SetUpdatedByIdentityName(v *string) {
	m.UpdatedByIdentityName = v
}

// GetID returns the ID value.
func (m UpdateNotificationEventParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated UpdateNotificationEventParams.
func (m *UpdateNotificationEventParams) SetID(v string) {
	m.ID = v
}

// GetIncludeDeleted returns the IncludeDeleted value.
func (m AllProfileSwitchesParams) GetIncludeDeleted() bool {
	return m.IncludeDeleted
}

// SetIncludeDeleted sets the IncludeDeleted value and returns the updated AllProfileSwitchesParams.
func (m *AllProfileSwitchesParams) SetIncludeDeleted(v bool) {
	m.IncludeDeleted = v
}

// GetSkip returns the Skip value.
func (m AllProfileSwitchesParams) GetSkip() int64 {
	return m.Skip
}

// SetSkip sets the Skip value and returns the updated AllProfileSwitchesParams.
func (m *AllProfileSwitchesParams) SetSkip(v int64) {
	m.Skip = v
}

// GetLimit returns the Limit value.
func (m AllProfileSwitchesParams) GetLimit() int64 {
	return m.Limit
}

// SetLimit sets the Limit value and returns the updated AllProfileSwitchesParams.
func (m *AllProfileSwitchesParams) SetLimit(v int64) {
	m.Limit = v
}

// GetID returns the ID value.
func (m CreateProfileSwitchParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated CreateProfileSwitchParams.
func (m *CreateProfileSwitchParams) SetID(v string) {
	m.ID = v
}

// GetName returns the Name value.
func (m CreateProfileSwitchParams) GetName() string {
	return m.Name
}

// SetName sets the Name value and returns the updated CreateProfileSwitchParams.
func (m *CreateProfileSwitchParams) SetName(v string) {
	m.Name = v
}

// GetDurationMinutes returns the DurationMinutes value.
func (m CreateProfileSwitchParams) GetDurationMinutes() int64 {
	return m.DurationMinutes
}

// SetDurationMinutes sets the DurationMinutes value and returns the updated CreateProfileSwitchParams.
func (m *CreateProfileSwitchParams) SetDurationMinutes(v int64) {
	m.DurationMinutes = v
}

// GetPercent returns the Percent value.
func (m CreateProfileSwitchParams) GetPercent() *float64 {
	return m.Percent
}

// SetPercent sets the Percent value and returns the updated CreateProfileSwitchParams.
func (m *CreateProfileSwitchParams) SetPercent(v *float64) {
	m.Percent = v
}

// GetDia returns the Dia value.
func (m CreateProfileSwitchParams) GetDia() *float64 {
	return m.Dia
}

// SetDia sets the Dia value and returns the updated CreateProfileSwitchParams.
func (m *CreateProfileSwitchParams) SetDia(v *float64) {
	m.Dia = v
}

// GetUnits returns the Units value.
func (m CreateProfileSwitchParams) GetUnits() *consts.Units {
	return m.Units
}

// SetUnits sets the Units value and returns the updated CreateProfileSwitchParams.
func (m *CreateProfileSwitchParams) SetUnits(v *consts.Units) {
	m.Units = v
}

// GetTimezone returns the Timezone value.
func (m CreateProfileSwitchParams) GetTimezone() *string {
	return m.Timezone
}

// SetTimezone sets the Timezone value and returns the updated CreateProfileSwitchParams.
func (m *CreateProfileSwitchParams) SetTimezone(v *string) {
	m.Timezone = v
}

// GetCarbsHr returns the CarbsHr value.
func (m CreateProfileSwitchParams) GetCarbsHr() *float64 {
	return m.CarbsHr
}

// SetCarbsHr sets the CarbsHr value and returns the updated CreateProfileSwitchParams.
func (m *CreateProfileSwitchParams) SetCarbsHr(v *float64) {
	m.CarbsHr = v
}

// GetDelay returns the Delay value.
func (m CreateProfileSwitchParams) GetDelay() *float64 {
	return m.Delay
}

// SetDelay sets the Delay value and returns the updated CreateProfileSwitchParams.
func (m *CreateProfileSwitchParams) SetDelay(v *float64) {
	m.Delay = v
}

// GetTimeshift returns the Timeshift value.
func (m CreateProfileSwitchParams) GetTimeshift() *int64 {
	return m.Timeshift
}

// SetTimeshift sets the Timeshift value and returns the updated CreateProfileSwitchParams.
func (m *CreateProfileSwitchParams) SetTimeshift(v *int64) {
	m.Timeshift = v
}

// GetCarbRatiosJson returns the CarbRatiosJson value.
func (m CreateProfileSwitchParams) GetCarbRatiosJson() *string {
	return m.CarbRatiosJson
}

// SetCarbRatiosJson sets the CarbRatiosJson value and returns the updated CreateProfileSwitchParams.
func (m *CreateProfileSwitchParams) SetCarbRatiosJson(v *string) {
	m.CarbRatiosJson = v
}

// GetInsulinSensitivitiesJson returns the InsulinSensitivitiesJson value.
func (m CreateProfileSwitchParams) GetInsulinSensitivitiesJson() *string {
	return m.InsulinSensitivitiesJson
}

// SetInsulinSensitivitiesJson sets the InsulinSensitivitiesJson value and returns the updated CreateProfileSwitchParams.
func (m *CreateProfileSwitchParams) SetInsulinSensitivitiesJson(v *string) {
	m.InsulinSensitivitiesJson = v
}

// GetBasalRatesJson returns the BasalRatesJson value.
func (m CreateProfileSwitchParams) GetBasalRatesJson() *string {
	return m.BasalRatesJson
}

// SetBasalRatesJson sets the BasalRatesJson value and returns the updated CreateProfileSwitchParams.
func (m *CreateProfileSwitchParams) SetBasalRatesJson(v *string) {
	m.BasalRatesJson = v
}

// GetTargetsHighJson returns the TargetsHighJson value.
func (m CreateProfileSwitchParams) GetTargetsHighJson() *string {
	return m.TargetsHighJson
}

// SetTargetsHighJson sets the TargetsHighJson value and returns the updated CreateProfileSwitchParams.
func (m *CreateProfileSwitchParams) SetTargetsHighJson(v *string) {
	m.TargetsHighJson = v
}

// GetTargetsLowJson returns the TargetsLowJson value.
func (m CreateProfileSwitchParams) GetTargetsLowJson() *string {
	return m.TargetsLowJson
}

// SetTargetsLowJson sets the TargetsLowJson value and returns the updated CreateProfileSwitchParams.
func (m *CreateProfileSwitchParams) SetTargetsLowJson(v *string) {
	m.TargetsLowJson = v
}

// GetNotes returns the Notes value.
func (m CreateProfileSwitchParams) GetNotes() *string {
	return m.Notes
}

// SetNotes sets the Notes value and returns the updated CreateProfileSwitchParams.
func (m *CreateProfileSwitchParams) SetNotes(v *string) {
	m.Notes = v
}

// GetReason returns the Reason value.
func (m CreateProfileSwitchParams) GetReason() *string {
	return m.Reason
}

// SetReason sets the Reason value and returns the updated CreateProfileSwitchParams.
func (m *CreateProfileSwitchParams) SetReason(v *string) {
	m.Reason = v
}

// GetUtcOffsetMinutes returns the UtcOffsetMinutes value.
func (m CreateProfileSwitchParams) GetUtcOffsetMinutes() *int64 {
	return m.UtcOffsetMinutes
}

// SetUtcOffsetMinutes sets the UtcOffsetMinutes value and returns the updated CreateProfileSwitchParams.
func (m *CreateProfileSwitchParams) SetUtcOffsetMinutes(v *int64) {
	m.UtcOffsetMinutes = v
}

// GetOccurredAt returns the OccurredAt value.
func (m CreateProfileSwitchParams) GetOccurredAt() int64 {
	return m.OccurredAt
}

// SetOccurredAt sets the OccurredAt value and returns the updated CreateProfileSwitchParams.
func (m *CreateProfileSwitchParams) SetOccurredAt(v int64) {
	m.OccurredAt = v
}

// GetRawJsonData returns the RawJsonData value.
func (m CreateProfileSwitchParams) GetRawJsonData() *string {
	return m.RawJsonData
}

// SetRawJsonData sets the RawJsonData value and returns the updated CreateProfileSwitchParams.
func (m *CreateProfileSwitchParams) SetRawJsonData(v *string) {
	m.RawJsonData = v
}

// GetCreatedByAppName returns the CreatedByAppName value.
func (m CreateProfileSwitchParams) GetCreatedByAppName() *string {
	return m.CreatedByAppName
}

// SetCreatedByAppName sets the CreatedByAppName value and returns the updated CreateProfileSwitchParams.
func (m *CreateProfileSwitchParams) SetCreatedByAppName(v *string) {
	m.CreatedByAppName = v
}

// GetCreatedByDeviceName returns the CreatedByDeviceName value.
func (m CreateProfileSwitchParams) GetCreatedByDeviceName() *string {
	return m.CreatedByDeviceName
}

// SetCreatedByDeviceName sets the CreatedByDeviceName value and returns the updated CreateProfileSwitchParams.
func (m *CreateProfileSwitchParams) SetCreatedByDeviceName(v *string) {
	m.CreatedByDeviceName = v
}

// GetCreatedByIdentityName returns the CreatedByIdentityName value.
func (m CreateProfileSwitchParams) GetCreatedByIdentityName() string {
	return m.CreatedByIdentityName
}

// SetCreatedByIdentityName sets the CreatedByIdentityName value and returns the updated CreateProfileSwitchParams.
func (m *CreateProfileSwitchParams) SetCreatedByIdentityName(v string) {
	m.CreatedByIdentityName = v
}

// GetCreatedByIdentityID returns the CreatedByIdentityID value.
func (m CreateProfileSwitchParams) GetCreatedByIdentityID() *int64 {
	return m.CreatedByIdentityID
}

// SetCreatedByIdentityID sets the CreatedByIdentityID value and returns the updated CreateProfileSwitchParams.
func (m *CreateProfileSwitchParams) SetCreatedByIdentityID(v *int64) {
	m.CreatedByIdentityID = v
}

// GetID returns the ID value.
func (m GetProfileSwitchParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated GetProfileSwitchParams.
func (m *GetProfileSwitchParams) SetID(v string) {
	m.ID = v
}

// GetIncludeDeleted returns the IncludeDeleted value.
func (m GetProfileSwitchParams) GetIncludeDeleted() bool {
	return m.IncludeDeleted
}

// SetIncludeDeleted sets the IncludeDeleted value and returns the updated GetProfileSwitchParams.
func (m *GetProfileSwitchParams) SetIncludeDeleted(v bool) {
	m.IncludeDeleted = v
}

// GetID returns the ID value.
func (m HardDeleteProfileSwitchParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated HardDeleteProfileSwitchParams.
func (m *HardDeleteProfileSwitchParams) SetID(v string) {
	m.ID = v
}

// GetDateGTE returns the DateGTE value.
func (m SearchProfileSwitchesParams) GetDateGTE() *int64 {
	return m.DateGTE
}

// SetDateGTE sets the DateGTE value and returns the updated SearchProfileSwitchesParams.
func (m *SearchProfileSwitchesParams) SetDateGTE(v *int64) {
	m.DateGTE = v
}

// GetDateGT returns the DateGT value.
func (m SearchProfileSwitchesParams) GetDateGT() *int64 {
	return m.DateGT
}

// SetDateGT sets the DateGT value and returns the updated SearchProfileSwitchesParams.
func (m *SearchProfileSwitchesParams) SetDateGT(v *int64) {
	m.DateGT = v
}

// GetDateEQ returns the DateEQ value.
func (m SearchProfileSwitchesParams) GetDateEQ() *int64 {
	return m.DateEQ
}

// SetDateEQ sets the DateEQ value and returns the updated SearchProfileSwitchesParams.
func (m *SearchProfileSwitchesParams) SetDateEQ(v *int64) {
	m.DateEQ = v
}

// GetDateLTE returns the DateLTE value.
func (m SearchProfileSwitchesParams) GetDateLTE() *int64 {
	return m.DateLTE
}

// SetDateLTE sets the DateLTE value and returns the updated SearchProfileSwitchesParams.
func (m *SearchProfileSwitchesParams) SetDateLTE(v *int64) {
	m.DateLTE = v
}

// GetDateLT returns the DateLT value.
func (m SearchProfileSwitchesParams) GetDateLT() *int64 {
	return m.DateLT
}

// SetDateLT sets the DateLT value and returns the updated SearchProfileSwitchesParams.
func (m *SearchProfileSwitchesParams) SetDateLT(v *int64) {
	m.DateLT = v
}

// GetCreatedAtGTE returns the CreatedAtGTE value.
func (m SearchProfileSwitchesParams) GetCreatedAtGTE() *int64 {
	return m.CreatedAtGTE
}

// SetCreatedAtGTE sets the CreatedAtGTE value and returns the updated SearchProfileSwitchesParams.
func (m *SearchProfileSwitchesParams) SetCreatedAtGTE(v *int64) {
	m.CreatedAtGTE = v
}

// GetCreatedAtLTE returns the CreatedAtLTE value.
func (m SearchProfileSwitchesParams) GetCreatedAtLTE() *int64 {
	return m.CreatedAtLTE
}

// SetCreatedAtLTE sets the CreatedAtLTE value and returns the updated SearchProfileSwitchesParams.
func (m *SearchProfileSwitchesParams) SetCreatedAtLTE(v *int64) {
	m.CreatedAtLTE = v
}

// GetCreatedAtGT returns the CreatedAtGT value.
func (m SearchProfileSwitchesParams) GetCreatedAtGT() *int64 {
	return m.CreatedAtGT
}

// SetCreatedAtGT sets the CreatedAtGT value and returns the updated SearchProfileSwitchesParams.
func (m *SearchProfileSwitchesParams) SetCreatedAtGT(v *int64) {
	m.CreatedAtGT = v
}

// GetCreatedAtLT returns the CreatedAtLT value.
func (m SearchProfileSwitchesParams) GetCreatedAtLT() *int64 {
	return m.CreatedAtLT
}

// SetCreatedAtLT sets the CreatedAtLT value and returns the updated SearchProfileSwitchesParams.
func (m *SearchProfileSwitchesParams) SetCreatedAtLT(v *int64) {
	m.CreatedAtLT = v
}

// GetUpdatedGTE returns the UpdatedGTE value.
func (m SearchProfileSwitchesParams) GetUpdatedGTE() *int64 {
	return m.UpdatedGTE
}

// SetUpdatedGTE sets the UpdatedGTE value and returns the updated SearchProfileSwitchesParams.
func (m *SearchProfileSwitchesParams) SetUpdatedGTE(v *int64) {
	m.UpdatedGTE = v
}

// GetUpdatedLTE returns the UpdatedLTE value.
func (m SearchProfileSwitchesParams) GetUpdatedLTE() *int64 {
	return m.UpdatedLTE
}

// SetUpdatedLTE sets the UpdatedLTE value and returns the updated SearchProfileSwitchesParams.
func (m *SearchProfileSwitchesParams) SetUpdatedLTE(v *int64) {
	m.UpdatedLTE = v
}

// GetUpdatedGT returns the UpdatedGT value.
func (m SearchProfileSwitchesParams) GetUpdatedGT() *int64 {
	return m.UpdatedGT
}

// SetUpdatedGT sets the UpdatedGT value and returns the updated SearchProfileSwitchesParams.
func (m *SearchProfileSwitchesParams) SetUpdatedGT(v *int64) {
	m.UpdatedGT = v
}

// GetUpdatedLT returns the UpdatedLT value.
func (m SearchProfileSwitchesParams) GetUpdatedLT() *int64 {
	return m.UpdatedLT
}

// SetUpdatedLT sets the UpdatedLT value and returns the updated SearchProfileSwitchesParams.
func (m *SearchProfileSwitchesParams) SetUpdatedLT(v *int64) {
	m.UpdatedLT = v
}

// GetIDEq returns the IDEq value.
func (m SearchProfileSwitchesParams) GetIDEq() *string {
	return m.IDEq
}

// SetIDEq sets the IDEq value and returns the updated SearchProfileSwitchesParams.
func (m *SearchProfileSwitchesParams) SetIDEq(v *string) {
	m.IDEq = v
}

// GetIDIn returns the IDIn value.
func (m SearchProfileSwitchesParams) GetIDIn() *string {
	return m.IDIn
}

// SetIDIn sets the IDIn value and returns the updated SearchProfileSwitchesParams.
func (m *SearchProfileSwitchesParams) SetIDIn(v *string) {
	m.IDIn = v
}

// GetNameEq returns the NameEq value.
func (m SearchProfileSwitchesParams) GetNameEq() *string {
	return m.NameEq
}

// SetNameEq sets the NameEq value and returns the updated SearchProfileSwitchesParams.
func (m *SearchProfileSwitchesParams) SetNameEq(v *string) {
	m.NameEq = v
}

// GetDeviceEq returns the DeviceEq value.
func (m SearchProfileSwitchesParams) GetDeviceEq() *string {
	return m.DeviceEq
}

// SetDeviceEq sets the DeviceEq value and returns the updated SearchProfileSwitchesParams.
func (m *SearchProfileSwitchesParams) SetDeviceEq(v *string) {
	m.DeviceEq = v
}

// GetAppEq returns the AppEq value.
func (m SearchProfileSwitchesParams) GetAppEq() *string {
	return m.AppEq
}

// SetAppEq sets the AppEq value and returns the updated SearchProfileSwitchesParams.
func (m *SearchProfileSwitchesParams) SetAppEq(v *string) {
	m.AppEq = v
}

// GetDurationEQ returns the DurationEQ value.
func (m SearchProfileSwitchesParams) GetDurationEQ() *int64 {
	return m.DurationEQ
}

// SetDurationEQ sets the DurationEQ value and returns the updated SearchProfileSwitchesParams.
func (m *SearchProfileSwitchesParams) SetDurationEQ(v *int64) {
	m.DurationEQ = v
}

// GetDurationGTE returns the DurationGTE value.
func (m SearchProfileSwitchesParams) GetDurationGTE() *int64 {
	return m.DurationGTE
}

// SetDurationGTE sets the DurationGTE value and returns the updated SearchProfileSwitchesParams.
func (m *SearchProfileSwitchesParams) SetDurationGTE(v *int64) {
	m.DurationGTE = v
}

// GetDurationGT returns the DurationGT value.
func (m SearchProfileSwitchesParams) GetDurationGT() *int64 {
	return m.DurationGT
}

// SetDurationGT sets the DurationGT value and returns the updated SearchProfileSwitchesParams.
func (m *SearchProfileSwitchesParams) SetDurationGT(v *int64) {
	m.DurationGT = v
}

// GetDurationLTE returns the DurationLTE value.
func (m SearchProfileSwitchesParams) GetDurationLTE() *int64 {
	return m.DurationLTE
}

// SetDurationLTE sets the DurationLTE value and returns the updated SearchProfileSwitchesParams.
func (m *SearchProfileSwitchesParams) SetDurationLTE(v *int64) {
	m.DurationLTE = v
}

// GetDurationLT returns the DurationLT value.
func (m SearchProfileSwitchesParams) GetDurationLT() *int64 {
	return m.DurationLT
}

// SetDurationLT sets the DurationLT value and returns the updated SearchProfileSwitchesParams.
func (m *SearchProfileSwitchesParams) SetDurationLT(v *int64) {
	m.DurationLT = v
}

// GetIncludeDeleted returns the IncludeDeleted value.
func (m SearchProfileSwitchesParams) GetIncludeDeleted() *bool {
	return m.IncludeDeleted
}

// SetIncludeDeleted sets the IncludeDeleted value and returns the updated SearchProfileSwitchesParams.
func (m *SearchProfileSwitchesParams) SetIncludeDeleted(v *bool) {
	m.IncludeDeleted = v
}

// GetSkip returns the Skip value.
func (m SearchProfileSwitchesParams) GetSkip() int64 {
	return m.Skip
}

// SetSkip sets the Skip value and returns the updated SearchProfileSwitchesParams.
func (m *SearchProfileSwitchesParams) SetSkip(v int64) {
	m.Skip = v
}

// GetLimit returns the Limit value.
func (m SearchProfileSwitchesParams) GetLimit() int64 {
	return m.Limit
}

// SetLimit sets the Limit value and returns the updated SearchProfileSwitchesParams.
func (m *SearchProfileSwitchesParams) SetLimit(v int64) {
	m.Limit = v
}

// GetUntil returns the Until value.
func (m SearchSummaryProfileSwitchesParams) GetUntil() int64 {
	return m.Until
}

// SetUntil sets the Until value and returns the updated SearchSummaryProfileSwitchesParams.
func (m *SearchSummaryProfileSwitchesParams) SetUntil(v int64) {
	m.Until = v
}

// GetSince returns the Since value.
func (m SearchSummaryProfileSwitchesParams) GetSince() int64 {
	return m.Since
}

// SetSince sets the Since value and returns the updated SearchSummaryProfileSwitchesParams.
func (m *SearchSummaryProfileSwitchesParams) SetSince(v int64) {
	m.Since = v
}

// GetEarliest returns the Earliest value.
func (m SearchSummaryProfileSwitchesParams) GetEarliest() int64 {
	return m.Earliest
}

// SetEarliest sets the Earliest value and returns the updated SearchSummaryProfileSwitchesParams.
func (m *SearchSummaryProfileSwitchesParams) SetEarliest(v int64) {
	m.Earliest = v
}

// GetLimit returns the Limit value.
func (m SearchSummaryProfileSwitchesParams) GetLimit() int64 {
	return m.Limit
}

// SetLimit sets the Limit value and returns the updated SearchSummaryProfileSwitchesParams.
func (m *SearchSummaryProfileSwitchesParams) SetLimit(v int64) {
	m.Limit = v
}

// GetUpdatedByIdentityName returns the UpdatedByIdentityName value.
func (m SoftDeleteProfileSwitchParams) GetUpdatedByIdentityName() *string {
	return m.UpdatedByIdentityName
}

// SetUpdatedByIdentityName sets the UpdatedByIdentityName value and returns the updated SoftDeleteProfileSwitchParams.
func (m *SoftDeleteProfileSwitchParams) SetUpdatedByIdentityName(v *string) {
	m.UpdatedByIdentityName = v
}

// GetID returns the ID value.
func (m SoftDeleteProfileSwitchParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated SoftDeleteProfileSwitchParams.
func (m *SoftDeleteProfileSwitchParams) SetID(v string) {
	m.ID = v
}

// GetName returns the Name value.
func (m UpdateProfileSwitchParams) GetName() *string {
	return m.Name
}

// SetName sets the Name value and returns the updated UpdateProfileSwitchParams.
func (m *UpdateProfileSwitchParams) SetName(v *string) {
	m.Name = v
}

// GetDurationMinutes returns the DurationMinutes value.
func (m UpdateProfileSwitchParams) GetDurationMinutes() *int64 {
	return m.DurationMinutes
}

// SetDurationMinutes sets the DurationMinutes value and returns the updated UpdateProfileSwitchParams.
func (m *UpdateProfileSwitchParams) SetDurationMinutes(v *int64) {
	m.DurationMinutes = v
}

// GetPercent returns the Percent value.
func (m UpdateProfileSwitchParams) GetPercent() *float64 {
	return m.Percent
}

// SetPercent sets the Percent value and returns the updated UpdateProfileSwitchParams.
func (m *UpdateProfileSwitchParams) SetPercent(v *float64) {
	m.Percent = v
}

// GetTimeshift returns the Timeshift value.
func (m UpdateProfileSwitchParams) GetTimeshift() *int64 {
	return m.Timeshift
}

// SetTimeshift sets the Timeshift value and returns the updated UpdateProfileSwitchParams.
func (m *UpdateProfileSwitchParams) SetTimeshift(v *int64) {
	m.Timeshift = v
}

// GetNotes returns the Notes value.
func (m UpdateProfileSwitchParams) GetNotes() *string {
	return m.Notes
}

// SetNotes sets the Notes value and returns the updated UpdateProfileSwitchParams.
func (m *UpdateProfileSwitchParams) SetNotes(v *string) {
	m.Notes = v
}

// GetReason returns the Reason value.
func (m UpdateProfileSwitchParams) GetReason() *string {
	return m.Reason
}

// SetReason sets the Reason value and returns the updated UpdateProfileSwitchParams.
func (m *UpdateProfileSwitchParams) SetReason(v *string) {
	m.Reason = v
}

// GetRawJsonData returns the RawJsonData value.
func (m UpdateProfileSwitchParams) GetRawJsonData() *string {
	return m.RawJsonData
}

// SetRawJsonData sets the RawJsonData value and returns the updated UpdateProfileSwitchParams.
func (m *UpdateProfileSwitchParams) SetRawJsonData(v *string) {
	m.RawJsonData = v
}

// GetUpdatedByIdentityID returns the UpdatedByIdentityID value.
func (m UpdateProfileSwitchParams) GetUpdatedByIdentityID() *int64 {
	return m.UpdatedByIdentityID
}

// SetUpdatedByIdentityID sets the UpdatedByIdentityID value and returns the updated UpdateProfileSwitchParams.
func (m *UpdateProfileSwitchParams) SetUpdatedByIdentityID(v *int64) {
	m.UpdatedByIdentityID = v
}

// GetUpdatedByIdentityName returns the UpdatedByIdentityName value.
func (m UpdateProfileSwitchParams) GetUpdatedByIdentityName() *string {
	return m.UpdatedByIdentityName
}

// SetUpdatedByIdentityName sets the UpdatedByIdentityName value and returns the updated UpdateProfileSwitchParams.
func (m *UpdateProfileSwitchParams) SetUpdatedByIdentityName(v *string) {
	m.UpdatedByIdentityName = v
}

// GetID returns the ID value.
func (m UpdateProfileSwitchParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated UpdateProfileSwitchParams.
func (m *UpdateProfileSwitchParams) SetID(v string) {
	m.ID = v
}

// GetIncludeDeleted returns the IncludeDeleted value.
func (m AllTempBasalsParams) GetIncludeDeleted() bool {
	return m.IncludeDeleted
}

// SetIncludeDeleted sets the IncludeDeleted value and returns the updated AllTempBasalsParams.
func (m *AllTempBasalsParams) SetIncludeDeleted(v bool) {
	m.IncludeDeleted = v
}

// GetSkip returns the Skip value.
func (m AllTempBasalsParams) GetSkip() int64 {
	return m.Skip
}

// SetSkip sets the Skip value and returns the updated AllTempBasalsParams.
func (m *AllTempBasalsParams) SetSkip(v int64) {
	m.Skip = v
}

// GetLimit returns the Limit value.
func (m AllTempBasalsParams) GetLimit() int64 {
	return m.Limit
}

// SetLimit sets the Limit value and returns the updated AllTempBasalsParams.
func (m *AllTempBasalsParams) SetLimit(v int64) {
	m.Limit = v
}

// GetID returns the ID value.
func (m CreateTempBasalParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated CreateTempBasalParams.
func (m *CreateTempBasalParams) SetID(v string) {
	m.ID = v
}

// GetDurationMinutes returns the DurationMinutes value.
func (m CreateTempBasalParams) GetDurationMinutes() int64 {
	return m.DurationMinutes
}

// SetDurationMinutes sets the DurationMinutes value and returns the updated CreateTempBasalParams.
func (m *CreateTempBasalParams) SetDurationMinutes(v int64) {
	m.DurationMinutes = v
}

// GetAbsolute returns the Absolute value.
func (m CreateTempBasalParams) GetAbsolute() float64 {
	return m.Absolute
}

// SetAbsolute sets the Absolute value and returns the updated CreateTempBasalParams.
func (m *CreateTempBasalParams) SetAbsolute(v float64) {
	m.Absolute = v
}

// GetPercent returns the Percent value.
func (m CreateTempBasalParams) GetPercent() *float64 {
	return m.Percent
}

// SetPercent sets the Percent value and returns the updated CreateTempBasalParams.
func (m *CreateTempBasalParams) SetPercent(v *float64) {
	m.Percent = v
}

// GetNotes returns the Notes value.
func (m CreateTempBasalParams) GetNotes() *string {
	return m.Notes
}

// SetNotes sets the Notes value and returns the updated CreateTempBasalParams.
func (m *CreateTempBasalParams) SetNotes(v *string) {
	m.Notes = v
}

// GetReason returns the Reason value.
func (m CreateTempBasalParams) GetReason() *string {
	return m.Reason
}

// SetReason sets the Reason value and returns the updated CreateTempBasalParams.
func (m *CreateTempBasalParams) SetReason(v *string) {
	m.Reason = v
}

// GetUtcOffsetMinutes returns the UtcOffsetMinutes value.
func (m CreateTempBasalParams) GetUtcOffsetMinutes() *int64 {
	return m.UtcOffsetMinutes
}

// SetUtcOffsetMinutes sets the UtcOffsetMinutes value and returns the updated CreateTempBasalParams.
func (m *CreateTempBasalParams) SetUtcOffsetMinutes(v *int64) {
	m.UtcOffsetMinutes = v
}

// GetOccurredAt returns the OccurredAt value.
func (m CreateTempBasalParams) GetOccurredAt() int64 {
	return m.OccurredAt
}

// SetOccurredAt sets the OccurredAt value and returns the updated CreateTempBasalParams.
func (m *CreateTempBasalParams) SetOccurredAt(v int64) {
	m.OccurredAt = v
}

// GetRawJsonData returns the RawJsonData value.
func (m CreateTempBasalParams) GetRawJsonData() *string {
	return m.RawJsonData
}

// SetRawJsonData sets the RawJsonData value and returns the updated CreateTempBasalParams.
func (m *CreateTempBasalParams) SetRawJsonData(v *string) {
	m.RawJsonData = v
}

// GetCreatedByAppName returns the CreatedByAppName value.
func (m CreateTempBasalParams) GetCreatedByAppName() *string {
	return m.CreatedByAppName
}

// SetCreatedByAppName sets the CreatedByAppName value and returns the updated CreateTempBasalParams.
func (m *CreateTempBasalParams) SetCreatedByAppName(v *string) {
	m.CreatedByAppName = v
}

// GetCreatedByDeviceName returns the CreatedByDeviceName value.
func (m CreateTempBasalParams) GetCreatedByDeviceName() *string {
	return m.CreatedByDeviceName
}

// SetCreatedByDeviceName sets the CreatedByDeviceName value and returns the updated CreateTempBasalParams.
func (m *CreateTempBasalParams) SetCreatedByDeviceName(v *string) {
	m.CreatedByDeviceName = v
}

// GetCreatedByIdentityName returns the CreatedByIdentityName value.
func (m CreateTempBasalParams) GetCreatedByIdentityName() string {
	return m.CreatedByIdentityName
}

// SetCreatedByIdentityName sets the CreatedByIdentityName value and returns the updated CreateTempBasalParams.
func (m *CreateTempBasalParams) SetCreatedByIdentityName(v string) {
	m.CreatedByIdentityName = v
}

// GetCreatedByIdentityID returns the CreatedByIdentityID value.
func (m CreateTempBasalParams) GetCreatedByIdentityID() *int64 {
	return m.CreatedByIdentityID
}

// SetCreatedByIdentityID sets the CreatedByIdentityID value and returns the updated CreateTempBasalParams.
func (m *CreateTempBasalParams) SetCreatedByIdentityID(v *int64) {
	m.CreatedByIdentityID = v
}

// GetID returns the ID value.
func (m GetTempBasalParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated GetTempBasalParams.
func (m *GetTempBasalParams) SetID(v string) {
	m.ID = v
}

// GetIncludeDeleted returns the IncludeDeleted value.
func (m GetTempBasalParams) GetIncludeDeleted() bool {
	return m.IncludeDeleted
}

// SetIncludeDeleted sets the IncludeDeleted value and returns the updated GetTempBasalParams.
func (m *GetTempBasalParams) SetIncludeDeleted(v bool) {
	m.IncludeDeleted = v
}

// GetID returns the ID value.
func (m HardDeleteTempBasalParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated HardDeleteTempBasalParams.
func (m *HardDeleteTempBasalParams) SetID(v string) {
	m.ID = v
}

// GetDateGTE returns the DateGTE value.
func (m SearchTempBasalsParams) GetDateGTE() *int64 {
	return m.DateGTE
}

// SetDateGTE sets the DateGTE value and returns the updated SearchTempBasalsParams.
func (m *SearchTempBasalsParams) SetDateGTE(v *int64) {
	m.DateGTE = v
}

// GetDateGT returns the DateGT value.
func (m SearchTempBasalsParams) GetDateGT() *int64 {
	return m.DateGT
}

// SetDateGT sets the DateGT value and returns the updated SearchTempBasalsParams.
func (m *SearchTempBasalsParams) SetDateGT(v *int64) {
	m.DateGT = v
}

// GetDateEQ returns the DateEQ value.
func (m SearchTempBasalsParams) GetDateEQ() *int64 {
	return m.DateEQ
}

// SetDateEQ sets the DateEQ value and returns the updated SearchTempBasalsParams.
func (m *SearchTempBasalsParams) SetDateEQ(v *int64) {
	m.DateEQ = v
}

// GetDateLTE returns the DateLTE value.
func (m SearchTempBasalsParams) GetDateLTE() *int64 {
	return m.DateLTE
}

// SetDateLTE sets the DateLTE value and returns the updated SearchTempBasalsParams.
func (m *SearchTempBasalsParams) SetDateLTE(v *int64) {
	m.DateLTE = v
}

// GetDateLT returns the DateLT value.
func (m SearchTempBasalsParams) GetDateLT() *int64 {
	return m.DateLT
}

// SetDateLT sets the DateLT value and returns the updated SearchTempBasalsParams.
func (m *SearchTempBasalsParams) SetDateLT(v *int64) {
	m.DateLT = v
}

// GetCreatedAtGTE returns the CreatedAtGTE value.
func (m SearchTempBasalsParams) GetCreatedAtGTE() *int64 {
	return m.CreatedAtGTE
}

// SetCreatedAtGTE sets the CreatedAtGTE value and returns the updated SearchTempBasalsParams.
func (m *SearchTempBasalsParams) SetCreatedAtGTE(v *int64) {
	m.CreatedAtGTE = v
}

// GetCreatedAtLTE returns the CreatedAtLTE value.
func (m SearchTempBasalsParams) GetCreatedAtLTE() *int64 {
	return m.CreatedAtLTE
}

// SetCreatedAtLTE sets the CreatedAtLTE value and returns the updated SearchTempBasalsParams.
func (m *SearchTempBasalsParams) SetCreatedAtLTE(v *int64) {
	m.CreatedAtLTE = v
}

// GetCreatedAtGT returns the CreatedAtGT value.
func (m SearchTempBasalsParams) GetCreatedAtGT() *int64 {
	return m.CreatedAtGT
}

// SetCreatedAtGT sets the CreatedAtGT value and returns the updated SearchTempBasalsParams.
func (m *SearchTempBasalsParams) SetCreatedAtGT(v *int64) {
	m.CreatedAtGT = v
}

// GetCreatedAtLT returns the CreatedAtLT value.
func (m SearchTempBasalsParams) GetCreatedAtLT() *int64 {
	return m.CreatedAtLT
}

// SetCreatedAtLT sets the CreatedAtLT value and returns the updated SearchTempBasalsParams.
func (m *SearchTempBasalsParams) SetCreatedAtLT(v *int64) {
	m.CreatedAtLT = v
}

// GetUpdatedGTE returns the UpdatedGTE value.
func (m SearchTempBasalsParams) GetUpdatedGTE() *int64 {
	return m.UpdatedGTE
}

// SetUpdatedGTE sets the UpdatedGTE value and returns the updated SearchTempBasalsParams.
func (m *SearchTempBasalsParams) SetUpdatedGTE(v *int64) {
	m.UpdatedGTE = v
}

// GetUpdatedLTE returns the UpdatedLTE value.
func (m SearchTempBasalsParams) GetUpdatedLTE() *int64 {
	return m.UpdatedLTE
}

// SetUpdatedLTE sets the UpdatedLTE value and returns the updated SearchTempBasalsParams.
func (m *SearchTempBasalsParams) SetUpdatedLTE(v *int64) {
	m.UpdatedLTE = v
}

// GetUpdatedGT returns the UpdatedGT value.
func (m SearchTempBasalsParams) GetUpdatedGT() *int64 {
	return m.UpdatedGT
}

// SetUpdatedGT sets the UpdatedGT value and returns the updated SearchTempBasalsParams.
func (m *SearchTempBasalsParams) SetUpdatedGT(v *int64) {
	m.UpdatedGT = v
}

// GetUpdatedLT returns the UpdatedLT value.
func (m SearchTempBasalsParams) GetUpdatedLT() *int64 {
	return m.UpdatedLT
}

// SetUpdatedLT sets the UpdatedLT value and returns the updated SearchTempBasalsParams.
func (m *SearchTempBasalsParams) SetUpdatedLT(v *int64) {
	m.UpdatedLT = v
}

// GetIDEq returns the IDEq value.
func (m SearchTempBasalsParams) GetIDEq() *string {
	return m.IDEq
}

// SetIDEq sets the IDEq value and returns the updated SearchTempBasalsParams.
func (m *SearchTempBasalsParams) SetIDEq(v *string) {
	m.IDEq = v
}

// GetIDIn returns the IDIn value.
func (m SearchTempBasalsParams) GetIDIn() *string {
	return m.IDIn
}

// SetIDIn sets the IDIn value and returns the updated SearchTempBasalsParams.
func (m *SearchTempBasalsParams) SetIDIn(v *string) {
	m.IDIn = v
}

// GetDeviceEq returns the DeviceEq value.
func (m SearchTempBasalsParams) GetDeviceEq() *string {
	return m.DeviceEq
}

// SetDeviceEq sets the DeviceEq value and returns the updated SearchTempBasalsParams.
func (m *SearchTempBasalsParams) SetDeviceEq(v *string) {
	m.DeviceEq = v
}

// GetAppEq returns the AppEq value.
func (m SearchTempBasalsParams) GetAppEq() *string {
	return m.AppEq
}

// SetAppEq sets the AppEq value and returns the updated SearchTempBasalsParams.
func (m *SearchTempBasalsParams) SetAppEq(v *string) {
	m.AppEq = v
}

// GetDurationEQ returns the DurationEQ value.
func (m SearchTempBasalsParams) GetDurationEQ() *int64 {
	return m.DurationEQ
}

// SetDurationEQ sets the DurationEQ value and returns the updated SearchTempBasalsParams.
func (m *SearchTempBasalsParams) SetDurationEQ(v *int64) {
	m.DurationEQ = v
}

// GetDurationGTE returns the DurationGTE value.
func (m SearchTempBasalsParams) GetDurationGTE() *int64 {
	return m.DurationGTE
}

// SetDurationGTE sets the DurationGTE value and returns the updated SearchTempBasalsParams.
func (m *SearchTempBasalsParams) SetDurationGTE(v *int64) {
	m.DurationGTE = v
}

// GetDurationGT returns the DurationGT value.
func (m SearchTempBasalsParams) GetDurationGT() *int64 {
	return m.DurationGT
}

// SetDurationGT sets the DurationGT value and returns the updated SearchTempBasalsParams.
func (m *SearchTempBasalsParams) SetDurationGT(v *int64) {
	m.DurationGT = v
}

// GetDurationLTE returns the DurationLTE value.
func (m SearchTempBasalsParams) GetDurationLTE() *int64 {
	return m.DurationLTE
}

// SetDurationLTE sets the DurationLTE value and returns the updated SearchTempBasalsParams.
func (m *SearchTempBasalsParams) SetDurationLTE(v *int64) {
	m.DurationLTE = v
}

// GetDurationLT returns the DurationLT value.
func (m SearchTempBasalsParams) GetDurationLT() *int64 {
	return m.DurationLT
}

// SetDurationLT sets the DurationLT value and returns the updated SearchTempBasalsParams.
func (m *SearchTempBasalsParams) SetDurationLT(v *int64) {
	m.DurationLT = v
}

// GetIncludeDeleted returns the IncludeDeleted value.
func (m SearchTempBasalsParams) GetIncludeDeleted() *bool {
	return m.IncludeDeleted
}

// SetIncludeDeleted sets the IncludeDeleted value and returns the updated SearchTempBasalsParams.
func (m *SearchTempBasalsParams) SetIncludeDeleted(v *bool) {
	m.IncludeDeleted = v
}

// GetSkip returns the Skip value.
func (m SearchTempBasalsParams) GetSkip() int64 {
	return m.Skip
}

// SetSkip sets the Skip value and returns the updated SearchTempBasalsParams.
func (m *SearchTempBasalsParams) SetSkip(v int64) {
	m.Skip = v
}

// GetLimit returns the Limit value.
func (m SearchTempBasalsParams) GetLimit() int64 {
	return m.Limit
}

// SetLimit sets the Limit value and returns the updated SearchTempBasalsParams.
func (m *SearchTempBasalsParams) SetLimit(v int64) {
	m.Limit = v
}

// GetUpdatedByIdentityName returns the UpdatedByIdentityName value.
func (m SoftDeleteTempBasalParams) GetUpdatedByIdentityName() *string {
	return m.UpdatedByIdentityName
}

// SetUpdatedByIdentityName sets the UpdatedByIdentityName value and returns the updated SoftDeleteTempBasalParams.
func (m *SoftDeleteTempBasalParams) SetUpdatedByIdentityName(v *string) {
	m.UpdatedByIdentityName = v
}

// GetID returns the ID value.
func (m SoftDeleteTempBasalParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated SoftDeleteTempBasalParams.
func (m *SoftDeleteTempBasalParams) SetID(v string) {
	m.ID = v
}

// GetDurationMinutes returns the DurationMinutes value.
func (m UpdateTempBasalParams) GetDurationMinutes() *int64 {
	return m.DurationMinutes
}

// SetDurationMinutes sets the DurationMinutes value and returns the updated UpdateTempBasalParams.
func (m *UpdateTempBasalParams) SetDurationMinutes(v *int64) {
	m.DurationMinutes = v
}

// GetAbsolute returns the Absolute value.
func (m UpdateTempBasalParams) GetAbsolute() *float64 {
	return m.Absolute
}

// SetAbsolute sets the Absolute value and returns the updated UpdateTempBasalParams.
func (m *UpdateTempBasalParams) SetAbsolute(v *float64) {
	m.Absolute = v
}

// GetPercent returns the Percent value.
func (m UpdateTempBasalParams) GetPercent() *float64 {
	return m.Percent
}

// SetPercent sets the Percent value and returns the updated UpdateTempBasalParams.
func (m *UpdateTempBasalParams) SetPercent(v *float64) {
	m.Percent = v
}

// GetNotes returns the Notes value.
func (m UpdateTempBasalParams) GetNotes() *string {
	return m.Notes
}

// SetNotes sets the Notes value and returns the updated UpdateTempBasalParams.
func (m *UpdateTempBasalParams) SetNotes(v *string) {
	m.Notes = v
}

// GetReason returns the Reason value.
func (m UpdateTempBasalParams) GetReason() *string {
	return m.Reason
}

// SetReason sets the Reason value and returns the updated UpdateTempBasalParams.
func (m *UpdateTempBasalParams) SetReason(v *string) {
	m.Reason = v
}

// GetRawJsonData returns the RawJsonData value.
func (m UpdateTempBasalParams) GetRawJsonData() *string {
	return m.RawJsonData
}

// SetRawJsonData sets the RawJsonData value and returns the updated UpdateTempBasalParams.
func (m *UpdateTempBasalParams) SetRawJsonData(v *string) {
	m.RawJsonData = v
}

// GetUpdatedByIdentityID returns the UpdatedByIdentityID value.
func (m UpdateTempBasalParams) GetUpdatedByIdentityID() *int64 {
	return m.UpdatedByIdentityID
}

// SetUpdatedByIdentityID sets the UpdatedByIdentityID value and returns the updated UpdateTempBasalParams.
func (m *UpdateTempBasalParams) SetUpdatedByIdentityID(v *int64) {
	m.UpdatedByIdentityID = v
}

// GetUpdatedByIdentityName returns the UpdatedByIdentityName value.
func (m UpdateTempBasalParams) GetUpdatedByIdentityName() *string {
	return m.UpdatedByIdentityName
}

// SetUpdatedByIdentityName sets the UpdatedByIdentityName value and returns the updated UpdateTempBasalParams.
func (m *UpdateTempBasalParams) SetUpdatedByIdentityName(v *string) {
	m.UpdatedByIdentityName = v
}

// GetID returns the ID value.
func (m UpdateTempBasalParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated UpdateTempBasalParams.
func (m *UpdateTempBasalParams) SetID(v string) {
	m.ID = v
}

// GetIncludeDeleted returns the IncludeDeleted value.
func (m AllTempTargetsParams) GetIncludeDeleted() bool {
	return m.IncludeDeleted
}

// SetIncludeDeleted sets the IncludeDeleted value and returns the updated AllTempTargetsParams.
func (m *AllTempTargetsParams) SetIncludeDeleted(v bool) {
	m.IncludeDeleted = v
}

// GetSkip returns the Skip value.
func (m AllTempTargetsParams) GetSkip() int64 {
	return m.Skip
}

// SetSkip sets the Skip value and returns the updated AllTempTargetsParams.
func (m *AllTempTargetsParams) SetSkip(v int64) {
	m.Skip = v
}

// GetLimit returns the Limit value.
func (m AllTempTargetsParams) GetLimit() int64 {
	return m.Limit
}

// SetLimit sets the Limit value and returns the updated AllTempTargetsParams.
func (m *AllTempTargetsParams) SetLimit(v int64) {
	m.Limit = v
}

// GetID returns the ID value.
func (m CreateTempTargetParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated CreateTempTargetParams.
func (m *CreateTempTargetParams) SetID(v string) {
	m.ID = v
}

// GetDurationMinutes returns the DurationMinutes value.
func (m CreateTempTargetParams) GetDurationMinutes() int64 {
	return m.DurationMinutes
}

// SetDurationMinutes sets the DurationMinutes value and returns the updated CreateTempTargetParams.
func (m *CreateTempTargetParams) SetDurationMinutes(v int64) {
	m.DurationMinutes = v
}

// GetHighTargetMgdl returns the HighTargetMgdl value.
func (m CreateTempTargetParams) GetHighTargetMgdl() float64 {
	return m.HighTargetMgdl
}

// SetHighTargetMgdl sets the HighTargetMgdl value and returns the updated CreateTempTargetParams.
func (m *CreateTempTargetParams) SetHighTargetMgdl(v float64) {
	m.HighTargetMgdl = v
}

// GetLowTargetMgdl returns the LowTargetMgdl value.
func (m CreateTempTargetParams) GetLowTargetMgdl() float64 {
	return m.LowTargetMgdl
}

// SetLowTargetMgdl sets the LowTargetMgdl value and returns the updated CreateTempTargetParams.
func (m *CreateTempTargetParams) SetLowTargetMgdl(v float64) {
	m.LowTargetMgdl = v
}

// GetInsulinNeedsScaleFactor returns the InsulinNeedsScaleFactor value.
func (m CreateTempTargetParams) GetInsulinNeedsScaleFactor() *float64 {
	return m.InsulinNeedsScaleFactor
}

// SetInsulinNeedsScaleFactor sets the InsulinNeedsScaleFactor value and returns the updated CreateTempTargetParams.
func (m *CreateTempTargetParams) SetInsulinNeedsScaleFactor(v *float64) {
	m.InsulinNeedsScaleFactor = v
}

// GetNotes returns the Notes value.
func (m CreateTempTargetParams) GetNotes() *string {
	return m.Notes
}

// SetNotes sets the Notes value and returns the updated CreateTempTargetParams.
func (m *CreateTempTargetParams) SetNotes(v *string) {
	m.Notes = v
}

// GetReason returns the Reason value.
func (m CreateTempTargetParams) GetReason() *string {
	return m.Reason
}

// SetReason sets the Reason value and returns the updated CreateTempTargetParams.
func (m *CreateTempTargetParams) SetReason(v *string) {
	m.Reason = v
}

// GetUtcOffsetMinutes returns the UtcOffsetMinutes value.
func (m CreateTempTargetParams) GetUtcOffsetMinutes() *int64 {
	return m.UtcOffsetMinutes
}

// SetUtcOffsetMinutes sets the UtcOffsetMinutes value and returns the updated CreateTempTargetParams.
func (m *CreateTempTargetParams) SetUtcOffsetMinutes(v *int64) {
	m.UtcOffsetMinutes = v
}

// GetOccurredAt returns the OccurredAt value.
func (m CreateTempTargetParams) GetOccurredAt() int64 {
	return m.OccurredAt
}

// SetOccurredAt sets the OccurredAt value and returns the updated CreateTempTargetParams.
func (m *CreateTempTargetParams) SetOccurredAt(v int64) {
	m.OccurredAt = v
}

// GetRawJsonData returns the RawJsonData value.
func (m CreateTempTargetParams) GetRawJsonData() *string {
	return m.RawJsonData
}

// SetRawJsonData sets the RawJsonData value and returns the updated CreateTempTargetParams.
func (m *CreateTempTargetParams) SetRawJsonData(v *string) {
	m.RawJsonData = v
}

// GetCreatedByAppName returns the CreatedByAppName value.
func (m CreateTempTargetParams) GetCreatedByAppName() *string {
	return m.CreatedByAppName
}

// SetCreatedByAppName sets the CreatedByAppName value and returns the updated CreateTempTargetParams.
func (m *CreateTempTargetParams) SetCreatedByAppName(v *string) {
	m.CreatedByAppName = v
}

// GetCreatedByDeviceName returns the CreatedByDeviceName value.
func (m CreateTempTargetParams) GetCreatedByDeviceName() *string {
	return m.CreatedByDeviceName
}

// SetCreatedByDeviceName sets the CreatedByDeviceName value and returns the updated CreateTempTargetParams.
func (m *CreateTempTargetParams) SetCreatedByDeviceName(v *string) {
	m.CreatedByDeviceName = v
}

// GetCreatedByIdentityName returns the CreatedByIdentityName value.
func (m CreateTempTargetParams) GetCreatedByIdentityName() string {
	return m.CreatedByIdentityName
}

// SetCreatedByIdentityName sets the CreatedByIdentityName value and returns the updated CreateTempTargetParams.
func (m *CreateTempTargetParams) SetCreatedByIdentityName(v string) {
	m.CreatedByIdentityName = v
}

// GetCreatedByIdentityID returns the CreatedByIdentityID value.
func (m CreateTempTargetParams) GetCreatedByIdentityID() *int64 {
	return m.CreatedByIdentityID
}

// SetCreatedByIdentityID sets the CreatedByIdentityID value and returns the updated CreateTempTargetParams.
func (m *CreateTempTargetParams) SetCreatedByIdentityID(v *int64) {
	m.CreatedByIdentityID = v
}

// GetID returns the ID value.
func (m GetTempTargetParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated GetTempTargetParams.
func (m *GetTempTargetParams) SetID(v string) {
	m.ID = v
}

// GetIncludeDeleted returns the IncludeDeleted value.
func (m GetTempTargetParams) GetIncludeDeleted() bool {
	return m.IncludeDeleted
}

// SetIncludeDeleted sets the IncludeDeleted value and returns the updated GetTempTargetParams.
func (m *GetTempTargetParams) SetIncludeDeleted(v bool) {
	m.IncludeDeleted = v
}

// GetID returns the ID value.
func (m HardDeleteTempTargetParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated HardDeleteTempTargetParams.
func (m *HardDeleteTempTargetParams) SetID(v string) {
	m.ID = v
}

// GetDateGTE returns the DateGTE value.
func (m SearchTempTargetsParams) GetDateGTE() *int64 {
	return m.DateGTE
}

// SetDateGTE sets the DateGTE value and returns the updated SearchTempTargetsParams.
func (m *SearchTempTargetsParams) SetDateGTE(v *int64) {
	m.DateGTE = v
}

// GetDateGT returns the DateGT value.
func (m SearchTempTargetsParams) GetDateGT() *int64 {
	return m.DateGT
}

// SetDateGT sets the DateGT value and returns the updated SearchTempTargetsParams.
func (m *SearchTempTargetsParams) SetDateGT(v *int64) {
	m.DateGT = v
}

// GetDateEQ returns the DateEQ value.
func (m SearchTempTargetsParams) GetDateEQ() *int64 {
	return m.DateEQ
}

// SetDateEQ sets the DateEQ value and returns the updated SearchTempTargetsParams.
func (m *SearchTempTargetsParams) SetDateEQ(v *int64) {
	m.DateEQ = v
}

// GetDateLTE returns the DateLTE value.
func (m SearchTempTargetsParams) GetDateLTE() *int64 {
	return m.DateLTE
}

// SetDateLTE sets the DateLTE value and returns the updated SearchTempTargetsParams.
func (m *SearchTempTargetsParams) SetDateLTE(v *int64) {
	m.DateLTE = v
}

// GetDateLT returns the DateLT value.
func (m SearchTempTargetsParams) GetDateLT() *int64 {
	return m.DateLT
}

// SetDateLT sets the DateLT value and returns the updated SearchTempTargetsParams.
func (m *SearchTempTargetsParams) SetDateLT(v *int64) {
	m.DateLT = v
}

// GetCreatedAtGTE returns the CreatedAtGTE value.
func (m SearchTempTargetsParams) GetCreatedAtGTE() *int64 {
	return m.CreatedAtGTE
}

// SetCreatedAtGTE sets the CreatedAtGTE value and returns the updated SearchTempTargetsParams.
func (m *SearchTempTargetsParams) SetCreatedAtGTE(v *int64) {
	m.CreatedAtGTE = v
}

// GetCreatedAtLTE returns the CreatedAtLTE value.
func (m SearchTempTargetsParams) GetCreatedAtLTE() *int64 {
	return m.CreatedAtLTE
}

// SetCreatedAtLTE sets the CreatedAtLTE value and returns the updated SearchTempTargetsParams.
func (m *SearchTempTargetsParams) SetCreatedAtLTE(v *int64) {
	m.CreatedAtLTE = v
}

// GetCreatedAtGT returns the CreatedAtGT value.
func (m SearchTempTargetsParams) GetCreatedAtGT() *int64 {
	return m.CreatedAtGT
}

// SetCreatedAtGT sets the CreatedAtGT value and returns the updated SearchTempTargetsParams.
func (m *SearchTempTargetsParams) SetCreatedAtGT(v *int64) {
	m.CreatedAtGT = v
}

// GetCreatedAtLT returns the CreatedAtLT value.
func (m SearchTempTargetsParams) GetCreatedAtLT() *int64 {
	return m.CreatedAtLT
}

// SetCreatedAtLT sets the CreatedAtLT value and returns the updated SearchTempTargetsParams.
func (m *SearchTempTargetsParams) SetCreatedAtLT(v *int64) {
	m.CreatedAtLT = v
}

// GetUpdatedGTE returns the UpdatedGTE value.
func (m SearchTempTargetsParams) GetUpdatedGTE() *int64 {
	return m.UpdatedGTE
}

// SetUpdatedGTE sets the UpdatedGTE value and returns the updated SearchTempTargetsParams.
func (m *SearchTempTargetsParams) SetUpdatedGTE(v *int64) {
	m.UpdatedGTE = v
}

// GetUpdatedLTE returns the UpdatedLTE value.
func (m SearchTempTargetsParams) GetUpdatedLTE() *int64 {
	return m.UpdatedLTE
}

// SetUpdatedLTE sets the UpdatedLTE value and returns the updated SearchTempTargetsParams.
func (m *SearchTempTargetsParams) SetUpdatedLTE(v *int64) {
	m.UpdatedLTE = v
}

// GetUpdatedGT returns the UpdatedGT value.
func (m SearchTempTargetsParams) GetUpdatedGT() *int64 {
	return m.UpdatedGT
}

// SetUpdatedGT sets the UpdatedGT value and returns the updated SearchTempTargetsParams.
func (m *SearchTempTargetsParams) SetUpdatedGT(v *int64) {
	m.UpdatedGT = v
}

// GetUpdatedLT returns the UpdatedLT value.
func (m SearchTempTargetsParams) GetUpdatedLT() *int64 {
	return m.UpdatedLT
}

// SetUpdatedLT sets the UpdatedLT value and returns the updated SearchTempTargetsParams.
func (m *SearchTempTargetsParams) SetUpdatedLT(v *int64) {
	m.UpdatedLT = v
}

// GetIDEq returns the IDEq value.
func (m SearchTempTargetsParams) GetIDEq() *string {
	return m.IDEq
}

// SetIDEq sets the IDEq value and returns the updated SearchTempTargetsParams.
func (m *SearchTempTargetsParams) SetIDEq(v *string) {
	m.IDEq = v
}

// GetIDIn returns the IDIn value.
func (m SearchTempTargetsParams) GetIDIn() *string {
	return m.IDIn
}

// SetIDIn sets the IDIn value and returns the updated SearchTempTargetsParams.
func (m *SearchTempTargetsParams) SetIDIn(v *string) {
	m.IDIn = v
}

// GetDeviceEq returns the DeviceEq value.
func (m SearchTempTargetsParams) GetDeviceEq() *string {
	return m.DeviceEq
}

// SetDeviceEq sets the DeviceEq value and returns the updated SearchTempTargetsParams.
func (m *SearchTempTargetsParams) SetDeviceEq(v *string) {
	m.DeviceEq = v
}

// GetAppEq returns the AppEq value.
func (m SearchTempTargetsParams) GetAppEq() *string {
	return m.AppEq
}

// SetAppEq sets the AppEq value and returns the updated SearchTempTargetsParams.
func (m *SearchTempTargetsParams) SetAppEq(v *string) {
	m.AppEq = v
}

// GetDurationEQ returns the DurationEQ value.
func (m SearchTempTargetsParams) GetDurationEQ() *int64 {
	return m.DurationEQ
}

// SetDurationEQ sets the DurationEQ value and returns the updated SearchTempTargetsParams.
func (m *SearchTempTargetsParams) SetDurationEQ(v *int64) {
	m.DurationEQ = v
}

// GetDurationGTE returns the DurationGTE value.
func (m SearchTempTargetsParams) GetDurationGTE() *int64 {
	return m.DurationGTE
}

// SetDurationGTE sets the DurationGTE value and returns the updated SearchTempTargetsParams.
func (m *SearchTempTargetsParams) SetDurationGTE(v *int64) {
	m.DurationGTE = v
}

// GetDurationGT returns the DurationGT value.
func (m SearchTempTargetsParams) GetDurationGT() *int64 {
	return m.DurationGT
}

// SetDurationGT sets the DurationGT value and returns the updated SearchTempTargetsParams.
func (m *SearchTempTargetsParams) SetDurationGT(v *int64) {
	m.DurationGT = v
}

// GetDurationLTE returns the DurationLTE value.
func (m SearchTempTargetsParams) GetDurationLTE() *int64 {
	return m.DurationLTE
}

// SetDurationLTE sets the DurationLTE value and returns the updated SearchTempTargetsParams.
func (m *SearchTempTargetsParams) SetDurationLTE(v *int64) {
	m.DurationLTE = v
}

// GetDurationLT returns the DurationLT value.
func (m SearchTempTargetsParams) GetDurationLT() *int64 {
	return m.DurationLT
}

// SetDurationLT sets the DurationLT value and returns the updated SearchTempTargetsParams.
func (m *SearchTempTargetsParams) SetDurationLT(v *int64) {
	m.DurationLT = v
}

// GetIncludeDeleted returns the IncludeDeleted value.
func (m SearchTempTargetsParams) GetIncludeDeleted() *bool {
	return m.IncludeDeleted
}

// SetIncludeDeleted sets the IncludeDeleted value and returns the updated SearchTempTargetsParams.
func (m *SearchTempTargetsParams) SetIncludeDeleted(v *bool) {
	m.IncludeDeleted = v
}

// GetSkip returns the Skip value.
func (m SearchTempTargetsParams) GetSkip() int64 {
	return m.Skip
}

// SetSkip sets the Skip value and returns the updated SearchTempTargetsParams.
func (m *SearchTempTargetsParams) SetSkip(v int64) {
	m.Skip = v
}

// GetLimit returns the Limit value.
func (m SearchTempTargetsParams) GetLimit() int64 {
	return m.Limit
}

// SetLimit sets the Limit value and returns the updated SearchTempTargetsParams.
func (m *SearchTempTargetsParams) SetLimit(v int64) {
	m.Limit = v
}

// GetUpdatedByIdentityName returns the UpdatedByIdentityName value.
func (m SoftDeleteTempTargetParams) GetUpdatedByIdentityName() *string {
	return m.UpdatedByIdentityName
}

// SetUpdatedByIdentityName sets the UpdatedByIdentityName value and returns the updated SoftDeleteTempTargetParams.
func (m *SoftDeleteTempTargetParams) SetUpdatedByIdentityName(v *string) {
	m.UpdatedByIdentityName = v
}

// GetID returns the ID value.
func (m SoftDeleteTempTargetParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated SoftDeleteTempTargetParams.
func (m *SoftDeleteTempTargetParams) SetID(v string) {
	m.ID = v
}

// GetDurationMinutes returns the DurationMinutes value.
func (m UpdateTempTargetParams) GetDurationMinutes() *int64 {
	return m.DurationMinutes
}

// SetDurationMinutes sets the DurationMinutes value and returns the updated UpdateTempTargetParams.
func (m *UpdateTempTargetParams) SetDurationMinutes(v *int64) {
	m.DurationMinutes = v
}

// GetHighTargetMgdl returns the HighTargetMgdl value.
func (m UpdateTempTargetParams) GetHighTargetMgdl() *float64 {
	return m.HighTargetMgdl
}

// SetHighTargetMgdl sets the HighTargetMgdl value and returns the updated UpdateTempTargetParams.
func (m *UpdateTempTargetParams) SetHighTargetMgdl(v *float64) {
	m.HighTargetMgdl = v
}

// GetLowTargetMgdl returns the LowTargetMgdl value.
func (m UpdateTempTargetParams) GetLowTargetMgdl() *float64 {
	return m.LowTargetMgdl
}

// SetLowTargetMgdl sets the LowTargetMgdl value and returns the updated UpdateTempTargetParams.
func (m *UpdateTempTargetParams) SetLowTargetMgdl(v *float64) {
	m.LowTargetMgdl = v
}

// GetNotes returns the Notes value.
func (m UpdateTempTargetParams) GetNotes() *string {
	return m.Notes
}

// SetNotes sets the Notes value and returns the updated UpdateTempTargetParams.
func (m *UpdateTempTargetParams) SetNotes(v *string) {
	m.Notes = v
}

// GetReason returns the Reason value.
func (m UpdateTempTargetParams) GetReason() *string {
	return m.Reason
}

// SetReason sets the Reason value and returns the updated UpdateTempTargetParams.
func (m *UpdateTempTargetParams) SetReason(v *string) {
	m.Reason = v
}

// GetRawJsonData returns the RawJsonData value.
func (m UpdateTempTargetParams) GetRawJsonData() *string {
	return m.RawJsonData
}

// SetRawJsonData sets the RawJsonData value and returns the updated UpdateTempTargetParams.
func (m *UpdateTempTargetParams) SetRawJsonData(v *string) {
	m.RawJsonData = v
}

// GetUpdatedByIdentityID returns the UpdatedByIdentityID value.
func (m UpdateTempTargetParams) GetUpdatedByIdentityID() *int64 {
	return m.UpdatedByIdentityID
}

// SetUpdatedByIdentityID sets the UpdatedByIdentityID value and returns the updated UpdateTempTargetParams.
func (m *UpdateTempTargetParams) SetUpdatedByIdentityID(v *int64) {
	m.UpdatedByIdentityID = v
}

// GetUpdatedByIdentityName returns the UpdatedByIdentityName value.
func (m UpdateTempTargetParams) GetUpdatedByIdentityName() *string {
	return m.UpdatedByIdentityName
}

// SetUpdatedByIdentityName sets the UpdatedByIdentityName value and returns the updated UpdateTempTargetParams.
func (m *UpdateTempTargetParams) SetUpdatedByIdentityName(v *string) {
	m.UpdatedByIdentityName = v
}

// GetID returns the ID value.
func (m UpdateTempTargetParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated UpdateTempTargetParams.
func (m *UpdateTempTargetParams) SetID(v string) {
	m.ID = v
}

// GetSkip returns the Skip value.
func (m AllFoodPresetsParams) GetSkip() int64 {
	return m.Skip
}

// SetSkip sets the Skip value and returns the updated AllFoodPresetsParams.
func (m *AllFoodPresetsParams) SetSkip(v int64) {
	m.Skip = v
}

// GetLimit returns the Limit value.
func (m AllFoodPresetsParams) GetLimit() int64 {
	return m.Limit
}

// SetLimit sets the Limit value and returns the updated AllFoodPresetsParams.
func (m *AllFoodPresetsParams) SetLimit(v int64) {
	m.Limit = v
}

// GetID returns the ID value.
func (m CreateFoodPresetParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated CreateFoodPresetParams.
func (m *CreateFoodPresetParams) SetID(v string) {
	m.ID = v
}

// GetName returns the Name value.
func (m CreateFoodPresetParams) GetName() string {
	return m.Name
}

// SetName sets the Name value and returns the updated CreateFoodPresetParams.
func (m *CreateFoodPresetParams) SetName(v string) {
	m.Name = v
}

// GetFoodType returns the FoodType value.
func (m CreateFoodPresetParams) GetFoodType() *string {
	return m.FoodType
}

// SetFoodType sets the FoodType value and returns the updated CreateFoodPresetParams.
func (m *CreateFoodPresetParams) SetFoodType(v *string) {
	m.FoodType = v
}

// GetCategory returns the Category value.
func (m CreateFoodPresetParams) GetCategory() *string {
	return m.Category
}

// SetCategory sets the Category value and returns the updated CreateFoodPresetParams.
func (m *CreateFoodPresetParams) SetCategory(v *string) {
	m.Category = v
}

// GetSubcategory returns the Subcategory value.
func (m CreateFoodPresetParams) GetSubcategory() *string {
	return m.Subcategory
}

// SetSubcategory sets the Subcategory value and returns the updated CreateFoodPresetParams.
func (m *CreateFoodPresetParams) SetSubcategory(v *string) {
	m.Subcategory = v
}

// GetAbsorptionTimeMinutes returns the AbsorptionTimeMinutes value.
func (m CreateFoodPresetParams) GetAbsorptionTimeMinutes() *int64 {
	return m.AbsorptionTimeMinutes
}

// SetAbsorptionTimeMinutes sets the AbsorptionTimeMinutes value and returns the updated CreateFoodPresetParams.
func (m *CreateFoodPresetParams) SetAbsorptionTimeMinutes(v *int64) {
	m.AbsorptionTimeMinutes = v
}

// GetPreBolus returns the PreBolus value.
func (m CreateFoodPresetParams) GetPreBolus() *float64 {
	return m.PreBolus
}

// SetPreBolus sets the PreBolus value and returns the updated CreateFoodPresetParams.
func (m *CreateFoodPresetParams) SetPreBolus(v *float64) {
	m.PreBolus = v
}

// GetPortions returns the Portions value.
func (m CreateFoodPresetParams) GetPortions() *float64 {
	return m.Portions
}

// SetPortions sets the Portions value and returns the updated CreateFoodPresetParams.
func (m *CreateFoodPresetParams) SetPortions(v *float64) {
	m.Portions = v
}

// GetCarbs returns the Carbs value.
func (m CreateFoodPresetParams) GetCarbs() *float64 {
	return m.Carbs
}

// SetCarbs sets the Carbs value and returns the updated CreateFoodPresetParams.
func (m *CreateFoodPresetParams) SetCarbs(v *float64) {
	m.Carbs = v
}

// GetFat returns the Fat value.
func (m CreateFoodPresetParams) GetFat() *float64 {
	return m.Fat
}

// SetFat sets the Fat value and returns the updated CreateFoodPresetParams.
func (m *CreateFoodPresetParams) SetFat(v *float64) {
	m.Fat = v
}

// GetProtein returns the Protein value.
func (m CreateFoodPresetParams) GetProtein() *float64 {
	return m.Protein
}

// SetProtein sets the Protein value and returns the updated CreateFoodPresetParams.
func (m *CreateFoodPresetParams) SetProtein(v *float64) {
	m.Protein = v
}

// GetFibre returns the Fibre value.
func (m CreateFoodPresetParams) GetFibre() *float64 {
	return m.Fibre
}

// SetFibre sets the Fibre value and returns the updated CreateFoodPresetParams.
func (m *CreateFoodPresetParams) SetFibre(v *float64) {
	m.Fibre = v
}

// GetEnergy returns the Energy value.
func (m CreateFoodPresetParams) GetEnergy() *float64 {
	return m.Energy
}

// SetEnergy sets the Energy value and returns the updated CreateFoodPresetParams.
func (m *CreateFoodPresetParams) SetEnergy(v *float64) {
	m.Energy = v
}

// GetRawJsonData returns the RawJsonData value.
func (m CreateFoodPresetParams) GetRawJsonData() *string {
	return m.RawJsonData
}

// SetRawJsonData sets the RawJsonData value and returns the updated CreateFoodPresetParams.
func (m *CreateFoodPresetParams) SetRawJsonData(v *string) {
	m.RawJsonData = v
}

// GetCreatedByAppName returns the CreatedByAppName value.
func (m CreateFoodPresetParams) GetCreatedByAppName() *string {
	return m.CreatedByAppName
}

// SetCreatedByAppName sets the CreatedByAppName value and returns the updated CreateFoodPresetParams.
func (m *CreateFoodPresetParams) SetCreatedByAppName(v *string) {
	m.CreatedByAppName = v
}

// GetCreatedByDeviceName returns the CreatedByDeviceName value.
func (m CreateFoodPresetParams) GetCreatedByDeviceName() *string {
	return m.CreatedByDeviceName
}

// SetCreatedByDeviceName sets the CreatedByDeviceName value and returns the updated CreateFoodPresetParams.
func (m *CreateFoodPresetParams) SetCreatedByDeviceName(v *string) {
	m.CreatedByDeviceName = v
}

// GetCreatedByIdentityName returns the CreatedByIdentityName value.
func (m CreateFoodPresetParams) GetCreatedByIdentityName() string {
	return m.CreatedByIdentityName
}

// SetCreatedByIdentityName sets the CreatedByIdentityName value and returns the updated CreateFoodPresetParams.
func (m *CreateFoodPresetParams) SetCreatedByIdentityName(v string) {
	m.CreatedByIdentityName = v
}

// GetCreatedByIdentityID returns the CreatedByIdentityID value.
func (m CreateFoodPresetParams) GetCreatedByIdentityID() *int64 {
	return m.CreatedByIdentityID
}

// SetCreatedByIdentityID sets the CreatedByIdentityID value and returns the updated CreateFoodPresetParams.
func (m *CreateFoodPresetParams) SetCreatedByIdentityID(v *int64) {
	m.CreatedByIdentityID = v
}

// GetID returns the ID value.
func (m GetFoodPresetParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated GetFoodPresetParams.
func (m *GetFoodPresetParams) SetID(v string) {
	m.ID = v
}

// GetID returns the ID value.
func (m HardDeleteFoodPresetParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated HardDeleteFoodPresetParams.
func (m *HardDeleteFoodPresetParams) SetID(v string) {
	m.ID = v
}

// GetFoodType returns the FoodType value.
func (m HardDeleteFoodPresetsByTypeParams) GetFoodType() *string {
	return m.FoodType
}

// SetFoodType sets the FoodType value and returns the updated HardDeleteFoodPresetsByTypeParams.
func (m *HardDeleteFoodPresetsByTypeParams) SetFoodType(v *string) {
	m.FoodType = v
}

// GetDateGTE returns the DateGTE value.
func (m SearchFoodPresetsParams) GetDateGTE() *int64 {
	return m.DateGTE
}

// SetDateGTE sets the DateGTE value and returns the updated SearchFoodPresetsParams.
func (m *SearchFoodPresetsParams) SetDateGTE(v *int64) {
	m.DateGTE = v
}

// GetDateLTE returns the DateLTE value.
func (m SearchFoodPresetsParams) GetDateLTE() *int64 {
	return m.DateLTE
}

// SetDateLTE sets the DateLTE value and returns the updated SearchFoodPresetsParams.
func (m *SearchFoodPresetsParams) SetDateLTE(v *int64) {
	m.DateLTE = v
}

// GetDateEQ returns the DateEQ value.
func (m SearchFoodPresetsParams) GetDateEQ() *int64 {
	return m.DateEQ
}

// SetDateEQ sets the DateEQ value and returns the updated SearchFoodPresetsParams.
func (m *SearchFoodPresetsParams) SetDateEQ(v *int64) {
	m.DateEQ = v
}

// GetDateGT returns the DateGT value.
func (m SearchFoodPresetsParams) GetDateGT() *int64 {
	return m.DateGT
}

// SetDateGT sets the DateGT value and returns the updated SearchFoodPresetsParams.
func (m *SearchFoodPresetsParams) SetDateGT(v *int64) {
	m.DateGT = v
}

// GetDateLT returns the DateLT value.
func (m SearchFoodPresetsParams) GetDateLT() *int64 {
	return m.DateLT
}

// SetDateLT sets the DateLT value and returns the updated SearchFoodPresetsParams.
func (m *SearchFoodPresetsParams) SetDateLT(v *int64) {
	m.DateLT = v
}

// GetCreatedAtGTE returns the CreatedAtGTE value.
func (m SearchFoodPresetsParams) GetCreatedAtGTE() *int64 {
	return m.CreatedAtGTE
}

// SetCreatedAtGTE sets the CreatedAtGTE value and returns the updated SearchFoodPresetsParams.
func (m *SearchFoodPresetsParams) SetCreatedAtGTE(v *int64) {
	m.CreatedAtGTE = v
}

// GetCreatedAtLTE returns the CreatedAtLTE value.
func (m SearchFoodPresetsParams) GetCreatedAtLTE() *int64 {
	return m.CreatedAtLTE
}

// SetCreatedAtLTE sets the CreatedAtLTE value and returns the updated SearchFoodPresetsParams.
func (m *SearchFoodPresetsParams) SetCreatedAtLTE(v *int64) {
	m.CreatedAtLTE = v
}

// GetCreatedAtGT returns the CreatedAtGT value.
func (m SearchFoodPresetsParams) GetCreatedAtGT() *int64 {
	return m.CreatedAtGT
}

// SetCreatedAtGT sets the CreatedAtGT value and returns the updated SearchFoodPresetsParams.
func (m *SearchFoodPresetsParams) SetCreatedAtGT(v *int64) {
	m.CreatedAtGT = v
}

// GetCreatedAtLT returns the CreatedAtLT value.
func (m SearchFoodPresetsParams) GetCreatedAtLT() *int64 {
	return m.CreatedAtLT
}

// SetCreatedAtLT sets the CreatedAtLT value and returns the updated SearchFoodPresetsParams.
func (m *SearchFoodPresetsParams) SetCreatedAtLT(v *int64) {
	m.CreatedAtLT = v
}

// GetUpdatedGTE returns the UpdatedGTE value.
func (m SearchFoodPresetsParams) GetUpdatedGTE() *int64 {
	return m.UpdatedGTE
}

// SetUpdatedGTE sets the UpdatedGTE value and returns the updated SearchFoodPresetsParams.
func (m *SearchFoodPresetsParams) SetUpdatedGTE(v *int64) {
	m.UpdatedGTE = v
}

// GetUpdatedLTE returns the UpdatedLTE value.
func (m SearchFoodPresetsParams) GetUpdatedLTE() *int64 {
	return m.UpdatedLTE
}

// SetUpdatedLTE sets the UpdatedLTE value and returns the updated SearchFoodPresetsParams.
func (m *SearchFoodPresetsParams) SetUpdatedLTE(v *int64) {
	m.UpdatedLTE = v
}

// GetUpdatedGT returns the UpdatedGT value.
func (m SearchFoodPresetsParams) GetUpdatedGT() *int64 {
	return m.UpdatedGT
}

// SetUpdatedGT sets the UpdatedGT value and returns the updated SearchFoodPresetsParams.
func (m *SearchFoodPresetsParams) SetUpdatedGT(v *int64) {
	m.UpdatedGT = v
}

// GetUpdatedLT returns the UpdatedLT value.
func (m SearchFoodPresetsParams) GetUpdatedLT() *int64 {
	return m.UpdatedLT
}

// SetUpdatedLT sets the UpdatedLT value and returns the updated SearchFoodPresetsParams.
func (m *SearchFoodPresetsParams) SetUpdatedLT(v *int64) {
	m.UpdatedLT = v
}

// GetIDEq returns the IDEq value.
func (m SearchFoodPresetsParams) GetIDEq() *string {
	return m.IDEq
}

// SetIDEq sets the IDEq value and returns the updated SearchFoodPresetsParams.
func (m *SearchFoodPresetsParams) SetIDEq(v *string) {
	m.IDEq = v
}

// GetIDIn returns the IDIn value.
func (m SearchFoodPresetsParams) GetIDIn() *string {
	return m.IDIn
}

// SetIDIn sets the IDIn value and returns the updated SearchFoodPresetsParams.
func (m *SearchFoodPresetsParams) SetIDIn(v *string) {
	m.IDIn = v
}

// GetNameEq returns the NameEq value.
func (m SearchFoodPresetsParams) GetNameEq() *string {
	return m.NameEq
}

// SetNameEq sets the NameEq value and returns the updated SearchFoodPresetsParams.
func (m *SearchFoodPresetsParams) SetNameEq(v *string) {
	m.NameEq = v
}

// GetDeviceEq returns the DeviceEq value.
func (m SearchFoodPresetsParams) GetDeviceEq() *string {
	return m.DeviceEq
}

// SetDeviceEq sets the DeviceEq value and returns the updated SearchFoodPresetsParams.
func (m *SearchFoodPresetsParams) SetDeviceEq(v *string) {
	m.DeviceEq = v
}

// GetAppEq returns the AppEq value.
func (m SearchFoodPresetsParams) GetAppEq() *string {
	return m.AppEq
}

// SetAppEq sets the AppEq value and returns the updated SearchFoodPresetsParams.
func (m *SearchFoodPresetsParams) SetAppEq(v *string) {
	m.AppEq = v
}

// GetTypeEq returns the TypeEq value.
func (m SearchFoodPresetsParams) GetTypeEq() *string {
	return m.TypeEq
}

// SetTypeEq sets the TypeEq value and returns the updated SearchFoodPresetsParams.
func (m *SearchFoodPresetsParams) SetTypeEq(v *string) {
	m.TypeEq = v
}

// GetTypeNeq returns the TypeNeq value.
func (m SearchFoodPresetsParams) GetTypeNeq() *string {
	return m.TypeNeq
}

// SetTypeNeq sets the TypeNeq value and returns the updated SearchFoodPresetsParams.
func (m *SearchFoodPresetsParams) SetTypeNeq(v *string) {
	m.TypeNeq = v
}

// GetTypeIn returns the TypeIn value.
func (m SearchFoodPresetsParams) GetTypeIn() *string {
	return m.TypeIn
}

// SetTypeIn sets the TypeIn value and returns the updated SearchFoodPresetsParams.
func (m *SearchFoodPresetsParams) SetTypeIn(v *string) {
	m.TypeIn = v
}

// GetTypeNin returns the TypeNin value.
func (m SearchFoodPresetsParams) GetTypeNin() *string {
	return m.TypeNin
}

// SetTypeNin sets the TypeNin value and returns the updated SearchFoodPresetsParams.
func (m *SearchFoodPresetsParams) SetTypeNin(v *string) {
	m.TypeNin = v
}

// GetCarbsEQ returns the CarbsEQ value.
func (m SearchFoodPresetsParams) GetCarbsEQ() *int64 {
	return m.CarbsEQ
}

// SetCarbsEQ sets the CarbsEQ value and returns the updated SearchFoodPresetsParams.
func (m *SearchFoodPresetsParams) SetCarbsEQ(v *int64) {
	m.CarbsEQ = v
}

// GetCarbsGTE returns the CarbsGTE value.
func (m SearchFoodPresetsParams) GetCarbsGTE() *int64 {
	return m.CarbsGTE
}

// SetCarbsGTE sets the CarbsGTE value and returns the updated SearchFoodPresetsParams.
func (m *SearchFoodPresetsParams) SetCarbsGTE(v *int64) {
	m.CarbsGTE = v
}

// GetCarbsGT returns the CarbsGT value.
func (m SearchFoodPresetsParams) GetCarbsGT() *int64 {
	return m.CarbsGT
}

// SetCarbsGT sets the CarbsGT value and returns the updated SearchFoodPresetsParams.
func (m *SearchFoodPresetsParams) SetCarbsGT(v *int64) {
	m.CarbsGT = v
}

// GetCarbsLTE returns the CarbsLTE value.
func (m SearchFoodPresetsParams) GetCarbsLTE() *int64 {
	return m.CarbsLTE
}

// SetCarbsLTE sets the CarbsLTE value and returns the updated SearchFoodPresetsParams.
func (m *SearchFoodPresetsParams) SetCarbsLTE(v *int64) {
	m.CarbsLTE = v
}

// GetCarbsLT returns the CarbsLT value.
func (m SearchFoodPresetsParams) GetCarbsLT() *int64 {
	return m.CarbsLT
}

// SetCarbsLT sets the CarbsLT value and returns the updated SearchFoodPresetsParams.
func (m *SearchFoodPresetsParams) SetCarbsLT(v *int64) {
	m.CarbsLT = v
}

// GetSkip returns the Skip value.
func (m SearchFoodPresetsParams) GetSkip() int64 {
	return m.Skip
}

// SetSkip sets the Skip value and returns the updated SearchFoodPresetsParams.
func (m *SearchFoodPresetsParams) SetSkip(v int64) {
	m.Skip = v
}

// GetLimit returns the Limit value.
func (m SearchFoodPresetsParams) GetLimit() int64 {
	return m.Limit
}

// SetLimit sets the Limit value and returns the updated SearchFoodPresetsParams.
func (m *SearchFoodPresetsParams) SetLimit(v int64) {
	m.Limit = v
}

// GetName returns the Name value.
func (m UpdateFoodPresetParams) GetName() *string {
	return m.Name
}

// SetName sets the Name value and returns the updated UpdateFoodPresetParams.
func (m *UpdateFoodPresetParams) SetName(v *string) {
	m.Name = v
}

// GetFoodType returns the FoodType value.
func (m UpdateFoodPresetParams) GetFoodType() *string {
	return m.FoodType
}

// SetFoodType sets the FoodType value and returns the updated UpdateFoodPresetParams.
func (m *UpdateFoodPresetParams) SetFoodType(v *string) {
	m.FoodType = v
}

// GetCategory returns the Category value.
func (m UpdateFoodPresetParams) GetCategory() *string {
	return m.Category
}

// SetCategory sets the Category value and returns the updated UpdateFoodPresetParams.
func (m *UpdateFoodPresetParams) SetCategory(v *string) {
	m.Category = v
}

// GetSubcategory returns the Subcategory value.
func (m UpdateFoodPresetParams) GetSubcategory() *string {
	return m.Subcategory
}

// SetSubcategory sets the Subcategory value and returns the updated UpdateFoodPresetParams.
func (m *UpdateFoodPresetParams) SetSubcategory(v *string) {
	m.Subcategory = v
}

// GetAbsorptionTimeMinutes returns the AbsorptionTimeMinutes value.
func (m UpdateFoodPresetParams) GetAbsorptionTimeMinutes() *int64 {
	return m.AbsorptionTimeMinutes
}

// SetAbsorptionTimeMinutes sets the AbsorptionTimeMinutes value and returns the updated UpdateFoodPresetParams.
func (m *UpdateFoodPresetParams) SetAbsorptionTimeMinutes(v *int64) {
	m.AbsorptionTimeMinutes = v
}

// GetPreBolus returns the PreBolus value.
func (m UpdateFoodPresetParams) GetPreBolus() *float64 {
	return m.PreBolus
}

// SetPreBolus sets the PreBolus value and returns the updated UpdateFoodPresetParams.
func (m *UpdateFoodPresetParams) SetPreBolus(v *float64) {
	m.PreBolus = v
}

// GetPortions returns the Portions value.
func (m UpdateFoodPresetParams) GetPortions() *float64 {
	return m.Portions
}

// SetPortions sets the Portions value and returns the updated UpdateFoodPresetParams.
func (m *UpdateFoodPresetParams) SetPortions(v *float64) {
	m.Portions = v
}

// GetCarbs returns the Carbs value.
func (m UpdateFoodPresetParams) GetCarbs() *float64 {
	return m.Carbs
}

// SetCarbs sets the Carbs value and returns the updated UpdateFoodPresetParams.
func (m *UpdateFoodPresetParams) SetCarbs(v *float64) {
	m.Carbs = v
}

// GetFat returns the Fat value.
func (m UpdateFoodPresetParams) GetFat() *float64 {
	return m.Fat
}

// SetFat sets the Fat value and returns the updated UpdateFoodPresetParams.
func (m *UpdateFoodPresetParams) SetFat(v *float64) {
	m.Fat = v
}

// GetProtein returns the Protein value.
func (m UpdateFoodPresetParams) GetProtein() *float64 {
	return m.Protein
}

// SetProtein sets the Protein value and returns the updated UpdateFoodPresetParams.
func (m *UpdateFoodPresetParams) SetProtein(v *float64) {
	m.Protein = v
}

// GetFibre returns the Fibre value.
func (m UpdateFoodPresetParams) GetFibre() *float64 {
	return m.Fibre
}

// SetFibre sets the Fibre value and returns the updated UpdateFoodPresetParams.
func (m *UpdateFoodPresetParams) SetFibre(v *float64) {
	m.Fibre = v
}

// GetEnergy returns the Energy value.
func (m UpdateFoodPresetParams) GetEnergy() *float64 {
	return m.Energy
}

// SetEnergy sets the Energy value and returns the updated UpdateFoodPresetParams.
func (m *UpdateFoodPresetParams) SetEnergy(v *float64) {
	m.Energy = v
}

// GetRawJsonData returns the RawJsonData value.
func (m UpdateFoodPresetParams) GetRawJsonData() *string {
	return m.RawJsonData
}

// SetRawJsonData sets the RawJsonData value and returns the updated UpdateFoodPresetParams.
func (m *UpdateFoodPresetParams) SetRawJsonData(v *string) {
	m.RawJsonData = v
}

// GetUpdatedByIdentityID returns the UpdatedByIdentityID value.
func (m UpdateFoodPresetParams) GetUpdatedByIdentityID() *int64 {
	return m.UpdatedByIdentityID
}

// SetUpdatedByIdentityID sets the UpdatedByIdentityID value and returns the updated UpdateFoodPresetParams.
func (m *UpdateFoodPresetParams) SetUpdatedByIdentityID(v *int64) {
	m.UpdatedByIdentityID = v
}

// GetUpdatedByIdentityName returns the UpdatedByIdentityName value.
func (m UpdateFoodPresetParams) GetUpdatedByIdentityName() *string {
	return m.UpdatedByIdentityName
}

// SetUpdatedByIdentityName sets the UpdatedByIdentityName value and returns the updated UpdateFoodPresetParams.
func (m *UpdateFoodPresetParams) SetUpdatedByIdentityName(v *string) {
	m.UpdatedByIdentityName = v
}

// GetID returns the ID value.
func (m UpdateFoodPresetParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated UpdateFoodPresetParams.
func (m *UpdateFoodPresetParams) SetID(v string) {
	m.ID = v
}

// GetSkip returns the Skip value.
func (m AllProfilePresetsParams) GetSkip() int64 {
	return m.Skip
}

// SetSkip sets the Skip value and returns the updated AllProfilePresetsParams.
func (m *AllProfilePresetsParams) SetSkip(v int64) {
	m.Skip = v
}

// GetLimit returns the Limit value.
func (m AllProfilePresetsParams) GetLimit() int64 {
	return m.Limit
}

// SetLimit sets the Limit value and returns the updated AllProfilePresetsParams.
func (m *AllProfilePresetsParams) SetLimit(v int64) {
	m.Limit = v
}

// GetID returns the ID value.
func (m CreateProfilePresetParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated CreateProfilePresetParams.
func (m *CreateProfilePresetParams) SetID(v string) {
	m.ID = v
}

// GetName returns the Name value.
func (m CreateProfilePresetParams) GetName() string {
	return m.Name
}

// SetName sets the Name value and returns the updated CreateProfilePresetParams.
func (m *CreateProfilePresetParams) SetName(v string) {
	m.Name = v
}

// GetProfileStoreID returns the ProfileStoreID value.
func (m CreateProfilePresetParams) GetProfileStoreID() string {
	return m.ProfileStoreID
}

// SetProfileStoreID sets the ProfileStoreID value and returns the updated CreateProfilePresetParams.
func (m *CreateProfilePresetParams) SetProfileStoreID(v string) {
	m.ProfileStoreID = v
}

// GetIsDefault returns the IsDefault value.
func (m CreateProfilePresetParams) GetIsDefault() bool {
	return m.IsDefault
}

// SetIsDefault sets the IsDefault value and returns the updated CreateProfilePresetParams.
func (m *CreateProfilePresetParams) SetIsDefault(v bool) {
	m.IsDefault = v
}

// GetDia returns the Dia value.
func (m CreateProfilePresetParams) GetDia() *float64 {
	return m.Dia
}

// SetDia sets the Dia value and returns the updated CreateProfilePresetParams.
func (m *CreateProfilePresetParams) SetDia(v *float64) {
	m.Dia = v
}

// GetUnits returns the Units value.
func (m CreateProfilePresetParams) GetUnits() *consts.Units {
	return m.Units
}

// SetUnits sets the Units value and returns the updated CreateProfilePresetParams.
func (m *CreateProfilePresetParams) SetUnits(v *consts.Units) {
	m.Units = v
}

// GetTimezone returns the Timezone value.
func (m CreateProfilePresetParams) GetTimezone() *string {
	return m.Timezone
}

// SetTimezone sets the Timezone value and returns the updated CreateProfilePresetParams.
func (m *CreateProfilePresetParams) SetTimezone(v *string) {
	m.Timezone = v
}

// GetCarbsHr returns the CarbsHr value.
func (m CreateProfilePresetParams) GetCarbsHr() *float64 {
	return m.CarbsHr
}

// SetCarbsHr sets the CarbsHr value and returns the updated CreateProfilePresetParams.
func (m *CreateProfilePresetParams) SetCarbsHr(v *float64) {
	m.CarbsHr = v
}

// GetDelay returns the Delay value.
func (m CreateProfilePresetParams) GetDelay() *float64 {
	return m.Delay
}

// SetDelay sets the Delay value and returns the updated CreateProfilePresetParams.
func (m *CreateProfilePresetParams) SetDelay(v *float64) {
	m.Delay = v
}

// GetCarbRatiosJson returns the CarbRatiosJson value.
func (m CreateProfilePresetParams) GetCarbRatiosJson() *string {
	return m.CarbRatiosJson
}

// SetCarbRatiosJson sets the CarbRatiosJson value and returns the updated CreateProfilePresetParams.
func (m *CreateProfilePresetParams) SetCarbRatiosJson(v *string) {
	m.CarbRatiosJson = v
}

// GetInsulinSensitivitiesJson returns the InsulinSensitivitiesJson value.
func (m CreateProfilePresetParams) GetInsulinSensitivitiesJson() *string {
	return m.InsulinSensitivitiesJson
}

// SetInsulinSensitivitiesJson sets the InsulinSensitivitiesJson value and returns the updated CreateProfilePresetParams.
func (m *CreateProfilePresetParams) SetInsulinSensitivitiesJson(v *string) {
	m.InsulinSensitivitiesJson = v
}

// GetBasalRatesJson returns the BasalRatesJson value.
func (m CreateProfilePresetParams) GetBasalRatesJson() *string {
	return m.BasalRatesJson
}

// SetBasalRatesJson sets the BasalRatesJson value and returns the updated CreateProfilePresetParams.
func (m *CreateProfilePresetParams) SetBasalRatesJson(v *string) {
	m.BasalRatesJson = v
}

// GetTargetsHighJson returns the TargetsHighJson value.
func (m CreateProfilePresetParams) GetTargetsHighJson() *string {
	return m.TargetsHighJson
}

// SetTargetsHighJson sets the TargetsHighJson value and returns the updated CreateProfilePresetParams.
func (m *CreateProfilePresetParams) SetTargetsHighJson(v *string) {
	m.TargetsHighJson = v
}

// GetTargetsLowJson returns the TargetsLowJson value.
func (m CreateProfilePresetParams) GetTargetsLowJson() *string {
	return m.TargetsLowJson
}

// SetTargetsLowJson sets the TargetsLowJson value and returns the updated CreateProfilePresetParams.
func (m *CreateProfilePresetParams) SetTargetsLowJson(v *string) {
	m.TargetsLowJson = v
}

// GetLoopSettingsJson returns the LoopSettingsJson value.
func (m CreateProfilePresetParams) GetLoopSettingsJson() *string {
	return m.LoopSettingsJson
}

// SetLoopSettingsJson sets the LoopSettingsJson value and returns the updated CreateProfilePresetParams.
func (m *CreateProfilePresetParams) SetLoopSettingsJson(v *string) {
	m.LoopSettingsJson = v
}

// GetRemoteSettingsJson returns the RemoteSettingsJson value.
func (m CreateProfilePresetParams) GetRemoteSettingsJson() *string {
	return m.RemoteSettingsJson
}

// SetRemoteSettingsJson sets the RemoteSettingsJson value and returns the updated CreateProfilePresetParams.
func (m *CreateProfilePresetParams) SetRemoteSettingsJson(v *string) {
	m.RemoteSettingsJson = v
}

// GetOccurredAt returns the OccurredAt value.
func (m CreateProfilePresetParams) GetOccurredAt() int64 {
	return m.OccurredAt
}

// SetOccurredAt sets the OccurredAt value and returns the updated CreateProfilePresetParams.
func (m *CreateProfilePresetParams) SetOccurredAt(v int64) {
	m.OccurredAt = v
}

// GetUtcOffsetMinutes returns the UtcOffsetMinutes value.
func (m CreateProfilePresetParams) GetUtcOffsetMinutes() *int64 {
	return m.UtcOffsetMinutes
}

// SetUtcOffsetMinutes sets the UtcOffsetMinutes value and returns the updated CreateProfilePresetParams.
func (m *CreateProfilePresetParams) SetUtcOffsetMinutes(v *int64) {
	m.UtcOffsetMinutes = v
}

// GetRawJsonData returns the RawJsonData value.
func (m CreateProfilePresetParams) GetRawJsonData() *string {
	return m.RawJsonData
}

// SetRawJsonData sets the RawJsonData value and returns the updated CreateProfilePresetParams.
func (m *CreateProfilePresetParams) SetRawJsonData(v *string) {
	m.RawJsonData = v
}

// GetCreatedByAppName returns the CreatedByAppName value.
func (m CreateProfilePresetParams) GetCreatedByAppName() *string {
	return m.CreatedByAppName
}

// SetCreatedByAppName sets the CreatedByAppName value and returns the updated CreateProfilePresetParams.
func (m *CreateProfilePresetParams) SetCreatedByAppName(v *string) {
	m.CreatedByAppName = v
}

// GetCreatedByDeviceName returns the CreatedByDeviceName value.
func (m CreateProfilePresetParams) GetCreatedByDeviceName() *string {
	return m.CreatedByDeviceName
}

// SetCreatedByDeviceName sets the CreatedByDeviceName value and returns the updated CreateProfilePresetParams.
func (m *CreateProfilePresetParams) SetCreatedByDeviceName(v *string) {
	m.CreatedByDeviceName = v
}

// GetCreatedByIdentityName returns the CreatedByIdentityName value.
func (m CreateProfilePresetParams) GetCreatedByIdentityName() string {
	return m.CreatedByIdentityName
}

// SetCreatedByIdentityName sets the CreatedByIdentityName value and returns the updated CreateProfilePresetParams.
func (m *CreateProfilePresetParams) SetCreatedByIdentityName(v string) {
	m.CreatedByIdentityName = v
}

// GetCreatedByIdentityID returns the CreatedByIdentityID value.
func (m CreateProfilePresetParams) GetCreatedByIdentityID() *int64 {
	return m.CreatedByIdentityID
}

// SetCreatedByIdentityID sets the CreatedByIdentityID value and returns the updated CreateProfilePresetParams.
func (m *CreateProfilePresetParams) SetCreatedByIdentityID(v *int64) {
	m.CreatedByIdentityID = v
}

// GetID returns the ID value.
func (m GetProfilePresetParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated GetProfilePresetParams.
func (m *GetProfilePresetParams) SetID(v string) {
	m.ID = v
}

// GetProfileStoreID returns the ProfileStoreID value.
func (m GetProfileStoreParams) GetProfileStoreID() string {
	return m.ProfileStoreID
}

// SetProfileStoreID sets the ProfileStoreID value and returns the updated GetProfileStoreParams.
func (m *GetProfileStoreParams) SetProfileStoreID(v string) {
	m.ProfileStoreID = v
}

// GetID returns the ID value.
func (m HardDeleteProfilePresetParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated HardDeleteProfilePresetParams.
func (m *HardDeleteProfilePresetParams) SetID(v string) {
	m.ID = v
}

// GetProfileStoreID returns the ProfileStoreID value.
func (m HardDeleteProfileStoreParams) GetProfileStoreID() string {
	return m.ProfileStoreID
}

// SetProfileStoreID sets the ProfileStoreID value and returns the updated HardDeleteProfileStoreParams.
func (m *HardDeleteProfileStoreParams) SetProfileStoreID(v string) {
	m.ProfileStoreID = v
}

// GetDateGTE returns the DateGTE value.
func (m SearchProfileStoresParams) GetDateGTE() *int64 {
	return m.DateGTE
}

// SetDateGTE sets the DateGTE value and returns the updated SearchProfileStoresParams.
func (m *SearchProfileStoresParams) SetDateGTE(v *int64) {
	m.DateGTE = v
}

// GetDateLTE returns the DateLTE value.
func (m SearchProfileStoresParams) GetDateLTE() *int64 {
	return m.DateLTE
}

// SetDateLTE sets the DateLTE value and returns the updated SearchProfileStoresParams.
func (m *SearchProfileStoresParams) SetDateLTE(v *int64) {
	m.DateLTE = v
}

// GetDateEQ returns the DateEQ value.
func (m SearchProfileStoresParams) GetDateEQ() *int64 {
	return m.DateEQ
}

// SetDateEQ sets the DateEQ value and returns the updated SearchProfileStoresParams.
func (m *SearchProfileStoresParams) SetDateEQ(v *int64) {
	m.DateEQ = v
}

// GetDateGT returns the DateGT value.
func (m SearchProfileStoresParams) GetDateGT() *int64 {
	return m.DateGT
}

// SetDateGT sets the DateGT value and returns the updated SearchProfileStoresParams.
func (m *SearchProfileStoresParams) SetDateGT(v *int64) {
	m.DateGT = v
}

// GetDateLT returns the DateLT value.
func (m SearchProfileStoresParams) GetDateLT() *int64 {
	return m.DateLT
}

// SetDateLT sets the DateLT value and returns the updated SearchProfileStoresParams.
func (m *SearchProfileStoresParams) SetDateLT(v *int64) {
	m.DateLT = v
}

// GetCreatedAtGTE returns the CreatedAtGTE value.
func (m SearchProfileStoresParams) GetCreatedAtGTE() *int64 {
	return m.CreatedAtGTE
}

// SetCreatedAtGTE sets the CreatedAtGTE value and returns the updated SearchProfileStoresParams.
func (m *SearchProfileStoresParams) SetCreatedAtGTE(v *int64) {
	m.CreatedAtGTE = v
}

// GetCreatedAtLTE returns the CreatedAtLTE value.
func (m SearchProfileStoresParams) GetCreatedAtLTE() *int64 {
	return m.CreatedAtLTE
}

// SetCreatedAtLTE sets the CreatedAtLTE value and returns the updated SearchProfileStoresParams.
func (m *SearchProfileStoresParams) SetCreatedAtLTE(v *int64) {
	m.CreatedAtLTE = v
}

// GetCreatedAtGT returns the CreatedAtGT value.
func (m SearchProfileStoresParams) GetCreatedAtGT() *int64 {
	return m.CreatedAtGT
}

// SetCreatedAtGT sets the CreatedAtGT value and returns the updated SearchProfileStoresParams.
func (m *SearchProfileStoresParams) SetCreatedAtGT(v *int64) {
	m.CreatedAtGT = v
}

// GetCreatedAtLT returns the CreatedAtLT value.
func (m SearchProfileStoresParams) GetCreatedAtLT() *int64 {
	return m.CreatedAtLT
}

// SetCreatedAtLT sets the CreatedAtLT value and returns the updated SearchProfileStoresParams.
func (m *SearchProfileStoresParams) SetCreatedAtLT(v *int64) {
	m.CreatedAtLT = v
}

// GetUpdatedGTE returns the UpdatedGTE value.
func (m SearchProfileStoresParams) GetUpdatedGTE() *int64 {
	return m.UpdatedGTE
}

// SetUpdatedGTE sets the UpdatedGTE value and returns the updated SearchProfileStoresParams.
func (m *SearchProfileStoresParams) SetUpdatedGTE(v *int64) {
	m.UpdatedGTE = v
}

// GetUpdatedLTE returns the UpdatedLTE value.
func (m SearchProfileStoresParams) GetUpdatedLTE() *int64 {
	return m.UpdatedLTE
}

// SetUpdatedLTE sets the UpdatedLTE value and returns the updated SearchProfileStoresParams.
func (m *SearchProfileStoresParams) SetUpdatedLTE(v *int64) {
	m.UpdatedLTE = v
}

// GetUpdatedGT returns the UpdatedGT value.
func (m SearchProfileStoresParams) GetUpdatedGT() *int64 {
	return m.UpdatedGT
}

// SetUpdatedGT sets the UpdatedGT value and returns the updated SearchProfileStoresParams.
func (m *SearchProfileStoresParams) SetUpdatedGT(v *int64) {
	m.UpdatedGT = v
}

// GetUpdatedLT returns the UpdatedLT value.
func (m SearchProfileStoresParams) GetUpdatedLT() *int64 {
	return m.UpdatedLT
}

// SetUpdatedLT sets the UpdatedLT value and returns the updated SearchProfileStoresParams.
func (m *SearchProfileStoresParams) SetUpdatedLT(v *int64) {
	m.UpdatedLT = v
}

// GetIDEq returns the IDEq value.
func (m SearchProfileStoresParams) GetIDEq() *string {
	return m.IDEq
}

// SetIDEq sets the IDEq value and returns the updated SearchProfileStoresParams.
func (m *SearchProfileStoresParams) SetIDEq(v *string) {
	m.IDEq = v
}

// GetIDIn returns the IDIn value.
func (m SearchProfileStoresParams) GetIDIn() *string {
	return m.IDIn
}

// SetIDIn sets the IDIn value and returns the updated SearchProfileStoresParams.
func (m *SearchProfileStoresParams) SetIDIn(v *string) {
	m.IDIn = v
}

// GetNameEq returns the NameEq value.
func (m SearchProfileStoresParams) GetNameEq() *string {
	return m.NameEq
}

// SetNameEq sets the NameEq value and returns the updated SearchProfileStoresParams.
func (m *SearchProfileStoresParams) SetNameEq(v *string) {
	m.NameEq = v
}

// GetDeviceEq returns the DeviceEq value.
func (m SearchProfileStoresParams) GetDeviceEq() *string {
	return m.DeviceEq
}

// SetDeviceEq sets the DeviceEq value and returns the updated SearchProfileStoresParams.
func (m *SearchProfileStoresParams) SetDeviceEq(v *string) {
	m.DeviceEq = v
}

// GetAppEq returns the AppEq value.
func (m SearchProfileStoresParams) GetAppEq() *string {
	return m.AppEq
}

// SetAppEq sets the AppEq value and returns the updated SearchProfileStoresParams.
func (m *SearchProfileStoresParams) SetAppEq(v *string) {
	m.AppEq = v
}

// GetSkip returns the Skip value.
func (m SearchProfileStoresParams) GetSkip() int64 {
	return m.Skip
}

// SetSkip sets the Skip value and returns the updated SearchProfileStoresParams.
func (m *SearchProfileStoresParams) SetSkip(v int64) {
	m.Skip = v
}

// GetLimit returns the Limit value.
func (m SearchProfileStoresParams) GetLimit() int64 {
	return m.Limit
}

// SetLimit sets the Limit value and returns the updated SearchProfileStoresParams.
func (m *SearchProfileStoresParams) SetLimit(v int64) {
	m.Limit = v
}

// GetUntil returns the Until value.
func (m SearchSummaryDefaultProfilesParams) GetUntil() int64 {
	return m.Until
}

// SetUntil sets the Until value and returns the updated SearchSummaryDefaultProfilesParams.
func (m *SearchSummaryDefaultProfilesParams) SetUntil(v int64) {
	m.Until = v
}

// GetSince returns the Since value.
func (m SearchSummaryDefaultProfilesParams) GetSince() int64 {
	return m.Since
}

// SetSince sets the Since value and returns the updated SearchSummaryDefaultProfilesParams.
func (m *SearchSummaryDefaultProfilesParams) SetSince(v int64) {
	m.Since = v
}

// GetLimit returns the Limit value.
func (m SearchSummaryDefaultProfilesParams) GetLimit() int64 {
	return m.Limit
}

// SetLimit sets the Limit value and returns the updated SearchSummaryDefaultProfilesParams.
func (m *SearchSummaryDefaultProfilesParams) SetLimit(v int64) {
	m.Limit = v
}

// GetName returns the Name value.
func (m UpdateProfilePresetParams) GetName() *string {
	return m.Name
}

// SetName sets the Name value and returns the updated UpdateProfilePresetParams.
func (m *UpdateProfilePresetParams) SetName(v *string) {
	m.Name = v
}

// GetRawJsonData returns the RawJsonData value.
func (m UpdateProfilePresetParams) GetRawJsonData() *string {
	return m.RawJsonData
}

// SetRawJsonData sets the RawJsonData value and returns the updated UpdateProfilePresetParams.
func (m *UpdateProfilePresetParams) SetRawJsonData(v *string) {
	m.RawJsonData = v
}

// GetUpdatedByIdentityID returns the UpdatedByIdentityID value.
func (m UpdateProfilePresetParams) GetUpdatedByIdentityID() *int64 {
	return m.UpdatedByIdentityID
}

// SetUpdatedByIdentityID sets the UpdatedByIdentityID value and returns the updated UpdateProfilePresetParams.
func (m *UpdateProfilePresetParams) SetUpdatedByIdentityID(v *int64) {
	m.UpdatedByIdentityID = v
}

// GetUpdatedByIdentityName returns the UpdatedByIdentityName value.
func (m UpdateProfilePresetParams) GetUpdatedByIdentityName() *string {
	return m.UpdatedByIdentityName
}

// SetUpdatedByIdentityName sets the UpdatedByIdentityName value and returns the updated UpdateProfilePresetParams.
func (m *UpdateProfilePresetParams) SetUpdatedByIdentityName(v *string) {
	m.UpdatedByIdentityName = v
}

// GetID returns the ID value.
func (m UpdateProfilePresetParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated UpdateProfilePresetParams.
func (m *UpdateProfilePresetParams) SetID(v string) {
	m.ID = v
}

// GetPeriod returns the Period value.
func (m DeleteSummaryPeriodsParams) GetPeriod() *string {
	return m.Period
}

// SetPeriod sets the Period value and returns the updated DeleteSummaryPeriodsParams.
func (m *DeleteSummaryPeriodsParams) SetPeriod(v *string) {
	m.Period = v
}

// GetPeriod returns the Period value.
func (m GetSummaryPeriodParams) GetPeriod() string {
	return m.Period
}

// SetPeriod sets the Period value and returns the updated GetSummaryPeriodParams.
func (m *GetSummaryPeriodParams) SetPeriod(v string) {
	m.Period = v
}

// GetDate returns the Date value.
func (m GetSummaryPeriodParams) GetDate() string {
	return m.Date
}

// SetDate sets the Date value and returns the updated GetSummaryPeriodParams.
func (m *GetSummaryPeriodParams) SetDate(v string) {
	m.Date = v
}

// GetPeriod returns the Period value.
func (m HardDeleteSummaryPeriodParams) GetPeriod() string {
	return m.Period
}

// SetPeriod sets the Period value and returns the updated HardDeleteSummaryPeriodParams.
func (m *HardDeleteSummaryPeriodParams) SetPeriod(v string) {
	m.Period = v
}

// GetDate returns the Date value.
func (m HardDeleteSummaryPeriodParams) GetDate() string {
	return m.Date
}

// SetDate sets the Date value and returns the updated HardDeleteSummaryPeriodParams.
func (m *HardDeleteSummaryPeriodParams) SetDate(v string) {
	m.Date = v
}

// GetPeriod returns the Period value.
func (m SearchSummaryPeriodsParams) GetPeriod() string {
	return m.Period
}

// SetPeriod sets the Period value and returns the updated SearchSummaryPeriodsParams.
func (m *SearchSummaryPeriodsParams) SetPeriod(v string) {
	m.Period = v
}

// GetDateGTE returns the DateGTE value.
func (m SearchSummaryPeriodsParams) GetDateGTE() *string {
	return m.DateGTE
}

// SetDateGTE sets the DateGTE value and returns the updated SearchSummaryPeriodsParams.
func (m *SearchSummaryPeriodsParams) SetDateGTE(v *string) {
	m.DateGTE = v
}

// GetDateLTE returns the DateLTE value.
func (m SearchSummaryPeriodsParams) GetDateLTE() *string {
	return m.DateLTE
}

// SetDateLTE sets the DateLTE value and returns the updated SearchSummaryPeriodsParams.
func (m *SearchSummaryPeriodsParams) SetDateLTE(v *string) {
	m.DateLTE = v
}

// GetSkip returns the Skip value.
func (m SearchSummaryPeriodsParams) GetSkip() int64 {
	return m.Skip
}

// SetSkip sets the Skip value and returns the updated SearchSummaryPeriodsParams.
func (m *SearchSummaryPeriodsParams) SetSkip(v int64) {
	m.Skip = v
}

// GetLimit returns the Limit value.
func (m SearchSummaryPeriodsParams) GetLimit() int64 {
	return m.Limit
}

// SetLimit sets the Limit value and returns the updated SearchSummaryPeriodsParams.
func (m *SearchSummaryPeriodsParams) SetLimit(v int64) {
	m.Limit = v
}

// GetPeriod returns the Period value.
func (m UpsertSummaryPeriodParams) GetPeriod() string {
	return m.Period
}

// SetPeriod sets the Period value and returns the updated UpsertSummaryPeriodParams.
func (m *UpsertSummaryPeriodParams) SetPeriod(v string) {
	m.Period = v
}

// GetDate returns the Date value.
func (m UpsertSummaryPeriodParams) GetDate() string {
	return m.Date
}

// SetDate sets the Date value and returns the updated UpsertSummaryPeriodParams.
func (m *UpsertSummaryPeriodParams) SetDate(v string) {
	m.Date = v
}

// GetTimezone returns the Timezone value.
func (m UpsertSummaryPeriodParams) GetTimezone() string {
	return m.Timezone
}

// SetTimezone sets the Timezone value and returns the updated UpsertSummaryPeriodParams.
func (m *UpsertSummaryPeriodParams) SetTimezone(v string) {
	m.Timezone = v
}

// GetStartAt returns the StartAt value.
func (m UpsertSummaryPeriodParams) GetStartAt() int64 {
	return m.StartAt
}

// SetStartAt sets the StartAt value and returns the updated UpsertSummaryPeriodParams.
func (m *UpsertSummaryPeriodParams) SetStartAt(v int64) {
	m.StartAt = v
}

// GetEndAt returns the EndAt value.
func (m UpsertSummaryPeriodParams) GetEndAt() int64 {
	return m.EndAt
}

// SetEndAt sets the EndAt value and returns the updated UpsertSummaryPeriodParams.
func (m *UpsertSummaryPeriodParams) SetEndAt(v int64) {
	m.EndAt = v
}

// GetDays returns the Days value.
func (m UpsertSummaryPeriodParams) GetDays() int64 {
	return m.Days
}

// SetDays sets the Days value and returns the updated UpsertSummaryPeriodParams.
func (m *UpsertSummaryPeriodParams) SetDays(v int64) {
	m.Days = v
}

// GetTimeInRange returns the TimeInRange value.
func (m UpsertSummaryPeriodParams) GetTimeInRange() *float64 {
	return m.TimeInRange
}

// SetTimeInRange sets the TimeInRange value and returns the updated UpsertSummaryPeriodParams.
func (m *UpsertSummaryPeriodParams) SetTimeInRange(v *float64) {
	m.TimeInRange = v
}

// GetTotalDailyDose returns the TotalDailyDose value.
func (m UpsertSummaryPeriodParams) GetTotalDailyDose() float64 {
	return m.TotalDailyDose
}

// SetTotalDailyDose sets the TotalDailyDose value and returns the updated UpsertSummaryPeriodParams.
func (m *UpsertSummaryPeriodParams) SetTotalDailyDose(v float64) {
	m.TotalDailyDose = v
}

// GetTotalDailyCarbs returns the TotalDailyCarbs value.
func (m UpsertSummaryPeriodParams) GetTotalDailyCarbs() float64 {
	return m.TotalDailyCarbs
}

// SetTotalDailyCarbs sets the TotalDailyCarbs value and returns the updated UpsertSummaryPeriodParams.
func (m *UpsertSummaryPeriodParams) SetTotalDailyCarbs(v float64) {
	m.TotalDailyCarbs = v
}

// GetGlucoseMgdlJson returns the GlucoseMgdlJson value.
func (m UpsertSummaryPeriodParams) GetGlucoseMgdlJson() string {
	return m.GlucoseMgdlJson
}

// SetGlucoseMgdlJson sets the GlucoseMgdlJson value and returns the updated UpsertSummaryPeriodParams.
func (m *UpsertSummaryPeriodParams) SetGlucoseMgdlJson(v string) {
	m.GlucoseMgdlJson = v
}

// GetAgpMgdlJson returns the AgpMgdlJson value.
func (m UpsertSummaryPeriodParams) GetAgpMgdlJson() *string {
	return m.AgpMgdlJson
}

// SetAgpMgdlJson sets the AgpMgdlJson value and returns the updated UpsertSummaryPeriodParams.
func (m *UpsertSummaryPeriodParams) SetAgpMgdlJson(v *string) {
	m.AgpMgdlJson = v
}

// GetBefore returns the Before value.
func (m ClearActivitiesRawJSONParams) GetBefore() int64 {
	return m.Before
}

// SetBefore sets the Before value and returns the updated ClearActivitiesRawJSONParams.
func (m *ClearActivitiesRawJSONParams) SetBefore(v int64) {
	m.Before = v
}

// GetBefore returns the Before value.
func (m ClearComponentEventsRawJSONParams) GetBefore() int64 {
	return m.Before
}

// SetBefore sets the Before value and returns the updated ClearComponentEventsRawJSONParams.
func (m *ClearComponentEventsRawJSONParams) SetBefore(v int64) {
	m.Before = v
}

// GetBefore returns the Before value.
func (m ClearDeviceReportsRawJSONParams) GetBefore() int64 {
	return m.Before
}

// SetBefore sets the Before value and returns the updated ClearDeviceReportsRawJSONParams.
func (m *ClearDeviceReportsRawJSONParams) SetBefore(v int64) {
	m.Before = v
}

// GetBefore returns the Before value.
func (m ClearEventGroupsRawJSONParams) GetBefore() int64 {
	return m.Before
}

// SetBefore sets the Before value and returns the updated ClearEventGroupsRawJSONParams.
func (m *ClearEventGroupsRawJSONParams) SetBefore(v int64) {
	m.Before = v
}

// GetBefore returns the Before value.
func (m ClearFoodIntakesRawJSONParams) GetBefore() int64 {
	return m.Before
}

// SetBefore sets the Before value and returns the updated ClearFoodIntakesRawJSONParams.
func (m *ClearFoodIntakesRawJSONParams) SetBefore(v int64) {
	m.Before = v
}

// GetBefore returns the Before value.
func (m ClearFoodPresetsRawJSONParams) GetBefore() int64 {
	return m.Before
}

// SetBefore sets the Before value and returns the updated ClearFoodPresetsRawJSONParams.
func (m *ClearFoodPresetsRawJSONParams) SetBefore(v int64) {
	m.Before = v
}

// GetBefore returns the Before value.
func (m ClearGlucoseReadingsRawJSONParams) GetBefore() int64 {
	return m.Before
}

// SetBefore sets the Before value and returns the updated ClearGlucoseReadingsRawJSONParams.
func (m *ClearGlucoseReadingsRawJSONParams) SetBefore(v int64) {
	m.Before = v
}

// GetBefore returns the Before value.
func (m ClearInsulinDosesRawJSONParams) GetBefore() int64 {
	return m.Before
}

// SetBefore sets the Before value and returns the updated ClearInsulinDosesRawJSONParams.
func (m *ClearInsulinDosesRawJSONParams) SetBefore(v int64) {
	m.Before = v
}

// GetBefore returns the Before value.
func (m ClearNotificationEventsRawJSONParams) GetBefore() int64 {
	return m.Before
}

// SetBefore sets the Before value and returns the updated ClearNotificationEventsRawJSONParams.
func (m *ClearNotificationEventsRawJSONParams) SetBefore(v int64) {
	m.Before = v
}

// GetBefore returns the Before value.
func (m ClearProfilePresetsRawJSONParams) GetBefore() int64 {
	return m.Before
}

// SetBefore sets the Before value and returns the updated ClearProfilePresetsRawJSONParams.
func (m *ClearProfilePresetsRawJSONParams) SetBefore(v int64) {
	m.Before = v
}

// GetBefore returns the Before value.
func (m ClearProfileSwitchesRawJSONParams) GetBefore() int64 {
	return m.Before
}

// SetBefore sets the Before value and returns the updated ClearProfileSwitchesRawJSONParams.
func (m *ClearProfileSwitchesRawJSONParams) SetBefore(v int64) {
	m.Before = v
}

// GetBefore returns the Before value.
func (m ClearTempBasalsRawJSONParams) GetBefore() int64 {
	return m.Before
}

// SetBefore sets the Before value and returns the updated ClearTempBasalsRawJSONParams.
func (m *ClearTempBasalsRawJSONParams) SetBefore(v int64) {
	m.Before = v
}

// GetBefore returns the Before value.
func (m ClearTempTargetsRawJSONParams) GetBefore() int64 {
	return m.Before
}

// SetBefore sets the Before value and returns the updated ClearTempTargetsRawJSONParams.
func (m *ClearTempTargetsRawJSONParams) SetBefore(v int64) {
	m.Before = v
}

