package sitedb

import "nighttroop.t1d.nz/utils"

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

func (model EventGroupsBundle) GetRawJsonData() *string {
	return model.EventGroup.RawJsonData
}

type EventGroupsBundle struct {
	EventGroup        EventGroup
	GlucoseReading    *GlucoseReading
	InsulinDose       *InsulinDose
	FoodIntake        *FoodIntake
	TempBasal         *TempBasal
	TempTarget        *TempTarget
	ProfileSwitch     *ProfileSwitch
	ComponentEvent    *ComponentEvent
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
	if model.ComponentEvent != nil {
		subEvents = append(subEvents, model.ComponentEvent)
	}
	if model.Activity != nil {
		subEvents = append(subEvents, model.Activity)
	}
	if model.NotificationEvent != nil {
		subEvents = append(subEvents, model.NotificationEvent)
	}
	return subEvents
}

func GetEventGroupToEventGroupsBundle(source GetEventGroupRow) EventGroupsBundle {
	egb := EventGroupsBundle{
		EventGroup: source.EventGroup,
	}

	if source.GlucoseReadingsVw.ID != nil {
		utils.ConvertViaJSON(source.GlucoseReadingsVw, &egb.GlucoseReading)
	}

	if source.InsulinDosesVw.ID != nil {
		utils.ConvertViaJSON(source.InsulinDosesVw, &egb.InsulinDose)

	}

	if source.FoodIntakesVw.ID != nil {
		utils.ConvertViaJSON(source.FoodIntakesVw, &egb.FoodIntake)
	}

	if source.TempBasalsVw.ID != nil {
		utils.ConvertViaJSON(source.TempBasalsVw, &egb.TempBasal)
	}

	if source.TempTargetsVw.ID != nil {
		utils.ConvertViaJSON(source.TempTargetsVw, &egb.TempTarget)
	}

	if source.ProfileSwitchesVw.ID != nil {
		utils.ConvertViaJSON(source.ProfileSwitchesVw, &egb.ProfileSwitch)

	}

	if source.ComponentEventsVw.ID != nil {
		utils.ConvertViaJSON(source.ComponentEventsVw, &egb.ComponentEvent)
	}

	if source.ActivitiesVw.ID != nil {
		utils.ConvertViaJSON(source.ActivitiesVw, &egb.Activity)
	}

	if source.NotificationEventsVw.ID != nil {
		utils.ConvertViaJSON(source.NotificationEventsVw, &egb.NotificationEvent)
	}

	return egb
}
