package settings_test

import (
	"testing"

	"github.com/stretchr/testify/require"
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
