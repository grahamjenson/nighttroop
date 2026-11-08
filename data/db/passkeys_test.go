package db_test

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/testutils"
)

// Site identities' and control admins' passkeys share one table, but each is only ever found, listed or removed as its
// own owner's: a control admin's passkey never signs in to a site, nor a site's to control
func TestPasskeysKeepTheirOwners(t *testing.T) {
	_, controlDB, site := testutils.NewControlDBWithSite(t)
	ctx, q, siteID := t.Context(), controlDB.Queries, site.SiteDB.Subdomain
	siteAdmin, err := q.GetSiteIdentity(ctx, controldb.GetSiteIdentityParams{SiteID: siteID, Name: consts.FIRST_ADMIN_NAME})
	require.NoError(t, err)
	controlAdmin, err := q.GetControlIdentity(ctx, controldb.GetControlIdentityParams{Name: consts.FIRST_ADMIN_NAME})
	require.NoError(t, err)
	require.NoError(t, q.CreateSiteIdentityPasskey(ctx, controldb.CreateSiteIdentityPasskeyParams{SiteID: siteID, IdentityID: siteAdmin.ID, CredentialID: "site-key", CredentialJson: "{}"}))
	require.NoError(t, q.CreateControlPasskey(ctx, controldb.CreateControlPasskeyParams{ControlIdentityID: controlAdmin.ID, CredentialID: "control-key", CredentialJson: "{}"}))

	// Each is found by its own owner's lookup only
	_, err = q.GetSiteIdentityPasskey(ctx, controldb.GetSiteIdentityPasskeyParams{SiteID: siteID, CredentialID: "control-key"})
	require.ErrorIs(t, err, sql.ErrNoRows)
	_, err = q.GetControlPasskey(ctx, controldb.GetControlPasskeyParams{CredentialID: "site-key"})
	require.ErrorIs(t, err, sql.ErrNoRows)
	controlKey, err := q.GetControlPasskey(ctx, controldb.GetControlPasskeyParams{CredentialID: "control-key"})
	require.NoError(t, err)
	require.Equal(t, controlAdmin.ID, *controlKey.ControlIdentityID)
	require.Nil(t, controlKey.SiteID)
	siteKey, err := q.GetSiteIdentityPasskey(ctx, controldb.GetSiteIdentityPasskeyParams{SiteID: siteID, CredentialID: "site-key"})
	require.NoError(t, err)
	require.Nil(t, siteKey.ControlIdentityID)

	// Removing one as the other's owner removes nothing
	require.NoError(t, q.DeleteSiteIdentityPasskey(ctx, controldb.DeleteSiteIdentityPasskeyParams{SiteID: siteID, IdentityID: siteAdmin.ID, ID: controlKey.ID}))
	require.NoError(t, q.DeleteControlPasskey(ctx, controldb.DeleteControlPasskeyParams{ControlIdentityID: controlAdmin.ID, ID: siteKey.ID}))
	controlKeys, err := q.ListControlPasskeys(ctx, controldb.ListControlPasskeysParams{ControlIdentityID: controlAdmin.ID})
	require.NoError(t, err)
	require.Len(t, controlKeys, 1)
	siteKeys, err := q.ListSiteIdentityPasskeys(ctx, controldb.ListSiteIdentityPasskeysParams{SiteID: siteID, IdentityID: siteAdmin.ID})
	require.NoError(t, err)
	require.Len(t, siteKeys, 1)

	// And as its own, it is removed
	require.NoError(t, q.DeleteControlPasskey(ctx, controldb.DeleteControlPasskeyParams{ControlIdentityID: controlAdmin.ID, ID: controlKey.ID}))
	controlKeys, err = q.ListControlPasskeys(ctx, controldb.ListControlPasskeysParams{ControlIdentityID: controlAdmin.ID})
	require.NoError(t, err)
	require.Empty(t, controlKeys)
}
