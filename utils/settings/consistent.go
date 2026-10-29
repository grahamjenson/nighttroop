package settings

import "fmt"

// orders are settings each of which must be under another, compared as the second times by, e.g. 60 for an
// insulin's peak in minutes under its duration in hours
var orders = []struct {
	under, over string
	by          float64
	rule        string
}{
	{"range_urgent_low_mgdl", "range_low_mgdl", 1, "the urgent low must be under the low"},
	{"range_low_mgdl", "range_high_mgdl", 1, "the low must be under the high"},
	{"range_high_mgdl", "range_urgent_high_mgdl", 1, "the high must be under the urgent high"},
	{"bolus_onset_minutes", "bolus_peak_minutes", 1, "bolus insulin must start acting before its peak"},
	{"bolus_peak_minutes", "bolus_dia_hours", 60, "bolus insulin must peak before it stops acting"},
	{"basal_onset_minutes", "basal_peak_minutes", 1, "basal insulin must start acting before its peak"},
	{"basal_peak_minutes", "basal_dia_hours", 60, "basal insulin must peak before it stops acting"},
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
			return fmt.Errorf("%s", order.rule)
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
