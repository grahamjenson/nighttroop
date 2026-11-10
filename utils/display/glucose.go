package display

import (
	"fmt"
	"math"
	"strconv"

	"nighttroop.t1d.nz/utils/consts"
)

// Percent is a percent, as a time in range, rounded to a whole one, "<1%" for one above 0 that rounds to 0, and "–"
// for none, so one reads the same on every page
func Percent(v *float64) string {
	switch {
	case v == nil:
		return "–"
	case *v > 0 && math.Round(*v) == 0:
		return "<1%"
	}
	return fmt.Sprintf("%.0f%%", math.Round(*v))
}

// Glucose is glucose in its units: mmol/L to one decimal place, and mg/dL, or glucose of no units, as a whole number
func Glucose(v float64, units *consts.Units) string {
	if units != nil && *units == consts.UnitsMMOLL {
		return strconv.FormatFloat(v, 'f', 1, 64)
	}
	return strconv.FormatFloat(v, 'f', 0, 64)
}

// GlucoseOr is Glucose, or "–" for none
func GlucoseOr(v *float64, units *consts.Units) string {
	if v == nil {
		return "–"
	}
	return Glucose(*v, units)
}

// Units are glucose's units as they are written, or "" for none
func Units(units *consts.Units) string {
	if units == nil {
		return ""
	}
	return string(*units)
}

// GlucoseRange is "Low" for a reading below low, "High" for one above high, and "" for one between them
func GlucoseRange(reading, low, high float64) string {
	switch {
	case reading < low:
		return "Low"
	case reading > high:
		return "High"
	}
	return ""
}

// Scale is how far a value is from bottom to top, as a fraction from 0 to 1, one outside them kept at their end, for
// drawing it on a graph's axis
func Scale(value, bottom, top float64) float64 {
	return (min(max(value, bottom), top) - bottom) / (top - bottom)
}
