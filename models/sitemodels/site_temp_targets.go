package sitemodels

import "nighttroop.t1d.nz/utils/consts"

type SiteTempTarget struct {
	ID              string `json:"id"`
	DurationMinutes int64  `json:"duration_minutes"`
	// The targets are in Units, the identity's, site's or default units setting
	HighTarget              float64       `json:"high_target"`
	LowTarget               float64       `json:"low_target"`
	Units                   *consts.Units `json:"units"`
	InsulinNeedsScaleFactor *float64      `json:"insulin_needs_scale_factor"`
	Notes                   *string       `json:"notes"`
	Reason                  *string       `json:"reason"`
	UtcOffsetMinutes        *int64        `json:"utc_offset_minutes"`
	OccurredAt              int64         `json:"occurred_at"`
	CreatedAt               int64         `json:"created_at"`
	UpdatedAt               int64         `json:"updated_at"`
	CreatedByAppName        *string       `json:"created_by_app_name"`
	CreatedByDeviceName     *string       `json:"created_by_device_name"`
	CreatedByIdentityName   string        `json:"created_by_identity_name"`
	UpdatedByIdentityName   *string       `json:"updated_by_identity_name"`
}
