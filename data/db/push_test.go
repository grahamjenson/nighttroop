package db_test

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/testutils"
)

// A push to an app on a site's device uses the site's own credential for it, else one for every site, and only one
// for its provider and app that allows its bundle ID
func TestPushCredentialFor(t *testing.T) {
	_, controlDB, site := testutils.NewControlDBWithSite(t)
	other, err := controlDB.CreateSite(t.Context())
	require.NoError(t, err)
	q, siteID, otherID := controlDB.Queries, site.SiteDB.Subdomain, other.SiteDB.Subdomain

	create := func(siteID *string, name string, provider consts.PushProvider, apps string, bundleIDs *string) int64 {
		credential, err := q.CreatePushCredential(t.Context(), controldb.CreatePushCredentialParams{SiteID: siteID, Name: name,
			Provider: provider, TeamOrProjectID: "TEAM", SecretEnc: new("secret"), AppsJson: apps, BundleIdsJson: bundleIDs,
			CreatedByIdentityName: "admin"})
		require.NoError(t, err)
		return credential.ID
	}
	create(nil, "global loop", consts.PushProviderAPNs, `["loop","t1dcarer"]`, nil)
	create(&siteID, "site loop", consts.PushProviderAPNs, `["loop"]`, nil)
	create(nil, "android", consts.PushProviderFCM, `["t1dcarer"]`, new(`["nz.t1d.carer"]`))
	deleted := create(&siteID, "deleted", consts.PushProviderAPNs, `["trio"]`, nil)
	rows, err := q.DeletePushCredential(t.Context(), controldb.DeletePushCredentialParams{ID: deleted, SiteID: &siteID})
	require.NoError(t, err)
	require.EqualValues(t, 1, rows)

	tests := []struct {
		name     string
		siteID   string
		provider consts.PushProvider
		app      string
		bundleID string
		want     string
	}{
		{"a site's own credential first", siteID, consts.PushProviderAPNs, "loop", "com.TEAM.loopkit.Loop", "site loop"},
		{"another site's is not used, so the one for every site is", otherID, consts.PushProviderAPNs, "loop", "com.TEAM.loopkit.Loop", "global loop"},
		{"one for every site pushes to each of its apps", siteID, consts.PushProviderAPNs, "t1dcarer", "nz.t1d.carer", "global loop"},
		{"by provider", siteID, consts.PushProviderFCM, "t1dcarer", "nz.t1d.carer", "android"},
		{"only to its bundle IDs", siteID, consts.PushProviderFCM, "t1dcarer", "nz.t1d.other", ""},
		{"not to an app it does not list", siteID, consts.PushProviderAPNs, "loopfollow", "com.TEAM.LoopFollow", ""},
		{"not a deleted one", siteID, consts.PushProviderAPNs, "trio", "com.TEAM.Trio", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			credential, err := q.PushCredentialFor(t.Context(), controldb.PushCredentialForParams{SiteID: &tt.siteID,
				Provider: tt.provider, App: tt.app, BundleID: tt.bundleID})
			if tt.want == "" {
				require.ErrorIs(t, err, sql.ErrNoRows)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, credential.Name)
		})
	}
}

// A site's credentials are only its own, and a deleted one keeps no secret
func TestPushCredentialsBySite(t *testing.T) {
	_, controlDB, site := testutils.NewControlDBWithSite(t)
	q, siteID := controlDB.Queries, site.SiteDB.Subdomain
	credential, err := q.CreatePushCredential(t.Context(), controldb.CreatePushCredentialParams{SiteID: &siteID, Name: "mine",
		Provider: consts.PushProviderAPNs, TeamOrProjectID: "TEAM", SecretEnc: new("secret"), AppsJson: `["loop"]`,
		CreatedByIdentityName: "admin"})
	require.NoError(t, err)

	// Not one for every site
	_, err = q.GetPushCredential(t.Context(), controldb.GetPushCredentialParams{ID: credential.ID})
	require.ErrorIs(t, err, sql.ErrNoRows)
	rows, err := q.DeletePushCredential(t.Context(), controldb.DeletePushCredentialParams{ID: credential.ID})
	require.NoError(t, err)
	require.Zero(t, rows)
	global, err := q.SearchPushCredentials(t.Context(), controldb.SearchPushCredentialsParams{Limit: 10})
	require.NoError(t, err)
	require.Empty(t, global)

	count, err := q.CountPushCredentials(t.Context(), controldb.CountPushCredentialsParams{SiteID: &siteID})
	require.NoError(t, err)
	require.EqualValues(t, 1, count)

	// Deleting a site's credentials removes their secrets
	require.NoError(t, q.DeleteSitePushCredentials(t.Context(), controldb.DeleteSitePushCredentialsParams{SiteID: &siteID}))
	_, err = q.GetPushCredential(t.Context(), controldb.GetPushCredentialParams{ID: credential.ID, SiteID: &siteID})
	require.ErrorIs(t, err, sql.ErrNoRows)
	found, err := q.SearchPushCredentials(t.Context(), controldb.SearchPushCredentialsParams{SiteID: &siteID, Limit: 10})
	require.NoError(t, err)
	require.Empty(t, found)
}

