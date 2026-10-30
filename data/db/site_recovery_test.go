package db_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/testutils"
)

// A backup key works once, and gives the site an admin that can sign in, even when its admin was disabled or deleted
func TestRecoverSiteAdmin(t *testing.T) {
	_, controlDB, site := testutils.NewControlDBWithSite(t)
	ctx, siteID, q := context.Background(), site.SiteDB.Subdomain, controlDB.Queries
	admin := func() *controldb.SiteIdentity {
		identity, err := q.GetSiteIdentity(ctx, controldb.GetSiteIdentityParams{SiteID: siteID, Name: consts.FIRST_ADMIN_NAME})
		require.NoError(t, err)
		return identity
	}

	_, err := controlDB.RecoverSiteAdmin(ctx, siteID, siteID+":notthekey123")
	require.ErrorIs(t, err, db.ErrBadBackupKey)

	// A disabled admin is turned back on, with a new password hash
	before := admin()
	require.NoError(t, q.UpdateSiteIdentityEnabled(ctx, controldb.UpdateSiteIdentityEnabledParams{Enabled: false, SiteID: siteID, ID: before.ID}))
	recovered, err := controlDB.RecoverSiteAdmin(ctx, siteID, site.BackupPassword)
	require.NoError(t, err)
	require.True(t, admin().Enabled)
	require.NotEqual(t, before.PasswordHash, admin().PasswordHash)

	// The key it used works no more, and the one it gave works in its place, making a deleted admin again
	_, err = controlDB.RecoverSiteAdmin(ctx, siteID, site.BackupPassword)
	require.ErrorIs(t, err, db.ErrBadBackupKey)
	require.NoError(t, q.DeleteSiteIdentity(ctx, controldb.DeleteSiteIdentityParams{SiteID: siteID, ID: admin().ID}))
	again, err := controlDB.RecoverSiteAdmin(ctx, siteID, recovered.BackupKey)
	require.NoError(t, err)
	require.Equal(t, string(consts.IdentityRoleAdmin), string(admin().Role))
	require.NotEqual(t, recovered.BackupKey, again.BackupKey)
}
