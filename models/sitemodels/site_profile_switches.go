package sitemodels

import "nighttroop.t1d.nz/utils/consts"

type SiteProfileSwitch struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	// DurationMinutes is a pointer so 0, until the next switch, can be sent
	DurationMinutes *int64 `json:"duration_minutes"`
	// Multiplier is how many times the profile's insulin, so 1.2 is 120%
	Multiplier *float64 `json:"multiplier"`
	// Timeshift is in minutes, e.g. 60 shifts the profile forward an hour
	Timeshift *int64 `json:"timeshift"`
	// The profile's settings and schedules are only read on the site
	Dia                      *float64      `json:"dia"`
	Units                    *consts.Units `json:"units"`
	Timezone                 *string       `json:"timezone"`
	CarbsHr                  *float64      `json:"carbs_hr"`
	Delay                    *float64      `json:"delay"`
	CarbRatiosJson           *string       `json:"carb_ratios_json"`
	InsulinSensitivitiesJson *string       `json:"insulin_sensitivities_json"`
	BasalRatesJson           *string       `json:"basal_rates_json"`
	TargetsHighJson          *string       `json:"targets_high_json"`
	TargetsLowJson           *string       `json:"targets_low_json"`
	Notes                    *string       `json:"notes"`
	Reason                   *string       `json:"reason"`
	UtcOffsetMinutes         *int64        `json:"utc_offset_minutes"`
	OccurredAt               int64         `json:"occurred_at"`
	CreatedAt                int64         `json:"created_at"`
	UpdatedAt                int64         `json:"updated_at"`
	CreatedByAppName         *string       `json:"created_by_app_name"`
	CreatedByDeviceName      *string       `json:"created_by_device_name"`
	CreatedByIdentityName    string        `json:"created_by_identity_name"`
	UpdatedByIdentityName    *string       `json:"updated_by_identity_name"`
}
