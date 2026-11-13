package settings

import "nighttroop.t1d.nz/utils/i18n"

// orders are settings each of which must be under another, compared as the second times by, e.g. 60 for an
// insulin's peak in minutes under its duration in hours
var orders = []struct {
	under, over string
	by          float64
	rule        string // the id of the message saying the order
}{
	{"range_urgent_low_mgdl", "range_low_mgdl", 1, "errors.settings.urgent_low_under_low"},
	{"range_low_mgdl", "range_high_mgdl", 1, "errors.settings.low_under_high"},
	{"range_high_mgdl", "range_urgent_high_mgdl", 1, "errors.settings.high_under_urgent_high"},
	{"bolus_onset_minutes", "bolus_peak_minutes", 1, "errors.settings.bolus_onset"},
	{"bolus_peak_minutes", "bolus_dia_hours", 60, "errors.settings.bolus_peak"},
	{"basal_onset_minutes", "basal_peak_minutes", 1, "errors.settings.basal_onset"},
	{"basal_peak_minutes", "basal_dia_hours", 60, "errors.settings.basal_peak"},
}

// Consistent is whether the settings in effect, each the identity's, the site's or the default, make sense
// together, e.g. a low alarm under the high alarm
func (m *CombinedSettingsMap) Consistent() error {
	for _, order := range orders {
		under, err := m.GetSettingAnyTypeByName(order.under)
		if err != nil {
			return err
		}
		over, err := m.GetSettingAnyTypeByName(order.over)
		if err != nil {
			return err
		}
		if number(under.Value) >= number(over.Value)*order.by {
			return i18n.NewError(order.rule)
		}
	}
	return nil
}

func number(value any) float64 {
	switch v := value.(type) {
	case int64:
		return float64(v)
	case float64:
		return v
	}
	return 0
}
