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

func (model FoodPreset) GetDeletedAt() *int64 {
	return nil
}

func (model FoodPreset) GetUtcOffsetMinutes() *int64 {
	utcOffset := int64(0)
	return &utcOffset
}

func (model FoodPreset) GetOccurredAt() int64 {
	return model.CreatedAt
}

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

func (model EventGroupsBundle) GetOccurredAt() int64 {
	return model.EventGroup.OccurredAt
}

func (model EventGroupsBundle) GetUtcOffsetMinutes() *int64 {
	return model.EventGroup.UtcOffsetMinutes
}

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

type EventGroupsBundle struct {
	EventGroup     EventGroup
	GlucoseReading *GlucoseReading
	InsulinDose    *InsulinDose
	FoodIntake     *FoodIntake
	TempBasal      *TempBasal
	TempTarget     *TempTarget
	ProfileSwitch  *ProfileSwitch
	DeviceState    *DeviceState
	LoopState      *LoopState
	PumpState      *PumpState
	CgmState       *CgmState
	Activity       *Activity
	Note           *Note
}

func GetEventGroupToEventGroupsBundle(source GetEventGroupRow) EventGroupsBundle {
	egb := EventGroupsBundle{
		EventGroup: source.EventGroup,
	}

	if source.GlucoseReadingID != nil && source.GlucoseReading != nil && source.GlucoseReadingType != nil {
		egb.GlucoseReading = &GlucoseReading{
			ID:          *source.GlucoseReadingID,
			Reading:     *source.GlucoseReading,
			ReadingType: *source.GlucoseReadingType,
		}
	}

	if source.InsulinDoseID != nil && source.InsulinDose != nil {
		egb.InsulinDose = &InsulinDose{
			ID:       *source.InsulinDoseID,
			Dose:     *source.InsulinDose,
			DoseType: source.InsulinDoseType,
		}
	}

	if source.FoodIntakeID != nil {
		egb.FoodIntake = &FoodIntake{
			ID:       *source.FoodIntakeID,
			Name:     source.FoodIntakeName,
			Carbs:    source.FoodIntakeCarbs,
			Fat:      source.FoodIntakeFat,
			Protein:  source.FoodIntakeProtein,
			Fibre:    source.FoodIntakeFibre,
			Calories: source.FoodIntakeCalories,
		}
	}

	if source.TempBasalID != nil && source.TempBasalDurationMinutes != nil && source.TempBasalAbsolute != nil {
		egb.TempBasal = &TempBasal{
			ID:              *source.TempBasalID,
			Absolute:        *source.TempBasalAbsolute,
			DurationMinutes: *source.TempBasalDurationMinutes,
			Percent:         source.TempBasalPercent,
		}
	}

	if source.TempTargetID != nil && source.TempTargetDurationMinutes != nil && source.TempTargetHighTarget != nil && source.TempTargetLowTarget != nil {
		egb.TempTarget = &TempTarget{
			ID:              *source.TempTargetID,
			DurationMinutes: *source.TempTargetDurationMinutes,
			HighTarget:      *source.TempTargetHighTarget,
			LowTarget:       *source.TempTargetLowTarget,
		}
	}

	if source.ProfileSwitchID != nil && source.ProfileSwitchName != nil && source.ProfileSwitchDurationMinutes != nil && source.ProfileSwitchPercent != nil {
		egb.ProfileSwitch = &ProfileSwitch{
			ID:                       *source.ProfileSwitchID,
			Name:                     *source.ProfileSwitchName,
			DurationMinutes:          *source.ProfileSwitchDurationMinutes,
			Percent:                  *source.ProfileSwitchPercent,
			CarbRatiosJson:           source.ProfileSwitchCarbRatiosJson,
			InsulinSensitivitiesJson: source.ProfileSwitchInsulinSensitivitiesJson,
			BasalRatesJson:           source.ProfileSwitchBasalRatesJson,
			TargetsJson:              source.ProfileSwitchTargetsJson,
		}

	}

	if source.DeviceStateID != nil {
		egb.DeviceState = &DeviceState{
			ID:                   *source.DeviceStateID,
			DeviceState:          source.DeviceStateDeviceState,
			DeviceTimestamp:      source.DeviceStateDeviceTimestamp,
			DeviceBatteryPercent: source.DeviceStateDeviceBatteryPercent,
			DeviceChargingState:  source.DeviceStateDeviceChargingState,
		}
	}

	if source.LoopStateID != nil {
		egb.LoopState = &LoopState{
			ID:            *source.LoopStateID,
			LoopState:     source.LoopStateLoopState,
			LoopTimestamp: source.LoopStateLoopTimestamp,
			Iob:           source.LoopStateIob,
			IobBasal:      source.LoopStateIobBasal,
			IobTreatment:  source.LoopStateIobTreatment,
			Cob:           source.LoopStateCob,
		}
	}

	if source.PumpStateID != nil {
		egb.PumpState = &PumpState{
			ID:                 *source.PumpStateID,
			PumpState:          source.PumpStatePumpState,
			PumpTimestamp:      source.PumpStatePumpTimestamp,
			PumpInsertionTime:  source.PumpStatePumpInsertionTime,
			PumpReservoir:      source.PumpStatePumpReservoir,
			PumpBatteryPercent: source.PumpStatePumpBatteryPercent,
			PumpChargingState:  source.PumpStatePumpChargingState,
		}
	}

	if source.CgmStateID != nil {
		egb.CgmState = &CgmState{
			ID:               *source.CgmStateID,
			CgmState:         source.CgmStateCgmState,
			CgmTimestamp:     source.CgmStateCgmTimestamp,
			CgmInsertionTime: source.CgmStateCgmInsertionTime,
		}
	}

	if source.ActivityID != nil {
		egb.Activity = &Activity{
			ID:              *source.ActivityID,
			Name:            source.ActivityName,
			ActivityType:    source.ActivityType,
			DurationMinutes: source.ActivityDurationMinutes,
			Calories:        source.ActivityCalories,
			Steps:           source.ActivitySteps,
			HeartRate:       source.ActivityHeartRate,
		}
	}

	if source.NoteID != nil && source.NoteContent != nil {
		egb.Note = &Note{
			ID:      *source.NoteID,
			Content: *source.NoteContent,
			Reason:  source.NoteReason,
		}
	}

	return egb
}
