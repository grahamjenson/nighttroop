package sitemodels

// SiteEventGroup is a Nightscout treatment: one or more events recorded together, like a meal bolus's food and insulin.
// Each event is set only when the group has one.
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
	ProfileSwitchID     *string             `json:"profile_switch_id"`
	ComponentEventID    *string             `json:"component_event_id"`
	ActivityID          *string             `json:"activity_id"`
	NotificationEventID *string             `json:"notification_event_id"`
}
