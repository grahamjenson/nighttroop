package db_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/utils/settings"
	"nighttroop.t1d.nz/utils/testutils"
)

// Two saves of a site's settings made from the same row, as two admins saving at once, each write only what they
// changed, so neither undoes the other. A stored value the code no longer accepts stays in the row through both.
func TestSavesAtOnceKeepEachOthersSettings(t *testing.T) {
	_, controlDB, site := testutils.NewControlDBWithSite(t)
	ctx, q, id := t.Context(), controlDB.Queries, site.SiteDB.Subdomain
	read := `{"timezone":"Pacific/Auckland","units":"stone"}`
	require.NoError(t, q.UpdateSiteSettings(ctx, controldb.UpdateSiteSettingsParams{ID: id, SettingsJson: read}))

	first, second := settings.FromJSON([]byte(read), nil, nil), settings.FromJSON([]byte(read), nil, nil)
	first.SetTimezone(new("Europe/London"))
	require.NoError(t, second.SetSettingByName("alarm_low_enabled", true, nil, nil))
	for _, level := range []*settings.SettingMap{first, second} {
		patch, _, err := level.Changes(read)
		require.NoError(t, err)
		require.NoError(t, q.PatchSiteSettings(ctx, controldb.PatchSiteSettingsParams{ID: id, Patch: patch}))
	}

	row, err := q.GetSite(ctx, controldb.GetSiteParams{ID: id})
	require.NoError(t, err)
	assert.JSONEq(t, `{"timezone":"Europe/London","alarm_low_enabled":true,"units":"stone"}`, row.SettingsJson)

	// A setting cleared is taken out of the row, and a patch that is not JSON changes nothing
	third := settings.FromJSON([]byte(row.SettingsJson), nil, nil)
	third.SetTimezone(nil)
	patch, _, err := third.Changes(row.SettingsJson)
	require.NoError(t, err)
	require.NoError(t, q.PatchSiteSettings(ctx, controldb.PatchSiteSettingsParams{ID: id, Patch: patch}))
	require.NoError(t, q.PatchSiteSettings(ctx, controldb.PatchSiteSettingsParams{ID: id, Patch: "not json"}))
	row, err = q.GetSite(ctx, controldb.GetSiteParams{ID: id})
	require.NoError(t, err)
	assert.JSONEq(t, `{"alarm_low_enabled":true,"units":"stone"}`, row.SettingsJson)

	// The row cannot grow past its limit, whatever is patched in
	long := `{"pad":"` + strings.Repeat("x", 5000) + `"}`
	assert.Error(t, q.PatchSiteSettings(ctx, controldb.PatchSiteSettingsParams{ID: id, Patch: long}))
}
