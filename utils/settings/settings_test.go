package settings_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/types"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/settings"
)

func TestCombinedSettingsUnits(t *testing.T) {
	mmol, mgdl := consts.UnitsMMOLL, consts.UnitsMGDL
	tests := []struct {
		name     string
		site     *consts.Units
		identity *consts.Units
		want     consts.Units
	}{
		{name: "nothing set uses the default", want: mmol},
		{name: "site set", site: &mgdl, want: mgdl},
		{name: "identity set", identity: &mgdl, want: mgdl},
		{name: "identity overrides site mg/dL", site: &mmol, identity: &mgdl, want: mgdl},
		{name: "identity overrides site mmol/L", site: &mgdl, identity: &mmol, want: mmol},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			site, identity := &settings.SettingMap{}, &settings.SettingMap{}
			site.SetUnits(test.site)
			identity.SetUnits(test.identity)
			combined := &settings.CombinedSettingsMap{
				SiteSettingsMap:     site,
				IdentitySettingsMap: identity,
				DefaultSettingsMap:  settings.DefaultSettingsMap,
			}

			require.Equal(t, test.want, *combined.GetUnits())
			// The same order as looking a setting up by name
			byName, err := combined.GetSettingAnyTypeByName("units")
			require.NoError(t, err)
			require.Equal(t, test.want, byName.Value)
		})
	}
}

func TestCombinedSettingsTimezone(t *testing.T) {
	auckland, london, notAZone := "Pacific/Auckland", "Europe/London", "Not/A_Zone"
	tests := []struct {
		name     string
		site     *string
		identity *string
		want     string
	}{
		{name: "nothing set is UTC", want: "UTC"},
		{name: "site set", site: &auckland, want: auckland},
		{name: "identity set", identity: &london, want: london},
		{name: "identity overrides site", site: &auckland, identity: &london, want: london},
		{name: "a zone that cannot be loaded is UTC", identity: &notAZone, want: "UTC"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			site, identity := &settings.SettingMap{}, &settings.SettingMap{}
			site.SetTimezone(test.site)
			identity.SetTimezone(test.identity)
			combined := &settings.CombinedSettingsMap{
				SiteSettingsMap:     site,
				IdentitySettingsMap: identity,
				DefaultSettingsMap:  settings.DefaultSettingsMap,
			}

			require.Equal(t, test.want, combined.Location().String())
		})
	}
}

func TestDeviceBatteryLowDevices(t *testing.T) {
	tests := []struct {
		name  string
		value any
		want  types.DeviceNames
		err   bool
	}{
		{"a list of names", []any{"openaps://pixel", "SugarPixel"}, types.DeviceNames{"openaps://pixel", "SugarPixel"}, false},
		{"none", []any{}, types.DeviceNames{}, false},
		{"not a list", "openaps://pixel", nil, true},
		{"a name that is not a string", []any{"phone", 3}, nil, true},
		{"an empty name", []any{""}, nil, true},
		{"a name too long", []any{strings.Repeat("a", 65)}, nil, true},
		{"a name twice", []any{"phone", "phone"}, nil, true},
		{"too many names", []any{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11"}, nil, true},
	}
	for _, tt := range tests {
		got, err := settings.AlarmDeviceBatteryLowDevicesJson.ParseValidate(tt.value)
		if tt.err {
			require.Error(t, err, tt.name)
			continue
		}
		require.NoError(t, err, tt.name)
		require.Equal(t, tt.want, *got, tt.name)
	}
}

// A fast rise or drop is measured over at least 5 minutes, as no change is seen over none
func TestFastRiseAndDropMinutes(t *testing.T) {
	for _, minutes := range []*settings.SettingDefinition[int64]{settings.AlarmFastRiseMinutes, settings.AlarmFastDropMinutes} {
		require.Error(t, minutes.ValidateFunc(4), minutes.Identifier)
		require.NoError(t, minutes.ValidateFunc(5), minutes.Identifier)
		require.NoError(t, minutes.ValidateFunc(settings.MaxAlarmMinutes), minutes.Identifier)
	}
}

// Times and dates are shown in the identity's format, then the site's, then the default, each as its Go layout
func TestCombinedSettingsLayouts(t *testing.T) {
	layoutOf := map[string]func(*settings.CombinedSettingsMap) string{
		"time_format": (*settings.CombinedSettingsMap).TimeLayout,
		"date_format": (*settings.CombinedSettingsMap).DateLayout,
	}
	tests := []struct {
		setting        string
		site, identity string
		layout         string
	}{
		{"time_format", "", "", "15:04:05"},
		{"time_format", "hh:mm:ss a", "", "03:04:05 PM"},
		{"time_format", "hh:mm:ss a", "HH:mm", "15:04"},
		{"time_format", "", "hh:mm a", "03:04 PM"},
		{"time_format", "", "HH:mm:ss", "15:04:05"},
		{"date_format", "", "", "2006-01-02"},
		{"date_format", "DD/MM/YYYY", "", "02/01/2006"},
		{"date_format", "DD/MM/YYYY", "MM/DD/YYYY", "01/02/2006"},
		{"date_format", "", "YYYY/MM/DD", "2006/01/02"},
		{"date_format", "", "DD-MM-YYYY", "02-01-2006"},
		{"date_format", "", "MM-DD-YYYY", "01-02-2006"},
		{"date_format", "", "YYYY-MM-DD", "2006-01-02"},
	}
	for _, test := range tests {
		t.Run(test.setting+"/"+test.site+"_"+test.identity, func(t *testing.T) {
			site, identity := &settings.SettingMap{}, &settings.SettingMap{}
			for level, format := range map[*settings.SettingMap]string{site: test.site, identity: test.identity} {
				if format != "" {
					require.NoError(t, level.SetSettingByName(test.setting, format, nil, nil))
				}
			}
			combined := &settings.CombinedSettingsMap{SiteSettingsMap: site, IdentitySettingsMap: identity, DefaultSettingsMap: settings.DefaultSettingsMap}
			require.Equal(t, test.layout, layoutOf[test.setting](combined))
		})
	}
}

func TestLoadLocation(t *testing.T) {
	tests := []struct {
		name, timezone string
		loads          bool
	}{
		{name: "a timezone loads", timezone: "Pacific/Auckland", loads: true},
		{name: "UTC loads", timezone: "UTC", loads: true},
		{name: "an unknown timezone does not load", timezone: "Nowhere/Atlantis"},
		{name: "a path is not a timezone", timezone: "../../etc/passwd"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			first, err := settings.LoadLocation(test.timezone)
			if !test.loads {
				require.Error(t, err)
				require.Nil(t, first)
				return
			}
			require.NoError(t, err)
			require.Equal(t, test.timezone, first.String())
			again, err := settings.LoadLocation(test.timezone)
			require.NoError(t, err)
			require.Same(t, first, again, "a timezone is loaded once and then kept")
		})
	}
}
