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

// Resetting an identity's password or TOTP, by an admin on the site or in control, removes its passkeys, as for a lost
// or taken device, so it then signs in only with its new credentials; another identity's are kept
func TestResettingAnIdentityRemovesItsPasskeys(t *testing.T) {
	_, controlDB, site := testutils.NewControlDBWithSite(t)
	ctx, q, siteID := t.Context(), controlDB.Queries, site.SiteDB.Subdomain
	identity := func(name string) *controldb.SiteIdentity {
		found, err := q.GetSiteIdentity(ctx, controldb.GetSiteIdentityParams{SiteID: siteID, Name: name})
		require.NoError(t, err)
		return found
	}
	admin, phone := identity(consts.FIRST_ADMIN_NAME), identity(consts.FIRST_PWD_NAME)
	add := func(owner *controldb.SiteIdentity, credentialID string) {
		require.NoError(t, q.CreateSiteIdentityPasskey(ctx, controldb.CreateSiteIdentityPasskeyParams{SiteID: siteID, IdentityID: owner.ID, CredentialID: credentialID, CredentialJson: "{}"}))
	}
	count := func(owner *controldb.SiteIdentity) int {
		passkeys, err := q.ListSiteIdentityPasskeys(ctx, controldb.ListSiteIdentityPasskeysParams{SiteID: siteID, IdentityID: owner.ID})
		require.NoError(t, err)
		return len(passkeys)
	}
	add(admin, "admin-key")
	add(phone, "phone-key")

	_, err := controlDB.ResetSiteIdentityPassword(ctx, admin, "control admin")
	require.NoError(t, err)
	require.Zero(t, count(admin))
	require.Equal(t, 1, count(phone))

	add(admin, "admin-key-again")
	_, err = controlDB.ResetSiteIdentityTOTP(ctx, identity(consts.FIRST_ADMIN_NAME), "control admin")
	require.NoError(t, err)
	require.Zero(t, count(admin))
	require.Equal(t, 1, count(phone))
}
