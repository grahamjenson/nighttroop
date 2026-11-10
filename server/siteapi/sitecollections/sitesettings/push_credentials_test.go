package sitesettings_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/models/sitemodels"
	"nighttroop.t1d.nz/server/siteapi/sitecollections/sitesettings"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/testutils"
)

const p8 = "-----BEGIN PRIVATE KEY-----\nMIGHAgEAMBMGByqGSM49AgEGCCqGSM49AwEHBG0wawIBAQQgofoAD0qFTDB8eVl2\n" +
	"LKYt19HfOaUvrvTYGMMR98DDoDihRANCAAROFNjYuI2rlnHoAy4UP4PVhQiFxz1L\n" +
	"SaETlxByVnuzazzPdJ6qtAO+GOvUV5lPZn62Rv3vFb9ozLi/OFfob09a\n-----END PRIVATE KEY-----\n"

// A site adding and deleting push credentials in a loop, which only counts its live ones against its limit, deletes
// each for good, so it cannot grow the control DB's push_credentials table without bound, and another site's are kept
func TestPushCredentialsDeletedInALoopAreBounded(t *testing.T) {
	cfg, controlDB, site := testutils.NewControlDBWithSite(t)
	other, err := controlDB.CreateSite(t.Context())
	require.NoError(t, err)
	q := controlDB.Queries

	create := func(siteID string) *sitemodels.PushCredential {
		model := &sitemodels.PushCredential{Name: "My Loop", Provider: consts.PushProviderAPNs, TeamOrProjectID: "TEAM123456",
			KeyID: new("ABC123DEFG"), Secret: new(p8), Apps: []string{"loop"}}
		created, err := sitesettings.CreatePushCredential(t.Context(), q, cfg.Security, model, controldb.CreatePushCredentialParams{
			SiteID: &siteID, Name: model.Name, Provider: model.Provider, TeamOrProjectID: model.TeamOrProjectID,
			KeyID: model.KeyID, AppsJson: `["loop"]`, CreatedByIdentityName: "admin"})
		require.NoError(t, err)
		return created
	}
	create(other.SiteDB.Subdomain)
	for range 30 {
		siteID := site.SiteDB.Subdomain
		rows, err := q.DeletePushCredential(t.Context(), controldb.DeletePushCredentialParams{ID: create(siteID).ID, SiteID: &siteID})
		require.NoError(t, err)
		require.EqualValues(t, 1, rows)
	}

	sizes, err := controlDB.TableSizes(t.Context())
	require.NoError(t, err)
	for _, size := range sizes {
		if size.Name == "push_credentials" {
			// Only the other site's
			require.EqualValues(t, 1, size.Records)
			return
		}
	}
	t.Fatal("no push_credentials table")
}
