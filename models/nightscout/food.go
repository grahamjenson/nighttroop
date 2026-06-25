package nightscout

import (
	"encoding/json"
	"time"
)

type Food struct {
	Identifier *IdentifierString `json:"identifier,omitempty"`

	Name        *SanitizedString `json:"name,omitempty"`
	Category    *SanitizedString `json:"category,omitempty"`
	Subcategory *SanitizedString `json:"subcategory,omitempty"`

	Portion *FlexibleFloat64 `json:"portion,omitempty"`
	Unit    *FoodUnit        `json:"unit,omitempty"`

	Carbs   *FlexibleFloat64 `json:"carbs,omitempty"`
	Fat     *FlexibleFloat64 `json:"fat,omitempty"`
	Protein *FlexibleFloat64 `json:"protein,omitempty"`
	Energy  *FlexibleFloat64 `json:"energy,omitempty"`
	GI      *FlexibleFloat64 `json:"gi,omitempty"`

	// API v3 additions
	SrvCreated  *DateInt `json:"srvCreated,omitempty"`
	SrvModified *DateInt `json:"srvModified,omitempty"`

	// Non-JSON values to make mapping easier
	Date        *DateInt `json:"-"` // Unix timestamp in milliseconds
	RawJsonData string   `json:"-"`
}

func (t *Food) UnmarshalJSON(data []byte) error {
	type Alias Food
	aux := struct {
		ID IdentifierString `json:"_id"`
		*Alias
	}{
		Alias: (*Alias)(t),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Raw Value
	t.RawJsonData = string(data)

	t.Date = ptr(DateInt(time.Now().UnixMilli()))

	// Identitfiers
	if t.Identifier == nil {
		t.Identifier = &aux.ID
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
