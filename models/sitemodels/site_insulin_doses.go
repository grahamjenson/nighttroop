package sitemodels

type SiteInsulinDose struct {
	ID                    string  `json:"id"`
	Dose                  float64 `json:"dose"`
	DoseType              *string `json:"dose_type"`
	Automatic             *bool   `json:"automatic"` // whether a loop gave it, as Loop sends, so it is an SMB
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

	// ExtendedDose is a combo or extended bolus's units given evenly over ExtendedDurationMinutes after the dose
	ExtendedDose            *float64 `json:"extended_dose"`
	ExtendedDurationMinutes *int64   `json:"extended_duration_minutes"`
}

// MsToMinutes are milliseconds as the nearest whole minutes
func MsToMinutes(ms *int64) *int64 {
	if ms == nil {
		return nil
	}
	return new((*ms + 30_000) / 60_000)
}

// MinutesToMs are minutes as milliseconds
func MinutesToMs(minutes *int64) *int64 {
	if minutes == nil {
		return nil
	}
	return new(*minutes * 60_000)
}
