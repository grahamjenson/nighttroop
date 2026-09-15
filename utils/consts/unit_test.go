package consts_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"nighttroop.t1d.nz/utils/consts"
)

func TestUnitConversionBackAndForth(t *testing.T) {
	mgdlTommol := map[float64]float64{
		70: 3.885,
		80: 4.44,
	}

	mgdl := consts.UnitsMGDL
	mmol := consts.UnitsMMOLL
	var nilu *consts.Units

	for mgdlValue, mmolValue := range mgdlTommol {
		// TO MGDL
		// MMol to MGDL conversion
		mgdl1 := mmol.ToMGDL(mmolValue)
		assert.InDelta(t, mgdlValue, mgdl1, 0.01)

		mgdl2 := mgdl.ToMGDL(mgdlValue)
		assert.InDelta(t, mgdlValue, mgdl2, 0.01)

		mgdl3 := nilu.ToMGDL(mmolValue)
		assert.InDelta(t, mgdlValue, mgdl3, 0.01)

		mgdl4 := nilu.ToMGDL(mgdlValue)
		assert.InDelta(t, mgdlValue, mgdl4, 0.01)

		// FROM MGDL
		fmmol1 := mmol.FromMGDL(mgdlValue)
		assert.InDelta(t, mmolValue, fmmol1, 0.01)

		fmgdl1 := mgdl.FromMGDL(mgdlValue)
		assert.InDelta(t, mgdlValue, fmgdl1, 0.01)

		// Second Derivitave
		mgdlSecond := mmol.ToMGDL(fmmol1)
		assert.InDelta(t, mgdlValue, mgdlSecond, 0.01)

		mmolSecond := mmol.FromMGDL(mgdlSecond)
		assert.InDelta(t, mmolValue, mmolSecond, 0.01)

		// Third
		mgdlThird := mmol.ToMGDL(mmolSecond)
		assert.InDelta(t, mgdlValue, mgdlThird, 0.01)

		mmolThird := mmol.FromMGDL(mgdlThird)
		assert.InDelta(t, mmolValue, mmolThird, 0.01)

		// Fourth
		mgdlFourth := mmol.ToMGDL(mmolThird)
		assert.InDelta(t, mgdlValue, mgdlFourth, 0.01)

		mmolFourth := mmol.FromMGDL(mgdlFourth)
		assert.InDelta(t, mmolValue, mmolFourth, 0.01)
	}
}
