package consts

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
)

type Units string

const (
	UnitsMGDL  Units = "mg/dL"
	UnitsMMOLL Units = "mmol/L"
)
const MMOL_L_TO_MG_DL = 18.018

var unitAliases = map[string]Units{
	// mg/dL
	"mg":      UnitsMGDL,
	"mgdl":    UnitsMGDL,
	"mgperdl": UnitsMGDL,

	// mmol/L
	"mmol":     UnitsMMOLL,
	"mmoll":    UnitsMMOLL,
	"mmolperl": UnitsMMOLL,
}

func ParseUnits(s string) (Units, error) {
	// Clean up the string a bit
	key := strings.ToLower(s)
	key = strings.ReplaceAll(key, " ", "")
	key = strings.ReplaceAll(key, "/", "")
	key = strings.ReplaceAll(key, `\`, "")

	if units, ok := unitAliases[key]; ok {
		return units, nil
	}

	return "", fmt.Errorf("invalid glucose units: %q", s)
}

func (t *Units) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return fmt.Errorf("expected units string, got %s", data)
	}

	parsed, err := ParseUnits(str)
	if err != nil {
		return err
	}
	*t = parsed
	return nil
}

func (u *Units) Stringp() *string {
	if u == nil {
		return nil
	}
	str := string(*u)
	return &str
}

func (u *Units) FromMGDL(mgdl float64) float64 {
	if u == nil {
		u = ptr(UnitsMGDL)
	}

	if *u == UnitsMGDL {
		return Round(mgdl, 3)
	}

	return Round(mgdl/MMOL_L_TO_MG_DL, 4)
}

// This will do its best to conver to MGDL
func (u *Units) ToMGDL(unitValue float64) float64 {
	if u == nil {
		// Dont know the unit, better guess
		u = ptr(GuessUnits(unitValue))
	}
	if *u == UnitsMGDL {
		return Round(unitValue, 3)
	}

	return Round(unitValue*MMOL_L_TO_MG_DL, 3)
}

func (u *Units) ToMGDLMultiplier() float64 {
	if u == nil {
		return 1.0
	}
	switch *u {
	case UnitsMGDL:
		return 1.0
	case UnitsMMOLL:
		return MMOL_L_TO_MG_DL
	default:
		return 1.0
	}
}

func Round(value float64, dp int) float64 {
	m := math.Pow10(dp)
	return math.Round(value*m) / m
}

func GuessUnits(value float64) Units {
	if value > 30 {
		return UnitsMGDL
	}
	return UnitsMMOLL
}

func ptr[T any](v T) *T {
	return &v
}
