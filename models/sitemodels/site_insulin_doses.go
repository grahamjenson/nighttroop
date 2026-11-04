package sitemodels

type SiteInsulinDose struct {
	ID                    string   `json:"id"`
	Dose                  float64  `json:"dose"`
	DoseType              *string  `json:"dose_type"`
	SplitNow              *float64 `json:"split_now"` // the fraction of a combo bolus given at once, so 0.7 is 70%
	SplitExt              *float64 `json:"split_ext"` // the fraction of it given over time
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
