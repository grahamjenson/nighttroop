package alarms

import (
	"testing"

	"github.com/stretchr/testify/require"

	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/data/types"
	"nighttroop.t1d.nz/utils/settings"
)

func TestFromSettings(t *testing.T) {
	tests := []struct {
		name     string
		site     controldb.Site
		identity controldb.SiteIdentity
		want     []Alarm
	}{
		{
			name: "every alarm is off by default",
		},
		{
			name: "the site's alarm, with the default settings it leaves out",
			site: controldb.Site{SettingsJson: settings.StoredOf(settings.Settings{Alarms: settings.Alarms{Low: settings.TimedAlarm{Enabled: new(true)}}, Range: settings.GlucoseRange{LowMgdl: new(int64(70))}})},
			want: []Alarm{LowAlarm{Mgdl: 70, Minutes: 5}},
		},
		{
			name:     "the identity's settings win over the site's",
			site:     controldb.Site{SettingsJson: settings.StoredOf(settings.Settings{Alarms: settings.Alarms{UrgentLow: settings.Alarm{Enabled: new(true)}}, Range: settings.GlucoseRange{UrgentLowMgdl: new(int64(50))}})},
			identity: controldb.SiteIdentity{SettingsJson: settings.StoredOf(settings.Settings{Range: settings.GlucoseRange{UrgentLowMgdl: new(int64(60))}})},
			want:     []Alarm{UrgentLowAlarm{Mgdl: 60}},
		},
		{
			name:     "the identity can turn off the site's alarm",
			site:     controldb.Site{SettingsJson: settings.StoredOf(settings.Settings{Alarms: settings.Alarms{NoData: settings.TimedAlarm{Enabled: new(true)}}})},
			identity: controldb.SiteIdentity{SettingsJson: settings.StoredOf(settings.Settings{Alarms: settings.Alarms{NoData: settings.TimedAlarm{Enabled: new(false)}}})},
		},
		{
			name:     "every alarm on, at the defaults, in the order they are checked",
			identity: controldb.SiteIdentity{SettingsJson: settings.StoredOf(settings.Settings{Alarms: settings.Alarms{UrgentLow: settings.Alarm{Enabled: new(true)}, Low: settings.TimedAlarm{Enabled: new(true)}, UrgentHigh: settings.Alarm{Enabled: new(true)}, High: settings.TimedAlarm{Enabled: new(true)}, NoData: settings.TimedAlarm{Enabled: new(true)}}})},
			want: []Alarm{
				UrgentLowAlarm{Mgdl: 56}, LowAlarm{Mgdl: 72, Minutes: 5}, UrgentHighAlarm{Mgdl: 270},
				HighAlarm{Mgdl: 180, Minutes: 30}, NoDataAlarm{Minutes: 20},
			},
		},
		{
			name:     "the predicted low follows the trend 20 minutes on to the low's glucose, by default",
			site:     controldb.Site{SettingsJson: settings.StoredOf(settings.Settings{Range: settings.GlucoseRange{LowMgdl: new(int64(80))}})},
			identity: controldb.SiteIdentity{SettingsJson: settings.StoredOf(settings.Settings{Alarms: settings.Alarms{LowPrediction: settings.TimedAlarm{Enabled: new(true)}}})},
			want:     []Alarm{PredictedLowAlarm{Mgdl: 80, Minutes: 20}},
		},
		{
			name:     "the fast rise and drop, at the defaults, rising or dropping 18 mg/dL within 15 minutes",
			identity: controldb.SiteIdentity{SettingsJson: settings.StoredOf(settings.Settings{Alarms: settings.Alarms{FastRise: settings.ChangeAlarm{Enabled: new(true)}, FastDrop: settings.ChangeAlarm{Enabled: new(true)}}})},
			want:     []Alarm{FastRiseAlarm{Mgdl: 18, Minutes: 15}, FastDropAlarm{Mgdl: 18, Minutes: 15}},
		},
		{
			name:     "the site's fast drop, the identity's minutes",
			site:     controldb.Site{SettingsJson: settings.StoredOf(settings.Settings{Alarms: settings.Alarms{FastDrop: settings.ChangeAlarm{Enabled: new(true), Mgdl: new(int64(27))}}})},
			identity: controldb.SiteIdentity{SettingsJson: settings.StoredOf(settings.Settings{Alarms: settings.Alarms{FastDrop: settings.ChangeAlarm{Minutes: new(int64(10))}}})},
			want:     []Alarm{FastDropAlarm{Mgdl: 27, Minutes: 10}},
		},
		{
			name:     "the device battery alarm checks nothing with no devices listed",
			identity: controldb.SiteIdentity{SettingsJson: settings.StoredOf(settings.Settings{Alarms: settings.Alarms{DeviceBatteryLow: settings.BatteryAlarm{Enabled: new(true)}}})},
		},
		{
			name:     "the device battery alarm checks the devices listed, at the defaults",
			site:     controldb.Site{SettingsJson: settings.StoredOf(settings.Settings{Alarms: settings.Alarms{DeviceBatteryLow: settings.BatteryAlarm{Devices: &types.DeviceNames{"openaps://pixel", "SugarPixel"}}}})},
			identity: controldb.SiteIdentity{SettingsJson: settings.StoredOf(settings.Settings{Alarms: settings.Alarms{DeviceBatteryLow: settings.BatteryAlarm{Enabled: new(true)}}})},
			want:     []Alarm{DeviceBatteryAlarm{Percent: 20, Minutes: 30, Devices: []string{"openaps://pixel", "SugarPixel"}}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, FromSettings(settings.NewContextSettings(&tt.site, &tt.identity).Values()))
		})
	}
}
