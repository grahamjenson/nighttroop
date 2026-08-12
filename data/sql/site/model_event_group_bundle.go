package site

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
	ComponentEvent    *ComponentEvent
	Activity          *Activity
	NotificationEvent *NotificationEvent
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

	if source.InsulinDoseID != nil && source.InsulinDose != nil && source.InsulinDoseType != nil {
		egb.InsulinDose = &InsulinDose{
			ID:       *source.InsulinDoseID,
			Dose:     *source.InsulinDose,
			DoseType: *source.InsulinDoseType,
		}
	}

	if source.FoodIntakeID != nil {
		egb.FoodIntake = &FoodIntake{
			ID:          *source.FoodIntakeID,
			Name:        source.FoodIntakeName,
			Category:    source.FoodIntakeCategory,
			Subcategory: source.FoodIntakeSubcategory,
			Carbs:       source.FoodIntakeCarbs,
			Fat:         source.FoodIntakeFat,
			Protein:     source.FoodIntakeProtein,
			Fibre:       source.FoodIntakeFibre,
			Energy:      source.FoodIntakeEnergy,
			Portions:    source.FoodIntakePortions,
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
			TargetsHighJson:          source.ProfileSwitchTargetsHighJson,
			TargetsLowJson:           source.ProfileSwitchTargetsLowJson,
		}

	}

	if source.ComponentEventID != nil {
		egb.ComponentEvent = &ComponentEvent{
			ID:        *source.ComponentEventID,
			EventType: source.EventGroup.GroupType,
		}
	}

	if source.ActivityID != nil && source.ActivityType != nil {
		egb.Activity = &Activity{
			ID:              *source.ActivityID,
			Name:            source.ActivityName,
			ActivityType:    *source.ActivityType,
			DurationMinutes: source.ActivityDurationMinutes,
			Calories:        source.ActivityCalories,
			Steps:           source.ActivitySteps,
			HeartRate:       source.ActivityHeartRate,
		}
	}

	if source.NotificationEventID != nil {
		egb.NotificationEvent = &NotificationEvent{
			ID:               *source.NotificationEventID,
			Title:            *source.NotificationEventTitle,
			Severity:         *source.NotificationEventSeverity,
			Content:          source.NotificationEventContent,
			NotificationType: *source.NotificationEventNotificationType,
		}
	}

	return egb
}

// Update if EventGroup has a Reference and
func (model EventGroupsBundle) UpdateGlucoseReading() bool {
	return model.EventGroup.GlucoseReadingID != nil && model.GlucoseReading != nil
}
func (model EventGroupsBundle) UpdateInsulinDose() bool {
	return model.EventGroup.InsulinDoseID != nil && model.InsulinDose != nil
}
func (model EventGroupsBundle) UpdateFoodIntake() bool {
	return model.EventGroup.FoodIntakeID != nil && model.FoodIntake != nil
}
func (model EventGroupsBundle) UpdateTempBasal() bool {
	return model.EventGroup.TempBasalID != nil && model.TempBasal != nil
}
func (model EventGroupsBundle) UpdateTempTarget() bool {
	return model.EventGroup.TempTargetID != nil && model.TempTarget != nil
}
func (model EventGroupsBundle) UpdateProfileSwitch() bool {
	return model.EventGroup.ProfileSwitchID != nil && model.ProfileSwitch != nil
}
func (model EventGroupsBundle) UpdateComponentEvent() bool {
	return model.EventGroup.ComponentEventID != nil && model.ComponentEvent != nil
}
func (model EventGroupsBundle) UpdateActivity() bool {
	return model.EventGroup.ActivityID != nil && model.Activity != nil
}
func (model EventGroupsBundle) UpdateNotificationEvent() bool {
	return model.EventGroup.NotificationEventID != nil && model.NotificationEvent != nil
}

// Create if EventGroup has model but no reference
func (model EventGroupsBundle) CreateGlucoseReading() bool {
	return model.EventGroup.GlucoseReadingID == nil && model.GlucoseReading != nil
}
func (model EventGroupsBundle) CreateInsulinDose() bool {
	return model.EventGroup.InsulinDoseID == nil && model.InsulinDose != nil
}
func (model EventGroupsBundle) CreateFoodIntake() bool {
	return model.EventGroup.FoodIntakeID == nil && model.FoodIntake != nil
}
func (model EventGroupsBundle) CreateTempBasal() bool {
	return model.EventGroup.TempBasalID == nil && model.TempBasal != nil
}
func (model EventGroupsBundle) CreateTempTarget() bool {
	return model.EventGroup.TempTargetID == nil && model.TempTarget != nil
}
func (model EventGroupsBundle) CreateProfileSwitch() bool {
	return model.EventGroup.ProfileSwitchID == nil && model.ProfileSwitch != nil
}
func (model EventGroupsBundle) CreateComponentEvent() bool {
	return model.EventGroup.ComponentEventID == nil && model.ComponentEvent != nil
}
func (model EventGroupsBundle) CreateActivity() bool {
	return model.EventGroup.ActivityID == nil && model.Activity != nil
}
func (model EventGroupsBundle) CreateNotificationEvent() bool {
	return model.EventGroup.NotificationEventID == nil && model.NotificationEvent != nil
}

// Delete if EventGroup has reference but no model
func (model EventGroupsBundle) DeleteGlucoseReading() bool {
	return model.EventGroup.GlucoseReadingID != nil && model.GlucoseReading == nil
}
func (model EventGroupsBundle) DeleteInsulinDose() bool {
	return model.EventGroup.InsulinDoseID != nil && model.InsulinDose == nil
}
func (model EventGroupsBundle) DeleteFoodIntake() bool {
	return model.EventGroup.FoodIntakeID != nil && model.FoodIntake == nil
}
func (model EventGroupsBundle) DeleteTempBasal() bool {
	return model.EventGroup.TempBasalID != nil && model.TempBasal == nil
}
func (model EventGroupsBundle) DeleteTempTarget() bool {
	return model.EventGroup.TempTargetID != nil && model.TempTarget == nil
}
func (model EventGroupsBundle) DeleteProfileSwitch() bool {
	return model.EventGroup.ProfileSwitchID != nil && model.ProfileSwitch == nil
}
func (model EventGroupsBundle) DeleteComponentEvent() bool {
	return model.EventGroup.ComponentEventID != nil && model.ComponentEvent == nil
}
func (model EventGroupsBundle) DeleteActivity() bool {
	return model.EventGroup.ActivityID != nil && model.Activity == nil
}
func (model EventGroupsBundle) DeleteNotificationEvent() bool {
	return model.EventGroup.NotificationEventID != nil && model.NotificationEvent == nil
}
