package apiutils

import "nighttroop.t1d.nz/utils/i18n"

// CheckRange refuses a value sent outside min to max with the message with the id, which names the value and its unit
// and has the range as {{.Min}} and {{.Max}}; a value not sent passes
func CheckRange[T int64 | float64](id string, value *T, min, max T) *HTTPError {
	if value == nil || (*value >= min && *value <= max) {
		return nil
	}
	return ErrConstraintCheckFailedWithHint(i18n.NewError(id, i18n.Data{"Min": min, "Max": max}))
}
