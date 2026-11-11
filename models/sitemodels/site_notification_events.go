package sitemodels

type SiteNotificationEvent struct {
	ID string `json:"id"`
	// NotificationType is e.g. announcement or question, announcement when none is sent
	NotificationType string `json:"notification_type"`
	Title            string `json:"title"`
	// Severity is from 0, the lowest, to 5, the highest
	Severity              int64   `json:"severity"`
	Content               *string `json:"content"`
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
}
