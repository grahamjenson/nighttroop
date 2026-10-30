package settings

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConsistent(t *testing.T) {
	tests := []struct {
		name     string
		site     map[string]any
		identity map[string]any
		want     string
	}{
		{name: "the defaults"},
		{name: "a low under the high", identity: map[string]any{"range_low_mgdl": "80", "range_high_mgdl": "200"}},
		{name: "a low over the high", identity: map[string]any{"range_low_mgdl": "150", "range_high_mgdl": "140"}, want: "the low must be under the high"},
		{name: "a low at the high", identity: map[string]any{"range_low_mgdl": "100", "range_high_mgdl": "100"}, want: "the low must be under the high"},
		{name: "the identity's low over the site's high", site: map[string]any{"range_high_mgdl": "120"}, identity: map[string]any{"range_low_mgdl": "130"}, want: "the low must be under the high"},
		{name: "an urgent low over the default low", identity: map[string]any{"range_urgent_low_mgdl": "80"}, want: "the urgent low must be under the low"},
		{name: "an urgent high under the default high", site: map[string]any{"range_urgent_high_mgdl": "150"}, want: "the high must be under the urgent high"},
		{name: "bolus onset after its peak", identity: map[string]any{"bolus_onset_minutes": "90", "bolus_peak_minutes": "60"}, want: "bolus insulin must start acting before its peak"},
		{name: "bolus peak after its duration", identity: map[string]any{"bolus_peak_minutes": "200", "bolus_dia_hours": "3"}, want: "bolus insulin must peak before it stops acting"},
		{name: "basal onset after its peak", site: map[string]any{"basal_onset_minutes": "200"}, want: "basal insulin must start acting before its peak"},
		{name: "basal peak after its duration", identity: map[string]any{"basal_dia_hours": "1", "basal_peak_minutes": "90"}, want: "basal insulin must peak before it stops acting"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			site, identity := &SettingMap{}, &SettingMap{}
			for name, value := range tt.site {
				require.NoError(t, site.SetSettingByName(name, value, nil, nil))
			}
			for name, value := range tt.identity {
				require.NoError(t, identity.SetSettingByName(name, value, nil, nil))
			}
			err := (&CombinedSettingsMap{SiteSettingsMap: site, IdentitySettingsMap: identity, DefaultSettingsMap: DefaultSettingsMap}).Consistent()
			if tt.want == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tt.want)
			}
		})
	}
}

// TestNoNumberSettingIsNegative checks every number setting refuses a negative value, such as minutes before now
func TestNoNumberSettingIsNegative(t *testing.T) {
	for _, name := range allSettingNames {
		setting := DefaultSettingsMap.GetSettingAnyTypeByName(name)
		if setting == nil || setting.TypeString != "number" {
			continue
		}
		t.Run(name, func(t *testing.T) {
			require.Error(t, (&SettingMap{}).SetSettingByName(name, "-1", nil, nil))
		})
	}
}

func TestNoDataAlarmIsAtLeastTenMinutes(t *testing.T) {
	require.Error(t, (&SettingMap{}).SetSettingByName("alarm_no_data_minutes", "5", nil, nil))
	require.NoError(t, (&SettingMap{}).SetSettingByName("alarm_no_data_minutes", "10", nil, nil))
}
