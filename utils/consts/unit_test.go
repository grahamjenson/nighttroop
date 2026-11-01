package consts_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"nighttroop.t1d.nz/utils/consts"
)

func TestUnitConversion(t *testing.T) {
	mgdl, mmol := consts.UnitsMGDL, consts.UnitsMMOLL
	tests := []struct {
		name    string
		convert func(float64) float64
		value   float64
		want    float64
	}{
		{"mmol/L to mg/dL", mmol.ToMGDL, 3.885, 70},
		{"mmol/L to mg/dL", mmol.ToMGDL, 4.44, 80},
		{"mg/dL to mg/dL", mgdl.ToMGDL, 70, 70},
		{"mg/dL to mg/dL", mgdl.ToMGDL, 80, 80},
		{"no units, guessed mmol/L, to mg/dL", (*consts.Units)(nil).ToMGDL, 3.885, 70},
		{"no units, guessed mmol/L, to mg/dL", (*consts.Units)(nil).ToMGDL, 4.44, 80},
		{"no units, guessed mg/dL, to mg/dL", (*consts.Units)(nil).ToMGDL, 70, 70},
		{"no units, guessed mg/dL, to mg/dL", (*consts.Units)(nil).ToMGDL, 80, 80},
		{"mg/dL to mmol/L", mmol.FromMGDL, 70, 3.885},
		{"mg/dL to mmol/L", mmol.FromMGDL, 80, 4.44},
		{"mg/dL to mg/dL", mgdl.FromMGDL, 70, 70},
		{"mg/dL to mg/dL", mgdl.FromMGDL, 80, 80},
		{"mg/dL to no units, as mg/dL", (*consts.Units)(nil).FromMGDL, 70, 70},
	}
	for _, tt := range tests {
		assert.InDelta(t, tt.want, tt.convert(tt.value), 0.01, "%s %v", tt.name, tt.value)
	}
}

// Going back and forth between mg/dL and mmol/L many times keeps the value, as rounding does not add up
func TestUnitConversionBackAndForth(t *testing.T) {
	mmol := consts.UnitsMMOLL
	for mgdlValue, mmolValue := range map[float64]float64{70: 3.885, 80: 4.44} {
		value := mmol.FromMGDL(mgdlValue)
		for range 3 {
			back := mmol.ToMGDL(value)
			assert.InDelta(t, mgdlValue, back, 0.01, "%v mg/dL", mgdlValue)
			value = mmol.FromMGDL(back)
			assert.InDelta(t, mmolValue, value, 0.01, "%v mmol/L", mmolValue)
		}
	}
}
