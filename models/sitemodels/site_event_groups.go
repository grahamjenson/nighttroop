package sitemodels

// SiteEventGroup is a Nightscout treatment: one or more events recorded together, like a meal bolus's food and insulin.
// Each event, and its ID, is set only when the group has one. Creating or editing a group picks its events by ID.
type SiteEventGroup struct {
	ID                    string  `json:"id"`
	GroupType             string  `json:"group_type"`
	DurationMinutes       *int64  `json:"duration_minutes"`
	Notes                 *string `json:"notes"`
	Reason                *string `json:"reason"`
	RawJsonData           *string `json:"raw_json_data"`
	UtcOffsetMinutes      *int64  `json:"utc_offset_minutes"`
	OccurredAt            int64   `json:"occurred_at"`
	CreatedAt             int64   `json:"created_at"`
	UpdatedAt             int64   `json:"updated_at"`
	CreatedByAppName      *string `json:"created_by_app_name"`
	CreatedByDeviceName   *string `json:"created_by_device_name"`
	CreatedByIdentityName string  `json:"created_by_identity_name"`
	UpdatedByIdentityName *string `json:"updated_by_identity_name"`

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
