//go:build !goverter

// goverter builds this package without the converters it makes, so it leaves out this file, which calls them, too

package sitedb

// GetEventGroupToEventGroupsBundle is an event group's row as its bundle, with each event the row joined
func GetEventGroupToEventGroupsBundle(source GetEventGroupRow) EventGroupsBundle {
	egb := EventGroupsBundle{
		EventGroup: source.EventGroup,
	}

	if source.GlucoseReadingsVw.ID != nil {
		egb.GlucoseReading = new(GlucoseReadingFromView(source.GlucoseReadingsVw))
	}

	if source.InsulinDosesVw.ID != nil {
		egb.InsulinDose = new(InsulinDoseFromView(source.InsulinDosesVw))
	}

	if source.FoodIntakesVw.ID != nil {
		egb.FoodIntake = new(FoodIntakeFromView(source.FoodIntakesVw))
	}

	if source.TempBasalsVw.ID != nil {
		egb.TempBasal = new(TempBasalFromView(source.TempBasalsVw))
	}

	if source.TempTargetsVw.ID != nil {
		egb.TempTarget = new(TempTargetFromView(source.TempTargetsVw))
	}

	if source.ProfileSwitchesVw.ID != nil {
		egb.ProfileSwitch = new(ProfileSwitchFromView(source.ProfileSwitchesVw))
	}

	if source.ComponentEventsVw.ID != nil {
		egb.ComponentEvent = new(ComponentEventFromView(source.ComponentEventsVw))
	}

	if source.ActivitiesVw.ID != nil {
		egb.Activity = new(ActivityFromView(source.ActivitiesVw))
	}

	if source.NotificationEventsVw.ID != nil {
		egb.NotificationEvent = new(NotificationEventFromView(source.NotificationEventsVw))
	}

	return egb
}
