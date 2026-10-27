package apiutils

import "fmt"

// CheckRange refuses a value sent outside min to max, naming it and its unit in the hint; a value not sent passes
func CheckRange[T int64 | float64](name string, value *T, min, max T, unit string) *HTTPError {
	if value == nil || (*value >= min && *value <= max) {
		return nil
	}
	hint := fmt.Sprintf("%s must be from %v to %v", name, min, max)
	if unit != "" {
		hint += " " + unit
	}
	return ErrConstraintCheckFailedWithHint(hint)
}
