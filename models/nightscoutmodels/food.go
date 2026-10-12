package nightscoutmodels

import (
	"encoding/json"
)

type Food struct {
	Identifier *IdentifierString `json:"identifier,omitempty"`

	Name *ShortString `json:"name,omitempty"`

	FoodType    *ShortString `json:"foodType,omitempty"` // Type of food
	Category    *ShortString `json:"category,omitempty"`
	Subcategory *ShortString `json:"subcategory,omitempty"`

	Unit *FoodUnit `json:"unit,omitempty"`

	AbsorptionTimeMinutes *FlexibleInt64   `json:"absorptionTime,omitempty"`
	PreBolus              *FlexibleFloat64 `json:"preBolus,omitempty"`
	Carbs                 *FlexibleFloat64 `json:"carbs,omitempty"`
	Fat                   *FlexibleFloat64 `json:"fat,omitempty"`
	Protein               *FlexibleFloat64 `json:"protein,omitempty"`
	Fibre                 *FlexibleFloat64 `json:"fibre,omitempty"`
	Energy                *FlexibleFloat64 `json:"energy,omitempty"`
	GI                    *FlexibleFloat64 `json:"gi,omitempty"`
	Portions              *FlexibleFloat64 `json:"portions,omitempty"`

	// QuickPick
	HideAfterUse *FlexibleBool  `json:"hideafteruse,omitempty"`
	Hidden       *FlexibleBool  `json:"hidden,omitempty"`
	Position     *FlexibleInt64 `json:"position,omitempty"`

	// Default String Fields
	Device  *ShortString `json:"device,omitempty"`  // Device used for the activity
	App     *ShortString `json:"app,omitempty"`     // App used for the activity
	Subject *ShortString `json:"subject,omitempty"` // Subject of the activity

	// Non-JSON values to make mapping easier
	Date        *DateInt `json:"-"` // Unix timestamp in milliseconds
	RawJsonData string   `json:"-"`
}

func (t *Food) UnmarshalJSON(data []byte) error {
	type Alias Food
	aux := struct {
		ID *IdentifierString `json:"_id"`
		*Alias
	}{
		Alias: (*Alias)(t),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Raw Value
	t.RawJsonData = string(data)

	// Identitfiers
	if t.Identifier == nil {
		t.Identifier = aux.ID
	}

	return nil
}

type FoodUnit string

const (
	UnitGram      FoodUnit = "g"
	UnitOunce     FoodUnit = "oz"
	UnitPortion   FoodUnit = "portion"
	UnitServing   FoodUnit = "serving"
	UnitItem      FoodUnit = "item"
	UnitPackage   FoodUnit = "package"
	UnitContainer FoodUnit = "container"
)

func (t Food) GetDevice() *string {
	return t.Device.Stringp()
}

func (t Food) GetApp() *string {
	return t.App.Stringp()
}

func (t Food) GetSubject() *string {
	return t.Subject.Stringp()
}

func (t Food) GetDate() *int64 {
	return t.Date.Int64p()
}

func (t Food) GetUtcOffsetMinutes() *int64 {
	return nil
}

func (t Food) GetIdentifier() *string {
	return t.Identifier.Stringp()
}
