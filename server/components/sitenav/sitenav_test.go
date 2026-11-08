package sitenav

import (
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/utils/consts"
)

// A path is of the page whose path matches most of it: its own path, the paths under it, and the paths it counts as
// its own, but not a page of no hub
func TestAt(t *testing.T) {
	tests := []struct{ path, want string }{
		{"/data/glucose", "site--data--glucose-button"},
		{"/data/glucose/123", "site--data--glucose-button"},
		{"/reports", "site--reports--agp-button"},
		{"/reports/week-2026-11-01", "site--reports--agp-button"},
		{"/reports/tdd", "site--reports--tdd-menu-button"},
		{"/reports/summaries", "site--reports--summaries-button"},
		{"/settings", "site--settings--settings-button"},
		{"/settings/site", "site--settings--settings-button"},
		{"/settings/identities", "site--settings--identities-button"},
		{"/settings/identity-settings/abc", "site--settings--identities-button"},
		{"/settings/push-devices", "site--settings--push-devices-button"},
		{"/logs/errors", "site--logs--errors-button"},
		{"/presets/food/1", "site--presets--food-button"},
		{"/", ""},
		{"/summary", ""},
		{"/settings/account", ""},
		{"/getting-started", ""},
		{"/data/glucoses", ""},
	}
	for _, test := range tests {
		t.Run(test.path, func(t *testing.T) {
			got := ""
			if page := At(test.path); page != nil {
				got = page.ID
			}
			require.Equal(t, test.want, got)
		})
	}
}

// Only an admin sees the pages that administer the site; everyone sees the data, reports and their own settings
func TestVisible(t *testing.T) {
	ids := func(pages []Page) []string {
		var ids []string
		for _, page := range pages {
			ids = append(ids, page.ID)
		}
		return ids
	}
	admin := &controldb.SiteIdentity{Role: consts.IdentityRoleAdmin}
	require.Len(t, Visible(admin), len(Pages))
	require.Len(t, Visible(admin, Data), 13)
	require.Len(t, Visible(admin, Reports), 3)
	for _, role := range []consts.IdentityRole{consts.IdentityRoleUser, consts.IdentityRoleUserReadOnly, consts.IdentityRoleDeviceRead} {
		settings := ids(Visible(&controldb.SiteIdentity{Role: role}, Settings))
		require.Equal(t, []string{"site--settings--settings-button", "site--settings--alarms-button", "site--settings--push-devices-button"}, settings, role)
		require.Len(t, Visible(&controldb.SiteIdentity{Role: role}, Data), 13, role)
	}
	require.Empty(t, Visible(nil))
}

// Side tabs are in their groups, and search results under their hubs, in order
func TestGroups(t *testing.T) {
	var names []string
	for _, group := range Groups(Visible(&controldb.SiteIdentity{Role: consts.IdentityRoleAdmin}, Settings), false) {
		names = append(names, group.Name)
	}
	require.Equal(t, []string{"This site", "People & devices", "Connections", "Troubleshooting"}, names)
	names = nil
	for _, group := range Groups(Visible(&controldb.SiteIdentity{Role: consts.IdentityRoleAdmin}, Data), false) {
		names = append(names, group.Name)
	}
	require.Equal(t, []string{"Readings & treatments", "Events", "Devices", "Presets"}, names)
	// Food presets and profiles are with the data they are used to enter, not the site's settings
	require.Equal(t, Data, At("/presets/food/1").Hub)
	require.Equal(t, Data, At("/presets/profiles").Hub)
	names = nil
	for _, group := range Groups(Pages, true) {
		names = append(names, group.Name)
	}
	require.Equal(t, []string{"Data", "Reports", "Settings"}, names)
	require.Equal(t, "site--nav--search--settings--push-devices", At("/settings/push-devices").SearchID())
}
