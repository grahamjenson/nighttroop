package db_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/settings"
	"nighttroop.t1d.nz/utils/testutils"
)

// A site's settings and an identity's are each one row, read together, and neither is another identity's
func TestSettingsRows(t *testing.T) {
	_, _, site := testutils.NewControlDBWithSite(t)
	ctx, q := t.Context(), site.SiteDB.Queries()

	siteLevel, identityLevel, err := db.LoadSettings(ctx, q, 7)
	require.NoError(t, err)
	assert.Nil(t, siteLevel.GetTimezone(), "a site with no row has no settings of its own")
	assert.Nil(t, identityLevel.GetUnits())

	ofSite, ofSeven, ofEight := &settings.SettingMap{}, &settings.SettingMap{}, &settings.SettingMap{}
	ofSite.SetTimezone(new("Pacific/Auckland"))
	ofSite.SetUnits(new(consts.UnitsMMOLL))
	ofSeven.SetUnits(new(consts.UnitsMGDL))
	ofEight.SetTimezone(new("Europe/London"))
	require.NoError(t, db.SaveSettings(ctx, q, 0, ofSite, "admin"))
	require.NoError(t, db.SaveSettings(ctx, q, 7, ofSeven, "admin"))
	require.NoError(t, db.SaveSettings(ctx, q, 8, ofEight, "admin"))

	siteLevel, identityLevel, err = db.LoadSettings(ctx, q, 7)
	require.NoError(t, err)
	assert.Equal(t, "Pacific/Auckland", *siteLevel.GetTimezone())
	assert.Equal(t, consts.UnitsMMOLL, *siteLevel.GetUnits())
	assert.Equal(t, consts.UnitsMGDL, *identityLevel.GetUnits())
	assert.Nil(t, identityLevel.GetTimezone(), "identity 8's timezone is not identity 7's")

	// No identity reads only the site's, and a save replaces a level whole, so a setting cleared is gone
	_, identityLevel, err = db.LoadSettings(ctx, q, 0)
	require.NoError(t, err)
	assert.Nil(t, identityLevel.GetUnits())
	require.NoError(t, db.SaveSettings(ctx, q, 7, &settings.SettingMap{}, "admin"))
	_, identityLevel, err = db.LoadSettings(ctx, q, 7)
	require.NoError(t, err)
	assert.Nil(t, identityLevel.GetUnits())

	// Saving again and again grows nothing: a level is one row, and an identity's goes with it
	for range 50 {
		require.NoError(t, db.SaveSettings(ctx, q, 8, ofEight, "admin"))
	}
	count, err := q.CountSettings(ctx)
	require.NoError(t, err)
	assert.EqualValues(t, 3, count)
	require.NoError(t, db.DeleteIdentitySettings(ctx, q, 8))
	count, err = q.CountSettings(ctx)
	require.NoError(t, err)
	assert.EqualValues(t, 2, count)
}

// The Nightscout settings documents (AAPS's) and Nighttroop's settings share a table and nothing else: the documents'
// queries never see a settings row, so AAPS cannot read, count, replace or delete one, and a document cannot be read
// as settings
func TestSettingsRowsAreNotNightscoutDocuments(t *testing.T) {
	_, _, site := testutils.NewControlDBWithSite(t)
	ctx, q := t.Context(), site.SiteDB.Queries()
	ofSite := &settings.SettingMap{}
	ofSite.SetTimezone(new("Pacific/Auckland"))
	require.NoError(t, db.SaveSettings(ctx, q, 0, ofSite, "admin"))
	_, err := q.UpsertGeneralSetting(ctx, sitedb.UpsertGeneralSettingParams{Key: "aaps", Type: "aaps", ValueJson: `{"timezone":"Europe/London"}`, CreatedByIdentityName: "phone"})
	require.NoError(t, err)

	documents, err := q.SearchGeneralSettings(ctx, sitedb.SearchGeneralSettingsParams{Limit: 100})
	require.NoError(t, err)
	require.Len(t, documents, 1)
	assert.Equal(t, "aaps", documents[0].Key)
	count, err := q.CountGeneralSettings(ctx)
	require.NoError(t, err)
	assert.EqualValues(t, 1, count)
	bytes, err := q.GeneralSettingsBytes(ctx, sitedb.GeneralSettingsBytesParams{Key: "none"})
	require.NoError(t, err)
	assert.EqualValues(t, len(`{"timezone":"Europe/London"}`), bytes)
	_, err = q.GetGeneralSetting(ctx, sitedb.GetGeneralSettingParams{Key: "nighttroop_site"})
	assert.Error(t, err)
	_, err = q.DeleteGeneralSetting(ctx, sitedb.DeleteGeneralSettingParams{Key: "nighttroop_site"})
	assert.Error(t, err)

	siteLevel, _, err := db.LoadSettings(ctx, q, 0)
	require.NoError(t, err)
	assert.Equal(t, "Pacific/Auckland", *siteLevel.GetTimezone())
}

// A settings row written around SaveSettings, as in an uploaded site DB, is read as far as it is valid and no further
func TestSettingsRowsAreValidatedWhenRead(t *testing.T) {
	_, _, site := testutils.NewControlDBWithSite(t)
	ctx, q := t.Context(), site.SiteDB.Queries()
	require.NoError(t, q.UpsertSettings(ctx, sitedb.UpsertSettingsParams{Key: "nighttroop_site", By: "upload",
		ValueJson: `{"timezone":"Pacific/Auckland","units":"stone","pad":"` + strings.Repeat("x", 100) + `"}`}))
	siteLevel, _, err := db.LoadSettings(ctx, q, 0)
	require.NoError(t, err)
	assert.Equal(t, "Pacific/Auckland", *siteLevel.GetTimezone())
	assert.Nil(t, siteLevel.GetUnits())

	// The column refuses a row no settings could need, whoever writes it
	assert.Error(t, q.UpsertSettings(ctx, sitedb.UpsertSettingsParams{Key: "nighttroop_site", By: "upload",
		ValueJson: `{"pad":"` + strings.Repeat("x", 300_000) + `"}`}))
}

// BenchmarkLoadSettings is what a request pays to read its site's and identity's settings from their rows, each with
// every setting set, among the rows of a site with its most identities and AAPS's documents. Today a request pays no
// query, as the settings are columns of the site and identity rows it has already read.
// Run with go test -run XXX -bench LoadSettings ./data/db
func BenchmarkLoadSettings(b *testing.B) {
	_, _, site := testutils.NewControlDBWithSite(b)
	ctx, q := b.Context(), site.SiteDB.Queries()
	require.NoError(b, db.SaveSettings(ctx, q, 0, settings.DefaultSettingsMap, "admin"))
	for identityID := range int64(100) {
		require.NoError(b, db.SaveSettings(ctx, q, identityID+1, settings.DefaultSettingsMap, "admin"))
	}
	for i := range 100 {
		_, err := q.UpsertGeneralSetting(ctx, sitedb.UpsertGeneralSettingParams{Key: "aaps_clientcontrol_" + strings.Repeat("x", i+1), Type: "aaps_clientcontrol",
			ValueJson: `{"envelope":"` + strings.Repeat("x", 4000) + `"}`, CreatedByIdentityName: "phone"})
		require.NoError(b, err)
	}
	b.ResetTimer()
	for b.Loop() {
		siteLevel, identityLevel, err := db.LoadSettings(ctx, q, 50)
		if err != nil || siteLevel.GetTimezone() == nil || identityLevel.GetTimezone() == nil {
			b.Fatal("settings not read", err)
		}
	}
}
