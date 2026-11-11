package sitemodels_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/models/sitemodels"
	"nighttroop.t1d.nz/utils/consts"
)

// LoopSettingsChanges are the settings that differ between two sets, old to new, with glucose in the site's units
func TestLoopSettingsChanges(t *testing.T) {
	mmol, mgdl := new(consts.UnitsMMOLL), new(consts.UnitsMGDL)
	tests := []struct {
		name    string
		app     string
		before  string
		after   string
		units   *consts.Units
		changes [][3]string // field, old and new, "-" when not sent
	}{
		{name: "the same settings have no changes", app: "loop", before: `{"maximumBolus":10}`, after: `{"maximumBolus":10}`, units: mmol},
		{name: "a changed number", app: "loop", before: `{"maximumBolus":10,"dosingEnabled":true}`, after: `{"maximumBolus":12.5,"dosingEnabled":true}`, units: mmol,
			changes: [][3]string{{"maximumBolus", "10", "12.5"}}},
		{name: "text and a switch", app: "loop", before: `{"dosingStrategy":"automaticBolus","dosingEnabled":true}`, after: `{"dosingStrategy":"tempBasalOnly","dosingEnabled":false}`, units: mmol,
			changes: [][3]string{{"dosingEnabled", "true", "false"}, {"dosingStrategy", "automaticBolus", "tempBasalOnly"}}},
		{name: "added and removed", app: "loop", before: `{"maximumBolus":10}`, after: `{"maximumBasalRatePerHour":3}`, units: mmol,
			changes: [][3]string{{"maximumBasalRatePerHour", "-", "3"}, {"maximumBolus", "10", "-"}}},
		{name: "glucose sent in mg/dL is in mmol/L", app: "loop", before: `{"minimumBGGuard":80,"preMealTargetRange":[80,90]}`, after: `{"minimumBGGuard":90,"preMealTargetRange":[90,100]}`, units: mmol,
			changes: [][3]string{{"minimumBGGuard", "4.4 mmol/L", "5.0 mmol/L"}, {"preMealTargetRange", "4.4–5.0 mmol/L", "5.0–5.5 mmol/L"}}},
		{name: "glucose sent in mmol/L is in mg/dL", app: "loop", before: `{"minimumBGGuard":4.5}`, after: `{"minimumBGGuard":5}`, units: mgdl,
			changes: [][3]string{{"minimumBGGuard", "81 mg/dL", "90 mg/dL"}}},
		{name: "glucose that reads the same is no change", app: "loop", before: `{"minimumBGGuard":80}`, after: `{"minimumBGGuard":80.1}`, units: mgdl},
		{name: "presets are matched by name, wherever they are in the list", app: "loop",
			before: `{"overridePresets":[{"name":"Run","targetRange":[140,160],"duration":3600},{"name":"Sick","insulinNeedsScaleFactor":1.2}]}`,
			after:  `{"overridePresets":[{"name":"Eat","targetRange":[100,110]},{"name":"Run","targetRange":[150,160],"duration":3600},{"name":"Sick","insulinNeedsScaleFactor":1.2}]}`,
			units:  mgdl,
			changes: [][3]string{{"overridePresets.Eat.name", "-", "Eat"}, {"overridePresets.Eat.targetRange", "-", "100–110 mg/dL"},
				{"overridePresets.Run.targetRange", "140–160 mg/dL", "150–160 mg/dL"}}},
		{name: "presets without names, or with one used before, are by their place", app: "loop",
			before:  `{"overridePresets":[{"duration":60},{"name":"Run","duration":30},{"name":"Run","duration":30}]}`,
			after:   `{"overridePresets":[{"duration":90},{"name":"Run","duration":30},{"name":"Run","duration":45}]}`,
			units:   mgdl,
			changes: [][3]string{{"overridePresets.1.duration", "60", "90"}, {"overridePresets.3.duration", "30", "45"}}},
		{name: "Trio's override targets are glucose", app: "trio", before: `{"overridePresets":[{"name":"Gym","target":120}]}`, after: `{"overridePresets":[{"name":"Gym","target":135}]}`, units: mmol,
			changes: [][3]string{{"overridePresets.Gym.target", "6.7 mmol/L", "7.5 mmol/L"}}},
		{name: "AAPS's nested configuration, which has no glucose", app: "aaps",
			before: `{"insulin":{"peak":75},"target":100,"plugins":["a","b"]}`, after: `{"insulin":{"peak":55},"target":110,"plugins":["a"]}`, units: mmol,
			changes: [][3]string{{"insulin.peak", "75", "55"}, {"plugins", `["a","b"]`, `["a"]`}, {"target", "100", "110"}}},
		{name: "settings that cannot be read are none", app: "loop", before: `not json`, after: `{"maximumBolus":10}`, units: mmol,
			changes: [][3]string{{"maximumBolus", "-", "10"}}},
	}
	text := func(value *string) string {
		if value == nil {
			return "-"
		}
		return *value
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			changes := [][3]string{}
			for _, change := range sitemodels.LoopSettingsChanges(tt.app, json.RawMessage(tt.before), json.RawMessage(tt.after), tt.units) {
				changes = append(changes, [3]string{change.Field, text(change.Old), text(change.New)})
			}
			if tt.changes == nil {
				tt.changes = [][3]string{}
			}
			require.Equal(t, tt.changes, changes)
		})
	}
}

// LoopSettingsFor leaves the device token out of the settings for every identity that cannot change the site's data
func TestLoopSettingsFor(t *testing.T) {
	loop := `{"bundleIdentifier":"com.X.loopkit.Loop","deviceToken":"abcd","maximumBolus":10}`
	hidden := `{"bundleIdentifier":"com.X.loopkit.Loop","maximumBolus":10}`
	tests := []struct {
		name     string
		identity *controldb.SiteIdentity
		settings string
		want     string // "" for none
	}{
		{name: "an admin sees it", identity: &controldb.SiteIdentity{Role: consts.IdentityRoleAdmin}, settings: loop, want: loop},
		{name: "a user sees it", identity: &controldb.SiteIdentity{Role: consts.IdentityRoleUser}, settings: loop, want: loop},
		{name: "the PWD phone sees it", identity: &controldb.SiteIdentity{Role: consts.IdentityRoleDevicePWDPhone}, settings: loop, want: loop},
		{name: "a read and write device sees it", identity: &controldb.SiteIdentity{Role: consts.IdentityRoleDeviceReadWrite}, settings: loop, want: loop},
		{name: "a read only user does not", identity: &controldb.SiteIdentity{Role: consts.IdentityRoleUserReadOnly}, settings: loop, want: hidden},
		{name: "a read only device does not", identity: &controldb.SiteIdentity{Role: consts.IdentityRoleDeviceRead}, settings: loop, want: hidden},
		{name: "a clinician does not", identity: &controldb.SiteIdentity{Role: consts.IdentityRoleClinician}, settings: loop, want: hidden},
		{name: "no identity does not", settings: loop, want: hidden},
		{name: "settings without one are as they were", settings: `{"maximumBolus":10, "dosingEnabled":true}`, want: `{"maximumBolus":10, "dosingEnabled":true}`},
		{name: "settings that cannot be read are none", settings: `[{"deviceToken":"abcd"}]`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sitemodels.LoopSettingsFor(tt.identity, json.RawMessage(tt.settings))
			if tt.want == "" {
				require.Nil(t, got)
				return
			}
			require.Equal(t, tt.want, string(got))
		})
	}
}
