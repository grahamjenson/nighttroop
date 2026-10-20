package sitemodels

type SiteTempBasal struct {
	ID              string `json:"id"`
	DurationMinutes int64  `json:"duration_minutes"`
	// Absolute is U/hr, a pointer so 0, a suspended basal, can be sent
	Absolute              *float64 `json:"absolute"`
	Percent               *float64 `json:"percent"`
	Notes                 *string  `json:"notes"`
	Reason                *string  `json:"reason"`
	RawJsonData           *string  `json:"raw_json_data"`
	UtcOffsetMinutes      *int64   `json:"utc_offset_minutes"`
	OccurredAt            int64    `json:"occurred_at"`
	CreatedAt             int64    `json:"created_at"`
	UpdatedAt             int64    `json:"updated_at"`
	CreatedByAppName      *string  `json:"created_by_app_name"`
	CreatedByDeviceName   *string  `json:"created_by_device_name"`
	CreatedByIdentityName string   `json:"created_by_identity_name"`
	UpdatedByIdentityName *string  `json:"updated_by_identity_name"`
}
