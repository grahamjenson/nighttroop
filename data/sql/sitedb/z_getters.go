
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

// GetCreatedByIdentityName returns the CreatedByIdentityName value.
func (m Activity) GetCreatedByIdentityName() string {
	return m.CreatedByIdentityName
}

// SetCreatedByIdentityName sets the CreatedByIdentityName value and returns the updated Activity.
func (m *Activity) SetCreatedByIdentityName(v string) {
	m.CreatedByIdentityName = v
}

// GetDate returns the Date value.
func (m ColdDay) GetDate() string {
	return m.Date
}

// GetCreatedAt returns the CreatedAt value.
func (m ColdDay) GetCreatedAt() int64 {
	return m.CreatedAt
}

// SetCreatedAt sets the CreatedAt value and returns the updated ColdDay.
func (m *ColdDay) SetCreatedAt(v int64) {
	m.CreatedAt = v
}

// GetUpdatedAt returns the UpdatedAt value.
func (m ColdDay) GetUpdatedAt() int64 {
	return m.UpdatedAt
}

// SetUpdatedAt sets the UpdatedAt value and returns the updated ColdDay.
func (m *ColdDay) SetUpdatedAt(v int64) {
	m.UpdatedAt = v
}

// GetID returns the ID value.
func (m DeviceReport) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated DeviceReport.
func (m *DeviceReport) SetID(v string) {
	m.ID = v
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

// GetCreatedByIdentityName returns the CreatedByIdentityName value.
func (m FoodPreset) GetCreatedByIdentityName() string {
	return m.CreatedByIdentityName
}

// SetCreatedByIdentityName sets the CreatedByIdentityName value and returns the updated FoodPreset.
func (m *FoodPreset) SetCreatedByIdentityName(v string) {
	m.CreatedByIdentityName = v
}

// GetKey returns the Key value.
func (m GeneralSetting) GetKey() string {
	return m.Key
}

// GetType returns the Type value.
func (m GeneralSetting) GetType() string {
	return m.Type
}

// GetCreatedByIdentityName returns the CreatedByIdentityName value.
func (m GeneralSetting) GetCreatedByIdentityName() string {
	return m.CreatedByIdentityName
}

// SetCreatedByIdentityName sets the CreatedByIdentityName value and returns the updated GeneralSetting.
func (m *GeneralSetting) SetCreatedByIdentityName(v string) {
	m.CreatedByIdentityName = v
}

// GetCreatedAt returns the CreatedAt value.
func (m GeneralSetting) GetCreatedAt() int64 {
	return m.CreatedAt
}

// SetCreatedAt sets the CreatedAt value and returns the updated GeneralSetting.
func (m *GeneralSetting) SetCreatedAt(v int64) {
	m.CreatedAt = v
}

// GetUpdatedAt returns the UpdatedAt value.
func (m GeneralSetting) GetUpdatedAt() int64 {
	return m.UpdatedAt
}

// SetUpdatedAt sets the UpdatedAt value and returns the updated GeneralSetting.
func (m *GeneralSetting) SetUpdatedAt(v int64) {
	m.UpdatedAt = v
}

// GetID returns the ID value.
func (m GlucoseReading) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated GlucoseReading.
func (m *GlucoseReading) SetID(v string) {
	m.ID = v
}

// SetReadingType sets the ReadingType value and returns the updated GlucoseReading.
func (m *GlucoseReading) SetReadingType(v string) {
	m.ReadingType = v
}

// GetUnits returns the Units value.
func (m GlucoseReading) GetUnits() *consts.Units {
	return m.Units
}

// GetDelta returns the Delta value.
func (m GlucoseReading) GetDelta() *float64 {
	return m.Delta
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

// SetReadingType sets the ReadingType value and returns the updated GlucoseReadingsVw.
func (m *GlucoseReadingsVw) SetReadingType(v *string) {
	m.ReadingType = v
}

// GetUnits returns the Units value.
func (m GlucoseReadingsVw) GetUnits() *consts.Units {
	return m.Units
}

// GetDelta returns the Delta value.
func (m GlucoseReadingsVw) GetDelta() *float64 {
	return m.Delta
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

// GetIcfgJson returns the IcfgJson value.
func (m InsulinDose) GetIcfgJson() *string {
	return m.IcfgJson
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

// GetIcfgJson returns the IcfgJson value.
func (m InsulinDosesVw) GetIcfgJson() *string {
	return m.IcfgJson
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

// GetCreatedByIdentityName returns the CreatedByIdentityName value.
func (m InsulinDosesVw) GetCreatedByIdentityName() *string {
	return m.CreatedByIdentityName
}

// SetCreatedByIdentityName sets the CreatedByIdentityName value and returns the updated InsulinDosesVw.
func (m *InsulinDosesVw) SetCreatedByIdentityName(v *string) {
	m.CreatedByIdentityName = v
}

// GetID returns the ID value.
func (m LoopSetting) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated LoopSetting.
func (m *LoopSetting) SetID(v string) {
	m.ID = v
}

// GetApp returns the App value.
func (m LoopSetting) GetApp() string {
	return m.App
}

// GetSource returns the Source value.
func (m LoopSetting) GetSource() string {
	return m.Source
}

// GetCreatedByDeviceName returns the CreatedByDeviceName value.
func (m LoopSetting) GetCreatedByDeviceName() *string {
	return m.CreatedByDeviceName
}

// SetCreatedByDeviceName sets the CreatedByDeviceName value and returns the updated LoopSetting.
func (m *LoopSetting) SetCreatedByDeviceName(v *string) {
	m.CreatedByDeviceName = v
}

// GetCreatedAt returns the CreatedAt value.
func (m LoopSetting) GetCreatedAt() int64 {
	return m.CreatedAt
}

// SetCreatedAt sets the CreatedAt value and returns the updated LoopSetting.
func (m *LoopSetting) SetCreatedAt(v int64) {
	m.CreatedAt = v
}

// GetID returns the ID value.
func (m NotificationEvent) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated NotificationEvent.
func (m *NotificationEvent) SetID(v string) {
	m.ID = v
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

// GetDia returns the Dia value.
func (m ProfilePreset) GetDia() *float64 {
	return m.Dia
}

// GetUnits returns the Units value.
func (m ProfilePreset) GetUnits() *consts.Units {
	return m.Units
}

// GetTimezone returns the Timezone value.
func (m ProfilePreset) GetTimezone() *string {
	return m.Timezone
}

// GetCarbRatiosJson returns the CarbRatiosJson value.
func (m ProfilePreset) GetCarbRatiosJson() *string {
	return m.CarbRatiosJson
}

// GetInsulinSensitivitiesJson returns the InsulinSensitivitiesJson value.
func (m ProfilePreset) GetInsulinSensitivitiesJson() *string {
	return m.InsulinSensitivitiesJson
}

// GetBasalRatesJson returns the BasalRatesJson value.
func (m ProfilePreset) GetBasalRatesJson() *string {
	return m.BasalRatesJson
}

// GetTargetsHighJson returns the TargetsHighJson value.
func (m ProfilePreset) GetTargetsHighJson() *string {
	return m.TargetsHighJson
}

// GetTargetsLowJson returns the TargetsLowJson value.
func (m ProfilePreset) GetTargetsLowJson() *string {
	return m.TargetsLowJson
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

// GetIcfgJson returns the IcfgJson value.
func (m ProfileSwitch) GetIcfgJson() *string {
	return m.IcfgJson
}

// GetDia returns the Dia value.
func (m ProfileSwitch) GetDia() *float64 {
	return m.Dia
}

// GetUnits returns the Units value.
func (m ProfileSwitch) GetUnits() *consts.Units {
	return m.Units
}

// GetTimezone returns the Timezone value.
func (m ProfileSwitch) GetTimezone() *string {
	return m.Timezone
}

// GetCarbRatiosJson returns the CarbRatiosJson value.
func (m ProfileSwitch) GetCarbRatiosJson() *string {
	return m.CarbRatiosJson
}

// GetInsulinSensitivitiesJson returns the InsulinSensitivitiesJson value.
func (m ProfileSwitch) GetInsulinSensitivitiesJson() *string {
	return m.InsulinSensitivitiesJson
}

// GetBasalRatesJson returns the BasalRatesJson value.
func (m ProfileSwitch) GetBasalRatesJson() *string {
	return m.BasalRatesJson
}

// GetTargetsHighJson returns the TargetsHighJson value.
func (m ProfileSwitch) GetTargetsHighJson() *string {
	return m.TargetsHighJson
}

// GetTargetsLowJson returns the TargetsLowJson value.
func (m ProfileSwitch) GetTargetsLowJson() *string {
	return m.TargetsLowJson
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

// GetIcfgJson returns the IcfgJson value.
func (m ProfileSwitchesVw) GetIcfgJson() *string {
	return m.IcfgJson
}

// GetDia returns the Dia value.
func (m ProfileSwitchesVw) GetDia() *float64 {
	return m.Dia
}

// GetUnits returns the Units value.
func (m ProfileSwitchesVw) GetUnits() *consts.Units {
	return m.Units
}

// GetTimezone returns the Timezone value.
func (m ProfileSwitchesVw) GetTimezone() *string {
	return m.Timezone
}

// GetCarbRatiosJson returns the CarbRatiosJson value.
func (m ProfileSwitchesVw) GetCarbRatiosJson() *string {
	return m.CarbRatiosJson
}

// GetInsulinSensitivitiesJson returns the InsulinSensitivitiesJson value.
func (m ProfileSwitchesVw) GetInsulinSensitivitiesJson() *string {
	return m.InsulinSensitivitiesJson
}

// GetBasalRatesJson returns the BasalRatesJson value.
func (m ProfileSwitchesVw) GetBasalRatesJson() *string {
	return m.BasalRatesJson
}

// GetTargetsHighJson returns the TargetsHighJson value.
func (m ProfileSwitchesVw) GetTargetsHighJson() *string {
	return m.TargetsHighJson
}

// GetTargetsLowJson returns the TargetsLowJson value.
func (m ProfileSwitchesVw) GetTargetsLowJson() *string {
	return m.TargetsLowJson
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

// GetCreatedByIdentityName returns the CreatedByIdentityName value.
func (m ProfileSwitchesVw) GetCreatedByIdentityName() *string {
	return m.CreatedByIdentityName
}

// SetCreatedByIdentityName sets the CreatedByIdentityName value and returns the updated ProfileSwitchesVw.
func (m *ProfileSwitchesVw) SetCreatedByIdentityName(v *string) {
	m.CreatedByIdentityName = v
}

// GetID returns the ID value.
func (m TempBasal) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated TempBasal.
func (m *TempBasal) SetID(v string) {
	m.ID = v
}

// GetType returns the Type value.
func (m TempBasal) GetType() *string {
	return m.Type
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

// GetType returns the Type value.
func (m TempBasalsVw) GetType() *string {
	return m.Type
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

// GetHighTargetMgdl returns the HighTargetMgdl value.
func (m TempTarget) GetHighTargetMgdl() float64 {
	return m.HighTargetMgdl
}

// GetLowTargetMgdl returns the LowTargetMgdl value.
func (m TempTarget) GetLowTargetMgdl() float64 {
	return m.LowTargetMgdl
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

// GetHighTargetMgdl returns the HighTargetMgdl value.
func (m TempTargetsVw) GetHighTargetMgdl() *float64 {
	return m.HighTargetMgdl
}

// GetLowTargetMgdl returns the LowTargetMgdl value.
func (m TempTargetsVw) GetLowTargetMgdl() *float64 {
	return m.LowTargetMgdl
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

// GetCreatedByIdentityName returns the CreatedByIdentityName value.
func (m TempTargetsVw) GetCreatedByIdentityName() *string {
	return m.CreatedByIdentityName
}

// SetCreatedByIdentityName sets the CreatedByIdentityName value and returns the updated TempTargetsVw.
func (m *TempTargetsVw) SetCreatedByIdentityName(v *string) {
	m.CreatedByIdentityName = v
}

// SetLimit sets the Limit value and returns the updated ColdDaysNewestFirstParams.
func (m *ColdDaysNewestFirstParams) SetLimit(v int64) {
	m.Limit = v
}

// SetLimit sets the Limit value and returns the updated ColdDaysPageParams.
func (m *ColdDaysPageParams) SetLimit(v int64) {
	m.Limit = v
}

// SetLimit sets the Limit value and returns the updated ColdDaysPageOldestParams.
func (m *ColdDaysPageOldestParams) SetLimit(v int64) {
	m.Limit = v
}

// GetDate returns the Date value.
func (m GetColdDayParams) GetDate() string {
	return m.Date
}

// GetDate returns the Date value.
func (m MarkColdDaysStaleParams) GetDate() string {
	return m.Date
}

// GetDate returns the Date value.
func (m UpsertColdDayParams) GetDate() string {
	return m.Date
}

// SetLimit sets the Limit value and returns the updated AllDeviceReportsParams.
func (m *AllDeviceReportsParams) SetLimit(v int64) {
	m.Limit = v
}

// GetID returns the ID value.
func (m ClearDeviceReportStatusesParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated ClearDeviceReportStatusesParams.
func (m *ClearDeviceReportStatusesParams) SetID(v string) {
	m.ID = v
}

// GetID returns the ID value.
func (m CreateDeviceReportParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated CreateDeviceReportParams.
func (m *CreateDeviceReportParams) SetID(v string) {
	m.ID = v
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

// SetLimit sets the Limit value and returns the updated DeviceReportAscendingIDsParams.
func (m *DeviceReportAscendingIDsParams) SetLimit(v int64) {
	m.Limit = v
}

// SetLimit sets the Limit value and returns the updated DeviceReportHistoryIDsParams.
func (m *DeviceReportHistoryIDsParams) SetLimit(v int64) {
	m.Limit = v
}

// SetLimit sets the Limit value and returns the updated DeviceReportSortedIDsParams.
func (m *DeviceReportSortedIDsParams) SetLimit(v int64) {
	m.Limit = v
}

// GetDevice returns the Device value.
func (m DeviceReportTotalsParams) GetDevice() *string {
	return m.Device
}

// GetDevice returns the Device value.
func (m DeviceReportsForDeviceParams) GetDevice() *string {
	return m.Device
}

// SetLimit sets the Limit value and returns the updated DeviceReportsForDeviceParams.
func (m *DeviceReportsForDeviceParams) SetLimit(v int64) {
	m.Limit = v
}

// SetLimit sets the Limit value and returns the updated DeviceReportsLoopBetweenParams.
func (m *DeviceReportsLoopBetweenParams) SetLimit(v int64) {
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

// GetID returns the ID value.
func (m HardDeleteDeviceReportParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated HardDeleteDeviceReportParams.
func (m *HardDeleteDeviceReportParams) SetID(v string) {
	m.ID = v
}

// SetLimit sets the Limit value and returns the updated RecentDeviceNamesParams.
func (m *RecentDeviceNamesParams) SetLimit(v int64) {
	m.Limit = v
}

// GetID returns the ID value.
func (m RewriteDeviceReportParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated RewriteDeviceReportParams.
func (m *RewriteDeviceReportParams) SetID(v string) {
	m.ID = v
}

// GetUtcOffsetMinutes returns the UtcOffsetMinutes value.
func (m RewriteDeviceReportParams) GetUtcOffsetMinutes() *int64 {
	return m.UtcOffsetMinutes
}

// SetUtcOffsetMinutes sets the UtcOffsetMinutes value and returns the updated RewriteDeviceReportParams.
func (m *RewriteDeviceReportParams) SetUtcOffsetMinutes(v *int64) {
	m.UtcOffsetMinutes = v
}

// GetOccurredAt returns the OccurredAt value.
func (m RewriteDeviceReportParams) GetOccurredAt() int64 {
	return m.OccurredAt
}

// SetOccurredAt sets the OccurredAt value and returns the updated RewriteDeviceReportParams.
func (m *RewriteDeviceReportParams) SetOccurredAt(v int64) {
	m.OccurredAt = v
}

// GetCreatedByAppName returns the CreatedByAppName value.
func (m RewriteDeviceReportParams) GetCreatedByAppName() *string {
	return m.CreatedByAppName
}

// SetCreatedByAppName sets the CreatedByAppName value and returns the updated RewriteDeviceReportParams.
func (m *RewriteDeviceReportParams) SetCreatedByAppName(v *string) {
	m.CreatedByAppName = v
}

// GetCreatedByDeviceName returns the CreatedByDeviceName value.
func (m RewriteDeviceReportParams) GetCreatedByDeviceName() *string {
	return m.CreatedByDeviceName
}

// SetCreatedByDeviceName sets the CreatedByDeviceName value and returns the updated RewriteDeviceReportParams.
func (m *RewriteDeviceReportParams) SetCreatedByDeviceName(v *string) {
	m.CreatedByDeviceName = v
}

// GetCreatedByIdentityName returns the CreatedByIdentityName value.
func (m RewriteDeviceReportParams) GetCreatedByIdentityName() string {
	return m.CreatedByIdentityName
}

// SetCreatedByIdentityName sets the CreatedByIdentityName value and returns the updated RewriteDeviceReportParams.
func (m *RewriteDeviceReportParams) SetCreatedByIdentityName(v string) {
	m.CreatedByIdentityName = v
}

// SetLimit sets the Limit value and returns the updated SearchDeviceReportsParams.
func (m *SearchDeviceReportsParams) SetLimit(v int64) {
	m.Limit = v
}

// GetID returns the ID value.
func (m SoftDeleteDeviceReportParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated SoftDeleteDeviceReportParams.
func (m *SoftDeleteDeviceReportParams) SetID(v string) {
	m.ID = v
}

// SetLimit sets the Limit value and returns the updated ActivityAscendingIDsParams.
func (m *ActivityAscendingIDsParams) SetLimit(v int64) {
	m.Limit = v
}

// SetLimit sets the Limit value and returns the updated ActivityHistoryIDsParams.
func (m *ActivityHistoryIDsParams) SetLimit(v int64) {
	m.Limit = v
}

// SetLimit sets the Limit value and returns the updated ActivitySortedIDsParams.
func (m *ActivitySortedIDsParams) SetLimit(v int64) {
	m.Limit = v
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

// GetID returns the ID value.
func (m GetActivityParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated GetActivityParams.
func (m *GetActivityParams) SetID(v string) {
	m.ID = v
}

// GetID returns the ID value.
func (m HardDeleteActivityParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated HardDeleteActivityParams.
func (m *HardDeleteActivityParams) SetID(v string) {
	m.ID = v
}

// GetID returns the ID value.
func (m RewriteActivityParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated RewriteActivityParams.
func (m *RewriteActivityParams) SetID(v string) {
	m.ID = v
}

// GetName returns the Name value.
func (m RewriteActivityParams) GetName() *string {
	return m.Name
}

// GetNotes returns the Notes value.
func (m RewriteActivityParams) GetNotes() *string {
	return m.Notes
}

// SetNotes sets the Notes value and returns the updated RewriteActivityParams.
func (m *RewriteActivityParams) SetNotes(v *string) {
	m.Notes = v
}

// GetReason returns the Reason value.
func (m RewriteActivityParams) GetReason() *string {
	return m.Reason
}

// SetReason sets the Reason value and returns the updated RewriteActivityParams.
func (m *RewriteActivityParams) SetReason(v *string) {
	m.Reason = v
}

// GetOccurredAt returns the OccurredAt value.
func (m RewriteActivityParams) GetOccurredAt() int64 {
	return m.OccurredAt
}

// SetOccurredAt sets the OccurredAt value and returns the updated RewriteActivityParams.
func (m *RewriteActivityParams) SetOccurredAt(v int64) {
	m.OccurredAt = v
}

// GetUtcOffsetMinutes returns the UtcOffsetMinutes value.
func (m RewriteActivityParams) GetUtcOffsetMinutes() *int64 {
	return m.UtcOffsetMinutes
}

// SetUtcOffsetMinutes sets the UtcOffsetMinutes value and returns the updated RewriteActivityParams.
func (m *RewriteActivityParams) SetUtcOffsetMinutes(v *int64) {
	m.UtcOffsetMinutes = v
}

// GetCreatedByAppName returns the CreatedByAppName value.
func (m RewriteActivityParams) GetCreatedByAppName() *string {
	return m.CreatedByAppName
}

// SetCreatedByAppName sets the CreatedByAppName value and returns the updated RewriteActivityParams.
func (m *RewriteActivityParams) SetCreatedByAppName(v *string) {
	m.CreatedByAppName = v
}

// GetCreatedByDeviceName returns the CreatedByDeviceName value.
func (m RewriteActivityParams) GetCreatedByDeviceName() *string {
	return m.CreatedByDeviceName
}

// SetCreatedByDeviceName sets the CreatedByDeviceName value and returns the updated RewriteActivityParams.
func (m *RewriteActivityParams) SetCreatedByDeviceName(v *string) {
	m.CreatedByDeviceName = v
}

// GetCreatedByIdentityName returns the CreatedByIdentityName value.
func (m RewriteActivityParams) GetCreatedByIdentityName() string {
	return m.CreatedByIdentityName
}

// SetCreatedByIdentityName sets the CreatedByIdentityName value and returns the updated RewriteActivityParams.
func (m *RewriteActivityParams) SetCreatedByIdentityName(v string) {
	m.CreatedByIdentityName = v
}

// SetLimit sets the Limit value and returns the updated SearchActivitiesParams.
func (m *SearchActivitiesParams) SetLimit(v int64) {
	m.Limit = v
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

// GetID returns the ID value.
func (m UpdateActivityParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated UpdateActivityParams.
func (m *UpdateActivityParams) SetID(v string) {
	m.ID = v
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

// SetLimit sets the Limit value and returns the updated FoodIntakeDurationsParams.
func (m *FoodIntakeDurationsParams) SetLimit(v int64) {
	m.Limit = v
}

// GetID returns the ID value.
func (m GetFoodIntakeParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated GetFoodIntakeParams.
func (m *GetFoodIntakeParams) SetID(v string) {
	m.ID = v
}

// GetID returns the ID value.
func (m HardDeleteFoodIntakeParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated HardDeleteFoodIntakeParams.
func (m *HardDeleteFoodIntakeParams) SetID(v string) {
	m.ID = v
}

// GetID returns the ID value.
func (m RewriteFoodIntakeParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated RewriteFoodIntakeParams.
func (m *RewriteFoodIntakeParams) SetID(v string) {
	m.ID = v
}

// GetName returns the Name value.
func (m RewriteFoodIntakeParams) GetName() *string {
	return m.Name
}

// GetNotes returns the Notes value.
func (m RewriteFoodIntakeParams) GetNotes() *string {
	return m.Notes
}

// SetNotes sets the Notes value and returns the updated RewriteFoodIntakeParams.
func (m *RewriteFoodIntakeParams) SetNotes(v *string) {
	m.Notes = v
}

// GetReason returns the Reason value.
func (m RewriteFoodIntakeParams) GetReason() *string {
	return m.Reason
}

// SetReason sets the Reason value and returns the updated RewriteFoodIntakeParams.
func (m *RewriteFoodIntakeParams) SetReason(v *string) {
	m.Reason = v
}

// GetUtcOffsetMinutes returns the UtcOffsetMinutes value.
func (m RewriteFoodIntakeParams) GetUtcOffsetMinutes() *int64 {
	return m.UtcOffsetMinutes
}

// SetUtcOffsetMinutes sets the UtcOffsetMinutes value and returns the updated RewriteFoodIntakeParams.
func (m *RewriteFoodIntakeParams) SetUtcOffsetMinutes(v *int64) {
	m.UtcOffsetMinutes = v
}

// GetOccurredAt returns the OccurredAt value.
func (m RewriteFoodIntakeParams) GetOccurredAt() int64 {
	return m.OccurredAt
}

// SetOccurredAt sets the OccurredAt value and returns the updated RewriteFoodIntakeParams.
func (m *RewriteFoodIntakeParams) SetOccurredAt(v int64) {
	m.OccurredAt = v
}

// GetCreatedByAppName returns the CreatedByAppName value.
func (m RewriteFoodIntakeParams) GetCreatedByAppName() *string {
	return m.CreatedByAppName
}

// SetCreatedByAppName sets the CreatedByAppName value and returns the updated RewriteFoodIntakeParams.
func (m *RewriteFoodIntakeParams) SetCreatedByAppName(v *string) {
	m.CreatedByAppName = v
}

// GetCreatedByDeviceName returns the CreatedByDeviceName value.
func (m RewriteFoodIntakeParams) GetCreatedByDeviceName() *string {
	return m.CreatedByDeviceName
}

// SetCreatedByDeviceName sets the CreatedByDeviceName value and returns the updated RewriteFoodIntakeParams.
func (m *RewriteFoodIntakeParams) SetCreatedByDeviceName(v *string) {
	m.CreatedByDeviceName = v
}

// GetCreatedByIdentityName returns the CreatedByIdentityName value.
func (m RewriteFoodIntakeParams) GetCreatedByIdentityName() string {
	return m.CreatedByIdentityName
}

// SetCreatedByIdentityName sets the CreatedByIdentityName value and returns the updated RewriteFoodIntakeParams.
func (m *RewriteFoodIntakeParams) SetCreatedByIdentityName(v string) {
	m.CreatedByIdentityName = v
}

// SetLimit sets the Limit value and returns the updated SearchFoodIntakesParams.
func (m *SearchFoodIntakesParams) SetLimit(v int64) {
	m.Limit = v
}

// GetID returns the ID value.
func (m SoftDeleteFoodIntakeParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated SoftDeleteFoodIntakeParams.
func (m *SoftDeleteFoodIntakeParams) SetID(v string) {
	m.ID = v
}

// GetName returns the Name value.
func (m UpdateFoodIntakeParams) GetName() *string {
	return m.Name
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

// GetID returns the ID value.
func (m UpdateFoodIntakeParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated UpdateFoodIntakeParams.
func (m *UpdateFoodIntakeParams) SetID(v string) {
	m.ID = v
}

// SetLimit sets the Limit value and returns the updated AlarmGlucoseReadingsParams.
func (m *AlarmGlucoseReadingsParams) SetLimit(v int64) {
	m.Limit = v
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

// SetReadingType sets the ReadingType value and returns the updated CreateGlucoseReadingParams.
func (m *CreateGlucoseReadingParams) SetReadingType(v string) {
	m.ReadingType = v
}

// GetUnits returns the Units value.
func (m CreateGlucoseReadingParams) GetUnits() *consts.Units {
	return m.Units
}

// GetDelta returns the Delta value.
func (m CreateGlucoseReadingParams) GetDelta() *float64 {
	return m.Delta
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

// GetID returns the ID value.
func (m GetGlucoseReadingParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated GetGlucoseReadingParams.
func (m *GetGlucoseReadingParams) SetID(v string) {
	m.ID = v
}

// SetLimit sets the Limit value and returns the updated GlucoseReadingAscendingIDsParams.
func (m *GlucoseReadingAscendingIDsParams) SetLimit(v int64) {
	m.Limit = v
}

// SetLimit sets the Limit value and returns the updated GlucoseReadingHistoryIDsParams.
func (m *GlucoseReadingHistoryIDsParams) SetLimit(v int64) {
	m.Limit = v
}

// SetLimit sets the Limit value and returns the updated GlucoseReadingSortedIDsParams.
func (m *GlucoseReadingSortedIDsParams) SetLimit(v int64) {
	m.Limit = v
}

// GetID returns the ID value.
func (m HardDeleteGlucoseReadingParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated HardDeleteGlucoseReadingParams.
func (m *HardDeleteGlucoseReadingParams) SetID(v string) {
	m.ID = v
}

// SetReadingType sets the ReadingType value and returns the updated HardDeleteGlucoseReadingsByTypeParams.
func (m *HardDeleteGlucoseReadingsByTypeParams) SetReadingType(v string) {
	m.ReadingType = v
}

// SetLimit sets the Limit value and returns the updated HardDeleteGlucoseReadingsByTypeParams.
func (m *HardDeleteGlucoseReadingsByTypeParams) SetLimit(v int64) {
	m.Limit = v
}

// GetID returns the ID value.
func (m RewriteGlucoseReadingParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated RewriteGlucoseReadingParams.
func (m *RewriteGlucoseReadingParams) SetID(v string) {
	m.ID = v
}

// SetReadingType sets the ReadingType value and returns the updated RewriteGlucoseReadingParams.
func (m *RewriteGlucoseReadingParams) SetReadingType(v string) {
	m.ReadingType = v
}

// GetUnits returns the Units value.
func (m RewriteGlucoseReadingParams) GetUnits() *consts.Units {
	return m.Units
}

// GetDelta returns the Delta value.
func (m RewriteGlucoseReadingParams) GetDelta() *float64 {
	return m.Delta
}

// GetNotes returns the Notes value.
func (m RewriteGlucoseReadingParams) GetNotes() *string {
	return m.Notes
}

// SetNotes sets the Notes value and returns the updated RewriteGlucoseReadingParams.
func (m *RewriteGlucoseReadingParams) SetNotes(v *string) {
	m.Notes = v
}

// GetReason returns the Reason value.
func (m RewriteGlucoseReadingParams) GetReason() *string {
	return m.Reason
}

// SetReason sets the Reason value and returns the updated RewriteGlucoseReadingParams.
func (m *RewriteGlucoseReadingParams) SetReason(v *string) {
	m.Reason = v
}

// GetUtcOffsetMinutes returns the UtcOffsetMinutes value.
func (m RewriteGlucoseReadingParams) GetUtcOffsetMinutes() *int64 {
	return m.UtcOffsetMinutes
}

// SetUtcOffsetMinutes sets the UtcOffsetMinutes value and returns the updated RewriteGlucoseReadingParams.
func (m *RewriteGlucoseReadingParams) SetUtcOffsetMinutes(v *int64) {
	m.UtcOffsetMinutes = v
}

// GetOccurredAt returns the OccurredAt value.
func (m RewriteGlucoseReadingParams) GetOccurredAt() int64 {
	return m.OccurredAt
}

// SetOccurredAt sets the OccurredAt value and returns the updated RewriteGlucoseReadingParams.
func (m *RewriteGlucoseReadingParams) SetOccurredAt(v int64) {
	m.OccurredAt = v
}

// GetCreatedByAppName returns the CreatedByAppName value.
func (m RewriteGlucoseReadingParams) GetCreatedByAppName() *string {
	return m.CreatedByAppName
}

// SetCreatedByAppName sets the CreatedByAppName value and returns the updated RewriteGlucoseReadingParams.
func (m *RewriteGlucoseReadingParams) SetCreatedByAppName(v *string) {
	m.CreatedByAppName = v
}

// GetCreatedByDeviceName returns the CreatedByDeviceName value.
func (m RewriteGlucoseReadingParams) GetCreatedByDeviceName() *string {
	return m.CreatedByDeviceName
}

// SetCreatedByDeviceName sets the CreatedByDeviceName value and returns the updated RewriteGlucoseReadingParams.
func (m *RewriteGlucoseReadingParams) SetCreatedByDeviceName(v *string) {
	m.CreatedByDeviceName = v
}

// GetCreatedByIdentityName returns the CreatedByIdentityName value.
func (m RewriteGlucoseReadingParams) GetCreatedByIdentityName() string {
	return m.CreatedByIdentityName
}

// SetCreatedByIdentityName sets the CreatedByIdentityName value and returns the updated RewriteGlucoseReadingParams.
func (m *RewriteGlucoseReadingParams) SetCreatedByIdentityName(v string) {
	m.CreatedByIdentityName = v
}

// SetLimit sets the Limit value and returns the updated SearchGlucoseReadingsParams.
func (m *SearchGlucoseReadingsParams) SetLimit(v int64) {
	m.Limit = v
}

// GetID returns the ID value.
func (m SoftDeleteGlucoseReadingParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated SoftDeleteGlucoseReadingParams.
func (m *SoftDeleteGlucoseReadingParams) SetID(v string) {
	m.ID = v
}

// SetReadingType sets the ReadingType value and returns the updated SoftDeleteGlucoseReadingsByTypeParams.
func (m *SoftDeleteGlucoseReadingsByTypeParams) SetReadingType(v string) {
	m.ReadingType = v
}

// SetLimit sets the Limit value and returns the updated SoftDeleteGlucoseReadingsByTypeParams.
func (m *SoftDeleteGlucoseReadingsByTypeParams) SetLimit(v int64) {
	m.Limit = v
}

// GetDelta returns the Delta value.
func (m UpdateGlucoseReadingParams) GetDelta() *float64 {
	return m.Delta
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

// GetID returns the ID value.
func (m UpdateGlucoseReadingParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated UpdateGlucoseReadingParams.
func (m *UpdateGlucoseReadingParams) SetID(v string) {
	m.ID = v
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

// SetLimit sets the Limit value and returns the updated EventGroupAscendingIDsParams.
func (m *EventGroupAscendingIDsParams) SetLimit(v int64) {
	m.Limit = v
}

// SetLimit sets the Limit value and returns the updated EventGroupHistoryIDsParams.
func (m *EventGroupHistoryIDsParams) SetLimit(v int64) {
	m.Limit = v
}

// SetLimit sets the Limit value and returns the updated EventGroupSortedIDsParams.
func (m *EventGroupSortedIDsParams) SetLimit(v int64) {
	m.Limit = v
}

// GetID returns the ID value.
func (m GetEventGroupParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated GetEventGroupParams.
func (m *GetEventGroupParams) SetID(v string) {
	m.ID = v
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

// GetID returns the ID value.
func (m HardDeleteEventGroupParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated HardDeleteEventGroupParams.
func (m *HardDeleteEventGroupParams) SetID(v string) {
	m.ID = v
}

// GetOccurredAt returns the OccurredAt value.
func (m RegroupEventGroupParams) GetOccurredAt() *int64 {
	return m.OccurredAt
}

// SetOccurredAt sets the OccurredAt value and returns the updated RegroupEventGroupParams.
func (m *RegroupEventGroupParams) SetOccurredAt(v *int64) {
	m.OccurredAt = v
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

// GetID returns the ID value.
func (m RegroupEventGroupParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated RegroupEventGroupParams.
func (m *RegroupEventGroupParams) SetID(v string) {
	m.ID = v
}

// GetID returns the ID value.
func (m RewriteEventGroupParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated RewriteEventGroupParams.
func (m *RewriteEventGroupParams) SetID(v string) {
	m.ID = v
}

// GetNotes returns the Notes value.
func (m RewriteEventGroupParams) GetNotes() *string {
	return m.Notes
}

// SetNotes sets the Notes value and returns the updated RewriteEventGroupParams.
func (m *RewriteEventGroupParams) SetNotes(v *string) {
	m.Notes = v
}

// GetReason returns the Reason value.
func (m RewriteEventGroupParams) GetReason() *string {
	return m.Reason
}

// SetReason sets the Reason value and returns the updated RewriteEventGroupParams.
func (m *RewriteEventGroupParams) SetReason(v *string) {
	m.Reason = v
}

// GetOccurredAt returns the OccurredAt value.
func (m RewriteEventGroupParams) GetOccurredAt() int64 {
	return m.OccurredAt
}

// SetOccurredAt sets the OccurredAt value and returns the updated RewriteEventGroupParams.
func (m *RewriteEventGroupParams) SetOccurredAt(v int64) {
	m.OccurredAt = v
}

// GetUtcOffsetMinutes returns the UtcOffsetMinutes value.
func (m RewriteEventGroupParams) GetUtcOffsetMinutes() *int64 {
	return m.UtcOffsetMinutes
}

// SetUtcOffsetMinutes sets the UtcOffsetMinutes value and returns the updated RewriteEventGroupParams.
func (m *RewriteEventGroupParams) SetUtcOffsetMinutes(v *int64) {
	m.UtcOffsetMinutes = v
}

// GetCreatedByAppName returns the CreatedByAppName value.
func (m RewriteEventGroupParams) GetCreatedByAppName() *string {
	return m.CreatedByAppName
}

// SetCreatedByAppName sets the CreatedByAppName value and returns the updated RewriteEventGroupParams.
func (m *RewriteEventGroupParams) SetCreatedByAppName(v *string) {
	m.CreatedByAppName = v
}

// GetCreatedByDeviceName returns the CreatedByDeviceName value.
func (m RewriteEventGroupParams) GetCreatedByDeviceName() *string {
	return m.CreatedByDeviceName
}

// SetCreatedByDeviceName sets the CreatedByDeviceName value and returns the updated RewriteEventGroupParams.
func (m *RewriteEventGroupParams) SetCreatedByDeviceName(v *string) {
	m.CreatedByDeviceName = v
}

// GetCreatedByIdentityName returns the CreatedByIdentityName value.
func (m RewriteEventGroupParams) GetCreatedByIdentityName() string {
	return m.CreatedByIdentityName
}

// SetCreatedByIdentityName sets the CreatedByIdentityName value and returns the updated RewriteEventGroupParams.
func (m *RewriteEventGroupParams) SetCreatedByIdentityName(v string) {
	m.CreatedByIdentityName = v
}

// SetLimit sets the Limit value and returns the updated SearchEventGroupsParams.
func (m *SearchEventGroupsParams) SetLimit(v int64) {
	m.Limit = v
}

// GetID returns the ID value.
func (m SoftDeleteEventGroupParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated SoftDeleteEventGroupParams.
func (m *SoftDeleteEventGroupParams) SetID(v string) {
	m.ID = v
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

// GetIcfgJson returns the IcfgJson value.
func (m CreateInsulinDoseParams) GetIcfgJson() *string {
	return m.IcfgJson
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

// GetID returns the ID value.
func (m GetInsulinDoseParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated GetInsulinDoseParams.
func (m *GetInsulinDoseParams) SetID(v string) {
	m.ID = v
}

// GetID returns the ID value.
func (m HardDeleteInsulinDoseParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated HardDeleteInsulinDoseParams.
func (m *HardDeleteInsulinDoseParams) SetID(v string) {
	m.ID = v
}

// SetLimit sets the Limit value and returns the updated InsulinDosesExtendedIntoParams.
func (m *InsulinDosesExtendedIntoParams) SetLimit(v int64) {
	m.Limit = v
}

// GetID returns the ID value.
func (m RewriteInsulinDoseParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated RewriteInsulinDoseParams.
func (m *RewriteInsulinDoseParams) SetID(v string) {
	m.ID = v
}

// GetIcfgJson returns the IcfgJson value.
func (m RewriteInsulinDoseParams) GetIcfgJson() *string {
	return m.IcfgJson
}

// GetNotes returns the Notes value.
func (m RewriteInsulinDoseParams) GetNotes() *string {
	return m.Notes
}

// SetNotes sets the Notes value and returns the updated RewriteInsulinDoseParams.
func (m *RewriteInsulinDoseParams) SetNotes(v *string) {
	m.Notes = v
}

// GetReason returns the Reason value.
func (m RewriteInsulinDoseParams) GetReason() *string {
	return m.Reason
}

// SetReason sets the Reason value and returns the updated RewriteInsulinDoseParams.
func (m *RewriteInsulinDoseParams) SetReason(v *string) {
	m.Reason = v
}

// GetUtcOffsetMinutes returns the UtcOffsetMinutes value.
func (m RewriteInsulinDoseParams) GetUtcOffsetMinutes() *int64 {
	return m.UtcOffsetMinutes
}

// SetUtcOffsetMinutes sets the UtcOffsetMinutes value and returns the updated RewriteInsulinDoseParams.
func (m *RewriteInsulinDoseParams) SetUtcOffsetMinutes(v *int64) {
	m.UtcOffsetMinutes = v
}

// GetOccurredAt returns the OccurredAt value.
func (m RewriteInsulinDoseParams) GetOccurredAt() int64 {
	return m.OccurredAt
}

// SetOccurredAt sets the OccurredAt value and returns the updated RewriteInsulinDoseParams.
func (m *RewriteInsulinDoseParams) SetOccurredAt(v int64) {
	m.OccurredAt = v
}

// GetCreatedByAppName returns the CreatedByAppName value.
func (m RewriteInsulinDoseParams) GetCreatedByAppName() *string {
	return m.CreatedByAppName
}

// SetCreatedByAppName sets the CreatedByAppName value and returns the updated RewriteInsulinDoseParams.
func (m *RewriteInsulinDoseParams) SetCreatedByAppName(v *string) {
	m.CreatedByAppName = v
}

// GetCreatedByDeviceName returns the CreatedByDeviceName value.
func (m RewriteInsulinDoseParams) GetCreatedByDeviceName() *string {
	return m.CreatedByDeviceName
}

// SetCreatedByDeviceName sets the CreatedByDeviceName value and returns the updated RewriteInsulinDoseParams.
func (m *RewriteInsulinDoseParams) SetCreatedByDeviceName(v *string) {
	m.CreatedByDeviceName = v
}

// GetCreatedByIdentityName returns the CreatedByIdentityName value.
func (m RewriteInsulinDoseParams) GetCreatedByIdentityName() string {
	return m.CreatedByIdentityName
}

// SetCreatedByIdentityName sets the CreatedByIdentityName value and returns the updated RewriteInsulinDoseParams.
func (m *RewriteInsulinDoseParams) SetCreatedByIdentityName(v string) {
	m.CreatedByIdentityName = v
}

// SetLimit sets the Limit value and returns the updated SearchInsulinDosesParams.
func (m *SearchInsulinDosesParams) SetLimit(v int64) {
	m.Limit = v
}

// GetID returns the ID value.
func (m SoftDeleteInsulinDoseParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated SoftDeleteInsulinDoseParams.
func (m *SoftDeleteInsulinDoseParams) SetID(v string) {
	m.ID = v
}

// GetIcfgJson returns the IcfgJson value.
func (m UpdateInsulinDoseParams) GetIcfgJson() *string {
	return m.IcfgJson
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

// GetID returns the ID value.
func (m UpdateInsulinDoseParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated UpdateInsulinDoseParams.
func (m *UpdateInsulinDoseParams) SetID(v string) {
	m.ID = v
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

// SetContent sets the Content value and returns the updated CreateNotificationEventParams.
func (m *CreateNotificationEventParams) SetContent(v *string) {
	m.Content = v
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

// GetID returns the ID value.
func (m GetNotificationEventParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated GetNotificationEventParams.
func (m *GetNotificationEventParams) SetID(v string) {
	m.ID = v
}

// GetID returns the ID value.
func (m HardDeleteNotificationEventParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated HardDeleteNotificationEventParams.
func (m *HardDeleteNotificationEventParams) SetID(v string) {
	m.ID = v
}

// GetID returns the ID value.
func (m RewriteNotificationEventParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated RewriteNotificationEventParams.
func (m *RewriteNotificationEventParams) SetID(v string) {
	m.ID = v
}

// SetContent sets the Content value and returns the updated RewriteNotificationEventParams.
func (m *RewriteNotificationEventParams) SetContent(v *string) {
	m.Content = v
}

// GetNotes returns the Notes value.
func (m RewriteNotificationEventParams) GetNotes() *string {
	return m.Notes
}

// SetNotes sets the Notes value and returns the updated RewriteNotificationEventParams.
func (m *RewriteNotificationEventParams) SetNotes(v *string) {
	m.Notes = v
}

// GetReason returns the Reason value.
func (m RewriteNotificationEventParams) GetReason() *string {
	return m.Reason
}

// SetReason sets the Reason value and returns the updated RewriteNotificationEventParams.
func (m *RewriteNotificationEventParams) SetReason(v *string) {
	m.Reason = v
}

// GetUtcOffsetMinutes returns the UtcOffsetMinutes value.
func (m RewriteNotificationEventParams) GetUtcOffsetMinutes() *int64 {
	return m.UtcOffsetMinutes
}

// SetUtcOffsetMinutes sets the UtcOffsetMinutes value and returns the updated RewriteNotificationEventParams.
func (m *RewriteNotificationEventParams) SetUtcOffsetMinutes(v *int64) {
	m.UtcOffsetMinutes = v
}

// GetOccurredAt returns the OccurredAt value.
func (m RewriteNotificationEventParams) GetOccurredAt() int64 {
	return m.OccurredAt
}

// SetOccurredAt sets the OccurredAt value and returns the updated RewriteNotificationEventParams.
func (m *RewriteNotificationEventParams) SetOccurredAt(v int64) {
	m.OccurredAt = v
}

// GetCreatedByAppName returns the CreatedByAppName value.
func (m RewriteNotificationEventParams) GetCreatedByAppName() *string {
	return m.CreatedByAppName
}

// SetCreatedByAppName sets the CreatedByAppName value and returns the updated RewriteNotificationEventParams.
func (m *RewriteNotificationEventParams) SetCreatedByAppName(v *string) {
	m.CreatedByAppName = v
}

// GetCreatedByDeviceName returns the CreatedByDeviceName value.
func (m RewriteNotificationEventParams) GetCreatedByDeviceName() *string {
	return m.CreatedByDeviceName
}

// SetCreatedByDeviceName sets the CreatedByDeviceName value and returns the updated RewriteNotificationEventParams.
func (m *RewriteNotificationEventParams) SetCreatedByDeviceName(v *string) {
	m.CreatedByDeviceName = v
}

// GetCreatedByIdentityName returns the CreatedByIdentityName value.
func (m RewriteNotificationEventParams) GetCreatedByIdentityName() string {
	return m.CreatedByIdentityName
}

// SetCreatedByIdentityName sets the CreatedByIdentityName value and returns the updated RewriteNotificationEventParams.
func (m *RewriteNotificationEventParams) SetCreatedByIdentityName(v string) {
	m.CreatedByIdentityName = v
}

// SetLimit sets the Limit value and returns the updated SearchNotificationEventsParams.
func (m *SearchNotificationEventsParams) SetLimit(v int64) {
	m.Limit = v
}

// GetID returns the ID value.
func (m SoftDeleteNotificationEventParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated SoftDeleteNotificationEventParams.
func (m *SoftDeleteNotificationEventParams) SetID(v string) {
	m.ID = v
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

// GetID returns the ID value.
func (m UpdateNotificationEventParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated UpdateNotificationEventParams.
func (m *UpdateNotificationEventParams) SetID(v string) {
	m.ID = v
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

// GetDia returns the Dia value.
func (m CreateProfileSwitchParams) GetDia() *float64 {
	return m.Dia
}

// GetUnits returns the Units value.
func (m CreateProfileSwitchParams) GetUnits() *consts.Units {
	return m.Units
}

// GetTimezone returns the Timezone value.
func (m CreateProfileSwitchParams) GetTimezone() *string {
	return m.Timezone
}

// GetIcfgJson returns the IcfgJson value.
func (m CreateProfileSwitchParams) GetIcfgJson() *string {
	return m.IcfgJson
}

// GetCarbRatiosJson returns the CarbRatiosJson value.
func (m CreateProfileSwitchParams) GetCarbRatiosJson() *string {
	return m.CarbRatiosJson
}

// GetInsulinSensitivitiesJson returns the InsulinSensitivitiesJson value.
func (m CreateProfileSwitchParams) GetInsulinSensitivitiesJson() *string {
	return m.InsulinSensitivitiesJson
}

// GetBasalRatesJson returns the BasalRatesJson value.
func (m CreateProfileSwitchParams) GetBasalRatesJson() *string {
	return m.BasalRatesJson
}

// GetTargetsHighJson returns the TargetsHighJson value.
func (m CreateProfileSwitchParams) GetTargetsHighJson() *string {
	return m.TargetsHighJson
}

// GetTargetsLowJson returns the TargetsLowJson value.
func (m CreateProfileSwitchParams) GetTargetsLowJson() *string {
	return m.TargetsLowJson
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

// GetID returns the ID value.
func (m GetProfileSwitchParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated GetProfileSwitchParams.
func (m *GetProfileSwitchParams) SetID(v string) {
	m.ID = v
}

// GetID returns the ID value.
func (m HardDeleteProfileSwitchParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated HardDeleteProfileSwitchParams.
func (m *HardDeleteProfileSwitchParams) SetID(v string) {
	m.ID = v
}

// GetID returns the ID value.
func (m RewriteProfileSwitchParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated RewriteProfileSwitchParams.
func (m *RewriteProfileSwitchParams) SetID(v string) {
	m.ID = v
}

// GetName returns the Name value.
func (m RewriteProfileSwitchParams) GetName() string {
	return m.Name
}

// GetDia returns the Dia value.
func (m RewriteProfileSwitchParams) GetDia() *float64 {
	return m.Dia
}

// GetUnits returns the Units value.
func (m RewriteProfileSwitchParams) GetUnits() *consts.Units {
	return m.Units
}

// GetTimezone returns the Timezone value.
func (m RewriteProfileSwitchParams) GetTimezone() *string {
	return m.Timezone
}

// GetIcfgJson returns the IcfgJson value.
func (m RewriteProfileSwitchParams) GetIcfgJson() *string {
	return m.IcfgJson
}

// GetCarbRatiosJson returns the CarbRatiosJson value.
func (m RewriteProfileSwitchParams) GetCarbRatiosJson() *string {
	return m.CarbRatiosJson
}

// GetInsulinSensitivitiesJson returns the InsulinSensitivitiesJson value.
func (m RewriteProfileSwitchParams) GetInsulinSensitivitiesJson() *string {
	return m.InsulinSensitivitiesJson
}

// GetBasalRatesJson returns the BasalRatesJson value.
func (m RewriteProfileSwitchParams) GetBasalRatesJson() *string {
	return m.BasalRatesJson
}

// GetTargetsHighJson returns the TargetsHighJson value.
func (m RewriteProfileSwitchParams) GetTargetsHighJson() *string {
	return m.TargetsHighJson
}

// GetTargetsLowJson returns the TargetsLowJson value.
func (m RewriteProfileSwitchParams) GetTargetsLowJson() *string {
	return m.TargetsLowJson
}

// GetNotes returns the Notes value.
func (m RewriteProfileSwitchParams) GetNotes() *string {
	return m.Notes
}

// SetNotes sets the Notes value and returns the updated RewriteProfileSwitchParams.
func (m *RewriteProfileSwitchParams) SetNotes(v *string) {
	m.Notes = v
}

// GetReason returns the Reason value.
func (m RewriteProfileSwitchParams) GetReason() *string {
	return m.Reason
}

// SetReason sets the Reason value and returns the updated RewriteProfileSwitchParams.
func (m *RewriteProfileSwitchParams) SetReason(v *string) {
	m.Reason = v
}

// GetUtcOffsetMinutes returns the UtcOffsetMinutes value.
func (m RewriteProfileSwitchParams) GetUtcOffsetMinutes() *int64 {
	return m.UtcOffsetMinutes
}

// SetUtcOffsetMinutes sets the UtcOffsetMinutes value and returns the updated RewriteProfileSwitchParams.
func (m *RewriteProfileSwitchParams) SetUtcOffsetMinutes(v *int64) {
	m.UtcOffsetMinutes = v
}

// GetOccurredAt returns the OccurredAt value.
func (m RewriteProfileSwitchParams) GetOccurredAt() int64 {
	return m.OccurredAt
}

// SetOccurredAt sets the OccurredAt value and returns the updated RewriteProfileSwitchParams.
func (m *RewriteProfileSwitchParams) SetOccurredAt(v int64) {
	m.OccurredAt = v
}

// GetCreatedByAppName returns the CreatedByAppName value.
func (m RewriteProfileSwitchParams) GetCreatedByAppName() *string {
	return m.CreatedByAppName
}

// SetCreatedByAppName sets the CreatedByAppName value and returns the updated RewriteProfileSwitchParams.
func (m *RewriteProfileSwitchParams) SetCreatedByAppName(v *string) {
	m.CreatedByAppName = v
}

// GetCreatedByDeviceName returns the CreatedByDeviceName value.
func (m RewriteProfileSwitchParams) GetCreatedByDeviceName() *string {
	return m.CreatedByDeviceName
}

// SetCreatedByDeviceName sets the CreatedByDeviceName value and returns the updated RewriteProfileSwitchParams.
func (m *RewriteProfileSwitchParams) SetCreatedByDeviceName(v *string) {
	m.CreatedByDeviceName = v
}

// GetCreatedByIdentityName returns the CreatedByIdentityName value.
func (m RewriteProfileSwitchParams) GetCreatedByIdentityName() string {
	return m.CreatedByIdentityName
}

// SetCreatedByIdentityName sets the CreatedByIdentityName value and returns the updated RewriteProfileSwitchParams.
func (m *RewriteProfileSwitchParams) SetCreatedByIdentityName(v string) {
	m.CreatedByIdentityName = v
}

// SetLimit sets the Limit value and returns the updated SearchProfileSwitchesParams.
func (m *SearchProfileSwitchesParams) SetLimit(v int64) {
	m.Limit = v
}

// SetLimit sets the Limit value and returns the updated SearchSummaryProfileSwitchesParams.
func (m *SearchSummaryProfileSwitchesParams) SetLimit(v int64) {
	m.Limit = v
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

// GetIcfgJson returns the IcfgJson value.
func (m UpdateProfileSwitchParams) GetIcfgJson() *string {
	return m.IcfgJson
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

// GetID returns the ID value.
func (m UpdateProfileSwitchParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated UpdateProfileSwitchParams.
func (m *UpdateProfileSwitchParams) SetID(v string) {
	m.ID = v
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

// GetType returns the Type value.
func (m CreateTempBasalParams) GetType() *string {
	return m.Type
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

// GetID returns the ID value.
func (m GetTempBasalParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated GetTempBasalParams.
func (m *GetTempBasalParams) SetID(v string) {
	m.ID = v
}

// GetID returns the ID value.
func (m HardDeleteTempBasalParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated HardDeleteTempBasalParams.
func (m *HardDeleteTempBasalParams) SetID(v string) {
	m.ID = v
}

// GetID returns the ID value.
func (m RewriteTempBasalParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated RewriteTempBasalParams.
func (m *RewriteTempBasalParams) SetID(v string) {
	m.ID = v
}

// GetType returns the Type value.
func (m RewriteTempBasalParams) GetType() *string {
	return m.Type
}

// GetNotes returns the Notes value.
func (m RewriteTempBasalParams) GetNotes() *string {
	return m.Notes
}

// SetNotes sets the Notes value and returns the updated RewriteTempBasalParams.
func (m *RewriteTempBasalParams) SetNotes(v *string) {
	m.Notes = v
}

// GetReason returns the Reason value.
func (m RewriteTempBasalParams) GetReason() *string {
	return m.Reason
}

// SetReason sets the Reason value and returns the updated RewriteTempBasalParams.
func (m *RewriteTempBasalParams) SetReason(v *string) {
	m.Reason = v
}

// GetUtcOffsetMinutes returns the UtcOffsetMinutes value.
func (m RewriteTempBasalParams) GetUtcOffsetMinutes() *int64 {
	return m.UtcOffsetMinutes
}

// SetUtcOffsetMinutes sets the UtcOffsetMinutes value and returns the updated RewriteTempBasalParams.
func (m *RewriteTempBasalParams) SetUtcOffsetMinutes(v *int64) {
	m.UtcOffsetMinutes = v
}

// GetOccurredAt returns the OccurredAt value.
func (m RewriteTempBasalParams) GetOccurredAt() int64 {
	return m.OccurredAt
}

// SetOccurredAt sets the OccurredAt value and returns the updated RewriteTempBasalParams.
func (m *RewriteTempBasalParams) SetOccurredAt(v int64) {
	m.OccurredAt = v
}

// GetCreatedByAppName returns the CreatedByAppName value.
func (m RewriteTempBasalParams) GetCreatedByAppName() *string {
	return m.CreatedByAppName
}

// SetCreatedByAppName sets the CreatedByAppName value and returns the updated RewriteTempBasalParams.
func (m *RewriteTempBasalParams) SetCreatedByAppName(v *string) {
	m.CreatedByAppName = v
}

// GetCreatedByDeviceName returns the CreatedByDeviceName value.
func (m RewriteTempBasalParams) GetCreatedByDeviceName() *string {
	return m.CreatedByDeviceName
}

// SetCreatedByDeviceName sets the CreatedByDeviceName value and returns the updated RewriteTempBasalParams.
func (m *RewriteTempBasalParams) SetCreatedByDeviceName(v *string) {
	m.CreatedByDeviceName = v
}

// GetCreatedByIdentityName returns the CreatedByIdentityName value.
func (m RewriteTempBasalParams) GetCreatedByIdentityName() string {
	return m.CreatedByIdentityName
}

// SetCreatedByIdentityName sets the CreatedByIdentityName value and returns the updated RewriteTempBasalParams.
func (m *RewriteTempBasalParams) SetCreatedByIdentityName(v string) {
	m.CreatedByIdentityName = v
}

// SetLimit sets the Limit value and returns the updated SearchTempBasalsParams.
func (m *SearchTempBasalsParams) SetLimit(v int64) {
	m.Limit = v
}

// GetID returns the ID value.
func (m SoftDeleteTempBasalParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated SoftDeleteTempBasalParams.
func (m *SoftDeleteTempBasalParams) SetID(v string) {
	m.ID = v
}

// GetType returns the Type value.
func (m UpdateTempBasalParams) GetType() *string {
	return m.Type
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

// GetID returns the ID value.
func (m UpdateTempBasalParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated UpdateTempBasalParams.
func (m *UpdateTempBasalParams) SetID(v string) {
	m.ID = v
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

// GetHighTargetMgdl returns the HighTargetMgdl value.
func (m CreateTempTargetParams) GetHighTargetMgdl() float64 {
	return m.HighTargetMgdl
}

// GetLowTargetMgdl returns the LowTargetMgdl value.
func (m CreateTempTargetParams) GetLowTargetMgdl() float64 {
	return m.LowTargetMgdl
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

// GetID returns the ID value.
func (m GetTempTargetParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated GetTempTargetParams.
func (m *GetTempTargetParams) SetID(v string) {
	m.ID = v
}

// GetID returns the ID value.
func (m HardDeleteTempTargetParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated HardDeleteTempTargetParams.
func (m *HardDeleteTempTargetParams) SetID(v string) {
	m.ID = v
}

// GetID returns the ID value.
func (m RewriteTempTargetParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated RewriteTempTargetParams.
func (m *RewriteTempTargetParams) SetID(v string) {
	m.ID = v
}

// GetHighTargetMgdl returns the HighTargetMgdl value.
func (m RewriteTempTargetParams) GetHighTargetMgdl() float64 {
	return m.HighTargetMgdl
}

// GetLowTargetMgdl returns the LowTargetMgdl value.
func (m RewriteTempTargetParams) GetLowTargetMgdl() float64 {
	return m.LowTargetMgdl
}

// GetNotes returns the Notes value.
func (m RewriteTempTargetParams) GetNotes() *string {
	return m.Notes
}

// SetNotes sets the Notes value and returns the updated RewriteTempTargetParams.
func (m *RewriteTempTargetParams) SetNotes(v *string) {
	m.Notes = v
}

// GetReason returns the Reason value.
func (m RewriteTempTargetParams) GetReason() *string {
	return m.Reason
}

// SetReason sets the Reason value and returns the updated RewriteTempTargetParams.
func (m *RewriteTempTargetParams) SetReason(v *string) {
	m.Reason = v
}

// GetUtcOffsetMinutes returns the UtcOffsetMinutes value.
func (m RewriteTempTargetParams) GetUtcOffsetMinutes() *int64 {
	return m.UtcOffsetMinutes
}

// SetUtcOffsetMinutes sets the UtcOffsetMinutes value and returns the updated RewriteTempTargetParams.
func (m *RewriteTempTargetParams) SetUtcOffsetMinutes(v *int64) {
	m.UtcOffsetMinutes = v
}

// GetOccurredAt returns the OccurredAt value.
func (m RewriteTempTargetParams) GetOccurredAt() int64 {
	return m.OccurredAt
}

// SetOccurredAt sets the OccurredAt value and returns the updated RewriteTempTargetParams.
func (m *RewriteTempTargetParams) SetOccurredAt(v int64) {
	m.OccurredAt = v
}

// GetCreatedByAppName returns the CreatedByAppName value.
func (m RewriteTempTargetParams) GetCreatedByAppName() *string {
	return m.CreatedByAppName
}

// SetCreatedByAppName sets the CreatedByAppName value and returns the updated RewriteTempTargetParams.
func (m *RewriteTempTargetParams) SetCreatedByAppName(v *string) {
	m.CreatedByAppName = v
}

// GetCreatedByDeviceName returns the CreatedByDeviceName value.
func (m RewriteTempTargetParams) GetCreatedByDeviceName() *string {
	return m.CreatedByDeviceName
}

// SetCreatedByDeviceName sets the CreatedByDeviceName value and returns the updated RewriteTempTargetParams.
func (m *RewriteTempTargetParams) SetCreatedByDeviceName(v *string) {
	m.CreatedByDeviceName = v
}

// GetCreatedByIdentityName returns the CreatedByIdentityName value.
func (m RewriteTempTargetParams) GetCreatedByIdentityName() string {
	return m.CreatedByIdentityName
}

// SetCreatedByIdentityName sets the CreatedByIdentityName value and returns the updated RewriteTempTargetParams.
func (m *RewriteTempTargetParams) SetCreatedByIdentityName(v string) {
	m.CreatedByIdentityName = v
}

// SetLimit sets the Limit value and returns the updated SearchTempTargetsParams.
func (m *SearchTempTargetsParams) SetLimit(v int64) {
	m.Limit = v
}

// GetID returns the ID value.
func (m SoftDeleteTempTargetParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated SoftDeleteTempTargetParams.
func (m *SoftDeleteTempTargetParams) SetID(v string) {
	m.ID = v
}

// GetHighTargetMgdl returns the HighTargetMgdl value.
func (m UpdateTempTargetParams) GetHighTargetMgdl() *float64 {
	return m.HighTargetMgdl
}

// GetLowTargetMgdl returns the LowTargetMgdl value.
func (m UpdateTempTargetParams) GetLowTargetMgdl() *float64 {
	return m.LowTargetMgdl
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

// GetID returns the ID value.
func (m UpdateTempTargetParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated UpdateTempTargetParams.
func (m *UpdateTempTargetParams) SetID(v string) {
	m.ID = v
}

// GetKey returns the Key value.
func (m DeleteGeneralSettingParams) GetKey() string {
	return m.Key
}

// GetType returns the Type value.
func (m DeleteGeneralSettingsBeforeParams) GetType() string {
	return m.Type
}

// GetUpdatedAt returns the UpdatedAt value.
func (m DeleteGeneralSettingsBeforeParams) GetUpdatedAt() int64 {
	return m.UpdatedAt
}

// SetUpdatedAt sets the UpdatedAt value and returns the updated DeleteGeneralSettingsBeforeParams.
func (m *DeleteGeneralSettingsBeforeParams) SetUpdatedAt(v int64) {
	m.UpdatedAt = v
}

// GetKey returns the Key value.
func (m GeneralSettingsBytesParams) GetKey() string {
	return m.Key
}

// GetKey returns the Key value.
func (m GetGeneralSettingParams) GetKey() string {
	return m.Key
}

// SetLimit sets the Limit value and returns the updated SearchGeneralSettingsParams.
func (m *SearchGeneralSettingsParams) SetLimit(v int64) {
	m.Limit = v
}

// GetKey returns the Key value.
func (m UpsertGeneralSettingParams) GetKey() string {
	return m.Key
}

// GetType returns the Type value.
func (m UpsertGeneralSettingParams) GetType() string {
	return m.Type
}

// GetCreatedByIdentityName returns the CreatedByIdentityName value.
func (m UpsertGeneralSettingParams) GetCreatedByIdentityName() string {
	return m.CreatedByIdentityName
}

// SetCreatedByIdentityName sets the CreatedByIdentityName value and returns the updated UpsertGeneralSettingParams.
func (m *UpsertGeneralSettingParams) SetCreatedByIdentityName(v string) {
	m.CreatedByIdentityName = v
}

// GetID returns the ID value.
func (m DeleteLoopSettingsParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated DeleteLoopSettingsParams.
func (m *DeleteLoopSettingsParams) SetID(v string) {
	m.ID = v
}

// GetID returns the ID value.
func (m GetLoopSettingsParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated GetLoopSettingsParams.
func (m *GetLoopSettingsParams) SetID(v string) {
	m.ID = v
}

// GetApp returns the App value.
func (m GetPreviousLoopSettingsParams) GetApp() string {
	return m.App
}

// GetID returns the ID value.
func (m GetPreviousLoopSettingsParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated GetPreviousLoopSettingsParams.
func (m *GetPreviousLoopSettingsParams) SetID(v string) {
	m.ID = v
}

// GetApp returns the App value.
func (m SearchLoopSettingsParams) GetApp() *string {
	return m.App
}

// SetLimit sets the Limit value and returns the updated SearchLoopSettingsParams.
func (m *SearchLoopSettingsParams) SetLimit(v int64) {
	m.Limit = v
}

// GetID returns the ID value.
func (m UpsertLoopSettingsParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated UpsertLoopSettingsParams.
func (m *UpsertLoopSettingsParams) SetID(v string) {
	m.ID = v
}

// GetApp returns the App value.
func (m UpsertLoopSettingsParams) GetApp() string {
	return m.App
}

// GetSource returns the Source value.
func (m UpsertLoopSettingsParams) GetSource() string {
	return m.Source
}

// GetCreatedByDeviceName returns the CreatedByDeviceName value.
func (m UpsertLoopSettingsParams) GetCreatedByDeviceName() *string {
	return m.CreatedByDeviceName
}

// SetCreatedByDeviceName sets the CreatedByDeviceName value and returns the updated UpsertLoopSettingsParams.
func (m *UpsertLoopSettingsParams) SetCreatedByDeviceName(v *string) {
	m.CreatedByDeviceName = v
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

// SetLimit sets the Limit value and returns the updated FoodPresetAscendingIDsParams.
func (m *FoodPresetAscendingIDsParams) SetLimit(v int64) {
	m.Limit = v
}

// SetLimit sets the Limit value and returns the updated FoodPresetHistoryIDsParams.
func (m *FoodPresetHistoryIDsParams) SetLimit(v int64) {
	m.Limit = v
}

// SetLimit sets the Limit value and returns the updated FoodPresetSortedIDsParams.
func (m *FoodPresetSortedIDsParams) SetLimit(v int64) {
	m.Limit = v
}

// GetName returns the Name value.
func (m FoodPresetsNamedParams) GetName() string {
	return m.Name
}

// SetLimit sets the Limit value and returns the updated FoodPresetsNamedParams.
func (m *FoodPresetsNamedParams) SetLimit(v int64) {
	m.Limit = v
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

// GetID returns the ID value.
func (m RewriteFoodPresetParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated RewriteFoodPresetParams.
func (m *RewriteFoodPresetParams) SetID(v string) {
	m.ID = v
}

// GetName returns the Name value.
func (m RewriteFoodPresetParams) GetName() string {
	return m.Name
}

// GetCreatedByAppName returns the CreatedByAppName value.
func (m RewriteFoodPresetParams) GetCreatedByAppName() *string {
	return m.CreatedByAppName
}

// SetCreatedByAppName sets the CreatedByAppName value and returns the updated RewriteFoodPresetParams.
func (m *RewriteFoodPresetParams) SetCreatedByAppName(v *string) {
	m.CreatedByAppName = v
}

// GetCreatedByDeviceName returns the CreatedByDeviceName value.
func (m RewriteFoodPresetParams) GetCreatedByDeviceName() *string {
	return m.CreatedByDeviceName
}

// SetCreatedByDeviceName sets the CreatedByDeviceName value and returns the updated RewriteFoodPresetParams.
func (m *RewriteFoodPresetParams) SetCreatedByDeviceName(v *string) {
	m.CreatedByDeviceName = v
}

// GetCreatedByIdentityName returns the CreatedByIdentityName value.
func (m RewriteFoodPresetParams) GetCreatedByIdentityName() string {
	return m.CreatedByIdentityName
}

// SetCreatedByIdentityName sets the CreatedByIdentityName value and returns the updated RewriteFoodPresetParams.
func (m *RewriteFoodPresetParams) SetCreatedByIdentityName(v string) {
	m.CreatedByIdentityName = v
}

// SetLimit sets the Limit value and returns the updated SearchFoodPresetsParams.
func (m *SearchFoodPresetsParams) SetLimit(v int64) {
	m.Limit = v
}

// GetName returns the Name value.
func (m UpdateFoodPresetParams) GetName() *string {
	return m.Name
}

// GetID returns the ID value.
func (m UpdateFoodPresetParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated UpdateFoodPresetParams.
func (m *UpdateFoodPresetParams) SetID(v string) {
	m.ID = v
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

// GetDia returns the Dia value.
func (m CreateProfilePresetParams) GetDia() *float64 {
	return m.Dia
}

// GetUnits returns the Units value.
func (m CreateProfilePresetParams) GetUnits() *consts.Units {
	return m.Units
}

// GetTimezone returns the Timezone value.
func (m CreateProfilePresetParams) GetTimezone() *string {
	return m.Timezone
}

// GetCarbRatiosJson returns the CarbRatiosJson value.
func (m CreateProfilePresetParams) GetCarbRatiosJson() *string {
	return m.CarbRatiosJson
}

// GetInsulinSensitivitiesJson returns the InsulinSensitivitiesJson value.
func (m CreateProfilePresetParams) GetInsulinSensitivitiesJson() *string {
	return m.InsulinSensitivitiesJson
}

// GetBasalRatesJson returns the BasalRatesJson value.
func (m CreateProfilePresetParams) GetBasalRatesJson() *string {
	return m.BasalRatesJson
}

// GetTargetsHighJson returns the TargetsHighJson value.
func (m CreateProfilePresetParams) GetTargetsHighJson() *string {
	return m.TargetsHighJson
}

// GetTargetsLowJson returns the TargetsLowJson value.
func (m CreateProfilePresetParams) GetTargetsLowJson() *string {
	return m.TargetsLowJson
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

// GetID returns the ID value.
func (m GetProfilePresetParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated GetProfilePresetParams.
func (m *GetProfilePresetParams) SetID(v string) {
	m.ID = v
}

// GetID returns the ID value.
func (m HardDeleteProfilePresetParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated HardDeleteProfilePresetParams.
func (m *HardDeleteProfilePresetParams) SetID(v string) {
	m.ID = v
}

// SetLimit sets the Limit value and returns the updated ProfileStoreAscendingIDsParams.
func (m *ProfileStoreAscendingIDsParams) SetLimit(v int64) {
	m.Limit = v
}

// SetLimit sets the Limit value and returns the updated ProfileStoreHistoryIDsParams.
func (m *ProfileStoreHistoryIDsParams) SetLimit(v int64) {
	m.Limit = v
}

// SetLimit sets the Limit value and returns the updated ProfileStoreSortedIDsParams.
func (m *ProfileStoreSortedIDsParams) SetLimit(v int64) {
	m.Limit = v
}

// SetLimit sets the Limit value and returns the updated SearchProfileStoresParams.
func (m *SearchProfileStoresParams) SetLimit(v int64) {
	m.Limit = v
}

// SetLimit sets the Limit value and returns the updated SearchSummaryDefaultProfilesParams.
func (m *SearchSummaryDefaultProfilesParams) SetLimit(v int64) {
	m.Limit = v
}

// SetLimit sets the Limit value and returns the updated SearchSummaryStoreProfilesParams.
func (m *SearchSummaryStoreProfilesParams) SetLimit(v int64) {
	m.Limit = v
}

// GetName returns the Name value.
func (m UpdateProfilePresetParams) GetName() *string {
	return m.Name
}

// GetID returns the ID value.
func (m UpdateProfilePresetParams) GetID() string {
	return m.ID
}

// SetID sets the ID value and returns the updated UpdateProfilePresetParams.
func (m *UpdateProfilePresetParams) SetID(v string) {
	m.ID = v
}

