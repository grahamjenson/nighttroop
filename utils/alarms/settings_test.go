package alarms

import (
	"cmp"
	"testing"

	"github.com/stretchr/testify/require"

	"nighttroop.t1d.nz/utils/settings"
)

func TestFromSettings(t *testing.T) {
	tests := []struct {
		name     string
		site     *settings.SettingMap
		identity *settings.SettingMap
		want     []Alarm
	}{
		{
			name: "every alarm is off by default",
		},
		{
			name: "the site's alarm, with the default settings it leaves out",
			site: settings.Of(map[string]any{"alarm_low_enabled": true, "range_low_mgdl": 70}),
			want: []Alarm{LowAlarm{Mgdl: 70, Minutes: 5}},
		},
		{
			name:     "the identity's settings win over the site's",
			site:     settings.Of(map[string]any{"alarm_urgent_low_enabled": true, "range_urgent_low_mgdl": 50}),
			identity: settings.Of(map[string]any{"range_urgent_low_mgdl": 60}),
			want:     []Alarm{UrgentLowAlarm{Mgdl: 60}},
		},
		{
			name:     "the identity can turn off the site's alarm",
			site:     settings.Of(map[string]any{"alarm_no_data_enabled": true}),
			identity: settings.Of(map[string]any{"alarm_no_data_enabled": false}),
		},
		{
			name:     "every alarm on, at the defaults, in the order they are checked",
			identity: settings.Of(map[string]any{"alarm_urgent_low_enabled": true, "alarm_low_enabled": true, "alarm_urgent_high_enabled": true, "alarm_high_enabled": true, "alarm_no_data_enabled": true}),
			want: []Alarm{
				UrgentLowAlarm{Mgdl: 56}, LowAlarm{Mgdl: 72, Minutes: 5}, UrgentHighAlarm{Mgdl: 270},
				HighAlarm{Mgdl: 180, Minutes: 30}, NoDataAlarm{Minutes: 20},
			},
		},
		{
			name:     "the predicted low follows the trend 20 minutes on to the low's glucose, by default",
			site:     settings.Of(map[string]any{"range_low_mgdl": 80}),
			identity: settings.Of(map[string]any{"alarm_low_prediction_enabled": true}),
			want:     []Alarm{PredictedLowAlarm{Mgdl: 80, Minutes: 20}},
		},
		{
			name:     "the fast rise and drop, at the defaults, rising or dropping 18 mg/dL within 15 minutes",
			identity: settings.Of(map[string]any{"alarm_fast_rise_enabled": true, "alarm_fast_drop_enabled": true}),
			want:     []Alarm{FastRiseAlarm{Mgdl: 18, Minutes: 15}, FastDropAlarm{Mgdl: 18, Minutes: 15}},
		},
		{
			name:     "the site's fast drop, the identity's minutes",
			site:     settings.Of(map[string]any{"alarm_fast_drop_enabled": true, "alarm_fast_drop_mgdl": 27}),
			identity: settings.Of(map[string]any{"alarm_fast_drop_minutes": 10}),
			want:     []Alarm{FastDropAlarm{Mgdl: 27, Minutes: 10}},
		},
		{
			name:     "the device battery alarm checks nothing with no devices listed",
			identity: settings.Of(map[string]any{"alarm_device_battery_low_enabled": true}),
		},
		{
			name:     "the device battery alarm checks the devices listed, at the defaults",
			site:     settings.Of(map[string]any{"alarm_device_battery_low_devices": []any{"openaps://pixel", "SugarPixel"}}),
			identity: settings.Of(map[string]any{"alarm_device_battery_low_enabled": true}),
			want:     []Alarm{DeviceBatteryAlarm{Percent: 20, Minutes: 30, Devices: []string{"openaps://pixel", "SugarPixel"}}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, FromSettings(settings.NewCombinedSettingsMap(cmp.Or(tt.site, &settings.SettingMap{}), cmp.Or(tt.identity, &settings.SettingMap{}))))
		})
	}
}
