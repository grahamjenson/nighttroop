package types

import "database/sql/driver"

// FoodIntakeOther are a food intake's other attributes: Nightscout's food fields beyond its carbs, such as its fat,
// protein and the minutes bolused before eating
type FoodIntakeOther struct {
	Category    *string  `json:"category,omitempty"`
	Subcategory *string  `json:"subcategory,omitempty"`
	Fat         *float64 `json:"fat,omitempty"`
	Protein     *float64 `json:"protein,omitempty"`
	Fibre       *float64 `json:"fibre,omitempty"`
	Energy      *float64 `json:"energy,omitempty"`
	Portions    *float64 `json:"portions,omitempty"`
	PreBolus    *float64 `json:"pre_bolus,omitempty"`
}

func (o FoodIntakeOther) Value() (driver.Value, error) { return otherValue(o) }

func (o *FoodIntakeOther) Scan(src any) error {
	*o = FoodIntakeOther{}
	return otherScan(src, o)
}
