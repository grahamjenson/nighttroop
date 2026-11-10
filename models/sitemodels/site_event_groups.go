package sitemodels

import "nighttroop.t1d.nz/models/nightscoutmodels"

// SiteEventGroup is a Nightscout treatment: one or more events recorded together, like a meal bolus's food and insulin.
// Each event, and its ID, is set only when the group has one. Creating or editing a group picks its events by ID.
type SiteEventGroup struct {
	ID                    string  `json:"id"`
	GroupType             string  `json:"group_type"`
	DurationMinutes       *int64  `json:"duration_minutes"`
	Notes                 *string `json:"notes"`
	Reason                *string `json:"reason"`
	UtcOffsetMinutes      *int64  `json:"utc_offset_minutes"`
	OccurredAt            int64   `json:"occurred_at"`
	CreatedAt             int64   `json:"created_at"`
	UpdatedAt             int64   `json:"updated_at"`
	CreatedByAppName      *string `json:"created_by_app_name"`
	CreatedByDeviceName   *string `json:"created_by_device_name"`
	CreatedByIdentityName string  `json:"created_by_identity_name"`
	UpdatedByIdentityName *string `json:"updated_by_identity_name"`
	// The pump the app named, as AAPS does on every treatment; the site does not set it
	PumpID     *string `json:"pump_id"`
	PumpSerial *string `json:"pump_serial"`
	PumpType   *string `json:"pump_type"`

	GlucoseReading      *SiteGlucoseReading `json:"glucose_reading"`
	InsulinDose         *SiteInsulinDose    `json:"insulin_dose"`
	FoodIntake          *SiteFoodIntake     `json:"food_intake"`
	TempBasal           *SiteTempBasal      `json:"temp_basal"`
	TempTarget          *SiteTempTarget     `json:"temp_target"`
	GlucoseReadingID    *string             `json:"glucose_reading_id"`
	InsulinDoseID       *string             `json:"insulin_dose_id"`
	FoodIntakeID        *string             `json:"food_intake_id"`
	TempBasalID         *string             `json:"temp_basal_id"`
	TempTargetID        *string             `json:"temp_target_id"`
	ProfileSwitchID     *string             `json:"profile_switch_id"`
	ComponentEventID    *string             `json:"component_event_id"`
	ActivityID          *string             `json:"activity_id"`
	NotificationEventID *string             `json:"notification_event_id"`
}

// SiteEventChoice is an event that can be put in a group, as its edit or create form lists them. Kind is the group's
// field for it, like glucose_reading_id. Amount is its glucose (in the settings' units), dose, carbs, rate or low
// target, Minutes how long it lasts, and Detail what else tells it apart, like its type or name.
type SiteEventChoice struct {
	Kind       string   `json:"kind"`
	ID         string   `json:"id"`
	OccurredAt int64    `json:"occurred_at"`
	Amount     *float64 `json:"amount"`
	Minutes    *int64   `json:"minutes"`
	Detail     *string  `json:"detail"`
	// Picked is whether the group is already of it
	Picked bool `json:"picked"`
}

// EventGroupTypeKinds are the kinds of event, by the group's field for each, that each type of group can have; the
// Nightscout API drops a treatment's others, and the site refuses them
var EventGroupTypeKinds = map[nightscoutmodels.TreatmentEventType][]string{
	nightscoutmodels.Announcement:      {"glucose_reading_id", "notification_event_id"},
	nightscoutmodels.BGCheck:           {"glucose_reading_id"},
	nightscoutmodels.BolusWizard:       {"glucose_reading_id", "insulin_dose_id", "food_intake_id"},
	nightscoutmodels.CarbCorrection:    {"food_intake_id"},
	nightscoutmodels.ComboBolus:        {"glucose_reading_id", "insulin_dose_id", "food_intake_id"},
	nightscoutmodels.CorrectionBolus:   {"insulin_dose_id", "glucose_reading_id"},
	nightscoutmodels.DADAlert:          {},
	nightscoutmodels.Exercise:          {"activity_id"},
	nightscoutmodels.InsulinChange:     {"component_event_id"},
	nightscoutmodels.MealBolus:         {"glucose_reading_id", "insulin_dose_id", "food_intake_id"},
	nightscoutmodels.Note:              {"glucose_reading_id", "insulin_dose_id", "food_intake_id", "profile_switch_id", "activity_id", "component_event_id"},
	nightscoutmodels.OpenAPSOffline:    {},
	nightscoutmodels.ProfileSwitch:     {"profile_switch_id"},
	nightscoutmodels.PumpBatteryChange: {"component_event_id"},
	nightscoutmodels.Question:          {"notification_event_id"},
	nightscoutmodels.SensorChange:      {"component_event_id"},
	nightscoutmodels.SensorStart:       {"component_event_id"},
	nightscoutmodels.SensorStop:        {"component_event_id"},
	nightscoutmodels.SiteChange:        {"component_event_id"},
	nightscoutmodels.SnackBolus:        {"glucose_reading_id", "insulin_dose_id", "food_intake_id"},
	nightscoutmodels.TempBasal:         {"temp_basal_id"},
	nightscoutmodels.TemporaryTarget:   {"temp_target_id"},
	nightscoutmodels.TemporaryOverride: {"temp_target_id"},
	nightscoutmodels.SuspendPump:       {},
	nightscoutmodels.ResumePump:        {},
	nightscoutmodels.CannulaFill:       {"component_event_id"},
	nightscoutmodels.Prime:             {"component_event_id"},
	nightscoutmodels.ReservoirChange:   {"component_event_id"},
	nightscoutmodels.Calibration:       {"glucose_reading_id"},
	nightscoutmodels.Bolus:             {"insulin_dose_id"},
	nightscoutmodels.SMB:               {"insulin_dose_id"},
	nightscoutmodels.ExternalInsulin:   {"insulin_dose_id"},
}

// EventGroupTypes are the types of group the site can add, those that can have an event, in Nightscout's order
var EventGroupTypes = func() []string {
	types := []string{}
	for _, t := range nightscoutmodels.TreatmentEventTypes {
		if len(EventGroupTypeKinds[t]) > 0 {
			types = append(types, string(t))
		}
	}
	return types
}()

// EventPicks are the group's events picked, by its field for each kind
func (g *SiteEventGroup) EventPicks() map[string]*string {
	return map[string]*string{"glucose_reading_id": g.GlucoseReadingID, "insulin_dose_id": g.InsulinDoseID,
		"food_intake_id": g.FoodIntakeID, "temp_basal_id": g.TempBasalID, "temp_target_id": g.TempTargetID,
		"profile_switch_id": g.ProfileSwitchID, "component_event_id": g.ComponentEventID, "activity_id": g.ActivityID,
		"notification_event_id": g.NotificationEventID}
}
