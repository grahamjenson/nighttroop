package sitedb

/////////
// Event Group Stuff
/////////

func (model EventGroupsBundle) GetID() string {
	return model.EventGroup.ID
}

func (model EventGroupsBundle) GetDeletedAt() *int64 {
	return model.EventGroup.DeletedAt
}

func (model EventGroupsBundle) GetCreatedAt() int64 {
	return model.EventGroup.CreatedAt
}

func (model EventGroupsBundle) GetUpdatedAt() int64 {
	return model.EventGroup.UpdatedAt
}

func (model EventGroupsBundle) GetCreatedByIdentityName() string {
	return model.EventGroup.CreatedByIdentityName
}

func (model EventGroupsBundle) GetCreatedByAppName() *string {
	return model.EventGroup.CreatedByAppName
}

func (model EventGroupsBundle) GetCreatedByDeviceName() *string {
	return model.EventGroup.CreatedByDeviceName
}

func (model EventGroupsBundle) GetOccurredAt() int64 {
	return model.EventGroup.OccurredAt
}

func (model EventGroupsBundle) GetUtcOffsetMinutes() *int64 {
	return model.EventGroup.UtcOffsetMinutes
}

type EventGroupsBundle struct {
	EventGroup        EventGroup
	GlucoseReading    *GlucoseReading
	InsulinDose       *InsulinDose
	FoodIntake        *FoodIntake
	TempBasal         *TempBasal
	TempTarget        *TempTarget
	ProfileSwitch     *ProfileSwitch
	Activity          *Activity
	NotificationEvent *NotificationEvent
}

type SubEventInterface interface {
	GetID() string
	GetDeletedAt() *int64
	GetCreatedAt() int64
	GetUpdatedAt() int64
	GetCreatedByIdentityName() string
	GetCreatedByAppName() *string
	GetCreatedByDeviceName() *string
	GetOccurredAt() int64
	GetUtcOffsetMinutes() *int64
	GetNotes() *string
	GetReason() *string

	SetID(id string)
	SetDeletedAt(deletedAt *int64)
	SetCreatedAt(createdAt int64)
	SetUpdatedAt(updatedAt int64)
	SetCreatedByIdentityName(identityName string)
	SetCreatedByAppName(appName *string)
	SetCreatedByDeviceName(deviceName *string)
	SetOccurredAt(occurredAt int64)
	SetUtcOffsetMinutes(utcOffsetMinutes *int64)
	SetNotes(notes *string)
	SetReason(reason *string)
}

func (model *EventGroupsBundle) GetAllSubEvents() []SubEventInterface {
	subEvents := []SubEventInterface{}
	if model.GlucoseReading != nil {
		subEvents = append(subEvents, model.GlucoseReading)
	}
	if model.InsulinDose != nil {
		subEvents = append(subEvents, model.InsulinDose)
	}
	if model.FoodIntake != nil {
		subEvents = append(subEvents, model.FoodIntake)
	}
	if model.TempBasal != nil {
		subEvents = append(subEvents, model.TempBasal)
	}
	if model.TempTarget != nil {
		subEvents = append(subEvents, model.TempTarget)
	}
	if model.ProfileSwitch != nil {
		subEvents = append(subEvents, model.ProfileSwitch)
	}
	if model.Activity != nil {
		subEvents = append(subEvents, model.Activity)
	}
	if model.NotificationEvent != nil {
		subEvents = append(subEvents, model.NotificationEvent)
	}
	return subEvents
}
