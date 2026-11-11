package nightscoutmodels

import (
	"encoding/json"

	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/utils/types"
)

type Food struct {
	Identifier *types.IdentifierString `json:"identifier,omitempty"`

	Name *types.ShortString `json:"name,omitempty"`
	// Type is Nightscout's food or quickpick; Nighttroop keeps foods, and AAPS loads only those of type food
	Type *types.ShortString `json:"type,omitempty"`

	FoodType    *types.ShortString `json:"foodType,omitempty"` // Type of food
	Category    *types.ShortString `json:"category,omitempty"`
	Subcategory *types.ShortString `json:"subcategory,omitempty"`

	Unit *FoodUnit `json:"unit,omitempty"`

	AbsorptionTimeMinutes *types.FlexibleInt64   `json:"absorptionTime,omitempty"`
	PreBolus              *types.FlexibleFloat64 `json:"preBolus,omitempty"`
	Carbs                 *types.FlexibleFloat64 `json:"carbs,omitempty"`
	Fat                   *types.FlexibleFloat64 `json:"fat,omitempty"`
	Protein               *types.FlexibleFloat64 `json:"protein,omitempty"`
	Fibre                 *types.FlexibleFloat64 `json:"fibre,omitempty"`
	Energy                *types.FlexibleFloat64 `json:"energy,omitempty"`
	GI                    *types.FlexibleFloat64 `json:"gi,omitempty"`
	Portions              *types.FlexibleFloat64 `json:"portions,omitempty"`

	// QuickPick
	HideAfterUse *types.FlexibleBool  `json:"hideafteruse,omitempty"`
	Hidden       *types.FlexibleBool  `json:"hidden,omitempty"`
	Position     *types.FlexibleInt64 `json:"position,omitempty"`

	// Default String Fields
	Device  *types.ShortString `json:"device,omitempty"`  // Device used for the activity
	App     *types.ShortString `json:"app,omitempty"`     // App used for the activity
	Subject *types.ShortString `json:"subject,omitempty"` // Subject of the activity

	// Non-JSON values to make mapping easier
	Date        *types.DateInt `json:"-"` // Unix timestamp in milliseconds
}

func (t *Food) UnmarshalJSON(data []byte) error {
	type Alias Food
	aux := struct {
		ID *types.IdentifierString `json:"_id"`
		*Alias
	}{
		Alias: (*Alias)(t),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}


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

// FoodDocumentType is a stored food's Nightscout type, food, as Nighttroop keeps no quickpicks
func FoodDocumentType(_ *sitedb.FoodPreset) *types.ShortString {
	return new(types.ShortString("food"))
}
