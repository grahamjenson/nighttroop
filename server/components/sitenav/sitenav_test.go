package sitenav

import (
	"slices"
	"strings"
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
		{"/reports", "site--reports--summaries-button"},
		{"/reports/summaries", "site--reports--summaries-button"},
		{"/reports/overview/month-2026-11-01", "site--reports--overview-button"},
		{"/reports/agp", "site--reports--agp-button"},
		{"/reports/agp/week-2026-11-01", "site--reports--agp-button"},
		{"/reports/tdd", "site--reports--tdd-menu-button"},
		{"/settings", "site--settings--settings-button"},
		{"/settings/site", "site--settings--settings-button"},
		{"/settings/identities", "site--settings--identities-button"},
		{"/settings/identity-settings/abc", "site--settings--identities-button"},
		{"/settings/push-devices", "site--settings--push-devices-button"},
		{"/settings/alarms", "site--settings--alarms-button"},
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
	require.Len(t, Visible(admin, Reports), 4)
	require.Len(t, Visible(admin, Settings), 4)
	require.Equal(t, []string{"site--settings--connectors-button"}, ids(Visible(admin, Connectors)))
	require.Equal(t, []string{"site--settings--alarms-button"}, ids(Visible(admin, Alarms)))
	require.Equal(t, []string{"site--logs--events-button", "site--logs--errors-button", "site--logs--logs-button"}, ids(Visible(admin, Logs)))
	require.Equal(t, []string{"site--usage-button"}, ids(Visible(admin, Usage)))
	require.Len(t, Visible(admin, MenuHubs...), 10)
	for _, role := range []consts.IdentityRole{consts.IdentityRoleUser, consts.IdentityRoleUserReadOnly, consts.IdentityRoleDeviceRead} {
		settings := ids(Visible(&controldb.SiteIdentity{Role: role}, Settings))
		require.Equal(t, []string{"site--settings--settings-button", "site--settings--push-devices-button"}, settings, role)
		require.Equal(t, []string{"site--settings--alarms-button"}, ids(Visible(&controldb.SiteIdentity{Role: role}, Alarms)), role)
		require.Empty(t, Visible(&controldb.SiteIdentity{Role: role}, Logs), role)
		require.Empty(t, Visible(&controldb.SiteIdentity{Role: role}, Usage), role)
		require.Empty(t, Visible(&controldb.SiteIdentity{Role: role}, Connectors), role)
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
	require.Equal(t, []string{"This site", "People & devices", "Push"}, names)
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
	require.Equal(t, []string{"Data", "Reports", "Settings", "Connectors", "Alarms", "Logs & Events", "Usage"}, names)
	require.Equal(t, "site--nav--search--settings--push-devices", At("/settings/push-devices").SearchID())
	// Alarms and the logs and events are hubs of their own, out of Settings
	require.Equal(t, Alarms, At("/settings/alarms").Hub)
	require.Equal(t, Logs, At("/logs/errors").Hub)
}

// The header opens Data and Reports, and the account menu Settings, Connectors, Alarms, Logs & Events and Usage, every
// hub once, each
// named in IDs by its first word
func TestHubs(t *testing.T) {
	var all []Hub
	for _, group := range Groups(Pages, true) {
		all = append(all, Hub(group.Name))
	}
	require.Equal(t, all, append(slices.Clone(HeaderHubs), MenuHubs...))
	var slugs []string
	for _, hub := range all {
		slugs = append(slugs, hub.Slug())
	}
	require.Equal(t, []string{"data", "reports", "settings", "connectors", "alarms", "logs", "usage"}, slugs)
}

// Search finds a page by the headings and words on it and by what people also call it, as the header's search
// matches: every word typed is in the page's search text
func TestSearchText(t *testing.T) {
	tests := []struct {
		query string
		want  []string
	}{
		{"nightscout", []string{"site--settings--connectors-button"}},
		{"dexcom", []string{"site--settings--connectors-button"}},
		{"low", []string{"site--settings--alarms-button"}},
		{"urgent high", []string{"site--settings--alarms-button"}},
		{"bgl", []string{"site--data--glucose-button"}},
		{"blood sugar", []string{"site--data--glucose-button"}},
		{"bolus", []string{"site--data--insulin-button", "site--presets--food-button", "site--presets--loop-settings-button"}},
		{"carbs", []string{"site--data--food-button", "site--presets--food-button", "site--reports--summaries-button"}},
		{"timezone", []string{"site--settings--settings-button"}},
		{"iob", []string{"site--data--device-status-button"}},
		{"total daily dose", []string{"site--reports--tdd-menu-button"}},
		{"exercise", []string{"site--data--activities-button"}},
		{"push", []string{"site--settings--push-devices-button", "site--settings--push-button"}},
		{"temp", []string{"site--data--temp-targets-button", "site--data--temp-basals-button"}},
	}
	for _, test := range tests {
		t.Run(test.query, func(t *testing.T) {
			var got []string
			for _, page := range Pages {
				text := page.SearchText()
				if !slices.ContainsFunc(strings.Fields(test.query), func(word string) bool { return !strings.Contains(text, word) }) {
					got = append(got, page.ID)
				}
			}
			require.Equal(t, test.want, got)
		})
	}
	// A user who is not an admin finds none of the logs, as no page they may open mentions them
	for _, page := range Visible(&controldb.SiteIdentity{Role: consts.IdentityRoleUser}) {
		require.NotContains(t, page.SearchText(), "logs", page.ID)
	}
}

// A page's hints are its keywords, and a page with none has none
func TestHints(t *testing.T) {
	require.Equal(t, "Nightscout|Dexcom|Libre|LibreLinkUp|Virtual|Backfilling", Page{Path: "/settings/connectors"}.Hints())
	require.Empty(t, Page{Path: "/logs/errors"}.Hints())
}
