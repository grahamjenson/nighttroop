package consts

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"
)

type Units string

const (
	UnitsMGDL  Units = "mg/dL"
	UnitsMMOLL Units = "mmol/L"
)
const MMOL_L_TO_MG_DL = 18.018

// MaxGlucoseMgdl is far above any real reading, so a typo or a made up value is refused
const MaxGlucoseMgdl = 1000

// MaxFutureReading is how far ahead of now a reading can be dated, for a phone clock's error
const MaxFutureReading = 15 * time.Minute

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
		u = new(UnitsMGDL)
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
		u = new(GuessUnits(unitValue))
	}
	if *u == UnitsMGDL {
		return Round(unitValue, 3)
	}

	return Round(unitValue*MMOL_L_TO_MG_DL, 3)
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
