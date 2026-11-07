package types

import "database/sql/driver"

// FoodPresetOther are a food preset's other attributes: Nightscout's food fields beyond its carbs, such as its fat,
// protein and the minutes bolused before eating
type FoodPresetOther struct {
	Category    *string  `json:"category,omitempty"`
	Subcategory *string  `json:"subcategory,omitempty"`
	Fat         *float64 `json:"fat,omitempty"`
	Protein     *float64 `json:"protein,omitempty"`
	Fibre       *float64 `json:"fibre,omitempty"`
	Energy      *float64 `json:"energy,omitempty"`
	Portions    *float64 `json:"portions,omitempty"`
	PreBolus    *float64 `json:"pre_bolus,omitempty"`
}

func (o FoodPresetOther) Value() (driver.Value, error) { return otherValue(o) }

func (o *FoodPresetOther) Scan(src any) error {
	*o = FoodPresetOther{}
	return otherScan(src, o)
}