// A token is registered to the last identity that sent it, and its registrations go with its identity. The control
// DB is on disk, as a test's in-memory one does not enforce foreign keys.
func TestPushRegistrations(t *testing.T) {
	cfg := *testutils.GetTestServerConfig()
	cfg.Mode, cfg.DataDir = consts.Development, t.TempDir()
	controlDB, err := db.NewControlDB(t.Context(), &cfg)
	require.NoError(t, err)
	t.Cleanup(func() { controlDB.Close() })
	site, err := controlDB.CreateSite(t.Context())
	require.NoError(t, err)
	q, siteID := controlDB.Queries, site.SiteDB.Subdomain
	identities, err := q.AllSiteIdentities(t.Context(), controldb.AllSiteIdentitiesParams{SiteID: siteID})
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(identities), 2)
	first, second := identities[0].ID, identities[1].ID

	register := func(identityID int64, hash string) *controldb.PushRegistration {
		registration, err := q.UpsertPushRegistration(t.Context(), controldb.UpsertPushRegistrationParams{SiteID: siteID,
			IdentityID: identityID, App: consts.PushAppLoop, Provider: consts.PushProviderAPNs, BundleID: "com.TEAM.loopkit.Loop",
			Production: true, TokenEnc: "token", TokenHash: hash})
		require.NoError(t, err)
		return registration
	}
	hashA, hashB := "a000000000000000000000000000000000000000000000000000000000000000", "b000000000000000000000000000000000000000000000000000000000000000"
	moved := register(first, hashA)
	register(first, hashB)
	require.Equal(t, moved.ID, register(second, hashA).ID)

	byIdentity := func(identityID int64) []*controldb.PushRegistration {
		found, err := q.SearchPushRegistrations(t.Context(), controldb.SearchPushRegistrationsParams{SiteID: siteID,
			IdentityID: &identityID, Limit: 10})
		require.NoError(t, err)
		return found
	}
	require.Len(t, byIdentity(first), 1)
	require.Len(t, byIdentity(second), 1)

	// An identity deletes only its own
	rows, err := q.DeletePushRegistration(t.Context(), controldb.DeletePushRegistrationParams{ID: moved.ID, SiteID: siteID, IdentityID: &first})
	require.NoError(t, err)
	require.Zero(t, rows)

	require.NoError(t, q.RecordPush(t.Context(), controldb.RecordPushParams{ID: moved.ID, LastError: new("BadDeviceToken")}))
	require.NoError(t, q.DeletePushRegistrationByTokenHash(t.Context(), controldb.DeletePushRegistrationByTokenHashParams{TokenHash: hashA}))
	require.Empty(t, byIdentity(second))

	// Deleting the identity deletes its registrations
	require.NoError(t, q.DeleteSiteIdentity(t.Context(), controldb.DeleteSiteIdentityParams{SiteID: siteID, ID: first}))
	all, err := q.SearchPushRegistrations(t.Context(), controldb.SearchPushRegistrationsParams{SiteID: siteID, Limit: 10})
	require.NoError(t, err)
	require.Empty(t, all)
}

// A site deleted for good takes its push credentials and its devices' push tokens with it
func TestDeleteSitesRemovesPushData(t *testing.T) {
	_, controlDB, site := testutils.NewControlDBWithSite(t)
	q, siteID := controlDB.Queries, site.SiteDB.Subdomain
	_, err := q.CreatePushCredential(t.Context(), controldb.CreatePushCredentialParams{SiteID: &siteID, Name: "mine",
		Provider: consts.PushProviderAPNs, TeamOrProjectID: "TEAM", SecretEnc: new("secret"), AppsJson: `["loop"]`,
		CreatedByIdentityName: "admin"})
	require.NoError(t, err)
	identities, err := q.AllSiteIdentities(t.Context(), controldb.AllSiteIdentitiesParams{SiteID: siteID})
	require.NoError(t, err)
	_, err = q.UpsertPushRegistration(t.Context(), controldb.UpsertPushRegistrationParams{SiteID: siteID,
		IdentityID: identities[0].ID, App: consts.PushAppLoop, Provider: consts.PushProviderAPNs, BundleID: "com.TEAM.loopkit.Loop",
		Production: true, TokenEnc: "token", TokenHash: "c000000000000000000000000000000000000000000000000000000000000000"})
	require.NoError(t, err)

	require.NoError(t, controlDB.DeleteSites(t.Context(), []string{siteID}))

	credentials, err := q.SearchPushCredentials(t.Context(), controldb.SearchPushCredentialsParams{SiteID: &siteID, Limit: 10})
	require.NoError(t, err)
	require.Empty(t, credentials)
	registrations, err := q.SearchPushRegistrations(t.Context(), controldb.SearchPushRegistrationsParams{SiteID: siteID, Limit: 10})
	require.NoError(t, err)
	require.Empty(t, registrations)
}
