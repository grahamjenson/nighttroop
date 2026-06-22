package nightscout

type Food struct {
	Identifier *string `json:"identifier,omitempty"`

	Name        *string `json:"name,omitempty"`
	Category    *string `json:"category,omitempty"`
	Subcategory *string `json:"subcategory,omitempty"`

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

	RawJSON string `json:"-"`
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
