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
			site: controldb.Site{AlarmLowEnabled: new(true), AlarmLow: new(int64(70))},
			want: []Alarm{LowAlarm{Mgdl: 70, Minutes: 5}},
		},
		{
			name:     "the identity's settings win over the site's",
			site:     controldb.Site{AlarmUrgentLowEnabled: new(true), AlarmUrgentLow: new(int64(50))},
			identity: controldb.SiteIdentity{AlarmUrgentLow: new(int64(60))},
			want:     []Alarm{UrgentLowAlarm{Mgdl: 60}},
		},
		{
			name:     "the identity can turn off the site's alarm",
			site:     controldb.Site{AlarmNoDataEnabled: new(true)},
			identity: controldb.SiteIdentity{AlarmNoDataEnabled: new(false)},
		},
		{
			name: "every alarm on, at the defaults, in the order they are checked",
			identity: controldb.SiteIdentity{
				AlarmUrgentLowEnabled: new(true), AlarmLowEnabled: new(true), AlarmUrgentHighEnabled: new(true),
				AlarmHighEnabled: new(true), AlarmNoDataEnabled: new(true),
			},
			want: []Alarm{
				UrgentLowAlarm{Mgdl: 56}, LowAlarm{Mgdl: 72, Minutes: 5}, UrgentHighAlarm{Mgdl: 270},
				HighAlarm{Mgdl: 180, Minutes: 30}, NoDataAlarm{Minutes: 20},
			},
		},
		{
			name:     "the fast rise and drop, at the defaults, rising or dropping 18 mg/dL within 15 minutes",
			identity: controldb.SiteIdentity{AlarmFastRiseEnabled: new(true), AlarmFastDropEnabled: new(true)},
			want:     []Alarm{FastRiseAlarm{Mgdl: 18, Minutes: 15}, FastDropAlarm{Mgdl: 18, Minutes: 15}},
		},
		{
			name:     "the site's fast drop, the identity's minutes",
			site:     controldb.Site{AlarmFastDropEnabled: new(true), AlarmFastDrop: new(int64(27))},
			identity: controldb.SiteIdentity{AlarmFastDropMinutes: new(int64(10))},
			want:     []Alarm{FastDropAlarm{Mgdl: 27, Minutes: 10}},
		},
		{
			name:     "the device battery alarm checks nothing with no devices listed",
			identity: controldb.SiteIdentity{AlarmDeviceBatteryLowEnabled: new(true)},
		},
		{
			name:     "the device battery alarm checks the devices listed, at the defaults",
			site:     controldb.Site{AlarmDeviceBatteryLowDevicesJson: &types.DeviceNames{"openaps://pixel", "SugarPixel"}},
			identity: controldb.SiteIdentity{AlarmDeviceBatteryLowEnabled: new(true)},
			want:     []Alarm{DeviceBatteryAlarm{Percent: 20, Minutes: 30, Devices: []string{"openaps://pixel", "SugarPixel"}}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, FromSettings(settings.NewCombinedSettingsMap(&tt.site, &tt.identity)))
		})
	}
}
