package sitemodels

// SiteFoodPreset is a food kept to pick from, with what one portion of it has
type SiteFoodPreset struct {
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
	RawJsonData           *string  `json:"raw_json_data"`
	CreatedAt             int64    `json:"created_at"`
	UpdatedAt             int64    `json:"updated_at"`
	CreatedByAppName      *string  `json:"created_by_app_name"`
	CreatedByDeviceName   *string  `json:"created_by_device_name"`
	CreatedByIdentityName string   `json:"created_by_identity_name"`
	UpdatedByIdentityName *string  `json:"updated_by_identity_name"`
}
