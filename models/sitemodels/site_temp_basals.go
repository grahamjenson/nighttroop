package sitemodels

type SiteTempBasal struct {
	ID              string `json:"id"`
	DurationMinutes int64  `json:"duration_minutes"`
	// Rate is the U/hr given while it runs, a pointer so 0, a suspended basal, can be sent
	Rate *float64 `json:"rate"`
	// Multiplier is how many times the scheduled basal at its start, so 0.5 is half and 0 a zero temp
	Multiplier            *float64 `json:"multiplier"`
	Notes                 *string  `json:"notes"`
	Reason                *string  `json:"reason"`
	UtcOffsetMinutes      *int64   `json:"utc_offset_minutes"`
	OccurredAt            int64    `json:"occurred_at"`
	CreatedAt             int64    `json:"created_at"`
	UpdatedAt             int64    `json:"updated_at"`
	CreatedByAppName      *string  `json:"created_by_app_name"`
	CreatedByDeviceName   *string  `json:"created_by_device_name"`
	CreatedByIdentityName string   `json:"created_by_identity_name"`
	UpdatedByIdentityName *string  `json:"updated_by_identity_name"`
}
