package sitemodels

type SiteActivity struct {
	ID string `json:"id"`
	// ActivityType is e.g. exercise or walking, exercise when none is sent
	ActivityType          string  `json:"activity_type"`
	Name                  *string `json:"name"`
	DurationMinutes       *int64  `json:"duration_minutes"`
	Calories              *int64  `json:"calories"`
	Steps                 *int64  `json:"steps"`
	HeartRate             *int64  `json:"heart_rate"`
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
}
