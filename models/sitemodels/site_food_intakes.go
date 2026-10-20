package sitemodels

type SiteFoodIntake struct {
	ID                    string   `json:"id"`
	Name                  *string  `json:"name"`
	FoodType              *string  `json:"food_type"`
	Category              *string  `json:"category"`
	Subcategory           *string  `json:"subcategory"`
	AbsorptionTimeMinutes *int64   `json:"absorption_time_minutes"`
	PreBolus              *float64 `json:"pre_bolus"`
	Carbs                 *float64 `json:"carbs"`
	Fat                   *float64 `json:"fat"`
	Protein               *float64 `json:"protein"`
	Fibre                 *float64 `json:"fibre"`
	Energy                *float64 `json:"energy"`
	Portions              *float64 `json:"portions"`
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
