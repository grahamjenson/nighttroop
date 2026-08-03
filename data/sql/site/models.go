package site

import (
	"strings"
	"time"
)

func (model *GlucoseReading) CreatedAtDate() time.Time {
	t := time.UnixMilli(model.CreatedAt)
	return t
}

func IsDuplicateID(err error) bool {
	// Exmaple Error String is "UNIQUE constraint failed: glucose_readings.id"
	// TODO: probably slow
	return strings.Contains(err.Error(), "UNIQUE constraint failed") && strings.Contains(err.Error(), ".id")
}

func (model *GetEventGroupRow) GetID() string {
	return model.EventGroup.ID
}

func (model *GetEventGroupRow) GetDeletedAt() *int64 {
	return model.EventGroup.DeletedAt
}

func (model *GetEventGroupRow) GetCreatedAt() int64 {
	return model.EventGroup.CreatedAt
}

func (model *GetEventGroupRow) GetUpdatedAt() int64 {
	return model.EventGroup.UpdatedAt
}

func (model *GetEventGroupRow) GetCreatedByIdentityName() string {
	return model.EventGroup.CreatedByIdentityName
}

func (model *FoodPreset) GetDeletedAt() *int64 {
	return nil
}

/////////
// Event Group Stuff
/////////

type CreateEventGroupBundle struct {
	EventGroup     CreateEventGroupParams
	GlucoseReading *CreateGlucoseReadingParams
	InsulinDose    *CreateInsulinDoseParams
	FoodIntake     *CreateFoodIntakeParams
	TempBasal      *CreateTempBasalParams
	TempTarget     *CreateTempTargetParams
	ProfileSwitch  *CreateProfileSwitchParams
	DeviceState    *CreateDeviceStateParams
	LoopState      *CreateLoopStateParams
	PumpState      *CreatePumpStateParams
	CgmState       *CreateCGMStateParams
	Activity       *CreateActivityParams
	Note           *CreateNoteParams
}

type UpdateEventGroupBundle struct {
	EventGroup     UpdateEventGroupParams
	GlucoseReading *UpdateGlucoseReadingParams
	InsulinDose    *UpdateInsulinDoseParams
	FoodIntake     *UpdateFoodIntakeParams
	TempBasal      *UpdateTempBasalParams
	TempTarget     *UpdateTempTargetParams
	ProfileSwitch  *UpdateProfileSwitchParams
	DeviceState    *UpdateDeviceStateParams
	LoopState      *UpdateLoopStateParams
	PumpState      *UpdatePumpStateParams
	CgmState       *UpdateCGMStateParams
	Activity       *UpdateActivityParams
	Note           *UpdateNoteParams
}
